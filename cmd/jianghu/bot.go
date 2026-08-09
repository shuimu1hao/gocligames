// 江湖行 - AI 自动玩家（--bot 模式）
// 用于数值平衡自测与自动打榜：BFS 寻路 + 自动战斗 + 主线推进 + 死亡复活
package main

import (
	"fmt"
	"os"
)

type Bot struct {
	gm        *Game
	frames    int
	deadCount int
	farmLeft  int
	done      bool
	timeout   bool
	fights    int
	startLv   int
	afterRest bool
	quitDone  bool
	logFile   *os.File
}

func (b *Bot) logf(format string, args ...interface{}) {
	if b.logFile == nil {
		return
	}
	fmt.Fprintf(b.logFile, format+"\n", args...)
}

// think 每帧调用（非 tty 环境由 Update 驱动，不需要输入）
func (b *Bot) think() {
	b.frames++
	if b.done || b.timeout {
		if !b.quitDone {
			b.quitDone = true
			b.gm.g.Quit() // 退出引擎循环，触发 OnQuit 打印报告
		}
		return
	}
	if b.frames > 36000 { // 约 20 分钟上限
		b.timeout = true
		return
	}
	if b.frames%4 != 0 { // 降频决策
		return
	}
	gm := b.gm
	switch gm.mode {
	case "intro":
		gm.mode = "map"
	case "fight":
		b.fight()
	case "talk":
		if gm.talkI < len(gm.talk) {
			gm.talkI++
		}
		if gm.talkI >= len(gm.talk) {
			if gm.onTalkEnd != nil {
				fn := gm.onTalkEnd
				gm.onTalkEnd = nil
				fn()
			}
			if gm.mode == "talk" {
				gm.mode = "map"
			}
			if gm.won {
				b.done = true
			}
		}
	case "menu", "shop", "casino":
		gm.mode = "map"
	case "inn":
		gm.onKeyInn("1") // 住店休息
		b.afterRest = true
	case "dead":
		b.deadCount++
		b.logf("frame=%d DEAD count=%d at %s", b.frames, b.deadCount, gm.region)
		gm.newGame()
		gm.save()
	case "win":
		b.done = true
	default:
		b.mapStep()
	}
}

// ---------- 目标 ----------

// targetRegion 当前阶段应去的区域
func (b *Bot) targetRegion() string {
	p := b.gm.mainProg
	switch {
	case p == 2: // 有密信，回师父
		return "qingzhou"
	case p == 3: // 古墓
		return "tomb"
	case p == 4: // 洛阳捕头
		return "luoyang"
	case p == 5: // 天山
		return "peak"
	default: // p 0/1 去黑风寨
		if b.gm.flags["bot_got_quest"] {
			return "fort"
		}
		return "qingzhou"
	}
}

// target 当前区域内的目标坐标；-1 表示不在本区域
func (b *Bot) target() (int, int) {
	gm := b.gm
	if b.farmLeft > 0 {
		return -1, -1
	}
	switch gm.mainProg {
	case 2:
		return 10, 2 // 师父
	case 3:
		if !gm.hasMartial("天蚕神功") {
			return 10, 9 // 宝箱：天蚕神功残卷（先拿，别漏绝世武功）
		}
		if gm.countItem(ItemJade) == 0 {
			return 16, 8 // 宝箱：暖玉
		}
		return 10, 3 // 尸王
	case 4:
		return 10, 4 // 捕头
	case 5:
		return 7, 5 // 教主
	default:
		if gm.flags["bot_got_quest"] {
			return 11, 6 // 寨主
		}
		return 10, 2 // 师父
	}
}

// ---------- 移动 ----------

func (b *Bot) mapStep() {
	gm := b.gm
	if gm.won {
		b.done = true
		return
	}
	// 血量低于 1/3 先吃药
	if gm.pl.HP < gm.pl.MaxHP/3 {
		if gm.countItem(ItemGreat) > 0 {
			gm.useItem(ItemGreat)
			return
		}
		if gm.countItem(ItemGold) > 0 {
			gm.useItem(ItemGold)
			return
		}
	}
	// 等级不足：先练级（world 打怪 / 客栈休息重生 / 其他区域回 world）
	if gm.pl.Level < b.farmLevel() {
		b.farm()
		return
	}
	// farm 阶段：升级到打 Boss 更稳
	if b.farmLeft > 0 && gm.mode == "map" {
		b.doFarm()
		return
	}
	// 目标区域判断
	tr := b.targetRegion()
	if tr != gm.region {
		if b.frames%300 == 0 {
			b.logf("frame=%d nav: %s -> %s (mainProg=%d lv=%d)", b.frames, gm.region, tr, gm.mainProg, gm.pl.Level)
		}
		b.goToPortal(tr)
		return
	}
	tx, ty := b.target()
	if tx < 0 {
		// 区域内无目标（如 farm 完）：站住
		return
	}
	if tx == gm.px && ty == gm.py {
		// 到达目标：交互
		if o := gm.objAt(tx, ty); o != nil && o.Kind == SpawnNPC {
			gm.startTalk(o.NPC)
			if o.NPC == "师父李长风" && gm.mainProg == 0 {
				gm.flags["bot_got_quest"] = true
			}
			return
		}
		if o := gm.objAt(tx, ty); o != nil && o.Kind == SpawnChest {
			gm.openChest(o)
			return
		}
		// 目标位置无对象（如 Boss 已死）：推进 farm
		b.farmLeft = 3
		return
	}
	// 走到目标旁边（相邻即交互）
	dx, dy := b.stepTo(tx, ty)
	if dx == 0 && dy == 0 {
		// 已相邻：如果目标是 NPC/宝箱按 E
		if o := gm.objAt(tx, ty); o != nil && (o.Kind == SpawnNPC || o.Kind == SpawnChest) {
			gm.interact()
			return
		}
		if gm.tileAt(tx, ty) == '>' || gm.tileAt(tx, ty) == '<' {
			gm.tryMove(tx-gm.px, ty-gm.py)
			return
		}
		b.farmLeft = 2
		return
	}
	gm.tryMove(dx, dy)
}

// goToPortal 走向通往目标区域的传送门
func (b *Bot) goToPortal(tr string) {
	gm := b.gm
	// 当前区域找传送点
	var px, py = -1, -1
	if gm.region == "world" {
		for _, p := range portals {
			if p.From == "world" && p.To == tr {
				px, py = p.X, p.Y
				break
			}
		}
	} else {
		// 在子区域，找 '<' 回 world
		for _, p := range portals {
			if p.From == gm.region && p.To == "world" {
				px, py = p.X, p.Y
				break
			}
		}
	}
	if px < 0 {
		return
	}
	dx, dy := b.stepTo(px, py)
	if dx == 0 && dy == 0 {
		gm.tryMove(px-gm.px, py-gm.py)
		return
	}
	gm.tryMove(dx, dy)
}

// stepTo BFS 返回从当前位置到 (tx,ty) 的下一步偏移
func (b *Bot) stepTo(tx, ty int) (int, int) {
	gm := b.gm
	type node struct{ x, y int }
	prev := map[node]node{}
	start := node{gm.px, gm.py}
	q := []node{start}
	seen := map[node]bool{start: true}
	dirs4 := [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}
	for len(q) > 0 {
		cur := q[0]
		q = q[1:]
		if cur.x == tx && cur.y == ty {
			// 回溯第一步
			first := cur
			for prev[first] != start {
				first = prev[first]
				if p, ok := prev[first]; !ok || p == first {
					break
				}
			}
			return first.x - gm.px, first.y - gm.py
		}
		for _, d := range dirs4 {
			nx, ny := cur.x+d[0], cur.y+d[1]
			n := node{nx, ny}
			if seen[n] || !gm.walkable(nx, ny) {
				continue
			}
			// 避开传送门（除非目标就是传送门），防止中途误传
			if (gm.tileAt(nx, ny) == '>' || gm.tileAt(nx, ny) == '<') && !(nx == tx && ny == ty) {
				continue
			}
			// 怪物会开战，BFS 允许走怪物格（tryMove 会开战）
			seen[n] = true
			prev[n] = cur
			q = append(q, n)
		}
	}
	return 0, 0
}

// doFarm 在 world 打野怪升级
// farmLevel 当前阶段建议的最低等级
func (b *Bot) farmLevel() int {
	switch {
	case b.gm.mainProg >= 4:
		return 10 // 天山决战前（天蚕神功在手，Lv10 稳过教主）
	case b.gm.mainProg >= 3:
		return 8 // 古墓前
	case b.gm.mainProg >= 2:
		return 7 // 师父解读后
	default:
		return 6 // 黑风寨前
	}
}

// farm 练级：world 打怪 → 怪杀光去青州客栈休息重生 → 再打
func (b *Bot) farm() {
	gm := b.gm
	// 刚休息完：先回到 world 再继续练（标志在到达 world 后才清）
	if b.afterRest {
		if gm.region == "world" {
			b.afterRest = false
		} else {
			b.goToPortal("world")
			return
		}
	}
	if gm.region == "world" {
		// 血量不足先去客栈恢复（保证打 Boss 时满状态）
		if gm.pl.HP < gm.pl.MaxHP*2/3 && (gm.pl.Money >= 20 || gm.pl.HP < gm.pl.MaxHP/2) {
			b.goToPortal("qingzhou")
			return
		}
		if b.frames%300 == 0 {
			b.logf("frame=%d farming lv=%d/%d hp=%d/%d", b.frames, gm.pl.Level, b.farmLevel(), gm.pl.HP, gm.pl.MaxHP)
		}
		if b.doFarm() {
			return
		}
		// world 怪杀光还不够级：去青州客栈
		b.goToPortal("qingzhou")
		return
	}
	if gm.region == "qingzhou" {
		// 去客栈小二 (2,4) 休息重生野怪
		if gm.px == 2 && gm.py == 4 {
			gm.startTalk("客栈小二")
			return
		}
		dx, dy := b.stepTo(2, 4)
		if dx == 0 && dy == 0 {
			gm.interact()
			return
		}
		gm.tryMove(dx, dy)
		return
	}
	// 其他区域：回 world
	b.goToPortal("world")
}

func (b *Bot) doFarm() bool {
	gm := b.gm
	// 前期只打安全的怪，随等级扩大猎杀范围
	safe := map[int]bool{0: true, 1: true}
	if gm.pl.Level >= 3 {
		safe[2] = true // 山贼
	}
	if gm.pl.Level >= 5 {
		safe[3] = true // 毒蛇
	}
	bestX, bestY, bd := -1, -1, -1
	for _, o := range gm.objs {
		if o.Kind != SpawnMonster || o.Dead {
			continue
		}
		if !safe[o.MIdx] {
			continue
		}
		d := (o.X-gm.px)*(o.X-gm.px) + (o.Y-gm.py)*(o.Y-gm.py)
		if bd < 0 || d < bd {
			bestX, bestY, bd = o.X, o.Y, d
		}
	}
	if bestX < 0 {
		b.farmLeft = 0
		return false
	}
	if bestX == gm.px && bestY == gm.py {
		gm.tryMove(bestX-gm.px, bestY-gm.py) // 撞怪开战
		return true
	}
	dx, dy := b.stepTo(bestX, bestY)
	if dx == 0 && dy == 0 {
		gm.interact()
		return true
	}
	gm.tryMove(dx, dy)
	return true
}

// ---------- 战斗 ----------

func (b *Bot) fight() {
	gm := b.gm
	f := gm.fight
	if f == nil {
		gm.mode = "map"
		return
	}
	d := gm.curMonster()
	// 血低先吃药
	if gm.pl.HP < gm.pl.MaxHP/2 {
		if gm.countItem(ItemGreat) > 0 {
			gm.fightItem(ItemGreat)
			return
		}
		if gm.countItem(ItemGold) > 0 {
			gm.fightItem(ItemGold)
			return
		}
	}
	// 选招：MP 允许内伤害最高的攻击招
	bestIdx := -1
	bestDmg := -1
	for i, mname := range gm.pl.Martials {
		m := martialDefs[mname]
		if m.Kind == "heal" {
			continue
		}
		if m.Cost > gm.pl.MP {
			continue
		}
		dmg := m.Base + m.Rand + m.Combo*(m.Base/2+m.Rand/2)
		if dmg > bestDmg {
			bestDmg = dmg
			bestIdx = i
		}
	}
	if bestIdx < 0 {
		bestIdx = 0 // 江湖把式 0 耗
	}
	// 预估能否打赢：玩家每回合伤害 vs 敌人每回合伤害
	playerDmg := bestDmg + gm.pl.Atk - d.Def
	if playerDmg <= 0 {
		playerDmg = 1
	}
	turnsToKill := (f.hp + playerDmg - 1) / playerDmg
	enemyDmg := d.Atk - gm.pl.Def
	if enemyDmg <= 0 {
		enemyDmg = 1
	}
	turnsToDie := (gm.pl.HP + enemyDmg - 1) / enemyDmg
	// 打不过且能逃跑（非 Boss）
	if !d.Boss && turnsToDie < turnsToKill {
		gm.fightRun()
		return
	}
	b.fights++
	gm.fightAttack(bestIdx)
}

// ---------- 统计 ----------

func (b *Bot) report() string {
	gm := b.gm
	res := "❌超时未通关"
	if b.done {
		res = "✅通关"
	}
	return fmt.Sprintf("%s | 等级%d→%d | 境界%s | 战力%d | 战斗%d场 | 死亡%d次 | 银两%d | 用时%.1f分钟",
		res, b.startLv, gm.pl.Level, realmNames[gm.pl.Realm], gm.power(), b.fights, b.deadCount, gm.pl.Money, float64(b.frames)/30/60)
}
