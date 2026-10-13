package mq

import (
	"context"
	"errors"
	"log"
	"strconv"
	"time"

	"genbi-go-backend/internal/event"
	"genbi-go-backend/internal/model"
	"genbi-go-backend/internal/service"

	amqp "github.com/rabbitmq/amqp091-go"
)

// ChartProcessor 抽象图表处理逻辑（由 service.ChartService 实现）。
type ChartProcessor interface {
	ProcessChart(ctx context.Context, chartID int64) error
	MarkFailed(chartID int64, message string)
}

// Consumer 消费 BI_Queue，异步生成图表。
type Consumer struct {
	ch         *amqp.Channel
	processor  ChartProcessor
	genTimeout time.Duration        // 单次生成超时上限（来自 config.ChartGenTimeout）
	publisher  event.EventPublisher // 状态变更事件发布器（默认 NoopPublisher）
}

// NewConsumer 构造。
// genTimeout 为单次图表生成的超时上限；publisher 为状态变更事件发布器
// （传 event.NoopPublisher{} 即不发布，供未启用 SSE 时使用）。
func NewConsumer(client *Client, processor ChartProcessor, genTimeout time.Duration, publisher event.EventPublisher) *Consumer {
	if publisher == nil {
		publisher = event.NoopPublisher{}
	}
	return &Consumer{ch: client.Channel, processor: processor, genTimeout: genTimeout, publisher: publisher}
}

// Start 开始消费（阻塞）。手动 ack，应用层重试最多 2 次、退避 2s，耗尽走 onExhausted。
func (c *Consumer) Start() error {
	// 预取 1，避免单消费者堆积（对标 Java Qos）。
	if err := c.ch.Qos(1, 0, false); err != nil {
		return err
	}
	deliveries, err := c.ch.Consume(QueueName, "", false /* autoAck=false，手动 ack */, false, false, false, nil)
	if err != nil {
		return err
	}
	log.Printf("[consumer] 开始消费队列 %s", QueueName)
	for d := range deliveries {
		c.handle(d)
	}
	return nil
}

// handle 处理单条消息。
func (c *Consumer) handle(d amqp.Delivery) {
	message := string(d.Body)
	log.Printf("[consumer] 接收到图表id消息：%s", message)

	chartID, err := strconv.ParseInt(message, 10, 64)
	if err != nil {
		log.Printf("[consumer] 消息格式错误，无法解析 chartId：%s", message)
		_ = d.Ack(false) // 坏消息直接确认移出
		return
	}

	// 用配置的超时包一层 ctx，透传给 ProcessChart → AI 调用，使生成可被主动打断。
	ctx, cancel := context.WithTimeout(context.Background(), c.genTimeout)
	defer cancel()

	// 应用层重试：最多 2 次，间隔 2s（对标 @Retryable(maxAttempts=2, backoff=2000)）。
	const maxAttempts = 2
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		lastErr = c.processor.ProcessChart(ctx, chartID)
		if lastErr == nil {
			break
		}
		// 幂等命中：该图表已被其它消费者/重复消息处理过（CAS 未抢到）。
		// 直接 ack 跳过，不重试、不 MarkFailed、不发状态事件，避免重复生成与误发事件。
		if errors.Is(lastErr, service.ErrAlreadyProcessed) {
			log.Printf("[consumer] 图表id:%d 已被处理（幂等跳过），确认消息", chartID)
			if err := d.Ack(false); err != nil {
				log.Printf("[consumer] 确认消息失败：%v", err)
			}
			return
		}
		log.Printf("[consumer] 图表id:%d 第 %d 次处理失败：%v", chartID, attempt, lastErr)
		if attempt < maxAttempts {
			time.Sleep(2 * time.Second)
		}
	}

	// 重试耗尽：走 onExhausted（对标 @Recover）。
	if lastErr != nil {
		c.onExhausted(d, chartID, lastErr)
		return
	}

	// 成功：发布 succeed 事件（默认 no-op；wt-sse 注入 Redis 实现后推送 SSE）。
	c.publisher.PublishChartEvent(context.Background(), chartID, model.ChartStatusSucceed)
	if err := d.Ack(false); err != nil {
		log.Printf("[consumer] 确认消息失败：%v", err)
	}
}

// onExhausted 重试耗尽后的兜底处理点。
// 实现：先 MarkFailed + 发布 failed 事件（保留原有可观测性），
// 再 Nack(requeue=false) 让这条毒消息经主队列的 x-dead-letter-exchange 进入死信队列，
// 可留存排查而非直接丢弃。
func (c *Consumer) onExhausted(d amqp.Delivery, chartID int64, lastErr error) {
	log.Printf("[consumer] 图表id:%d 处理失败，已达最大重试次数", chartID)
	c.processor.MarkFailed(chartID, "AI 生成图表失败："+lastErr.Error())
	c.publisher.PublishChartEvent(context.Background(), chartID, model.ChartStatusFailed)

	// Nack 且 requeue=false：不重回主队列，交由死信 exchange 投递进死信队列保留。
	if err := d.Nack(false, false); err != nil {
		log.Printf("[consumer] Nack 消息失败：%v", err)
	}
}
