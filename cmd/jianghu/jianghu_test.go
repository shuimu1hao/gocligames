// 江湖行 - 测试套件：地图校验 / 移动 / 传送 / 对话 / 战斗 / 主线通关 / 存档
package main

import (
	"math/rand"
	"path/filepath"
	"testing"

	"gocligames/engine"
)

func tileWalkable(name string, x, y int) bool {
	r := regions[name]
	if x < 0 || y < 0 || x >= r.W || y >= r.H {
		return false
	}
	switch r.Tiles[y][x] {
	case '#', 'T', '~':
		return false
	}
	return true
}

func testGame(t *testing.T) *Game {
	t.Helper()
	dir := t.TempDir()
	gm := &Game{savePath: filepath.Join(dir, "save.json")}
	gm.newGame()
	gm.sb = engine.NewScoreboard(filepath.Join(dir, "rank.json"), 10)
	return gm
}

// 所有区域：12 行、每行 20 列、纯 ASCII
func TestMapShapes(t *testing.T) {
	for _, name := range regionOrder {
		r := regions[name]
		if r.H != 12 {
			t.Fatalf("%s rows=%d want 12", name, r.H)
		}
		for y, row := range r.Tiles {
			if len(row) != 20 {
				t.Fatalf("%s row%d width=%d want 20: %s", name, y, len(row), string(row))
			}
			for _, ch := range row {
				if ch > 0x7f {
					t.Fatalf("%s row%d non-ascii %q", name, y, ch)
				}
			}
		}
	}
}

// 所有 spawn 在可走格
func TestSpawnsWalkable(t *testing.T) {
	for _, name := range regionOrder {
		for _, sp := range regionSpawns[name] {
			if !tileWalkable(name, sp.X, sp.Y) {
				t.Fatalf("%s spawn (%d,%d) not walkable", name, sp.X, sp.Y)
			}
		}
	}
}

// portal 与地图 '<' '>' 一致、目标可走
func TestPortalsConsistent(t *testing.T) {
	for _, p := range portals {
		ch := regions[p.From].Tiles[p.Y][p.X]
		if ch != '>' && ch != '<' {
			t.Fatalf("portal %s(%d,%d) tile=%c", p.From, p.X, p.Y, ch)
		}
		if !tileWalkable(p.To, p.ToX, p.ToY) {
			t.Fatalf("portal dest %s(%d,%d) not walkable", p.To, p.ToX, p.ToY)
		}
	}
}

// 新游戏初始状态
func TestNewGame(t *testing.T) {
	gm := testGame(t)
	if gm.pl.Level != 1 || gm.pl.HP != 60 || gm.pl.Money != 50 {
		t.Fatalf("bad init: %+v", gm.pl)
	}
	if !gm.walkable(gm.px, gm.py) {
		t.Fatalf("spawn (%d,%d) not walkable", gm.px, gm.py)
	}
	if gm.region != "world" {
		t.Fatalf("start region=%s", gm.region)
	}
}

// 撞墙不动，空地能走
func TestMove(t *testing.T) {
	gm := testGame(t)
	gm.px, gm.py = 4, 3
	gm.tryMove(1, 0) // (5,3) 空地
	if gm.px != 5 {
		t.Fatalf("move right failed px=%d", gm.px)
	}
	gm.px, gm.py = 3, 3
	gm.tryMove(1, 0) // (4,3) 空地
	if gm.px != 4 {
		t.Fatalf("move to open failed px=%d", gm.px)
	}
	gm.px, gm.py = 3, 2
	gm.tryMove(1, 0) // (4,2) 空地
	gm.tryMove(1, 0) // (5,2) 空地
	gm.tryMove(1, 0) // (6,2) 空地
	gm.tryMove(1, 0) // (7,2) 空地
	gm.tryMove(1, 0) // (8,2) 是 T，撞树不动
	if gm.px != 7 {
		t.Fatalf("should hit tree at x8, px=%d", gm.px)
	}
	// 撞地图边界
	gm.px, gm.py = 1, 5
	gm.tryMove(-1, 0)
	if gm.px != 1 {
		t.Fatalf("should hit wall at x0, px=%d", gm.px)
	}
}

// 踩 '>' 进入青州城，'<' 返回
func TestPortalTravel(t *testing.T) {
	gm := testGame(t)
	gm.px, gm.py = 12, 2
	gm.tryMove(1, 0) // 踩到 (13,2) '>'
	if gm.region != "qingzhou" {
		t.Fatalf("enter qingzhou failed, region=%s", gm.region)
	}
	gm.px, gm.py = 15, 6
	gm.tryMove(1, 0) // 踩到 (16,6) '<'
	if gm.region != "world" {
		t.Fatalf("return world failed, region=%s", gm.region)
	}
	if gm.px != 12 || gm.py != 3 {
		t.Fatalf("return pos (%d,%d)", gm.px, gm.py)
	}
}

// 踩 '?' 触发奇遇并标记
func TestEncounter(t *testing.T) {
	gm := testGame(t)
	gm.px, gm.py = 2, 10
	gm.tryMove(1, 0) // 踩 (3,10) '?'
	if gm.encID != "world-3-10" || gm.mode == "map" {
		t.Fatalf("encounter not triggered: encID=%s mode=%s", gm.encID, gm.mode)
	}
	if gm.mode != "talk" {
		t.Fatalf("should enter talk mode, got %s", gm.mode)
	}
}

// 对话师父推进主线 0→1
func TestTalkMaster(t *testing.T) {
	gm := testGame(t)
	gm.startTalk("师父李长风")
	if gm.mode != "talk" {
		t.Fatal("should enter talk")
	}
	// 走完所有对话行
	for i := 0; i < len(gm.talk)+1; i++ {
		if gm.mode != "talk" {
			break
		}
		if gm.talkI < len(gm.talk) {
			gm.talkI++
		}
	}
	if gm.mode == "talk" {
		gm.onTalkEnd()
		gm.mode = "map"
	}
	if gm.mainProg != 1 {
		t.Fatalf("mainProg=%d want 1", gm.mainProg)
	}
	if gm.countItem(ItemGold) != 4 { // 初始 3 + 师父 1
		t.Fatalf("gold pills=%d want 4", gm.countItem(ItemGold))
	}
}

// 战斗：打赢小怪
func TestFightWin(t *testing.T) {
	gm := testGame(t)
	gm.startFight(0) // 野狗
	if gm.mode != "fight" {
		t.Fatal("should enter fight")
	}
	// 玩家攻击到死（防御高，循环防 miss）
	gm.pl.Atk = 30
	for i := 0; i < 10 && gm.mode == "fight"; i++ {
		gm.fightAttack(0)
	}
	if gm.mode != "map" {
		t.Fatalf("fight should end, mode=%s", gm.mode)
	}
	if gm.pl.XP <= 0 {
		t.Fatal("should gain xp")
	}
}

// 战斗：死亡 → dead 模式 + 排行榜记录
func TestFightLose(t *testing.T) {
	gm := testGame(t)
	gm.startFight(5) // 寨主
	gm.pl.HP = 1
	gm.pl.Def = 0
	gm.enemyTurn()
	if gm.mode != "dead" {
		t.Fatalf("should be dead, mode=%s", gm.mode)
	}
	if gm.sb.Top(10) == nil {
		t.Fatal("rank should have record after death")
	}
}

// 完整主线：打黑风寨主 → 密信 → 师父 → 古墓 → 捕头 → 天山教主通关
func TestFullMainQuest(t *testing.T) {
	gm := testGame(t)
	// 强化玩家保证能打赢
	gm.pl = Player{Name: "测试侠", Level: 30, HP: 999, MaxHP: 999, MP: 500, MaxMP: 500,
		Atk: 60, Def: 40, Money: 1000, Inventory: []ItemKind{ItemGold, ItemGreat},
		Martials: []string{"江湖把式", "天蚕神功"}, Sect: "武当派", Realm: 4}

	// 1. 对话师父开主线
	gm.startTalk("师父李长风")
	for i := 0; i < 10 && gm.mode == "talk"; i++ {
		gm.talkI++
		if gm.talkI >= len(gm.talk) {
			if gm.onTalkEnd != nil {
				fn := gm.onTalkEnd
				gm.onTalkEnd = nil
				fn()
			}
			gm.mode = "map"
		}
	}
	if gm.mainProg != 1 {
		t.Fatalf("step1 mainProg=%d", gm.mainProg)
	}

	// 2. 打黑风寨主拿密信
	gm.startFight(5)
	for gm.mode == "fight" {
		gm.fightAttack(1) // 用天蚕神功
	}
	if gm.countItem(ItemLetter) != 1 {
		t.Fatalf("should have letter, got %d", gm.countItem(ItemLetter))
	}
	if gm.mainProg != 2 {
		t.Fatalf("step2 mainProg=%d", gm.mainProg)
	}

	// 3. 师父解读密信
	gm.startTalk("师父李长风")
	for i := 0; i < 10 && gm.mode == "talk"; i++ {
		gm.talkI++
		if gm.talkI >= len(gm.talk) {
			if gm.onTalkEnd != nil {
				fn := gm.onTalkEnd
				gm.onTalkEnd = nil
				fn()
			}
			gm.mode = "map"
		}
	}
	if gm.mainProg != 3 {
		t.Fatalf("step3 mainProg=%d", gm.mainProg)
	}

	// 4. 古墓开箱拿暖玉
	gm.openChest(&Obj{Item: ItemJade, Key: "c_tomb_jade"})
	if gm.countItem(ItemJade) != 1 {
		t.Fatal("should have jade")
	}

	// 5. 捕头 → 天山
	gm.startTalk("捕头")
	for i := 0; i < 10 && gm.mode == "talk"; i++ {
		gm.talkI++
		if gm.talkI >= len(gm.talk) {
			if gm.onTalkEnd != nil {
				fn := gm.onTalkEnd
				gm.onTalkEnd = nil
				fn()
			}
			gm.mode = "map"
		}
	}
	if gm.mainProg != 5 {
		t.Fatalf("step5 mainProg=%d", gm.mainProg)
	}

	// 6. 打教主通关
	gm.startFight(10)
	for gm.mode == "fight" {
		gm.fightAttack(1)
	}
	if !gm.won {
		t.Fatal("should win")
	}
	if gm.mainProg != 6 {
		t.Fatalf("final mainProg=%d", gm.mainProg)
	}
	if len(gm.sb.Top(10)) == 0 {
		t.Fatal("win should record rank")
	}
}

// 存档读档
func TestSaveLoad(t *testing.T) {
	gm := testGame(t)
	gm.pl.Money = 777
	gm.pl.Weapon = "铁剑"
	gm.mainProg = 3
	gm.save()

	gm2 := testGame(t)
	gm2.savePath = gm.savePath
	if !gm2.load() {
		t.Fatal("load failed")
	}
	if gm2.pl.Money != 777 || gm2.pl.Weapon != "铁剑" || gm2.mainProg != 3 {
		t.Fatalf("load mismatch: %+v", gm2.pl)
	}
}

// 物品/装备/武功基础操作
func TestItemsAndMartials(t *testing.T) {
	gm := testGame(t)
	gm.addItem(ItemGold)
	if gm.countItem(ItemGold) != 4 {
		t.Fatalf("count=%d", gm.countItem(ItemGold))
	}
	if !gm.removeItem(ItemGold) {
		t.Fatal("remove failed")
	}
	gm.pl.HP = 10
	gm.useItem(ItemGold)
	if gm.pl.HP != gm.pl.MaxHP {
		t.Fatalf("HP=%d want %d", gm.pl.HP, gm.pl.MaxHP)
	}
	gm.learnMartial("太极剑法")
	if !gm.hasMartial("太极剑法") {
		t.Fatal("learn failed")
	}
	gm.pl.Money = 1000
	gm.buyEquip("铁剑")
	if gm.pl.Weapon != "铁剑" || gm.pl.Money != 880 {
		t.Fatalf("equip/money wrong: weapon=%s money=%d", gm.pl.Weapon, gm.pl.Money)
	}
}

// 境界计算
func TestRealm(t *testing.T) {
	cases := map[int]int{1: 0, 6: 1, 12: 2, 20: 3, 30: 4, 42: 5, 60: 6}
	for lv, want := range cases {
		if got := realmForLevel(lv); got != want {
			t.Fatalf("realmForLevel(%d)=%d want %d", lv, got, want)
		}
	}
}

// 战力为正
func TestPowerPositive(t *testing.T) {
	gm := testGame(t)
	if gm.power() <= 0 {
		t.Fatal("power should be positive")
	}
}

// 世界地图连通性：出生点 BFS 可达所有入口/NPC/怪物/奇遇
func TestWorldConnectivity(t *testing.T) {
	type pt struct{ x, y int }
	start := pt{4, 3}
	q := []pt{start}
	seen := map[pt]bool{start: true}
	for len(q) > 0 {
		p := q[0]
		q = q[1:]
		for _, d := range [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
			nx, ny := p.x+d[0], p.y+d[1]
			np := pt{nx, ny}
			if !seen[np] && tileWalkable("world", nx, ny) {
				seen[np] = true
				q = append(q, np)
			}
		}
	}
	targets := []pt{{13, 2}, {8, 4}, {16, 4}, {13, 7}, {13, 9},
		{2, 2}, {10, 5}, {17, 8}, {4, 1}, {15, 1}, {9, 3}, {17, 3},
		{4, 6}, {15, 6}, {4, 8}, {15, 8}, {7, 10}, {3, 10}, {9, 10}, {16, 10}}
	for _, tp := range targets {
		if !tileWalkable("world", tp.x, tp.y) {
			t.Fatalf("target (%d,%d) not walkable", tp.x, tp.y)
		}
		if !seen[tp] {
			t.Fatalf("target (%d,%d) unreachable from spawn", tp.x, tp.y)
		}
	}
}

// Boss 战胜率模拟：不同等级装备下打各 Boss 的胜率（数值平衡验证）
func TestBossWinRate(t *testing.T) {
	type setup struct {
		lv, atk, def, hp int
		weapon           string
	}
	setups := []setup{
		{10, 20, 10, 150, "铁剑"},
		{15, 28, 14, 220, "青锋剑"},
		{20, 36, 18, 300, "青锋剑"},
		{30, 52, 28, 480, "龙泉宝剑"},
	}
	bosses := []int{5, 7, 10} // 寨主/尸王/教主
	for _, s := range setups {
		for _, mi := range bosses {
			wins := 0
			trials := 50
			for i := 0; i < trials; i++ {
				if simBossFight(s, mi) {
					wins++
				}
			}
			t.Logf("Lv%d 攻%d 防%d HP%d %s vs %s: 胜率 %d/%d",
				s.lv, s.atk, s.def, s.hp, s.weapon, monsterDefs[mi].Name, wins, trials)
		}
	}
}

// simBossFight 模拟一场 Boss 战（纯数值）
func simBossFight(s struct {
	lv, atk, def, hp int
	weapon           string
}, mi int) bool {
	d := monsterDefs[mi]
	ph, pm := s.hp, 100
	atk := s.atk
	if e := findEquip(s.weapon); e != nil {
		atk += e.Atk
	}
	mh := d.HP
	for i := 0; i < 200; i++ {
		base := 12 // 江湖把式
		if pm >= 12 {
			base = 34 // 天蚕神功
			pm -= 12
		}
		dmg := base + atk - d.Def + rand.Intn(8)
		if dmg < 1 {
			dmg = 1
		}
		mh -= dmg
		if mh <= 0 {
			return true
		}
		ed := d.Atk - s.def + rand.Intn(4)
		if ed < 1 {
			ed = 1
		}
		ph -= ed
		if ph <= 0 {
			return false
		}
		pm += 8
		if pm > 100 {
			pm = 100
		}
	}
	return false
}

// Bot 模拟：多次随机种子验证 AI 通关稳定性
func TestBotSimulate(t *testing.T) {
	for run := 0; run < 5; run++ {
		gm := testGame(t)
		bot := &Bot{gm: gm, startLv: 1}
		for i := 0; i < 36000 && !bot.done && !bot.timeout; i++ {
			bot.think()
		}
		if !bot.done {
			t.Fatalf("run%d STUCK: mainProg=%d lv=%d region=%s dead=%d money=%d martials=%v",
				run, gm.mainProg, gm.pl.Level, gm.region, bot.deadCount, gm.pl.Money, gm.pl.Martials)
		}
		t.Logf("run%d %s", run, bot.report())
	}
}

// 支线任务全流程：药铺急单 / 洛阳乞酒 / 镖局押镖
func TestSideQuests(t *testing.T) {
	gm := testGame(t)
	// 药铺急单
	gm.startTalk("药铺掌柜")
	for i := 0; i < 10 && gm.mode == "talk"; i++ {
		gm.talkI++
		if gm.talkI >= len(gm.talk) {
			if gm.onTalkEnd != nil {
				fn := gm.onTalkEnd
				gm.onTalkEnd = nil
				fn()
			}
			if gm.mode == "talk" {
				gm.mode = "map"
			}
		}
	}
	if !gm.flags["quest10_started"] {
		t.Fatal("quest10 should start")
	}
	gm.addItem(ItemHerb)
	gm.startTalk("药铺掌柜")
	for i := 0; i < 10 && gm.mode == "talk"; i++ {
		gm.talkI++
		if gm.talkI >= len(gm.talk) {
			if gm.onTalkEnd != nil {
				fn := gm.onTalkEnd
				gm.onTalkEnd = nil
				fn()
			}
			if gm.mode == "talk" {
				gm.mode = "map"
			}
		}
	}
	if !gm.quests[10] {
		t.Fatal("quest10 should complete")
	}
	if gm.pl.Money < 200 {
		t.Fatalf("quest10 reward missing, money=%d", gm.pl.Money)
	}
	// 洛阳乞酒：买酒 → 交丐帮弟子
	gm.pl.Money = 500
	gm.startTalk("酒馆老板")
	for i := 0; i < 10 && gm.mode == "talk"; i++ {
		gm.talkI++
		if gm.talkI >= len(gm.talk) {
			if gm.onTalkEnd != nil {
				fn := gm.onTalkEnd
				gm.onTalkEnd = nil
				fn()
			}
			if gm.mode == "talk" {
				gm.mode = "map"
			}
		}
	}
	if gm.countItem(ItemWine) != 1 {
		t.Fatal("should buy wine")
	}
	gm.startTalk("丐帮弟子")
	for i := 0; i < 10 && gm.mode == "talk"; i++ {
		gm.talkI++
		if gm.talkI >= len(gm.talk) {
			if gm.onTalkEnd != nil {
				fn := gm.onTalkEnd
				gm.onTalkEnd = nil
				fn()
			}
			if gm.mode == "talk" {
				gm.mode = "map"
			}
		}
	}
	if !gm.quests[11] {
		t.Fatal("quest11 should complete")
	}
	// 镖局押镖：触发战斗 → 打赢 → 交任务
	gm.startTalk("镖局镖头")
	for i := 0; i < 10 && gm.mode == "talk"; i++ {
		gm.talkI++
		if gm.talkI >= len(gm.talk) {
			if gm.onTalkEnd != nil {
				fn := gm.onTalkEnd
				gm.onTalkEnd = nil
				fn()
			}
			if gm.mode == "talk" {
				gm.mode = "map"
			}
		}
	}
	if gm.mode != "fight" {
		t.Fatalf("quest12 should trigger fight, mode=%s", gm.mode)
	}
	gm.pl.Atk = 50
	for gm.mode == "fight" {
		gm.fightAttack(0)
	}
	gm.startTalk("镖局镖头")
	for i := 0; i < 10 && gm.mode == "talk"; i++ {
		gm.talkI++
		if gm.talkI >= len(gm.talk) {
			if gm.onTalkEnd != nil {
				fn := gm.onTalkEnd
				gm.onTalkEnd = nil
				fn()
			}
			if gm.mode == "talk" {
				gm.mode = "map"
			}
		}
	}
	if !gm.quests[12] {
		t.Fatal("quest12 should complete")
	}
}

// 商店购买/装备/客栈休息/赌场
func TestShopInnCasino(t *testing.T) {
	gm := testGame(t)
	gm.pl.Money = 1000
	// 商店买东西
	gm.shopID = "luoyang_shop"
	gm.buyItem(ItemGreat)
	if gm.countItem(ItemGreat) != 1 || gm.pl.Money != 910 {
		t.Fatalf("buy item failed: money=%d", gm.pl.Money)
	}
	gm.buyEquip("铁剑")
	if gm.pl.Weapon != "铁剑" || gm.pl.Money != 790 {
		t.Fatalf("buy equip failed: weapon=%s money=%d", gm.pl.Weapon, gm.pl.Money)
	}
	// 客栈休息
	gm.pl.HP = 10
	gm.innName = "测试客栈"
	gm.innCost = 20
	gm.onKeyInn("1")
	if gm.pl.HP != gm.pl.MaxHP || gm.pl.Money != 770 {
		t.Fatalf("inn rest failed: hp=%d money=%d", gm.pl.HP, gm.pl.Money)
	}
	// 赌场：押大（钱足够）
	gm.pl.Money = 100
	gm.mode = "casino"
	gm.casinoState = 0
	gm.onKeyCasino("1")
	if gm.mode != "map" {
		t.Fatal("casino should return to map")
	}
	if gm.pl.Money == 100 {
		t.Fatal("casino should change money")
	}
}

// 通关后：世界地图出现血月余孽（冲榜目标）
func TestAfterBossYunie(t *testing.T) {
	gm := testGame(t)
	gm.won = true
	gm.region = "world"
	gm.spawnRegion()
	found := 0
	for _, o := range gm.objs {
		if o.MIdx == 11 {
			found++
		}
	}
	if found != 3 {
		t.Fatalf("should spawn 3 yunie after win, got %d", found)
	}
	// 客栈休息后余孽也重生
	gm.dead["m_after_1"] = true
	gm.respawnWorld()
	if gm.dead["m_after_1"] {
		t.Fatal("respawnWorld should clear yunie dead flag")
	}
}
