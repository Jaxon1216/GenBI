package model

import "time"

// outbox 投递状态常量：pending 待投递 / sent 已投递。
const (
	OutboxStatusPending = "pending"
	OutboxStatusSent    = "sent"
)

// Outbox 事务消息表（表 outbox）。
//
// 它与 chart 的写入在同一个 DB 事务里完成，保证「业务入库 + 待投递消息落库」的原子性，
// 彻底消除原先 GenAsync「入库成功但发 MQ 失败 → 任务永远 wait」的双写不一致问题；
// 真正的 MQ 投递交由 server 进程内的 relay goroutine 轮询 pending 记录补发。
//
// 列名采用「前端口径」驼峰风格（aggregateId / retryCount / createTime / updateTime），
// 与 chart 表保持一致。
type Outbox struct {
	ID          int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	AggregateID int64     `gorm:"column:aggregateId" json:"aggregateId"` // 聚合根 id，此处为 chartId
	Payload     string    `gorm:"column:payload" json:"payload"`         // 待投递的消息体，此处为 chartId 字符串
	Status      string    `gorm:"column:status" json:"status"`           // pending / sent
	RetryCount  int       `gorm:"column:retryCount" json:"retryCount"`   // 投递重试次数
	CreateTime  time.Time `gorm:"column:createTime;autoCreateTime" json:"createTime"`
	UpdateTime  time.Time `gorm:"column:updateTime;autoUpdateTime" json:"updateTime"`
}

// TableName 指定表名。
func (Outbox) TableName() string { return "outbox" }
