// 弹球（Bounce）- gocligames 示例游戏。
// 演示引擎能力：Screen 双缓冲渲染 / Entity / Physics AABB /
// Scoreboard 排行榜 / 跨平台输入（WASD + 方向键 + 空格）。
package main

import (
	"fmt"
	"math"
	"os"
	"path/filepath"

	"gocligames/engine"
)

const (
	w, h    = 46, 22 // 画布尺寸（含边框）
	gTop    = 1      // 游戏区上界
	gBottom = h - 3  // 游戏区下界（不含提示行）
	paddleY = h - 4  // 挡板行
	paddleW = 6      // 挡板宽度
	fps     = 30
)

var scoresFile string

func main() {
	dir, _ := os.Getwd()
	scoresFile = filepath.Join(dir, "bounce_scores.json")

	g := engine.NewGame("弹球 (gocligames demo)", w, h, fps)
	b := &Bounce{}
	g.OnStart = b.onStart
	g.OnKey = b.onKey
	g.Update = b.update
	g.Render = b.render
	g.OnQuit = b.onQuit
	g.Run()
}

// Bounce 游戏状态。
type Bounce struct {
	paddle     *engine.Entity
	ball       *engine.Entity
	ballActive bool
	ballSpeed  float64
	dirX, dirY float64
	score      int
	lives      int
	hits       int
	settled    bool
	over       bool
	lastRank   int
	hint       string
	scores     *engine.Scoreboard
}

func (b *Bounce) reset() {
	b.paddle = engine.NewEntity(float64(w/2-paddleW/2), paddleY, paddleW, 1, '=', engine.ColorCyan)
	b.ball = engine.NewEntity(float64(w/2), float64(gTop+2), 1, 1, 'o', engine.ColorYellow)
	b.ballActive = false
	b.ballSpeed = 6.0
	b.dirX, b.dirY = 0.8, -0.6
	b.score, b.lives, b.hits = 0, 3, 0
	b.settled, b.over = false, false
	b.lastRank = -1
	b.hint = "空格发球 | AD/方向键移动 | P暂停 R重开 E结算 Q退出"
}

func (b *Bounce) onStart(g *engine.Game) {
	b.reset()
	b.scores = engine.NewScoreboard(scoresFile, 10)
}

func (b *Bounce) onKey(g *engine.Game, key string) {
	if b.settled || b.over {
		switch key {
		case "r":
			b.reset()
		case "q":
			g.Quit()
		}
		return
	}
	switch key {
	case "p":
		g.Paused = !g.Paused
	case "r":
		b.reset()
	case "q":
		b.quitSave()
		g.Quit()
	case "e":
		b.settle()
	case "a", "w", "left":
		b.paddle.X = engine.Clamp(b.paddle.X-1, 1, float64(w-1-paddleW))
	case "d", "s", "right":
		b.paddle.X = engine.Clamp(b.paddle.X+1, 1, float64(w-1-paddleW))
	case "space":
		if !b.ballActive && !b.settled {
			b.ballActive = true
			b.hint = "P暂停 R重开 E结算 Q退出"
		}
	}
}

func (b *Bounce) update(g *engine.Game, dt float64) {
	if g.Paused || b.settled || b.over || !b.ballActive {
		return
	}
	bl := b.ball
	bl.X += b.dirX * b.ballSpeed * dt
	bl.Y += b.dirY * b.ballSpeed * dt
	// 左右墙反弹
	if bl.X < 1 {
		bl.X = 1
		b.dirX = math.Abs(b.dirX)
	} else if bl.X > float64(w-2) {
		bl.X = float64(w - 2)
		b.dirX = -math.Abs(b.dirX)
	}
	// 上墙反弹
	if bl.Y < float64(gTop) {
		bl.Y = float64(gTop)
		b.dirY = math.Abs(b.dirY)
	}
	// 挡板碰撞：命中点越靠边缘，反弹角越大
	if bl.Y >= float64(paddleY) && b.dirY > 0 && engine.Overlap(bl, b.paddle) {
		bl.Y = float64(paddleY - 1)
		off := (bl.X + 0.5 - b.paddle.CenterX()) / (float64(paddleW) / 2)
		off = engine.Clamp(off, -1, 1)
		ang := off*math.Pi/4 + math.Pi/2
		b.dirX = math.Cos(ang)
		b.dirY = -math.Abs(math.Sin(ang))
		b.hits++
		b.score += 10 + int(math.Abs(off)*10)
		if b.hits%5 == 0 && b.ballSpeed < 12.0 {
			b.ballSpeed += 0.6
		}
	}
	// 漏球
	if bl.Y > float64(h-2) {
		b.lives--
		if b.lives <= 0 {
			b.over = true
			b.settle()
		} else {
			b.ball = engine.NewEntity(float64(w/2), float64(gTop+2), 1, 1, 'o', engine.ColorYellow)
			b.ballActive = false
			b.hint = "漏球！空格发球"
		}
	}
}

func (b *Bounce) settle() {
	if b.settled {
		return
	}
	b.settled = true
	if b.score > 0 {
		b.lastRank = b.scores.Add("PLAYER", b.score, fmt.Sprintf("%d hits", b.hits))
	}
}

func (b *Bounce) quitSave() {
	if !b.settled && !b.over && b.score > 0 {
		b.scores.Add("PLAYER", b.score, fmt.Sprintf("%d hits", b.hits))
	}
}

func (b *Bounce) render(g *engine.Game, s *engine.Screen) {
	s.Clear()
	engine.Box(s, 0, 0, w, h, engine.ColorGray)
	s.Text(2, 1, fmt.Sprintf("分数 %d", b.score), engine.ColorWhite)
	life := ""
	for i := 0; i < b.lives; i++ {
		life += "o"
	}
	s.Text(w-12, 1, fmt.Sprintf("生命 %s", life), engine.ColorGreen)
	s.Text(2, 2, fmt.Sprintf("速度 %.1f  连击 %d", b.ballSpeed, b.hits), engine.ColorGray)
	for i := 0; i < w-4; i++ {
		s.Set(2+i, 3, '-', engine.ColorDarkGray, -1)
	}
	s.Text(2, h-2, b.hint, engine.ColorGray)
	b.paddle.Draw(s)
	if b.ballActive {
		b.ball.Draw(s)
	}
	if g.Paused {
		engine.TextCentered(s, h/2, "=== 暂停中 (P 继续) ===", engine.ColorYellow)
	} else if b.settled || b.over {
		b.renderBoard(s)
	}
}

func (b *Bounce) renderBoard(s *engine.Screen) {
	title := "游戏结束"
	if !b.over {
		title = "已结算"
	}
	engine.TextCentered(s, h/2-6, "===== "+title+" =====", engine.ColorYellow)
	engine.TextCentered(s, h/2-5, fmt.Sprintf("最终分数 %d", b.score), engine.ColorWhite)
	top := b.scores.Top(8)
	if len(top) == 0 {
		engine.TextCentered(s, h/2-3, "排行榜暂无记录", engine.ColorGray)
	} else {
		for i, r := range top {
			line := fmt.Sprintf("#%d  %s  %d  %s", i+1, r.Name, r.Score, r.Extra)
			fg := engine.ColorGray
			if i+1 == b.lastRank+1 {
				fg = engine.ColorGreen
			}
			engine.TextCentered(s, h/2-3+i, line, fg)
		}
	}
	engine.TextCentered(s, h/2+6, "R 重开 | Q 退出", engine.ColorGray)
}

func (b *Bounce) onQuit(g *engine.Game) {}
