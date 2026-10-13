package service

import (
	"errors"
	"testing"
)

// TestClassifyCASResult 覆盖「置 running」CAS 的三类结果分类（纯逻辑，无 DB 依赖）。
//
// 为什么用纯函数测而非 sqlite/sqlmock：本仓库 go.mod 未引入 sqlite / sqlmock 驱动，
// mysql 驱动即使 DryRun 也会在 gorm.Open 时真实建连；go.mod 不在本次改动边界内，
// 故把 CAS 的「结果 → 返回错误」判定抽成纯函数 classifyCASResult 做最小可验证测试。
func TestClassifyCASResult(t *testing.T) {
	// 场景1：wait→running 第一次抢占成功（影响 1 行）→ 返回 nil，继续生成。
	if err := classifyCASResult(nil, 1); err != nil {
		t.Errorf("抢占成功应返回 nil，实际: %v", err)
	}

	// 场景2：第二次处理（状态已是 running/succeed，不再是 wait）→ 影响 0 行
	//        → 识别为 ErrAlreadyProcessed，供 consumer 幂等跳过。
	err := classifyCASResult(nil, 0)
	if !errors.Is(err, ErrAlreadyProcessed) {
		t.Errorf("RowsAffected==0 应返回 ErrAlreadyProcessed，实际: %v", err)
	}

	// 场景3：DB 执行出错 → 返回更新失败业务错误，且不能被误判为「已处理」。
	dbErr := errors.New("db down")
	got := classifyCASResult(dbErr, 0)
	if got == nil {
		t.Fatal("DB 出错应返回非 nil 错误")
	}
	if errors.Is(got, ErrAlreadyProcessed) {
		t.Errorf("DB 出错不应被判为 ErrAlreadyProcessed，实际: %v", got)
	}
}

// TestErrAlreadyProcessedIsStable 哨兵错误应可被 errors.Is 稳定识别
// （consumer 依赖 errors.Is(lastErr, service.ErrAlreadyProcessed) 做幂等跳过）。
func TestErrAlreadyProcessedIsStable(t *testing.T) {
	wrapped := errors.New("wrap: " + ErrAlreadyProcessed.Error())
	// 普通包装（非 %w）不应误命中。
	if errors.Is(wrapped, ErrAlreadyProcessed) {
		t.Error("非 %w 包装不应命中哨兵")
	}
	// 直接返回的哨兵必须命中。
	if !errors.Is(ErrAlreadyProcessed, ErrAlreadyProcessed) {
		t.Error("哨兵自身应命中 errors.Is")
	}
}
