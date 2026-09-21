// Package dto 定义请求/响应结构体，字段对齐前端 typings.d.ts。
package dto

// UserRegisterRequest 注册请求。
type UserRegisterRequest struct {
	UserAccount   string `json:"userAccount"`
	UserPassword  string `json:"userPassword"`
	CheckPassword string `json:"checkPassword"`
}

// UserLoginRequest 登录请求。
type UserLoginRequest struct {
	UserAccount  string `json:"userAccount"`
	UserPassword string `json:"userPassword"`
}

// UserAddRequest 管理员创建用户请求。
type UserAddRequest struct {
	UserAccount string `json:"userAccount"`
	UserName    string `json:"userName"`
	UserAvatar  string `json:"userAvatar"`
	UserRole    string `json:"userRole"`
}

// UserUpdateRequest 管理员更新用户请求。
type UserUpdateRequest struct {
	ID         int64  `json:"id"`
	UserName   string `json:"userName"`
	UserAvatar string `json:"userAvatar"`
	UserRole   string `json:"userRole"`
}

// UserUpdateMyRequest 更新个人信息请求。
type UserUpdateMyRequest struct {
	UserName   string `json:"userName"`
	UserAvatar string `json:"userAvatar"`
}

// UserQueryRequest 用户分页查询请求。
type UserQueryRequest struct {
	Current   int64  `json:"current"`
	PageSize  int64  `json:"pageSize"`
	ID        int64  `json:"id"`
	UserName  string `json:"userName"`
	UserRole  string `json:"userRole"`
	SortField string `json:"sortField"`
	SortOrder string `json:"sortOrder"`
}

// DeleteRequest 通用删除请求。
type DeleteRequest struct {
	ID int64 `json:"id"`
}
