package engine

import (
	"bufio"
	"os"
)

// Input 非阻塞按键输入：后台 goroutine 持续读 stdin 并归一化按键，
// 主循环通过 Poll 无阻塞取键。跨平台行为一致（Unix raw mode / Windows console）。
type Input struct {
	ch   chan string
	done chan struct{}
}

// NewInput 启动输入读取 goroutine。
func NewInput() *Input {
	in := &Input{
		ch:   make(chan string, 16),
		done: make(chan struct{}),
	}
	go in.readLoop()
	return in
}

func (in *Input) readLoop() {
	defer close(in.ch)
	r := bufio.NewReader(os.Stdin)
	for {
		select {
		case <-in.done:
			return
		default:
		}
		ch, _, err := r.ReadRune()
		if err != nil {
			return
		}
		if ch == 0x1b {
			in.handleEscape(r)
			continue
		}
		if k := normalizeKey(ch); k != "" {
			in.push(k)
		}
	}
}

// handleEscape 解析 ESC 开头的序列：Unix 方向键 ESC[A 等；Windows scan code ESC NUL H 等。
func (in *Input) handleEscape(r *bufio.Reader) {
	r1, _, err := r.ReadRune()
	if err != nil {
		in.push("esc")
		return
	}
	if r1 == '[' || r1 == 'O' {
		r2, _, err2 := r.ReadRune()
		if err2 != nil {
			in.push("esc")
			return
		}
		switch r2 {
		case 'A':
			in.push("up")
		case 'B':
			in.push("down")
		case 'C':
			in.push("right")
		case 'D':
			in.push("left")
		case 'H':
			in.push("home")
		case 'F':
			in.push("end")
		default:
			in.push("esc")
		}
		return
	}
	if r1 == 0x00 || r1 == 0xe0 {
		// Windows 方向键 scan code（NUL 或 0xE0 引导）
		r2, _, err2 := r.ReadRune()
		if err2 != nil {
			in.push("esc")
			return
		}
		switch r2 {
		case 'H':
			in.push("up")
		case 'P':
			in.push("down")
		case 'K':
			in.push("left")
		case 'M':
			in.push("right")
		default:
			in.push("esc")
		}
		return
	}
	_ = r.UnreadRune() // 普通字符在 ESC 后（如 Alt 组合），退回给下次读取
	in.push("esc")
}

// normalizeKey 把单个 rune 归一化为按键名。
func normalizeKey(r rune) string {
	switch r {
	case '\r', '\n':
		return "enter"
	case ' ':
		return "space"
	case '\t':
		return "tab"
	case 0x03:
		return "ctrl_c"
	case 0x04:
		return "ctrl_d"
	}
	if r >= 'A' && r <= 'Z' {
		return string(r - 'A' + 'a')
	}
	if r >= 0x20 && r < 0x7f {
		return string(r)
	}
	return ""
}

func (in *Input) push(k string) {
	select {
	case in.ch <- k:
	default: // 队列满则丢弃（游戏卡顿时按键丢失可接受）
	}
}

// Poll 非阻塞取一个按键；无按键返回空串。
func (in *Input) Poll() string {
	select {
	case k, ok := <-in.ch:
		if !ok {
			return ""
		}
		return k
	default:
		return ""
	}
}

// Close 停止读取 goroutine。
func (in *Input) Close() {
	select {
	case <-in.done:
	default:
		close(in.done)
	}
}
