package engine

// AABB 判断两个矩形（浮点坐标）是否重叠。
func AABB(ax0, ay0, ax1, ay1, bx0, by0, bx1, by1 float64) bool {
	return !(ax1 < bx0 || bx1 < ax0 || ay1 < by0 || by1 < ay0)
}

// Overlap 判断两个实体碰撞盒是否重叠。
func Overlap(a, b *Entity) bool {
	ax0, ay0, ax1, ay1 := a.Rect()
	bx0, by0, bx1, by1 := b.Rect()
	return AABB(ax0, ay0, ax1, ay1, bx0, by0, bx1, by1)
}
