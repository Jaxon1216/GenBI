package middleware

import (
	"net/http"

	"genbi-go-backend/internal/common"
	"genbi-go-backend/internal/model"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Auth 登录鉴权中间件（对照 Java getLoginUser）：
// 从会话取 userId → 查库 → 注入 *model.User 到 gin.Context；缺失/无效则 40100。
func Auth(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		v := session.Get(common.SessionUserIDKey)
		userID, ok := v.(int64)
		if !ok || userID <= 0 {
			c.AbortWithStatusJSON(http.StatusOK, common.Error(common.NOT_LOGIN_ERROR))
			return
		}
		var user model.User
		if err := db.Where("id = ?", userID).First(&user).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusOK, common.Error(common.NOT_LOGIN_ERROR))
			return
		}
		c.Set(common.ContextLoginUserKey, &user)
		c.Next()
	}
}

// Admin 管理员校验中间件（对照 @AuthCheck(mustRole="admin")）。需先经过 Auth。
func Admin() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := GetLoginUser(c)
		if user == nil || user.UserRole != common.RoleAdmin {
			c.AbortWithStatusJSON(http.StatusOK, common.Error(common.NO_AUTH_ERROR))
			return
		}
		c.Next()
	}
}

// GetLoginUser 从 gin.Context 取出 Auth 中间件注入的登录用户。
func GetLoginUser(c *gin.Context) *model.User {
	v, ok := c.Get(common.ContextLoginUserKey)
	if !ok {
		return nil
	}
	user, ok := v.(*model.User)
	if !ok {
		return nil
	}
	return user
}
