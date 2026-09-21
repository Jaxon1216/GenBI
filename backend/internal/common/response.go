// Package common 提供统一响应封装、错误码与分页结构（对齐 Java BaseResponse/ErrorCode）。
package common

// BaseResponse 是所有接口的统一返回封装。
// 对齐 Java：{ code, data, message }，code==0 表示成功。
type BaseResponse struct {
	Code    int    `json:"code"`
	Data    any    `json:"data"`
	Message string `json:"message"`
}

// Success 构造成功响应。
func Success(data any) BaseResponse {
	return BaseResponse{Code: 0, Data: data, Message: "ok"}
}

// Error 依据错误码构造失败响应（data 为 null）。
func Error(code ErrorCode) BaseResponse {
	return BaseResponse{Code: code.Code, Data: nil, Message: code.Message}
}

// ErrorWithMsg 依据错误码构造失败响应，但自定义 message。
func ErrorWithMsg(code ErrorCode, message string) BaseResponse {
	return BaseResponse{Code: code.Code, Data: nil, Message: message}
}
