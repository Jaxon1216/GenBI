package service

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestAnalyzeCsv_ContextTimeout 验证：当传入一个很短超时的 ctx 时，
// AnalyzeCsv 能被 ctx「真正打断」——
//
//	(a) 在远小于 client 60s 兜底的时间内返回（证明是 ctx 打断而非等客户端超时）；
//	(b) 返回的 error 满足 errors.Is(err, context.DeadlineExceeded)。
func TestAnalyzeCsv_ContextTimeout(t *testing.T) {
	// 慢 server：sleep 2s 远超 ctx 的 100ms，但远小于 client 的 60s 兜底。
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"{}"}}]}`))
	}))
	defer server.Close()

	// apiKey 非空才会真正发起 HTTP 请求。
	svc := NewDeepSeekService("test-key", server.URL)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, err := svc.AnalyzeCsv(ctx, "col1,col2\n1,2")
	elapsed := time.Since(start)

	// (b) error 必须可识别为超时。
	if err == nil {
		t.Fatalf("期望超时返回 error，但得到 nil")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("期望 errors.Is(err, context.DeadlineExceeded) 为真，实际 err=%v", err)
	}

	// (a) 必须在远小于 60s 的时间内返回；给足余量设为 5s。
	if elapsed >= 5*time.Second {
		t.Fatalf("期望被 ctx 在 ~100ms 打断（远小于 client 60s 兜底），实际耗时 %v", elapsed)
	}
	t.Logf("AnalyzeCsv 被 ctx 超时打断，耗时 %v，err=%v", elapsed, err)
}

// TestAnalyzeCsv_ContextCanceled 验证：主动 cancel 时同样被打断且可识别为 Canceled。
func TestAnalyzeCsv_ContextCanceled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"{}"}}]}`))
	}))
	defer server.Close()

	svc := NewDeepSeekService("test-key", server.URL)

	ctx, cancel := context.WithCancel(context.Background())
	// 50ms 后主动取消。
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	_, err := svc.AnalyzeCsv(ctx, "col1,col2\n1,2")
	elapsed := time.Since(start)

	if err == nil {
		t.Fatalf("期望取消返回 error，但得到 nil")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("期望 errors.Is(err, context.Canceled) 为真，实际 err=%v", err)
	}
	if elapsed >= 5*time.Second {
		t.Fatalf("期望被 cancel 快速打断，实际耗时 %v", elapsed)
	}
	t.Logf("AnalyzeCsv 被 ctx 取消打断，耗时 %v，err=%v", elapsed, err)
}

// TestAnalyzeCsv_NilCtx 验证 nil ctx 兜底不 panic（走 context.Background）。
func TestAnalyzeCsv_NilCtx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"{\"genChart\":\"{}\",\"genResult\":[\"ok\"]}"}}]}`))
	}))
	defer server.Close()

	svc := NewDeepSeekService("test-key", server.URL)

	//nolint:staticcheck // 故意传 nil 测试兜底分支
	res, err := svc.AnalyzeCsv(nil, "col1,col2\n1,2")
	if err != nil {
		t.Fatalf("nil ctx 兜底应正常返回，实际 err=%v", err)
	}
	if res == nil || res.GenResult == "" {
		t.Fatalf("期望解析出结果，实际 res=%+v", res)
	}
}
