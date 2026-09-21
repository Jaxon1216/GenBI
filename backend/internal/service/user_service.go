// Package service 实现业务逻辑层（对照 Java 的 *ServiceImpl）。
package service

import (
	"errors"
	"strings"

	"genbi-go-backend/internal/common"
	"genbi-go-backend/internal/dto"
	"genbi-go-backend/internal/model"
	"genbi-go-backend/internal/util"

	"gorm.io/gorm"
)

// UserService 用户业务。
type UserService struct {
	db *gorm.DB
}

// NewUserService 构造。
func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

// Register 用户注册，返回新用户 id（对照 UserServiceImpl.userRegister）。
func (s *UserService) Register(account, password, checkPassword string) (int64, error) {
	// 1. 校验
	if account == "" || password == "" || checkPassword == "" {
		return 0, common.NewBusinessError(common.PARAMS_ERROR, "参数为空")
	}
	if len(account) < 4 {
		return 0, common.NewBusinessError(common.PARAMS_ERROR, "用户账号过短")
	}
	if len(password) < 8 || len(checkPassword) < 8 {
		return 0, common.NewBusinessError(common.PARAMS_ERROR, "用户密码过短")
	}
	if password != checkPassword {
		return 0, common.NewBusinessError(common.PARAMS_ERROR, "两次输入的密码不一致")
	}
	// 2. 账号不能重复
	var count int64
	if err := s.db.Model(&model.User{}).Where("userAccount = ?", account).Count(&count).Error; err != nil {
		return 0, common.NewBusinessError(common.SYSTEM_ERROR, "注册失败，数据库错误")
	}
	if count > 0 {
		return 0, common.NewBusinessError(common.PARAMS_ERROR, "账号重复")
	}
	// 3. 加密并插入
	user := &model.User{
		UserAccount:  account,
		UserPassword: util.EncryptPassword(password),
		UserName:     account,
		UserAvatar:   "https://javatan.oss-cn-beijing.aliyuncs.com/man01.png",
		UserRole:     common.RoleUser,
	}
	if err := s.db.Create(user).Error; err != nil {
		return 0, common.NewBusinessError(common.SYSTEM_ERROR, "注册失败，数据库错误")
	}
	return user.ID, nil
}

// Login 校验账号密码，成功返回用户实体（会话由 handler 写入）。
func (s *UserService) Login(account, password string) (*model.User, error) {
	if account == "" || password == "" {
		return nil, common.NewBusinessError(common.PARAMS_ERROR, "参数为空")
	}
	if len(account) < 4 {
		return nil, common.NewBusinessError(common.PARAMS_ERROR, "账号错误")
	}
	if len(password) < 8 {
		return nil, common.NewBusinessError(common.PARAMS_ERROR, "密码错误")
	}
	encrypted := util.EncryptPassword(password)
	var user model.User
	err := s.db.Where("userAccount = ? AND userPassword = ?", account, encrypted).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, common.NewBusinessError(common.PARAMS_ERROR, "用户不存在或密码错误")
	}
	if err != nil {
		return nil, common.NewBusinessError(common.SYSTEM_ERROR)
	}
	return &user, nil
}

// GetByID 按 id 查询用户。
func (s *UserService) GetByID(id int64) (*model.User, error) {
	var user model.User
	err := s.db.Where("id = ?", id).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, common.NewBusinessError(common.NOT_FOUND_ERROR)
	}
	if err != nil {
		return nil, common.NewBusinessError(common.SYSTEM_ERROR)
	}
	return &user, nil
}

// Add 管理员创建用户，默认密码 12345678。
func (s *UserService) Add(req *dto.UserAddRequest) (int64, error) {
	user := &model.User{
		UserAccount: req.UserAccount,
		UserName:    req.UserName,
		UserAvatar:  req.UserAvatar,
		UserRole:    req.UserRole,
	}
	if user.UserRole == "" {
		user.UserRole = common.RoleUser
	}
	user.UserPassword = util.EncryptPassword("12345678")
	if err := s.db.Create(user).Error; err != nil {
		return 0, common.NewBusinessError(common.OPERATION_ERROR)
	}
	return user.ID, nil
}

// Delete 按 id 逻辑删除用户。
func (s *UserService) Delete(id int64) (bool, error) {
	if err := s.db.Delete(&model.User{}, id).Error; err != nil {
		return false, common.NewBusinessError(common.OPERATION_ERROR)
	}
	return true, nil
}

// Update 更新用户（管理员）。
func (s *UserService) Update(req *dto.UserUpdateRequest) (bool, error) {
	updates := map[string]any{
		"userName":   req.UserName,
		"userAvatar": req.UserAvatar,
		"userRole":   req.UserRole,
	}
	if err := s.db.Model(&model.User{}).Where("id = ?", req.ID).Updates(updates).Error; err != nil {
		return false, common.NewBusinessError(common.OPERATION_ERROR)
	}
	return true, nil
}

// UpdateMy 更新个人信息。
func (s *UserService) UpdateMy(id int64, req *dto.UserUpdateMyRequest) (bool, error) {
	updates := map[string]any{
		"userName":   req.UserName,
		"userAvatar": req.UserAvatar,
	}
	if err := s.db.Model(&model.User{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return false, common.NewBusinessError(common.OPERATION_ERROR)
	}
	return true, nil
}

// Page 分页查询用户列表。
func (s *UserService) Page(req *dto.UserQueryRequest) (*common.Page[model.User], error) {
	current, size := normalizePage(req.Current, req.PageSize)
	q := s.db.Model(&model.User{})
	if req.ID > 0 {
		q = q.Where("id = ?", req.ID)
	}
	if req.UserRole != "" {
		q = q.Where("userRole = ?", req.UserRole)
	}
	if req.UserName != "" {
		q = q.Where("userName like ?", "%"+req.UserName+"%")
	}
	if order := buildOrder(req.SortField, req.SortOrder); order != "" {
		q = q.Order(order)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, common.NewBusinessError(common.SYSTEM_ERROR)
	}
	var records []model.User
	if err := q.Offset(int((current - 1) * size)).Limit(int(size)).Find(&records).Error; err != nil {
		return nil, common.NewBusinessError(common.SYSTEM_ERROR)
	}
	return common.NewPage(records, total, current, size), nil
}

// ToLoginUserVO 脱敏为登录视图。
func ToLoginUserVO(u *model.User) *dto.LoginUserVO {
	if u == nil {
		return nil
	}
	return &dto.LoginUserVO{
		ID:         u.ID,
		UserName:   u.UserName,
		UserAvatar: u.UserAvatar,
		UserRole:   u.UserRole,
		CreateTime: u.CreateTime,
		UpdateTime: u.UpdateTime,
	}
}

// ToUserVO 脱敏为用户视图。
func ToUserVO(u *model.User) *dto.UserVO {
	if u == nil {
		return nil
	}
	return &dto.UserVO{
		ID:         u.ID,
		UserName:   u.UserName,
		UserAvatar: u.UserAvatar,
		UserRole:   u.UserRole,
		CreateTime: u.CreateTime,
		UpdateTime: u.UpdateTime,
	}
}

// IsAdmin 判断是否管理员。
func IsAdmin(u *model.User) bool {
	return u != nil && u.UserRole == common.RoleAdmin
}

// normalizePage 归一化分页参数。
func normalizePage(current, size int64) (int64, int64) {
	if current <= 0 {
		current = 1
	}
	if size <= 0 {
		size = 10
	}
	return current, size
}

// buildOrder 构造安全的 order by 子句（字段白名单 + 方向）。
func buildOrder(field, order string) string {
	field = strings.TrimSpace(field)
	if field == "" {
		return ""
	}
	// 仅允许字母、数字、下划线，防注入。
	for _, r := range field {
		if !(r == '_' || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')) {
			return ""
		}
	}
	dir := "DESC"
	if strings.EqualFold(order, "ascend") || strings.EqualFold(order, "asc") {
		dir = "ASC"
	}
	return field + " " + dir
}
