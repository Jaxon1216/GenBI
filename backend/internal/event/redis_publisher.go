package event

import (
	"context"
	"encoding/json"
	"log"

	"github.com/redis/go-redis/v9"
)

// chartEventPayload 是发布到 Redis 频道的消息体。
// 跨进程契约：consumer 发布、server 的 SSE handler 订阅并解析，字段名与前端口径一致。
type chartEventPayload struct {
	ChartID int64  `json:"chartId"`
	Status  string `json:"status"`
}

// RedisPublisher 基于 Redis Pub/Sub 的事件发布实现（供 consumer 进程使用）。
// 持有 Redis 客户端与频道名，向该频道 Publish 图表状态变更事件。
type RedisPublisher struct {
	rdb     *redis.Client
	channel string
}

// NewRedisPublisher 构造 Redis 事件发布器。
// channel 为 Pub/Sub 频道名（来自 config.ChartEventsChannel，默认 chart:events）。
func NewRedisPublisher(rdb *redis.Client, channel string) *RedisPublisher {
	return &RedisPublisher{rdb: rdb, channel: channel}
}

// PublishChartEvent 发布图表状态变更事件。
// 「尽力而为」语义：序列化或 Publish 失败只记录日志，不 panic、不阻断消费主流程。
func (p *RedisPublisher) PublishChartEvent(ctx context.Context, chartID int64, status string) {
	payload, err := json.Marshal(chartEventPayload{ChartID: chartID, Status: status})
	if err != nil {
		log.Printf("[event] 序列化图表事件失败 chartId=%d status=%s: %v", chartID, status, err)
		return
	}
	if err := p.rdb.Publish(ctx, p.channel, payload).Err(); err != nil {
		log.Printf("[event] 发布图表事件失败 chartId=%d status=%s: %v", chartID, status, err)
		return
	}
	log.Printf("[event] 已发布图表事件 chartId=%d status=%s", chartID, status)
}
