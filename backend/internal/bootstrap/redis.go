package bootstrap

import (
	"context"
	"log"

	"genbi-go-backend/internal/config"

	"github.com/redis/go-redis/v9"
)

// NewRedis 连接 Redis 并返回客户端。用于会话存储与限流。
func NewRedis(cfg *config.Config) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       0,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("[bootstrap] 连接 Redis 失败: %v", err)
	}
	log.Println("[bootstrap] Redis 连接成功")
	return rdb
}
