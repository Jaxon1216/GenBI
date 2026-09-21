package handler

import (
	"errors"
	"net/http"
	"strconv"

	"genbi-go-backend/internal/common"

	"github.com/gin-gonic/gin"
)

// fail 将 error 转成统一失败响应：BusinessError 用其 code/message，其余归为系统错误。
func fail(c *gin.Context, err error) {
	var be *common.BusinessError
	if errors.As(err, &be) {
		c.JSON(http.StatusOK, common.BaseResponse{Code: be.Code, Data: nil, Message: be.Message})
		return
	}
	c.JSON(http.StatusOK, common.Error(common.SYSTEM_ERROR))
}

// ok 输出成功响应。
func ok(c *gin.Context, data any) {
	c.JSON(http.StatusOK, common.Success(data))
}

// failCode 用指定错误码输出失败响应。
func failCode(c *gin.Context, code common.ErrorCode, msg ...string) {
	if len(msg) > 0 && msg[0] != "" {
		c.JSON(http.StatusOK, common.ErrorWithMsg(code, msg[0]))
		return
	}
	c.JSON(http.StatusOK, common.Error(code))
}

// parseIDQuery 从 query 参数 ?id= 解析出 int64（失败返回 0）。
func parseIDQuery(c *gin.Context) int64 {
	id, err := strconv.ParseInt(c.Query("id"), 10, 64)
	if err != nil {
		return 0
	}
	return id
}
