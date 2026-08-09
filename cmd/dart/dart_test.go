// 江湖·接暗器 - 测试：游戏逻辑 / 结算 / 排行榜
package main

import (
	"path/filepath"
	"testing"

	"gocligames/engine"
)

func testGame(t *testing.T) *Game {
	t.Helper()
	dir := t.TempDir()
	gm := &Game{mode: "title", sb: engine.NewScoreboard(filepath.Join(dir, "scores.json"), 10)}
	return gm
}

// 新局初始状态
func TestReset(t *testing.T) {
	gm := testGame(t)
	gm.reset()
	if gm.lives != 3 || gm.score != 0 || gm.mode != "play" {
		t.Fatalf("reset bad: %+v", gm)
	}
}

// 接金镖加分
func TestCatchGold(t *testing.T) {
	gm := testGame(t)
	gm.reset()
	gm.px = 10
	gm.spawnT = 999
	gm.darts = []*Dart{{X: 10, Y: float64(scrH - 1), Speed: 1, Kind: 0, Ch: '镖', Active: true}}
	before := gm.score
	gm.update(nil, 0.05)
	if gm.score != before+10 {
		t.Fatalf("should +10, score=%d", gm.score)
	}
	if len(gm.darts) != 0 {
		t.Fatalf("dart should be consumed, got %d", len(gm.darts))
	}
}

// 毒镖扣命，3 次死亡
func TestPoisonDeath(t *testing.T) {
	gm := testGame(t)
	gm.reset()
	for i := 0; i < 3; i++ {
		gm.px = 10
		gm.spawnT = 999
		gm.darts = []*Dart{{X: 10, Y: float64(scrH - 1), Speed: 1, Kind: 1, Ch: '毒', Active: true}}
		gm.update(nil, 0.05)
	}
	if gm.mode != "over" {
		t.Fatalf("should be over, mode=%s", gm.mode)
	}
}

// 结算写排行榜，重复结算防重
func TestSettle(t *testing.T) {
	gm := testGame(t)
	gm.reset()
	gm.score = 100
	gm.settle()
	if gm.rankIn != 0 {
		t.Fatalf("rankIn=%d", gm.rankIn)
	}
	if len(gm.sb.Top(10)) != 1 {
		t.Fatal("should have 1 record")
	}
	gm.score = 200
	gm.settle() // settled 防重
	if gm.rankIn != 0 {
		t.Fatal("settled should not re-add")
	}
	if len(gm.sb.Top(10)) != 1 {
		t.Fatal("should still 1 record")
	}
}

// 漏接金镖不扣命
func TestMissGold(t *testing.T) {
	gm := testGame(t)
	gm.reset()
	gm.px = 5
	gm.spawnT = 999
	gm.darts = []*Dart{{X: 30, Y: float64(scrH - 1), Speed: 1, Kind: 0, Ch: '镖', Active: true}}
	gm.update(nil, 0.05)
	if gm.lives != 3 {
		t.Fatalf("miss should not cost life, lives=%d", gm.lives)
	}
	if gm.score != 0 {
		t.Fatalf("miss should not score, score=%d", gm.score)
	}
}
