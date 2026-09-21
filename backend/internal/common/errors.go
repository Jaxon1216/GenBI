package common

import "fmt"

// BusinessError 业务异常（对齐 Java BusinessException）。
// service 层抛出，handler 层捕获并转成 BaseResponse。
type BusinessError struct {
	Code    int
	Message string
}

func (e *BusinessError) Error() string {
	return fmt.Sprintf("business error %d: %s", e.Code, e.Message)
}

// NewBusinessError 用错误码构造业务异常，可覆盖 message。
func NewBusinessError(code ErrorCode, message ...string) *BusinessError {
	msg := code.Message
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	return &BusinessError{Code: code.Code, Message: msg}
}

// Page 分页结构，字段对齐前端 typings（前端读 records 与 total）。
type Page[T any] struct {
	Records []T   `json:"records"`
	Total   int64 `json:"total"`
	Current int64 `json:"current"`
	Size    int64 `json:"size"`
}

// NewPage 构造分页结果。
func NewPage[T any](records []T, total, current, size int64) *Page[T] {
	if records == nil {
		records = []T{}
	}
	return &Page[T]{Records: records, Total: total, Current: current, Size: size}
}
