package handler

import (
	"strconv"

	"genbi-go-backend/internal/common"
	"genbi-go-backend/internal/dto"
	"genbi-go-backend/internal/middleware"
	"genbi-go-backend/internal/mq"
	"genbi-go-backend/internal/ratelimit"
	"genbi-go-backend/internal/service"
	"genbi-go-backend/internal/util"

	"github.com/gin-gonic/gin"
)

// ChartHandler 图表接口处理器。
type ChartHandler struct {
	chartService *service.ChartService
	limiter      *ratelimit.Limiter
	producer     *mq.Producer
}

// NewChartHandler 构造。producer 可为 nil（未启用异步时 /gen/async 会返回系统错误）。
func NewChartHandler(chartService *service.ChartService, limiter *ratelimit.Limiter, producer *mq.Producer) *ChartHandler {
	return &ChartHandler{chartService: chartService, limiter: limiter, producer: producer}
}

// Add POST /api/chart/add
func (h *ChartHandler) Add(c *gin.Context) {
	var req dto.ChartAddRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		failCode(c, common.PARAMS_ERROR)
		return
	}
	user := middleware.GetLoginUser(c)
	id, err := h.chartService.Add(user.ID, &req)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, id)
}

// Delete POST /api/chart/delete
func (h *ChartHandler) Delete(c *gin.Context) {
	var req dto.DeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.ID <= 0 {
		failCode(c, common.PARAMS_ERROR)
		return
	}
	user := middleware.GetLoginUser(c)
	b, err := h.chartService.Delete(req.ID, user)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, b)
}

// GetByID GET /api/chart/get?id=
func (h *ChartHandler) GetByID(c *gin.Context) {
	id := parseIDQuery(c)
	if id <= 0 {
		failCode(c, common.PARAMS_ERROR)
		return
	}
	chart, err := h.chartService.GetByID(id)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, chart)
}

// Page POST /api/chart/list/page
func (h *ChartHandler) Page(c *gin.Context) {
	var req dto.ChartQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		failCode(c, common.PARAMS_ERROR)
		return
	}
	page, err := h.chartService.Page(&req)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, page)
}

// MyPage POST /api/chart/my/list/page —— 强制限定为当前登录用户。
func (h *ChartHandler) MyPage(c *gin.Context) {
	var req dto.ChartQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		failCode(c, common.PARAMS_ERROR)
		return
	}
	user := middleware.GetLoginUser(c)
	req.UserID = user.ID
	page, err := h.chartService.Page(&req)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, page)
}

// Edit POST /api/chart/edit （本人）
func (h *ChartHandler) Edit(c *gin.Context) {
	var req dto.ChartEditRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		failCode(c, common.PARAMS_ERROR)
		return
	}
	user := middleware.GetLoginUser(c)
	b, err := h.chartService.Edit(user, &req)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, b)
}

// Update POST /api/chart/update （管理员）
func (h *ChartHandler) Update(c *gin.Context) {
	var req dto.ChartUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		failCode(c, common.PARAMS_ERROR)
		return
	}
	b, err := h.chartService.Update(&req)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, b)
}

// Gen POST /api/chart/gen —— multipart 同步 AI 生成。
func (h *ChartHandler) Gen(c *gin.Context) {
	user := middleware.GetLoginUser(c)

	req, csvData, ok2 := h.parseGenRequest(c)
	if !ok2 {
		return
	}

	// 限流：gen_chart:{userId}
	allowed, err := h.limiter.Allow("gen_chart:" + strconv.FormatInt(user.ID, 10))
	if err != nil {
		failCode(c, common.SYSTEM_ERROR)
		return
	}
	if !allowed {
		failCode(c, common.TOO_MANY_REQUEST_ERROR)
		return
	}

	res, err := h.chartService.GenSync(user.ID, req, csvData)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, res)
}

// GenAsync POST /api/chart/gen/async —— multipart 异步 AI 生成。
// 入库为 wait 后发消息到 RabbitMQ，返回 chartId；由消费者异步生成。
func (h *ChartHandler) GenAsync(c *gin.Context) {
	user := middleware.GetLoginUser(c)

	req, csvData, valid := h.parseGenRequest(c)
	if !valid {
		return
	}

	if h.producer == nil {
		failCode(c, common.SYSTEM_ERROR, "异步服务未启用")
		return
	}

	// 限流：gen_chart:{userId}
	allowed, err := h.limiter.Allow("gen_chart:" + strconv.FormatInt(user.ID, 10))
	if err != nil {
		failCode(c, common.SYSTEM_ERROR)
		return
	}
	if !allowed {
		failCode(c, common.TOO_MANY_REQUEST_ERROR)
		return
	}

	chartID, err := h.chartService.SaveWait(user.ID, req, csvData)
	if err != nil {
		fail(c, err)
		return
	}
	if err := h.producer.Publish(strconv.FormatInt(chartID, 10)); err != nil {
		failCode(c, common.SYSTEM_ERROR, "任务提交失败")
		return
	}
	ok(c, chartID)
}

// parseGenRequest 解析 multipart 表单：校验文件格式、必填参数，返回 (req, csv, true)。
// 校验失败时已写入响应并返回 false。
func (h *ChartHandler) parseGenRequest(c *gin.Context) (*dto.GenChartByAiRequest, string, bool) {
	var req dto.GenChartByAiRequest
	// name/goal/chartType 可来自 form 或 query（前端以 query 传参）。
	_ = c.ShouldBind(&req)
	if req.Name == "" {
		req.Name = c.Query("name")
	}
	if req.Goal == "" {
		req.Goal = c.Query("goal")
	}
	if req.ChartType == "" {
		req.ChartType = c.Query("chartType")
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		failCode(c, common.PARAMS_ERROR, "上传文件不能为空")
		return nil, "", false
	}
	// 参数校验（对齐 Java）
	if req.Name == "" {
		failCode(c, common.PARAMS_ERROR, "图表名称不能为空")
		return nil, "", false
	}
	if req.Goal == "" {
		failCode(c, common.PARAMS_ERROR, "分析目标不能为空")
		return nil, "", false
	}
	if req.ChartType == "" {
		failCode(c, common.PARAMS_ERROR, "图表类型不能为空")
		return nil, "", false
	}

	csvData, err := util.ExcelToCSV(fileHeader)
	if err != nil {
		failCode(c, common.PARAMS_ERROR, "上传失败，仅支持 .xlsx 和 .csv 格式文件，请重新上传！")
		return nil, "", false
	}
	return &req, csvData, true
}
