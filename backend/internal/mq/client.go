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

// declareTopology 声明主拓扑 + 死信拓扑（在 Java InitMain 基础上增加死信能力）。
//
// 主队列 BI_Queue.v2 带 x-dead-letter-exchange 参数：当消息被 Nack(requeue=false)
// 或过期时，RabbitMQ 会把它投递到死信 exchange，再路由进死信队列保留，便于排查。
func declareTopology(ch *amqp.Channel) error {
	// 1. 主 direct exchange（durable，不变）。
	if err := ch.ExchangeDeclare(ExchangeName, "direct", true, false, false, false, nil); err != nil {
		return err
	}

	// 2. 死信 exchange（direct，durable）。
	if err := ch.ExchangeDeclare(DeadLetterExchangeName, "direct", true, false, false, false, nil); err != nil {
		return err
	}

	// 3. 死信队列（durable）并绑定到死信 exchange。
	if _, err := ch.QueueDeclare(DeadLetterQueueName, true, false, false, false, nil); err != nil {
		return err
	}
	if err := ch.QueueBind(DeadLetterQueueName, DeadLetterRoutingKey, DeadLetterExchangeName, false, nil); err != nil {
		return err
	}

	// 4. 主队列（durable）声明死信参数，指向死信 exchange / routingKey。
	//    注意：队列参数一旦声明不可变，故主队列已换新名 BI_Queue.v2（见 constant.go 说明）。
	mainArgs := amqp.Table{
		"x-dead-letter-exchange":    DeadLetterExchangeName,
		"x-dead-letter-routing-key": DeadLetterRoutingKey,
	}
	if _, err := ch.QueueDeclare(QueueName, true, false, false, false, mainArgs); err != nil {
		return err
	}

	// 5. 主队列绑定到主 exchange（routingKey 不变，端到端链路不变）。
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
