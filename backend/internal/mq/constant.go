// Package mq 封装 RabbitMQ 连接、拓扑声明与生产/消费。
// 拓扑常量与 Java Constant 保持一致，便于对照迁移。
package mq

const (
	QueueName    = "BI_Queue"      // 队列
	ExchangeName = "BI_exchange"   // direct 交换机
	RoutingKey   = "BI_routingkey" // 路由键
)
