// Package outbox 实现「事务消息表（Transactional Outbox）」的后台投递 relay。
//
// 背景与定位：
//
//	GenAsync 原先是「先入库 chart(wait)，再向 RabbitMQ 发消息」的双写，一旦发消息失败，
//	任务就会永远停在 wait。改造后，chart(wait) 与 outbox(pending) 在同一个 DB 事务里写入，
//	业务入库与「待投递消息」要么同时成功、要么同时回滚，不再丢消息；真正的 MQ 投递被解耦
//	到本 relay：server 进程内一个后台 goroutine 按固定间隔轮询 outbox 中 status=pending 的
//	记录，逐条调用 producer.Publish(payload) 补发，成功置为 sent，失败则累加 retryCount 等下轮重试。
//
// ===== exactly-once 边界契约（务必阅读）=====
//
//	本 relay 的投递语义是 at-least-once（至少一次），而非 exactly-once：
//	  1. 补发成功后、UPDATE status='sent' 之前进程崩溃，下轮会再次扫描到该 pending 记录并重发；
//	  2. 轮询本身也可能在上一条消息已投递、本条标记未落库时与消费并发交叠。
//	因此「同一个 chartId 可能被投递多次」是本设计刻意接受的权衡（宁可重复，不可丢失）。
//	消息「有且仅处理一次」的最终保证，依赖消费端（wt-reliable）的幂等消费兜底：
//	消费者用状态 CAS 推进——UPDATE chart SET status='running' WHERE id=? AND status='wait'，
//	当 RowsAffected==0 说明该任务已被（其他投递/重复消息）处理过，直接跳过。
//	换言之：relay 负责「不丢」，幂等消费负责「不重」，两者配合才构成端到端的 exactly-once 效果。
//	relay 侧不做去重，也不应依赖去重；retryCount 仅用于观测与排障，不作为丢弃阈值（避免误丢）。
package outbox

import (
	"context"
	"log"
	"time"

	"genbi-go-backend/internal/model"
	"genbi-go-backend/internal/mq"

	"gorm.io/gorm"
)

// defaultBatchSize 单轮最多拉取的 pending 记录数，避免一次拉太多撑爆内存 / 长时间占用。
const defaultBatchSize = 100

// MessagePublisher 抽象「把消息投递出去」的能力，*mq.Producer 实现之。
// 抽象成接口是为了让 relay 的扫描/置位逻辑可用 fake 做单测，而不必连真实 RabbitMQ。
type MessagePublisher interface {
	Publish(message string) error
}

// pendingStore 抽象 relay 对 outbox 表的读写，便于单测注入内存 fake（无需 sqlite 依赖）。
type pendingStore interface {
	// fetchPending 取出至多 limit 条 status=pending 的记录。
	fetchPending(ctx context.Context, limit int) ([]model.Outbox, error)
	// markSent 将指定记录置为 sent。
	markSent(ctx context.Context, id int64) error
	// incrementRetry 将指定记录的 retryCount 加一（status 保持 pending，下轮重试）。
	incrementRetry(ctx context.Context, id int64) error
}

// gormStore 是基于 *gorm.DB 的 pendingStore 实现。
type gormStore struct {
	db *gorm.DB
}

func (s *gormStore) fetchPending(ctx context.Context, limit int) ([]model.Outbox, error) {
	var records []model.Outbox
	err := s.db.WithContext(ctx).
		Where("status = ?", model.OutboxStatusPending).
		Order("id ASC").
		Limit(limit).
		Find(&records).Error
	return records, err
}

func (s *gormStore) markSent(ctx context.Context, id int64) error {
	return s.db.WithContext(ctx).Model(&model.Outbox{}).
		Where("id = ?", id).
		Update("status", model.OutboxStatusSent).Error
}

func (s *gormStore) incrementRetry(ctx context.Context, id int64) error {
	return s.db.WithContext(ctx).Model(&model.Outbox{}).
		Where("id = ?", id).
		Update("retryCount", gorm.Expr("retryCount + ?", 1)).Error
}

// Relay 轮询 outbox 并把 pending 记录补发到 MQ。
type Relay struct {
	store     pendingStore
	publisher MessagePublisher
	batchSize int
}

// NewRelay 构造 relay。producer 为真实的 RabbitMQ 生产者（server 进程在其非 nil 时才启动 relay）。
func NewRelay(db *gorm.DB, producer *mq.Producer) *Relay {
	return &Relay{
		store:     &gormStore{db: db},
		publisher: producer,
		batchSize: defaultBatchSize,
	}
}

// Start 启动轮询（阻塞运行，通常由调用方 go 出去）。
// 按 interval 周期扫描一次 pending；收到 ctx.Done() 后优雅退出。
func (r *Relay) Start(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = 5 * time.Second
	}
	log.Printf("[outbox] relay 启动，轮询间隔 %s，批量上限 %d", interval, r.batchSize)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// 启动后立即跑一次，避免白等一个完整 interval 才补发历史 pending。
	r.runOnce(ctx)
	for {
		select {
		case <-ctx.Done():
			log.Printf("[outbox] relay 退出：%v", ctx.Err())
			return
		case <-ticker.C:
			r.runOnce(ctx)
		}
	}
}

// runOnce 执行一轮：取 pending → 逐条 Publish → 成功置 sent / 失败累加 retryCount。
// 返回本轮成功投递的条数，便于观测与单测断言。单条失败不影响同批其余记录。
func (r *Relay) runOnce(ctx context.Context) int {
	records, err := r.store.fetchPending(ctx, r.batchSize)
	if err != nil {
		log.Printf("[outbox] 扫描 pending 记录失败：%v", err)
		return 0
	}

	sent := 0
	for i := range records {
		rec := records[i]
		if err := r.publisher.Publish(rec.Payload); err != nil {
			// 投递失败：仅累加 retryCount，status 仍为 pending，等下一轮重试（at-least-once）。
			log.Printf("[outbox] 投递失败 outboxId=%d chartId=%d（第 %d 次重试将在下轮进行）：%v",
				rec.ID, rec.AggregateID, rec.RetryCount+1, err)
			if e := r.store.incrementRetry(ctx, rec.ID); e != nil {
				log.Printf("[outbox] 累加 retryCount 失败 outboxId=%d：%v", rec.ID, e)
			}
			continue
		}
		if err := r.store.markSent(ctx, rec.ID); err != nil {
			// 已投递但置 sent 失败：下轮会重发（重复由幂等消费兜底，见文件头契约）。
			log.Printf("[outbox] 置 sent 失败 outboxId=%d（下轮将重发，由幂等消费兜底）：%v", rec.ID, err)
			continue
		}
		sent++
	}
	if sent > 0 {
		log.Printf("[outbox] 本轮补发完成，成功投递 %d 条", sent)
	}
	return sent
}
