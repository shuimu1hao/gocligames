package engine

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestScreenSetAndBounds(t *testing.T) {
	s := NewScreen(10, 5)
	s.Set(9, 4, 'x', ColorWhite, -1)  // 合法角
	s.Set(10, 4, 'y', ColorWhite, -1) // 越界不 panic
	s.Set(-1, 0, 'y', ColorWhite, -1)
	if s.cells[4][9].Ch != 'x' {
		t.Fatal("Set 未生效")
	}
	frame := s.Frame()
	if !strings.HasPrefix(frame, "\x1b[H") {
		t.Fatal("Frame 应以光标归位开头")
	}
	if !strings.Contains(frame, "\x1b[38;5;231m") {
		t.Fatal("Frame 应包含颜色转义")
	}
}

func TestScoreboardSortAndPersist(t *testing.T) {
	p := filepath.Join(t.TempDir(), "scores.json")
	sb := NewScoreboard(p, 3)
	rank := sb.Add("A", 100, "")
	if rank != 0 {
		t.Fatalf("首个记录应为第 0 名，got %d", rank)
	}
	sb.Add("B", 300, "x")
	sb.Add("C", 200, "")
	sb.Add("D", 999, "") // 超限被截断
	if len(sb.Scores) != 3 {
		t.Fatalf("应截断到 3 条，got %d", len(sb.Scores))
	}
	if sb.Scores[0].Name != "D" || sb.Scores[2].Name != "C" {
		t.Fatalf("排序错误: %+v", sb.Scores)
	}
	// 重新加载验证持久化
	sb2 := NewScoreboard(p, 3)
	if len(sb2.Scores) != 3 || sb2.Scores[0].Name != "D" {
		t.Fatalf("持久化加载错误: %+v", sb2.Scores)
	}
}

func TestPhysicsAABB(t *testing.T) {
	a := NewEntity(0, 0, 2, 2, 'a', ColorWhite)
	b := NewEntity(1, 1, 2, 2, 'b', ColorWhite)
	c := NewEntity(5, 5, 1, 1, 'c', ColorWhite)
	if !Overlap(a, b) {
		t.Fatal("a/b 应重叠")
	}
	if Overlap(a, c) {
		t.Fatal("a/c 不应重叠")
	}
}

func TestNormalizeKey(t *testing.T) {
	cases := map[rune]string{
		'a': "a", 'Z': "z", '\r': "enter", ' ': "space", 0x03: "ctrl_c",
	}
	for in, want := range cases {
		if got := normalizeKey(in); got != want {
			t.Fatalf("normalizeKey(%q) = %q, want %q", in, got, want)
		}
	}
}
