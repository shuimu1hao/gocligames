package engine

import (
	"fmt"
	"os"
	"time"
)

// Game 游戏主循环与生命周期管理。
// 子游戏通过函数字段注入行为（OnStart/OnKey/Update/Render/OnQuit）。
type Game struct {
	Title  string
	Screen *Screen
	Input  *Input
	FPS    int
	Paused bool

	running bool

	OnStart func(g *Game)
	OnKey   func(g *Game, key string)
	Update  func(g *Game, dt float64)
	Render  func(g *Game, s *Screen)
	OnQuit  func(g *Game)
}

// NewGame 创建游戏实例。
func NewGame(title string, w, h, fps int) *Game {
	return &Game{
		Title:  title,
		Screen: NewScreen(w, h),
		Input:  NewInput(),
		FPS:    fps,
	}
}

// Run 进入主循环：设置终端 → 循环（按键/更新/渲染）→ 恢复终端。
// 自动处理 Ctrl+C 退出与终端状态恢复。
func (g *Game) Run() {
	restore, err := makeRaw()
	if err != nil {
		fmt.Fprintln(os.Stderr, "gocligames: raw mode unavailable:", err)
	} else {
		defer restore()
	}
	defer g.Input.Close()
	defer g.showCursor()
	g.hideCursor()
	if g.OnStart != nil {
		g.OnStart(g)
	}
	g.running = true
	frameTime := time.Second / time.Duration(g.FPS)
	last := time.Now()
	for g.running {
		now := time.Now()
		dt := now.Sub(last).Seconds()
		last = now
		if dt > 0.1 {
			dt = 0.1 // 防止卡顿后大跳跃
		}
		// 消费本帧所有排队按键
		for {
			k := g.Input.Poll()
			if k == "" {
				break
			}
			if k == "ctrl_c" {
				g.running = false
				break
			}
			if g.OnKey != nil {
				g.OnKey(g, k)
			}
		}
		if !g.Paused && g.Update != nil {
			g.Update(g, dt)
		}
		if g.Render != nil {
			g.Render(g, g.Screen)
		}
		g.Screen.Render()
		if sleep := frameTime - time.Since(now); sleep > 0 {
			time.Sleep(sleep)
		}
	}
	if g.OnQuit != nil {
		g.OnQuit(g)
	}
}

// Quit 请求退出主循环。
func (g *Game) Quit() {
	g.running = false
}

func (g *Game) hideCursor() {
	os.Stdout.WriteString("\x1b[?25l")
}

func (g *Game) showCursor() {
	os.Stdout.WriteString("\x1b[0m\x1b[?25h")
}

// ---------- 绘制工具 ----------

// Clamp 把 v 限制在 [lo, hi]。
func Clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// DispWidth 估算字符串显示宽度（中文/全角按 2 列）。
func DispWidth(str string) int {
	w := 0
	for _, r := range str {
		if r > 0x2E7F {
			w += 2
		} else {
			w++
		}
	}
	return w
}

// TextCentered 在 y 行水平居中画文本。
func TextCentered(s *Screen, y int, str string, fg int) {
	s.Text((s.W-DispWidth(str))/2, y, str, fg)
}

// Box 画矩形边框。
func Box(s *Screen, x, y, w, h int, fg int) {
	for i := 0; i < w; i++ {
		s.Set(x+i, y, '-', fg, -1)
		s.Set(x+i, y+h-1, '-', fg, -1)
	}
	for j := 0; j < h; j++ {
		s.Set(x, y+j, '|', fg, -1)
		s.Set(x+w-1, y+j, '|', fg, -1)
	}
	s.Set(x, y, '+', fg, -1)
	s.Set(x+w-1, y, '+', fg, -1)
	s.Set(x, y+h-1, '+', fg, -1)
	s.Set(x+w-1, y+h-1, '+', fg, -1)
}
