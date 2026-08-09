// 纹理实验室 - 展示 gocligames 的 ASCII 精细纹理能力。
// 按 1/2/3/4 切换纹理模式，Q 退出。
package main

import (
	"math"

	"gocligames/engine"
)

const (
	scrW, scrH = 60, 22
	artX, artY = 3, 3
)

// Game 演示状态
type Game struct {
	g    *engine.Game
	mode int
	t    int // 帧计数（动画用）
}

func (gm *Game) onKey(g *engine.Game, key string) {
	switch key {
	case "1":
		gm.mode = 1
	case "2":
		gm.mode = 2
	case "3":
		gm.mode = 3
	case "4":
		gm.mode = 4
	case "q":
		g.Quit()
	}
}

// 模式1：亮度字符渐变球（字符密度模拟明暗）
func drawGradientBall(s *engine.Screen, x, y, w, h int) {
	ramp := " .:-=+*#%@"
	for j := 0; j < h; j++ {
		for i := 0; i < w; i++ {
			nx := float64(i)/float64(w)*2 - 1
			ny := float64(j)/float64(h)*2 - 1
			d := math.Sqrt(nx*nx + ny*ny)
			if d > 1 {
				continue
			}
			z := math.Sqrt(1 - d*d)
			lum := z*0.85 + 0.15*math.Sin(nx*5)*math.Cos(ny*5) + 0.1*(1-d)
			if lum < 0 {
				lum = 0
			}
			if lum > 1 {
				lum = 1
			}
			idx := int(lum * float64(len(ramp)-1))
			s.Set(x+i, y+j, rune(ramp[idx]), engine.ColorCyan, -1)
		}
	}
}

// 模式2：彩色半块球（44x26 像素 → 13 行画布，垂直分辨率翻倍）
func halfSpherePixels(w, h int) [][]int {
	pix := make([][]int, h)
	for j := 0; j < h; j++ {
		pix[j] = make([]int, w)
		for i := 0; i < w; i++ {
			nx := float64(i)/float64(w)*2 - 1
			ny := float64(j)/float64(h)*2 - 1
			d := math.Sqrt(nx*nx + ny*ny)
			if d > 1 {
				pix[j][i] = -1
				continue
			}
			z := math.Sqrt(1 - d*d)
			lum := z*0.8 + 0.2*(1-d)
			c := 17 + int(lum*28) // 蓝紫渐变
			if lum > 0.9 {
				c = 231 // 高光白
			}
			pix[j][i] = c
		}
	}
	return pix
}

// 模式3：像素魔龙（Sprite 字符纹理贴图）
func dragonSprite() *engine.Sprite {
	sp := engine.NewSprite(
		"          ▄▄▄▄▄▄▄▄▄▄",
		"        ▄██████████████▄",
		"       ██████████████████▄",
		"      ████████████████████▄",
		"      ████▀▀██████▀▀██████",
		"      ████▀▀██████▀▀██████",
		"      ████████████████████",
		"       ███▓▓▓▓████▓▓▓▓███",
		"       ██▓▓▓▓▓▓██▓▓▓▓▓▓██",
		"        ██▓▓▓▓▓▓▓▓▓▓▓▓██",
		"         ██████████████",
		"          ████▄▄▄▄████",
	)
	sp.Palette = map[rune]int{
		'█': 40,  // 龙身绿
		'▄': 28,  // 尾部暗绿
		'▀': 46,  // 高光亮绿
		'▓': 51,  // 腹部亮青
	}
	return sp
}

// 模式4：程序化火焰（噪声 + 半块渲染，动态纹理）
func firePixels(w, h, t int) [][]int {
	pix := make([][]int, h)
	ramp := []int{52, 88, 124, 160, 196, 208, 226, 230}
	for j := 0; j < h; j++ {
		pix[j] = make([]int, w)
		for i := 0; i < w; i++ {
			yy := float64(h-j) / float64(h) // 0=顶 1=底
			noise := math.Sin(float64(i)*0.35+float64(t)*0.45)*0.22 +
				math.Sin(float64(i)*0.19-float64(t)*0.3+float64(j)*0.45)*0.14 +
				math.Sin(float64(t)*0.2)*0.06
			v := (yy + noise) * 1.15
			if v <= 0 {
				pix[j][i] = -1
				continue
			}
			idx := int(v * float64(len(ramp)-1))
			if idx >= len(ramp) {
				idx = len(ramp) - 1
			}
			pix[j][i] = ramp[idx]
		}
	}
	return pix
}

func (gm *Game) render(g *engine.Game, s *engine.Screen) {
	s.Clear()
	s.Text(1, 0, "gocligames 纹理实验室", engine.ColorYellow)
	s.Text(1, 1, "1渐变球 2半块彩球 3像素龙 4火焰 Q退", engine.ColorGray)
	gm.t++
	switch gm.mode {
	case 1:
		drawGradientBall(s, artX, artY, 44, 14)
	case 2:
		engine.HalfBlock(s, artX, artY, halfSpherePixels(44, 26))
	case 3:
		dragonSprite().Draw(s, artX, artY)
	case 4:
		engine.HalfBlock(s, artX, artY, firePixels(48, 30, gm.t))
	}
	var desc string
	switch gm.mode {
	case 1:
		desc = "亮度字符渐变球：字符密度模拟明暗"
	case 2:
		desc = "半块渲染：每格上下两色，垂直分辨率翻倍"
	case 3:
		desc = "Sprite 精灵：字符像素画加调色板贴图"
	case 4:
		desc = "程序化火焰：噪声加半块渲染，动态纹理"
	}
	s.Text(1, scrH-2, desc, engine.ColorCyan)
}

func main() {
	g := engine.NewGame("纹理实验室", scrW, scrH, 20)
	gm := &Game{g: g, mode: 2}
	g.OnKey = gm.onKey
	g.Render = gm.render
	g.Run()
}
