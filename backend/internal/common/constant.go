package common

// 会话与角色常量（对齐 Java UserConstant / UserRoleEnum）。
const (
	// SessionUserIDKey 是会话中存储登录用户 id 的 key。
	SessionUserIDKey = "userId"
	// ContextLoginUserKey 是 gin.Context 中存储登录用户的 key（由 auth 中间件注入）。
	ContextLoginUserKey = "loginUser"

	// 用户角色
	RoleUser  = "user"
	RoleAdmin = "admin"
	RoleBan   = "ban"
)
