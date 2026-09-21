package mq

import (
	"log"
	"strconv"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// ChartProcessor 抽象图表处理逻辑（由 service.ChartService 实现）。
type ChartProcessor interface {
	ProcessChart(chartID int64) error
	MarkFailed(chartID int64, message string)
}

// Consumer 消费 BI_Queue，异步生成图表。
type Consumer struct {
	ch        *amqp.Channel
	processor ChartProcessor
}

// NewConsumer 构造。
func NewConsumer(client *Client, processor ChartProcessor) *Consumer {
	return &Consumer{ch: client.Channel, processor: processor}
}

// Start 开始消费（阻塞）。手动 ack，应用层重试最多 2 次、退避 2s，耗尽置 failed。
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

	// 应用层重试：最多 2 次，间隔 2s（对标 @Retryable(maxAttempts=2, backoff=2000)）。
	const maxAttempts = 2
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if lastErr = c.processor.ProcessChart(chartID); lastErr == nil {
			break
		}
		log.Printf("[consumer] 图表id:%d 第 %d 次处理失败：%v", chartID, attempt, lastErr)
		if attempt < maxAttempts {
			time.Sleep(2 * time.Second)
		}
	}
	// 重试耗尽兜底（对标 @Recover）。
	if lastErr != nil {
		log.Printf("[consumer] 图表id:%d 处理失败，已达最大重试次数", chartID)
		c.processor.MarkFailed(chartID, "AI 生成图表失败："+lastErr.Error())
	}

	// 无论成功还是兜底，都确认消息移出队列（对齐 Java 手动 ack 逻辑）。
	if err := d.Ack(false); err != nil {
		log.Printf("[consumer] 确认消息失败：%v", err)
	}
}
