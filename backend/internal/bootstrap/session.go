package bootstrap

import (
	"genbi-go-backend/internal/config"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
)

// NewSessionStore 基于 Redis 创建会话存储（对照 Spring Session + Redis）。
func NewSessionStore(cfg *config.Config) (sessions.Store, error) {
	store, err := redis.NewStore(10, "tcp", cfg.RedisAddr, "", cfg.RedisPassword, []byte(cfg.SessionSecret))
	if err != nil {
		return nil, err
	}
	// Cookie 属性：HttpOnly、SameSite=Lax、30 天过期。
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   cfg.SessionMaxAge,
		HttpOnly: true,
		SameSite: 2, // http.SameSiteLaxMode
	})
	return store, nil
}

// SessionName 是会话 cookie 的名称。
const SessionName = "GENBI_SESSION"

// sessionMiddleware 返回挂载会话的中间件。
func sessionMiddleware(store sessions.Store) gin.HandlerFunc {
	return sessions.Sessions(SessionName, store)
}
