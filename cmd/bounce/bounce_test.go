package main

import (
	"math"
	"path/filepath"
	"testing"

	"gocligames/engine"
)

func newTestGame(t *testing.T) (*engine.Game, *Bounce) {
	t.Helper()
	g := engine.NewGame("test", w, h, fps)
	b := &Bounce{}
	b.reset()
	b.scores = engine.NewScoreboard(filepath.Join(t.TempDir(), "s.json"), 10)
	g.Paused = false
	return g, b
}

func TestPaddleMoveClamp(t *testing.T) {
	_, b := newTestGame(t)
	for i := 0; i < 30; i++ {
		b.onKey(nil, "a")
	}
	if b.paddle.X != 1 {
		t.Fatalf("左边界应 clamp 到 1，got %v", b.paddle.X)
	}
	for i := 0; i < 60; i++ {
		b.onKey(nil, "d")
	}
	if b.paddle.X != float64(w-1-paddleW) {
		t.Fatalf("右边界 clamp 错误，got %v", b.paddle.X)
	}
}

func TestServeAndMotion(t *testing.T) {
	g, b := newTestGame(t)
	b.onKey(g, "space")
	if !b.ballActive {
		t.Fatal("空格应发球")
	}
	x0, y0 := b.ball.X, b.ball.Y
	for i := 0; i < 10; i++ {
		b.update(g, 1.0/30)
	}
	if b.ball.X == x0 && b.ball.Y == y0 {
		t.Fatal("球应移动")
	}
}

func TestPaddleBounceAndScore(t *testing.T) {
	g, b := newTestGame(t)
	b.ball.X = b.paddle.CenterX() - 0.5
	b.ball.Y = float64(paddleY) - 0.15
	b.dirX, b.dirY = 0, 1
	b.ballActive = true
	before := b.score
	b.update(g, 1.0/30)
	if b.ball.Y > float64(paddleY)-1 {
		t.Fatalf("球应弹回，y=%v", b.ball.Y)
	}
	if b.dirY >= 0 {
		t.Fatalf("反弹后应向上，dirY=%v", b.dirY)
	}
	if b.score <= before {
		t.Fatal("接球应加分")
	}
}

func TestEdgeAngle(t *testing.T) {
	g, b := newTestGame(t)
	b.ball.X = b.paddle.X - 0.5 // 挡板最左边缘
	b.ball.Y = float64(paddleY) - 0.15
	b.dirX, b.dirY = 0, 1
	b.ballActive = true
	b.update(g, 1.0/30)
	angle := math.Acos(math.Abs(b.dirX)) * 180 / math.Pi
	if angle < 40 {
		t.Fatalf("边缘接球应大角度，got %.1f", angle)
	}
}

func TestMissLosesLife(t *testing.T) {
	g, b := newTestGame(t)
	b.onKey(g, "space")
	b.ball.Y = float64(h) - 1.5
	b.dirX, b.dirY = 0, 1
	b.update(g, 1.0/30)
	if b.lives != 2 {
		t.Fatalf("漏球应扣命，lives=%d", b.lives)
	}
	if b.ballActive {
		t.Fatal("漏球后应等待重新发球")
	}
}

func TestGameOverAutoSettle(t *testing.T) {
	g, b := newTestGame(t)
	b.lives = 1
	b.score = 30 // 有分数才写榜
	b.onKey(g, "space")
	b.ball.Y = float64(h) - 1.5
	b.dirX, b.dirY = 0, 1
	b.update(g, 1.0/30)
	if !b.over || !b.settled {
		t.Fatal("生命耗尽应 over 并自动结算")
	}
	if len(b.scores.Scores) == 0 {
		t.Fatal("结算后排行榜应有记录")
	}
}

func TestSettleDedup(t *testing.T) {
	_, b := newTestGame(t)
	b.score = 50
	b.settle()
	n := len(b.scores.Scores)
	b.settle()
	if len(b.scores.Scores) != n {
		t.Fatal("重复结算不应重复写榜")
	}
}

func TestRenderSmoke(t *testing.T) {
	g, b := newTestGame(t)
	b.onKey(g, "space")
	for i := 0; i < 20; i++ {
		b.update(g, 1.0/30)
	}
	g.Screen.Clear()
	b.render(g, g.Screen)
	if g.Screen.Height() != h {
		t.Fatalf("画布高度 %d != %d", g.Screen.Height(), h)
	}
	if g.Screen.Width() != w {
		t.Fatalf("画布宽度 %d != %d", g.Screen.Width(), w)
	}
	frame := g.Screen.Frame()
	if len(frame) == 0 {
		t.Fatal("渲染帧为空")
	}
}
