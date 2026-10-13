package handler

import (
	"context"
	"encoding/json"
	"time"

	"genbi-go-backend/internal/common"
	"genbi-go-backend/internal/middleware"
	"genbi-go-backend/internal/model"
	"genbi-go-backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// heartbeatInterval SSE 心跳间隔：定期写注释行保持连接存活（穿越代理/负载均衡的空闲超时）。
const heartbeatInterval = 15 * time.Second

// sseEventPayload 解析 consumer 经 Redis 发布的事件体（字段与 event.RedisPublisher 对齐）。
type sseEventPayload struct {
	ChartID int64  `json:"chartId"`
	Status  string `json:"status"`
}

// SSEHandler 图表状态变更的 SSE 推送处理器（server 进程使用）。
// 订阅 Redis Pub/Sub 频道，按 chartId 过滤后把事件以 text/event-stream 推给浏览器。
// 作为独立 handler 存在，避免给 ChartHandler 引入 Redis 依赖。
type SSEHandler struct {
	rdb          *redis.Client
	chartService *service.ChartService
	channel      string
}

// NewSSEHandler 构造。channel 为 Pub/Sub 频道名（来自 config.ChartEventsChannel）。
func NewSSEHandler(rdb *redis.Client, chartService *service.ChartService, channel string) *SSEHandler {
	return &SSEHandler{rdb: rdb, chartService: chartService, channel: channel}
}

// Stream GET /api/chart/sse?id= —— 建立 SSE 长连接推送指定图表的状态变更。
// 需经 auth 中间件：仅图表归属者可订阅；收到 succeed/failed 后结束流。
func (h *SSEHandler) Stream(c *gin.Context) {
	id := parseIDQuery(c)
	if id <= 0 {
		failCode(c, common.PARAMS_ERROR)
		return
	}

	// 鉴权：校验该图表归属当前登录用户，避免越权订阅他人任务。
	user := middleware.GetLoginUser(c)
	chart, err := h.chartService.GetByID(id)
	if err != nil {
		fail(c, err)
		return
	}
	if user == nil || chart.UserID != user.ID {
		failCode(c, common.NO_AUTH_ERROR)
		return
	}

	// SSE 响应头。X-Accel-Buffering 关闭 Nginx 等代理的缓冲，确保事件实时下发。
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	c.Writer.Flush()

	// 订阅 Redis 频道。独立 ctx 跟随请求生命周期，断连时取消以释放订阅。
	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()
	pubsub := h.rdb.Subscribe(ctx, h.channel)
	defer func() { _ = pubsub.Close() }()
	msgCh := pubsub.Channel()

	// 心跳定时器：空闲时写注释行维持连接。
	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.Request.Context().Done():
			// 客户端断连：退出循环，defer 清理 pubsub。
			return
		case <-ticker.C:
			if _, err := c.Writer.WriteString(": ping\n\n"); err != nil {
				return
			}
			c.Writer.Flush()
		case msg, ok := <-msgCh:
			if !ok {
				// 订阅通道关闭（如 Redis 断开）：结束流，前端据此回退轮询。
				return
			}
			var payload sseEventPayload
			if err := json.Unmarshal([]byte(msg.Payload), &payload); err != nil {
				// 坏消息跳过，不中断连接。
				continue
			}
			// 按 chartId 过滤：只推与本次订阅 id 匹配的事件。
			if payload.ChartID != id {
				continue
			}
			if _, err := c.Writer.WriteString("data: " + msg.Payload + "\n\n"); err != nil {
				return
			}
			c.Writer.Flush()
			// 终态（成功/失败）后主动结束流，前端收到后关闭 EventSource。
			if payload.Status == model.ChartStatusSucceed || payload.Status == model.ChartStatusFailed {
				return
			}
		}
	}
}
