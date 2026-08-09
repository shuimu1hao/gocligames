// 江湖行 - 武侠开放世界 RPG（gocligames 引擎）
// 玩法：WASD 走江湖，E 交互，I 背包 F 武功 T 任务 B 江湖榜 H 帮助
// 主线：黑风寨取密信 → 古墓寻天蚕神功 → 天山顶决战血月教主
// 存档：自动存 jianghu_save.json（换区/战斗/拾取/对话后）
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gocligames/engine"
)

const (
	scrW, scrH = 44, 20
	mapXOff    = 2 // 地图水平偏移
	mapOffY    = 2 // 地图起始行
	infoX      = 24
	infoY      = 2
	msgY1      = 14
	msgY2      = 15
	helpY1     = 16
	helpY2     = 17
)

// 常驻按键说明
const helpLine1 = "WASD移动 E交互 I背包 F武功 T任务"
const helpLine2 = "B江湖榜 H帮助 1金疮药 2大还丹 Q退出"

// ---------- 玩家 ----------

// Player 玩家状态（属性/装备/武功/任务进度）。
type Player struct {
	Name      string
	Level     int
	XP        int
	HP        int
	MaxHP     int
	MP        int
	MaxMP     int
	Atk       int
	Def       int
	Money     int
	Weapon    string
	Armor     string
	Accessory string
	Inventory []ItemKind
	Martials  []string
	Sect      string
	Realm     int
	Poisoned  int
}

// Fight 战斗临时状态
type Fight struct {
	mi     int // 怪物索引
	hp     int
	msg    string
	poison int // 怪物中毒剩余回合
	turns  int
}

// Game 主状态
type Game struct {
	g           *engine.Game
	region      string
	px, py      int
	objs        []*Obj
	mode        string // map talk fight menu shop inn casino enc dead win
	talk        []string
	talkI       int
	onTalkEnd   func()
	fight       *Fight
	menuTab     int
	menuIdx     int
	shopID      string
	innName     string
	innCost     int
	casinoState int
	encID       string
	encOpts     []string
	msg         []string
	pl          Player
	savePath    string
	dead        map[string]bool
	opened      map[string]bool
	flags       map[string]bool
	mainProg    int
	quests      map[int]bool
	rankIn      int
	won         bool
	sb          *engine.Scoreboard
}

// ---------- 工具 ----------

func keyOf(region string, x, y int) string { return fmt.Sprintf("%s-%d-%d", region, x, y) }

func cut(s string, max int) string {
	r := []rune(s)
	if len(r) > max {
		return string(r[:max])
	}
	return s
}

func (gm *Game) push(m string) {
	gm.msg = append(gm.msg, m)
	if len(gm.msg) > 2 {
		gm.msg = gm.msg[1:]
	}
}

// wrapText 按字符数折行（信息区按中文宽度显示）
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

func (gm *Game) curRegion() *Region { return regions[gm.region] }

func (gm *Game) tileAt(x, y int) rune {
	r := gm.curRegion()
	if x < 0 || y < 0 || x >= r.W || y >= r.H {
		return '#'
	}
	return r.Tiles[y][x]
}

func (gm *Game) objAt(x, y int) *Obj {
	for _, o := range gm.objs {
		if o.Dead || o.Opened {
			continue
		}
		if o.X == x && o.Y == y {
			return o
		}
	}
	return nil
}

func (gm *Game) adjacent(o *Obj) bool {
	dx := o.X - gm.px
	dy := o.Y - gm.py
	return dx*dx+dy*dy <= 2 && !(dx == 0 && dy == 0)
}

// ---------- 境界 / 战力 ----------

var realmNames = []string{"初入江湖", "三流高手", "二流高手", "一流高手", "宗师", "大宗师", "绝世高手"}

func realmForLevel(lv int) int {
	switch {
	case lv >= 55:
		return 6
	case lv >= 42:
		return 5
	case lv >= 30:
		return 4
	case lv >= 20:
		return 3
	case lv >= 12:
		return 2
	case lv >= 6:
		return 1
	default:
		return 0
	}
}

// equipBonus 装备加成
func (gm *Game) equipBonus() (atk, def, hp, mp int) {
	for _, n := range []string{gm.pl.Weapon, gm.pl.Armor, gm.pl.Accessory} {
		if n == "" {
			continue
		}
		if e := findEquip(n); e != nil {
			atk += e.Atk
			def += e.Def
			hp += e.HP
			mp += e.MP
		}
	}
	return
}

// power 江湖战力（排行榜用）
func (gm *Game) power() int {
	atk, def, hp, mp := gm.equipBonus()
	p := &gm.pl
	base := p.Level*10 + (p.Atk+atk)*3 + (p.Def+def)*2 + (p.MaxHP+hp)/8 + (p.MaxMP+mp)/4
	return base * (p.Realm + 1)
}

// ---------- 升级 ----------

func (gm *Game) xpToNext() int { return 15 + gm.pl.Level*12 }

func (gm *Game) addXP(n int) {
	p := &gm.pl
	p.XP += n
	for p.XP >= gm.xpToNext() {
		p.XP -= gm.xpToNext()
		p.Level++
		p.MaxHP += 12
		p.MaxMP += 4
		p.Atk += 2
		p.Def++
		p.HP = p.MaxHP
		p.MP = p.MaxMP
		newRealm := realmForLevel(p.Level)
		if newRealm > p.Realm {
			p.Realm = newRealm
			gm.push(fmt.Sprintf("境界突破！你已是%s！", realmNames[p.Realm]))
		} else {
			gm.push(fmt.Sprintf("武功精进！升到 %d 级", p.Level))
		}
	}
}

// ---------- 移动 / 交互 ----------

var dirs = map[string][2]int{
	"w": {0, -1}, "up": {0, -1}, "s": {0, 1}, "down": {0, 1},
	"a": {-1, 0}, "left": {-1, 0}, "d": {1, 0}, "right": {1, 0},
}

func (gm *Game) tryMove(dx, dy int) {
	nx, ny := gm.px+dx, gm.py+dy
	if !gm.walkable(nx, ny) {
		return
	}
	if o := gm.objAt(nx, ny); o != nil && o.Kind == SpawnMonster {
		gm.startFight(o.MIdx)
		return
	}
	gm.px, gm.py = nx, ny
	t := gm.tileAt(nx, ny)
	switch t {
	case '>':
		if p := portalAt(gm.region, nx, ny); p != nil {
			gm.enterRegion(p.To, p.ToX, p.ToY)
			return
		}
	case '<':
		if p := portalAt(gm.region, nx, ny); p != nil {
			gm.enterRegion(p.To, p.ToX, p.ToY)
			return
		}
	case '?':
		if !gm.flags["enc_"+keyOf(gm.region, nx, ny)] {
			gm.triggerEncounter(keyOf(gm.region, nx, ny))
		}
	}
	if o := gm.objAt(nx, ny); o != nil && o.Kind == SpawnChest {
		gm.openChest(o)
	}
}

func (gm *Game) walkable(x, y int) bool {
	t := gm.tileAt(x, y)
	switch t {
	case '#', 'T', '~':
		return false
	}
	return true
}

func (gm *Game) enterRegion(name string, x, y int) {
	gm.region = name
	gm.px, gm.py = x, y
	gm.spawnRegion()
	gm.save()
	gm.push("—— " + regions[name].Name + " ——")
}

func (gm *Game) interact() {
	// 优先：相邻 NPC / 宝箱
	for _, o := range gm.objs {
		if o.Dead || o.Opened {
			continue
		}
		if o.Kind == SpawnNPC && gm.adjacent(o) {
			gm.startTalk(o.NPC)
			return
		}
		if o.Kind == SpawnChest && gm.adjacent(o) {
			gm.openChest(o)
			return
		}
	}
	gm.push("四下无人，只有风声……")
}

func (gm *Game) openChest(o *Obj) {
	o.Opened = true
	gm.opened[o.Key] = true
	if o.Item != "" {
		gm.addItem(o.Item)
		gm.push("开启宝箱，获得 " + itemDefs[o.Item].Name + "！")
		if o.Item == ItemScroll {
			gm.learnMartial("天蚕神功")
			gm.push("你参悟了天蚕神功残卷，学会绝世武功！")
		}
		if o.Item == ItemJade && gm.mainProg == 3 {
			gm.mainProg = 4
			gm.push("暖玉入手！师父说洛阳捕头知晓天山之事")
		}
	}
	if o.Money > 0 {
		gm.pl.Money += o.Money
		gm.push(fmt.Sprintf("开启宝箱，获得 %d 两银钱！", o.Money))
	}
	gm.save()
}

// respawnWorld 客栈休息后，中原大地的野怪重新出没（练级资源再生）
func (gm *Game) respawnWorld() {
	for k := range gm.dead {
		if len(k) >= 2 && (k[:2] == "m_") {
			if k == "m_wild_dog1" || k == "m_wild_wolf1" || k == "m_bandit1" || k == "m_snake1" ||
				k == "m_wild_dog2" || k == "m_wild_wolf2" || k == "m_bandit2" || k == "m_snake2" || k == "m_bandit3" ||
				k == "m_after_1" || k == "m_after_2" || k == "m_after_3" {
				delete(gm.dead, k)
			}
		}
	}
	if gm.region == "world" {
		gm.spawnRegion()
	}
}

func (gm *Game) addItem(k ItemKind) {
	gm.pl.Inventory = append(gm.pl.Inventory, k)
}

func (gm *Game) removeItem(k ItemKind) bool {
	for i, it := range gm.pl.Inventory {
		if it == k {
			gm.pl.Inventory = append(gm.pl.Inventory[:i], gm.pl.Inventory[i+1:]...)
			return true
		}
	}
	return false
}

func (gm *Game) countItem(k ItemKind) int {
	n := 0
	for _, it := range gm.pl.Inventory {
		if it == k {
			n++
		}
	}
	return n
}

func (gm *Game) learnMartial(name string) {
	for _, m := range gm.pl.Martials {
		if m == name {
			return
		}
	}
	gm.pl.Martials = append(gm.pl.Martials, name)
}

// ---------- 使用物品 ----------

func (gm *Game) useItem(k ItemKind) {
	d := itemDefs[k]
	if d == nil || d.Quest {
		gm.push("这个物品不能直接使用")
		return
	}
	if d.HP > 0 && gm.pl.HP >= gm.pl.MaxHP && d.MP == 0 {
		gm.push("气血已满，无需用药")
		return
	}
	if d.MP > 0 && gm.pl.MP >= gm.pl.MaxMP && d.HP == 0 {
		gm.push("内力已满，无需用药")
		return
	}
	if !gm.removeItem(k) {
		gm.push("没有 " + d.Name)
		return
	}
	gm.pl.HP += d.HP
	if gm.pl.HP > gm.pl.MaxHP {
		gm.pl.HP = gm.pl.MaxHP
	}
	gm.pl.MP += d.MP
	if gm.pl.MP > gm.pl.MaxMP {
		gm.pl.MP = gm.pl.MaxMP
	}
	gm.push("使用" + d.Name + "！")
	gm.save()
}

// 快捷用药（地图上）
func (gm *Game) quickUse(k ItemKind) {
	if gm.countItem(k) <= 0 {
		gm.push("没有" + itemDefs[k].Name)
		return
	}
	gm.useItem(k)
}

// ---------- 存档 ----------

// SaveData 存档数据（JSON 序列化）。
type SaveData struct {
	Region   string
	PX, PY   int
	Pl       Player
	Dead     map[string]bool
	Opened   map[string]bool
	Flags    map[string]bool
	MainProg int
	Quests   map[int]bool
	Won      bool
}

func (gm *Game) save() {
	sd := SaveData{Region: gm.region, PX: gm.px, PY: gm.py, Pl: gm.pl,
		Dead: gm.dead, Opened: gm.opened, Flags: gm.flags,
		MainProg: gm.mainProg, Quests: gm.quests, Won: gm.won}
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
	gm.region = sd.Region
	gm.px, gm.py = sd.PX, sd.PY
	gm.pl = sd.Pl
	gm.dead = sd.Dead
	gm.opened = sd.Opened
	gm.flags = sd.Flags
	gm.mainProg = sd.MainProg
	gm.quests = sd.Quests
	gm.won = sd.Won
	if gm.dead == nil {
		gm.dead = map[string]bool{}
	}
	if gm.opened == nil {
		gm.opened = map[string]bool{}
	}
	if gm.flags == nil {
		gm.flags = map[string]bool{}
	}
	if gm.quests == nil {
		gm.quests = map[int]bool{}
	}
	return true
}

func (gm *Game) newGame() {
	gm.region = "world"
	gm.px, gm.py = 4, 3
	gm.pl = Player{Name: "无名侠客", Level: 1, XP: 0, HP: 60, MaxHP: 60, MP: 20, MaxMP: 20,
		Atk: 8, Def: 2, Money: 50, Weapon: "", Armor: "", Accessory: "",
		Inventory: []ItemKind{ItemGold, ItemGold, ItemGold}, Martials: []string{"江湖把式"},
		Sect: "", Realm: 0}
	gm.dead = map[string]bool{}
	gm.opened = map[string]bool{}
	gm.flags = map[string]bool{}
	gm.mainProg = 0
	gm.quests = map[int]bool{}
	gm.won = false
	gm.mode = "map"
	gm.spawnRegion()
	gm.msg = nil
	gm.push("你来到中原，江湖路远，万事小心。")
	gm.push("先去青州城拜见师父李长风吧！")
}

// ---------- 引擎钩子 ----------

func (gm *Game) onStart(g *engine.Game) {
	gm.dead, gm.opened, gm.flags = map[string]bool{}, map[string]bool{}, map[string]bool{}
	gm.quests = map[int]bool{}
	if gm.load() {
		gm.mode = "map"
		gm.spawnRegion()
		gm.push(fmt.Sprintf("（读档）回到%s", regions[gm.region].Name))
	} else {
		gm.newGame()
		gm.mode = "intro"
	}
	gm.sb = engine.NewScoreboard(filepath.Join(filepath.Dir(gm.savePath), "jianghu_rank.json"), 10)
}

func (gm *Game) onQuit(g *engine.Game) {
	gm.save()
}

// ---------- 输入 ----------

func (gm *Game) onKey(g *engine.Game, key string) {
	switch gm.mode {
	case "intro":
		gm.mode = "map"
	case "map":
		gm.onKeyMap(key)
	case "talk":
		if key == "enter" || key == "space" || key == "e" {
			gm.talkI++
			if gm.talkI >= len(gm.talk) {
				if gm.onTalkEnd != nil {
					fn := gm.onTalkEnd
					gm.onTalkEnd = nil
					fn()
				}
				// onTalkEnd 可能已切模式（inn/shop/casino），只有还在 talk 才回 map
				if gm.mode == "talk" {
					gm.mode = "map"
					if gm.won {
						gm.mode = "win"
					}
				}
			}
		} else if key == "q" {
			gm.mode = "map"
			if gm.won {
				gm.mode = "win"
			}
		}
	case "fight":
		gm.onKeyFight(key)
	case "menu":
		gm.onKeyMenu(key)
	case "shop":
		gm.onKeyShop(key)
	case "inn":
		gm.onKeyInn(key)
	case "casino":
		gm.onKeyCasino(key)
	case "enc":
		gm.onKeyEnc(key)
	case "dead", "win":
		switch key {
		case "r":
			gm.newGame()
			gm.save()
		case "q":
			g.Quit()
		}
	}
}

func (gm *Game) onKeyMap(key string) {
	if d, ok := dirs[key]; ok {
		gm.tryMove(d[0], d[1])
		return
	}
	switch key {
	case "e":
		gm.interact()
	case "1":
		gm.quickUse(ItemGold)
	case "2":
		gm.quickUse(ItemGreat)
	case "i":
		gm.menuTab, gm.menuIdx = 0, 0
		gm.mode = "menu"
	case "f":
		gm.menuTab, gm.menuIdx = 2, 0
		gm.mode = "menu"
	case "t":
		gm.menuTab, gm.menuIdx = 3, 0
		gm.mode = "menu"
	case "b":
		gm.menuTab, gm.menuIdx = 4, 0
		gm.mode = "menu"
	case "h":
		gm.menuTab, gm.menuIdx = 5, 0
		gm.mode = "menu"
	case "q":
		gm.save()
		gm.g.Quit()
	}
}

// ---------- 渲染 ----------

func (gm *Game) render(g *engine.Game, s *engine.Screen) {
	s.Clear()
	switch gm.mode {
	case "intro":
		gm.renderIntro(s)
	case "map", "talk", "enc":
		gm.renderStatus(s)
		gm.renderMap(s)
		gm.renderHelp(s)
		gm.renderMsg(s)
		if gm.mode == "talk" {
			gm.renderTalk(s)
		}
		if gm.mode == "enc" {
			gm.renderEnc(s)
		}
	case "fight":
		gm.renderFight(s)
	case "menu":
		gm.renderMenu(s)
	case "shop":
		gm.renderShop(s)
	case "inn":
		gm.renderInn(s)
	case "casino":
		gm.renderCasino(s)
	case "dead":
		gm.renderDead(s)
	case "win":
		gm.renderWin(s)
	}
}

func (gm *Game) renderStatus(s *engine.Screen) {
	p := &gm.pl
	atk, def, _, _ := gm.equipBonus()
	s.Text(1, 0, cut(fmt.Sprintf("%s · %s", p.Name, realmNames[p.Realm]), 11), engine.ColorCyan)
	s.Text(1, 1, cut(fmt.Sprintf("Lv%d HP%d/%d MP%d/%d", p.Level, p.HP, p.MaxHP, p.MP, p.MaxMP), 19), engine.ColorWhite)
	s.Text(23, 0, cut(fmt.Sprintf("攻%d 防%d", p.Atk+atk, p.Def+def), 10), engine.ColorYellow)
	s.Text(23, 1, cut(fmt.Sprintf("银%d两", p.Money), 10), engine.ColorYellow)
}

func (gm *Game) renderMap(s *engine.Screen) {
	r := gm.curRegion()
	for y := 0; y < r.H; y++ {
		for x := 0; x < r.W; x++ {
			t := r.Tiles[y][x]
			fg := engine.ColorGray
			switch t {
			case '#':
				fg = engine.ColorDarkGray
			case 'T':
				fg = engine.ColorGreen
			case '~':
				fg = engine.ColorBlue
			case '^':
				fg = engine.ColorWhite
			case '>', '<':
				fg = engine.ColorCyan
			case '?':
				fg = engine.ColorMagenta
			}
			if mapOffY+y < scrH && mapXOff+x < scrW {
				s.Set(mapXOff+x, mapOffY+y, t, fg, -1)
			}
		}
	}
	for _, o := range gm.objs {
		if o.Dead || o.Opened {
			continue
		}
		if o.X >= 0 && o.Y >= 0 && mapXOff+o.X < scrW && mapOffY+o.Y < scrH {
			s.Set(mapXOff+o.X, mapOffY+o.Y, o.Ch, o.Fg, -1)
		}
	}
	s.Set(mapXOff+gm.px, mapOffY+gm.py, '@', engine.ColorCyan, 17)
}

func (gm *Game) renderHelp(s *engine.Screen) {
	s.Text(1, helpY1, helpLine1, engine.ColorGray)
	s.Text(1, helpY2, helpLine2, engine.ColorGray)
}

func (gm *Game) renderMsg(s *engine.Screen) {
	for i, m := range gm.msg {
		lines := wrapText(m, 20)
		for j, ln := range lines {
			yy := msgY1 + i
			if yy >= 0 && yy < scrH && j == 0 {
				s.Text(1, yy, cut(ln, 20), engine.ColorGray)
			}
		}
	}
}

func (gm *Game) renderTalk(s *engine.Screen) {
	if gm.talkI >= len(gm.talk) {
		return
	}
	lines := wrapText(gm.talk[gm.talkI], 10)
	for i, ln := range lines {
		if infoY+i < 13 {
			s.Text(infoX, infoY+i, ln, engine.ColorYellow)
		}
	}
	ty := infoY + len(lines) + 1
	if ty > 12 {
		ty = 12
	}
	s.Text(infoX, ty, "Enter 继续  Q 跳过", engine.ColorGray)
}

// ---------- 主函数 ----------

func main() {
	dir, _ := os.Getwd()
	botMode := len(os.Args) > 1 && os.Args[1] == "--bot"
	gm := &Game{savePath: filepath.Join(dir, "jianghu_save.json")}
	bw, bh := scrW, scrH
	if botMode {
		bw, bh = 1, 1 // bot 不渲染，1x1 画布避免刷屏
	}
	g := engine.NewGame("江湖行 · 武侠开放世界", bw, bh, 30)
	gm.g = g
	gm.g.OnStart = gm.onStart
	gm.g.OnKey = gm.onKey
	gm.g.OnQuit = gm.onQuit
	if botMode {
		bot := &Bot{gm: gm, startLv: 1}
		if f, err := os.OpenFile(filepath.Join(os.Getenv("HOME"), "hermes11", "bot_tmp.log"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644); err == nil {
			bot.logFile = f
		}
		g.Update = func(gg *engine.Game, dt float64) { bot.think() }
		g.Render = func(gg *engine.Game, s *engine.Screen) {}
		gm.g.OnQuit = func(gg *engine.Game) {
			fmt.Println("=== AI 自动游玩报告 ===")
			fmt.Println(bot.report())
		}
	} else {
		g.Render = gm.render
	}
	gm.g.Run()
}
