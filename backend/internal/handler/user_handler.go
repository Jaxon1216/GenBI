package handler

import (
	"genbi-go-backend/internal/common"
	"genbi-go-backend/internal/dto"
	"genbi-go-backend/internal/middleware"
	"genbi-go-backend/internal/service"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// UserHandler 用户接口处理器。
type UserHandler struct {
	userService *service.UserService
}

// NewUserHandler 构造。
func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// Register POST /api/user/register
func (h *UserHandler) Register(c *gin.Context) {
	var req dto.UserRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		failCode(c, common.PARAMS_ERROR)
		return
	}
	id, err := h.userService.Register(req.UserAccount, req.UserPassword, req.CheckPassword)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, id)
}

// Login POST /api/user/login
func (h *UserHandler) Login(c *gin.Context) {
	var req dto.UserLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		failCode(c, common.PARAMS_ERROR)
		return
	}
	user, err := h.userService.Login(req.UserAccount, req.UserPassword)
	if err != nil {
		fail(c, err)
		return
	}
	// 写入会话
	session := sessions.Default(c)
	session.Set(common.SessionUserIDKey, user.ID)
	if err := session.Save(); err != nil {
		failCode(c, common.SYSTEM_ERROR)
		return
	}
	ok(c, service.ToLoginUserVO(user))
}

// Logout POST /api/user/logout
func (h *UserHandler) Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	if err := session.Save(); err != nil {
		failCode(c, common.SYSTEM_ERROR)
		return
	}
	ok(c, true)
}

// GetLoginUser GET /api/user/get/login
func (h *UserHandler) GetLoginUser(c *gin.Context) {
	user := middleware.GetLoginUser(c)
	ok(c, service.ToLoginUserVO(user))
}

// Add POST /api/user/add （管理员）
func (h *UserHandler) Add(c *gin.Context) {
	var req dto.UserAddRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		failCode(c, common.PARAMS_ERROR)
		return
	}
	id, err := h.userService.Add(&req)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, id)
}

// Delete POST /api/user/delete （管理员）
func (h *UserHandler) Delete(c *gin.Context) {
	var req dto.DeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.ID <= 0 {
		failCode(c, common.PARAMS_ERROR)
		return
	}
	b, err := h.userService.Delete(req.ID)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, b)
}

// Update POST /api/user/update （管理员）
func (h *UserHandler) Update(c *gin.Context) {
	var req dto.UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.ID <= 0 {
		failCode(c, common.PARAMS_ERROR)
		return
	}
	b, err := h.userService.Update(&req)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, b)
}

// UpdateMy POST /api/user/update/my
func (h *UserHandler) UpdateMy(c *gin.Context) {
	var req dto.UserUpdateMyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		failCode(c, common.PARAMS_ERROR)
		return
	}
	user := middleware.GetLoginUser(c)
	b, err := h.userService.UpdateMy(user.ID, &req)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, b)
}

// GetByID GET /api/user/get?id= （管理员）
func (h *UserHandler) GetByID(c *gin.Context) {
	id := parseIDQuery(c)
	if id <= 0 {
		failCode(c, common.PARAMS_ERROR)
		return
	}
	user, err := h.userService.GetByID(id)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, user)
}

// GetVOByID GET /api/user/get/vo?id=
func (h *UserHandler) GetVOByID(c *gin.Context) {
	id := parseIDQuery(c)
	if id <= 0 {
		failCode(c, common.PARAMS_ERROR)
		return
	}
	user, err := h.userService.GetByID(id)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, service.ToUserVO(user))
}

// Page POST /api/user/list/page （管理员）
func (h *UserHandler) Page(c *gin.Context) {
	var req dto.UserQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		failCode(c, common.PARAMS_ERROR)
		return
	}
	page, err := h.userService.Page(&req)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, page)
}

// PageVO POST /api/user/list/page/vo
func (h *UserHandler) PageVO(c *gin.Context) {
	var req dto.UserQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		failCode(c, common.PARAMS_ERROR)
		return
	}
	page, err := h.userService.Page(&req)
	if err != nil {
		fail(c, err)
		return
	}
	// 转 VO
	vos := make([]*dto.UserVO, 0, len(page.Records))
	for i := range page.Records {
		vos = append(vos, service.ToUserVO(&page.Records[i]))
	}
	ok(c, common.NewPage(vos, page.Total, page.Current, page.Size))
}
