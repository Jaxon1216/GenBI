// Package mq 封装 RabbitMQ 连接、拓扑声明与生产/消费。
// 拓扑常量与 Java Constant 保持一致，便于对照迁移。
package mq

const (
	// QueueName 主业务队列名。
	//
	// 注意：这里从旧的 "BI_Queue" 改为 "BI_Queue.v2"。
	// 原因：RabbitMQ 不允许修改一个【已存在】队列的声明参数（例如给它补上
	// x-dead-letter-exchange 死信参数），对已存在队列重复 QueueDeclare 不同参数会报
	// PRECONDITION_FAILED。为了给主队列加死信能力又不手动删除线上旧队列，这里直接换一个
	// 新队列名重新声明（带死信参数）。
	//
	// 端到端连通性保证：exchange 与 routingKey 保持不变，producer 仍发往
	// ExchangeName/RoutingKey；新队列用相同 RoutingKey 绑定到同一 exchange，
	// consumer 消费新队列，链路不变。
	// 已部署环境影响：旧 BI_Queue 中若有积压消息不会被新 consumer 消费，需人工处理旧队列。
	QueueName    = "BI_Queue.v2"   // 主业务队列（带死信参数，故换新名）
	ExchangeName = "BI_exchange"   // direct 交换机（不变）
	RoutingKey   = "BI_routingkey" // 路由键（不变）

	// 死信（Dead Letter）拓扑：
	// 主队列中重试耗尽、被 Nack(requeue=false) 的毒消息，会经由主队列声明的
	// x-dead-letter-exchange 投递到下面的死信 exchange，再路由进死信队列保留下来，
	// 便于人工排查而非直接丢弃。
	DeadLetterExchangeName = "BI_exchange.DLX"   // 死信 direct 交换机（durable）
	DeadLetterQueueName    = "BI_Queue.DLX"      // 死信队列（durable）
	DeadLetterRoutingKey   = "BI_routingkey.DLX" // 死信路由键
)
