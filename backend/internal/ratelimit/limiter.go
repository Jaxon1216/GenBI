// Package ratelimit 基于 Redis 令牌桶做限流（对照 Redisson RRateLimiter）。
package ratelimit

import (
	"context"
	"time"

	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
)

// Limiter 令牌桶限流器。
type Limiter struct {
	limiter *redis_rate.Limiter
}

// New 构造。
func New(rdb *redis.Client) *Limiter {
	return &Limiter{limiter: redis_rate.NewLimiter(rdb)}
}

// Allow 对 key 施加「每秒 3 次」限流（对齐 Java）。返回 true 表示放行。
func (l *Limiter) Allow(key string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	res, err := l.limiter.Allow(ctx, key, redis_rate.PerSecond(3))
	if err != nil {
		return false, err
	}
	return res.Allowed > 0, nil
}
