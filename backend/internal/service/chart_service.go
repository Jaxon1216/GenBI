package service

import (
	"errors"
	"fmt"

	"genbi-go-backend/internal/common"
	"genbi-go-backend/internal/dto"
	"genbi-go-backend/internal/model"
	"genbi-go-backend/internal/util"

	"gorm.io/gorm"
)

// ChartService 图表业务。
type ChartService struct {
	db       *gorm.DB
	deepSeek *DeepSeekService
}

// NewChartService 构造。
func NewChartService(db *gorm.DB, deepSeek *DeepSeekService) *ChartService {
	return &ChartService{db: db, deepSeek: deepSeek}
}

// DB 暴露底层 *gorm.DB，供同步/异步生成等场景直接操作（Phase 4/5 使用）。
func (s *ChartService) DB() *gorm.DB { return s.db }

// Add 新增图表，返回新 id。
func (s *ChartService) Add(userID int64, req *dto.ChartAddRequest) (int64, error) {
	chart := &model.Chart{
		UserID:    userID,
		Name:      req.Name,
		Goal:      req.Goal,
		ChartData: req.ChartData,
		ChartType: req.ChartType,
	}
	if err := s.db.Create(chart).Error; err != nil {
		return 0, common.NewBusinessError(common.OPERATION_ERROR)
	}
	return chart.ID, nil
}

// GetByID 按 id 查询图表。
func (s *ChartService) GetByID(id int64) (*model.Chart, error) {
	var chart model.Chart
	err := s.db.Where("id = ?", id).First(&chart).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, common.NewBusinessError(common.NOT_FOUND_ERROR)
	}
	if err != nil {
		return nil, common.NewBusinessError(common.SYSTEM_ERROR)
	}
	return &chart, nil
}

// Delete 逻辑删除图表：需为本人或管理员。
func (s *ChartService) Delete(id int64, loginUser *model.User) (bool, error) {
	old, err := s.GetByID(id)
	if err != nil {
		return false, err
	}
	if old.UserID != loginUser.ID && !IsAdmin(loginUser) {
		return false, common.NewBusinessError(common.NO_AUTH_ERROR)
	}
	if err := s.db.Delete(&model.Chart{}, id).Error; err != nil {
		return false, common.NewBusinessError(common.OPERATION_ERROR)
	}
	return true, nil
}

// Edit 编辑图表：仅本人。
func (s *ChartService) Edit(loginUser *model.User, req *dto.ChartEditRequest) (bool, error) {
	if req.ID <= 0 {
		return false, common.NewBusinessError(common.PARAMS_ERROR)
	}
	old, err := s.GetByID(req.ID)
	if err != nil {
		return false, err
	}
	if old.UserID != loginUser.ID && !IsAdmin(loginUser) {
		return false, common.NewBusinessError(common.NO_AUTH_ERROR)
	}
	updates := map[string]any{
		"name":      req.Name,
		"goal":      req.Goal,
		"chartData": req.ChartData,
		"chartType": req.ChartType,
	}
	if err := s.db.Model(&model.Chart{}).Where("id = ?", req.ID).Updates(updates).Error; err != nil {
		return false, common.NewBusinessError(common.OPERATION_ERROR)
	}
	return true, nil
}

// Update 更新图表（管理员）。
func (s *ChartService) Update(req *dto.ChartUpdateRequest) (bool, error) {
	if req.ID <= 0 {
		return false, common.NewBusinessError(common.PARAMS_ERROR)
	}
	if _, err := s.GetByID(req.ID); err != nil {
		return false, err
	}
	updates := map[string]any{
		"name":      req.Name,
		"goal":      req.Goal,
		"chartData": req.ChartData,
		"chartType": req.ChartType,
		"genChart":  req.GenChart,
		"genResult": req.GenResult,
	}
	if err := s.db.Model(&model.Chart{}).Where("id = ?", req.ID).Updates(updates).Error; err != nil {
		return false, common.NewBusinessError(common.OPERATION_ERROR)
	}
	return true, nil
}

// Page 分页查询图表。size 上限 20（对齐 Java 限制爬虫）。
func (s *ChartService) Page(req *dto.ChartQueryRequest) (*common.Page[model.Chart], error) {
	current, size := normalizePage(req.Current, req.PageSize)
	if size > 20 {
		return nil, common.NewBusinessError(common.PARAMS_ERROR)
	}
	q := s.db.Model(&model.Chart{})
	if req.ID > 0 {
		q = q.Where("id = ?", req.ID)
	}
	if req.Name != "" {
		q = q.Where("name like ?", "%"+req.Name+"%")
	}
	if req.Goal != "" {
		q = q.Where("goal like ?", "%"+req.Goal+"%")
	}
	if req.ChartType != "" {
		q = q.Where("chartType like ?", "%"+req.ChartType+"%")
	}
	if req.UserID > 0 {
		q = q.Where("userId = ?", req.UserID)
	}
	if order := buildOrder(req.SortField, req.SortOrder); order != "" {
		q = q.Order(order)
	} else {
		q = q.Order("createTime DESC")
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, common.NewBusinessError(common.SYSTEM_ERROR)
	}
	var records []model.Chart
	if err := q.Offset(int((current - 1) * size)).Limit(int(size)).Find(&records).Error; err != nil {
		return nil, common.NewBusinessError(common.SYSTEM_ERROR)
	}
	return common.NewPage(records, total, current, size), nil
}

// BuildPrompt 拼接给 AI 的完整指令（同步/异步共用，对齐 Java）。
func BuildPrompt(goal, chartType, csvData string) string {
	return fmt.Sprintf("分析目标：%s\n图表类型：%s\n数据如下：\n%s", goal, chartType, csvData)
}

// GenSync 同步生成：调用 AI → 清洗 → 入库 → 返回 BiResponse。
// csvData 由 handler 从上传文件解析得到。
func (s *ChartService) GenSync(userID int64, req *dto.GenChartByAiRequest, csvData string) (*dto.BiResponse, error) {
	prompt := BuildPrompt(req.Goal, req.ChartType, csvData)

	aiResult, err := s.deepSeek.AnalyzeCsv(prompt)
	if err != nil {
		return nil, common.NewBusinessError(common.SYSTEM_ERROR, "AI 生成图表失败："+err.Error())
	}
	genChart := util.CleanEchartsJSON(aiResult.GenChart)

	chart := &model.Chart{
		UserID:    userID,
		Name:      req.Name,
		Goal:      req.Goal,
		ChartType: req.ChartType,
		ChartData: csvData,
		GenChart:  genChart,
		GenResult: aiResult.GenResult,
		Status:    model.ChartStatusSucceed,
	}
	if err := s.db.Create(chart).Error; err != nil {
		return nil, common.NewBusinessError(common.OPERATION_ERROR, "图表信息保存失败！")
	}

	return &dto.BiResponse{
		ChartID:   chart.ID,
		GenChart:  genChart,
		GenResult: aiResult.GenResult,
	}, nil
}

// SaveWait 异步生成第一步：先入库为 wait 状态，返回新 chartId（由 handler 发消息）。
func (s *ChartService) SaveWait(userID int64, req *dto.GenChartByAiRequest, csvData string) (int64, error) {
	chart := &model.Chart{
		UserID:    userID,
		Name:      req.Name,
		Goal:      req.Goal,
		ChartType: req.ChartType,
		ChartData: csvData,
		Status:    model.ChartStatusWait,
	}
	if err := s.db.Create(chart).Error; err != nil {
		return 0, common.NewBusinessError(common.OPERATION_ERROR, "图表信息保存失败！")
	}
	return chart.ID, nil
}

// ProcessChart 消费者处理逻辑：置 running → 调 AI → 清洗 → 置 succeed。
// 任一步失败返回 error（由消费者按重试/兜底策略处理）。
func (s *ChartService) ProcessChart(chartID int64) error {
	var chart model.Chart
	if err := s.db.Where("id = ?", chartID).First(&chart).Error; err != nil {
		return common.NewBusinessError(common.NOT_FOUND_ERROR, "图表不存在")
	}

	// 置为执行中
	if err := s.db.Model(&model.Chart{}).Where("id = ?", chartID).
		Update("status", model.ChartStatusRunning).Error; err != nil {
		return common.NewBusinessError(common.OPERATION_ERROR, "图表状态更新为running失败")
	}

	prompt := BuildPrompt(chart.Goal, chart.ChartType, chart.ChartData)
	aiResult, err := s.deepSeek.AnalyzeCsv(prompt)
	if err != nil {
		return common.NewBusinessError(common.OPERATION_ERROR, "AI 生成图表失败："+err.Error())
	}
	genChart := util.CleanEchartsJSON(aiResult.GenChart)

	// 置为成功
	if err := s.db.Model(&model.Chart{}).Where("id = ?", chartID).Updates(map[string]any{
		"status":    model.ChartStatusSucceed,
		"genChart":  genChart,
		"genResult": aiResult.GenResult,
	}).Error; err != nil {
		return common.NewBusinessError(common.OPERATION_ERROR, "图表状态更新为succeed失败")
	}
	return nil
}

// MarkFailed 将图表置为失败并写入执行信息（重试耗尽后的兜底，对标 @Recover）。
func (s *ChartService) MarkFailed(chartID int64, message string) {
	s.db.Model(&model.Chart{}).Where("id = ?", chartID).Updates(map[string]any{
		"status":      model.ChartStatusFailed,
		"execMessage": message,
	})
}
