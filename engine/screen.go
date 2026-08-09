package engine

import (
	"fmt"
	"os"
	"strings"
)

// Cell 是画布上的一个字符单元。
type Cell struct {
	Ch rune
	Fg int // 256 色索引，-1 表示默认前景
	Bg int // 256 色索引，-1 表示默认背景
}

// Screen 双缓冲字符画布。
type Screen struct {
	W, H  int
	cells [][]Cell
}

// NewScreen 创建 w x h 的画布。
func NewScreen(w, h int) *Screen {
	s := &Screen{W: w, H: h}
	s.cells = make([][]Cell, h)
	for y := 0; y < h; y++ {
		s.cells[y] = make([]Cell, w)
	}
	s.Clear()
	return s
}

// Width 画布宽度。
func (s *Screen) Width() int { return s.W }

// Height 画布高度。
func (s *Screen) Height() int { return s.H }

// Clear 清空画布。
func (s *Screen) Clear() {
	for y := 0; y < s.H; y++ {
		for x := 0; x < s.W; x++ {
			s.cells[y][x] = Cell{Ch: ' ', Fg: -1, Bg: -1}
		}
	}
}

// Set 在 (x, y) 画一个字符；越界安全忽略。
func (s *Screen) Set(x, y int, ch rune, fg, bg int) {
	if x < 0 || y < 0 || x >= s.W || y >= s.H {
		return
	}
	s.cells[y][x] = Cell{Ch: ch, Fg: fg, Bg: bg}
}

// Text 在 (x, y) 画一行字符串。
func (s *Screen) Text(x, y int, str string, fg int) {
	for i, r := range str {
		s.Set(x+i, y, r, fg, -1)
	}
}

// Frame 构建整帧 ANSI 字符串（同色相邻字符自动合并转义，减少输出量）。
func (s *Screen) Frame() string {
	var b strings.Builder
	b.WriteString("\x1b[H") // 光标回原点
	curFg, curBg := -2, -2
	for y := 0; y < s.H; y++ {
		if y > 0 {
			b.WriteByte('\n')
		}
		for x := 0; x < s.W; x++ {
			c := s.cells[y][x]
			if c.Fg != curFg {
				if c.Fg < 0 {
					b.WriteString("\x1b[39m")
				} else {
					fmt.Fprintf(&b, "\x1b[38;5;%dm", c.Fg)
				}
				curFg = c.Fg
			}
			if c.Bg != curBg {
				if c.Bg < 0 {
					b.WriteString("\x1b[49m")
				} else {
					fmt.Fprintf(&b, "\x1b[48;5;%dm", c.Bg)
				}
				curBg = c.Bg
			}
			b.WriteRune(c.Ch)
		}
	}
	b.WriteString("\x1b[0m\x1b[J")
	return b.String()
}

// Render 输出整帧到终端。
func (s *Screen) Render() {
	os.Stdout.WriteString(s.Frame())
}
