package engine

// Entity 游戏实体：位置、速度、尺寸、外观与碰撞盒。
type Entity struct {
	X, Y   float64 // 位置（左上角，浮点便于平滑移动）
	W, H   int     // 尺寸（格数）
	Vx, Vy float64 // 速度（格/秒）
	Ch     rune    // 渲染字符
	Fg     int     // 颜色
	Active bool    // 是否参与逻辑/绘制
}

// NewEntity 创建实体。
func NewEntity(x, y float64, w, h int, ch rune, fg int) *Entity {
	return &Entity{X: x, Y: y, W: w, H: h, Ch: ch, Fg: fg, Active: true}
}

// Update 按速度位移。
func (e *Entity) Update(dt float64) {
	if !e.Active {
		return
	}
	e.X += e.Vx * dt
	e.Y += e.Vy * dt
}

// Rect 返回 AABB 边界 (x0, y0, x1, y1)。
func (e *Entity) Rect() (x0, y0, x1, y1 float64) {
	return e.X, e.Y, e.X + float64(e.W), e.Y + float64(e.H)
}

// CenterX 水平中心。
func (e *Entity) CenterX() float64 { return e.X + float64(e.W)/2 }

// CenterY 垂直中心。
func (e *Entity) CenterY() float64 { return e.Y + float64(e.H)/2 }

// Draw 把实体画到画布（多格宽/高逐格绘制）。
func (e *Entity) Draw(s *Screen) {
	if !e.Active {
		return
	}
	for j := 0; j < e.H; j++ {
		for i := 0; i < e.W; i++ {
			s.Set(int(e.X)+i, int(e.Y)+j, e.Ch, e.Fg, -1)
		}
	}
}
