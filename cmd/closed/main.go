// 江湖·闭关 - 放置修炼小游戏（gocligames 引擎）
// 玩法：选择闭关时长挂机，灵气积累突破境界，冲修行榜
// 排行榜：closed_scores.json Top10（按修为）
package main

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"

	"gocligames/engine"
)

const (
	scrW, scrH = 40, 18
)

var realmNames = []string{"炼气", "筑基", "金丹", "元婴", "化神", "渡劫", "大乘"}

// 修为阈值（每层境界所需累计修为）
var realmNeed = []int{0, 300, 800, 1500, 2600, 4000, 6000}

// Game 主状态
type Game struct {
	g        *engine.Game
	mode     string // menu closed result
	realm    int
	xiuwei   int
	closedT  float64 // 当前闭关剩余
	closedN  float64 // 当前闭关总时长
	lastGain int
	msg      string
	tick     float64
	sb       *engine.Scoreboard
	rankIn   int
	settled  bool
}

func newGame() *Game {
	gm := &Game{mode: "menu"}
	dir, _ := os.Getwd()
	gm.sb = engine.NewScoreboard(filepath.Join(dir, "closed_scores.json"), 10)
	return gm
}

// 闭关时长选项（秒）
var durations = []float64{10, 20, 40}

func (gm *Game) startClosed(n float64) {
	gm.closedN = n
	gm.closedT = n
	gm.lastGain = 0
	gm.mode = "closed"
}

// 结算闭关收益
func (gm *Game) finishClosed() {
	// 收益 = 时长 x 效率（境界越高灵气越浓）+ 随机
	efficiency := 1.0 + float64(gm.realm)*0.5
	gain := int(gm.closedN * float64(10+rand.Intn(10)) * efficiency / 10)
	gm.xiuwei += gain
	gm.lastGain = gain
	// 突破
	for gm.realm < len(realmNames)-1 && gm.xiuwei >= realmNeed[gm.realm+1] {
		gm.realm++
		gm.msg = fmt.Sprintf("突破！踏入%s境界！", realmNames[gm.realm])
	}
	if gm.msg == "" {
		gm.msg = fmt.Sprintf("修为 +%d，距离下一境界还差 %d 修为", gain, realmNeed[gm.realm+1]-gm.xiuwei)
	}
	gm.mode = "result"
}

func (gm *Game) settle() {
	if gm.settled {
		return
	}
	gm.settled = true
	score := gm.xiuwei + gm.realm*1000
	gm.rankIn = gm.sb.Add("无名侠客", score, realmNames[gm.realm])
}

// ---------- 引擎钩子 ----------

func (gm *Game) onKey(g *engine.Game, key string) {
	switch gm.mode {
	case "menu":
		switch key {
		case "1", "2", "3":
			idx := int(key[0] - '1')
			if idx < len(durations) {
				gm.startClosed(durations[idx])
			}
		case "e":
			gm.settle()
			gm.msg = "已写入修行榜！"
		case "q":
			g.Quit()
		}
	case "closed":
		if key == "q" {
			gm.mode = "menu"
		}
	case "result":
		switch key {
		case "enter", "space", "r":
			gm.mode = "menu"
		case "e":
			gm.settle()
			gm.msg = "已写入修行榜！"
		case "q":
			g.Quit()
		}
	}
}

func (gm *Game) update(g *engine.Game, dt float64) {
	if gm.mode != "closed" {
		return
	}
	gm.closedT -= dt
	gm.tick += dt
	if gm.closedT <= 0 {
		gm.finishClosed()
	}
}

// ---------- 渲染 ----------

func (gm *Game) render(g *engine.Game, s *engine.Screen) {
	s.Clear()
	switch gm.mode {
	case "menu":
		gm.renderMenu(s)
	case "closed":
		gm.renderClosed(s)
	case "result":
		gm.renderResult(s)
	}
}

func (gm *Game) renderMenu(s *engine.Screen) {
	engine.TextCentered(s, 2, "江 湖 · 闭 关", engine.ColorYellow)
	engine.TextCentered(s, 4, fmt.Sprintf("%s境界 · 修为 %d", realmNames[gm.realm], gm.xiuwei), engine.ColorCyan)
	engine.TextCentered(s, 6, "选择闭关时长：", engine.ColorWhite)
	engine.TextCentered(s, 8, "1. 闭关 10 秒（~100 修为）", engine.ColorGray)
	engine.TextCentered(s, 9, "2. 闭关 20 秒（~200 修为）", engine.ColorGray)
	engine.TextCentered(s, 10, "3. 闭关 40 秒（~400 修为）", engine.ColorGray)
	engine.TextCentered(s, 12, "E 写入修行榜  Q 退出", engine.ColorGray)
	if gm.msg != "" {
		engine.TextCentered(s, 14, gm.msg, engine.ColorMagenta)
	}
	top := gm.sb.Top(3)
	if len(top) > 0 {
		engine.TextCentered(s, 16, "修行榜 Top3", engine.ColorGray)
	}
}

func (gm *Game) renderClosed(s *engine.Screen) {
	engine.TextCentered(s, 3, "—— 闭 关 修 炼 中 ——", engine.ColorYellow)
	// 灵气光点
	for i := 0; i < 8; i++ {
		x := 4 + (gm.tick*float64(i+1)*7 + float64(i)*13)
		y := 6 + i%5
		s.Set(int(x)%(scrW-8)+4, y, '.', engine.ColorCyan, -1)
	}
	// 倒计时条
	left := gm.closedT / gm.closedN
	bar := int(left * 20)
	line := ""
	for i := 0; i < 20; i++ {
		if i < bar {
			line += "#"
		} else {
			line += "."
		}
	}
	engine.TextCentered(s, 8, "剩余", engine.ColorWhite)
	engine.TextCentered(s, 9, line, engine.ColorCyan)
	engine.TextCentered(s, 10, fmt.Sprintf("%.1f 秒", gm.closedT), engine.ColorWhite)
	engine.TextCentered(s, 12, fmt.Sprintf("%s境界 · 修为 %d", realmNames[gm.realm], gm.xiuwei), engine.ColorGray)
	engine.TextCentered(s, 14, "Q 中断闭关", engine.ColorGray)
}

func (gm *Game) renderResult(s *engine.Screen) {
	engine.TextCentered(s, 3, "—— 出 关 ——", engine.ColorYellow)
	engine.TextCentered(s, 5, fmt.Sprintf("本次修为 +%d", gm.lastGain), engine.ColorCyan)
	engine.TextCentered(s, 6, fmt.Sprintf("当前：%s境界 · 修为 %d", realmNames[gm.realm], gm.xiuwei), engine.ColorWhite)
	lines := engineWrap(gm.msg, 18)
	for i, ln := range lines {
		engine.TextCentered(s, 8+i, ln, engine.ColorMagenta)
	}
	engine.TextCentered(s, 12, "Enter 继续闭关  E 写榜  Q 退出", engine.ColorGray)
}

// engineWrap 按显示宽度折行
func engineWrap(s string, width int) []string {
	r := []rune(s)
	var out []string
	for len(r) > 0 {
		n := width
		if len(r) < n {
			n = len(r)
		}
		out = append(out, string(r[:n]))
		r = r[n:]
	}
	return out
}

func main() {
	gm := newGame()
	g := engine.NewGame("江湖·闭关", scrW, scrH, 30)
	gm.g = g
	g.OnKey = gm.onKey
	g.Update = gm.update
	g.Render = gm.render
	g.Run()
}
