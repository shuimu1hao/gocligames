// 修仙地牢 - 用 gocligames 写的迷你 RPG demo
// 玩法：WASD 走地牢，碰到怪物进入回合制战斗，
//
//	按 E 与 NPC 对话/开宝箱，1/2 服丹药回血回蓝，
//	走下楼梯深入，击败幽冥魔尊救出大师兄即通关。
//
// 存档：自动存到当前目录 xiuxian_save.json（换层/拾取/战斗后都存）。
package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"

	"gocligames/engine"
)

const (
	scrW, scrH     = 44, 20 // 画布（加宽：左侧地图 + 右侧信息区）
	mapXOff        = 2      // 地图水平偏移（20 列地图靠左）
	mapOffY        = 2      // 地图在画布中的起始行
	helpY1         = 14     // 按键说明行（下方）
	helpY2         = 15     // 按键说明行 2
	infoX          = 24     // 右侧信息区（对话/交互/战斗信息）起始列
	infoY          = 2      // 右侧信息区起始行
	startX, startY = 2, 2   // 每层出生点
)

// 常驻按键说明
const helpLine1 = "WASD移动 E交互(对话/开箱)"
const helpLine2 = "1回血 2回蓝 P帮助 R重开 Q退出"

// Player 玩家状态
type Player struct {
	Level, XP int
	HP, MaxHP int
	MP, MaxMP int
	Atk, Def  int
	Ling      int // 灵石
	Pills     int // 回元丹
	ManaPills int // 聚灵丹
}

// Fight 战斗临时状态
type Fight struct {
	o   *Obj
	hp  int
	msg string
}

// Game 游戏主状态
type Game struct {
	g        *engine.Game
	layer    int
	px, py   int
	objs     []*Obj
	mode     string // map talk fight over win
	talk     []string
	talkI    int
	fight    *Fight
	msg      []string
	pl       Player
	savePath string
	dead     map[string]bool // 已杀怪（存档）
	opened   map[string]bool // 已拾取/已开箱（存档）
	talked   map[string]bool // 已对话（存档）
	bossDead bool
	won      bool
}

// ---------- 小工具 ----------

func keyOf(layer, x, y int) string { return fmt.Sprintf("%d-%d-%d", layer, x, y) }

func cut(s string, max int) string {
	r := []rune(s)
	if len(r) > max {
		return string(r[:max])
	}
	return s
}

func (gm *Game) push(m string) {
	gm.msg = append(gm.msg, m)
	if len(gm.msg) > 6 {
		gm.msg = gm.msg[1:]
	}
}

// wrapText 按显示宽度折行（1 rune = 1 中文字 = 2 终端列）。
func wrapText(s string, width int) []string {
	r := []rune(s)
	var out []string
	for len(r) > 0 {
		n := width
		if len(r) < n {
			n = len(r)
		}
		out = append(out, string(r[:n]))
		r = r[n:]
	}
	return out
}

// ---------- 地图 ----------

func (gm *Game) tileAt(x, y int) rune {
	l := &layers[gm.layer]
	if x < 0 || y < 0 || x >= l.W || y >= l.H {
		return '#'
	}
	return l.Tiles[y][x]
}

func (gm *Game) objAt(x, y int) *Obj {
	for _, o := range gm.objs {
		if o.Dead || o.Opened {
			continue
		}
		if x >= o.X && x < o.X+o.W && y >= o.Y && y < o.Y+o.H {
			return o
		}
	}
	return nil
}

func (gm *Game) adjacent(o *Obj) bool {
	return (o.X-gm.px)*(o.X-gm.px)+(o.Y-gm.py)*(o.Y-gm.py) <= 2
}

// spawnLayer 按当前层重新生成对象（读档时过滤掉已死/已拾取）
func (gm *Game) spawnLayer() {
	gm.objs = nil
	for _, sp := range layerSpawns[gm.layer] {
		o := &Obj{X: sp.x, Y: sp.y, Kind: sp.kind, DefIdx: sp.defIdx, Item: sp.item, W: sp.w, H: sp.h}
		k := keyOf(gm.layer, sp.x, sp.y)
		switch sp.kind {
		case ObjMonster:
			d := monsterDefs[sp.defIdx]
			o.Ch, o.Fg = d.Ch, d.Fg
			if gm.dead[k] {
				o.Dead = true
			}
		case ObjNPC:
			if sp.defIdx == 1 {
				o.Ch, o.Fg = '老', engine.ColorWhite
			} else {
				o.Ch, o.Fg = '师', engine.ColorYellow
			}
		case ObjItem:
			switch sp.item {
			case ItemPill:
				o.Ch = '丹'
			case ItemManaPill:
				o.Ch = '丹'
			case ItemLing:
				o.Ch = '石'
			case ItemBox:
				o.Ch = '箱'
			}
			if gm.opened[k] {
				o.Opened = true
			}
		}
		o.Fg, o.Bg = objColor(o)
		gm.objs = append(gm.objs, o)
	}
}

func (gm *Game) gotoLayer(n int) {
	if n < 0 || n >= len(layers) {
		return
	}
	gm.layer = n
	gm.spawnLayer()
	gm.px, gm.py = startX, startY
	gm.save()
	gm.push("—— " + layers[n].Name + " ——")
}

// ---------- 移动 / 交互 ----------

var dirs = map[string][2]int{
	"w": {0, -1}, "up": {0, -1}, "s": {0, 1}, "down": {0, 1},
	"a": {-1, 0}, "left": {-1, 0}, "d": {1, 0}, "right": {1, 0},
}

func (gm *Game) tryMove(dx, dy int) {
	nx, ny := gm.px+dx, gm.py+dy
	if gm.tileAt(nx, ny) == '#' {
		return
	}
	if o := gm.objAt(nx, ny); o != nil && o.Kind == ObjMonster {
		gm.startFight(o)
		return
	}
	gm.px, gm.py = nx, ny
	if o := gm.objAt(gm.px, gm.py); o != nil && o.Kind == ObjItem {
		gm.pickup(o)
	}
	switch gm.tileAt(gm.px, gm.py) {
	case '>':
		gm.gotoLayer(gm.layer + 1)
	case '<':
		gm.gotoLayer(gm.layer - 1)
	}
}

func (gm *Game) interact() {
	for _, o := range gm.objs {
		if o.Dead || o.Opened {
			continue
		}
		if o.Kind == ObjNPC && gm.adjacent(o) {
			gm.startTalk(o)
			return
		}
		if o.Kind == ObjItem && o.Item == ItemBox && gm.adjacent(o) {
			gm.openBox(o)
			return
		}
	}
	gm.push("四下无人，只有风声……")
}

func (gm *Game) startTalk(o *Obj) {
	k := keyOf(gm.layer, o.X, o.Y)
	if o.DefIdx == 1 { // 守墓老者
		if !gm.talked[k] {
			gm.talk = talkOldMan[0]
			gm.pl.Pills++
			gm.talked[k] = true
			gm.save()
		} else {
			gm.talk = talkOldMan[1]
		}
	} else { // 大师兄
		if gm.bossDead {
			gm.talk = talkBrotherAfter[0]
			gm.won = true
		} else {
			gm.talk = talkBrotherBefore[0]
		}
	}
	gm.talkI = 0
	gm.mode = "talk"
}

func (gm *Game) openBox(o *Obj) {
	o.Opened = true
	gm.opened[keyOf(gm.layer, o.X, o.Y)] = true
	r := rand.Intn(100)
	switch {
	case r < 45:
		gm.pl.Pills++
		gm.push("宝箱开出一颗回元丹！")
	case r < 70:
		gm.pl.ManaPills++
		gm.push("宝箱开出一颗聚灵丹！")
	default:
		got := 2 + rand.Intn(3)
		gm.pl.Ling += got
		gm.push(fmt.Sprintf("宝箱里是 %d 颗灵石！", got))
	}
	gm.save()
}

func (gm *Game) pickup(o *Obj) {
	k := keyOf(gm.layer, o.X, o.Y)
	switch o.Item {
	case ItemPill:
		gm.pl.Pills++
		gm.push("拾取回元丹 ×1")
	case ItemManaPill:
		gm.pl.ManaPills++
		gm.push("拾取聚灵丹 ×1")
	case ItemLing:
		got := 1 + rand.Intn(3)
		gm.pl.Ling += got
		gm.push(fmt.Sprintf("拾取灵石 ×%d", got))
	}
	o.Dead = true
	gm.opened[k] = true
	gm.save()
}

// ---------- 战斗 ----------

func (gm *Game) startFight(o *Obj) {
	d := monsterDefs[o.DefIdx]
	gm.fight = &Fight{o: o, hp: d.HP}
	gm.mode = "fight"
	gm.push("遭遇 " + d.Name + "！")
}

func (gm *Game) fightAttack() {
	d := monsterDefs[gm.fight.o.DefIdx]
	dmg := gm.pl.Atk - d.Def + rand.Intn(4)
	if dmg < 1 {
		dmg = 1
	}
	gm.fight.hp -= dmg
	gm.fight.msg = fmt.Sprintf("青云剑诀！造成 %d 点伤害", dmg)
	if gm.fight.hp <= 0 {
		gm.fightWin()
		return
	}
	gm.enemyTurn()
}

func (gm *Game) fightPill() {
	if gm.pl.Pills <= 0 {
		gm.fight.msg = "回元丹已用完！"
		return
	}
	gm.pl.Pills--
	gm.pl.HP += 30
	if gm.pl.HP > gm.pl.MaxHP {
		gm.pl.HP = gm.pl.MaxHP
	}
	gm.fight.msg = "吞下回元丹，恢复 30 点气血"
	gm.enemyTurn()
}

func (gm *Game) fightRun() {
	if rand.Intn(100) < 70 {
		gm.fight = nil
		gm.mode = "map"
		gm.push("你撤身退出战斗！")
	} else {
		gm.fight.msg = "逃跑失败！"
		gm.enemyTurn()
	}
}

func (gm *Game) enemyTurn() {
	d := monsterDefs[gm.fight.o.DefIdx]
	dmg := d.Atk - gm.pl.Def + rand.Intn(3)
	if dmg < 1 {
		dmg = 1
	}
	gm.pl.HP -= dmg
	gm.fight.msg = fmt.Sprintf("%s反击，你受 %d 点伤害", d.Name, dmg)
	if gm.pl.HP <= 0 {
		gm.pl.HP = 0
		gm.mode = "over"
		gm.push("你倒下了……万骨窟又添一具白骨")
	}
}

func (gm *Game) fightWin() {
	d := monsterDefs[gm.fight.o.DefIdx]
	gm.fight.o.Dead = true
	gm.dead[keyOf(gm.layer, gm.fight.o.X, gm.fight.o.Y)] = true
	got := d.MinLing + rand.Intn(d.MaxLing-d.MinLing+1)
	gm.pl.Ling += got
	gm.pl.XP += d.XP
	gm.fight.msg = fmt.Sprintf("击破%s！+%d灵石 +%d经验", d.Name, got, d.XP)
	if rand.Float64() < d.DropPill {
		gm.pl.Pills++
		gm.fight.msg += " 掉出回元丹"
	}
	if d.DropMana > 0 && rand.Float64() < d.DropMana {
		gm.pl.ManaPills++
		gm.fight.msg += " 掉出聚灵丹"
	}
	if gm.fight.o.DefIdx == 3 {
		gm.bossDead = true
	}
	gm.mode = "map"
	gm.checkLevelUp()
	gm.save()
}

func (gm *Game) xpToNext() int { return 20 + (gm.pl.Level-1)*15 }

func (gm *Game) checkLevelUp() {
	for gm.pl.XP >= gm.xpToNext() {
		gm.pl.XP -= gm.xpToNext()
		gm.pl.Level++
		gm.pl.MaxHP += 15
		gm.pl.MaxMP += 5
		gm.pl.Atk += 2
		gm.pl.Def++
		gm.pl.HP = gm.pl.MaxHP
		gm.pl.MP = gm.pl.MaxMP
		gm.push(fmt.Sprintf("道行精进！突破至练气 %d 层！", gm.pl.Level))
	}
}

// ---------- 丹药（地图外） ----------

func (gm *Game) usePill(k ItemKind) {
	if k == ItemPill {
		if gm.pl.Pills <= 0 {
			gm.push("回元丹已用完")
			return
		}
		if gm.pl.HP >= gm.pl.MaxHP {
			gm.push("气血已满，无需服药")
			return
		}
		gm.pl.Pills--
		gm.pl.HP += 30
		if gm.pl.HP > gm.pl.MaxHP {
			gm.pl.HP = gm.pl.MaxHP
		}
		gm.push("服下回元丹，恢复 30 点气血")
	} else {
		if gm.pl.ManaPills <= 0 {
			gm.push("聚灵丹已用完")
			return
		}
		if gm.pl.MP >= gm.pl.MaxMP {
			gm.push("灵力已满，无需服药")
			return
		}
		gm.pl.ManaPills--
		gm.pl.MP += 15
		if gm.pl.MP > gm.pl.MaxMP {
			gm.pl.MP = gm.pl.MaxMP
		}
		gm.push("服下聚灵丹，恢复 15 点灵力")
	}
	gm.save()
}

// ---------- 存档 ----------

// SaveData 存档数据（JSON 序列化）。
type SaveData struct {
	Layer    int
	Pl       Player
	Dead     map[string]bool
	Opened   map[string]bool
	Talked   map[string]bool
	BossDead bool
}

func (gm *Game) save() {
	sd := SaveData{Layer: gm.layer, Pl: gm.pl, Dead: gm.dead, Opened: gm.opened,
		Talked: gm.talked, BossDead: gm.bossDead}
	data, err := json.Marshal(sd)
	if err != nil {
		return
	}
	_ = os.WriteFile(gm.savePath, data, 0o644)
}

func (gm *Game) load() bool {
	data, err := os.ReadFile(gm.savePath)
	if err != nil {
		return false
	}
	var sd SaveData
	if json.Unmarshal(data, &sd) != nil {
		return false
	}
	gm.layer = sd.Layer
	gm.pl = sd.Pl
	gm.dead = sd.Dead
	gm.opened = sd.Opened
	gm.talked = sd.Talked
	gm.bossDead = sd.BossDead
	if gm.dead == nil {
		gm.dead = map[string]bool{}
	}
	if gm.opened == nil {
		gm.opened = map[string]bool{}
	}
	if gm.talked == nil {
		gm.talked = map[string]bool{}
	}
	return true
}

func (gm *Game) newGame() {
	gm.layer = 0
	gm.pl = Player{Level: 1, XP: 0, HP: 50, MaxHP: 50, MP: 20, MaxMP: 20,
		Atk: 8, Def: 2, Ling: 0, Pills: 1, ManaPills: 0}
	gm.dead = map[string]bool{}
	gm.opened = map[string]bool{}
	gm.talked = map[string]bool{}
	gm.bossDead = false
	gm.won = false
	gm.mode = "map"
	gm.spawnLayer()
	gm.px, gm.py = startX, startY
	gm.msg = nil
	gm.push("你踏入万骨窟，阴风扑面而来……")
}

// ---------- 引擎钩子 ----------

func (gm *Game) onStart(g *engine.Game) {
	gm.dead, gm.opened, gm.talked = map[string]bool{}, map[string]bool{}, map[string]bool{}
	if gm.load() {
		gm.mode = "map"
		gm.spawnLayer()
		gm.px, gm.py = startX, startY
		gm.push(fmt.Sprintf("（读档）回到第 %d 层：%s", gm.layer+1, layers[gm.layer].Name))
	} else {
		gm.newGame()
		gm.push("（新旅程）师父的话还在耳边：找到你大师兄！")
	}
}

func (gm *Game) onKey(g *engine.Game, key string) {
	switch gm.mode {
	case "map":
		if d, ok := dirs[key]; ok {
			gm.tryMove(d[0], d[1])
			return
		}
		switch key {
		case "e":
			gm.interact()
		case "1":
			gm.usePill(ItemPill)
		case "2":
			gm.usePill(ItemManaPill)
		case "p":
			gm.push("WASD移动 E交互 1回血 2回蓝 R重开 Q退出")
		case "r":
			gm.newGame()
			gm.save()
		case "q":
			gm.save()
			g.Quit()
		}
	case "talk":
		if key == "enter" || key == "space" || key == "e" {
			gm.talkI++
			if gm.talkI >= len(gm.talk) {
				gm.mode = "map"
				if gm.won {
					gm.mode = "win"
				}
			}
		} else if key == "q" {
			gm.mode = "map"
			if gm.won {
				gm.mode = "win"
			}
		}
	case "fight":
		switch key {
		case "1":
			gm.fightAttack()
		case "2":
			gm.fightPill()
		case "3", "q":
			gm.fightRun()
		}
	case "over", "win":
		switch key {
		case "r":
			gm.newGame()
			gm.save()
		case "q":
			g.Quit()
		}
	}
}

// ---------- 渲染 ----------

func (gm *Game) render(g *engine.Game, s *engine.Screen) {
	s.Clear()
	gm.renderStatus(s)
	gm.renderMap(s)
	gm.renderHelp(s)
	gm.renderMsg(s)
	switch gm.mode {
	case "talk":
		gm.renderTalk(s)
	case "fight":
		gm.renderFight(s)
	case "over":
		gm.renderOver(s)
	case "win":
		gm.renderWin(s)
	}
}

func (gm *Game) renderStatus(s *engine.Screen) {
	p := &gm.pl
	s.Text(1, 0, fmt.Sprintf("青云 Lv.%d 石%d 丹%d 丹%d", p.Level, p.Ling, p.Pills, p.ManaPills), engine.ColorCyan)
	s.Text(1, 1, fmt.Sprintf("HP %d/%d MP %d/%d 攻%d 防%d", p.HP, p.MaxHP, p.MP, p.MaxMP, p.Atk, p.Def), engine.ColorWhite)
}

func (gm *Game) renderMap(s *engine.Screen) {
	l := &layers[gm.layer]
	for y := 0; y < l.H; y++ {
		for x := 0; x < l.W; x++ {
			t := l.Tiles[y][x]
			fg := engine.ColorGray
			switch t {
			case '#':
				fg = engine.ColorDarkGray
			case '>', '<':
				fg = engine.ColorCyan
			}
			s.Set(mapXOff+x, mapOffY+y, t, fg, -1)
		}
	}
	for _, o := range gm.objs {
		if o.Dead || o.Opened {
			continue
		}
		for j := 0; j < o.H; j++ {
			for i := 0; i < o.W; i++ {
				s.Set(mapXOff+o.X+i, mapOffY+o.Y+j, o.Ch, o.Fg, o.Bg)
			}
		}
	}
	s.Set(mapXOff+gm.px, mapOffY+gm.py, '@', engine.ColorCyan, bgDeepBlue)
}

func (gm *Game) renderHelp(s *engine.Screen) {
	// 战斗时底部显示战斗操作，其余显示常驻说明
	if gm.mode == "fight" {
		s.Text(1, helpY1, "1 攻击  2 回元丹  3 逃跑", engine.ColorYellow)
		s.Text(1, helpY2, "Q 逃跑", engine.ColorGray)
		return
	}
	s.Text(1, helpY1, helpLine1, engine.ColorGray)
	s.Text(1, helpY2, helpLine2, engine.ColorGray)
}

func (gm *Game) renderMsg(s *engine.Screen) {
	// 右侧信息区：最近 6 条交互消息
	for i, m := range gm.msg {
		s.Text(infoX, infoY+i, cut(m, 10), engine.ColorGray)
	}
}

func (gm *Game) renderTalk(s *engine.Screen) {
	// 右侧：对话内容折行显示
	lines := wrapText(gm.talk[gm.talkI], 10)
	for i, ln := range lines {
		s.Text(infoX, infoY+i, ln, engine.ColorYellow)
	}
	ty := infoY + len(lines) + 1
	if ty > 12 {
		ty = 12
	}
	s.Text(infoX, ty, "Enter 继续  Q 跳过", engine.ColorGray)
}

func (gm *Game) renderFight(s *engine.Screen) {
	d := monsterDefs[gm.fight.o.DefIdx]
	// 右侧信息区（x=infoX 起，20 列）
	s.Text(infoX, 2, "=== 战斗 ===", engine.ColorRed)
	s.Text(infoX, 3, d.Name, d.Fg)
	pct := float64(gm.fight.hp) / float64(d.HP)
	bar := int(pct * 10)
	if bar < 0 {
		bar = 0
	}
	hpbar := ""
	for i := 0; i < 10; i++ {
		if i < bar {
			hpbar += "#"
		} else {
			hpbar += "."
		}
	}
	s.Text(infoX, 4, hpbar, engine.ColorRed)
	s.Text(infoX, 5, fmt.Sprintf("HP %d/%d", gm.fight.hp, d.HP), engine.ColorRed)
	s.Text(infoX, 6, fmt.Sprintf("你 %d/%d", gm.pl.HP, gm.pl.MaxHP), engine.ColorWhite)
	s.Text(infoX, 7, fmt.Sprintf("攻%d 防%d", gm.pl.Atk, gm.pl.Def), engine.ColorWhite)
	if gm.fight.o.DefIdx == 3 {
		s.Text(infoX, 8, fmt.Sprintf("丹%d", gm.pl.Pills), engine.ColorYellow)
		// Boss：精细像素画覆盖地图区
		drawBoss(s)
	} else {
		s.Text(infoX, 8, fmt.Sprintf("丹%d", gm.pl.Pills), engine.ColorYellow)
	}
	if gm.fight.msg != "" {
		s.Text(infoX, 10, cut(gm.fight.msg, 10), engine.ColorGray)
	}
	// 操作提示在底部（renderHelp 已切换）
}

func (gm *Game) renderOver(s *engine.Screen) {
	s.Text(5, 6, "你倒在了万骨窟深处……", engine.ColorRed)
	s.Text(5, 8, fmt.Sprintf("灵石 %d  等级 %d", gm.pl.Ling, gm.pl.Level), engine.ColorWhite)
	s.Text(5, 10, "R 重入轮回  Q 归去", engine.ColorGray)
}

func (gm *Game) renderWin(s *engine.Screen) {
	s.Text(5, 5, "幽冥魔尊已灭！", engine.ColorYellow)
	s.Text(5, 7, "大师兄得救，万骨窟重归封印。", engine.ColorGreen)
	s.Text(5, 8, fmt.Sprintf("你带回 %d 颗灵石", gm.pl.Ling), engine.ColorWhite)
	s.Text(5, 10, "R 再闯一遭  Q 归去", engine.ColorGray)
}

func main() {
	dir, _ := os.Getwd()
	g := engine.NewGame("修仙地牢 demo", scrW, scrH, 30)
	gm := &Game{g: g, savePath: filepath.Join(dir, "xiuxian_save.json")}
	g.OnStart = gm.onStart
	g.OnKey = gm.onKey
	g.Render = gm.render
	g.Run()
}
