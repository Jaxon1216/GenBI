package mq

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Client 持有 RabbitMQ 连接与信道，并负责声明拓扑（exchange/queue/binding）。
type Client struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
}

// NewClient 连接 RabbitMQ 并声明拓扑（对照 Java InitMain）。
func NewClient(url string) (*Client, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	if err := declareTopology(ch); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, err
	}
	log.Println("[mq] RabbitMQ 连接成功，拓扑已声明")
	return &Client{Conn: conn, Channel: ch}, nil
}

// declareTopology 声明 direct exchange、持久化 queue 并绑定（与 Java InitMain 一致）。
func declareTopology(ch *amqp.Channel) error {
	if err := ch.ExchangeDeclare(ExchangeName, "direct", true, false, false, false, nil); err != nil {
		return err
	}
	if _, err := ch.QueueDeclare(QueueName, true, false, false, false, nil); err != nil {
		return err
	}
	return ch.QueueBind(QueueName, RoutingKey, ExchangeName, false, nil)
}

// Close 关闭信道与连接。
func (c *Client) Close() {
	if c.Channel != nil {
		_ = c.Channel.Close()
	}
	if c.Conn != nil {
		_ = c.Conn.Close()
	}
}
