// Package event 定义「图表任务状态变更事件」的发布抽象。
//
// 跨进程契约：consumer 进程在处理完成/失败后发布事件，server 进程订阅后经 SSE
// 推送给浏览器。生产实现（Redis Pub/Sub）由 wt-sse 在 redis_publisher.go 中提供，
// 此处仅定义接口与默认空实现，保证未启用 SSE 时链路照常工作。
package event

import "context"

// EventPublisher 发布图表状态变更事件。
// status 取值与前端口径一致：wait/running/succeed/failed（见 model 常量）。
// 约定为「尽力而为」语义：发布失败只应记录日志，不得阻断主流程或改变消费结果。
type EventPublisher interface {
	PublishChartEvent(ctx context.Context, chartID int64, status string)
}

// NoopPublisher 默认空实现：不做任何事。
// 未注入真实实现（未启用 SSE）时使用，使 consumer 的事件发布调用成为无副作用的空操作。
type NoopPublisher struct{}

// PublishChartEvent no-op。
func (NoopPublisher) PublishChartEvent(_ context.Context, _ int64, _ string) {}
