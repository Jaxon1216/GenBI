// Package handler 提供 gin HTTP 处理函数。
package handler

import (
	"net/http"

	"genbi-go-backend/internal/common"

	"github.com/gin-gonic/gin"
)

// Health 存活探针：GET /api/health。
func Health(c *gin.Context) {
	c.JSON(http.StatusOK, common.Success("ok"))
}
