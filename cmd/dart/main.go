// 江湖·接暗器 - 快节奏反应小游戏（gocligames 引擎）
// 玩法：A/D 或方向键移动，接金镖得分、躲毒镖保命，生命 3 条
// 排行榜：dart_scores.json Top10，E 结算 Q 退出（自动保存）
package main

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"

	"gocligames/engine"
)

const (
	scrW, scrH = 40, 20
	startLives = 3
)

// Dart 一枚暗器
type Dart struct {
	X, Y   float64
	Speed  float64
	Kind   int // 0 金镖 +10分, 1 毒镖 -1命
	Ch     rune
	Fg     int
	Active bool
}

// Game 主状态
type Game struct {
	g       *engine.Game
	px      float64 // 玩家 x
	lives   int
	score   int
	speedUp float64
	darts   []*Dart
	spawnT  float64
	flashT  float64
	flash   string
	mode    string // title play over
	overMsg string
	sb      *engine.Scoreboard
	rankIn  int
	settled bool
}

func newGame() *Game {
	gm := &Game{mode: "title"}
	dir, _ := os.Getwd()
	gm.sb = engine.NewScoreboard(filepath.Join(dir, "dart_scores.json"), 10)
	return gm
}

func (gm *Game) reset() {
	gm.px = float64(scrW) / 2
	gm.lives = startLives
	gm.score = 0
	gm.speedUp = 0
	gm.darts = nil
	gm.spawnT = 0
	gm.flashT = 0
	gm.settled = false
	gm.mode = "play"
}

func (gm *Game) spawnDart() {
	kind := 0
	if rand.Intn(100) < 30 {
		kind = 1 // 30% 毒镖
	}
	d := &Dart{
		X:      float64(1 + rand.Intn(scrW-2)),
		Y:      0,
		Speed:  2.0 + gm.speedUp + rand.Float64()*1.5,
		Kind:   kind,
		Active: true,
	}
	if kind == 0 {
		d.Ch, d.Fg = '镖', engine.ColorYellow
	} else {
		d.Ch, d.Fg = '毒', engine.ColorGreen
	}
	gm.darts = append(gm.darts, d)
}

// ---------- 引擎钩子 ----------

func (gm *Game) onKey(g *engine.Game, key string) {
	switch gm.mode {
	case "title":
		if key == "enter" || key == "space" || key == "e" {
			gm.reset()
		} else if key == "q" {
			g.Quit()
		}
	case "play":
		switch key {
		case "a", "left":
			gm.px -= 1.5
		case "d", "right":
			gm.px += 1.5
		case "e":
			gm.settle()
		case "q":
			gm.quitSave()
			g.Quit()
		}
	case "over":
		switch key {
		case "r":
			gm.reset()
		case "e":
			gm.settle()
		case "q":
			g.Quit()
		}
	}
}

func (gm *Game) settle() {
	if gm.settled {
		return
	}
	gm.settled = true
	gm.rankIn = gm.sb.Add("无名侠客", gm.score, "")
	gm.overMsg = "已写入排行榜！"
}

func (gm *Game) quitSave() {
	if !gm.settled && gm.score > 0 {
		gm.rankIn = gm.sb.Add("无名侠客", gm.score, "")
	}
}

func (gm *Game) update(g *engine.Game, dt float64) {
	if gm.mode != "play" {
		return
	}
	// 限位
	if gm.px < 0 {
		gm.px = 0
	}
	if gm.px > scrW-1 {
		gm.px = scrW - 1
	}
	// 生成暗器（间隔随速度缩短）
	gm.spawnT -= dt
	if gm.spawnT <= 0 {
		gm.spawnDart()
		gm.spawnT = 1.1 - gm.speedUp*0.02
		if gm.spawnT < 0.25 {
			gm.spawnT = 0.25
		}
	}
	// 移动暗器
	for _, d := range gm.darts {
		if !d.Active {
			continue
		}
		d.Y += d.Speed * dt
		// 接住判定（玩家在底部一行）
		if d.Y >= float64(scrH-1) && d.Active {
			if int(d.X) >= int(gm.px)-1 && int(d.X) <= int(gm.px)+1 {
				if d.Kind == 0 {
					gm.score += 10
					gm.speedUp += 0.04
					gm.flash = "+10"
					gm.flashT = 0.5
				} else {
					gm.lives--
					gm.flash = "中毒！-1命"
					gm.flashT = 0.5
				}
			} else {
				if d.Kind == 0 {
					gm.flash = "漏接金镖"
					gm.flashT = 0.4
				}
			}
			d.Active = false
			if gm.lives <= 0 {
				gm.mode = "over"
				gm.overMsg = "暗器无情，江湖梦断……"
				gm.quitSave()
			}
		}
		if d.Y > float64(scrH) {
			d.Active = false
		}
	}
	if gm.flashT > 0 {
		gm.flashT -= dt
	}
	// 清理
	alive := gm.darts[:0]
	for _, d := range gm.darts {
		if d.Active {
			alive = append(alive, d)
		}
	}
	gm.darts = alive
}

// ---------- 渲染 ----------

func (gm *Game) render(g *engine.Game, s *engine.Screen) {
	s.Clear()
	switch gm.mode {
	case "title":
		gm.renderTitle(s)
	case "play":
		gm.renderPlay(s)
	case "over":
		gm.renderOver(s)
	}
}

func (gm *Game) renderTitle(s *engine.Screen) {
	engine.TextCentered(s, 3, "江 湖 · 接 暗 器", engine.ColorYellow)
	engine.TextCentered(s, 5, "金镖 +10 分，毒镖 -1 命", engine.ColorWhite)
	engine.TextCentered(s, 7, "A/D 或 方向键 移动", engine.ColorCyan)
	engine.TextCentered(s, 9, "E 结算  Q 退出", engine.ColorGray)
	engine.TextCentered(s, 12, "按 Enter 开始", engine.ColorYellow)
	top := gm.sb.Top(3)
	if len(top) > 0 {
		engine.TextCentered(s, 15, "--- 江湖榜 Top3 ---", engine.ColorGray)
		for i, r := range top {
			engine.TextCentered(s, 16+i, fmt.Sprintf("%d. %s  %d分", i+1, r.Name, r.Score), engine.ColorWhite)
		}
	}
}

func (gm *Game) renderPlay(s *engine.Screen) {
	// 状态栏
	s.Text(1, 0, fmt.Sprintf("生命 %d", gm.lives), engine.ColorRed)
	s.Text(15, 0, fmt.Sprintf("得分 %d", gm.score), engine.ColorYellow)
	s.Text(30, 0, "E结算 Q退出", engine.ColorGray)
	// 暗器
	for _, d := range gm.darts {
		if d.Active {
			s.Set(int(d.X), int(d.Y), d.Ch, d.Fg, -1)
		}
	}
	// 玩家
	s.Set(int(gm.px), scrH-1, '@', engine.ColorCyan, 17)
	// 闪烁提示
	if gm.flashT > 0 && gm.flash != "" {
		engine.TextCentered(s, 2, gm.flash, engine.ColorMagenta)
	}
}

func (gm *Game) renderOver(s *engine.Screen) {
	engine.TextCentered(s, 5, "—— 暗器无情 ——", engine.ColorRed)
	engine.TextCentered(s, 7, fmt.Sprintf("最终得分 %d", gm.score), engine.ColorYellow)
	if gm.rankIn >= 0 && gm.rankIn < 10 {
		engine.TextCentered(s, 8, fmt.Sprintf("名列江湖榜第 %d 位", gm.rankIn+1), engine.ColorWhite)
	}
	engine.TextCentered(s, 9, gm.overMsg, engine.ColorGray)
	engine.TextCentered(s, 12, "R 再来一局  E 结算  Q 退出", engine.ColorGray)
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--bot" {
		RunBot()
		return
	}
	gm := newGame()
	g := engine.NewGame("江湖·接暗器", scrW, scrH, 30)
	gm.g = g
	g.OnKey = gm.onKey
	g.Update = gm.update
	g.Render = gm.render
	g.Run()
}
