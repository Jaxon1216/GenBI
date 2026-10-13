package outbox

import (
	"context"
	"errors"
	"testing"

	"genbi-go-backend/internal/model"
)

// fakePublisher 以内存方式记录投递调用，并可按需对指定 payload 返回错误。
// 用它替代真实 *mq.Producer，使 relay 的扫描/置位逻辑无需连 RabbitMQ 即可测。
type fakePublisher struct {
	published []string        // 成功 + 失败都会记录尝试过的 payload，顺序即调用顺序
	failOn    map[string]bool // 命中则该 payload 的 Publish 返回错误
}

func (p *fakePublisher) Publish(message string) error {
	p.published = append(p.published, message)
	if p.failOn[message] {
		return errors.New("模拟投递失败")
	}
	return nil
}

// fakeStore 以内存 map 模拟 outbox 表，实现 pendingStore，便于断言 status/retryCount 的变化。
type fakeStore struct {
	records map[int64]*model.Outbox
}

func newFakeStore(recs ...*model.Outbox) *fakeStore {
	m := make(map[int64]*model.Outbox, len(recs))
	for _, r := range recs {
		m[r.ID] = r
	}
	return &fakeStore{records: m}
}

func (s *fakeStore) fetchPending(_ context.Context, limit int) ([]model.Outbox, error) {
	var out []model.Outbox
	// 按 id 升序稳定返回，模拟 "Order(id ASC)"。
	for id := int64(1); id <= int64(len(s.records))+int64(limit); id++ {
		r, ok := s.records[id]
		if !ok {
			continue
		}
		if r.Status == model.OutboxStatusPending {
			out = append(out, *r)
			if len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

func (s *fakeStore) markSent(_ context.Context, id int64) error {
	if r, ok := s.records[id]; ok {
		r.Status = model.OutboxStatusSent
	}
	return nil
}

func (s *fakeStore) incrementRetry(_ context.Context, id int64) error {
	if r, ok := s.records[id]; ok {
		r.RetryCount++
	}
	return nil
}

// newTestRelay 用内存 fake 组装一个 Relay。
func newTestRelay(store pendingStore, pub MessagePublisher) *Relay {
	return &Relay{store: store, publisher: pub, batchSize: defaultBatchSize}
}

// 场景 1：扫描 pending → 调用 Publish → 全部置 sent。
func TestRunOnce_PublishesPendingAndMarksSent(t *testing.T) {
	store := newFakeStore(
		&model.Outbox{ID: 1, AggregateID: 101, Payload: "101", Status: model.OutboxStatusPending},
		&model.Outbox{ID: 2, AggregateID: 102, Payload: "102", Status: model.OutboxStatusPending},
	)
	pub := &fakePublisher{failOn: map[string]bool{}}
	relay := newTestRelay(store, pub)

	sent := relay.runOnce(context.Background())

	if sent != 2 {
		t.Fatalf("期望成功投递 2 条，实际 %d", sent)
	}
	if len(pub.published) != 2 {
		t.Fatalf("期望 Publish 被调用 2 次，实际 %d：%v", len(pub.published), pub.published)
	}
	for id, r := range store.records {
		if r.Status != model.OutboxStatusSent {
			t.Errorf("outboxId=%d 期望 status=sent，实际 %q", id, r.Status)
		}
		if r.RetryCount != 0 {
			t.Errorf("outboxId=%d 成功投递不应累加 retryCount，实际 %d", id, r.RetryCount)
		}
	}
}

// 场景 2：Publish 失败 → retryCount 递增、status 仍为 pending（下轮重试）。
func TestRunOnce_PublishFailsIncrementsRetryAndKeepsPending(t *testing.T) {
	store := newFakeStore(
		&model.Outbox{ID: 1, AggregateID: 101, Payload: "101", Status: model.OutboxStatusPending},
		&model.Outbox{ID: 2, AggregateID: 102, Payload: "102", Status: model.OutboxStatusPending},
	)
	// payload "101" 投递失败，"102" 成功。
	pub := &fakePublisher{failOn: map[string]bool{"101": true}}
	relay := newTestRelay(store, pub)

	sent := relay.runOnce(context.Background())

	if sent != 1 {
		t.Fatalf("期望成功投递 1 条（仅 102），实际 %d", sent)
	}

	failed := store.records[1]
	if failed.Status != model.OutboxStatusPending {
		t.Errorf("投递失败的记录应保持 pending，实际 %q", failed.Status)
	}
	if failed.RetryCount != 1 {
		t.Errorf("投递失败应使 retryCount=1，实际 %d", failed.RetryCount)
	}

	ok := store.records[2]
	if ok.Status != model.OutboxStatusSent {
		t.Errorf("投递成功的记录应为 sent，实际 %q", ok.Status)
	}

	// 再跑一轮：这次让 101 也成功，验证它会被重新扫描并最终置 sent（下轮重试语义）。
	pub.failOn = map[string]bool{}
	sent2 := relay.runOnce(context.Background())
	if sent2 != 1 {
		t.Fatalf("第二轮期望补发剩余 1 条，实际 %d", sent2)
	}
	if failed.Status != model.OutboxStatusSent {
		t.Errorf("第二轮后 outboxId=1 应为 sent，实际 %q", failed.Status)
	}
	if failed.RetryCount != 1 {
		t.Errorf("第二轮成功投递不应再加 retryCount，期望仍为 1，实际 %d", failed.RetryCount)
	}
}

// 场景 3：没有 pending 记录时，不调用 Publish，返回 0。
func TestRunOnce_NoPendingNoop(t *testing.T) {
	store := newFakeStore(
		&model.Outbox{ID: 1, AggregateID: 101, Payload: "101", Status: model.OutboxStatusSent},
	)
	pub := &fakePublisher{failOn: map[string]bool{}}
	relay := newTestRelay(store, pub)

	if sent := relay.runOnce(context.Background()); sent != 0 {
		t.Fatalf("无 pending 时应返回 0，实际 %d", sent)
	}
	if len(pub.published) != 0 {
		t.Fatalf("无 pending 时不应调用 Publish，实际 %v", pub.published)
	}
}
