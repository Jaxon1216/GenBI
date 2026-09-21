package dto

import "time"

// LoginUserVO 已登录用户视图（脱敏，不含密码）。对齐前端 typings。
type LoginUserVO struct {
	ID          int64     `json:"id"`
	UserName    string    `json:"userName"`
	UserAvatar  string    `json:"userAvatar"`
	UserProfile string    `json:"userProfile"`
	UserRole    string    `json:"userRole"`
	CreateTime  time.Time `json:"createTime"`
	UpdateTime  time.Time `json:"updateTime"`
}

// UserVO 用户视图（脱敏）。
type UserVO struct {
	ID         int64     `json:"id"`
	UserName   string    `json:"userName"`
	UserAvatar string    `json:"userAvatar"`
	UserRole   string    `json:"userRole"`
	CreateTime time.Time `json:"createTime"`
	UpdateTime time.Time `json:"updateTime"`
}
