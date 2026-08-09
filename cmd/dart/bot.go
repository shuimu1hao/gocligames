// 江湖·接暗器 - AI 自动玩家（--bot 模式）
// 策略：追踪最近的底部金镖，毒镖靠近时躲避；自动重开 10 局统计成绩
package main

import (
	"fmt"

	"gocligames/engine"
)

// Bot 是自动玩家（AI 打榜/自测用，走游戏自身逻辑）。
type Bot struct {
	gm      *Game
	frames  int
	runFrm  int
	runs    int
	done    bool
	best    int
	total   int
	gold    int
	poison  int
	maxRunF int
}

func (b *Bot) think() {
	b.frames++
	if b.done {
		return
	}
	if b.frames%3 != 0 {
		return
	}
	gm := b.gm
	switch gm.mode {
	case "title":
		gm.reset()
	case "over":
		b.runs++
		b.total += gm.score
		if gm.score > b.best {
			b.best = gm.score
		}
		if b.runs >= 10 {
			b.done = true
			gm.g.Quit()
			return
		}
		gm.reset()
	case "play":
		b.runFrm++
		if b.runFrm > b.maxRunF {
			// 单局超时（AI 太强一直不死）：强制结算
			gm.mode = "over"
			gm.quitSave()
			return
		}
		b.play()
	}
}

func (b *Bot) play() {
	gm := b.gm
	// 毒镖威胁：底部 4 行内且水平距离 < 3 → 反向躲避
	threat := false
	threatDir := 1
	for _, d := range gm.darts {
		if !d.Active || d.Kind != 1 {
			continue
		}
		if d.Y > float64(scrH-4) && abs(int(d.X)-int(gm.px)) < 3 {
			threat = true
			if int(d.X) >= int(gm.px) {
				threatDir = -1
			} else {
				threatDir = 1
			}
		}
	}
	if threat {
		gm.px += float64(threatDir) * 2
		b.poison++
		return
	}
	// 追最近的底部金镖（Y 权重高，越近越好）
	bestScore := -1
	bestX := -1
	for _, d := range gm.darts {
		if !d.Active || d.Kind != 0 {
			continue
		}
		s := int(d.Y)*100 - abs(int(d.X)-int(gm.px))
		if s > bestScore {
			bestScore = s
			bestX = int(d.X)
		}
	}
	if bestX >= 0 {
		if bestX > int(gm.px) {
			gm.px += 1.5
		} else if bestX < int(gm.px) {
			gm.px -= 1.5
		}
		b.gold++
		return
	}
	// 无金镖：轻微晃动
	gm.px += 0.8
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// RunBot bot 模式入口
func RunBot() {
	gm := newGame()
	// 1x1 画布防刷屏（engine.Run 每帧无条件 Screen.Render）
	g := engine.NewGame("江湖·接暗器(bot)", 1, 1, 30)
	gm.g = g
	bot := &Bot{gm: gm, maxRunF: 300} // 每局约 10 秒上限
	g.OnKey = gm.onKey
	g.Update = func(gg *engine.Game, dt float64) {
		gm.update(gg, dt) // 游戏逻辑：生成/移动暗器
		bot.think()       // AI 决策
	}
	g.Render = func(gg *engine.Game, s *engine.Screen) {}
	g.OnQuit = func(gg *engine.Game) {
		avg := 0
		if bot.runs > 0 {
			avg = bot.total / bot.runs
		}
		fmt.Printf("AI 自动游玩 %d 局：最高 %d 分 | 平均 %d 分 | 金镖追踪 %d 次 | 毒镖躲避 %d 次\n",
			bot.runs, bot.best, avg, bot.gold, bot.poison)
	}
	g.Run()
}
