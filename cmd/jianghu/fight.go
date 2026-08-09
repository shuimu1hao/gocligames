// 江湖行 - 回合制战斗系统
package main

import (
	"fmt"
	"math/rand"

	"gocligames/engine"
)

// ---------- 战斗入口 ----------

func (gm *Game) startFight(mi int) {
	d := monsterDefs[mi]
	gm.fight = &Fight{mi: mi, hp: d.HP}
	gm.mode = "fight"
	gm.push("遭遇" + d.Name + "！")
}

func (gm *Game) curMonster() *MonsterDef { return &monsterDefs[gm.fight.mi] }

// ---------- 玩家回合 ----------

func (gm *Game) fightAttack(mi int) {
	f := gm.fight
	p := &gm.pl
	m := martialDefs[gm.pl.Martials[mi]]
	if m.Cost > p.MP {
		f.msg = "内力不足，无法施展" + m.Name
		return
	}
	p.MP -= m.Cost
	d := gm.curMonster()
	atk, _, _, _ := gm.equipBonus()
	switch m.Kind {
	case "heal":
		heal := m.Heal + p.Level
		p.HP += heal
		if p.HP > p.MaxHP {
			p.HP = p.MaxHP
		}
		f.msg = fmt.Sprintf("%s！恢复 %d 点气血", m.Name, heal)
	default:
		// 命中判定
		if rand.Intn(100) >= m.Acc {
			f.msg = m.Name + "落空了！"
		} else {
			dmg := m.Base + rand.Intn(m.Rand+1) + p.Atk + atk - d.Def
			if dmg < 1 {
				dmg = 1
			}
			f.hp -= dmg
			f.msg = fmt.Sprintf("%s！造成 %d 点伤害", m.Name, dmg)
			if m.Kind == "combo" {
				for i := 1; i <= m.Combo; i++ {
					cd := m.Base/2 + rand.Intn(m.Rand/2+1) + p.Atk + atk - d.Def
					if cd < 1 {
						cd = 1
					}
					f.hp -= cd
					f.msg += fmt.Sprintf("（连击%d：%d）", i+1, cd)
				}
			}
			if m.Kind == "poison" && m.Poison > 0 {
				f.poison = m.Poison
				f.msg += " 已中毒！"
			}
		}
	}
	if f.hp <= 0 {
		gm.fightWin()
		return
	}
	gm.enemyTurn()
}

func (gm *Game) fightDefend() {
	f := gm.fight
	f.msg = "你凝神防守，气沉丹田"
	gm.enemyTurnWith(0.5)
}

func (gm *Game) fightItem(k ItemKind) {
	d := itemDefs[k]
	if d == nil || d.Quest {
		gm.fight.msg = "这个物品无法在战斗中使用"
		return
	}
	if !gm.removeItem(k) {
		gm.fight.msg = "没有" + d.Name
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
	gm.fight.msg = "使用" + d.Name + "！"
	gm.enemyTurn()
}

func (gm *Game) fightRun() {
	d := gm.curMonster()
	// 轻功影响逃跑率（Boss 不可逃）
	if d.Boss {
		gm.fight.msg = "Boss 拦路，无法逃跑！"
		return
	}
	if rand.Intn(100) < 75 {
		gm.fight = nil
		gm.mode = "map"
		gm.push("你撤身退出战斗！")
	} else {
		gm.fight.msg = "逃跑失败！"
		gm.enemyTurn()
	}
}

// ---------- 敌人回合 ----------

func (gm *Game) enemyTurn() { gm.enemyTurnWith(1.0) }

func (gm *Game) enemyTurnWith(dmgScale float64) {
	f := gm.fight
	d := gm.curMonster()
	// 玩家中毒
	if gm.pl.Poisoned > 0 {
		gm.pl.HP -= 4
		gm.pl.Poisoned--
		f.msg = "你身中剧毒，流失 4 点气血"
		if gm.pl.HP <= 0 {
			gm.pl.HP = 0
			gm.fightLose()
			return
		}
	}
	// 怪物中毒
	if f.poison > 0 {
		dmg := 6
		f.hp -= dmg
		f.poison--
		if f.hp <= 0 {
			gm.fightWin()
			return
		}
		f.msg += " 敌人受毒伤" + fmt.Sprint(dmg)
	}
	p := &gm.pl
	def, _, _, _ := gm.equipBonus()
	dmg := float64(d.Atk-p.Def-def) + float64(rand.Intn(4))
	if dmg < 1 {
		dmg = 1
	}
	dmg *= dmgScale
	gd := int(dmg)
	if gd < 1 {
		gd = 1
	}
	p.HP -= gd
	f.msg = d.Name + "攻击，你受 " + fmt.Sprint(gd) + " 点伤害"
	if p.HP <= 0 {
		p.HP = 0
		gm.fightLose()
	}
}

// ---------- 胜负 ----------

func (gm *Game) fightWin() {
	f := gm.fight
	d := gm.curMonster()
	// 标记死亡
	for _, o := range gm.objs {
		if o.Kind == SpawnMonster && o.MIdx == f.mi && !o.Dead && gm.adjacent(o) {
			o.Dead = true
			gm.dead[o.Key] = true
		}
	}
	gm.addXP(d.XP)
	gold := d.GoldMin + rand.Intn(d.GoldMax-d.GoldMin+1)
	gm.pl.Money += gold
	f.msg = fmt.Sprintf("击破%s！+%d经验 +%d两", d.Name, d.XP, gold)
	// 掉落装备
	if d.DropEquip != "" && rand.Float64() < d.DropRate {
		eq := findEquip(d.DropEquip)
		if eq != nil && !gm.hasEquip(d.DropEquip) {
			gm.addItemEquip(d.DropEquip)
			f.msg += " 掉落" + eq.Name + "！"
		}
	}
	// 支线道具掉落：灵芝
	if (f.mi == 2 || f.mi == 4 || f.mi == 8) && gm.flags["quest10_started"] && !gm.quests[10] {
		if rand.Float64() < 0.35 {
			gm.addItem(ItemHerb)
			f.msg += " 拾获千年灵芝！"
		}
	}
	// 镖局押镖：打赢拦路山贼后回镖局领赏
	if f.mi == 2 && gm.flags["quest12_started"] && !gm.quests[12] {
		gm.flags["quest12_done_fight"] = true
		gm.push("击退劫镖山贼！可回镖局领赏")
	}
	gm.fight = nil
	gm.mode = "map"
	gm.onBossKilled(f.mi)
	gm.save()
}

func (gm *Game) fightLose() {
	gm.fight = nil
	gm.mode = "dead"
	gm.recordRank()
	gm.push("你倒下了……江湖险恶，重头再来")
	gm.save()
}

func (gm *Game) hasEquip(name string) bool {
	return gm.pl.Weapon == name || gm.pl.Armor == name || gm.pl.Accessory == name
}

func (gm *Game) addItemEquip(name string) {
	// 装备直接进背包（用 ItemKind 存装备名不行，这里用特殊处理：装备名存到 Inventory 会冲突）
	// 装备掉落直接穿不上，转为银两折算
	eq := findEquip(name)
	if eq == nil {
		return
	}
	gm.pl.Money += eq.Sell / 2
	gm.push("（" + eq.Name + "不合用，折银" + fmt.Sprint(eq.Sell/2) + "两）")
}

// ---------- 按键 ----------

func (gm *Game) onKeyFight(key string) {
	switch key {
	case "1", "2", "3", "4", "5", "6":
		idx := int(key[0] - '1')
		if idx < len(gm.pl.Martials) {
			gm.fightAttack(idx)
		}
	case "d", "e":
		gm.fightDefend()
	case "x":
		gm.fightItem(ItemGold)
	case "c":
		gm.fightItem(ItemGreat)
	case "q":
		gm.fightRun()
	}
}

// ---------- 渲染 ----------

func (gm *Game) renderFight(s *engine.Screen) {
	d := gm.curMonster()
	f := gm.fight
	s.Text(1, 0, "=== 战斗 ===", engine.ColorRed)
	s.Text(1, 1, cut(d.Name, 12), d.Fg)
	pct := float64(f.hp) / float64(d.HP)
	if pct < 0 {
		pct = 0
	}
	bar := int(pct * 12)
	hpbar := ""
	for i := 0; i < 12; i++ {
		if i < bar {
			hpbar += "#"
		} else {
			hpbar += "."
		}
	}
	s.Text(1, 2, hpbar, engine.ColorRed)
	s.Text(1, 3, fmt.Sprintf("敌 HP %d/%d", f.hp, d.HP), engine.ColorRed)
	atk, def, _, _ := gm.equipBonus()
	s.Text(1, 4, fmt.Sprintf("你 HP %d/%d MP %d/%d", gm.pl.HP, gm.pl.MaxHP, gm.pl.MP, gm.pl.MaxMP), engine.ColorWhite)
	s.Text(1, 5, fmt.Sprintf("攻%d 防%d", gm.pl.Atk+atk, gm.pl.Def+def), engine.ColorWhite)
	// 招式列表
	s.Text(24, 2, "招式", engine.ColorCyan)
	for i, m := range gm.pl.Martials {
		if i >= 6 {
			break
		}
		mm := martialDefs[m]
		line := fmt.Sprintf("%d.%s(%d内)", i+1, m, mm.Cost)
		s.Text(24, 3+i, cut(line, 10), engine.ColorYellow)
	}
	s.Text(24, 10, "D防守 X金疮药 C大还丹", engine.ColorGray)
	s.Text(24, 11, "Q逃跑", engine.ColorGray)
	if f.msg != "" {
		lines := wrapText(f.msg, 10)
		for i, ln := range lines {
			if i < 3 {
				s.Text(24, 12+i, ln, engine.ColorGray)
			}
		}
	}
	// 战斗操作提示
	s.Text(1, helpY1, "1-6 出招  D防守  Q逃跑", engine.ColorYellow)
	s.Text(1, helpY2, "X金疮药  C大还丹", engine.ColorGray)
}

func (gm *Game) renderDead(s *engine.Screen) {
	s.Text(6, 4, "—— 江湖梦断 ——", engine.ColorRed)
	s.Text(6, 6, fmt.Sprintf("你倒在了 %s……", regions[gm.region].Name), engine.ColorRed)
	s.Text(6, 7, fmt.Sprintf("等级 %d · 境界 %s · 战力 %d", gm.pl.Level, realmNames[gm.pl.Realm], gm.power()), engine.ColorWhite)
	if gm.rankIn >= 0 && gm.rankIn < 10 {
		s.Text(6, 8, fmt.Sprintf("名列江湖榜第 %d 位", gm.rankIn+1), engine.ColorYellow)
	}
	s.Text(6, 10, "R 重入江湖  Q 归隐", engine.ColorGray)
}

func (gm *Game) renderWin(s *engine.Screen) {
	s.Text(4, 3, "—— 血月教覆灭 ——", engine.ColorYellow)
	s.Text(4, 5, "天山顶上，血月教主伏诛。", engine.ColorGreen)
	s.Text(4, 6, "江湖重归太平，侠名远播！", engine.ColorGreen)
	s.Text(4, 7, fmt.Sprintf("等级 %d · %s · 战力 %d", gm.pl.Level, realmNames[gm.pl.Realm], gm.power()), engine.ColorWhite)
	if gm.rankIn >= 0 && gm.rankIn < 10 {
		s.Text(4, 8, fmt.Sprintf("名列江湖榜第 %d 位！", gm.rankIn+1), engine.ColorYellow)
	}
	s.Text(4, 10, "R 重入江湖  Q 归隐", engine.ColorGray)
}
