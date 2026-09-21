// Package middleware 提供 gin 中间件：recover、请求日志、登录鉴权、管理员校验。
package middleware

import (
	"log"
	"net/http"
	"runtime/debug"

	"genbi-go-backend/internal/common"

	"github.com/gin-gonic/gin"
)

// Recovery 捕获 panic，返回统一的系统错误响应（对照 Java GlobalExceptionHandler）。
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("[panic] %v\n%s", err, debug.Stack())
				c.AbortWithStatusJSON(http.StatusOK, common.Error(common.SYSTEM_ERROR))
			}
		}()
		c.Next()
	}
}
