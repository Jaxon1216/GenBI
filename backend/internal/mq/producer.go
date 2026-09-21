package mq

import (
	"context"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Producer 向 BI_exchange 发布图表生成任务。
type Producer struct {
	ch *amqp.Channel
}

// NewProducer 构造。
func NewProducer(client *Client) *Producer {
	return &Producer{ch: client.Channel}
}

// Publish 发布一条消息（消息体为 chartId 的字符串），持久化投递。
func (p *Producer) Publish(message string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return p.ch.PublishWithContext(ctx,
		ExchangeName, // exchange
		RoutingKey,   // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType:  "text/plain",
			DeliveryMode: amqp.Persistent, // 消息持久化
			Body:         []byte(message),
		},
	)
}
