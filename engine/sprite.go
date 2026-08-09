// 字符纹理与高分辨率渲染辅助（2026-08-08 新增）。
// 让游戏能用 ASCII 画出精细贴图：多字符精灵 + 半块高分辨率渲染。
package engine

// Sprite 字符纹理：多行字符 + 调色板，可绘制精细图案。
// 每个可见字符按 Palette 映射 256 色；' ' 或 Transparent 中的字符不绘制（透明）。
type Sprite struct {
	Lines       []string
	Palette     map[rune]int
	Transparent map[rune]bool
}

// NewSprite 创建精灵，每行一个字符串（可用 r”' 风格直接写多行）。
func NewSprite(lines ...string) *Sprite {
	return &Sprite{Lines: lines, Palette: map[rune]int{}, Transparent: map[rune]bool{}}
}

// W 精灵宽度（最长行）。
func (sp *Sprite) W() int {
	w := 0
	for _, l := range sp.Lines {
		if n := len([]rune(l)); n > w {
			w = n
		}
	}
	return w
}

// H 精灵高度（行数）。
func (sp *Sprite) H() int { return len(sp.Lines) }

// Draw 把精灵绘制到画布 (x, y) 为左上角；越界由 Screen.Set 安全忽略。
func (sp *Sprite) Draw(s *Screen, x, y int) {
	for j, line := range sp.Lines {
		for i, r := range line {
			if r == ' ' || sp.Transparent[r] {
				continue
			}
			fg, ok := sp.Palette[r]
			if !ok {
				fg = ColorWhite
			}
			s.Set(x+i, y+j, r, fg, -1)
		}
	}
}

// HalfBlock 半块高分辨率渲染：把 2*H 行虚拟像素画到 H 行画布（垂直分辨率翻倍）。
// pixels[py][px] 为 256 色索引，-1 表示透明。
// 原理：每格字符 '▀' 的上半块用前景色、下半块用背景色，一格显示两个像素。
func HalfBlock(s *Screen, x, y int, pixels [][]int) {
	for py := 0; py < len(pixels); py += 2 {
		row := y + py/2
		for px := 0; px < len(pixels[py]); px++ {
			up := pixels[py][px]
			dn := -1
			if py+1 < len(pixels) && px < len(pixels[py+1]) {
				dn = pixels[py+1][px]
			}
			ch := ' '
			fg, bg := -1, -1
			switch {
			case up >= 0 && dn >= 0:
				ch, fg, bg = '▀', up, dn
			case up >= 0:
				ch, fg = '▀', up
			case dn >= 0:
				ch, fg = '▄', dn
			}
			s.Set(x+px, row, ch, fg, bg)
		}
	}
}
