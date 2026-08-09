// 江湖·闭关 - 测试：修炼/突破/排行榜
package main

import (
	"path/filepath"
	"testing"

	"gocligames/engine"
)

func testGame(t *testing.T) *Game {
	t.Helper()
	dir := t.TempDir()
	gm := &Game{mode: "menu", sb: engine.NewScoreboard(filepath.Join(dir, "scores.json"), 10)}
	return gm
}

// 闭关结束获得修为
func TestClosedGain(t *testing.T) {
	gm := testGame(t)
	gm.startClosed(2)
	gm.closedT = 0.01
	gm.update(nil, 0.02)
	if gm.mode != "result" {
		t.Fatalf("should be result, mode=%s", gm.mode)
	}
	if gm.xiuwei <= 0 {
		t.Fatalf("should gain xiuwei, got %d", gm.xiuwei)
	}
}

// 修为累积突破境界
func TestRealmBreakthrough(t *testing.T) {
	gm := testGame(t)
	gm.xiuwei = 305
	gm.closedN = 1
	gm.finishClosed()
	if gm.realm < 1 {
		t.Fatalf("should break to 筑基, realm=%d", gm.realm)
	}
}

// 结算写榜 + 防重
func TestSettle(t *testing.T) {
	gm := testGame(t)
	gm.xiuwei = 100
	gm.settle()
	if len(gm.sb.Top(10)) != 1 {
		t.Fatal("should have 1 record")
	}
	gm.xiuwei = 500
	gm.settle()
	if len(gm.sb.Top(10)) != 1 {
		t.Fatal("settled should prevent re-add")
	}
}

// 中断闭关回菜单
func TestInterrupt(t *testing.T) {
	gm := testGame(t)
	gm.startClosed(10)
	gm.mode = "closed"
	gm.onKey(nil, "q")
	if gm.mode != "menu" {
		t.Fatalf("q should return menu, mode=%s", gm.mode)
	}
}
