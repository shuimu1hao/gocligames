// 江湖行 - 区域地图与对象生成
// 规则：tiles 一律纯 ASCII 单宽等宽网格（中文只放对象层，避免错位）
// 地形：#墙 .路 T树 ~水 ^雪 >传送(进入) <传送(返回) ?奇遇
package main

import "gocligames/engine"

// Region 一个区域：tiles 是等宽 ASCII 网格
type Region struct {
	Name  string
	Tiles [][]rune
	W, H  int
}

// SpawnKind 对象类型
type SpawnKind string

const (
	SpawnMonster SpawnKind = "monster"
	SpawnNPC     SpawnKind = "npc"
	SpawnChest   SpawnKind = "chest"
)

// Spawn 区域对象出生点（存档键 = key）
type Spawn struct {
	Kind  SpawnKind
	X, Y  int
	MIdx  int      // 怪物索引（monster）
	NPC   string   // NPC 名（npc）
	Ch    rune     // NPC 显示字符
	Fg    int      // NPC 颜色
	Item  ItemKind // 宝箱物品（chest，可空）
	Money int      // 宝箱金钱（chest，可 0）
	Key   string   // 唯一键
}

// Obj 运行时对象
type Obj struct {
	Kind   SpawnKind
	X, Y   int
	Ch     rune
	Fg     int
	MIdx   int
	NPC    string
	Item   ItemKind
	Money  int
	Key    string
	Dead   bool // 怪物已杀
	Opened bool // 宝箱已开
}

func mkTiles(rows []string) [][]rune {
	t := make([][]rune, len(rows))
	for i, r := range rows {
		t[i] = []rune(r)
	}
	return t
}

func init() {
	for i := range regions {
		regions[i].H = len(regions[i].Tiles)
		if regions[i].H > 0 {
			regions[i].W = len(regions[i].Tiles[0])
		}
	}
}

// regionOrder 遍历顺序
var regionOrder = []string{"world", "qingzhou", "luoyang", "fort", "tomb", "peak"}

// regions 区域表（名字 -> 区域）
var regions = map[string]*Region{
	"world": {
		Name: "中原大地",
		Tiles: mkTiles([]string{
			"####################",
			"#..T....T....T....T#",
			"#..T....T....>....T#",
			"#..T....T....T....T#",
			"#..T....>..T....>..#",
			"#..T....T....T....T#",
			"#..T....T....T....T#",
			"#..T....T....>....T#",
			"#..T....T....T....T#",
			"#..T....T....>....T#",
			"#..?....?....?....?#",
			"####################",
		}),
	},
	"qingzhou": {
		Name: "青州城",
		Tiles: mkTiles([]string{
			"####################",
			"#..................#",
			"#..##......##......#",
			"#..................#",
			"#..##......##......#",
			"#..................#",
			"#...............<..#",
			"#..................#",
			"#..##......##......#",
			"#..................#",
			"#..##......##......#",
			"####################",
		}),
	},
	"luoyang": {
		Name: "洛阳城",
		Tiles: mkTiles([]string{
			"####################",
			"#..................#",
			"#..##..##..##..##..#",
			"#..................#",
			"#..##..##..##..##..#",
			"#..................#",
			"#...............<..#",
			"#..................#",
			"#..##..##..##......#",
			"#..................#",
			"#..##..##..##......#",
			"####################",
		}),
	},
	"fort": {
		Name: "黑风寨",
		Tiles: mkTiles([]string{
			"####################",
			"#....####....####..#",
			"#....#..####..#....#",
			"#....#........#....#",
			"#..##..........##..#",
			"#.................<#",
			"#..##....##....##..#",
			"#....#........#....#",
			"#....#..####..#....#",
			"#....####....####..#",
			"#..................#",
			"####################",
		}),
	},
	"tomb": {
		Name: "幽暗古墓",
		Tiles: mkTiles([]string{
			"####################",
			"#....###....###....#",
			"#....#.#....#.#....#",
			"#....#.#......#....#",
			"#....###....###....#",
			"#......####........#",
			"#..................#",
			"#..##......##......#",
			"#..#........#......#",
			"#..#........#......#",
			"#..##......##...<..#",
			"####################",
		}),
	},
	"peak": {
		Name: "天山顶",
		Tiles: mkTiles([]string{
			"####################",
			"#..^^..^^..^^..^^..#",
			"#..^^..^^..^^..^^..#",
			"#..^^........^^....#",
			"#..................#",
			"#..^^..^^..^^..^^..#",
			"#..................#",
			"#..^^..^^..^^..^^..#",
			"#..^^..^^..^^..^^..#",
			"#..^^..^^..^^..^^..#",
			"#................<.#",
			"####################",
		}),
	},
}

// regionSpawns 各区域对象出生表
var regionSpawns = map[string][]Spawn{
	"world": {
		{SpawnNPC, 2, 2, 0, "方丈玄慈", '方', engine.ColorYellow, "", 0, "npc_shaolin"},
		{SpawnNPC, 10, 5, 0, "掌门冲虚", '掌', engine.ColorCyan, "", 0, "npc_wudang"},
		{SpawnNPC, 17, 8, 0, "帮主洪七公", '帮', engine.ColorOrange, "", 0, "npc_gaibang"},
		{SpawnMonster, 4, 1, 0, "", 0, 0, "", 0, "m_wild_dog1"},
		{SpawnMonster, 15, 1, 1, "", 0, 0, "", 0, "m_wild_wolf1"},
		{SpawnMonster, 9, 3, 2, "", 0, 0, "", 0, "m_bandit1"},
		{SpawnMonster, 17, 3, 3, "", 0, 0, "", 0, "m_snake1"},
		{SpawnMonster, 4, 6, 0, "", 0, 0, "", 0, "m_wild_dog2"},
		{SpawnMonster, 15, 6, 1, "", 0, 0, "", 0, "m_wild_wolf2"},
		{SpawnMonster, 4, 8, 2, "", 0, 0, "", 0, "m_bandit2"},
		{SpawnMonster, 15, 8, 3, "", 0, 0, "", 0, "m_snake2"},
		{SpawnMonster, 7, 10, 2, "", 0, 0, "", 0, "m_bandit3"},
	},
	"qingzhou": {
		{SpawnNPC, 2, 2, 0, "药铺掌柜", '药', engine.ColorGreen, "", 0, "npc_doctor"},
		{SpawnNPC, 10, 2, 0, "师父李长风", '师', engine.ColorWhite, "", 0, "npc_master"},
		{SpawnNPC, 2, 4, 0, "客栈小二", '栈', engine.ColorYellow, "", 0, "npc_inn_qz"},
		{SpawnNPC, 10, 4, 0, "武馆教头", '武', engine.ColorCyan, "", 0, "npc_wuguan"},
		{SpawnNPC, 2, 8, 0, "镖局镖头", '镖', engine.ColorOrange, "", 0, "npc_biaotou"},
		{SpawnChest, 9, 1, 0, "", 0, 0, ItemGold, 20, "c_qz_1"},
	},
	"luoyang": {
		{SpawnNPC, 2, 2, 0, "铁匠", '匠', engine.ColorOrange, "", 0, "npc_smith"},
		{SpawnNPC, 6, 2, 0, "赌场老板", '赌', engine.ColorMagenta, "", 0, "npc_gamble"},
		{SpawnNPC, 10, 2, 0, "当铺朝奉", '当', engine.ColorGray, "", 0, "npc_pawn"},
		{SpawnNPC, 2, 4, 0, "丐帮弟子", '丐', engine.ColorYellow, "", 0, "npc_gaizi"},
		{SpawnNPC, 6, 4, 0, "酒馆老板", '酒', engine.ColorGreen, "", 0, "npc_jiuguan"},
		{SpawnNPC, 10, 4, 0, "捕头", '捕', engine.ColorBlue, "", 0, "npc_butou"},
		{SpawnNPC, 2, 8, 0, "客栈小二", '栈', engine.ColorYellow, "", 0, "npc_inn_ly"},
		{SpawnChest, 16, 1, 0, "", 0, 0, ItemMana, 0, "c_ly_1"},
	},
	"fort": {
		{SpawnMonster, 4, 2, 4, "", 0, 0, "", 0, "m_fort_1"},
		{SpawnMonster, 13, 2, 4, "", 0, 0, "", 0, "m_fort_2"},
		{SpawnMonster, 4, 8, 4, "", 0, 0, "", 0, "m_fort_3"},
		{SpawnMonster, 11, 6, 5, "", 0, 0, "", 0, "boss_zhai"},
		{SpawnChest, 17, 3, 0, "", 0, 0, ItemLetter, 0, "c_fort_letter"},
		{SpawnChest, 3, 9, 0, "", 0, 0, ItemGreat, 0, "c_fort_1"},
	},
	"tomb": {
		{SpawnMonster, 4, 2, 6, "", 0, 0, "", 0, "m_tomb_1"},
		{SpawnMonster, 15, 2, 6, "", 0, 0, "", 0, "m_tomb_2"},
		{SpawnMonster, 5, 9, 6, "", 0, 0, "", 0, "m_tomb_3"},
		{SpawnMonster, 10, 3, 7, "", 0, 0, "", 0, "boss_shi"},
		{SpawnChest, 16, 8, 0, "", 0, 0, ItemJade, 0, "c_tomb_jade"},
		{SpawnChest, 10, 9, 0, "", 0, 0, ItemScroll, 0, "c_tomb_scroll"},
	},
	"peak": {
		{SpawnMonster, 4, 2, 8, "", 0, 0, "", 0, "m_peak_1"},
		{SpawnMonster, 14, 2, 8, "", 0, 0, "", 0, "m_peak_2"},
		{SpawnMonster, 9, 4, 9, "", 0, 0, "", 0, "m_peak_guard"},
		{SpawnMonster, 7, 5, 10, "", 0, 0, "", 0, "boss_jiao"},
		{SpawnChest, 15, 8, 0, "", 0, 0, ItemToken, 0, "c_peak_token"},
		{SpawnChest, 3, 3, 0, "", 0, 0, ItemGreat, 0, "c_peak_1"},
	},
}

// portals 传送表：踩到某区域某格 -> 目标区域出生点
type Portal struct {
	From string
	X, Y int
	To   string
	ToX  int
	ToY  int
}

var portals = []Portal{
	{"world", 13, 2, "qingzhou", 2, 3},
	{"world", 8, 4, "fort", 2, 4},
	{"world", 16, 4, "tomb", 2, 4},
	{"world", 13, 7, "luoyang", 2, 3},
	{"world", 13, 9, "peak", 2, 4},
	{"qingzhou", 16, 6, "world", 12, 3},
	{"luoyang", 16, 6, "world", 12, 8},
	{"fort", 18, 5, "world", 7, 5},
	{"tomb", 16, 10, "world", 16, 5},
	{"peak", 17, 10, "world", 12, 10},
}

func portalAt(region string, x, y int) *Portal {
	for i := range portals {
		if portals[i].From == region && portals[i].X == x && portals[i].Y == y {
			return &portals[i]
		}
	}
	return nil
}

// spawnRegion 生成某区域对象（读档时过滤已死/已开）
func (gm *Game) spawnRegion() {
	gm.objs = nil
	for _, sp := range regionSpawns[gm.region] {
		o := &Obj{Kind: sp.Kind, X: sp.X, Y: sp.Y, MIdx: sp.MIdx, NPC: sp.NPC,
			Item: sp.Item, Money: sp.Money, Key: sp.Key}
		switch sp.Kind {
		case SpawnMonster:
			d := monsterDefs[sp.MIdx]
			o.Ch, o.Fg = d.Ch, d.Fg
			if gm.dead[sp.Key] {
				o.Dead = true
			}
		case SpawnNPC:
			o.Ch, o.Fg = sp.Ch, sp.Fg
		case SpawnChest:
			o.Ch, o.Fg = '箱', engine.ColorYellow
			if gm.opened[sp.Key] {
				o.Opened = true
			}
		}
		gm.objs = append(gm.objs, o)
	}
	// 通关后：血月余孽出没中原（冲战力榜的新目标）
	if gm.region == "world" && gm.won {
		yunie := []Spawn{
			{SpawnMonster, 5, 5, 11, "", 0, 0, "", 0, "m_after_1"},
			{SpawnMonster, 13, 5, 11, "", 0, 0, "", 0, "m_after_2"},
			{SpawnMonster, 9, 9, 11, "", 0, 0, "", 0, "m_after_3"},
		}
		for _, sp := range yunie {
			if gm.dead[sp.Key] {
				continue
			}
			d := monsterDefs[sp.MIdx]
			gm.objs = append(gm.objs, &Obj{Kind: SpawnMonster, X: sp.X, Y: sp.Y,
				Ch: d.Ch, Fg: d.Fg, MIdx: sp.MIdx, Key: sp.Key})
		}
	}
}
