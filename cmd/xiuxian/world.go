// 修仙地牢 demo - 世界观与数据定义
//
// 世界观：你是青云宗外门弟子。宗门禁地「万骨窟」突生异变，
// 大师兄率队探查后失踪，师父命你入窟寻人。
// 万骨窟分三层：枯骨回廊 → 妖兽巢穴 → 幽冥大殿。
// 通关目标：击败幽冥魔尊，救出大师兄。
package main

import "gocligames/engine"

// ---------- 地形 ----------

// Layer 一层地牢：地形是纯 ASCII 等宽网格（#墙 .地 >下楼梯 <上楼梯），
// 怪物/物品/NPC 是对象层（见 Obj），不画进地形，保证网格不错位。
type Layer struct {
	Name  string
	Tiles [][]rune
	W, H  int
}

func mkTiles(rows []string) [][]rune {
	t := make([][]rune, len(rows))
	for i, r := range rows {
		t[i] = []rune(r)
	}
	return t
}

// init 补齐各层 W/H（tileAt / renderMap 依赖）
func init() {
	for i := range layers {
		layers[i].H = len(layers[i].Tiles)
		if layers[i].H > 0 {
			layers[i].W = len(layers[i].Tiles[0])
		}
	}
}

var layers = []Layer{
	{
		Name: "枯骨回廊",
		Tiles: mkTiles([]string{
			"####################",
			"#..................#",
			"#..................#",
			"#....####....####..#",
			"#....#..#....#..#..#",
			"#....#..#....#..#..#",
			"#....####....####..#",
			"#..................#",
			"#..................#",
			"#..................#",
			"#..........>.......#",
			"####################",
		}),
	},
	{
		Name: "妖兽巢穴",
		Tiles: mkTiles([]string{
			"####################",
			"#..................#",
			"#..##........##....#",
			"#..#..........#....#",
			"#..#..........#....#",
			"#..#..........#....#",
			"#..##........##..>.#",
			"#..................#",
			"#....##......##....#",
			"#....#........#....#",
			"#....#........#....#",
			"####################",
		}),
	},
	{
		Name: "幽冥大殿",
		Tiles: mkTiles([]string{
			"####################",
			"#..................#",
			"#..####....####....#",
			"#..#..#....#..#....#",
			"#..#..#....#..#....#",
			"#..####....####....#",
			"#..................#",
			"#..................#",
			"#..................#",
			"#.................<#",
			"#..................#",
			"####################",
		}),
	},
}

// ---------- 怪物 ----------

// MonsterDef 怪物定义
type MonsterDef struct {
	Name     string
	Ch       rune
	Fg       int
	HP       int
	Atk      int
	Def      int
	XP       int
	MinLing  int     // 掉灵石下限
	MaxLing  int     // 掉灵石上限
	DropPill float64 // 掉回元丹概率
	DropMana float64 // 掉聚灵丹概率
}

var monsterDefs = []MonsterDef{
	{Name: "赤炼蛇妖", Ch: '蛇', Fg: engine.ColorGreen, HP: 22, Atk: 5, Def: 1, XP: 15, MinLing: 1, MaxLing: 3, DropPill: 0.20},
	{Name: "白骨尸傀", Ch: '尸', Fg: engine.ColorGray, HP: 36, Atk: 8, Def: 2, XP: 25, MinLing: 2, MaxLing: 4, DropPill: 0.30},
	{Name: "黑风妖狼", Ch: '狼', Fg: engine.ColorOrange, HP: 26, Atk: 6, Def: 1, XP: 20, MinLing: 1, MaxLing: 4, DropPill: 0.15, DropMana: 0.15},
	{Name: "幽冥魔尊", Ch: '魔', Fg: engine.ColorMagenta, HP: 120, Atk: 12, Def: 4, XP: 100, MinLing: 8, MaxLing: 12, DropPill: 1.0},
}

// ---------- 物品 ----------

// ItemKind 物品类型枚举。
type ItemKind int

const (
	ItemPill     ItemKind = iota // 回元丹：+30 HP（战斗内外都能用）
	ItemLing                     // 灵石：拾取直接进背包计数
	ItemBox                      // 宝箱：按 E 打开随机给东西
	ItemManaPill                 // 聚灵丹：+15 MP（战斗外使用）
)

// ---------- 地图对象（怪/NPC/物品） ----------

// ObjKind 世界对象类型枚举。
type ObjKind int

const (
	ObjMonster ObjKind = iota
	ObjNPC
	ObjItem
)

// Obj 世界对象（地图上的可交互元素）。
type Obj struct {
	X, Y   int
	Ch     rune
	Fg     int
	Bg     int
	Kind   ObjKind
	W, H   int      // 体型（Boss 2x2）
	DefIdx int      // 怪物定义下标 / NPC id（1=守墓老者 2=大师兄）
	Item   ItemKind // 物品类型
	Dead   bool
	Talked bool
	Opened bool
}

// 深色背景色（256 色），让角色在暗色地牢里更醒目
const (
	bgDeepBlue   = 24  // 玩家脚下
	bgDeepGreen  = 22  // 蛇妖
	bgDeepGray   = 236 // 尸傀
	bgDeepOrange = 52  // 妖狼
	bgDeepRed    = 88  // 幽冥魔尊
	bgDarkYellow = 58  // 回元丹
	bgDeepCyan   = 23  // 聚灵丹
	bgDarkBrown  = 94  // 宝箱
	bgMidGray    = 240 // 守墓老者
)

// objColor 按对象类型返回前景/背景色
func objColor(o *Obj) (fg, bg int) {
	switch o.Kind {
	case ObjMonster:
		fg = monsterDefs[o.DefIdx].Fg
		switch o.DefIdx {
		case 0:
			bg = bgDeepGreen
		case 1:
			bg = bgDeepGray
		case 2:
			bg = bgDeepOrange
		case 3:
			bg = bgDeepRed
		}
	case ObjNPC:
		if o.DefIdx == 1 {
			fg, bg = engine.ColorWhite, bgMidGray
		} else {
			fg, bg = engine.ColorYellow, bgDeepGreen
		}
	case ObjItem:
		switch o.Item {
		case ItemPill:
			fg, bg = engine.ColorYellow, bgDarkYellow
		case ItemManaPill:
			fg, bg = engine.ColorCyan, bgDeepCyan
		case ItemLing:
			fg, bg = engine.ColorBlue, bgDeepBlue
		case ItemBox:
			fg, bg = engine.ColorYellow, bgDarkBrown
		}
	}
	return
}

// 各层对象出生表：{kind, x, y, defIdx/item, w, h}
type spawnSpec struct {
	kind   ObjKind
	x, y   int
	defIdx int
	item   ItemKind
	w, h   int
}

var layerSpawns = [][]spawnSpec{
	{ // 第 1 层：枯骨回廊
		{ObjNPC, 3, 4, 1, 0, 1, 1},     // 守墓老者
		{ObjMonster, 3, 9, 0, 0, 1, 1}, // 蛇妖
		{ObjMonster, 15, 9, 0, 0, 1, 1},
		{ObjMonster, 9, 2, 0, 0, 1, 1},
		{ObjItem, 9, 8, 0, ItemPill, 1, 1},
	},
	{ // 第 2 层：妖兽巢穴
		{ObjMonster, 6, 3, 1, 0, 1, 1},  // 尸傀
		{ObjMonster, 8, 9, 1, 0, 1, 1},  // 尸傀
		{ObjMonster, 11, 8, 2, 0, 1, 1}, // 妖狼
		{ObjItem, 16, 4, 0, ItemBox, 1, 1},
		{ObjItem, 3, 10, 0, ItemManaPill, 1, 1},
		{ObjItem, 15, 9, 0, ItemPill, 1, 1},
	},
	{ // 第 3 层：幽冥大殿
		{ObjMonster, 9, 6, 3, 0, 2, 2}, // 幽冥魔尊（Boss 2x2）
		{ObjNPC, 4, 9, 2, 0, 1, 1},     // 大师兄
		{ObjMonster, 16, 3, 1, 0, 1, 1},
		{ObjMonster, 2, 7, 1, 0, 1, 1},
		{ObjItem, 17, 6, 0, ItemBox, 1, 1},
		{ObjItem, 3, 8, 0, ItemPill, 1, 1},
		{ObjItem, 15, 8, 0, ItemPill, 1, 1},
	},
}

// ---------- 对话文本 ----------

// 守墓老者（NPC id=1）
var talkOldMan = [][]string{
	{
		"老朽乃万骨窟守墓人。",
		"年轻人，此窟凶险，越深越恶。",
		"送你一颗回元丹，危急时含服可回元气。",
		"记住：脚下尘土变黑之处，便是幽冥魔尊的地盘。",
	},
	{
		"窟中妖兽嗜血，避无可避。",
		"击杀它们可得灵石，凑足一百颗，",
		"可在山下坊市换一枚筑基丹。",
	},
}

// 大师兄（NPC id=2）
var talkBrotherBefore = [][]string{
	{
		"师弟！你怎么来了……",
		"此地已被幽冥魔尊封住，我走不脱。",
		"那魔头元神就寄在殿中石像上，",
		"你若能斩其肉身，我便能用宗门秘法反制它！",
	},
}

var talkBrotherAfter = [][]string{
	{
		"师弟，你做到了！",
		"幽冥魔尊既灭，万骨窟的封印正在消散。",
		"随我回山，师父定会重赏于你！",
		"（恭喜通关！按 Q 离开，或按 R 再闯一遭）",
	},
}

// ---------- Boss 精细纹理 ----------

// drawBoss 在战斗界面绘制幽冥魔尊的像素画（16x8，紫色魔头）。
// ⚠️ 全部用单宽 ASCII 字符：块元素（█▄▀▓）在 Termux 是 2 列宽，
//
//	16 个会撑爆 30 列画布导致终端自动换行、整屏错乱（2026-08-08 实测）。
//	明暗层次靠颜色表达：'#'=紫躯 '%%'=金牙 '@@'=白瞳 '-'=深红边缘。
func drawBoss(s *engine.Screen) {
	sp := engine.NewSprite(
		"    ------------",
		"  --############--",
		"  ##################",
		"  ####@@######@@####",
		"  ####@@######@@####",
		"  ##################",
		"  ###%%%%%%%%%%%%###",
		"  ##################",
	)
	sp.Palette = map[rune]int{
		'#': 129, // 魔躯品红紫
		'-': 88,  // 深红边缘
		'@': 231, // 白瞳
		'%': 226, // 金黄獠牙
	}
	sp.Draw(s, 2, 3)
}
