// 江湖行 - 菜单界面（背包/装备/武功/任务/江湖榜/帮助/奇遇）
package main

import (
	"fmt"

	"gocligames/engine"
)

var menuTitles = []string{"背包", "装备", "武功", "任务", "江湖榜", "帮助"}

// 物品分组（去重）
type itemGroup struct {
	Kind ItemKind
	Name string
	Num  int
}

func (gm *Game) groupItems() []itemGroup {
	seen := map[ItemKind]int{}
	for _, k := range gm.pl.Inventory {
		seen[k]++
	}
	var out []itemGroup
	// 固定顺序：消耗品在前，任务道具在后
	for _, k := range []ItemKind{ItemGold, ItemGreat, ItemMana, ItemHerb, ItemWine, ItemLetter, ItemJade, ItemToken, ItemScroll} {
		if seen[k] > 0 {
			out = append(out, itemGroup{k, itemDefs[k].Name, seen[k]})
		}
	}
	return out
}

// ---------- 按键 ----------

func (gm *Game) onKeyMenu(key string) {
	switch key {
	case "q", "esc":
		gm.mode = "map"
	case "e", "enter", "space":
		if gm.menuTab == 0 {
			items := gm.groupItems()
			if gm.menuIdx >= 0 && gm.menuIdx < len(items) {
				gm.useItem(items[gm.menuIdx].Kind)
			}
		}
	case "w", "up":
		gm.menuIdx--
		if gm.menuIdx < 0 {
			gm.menuIdx = 0
		}
	case "s", "down":
		gm.menuIdx++
		gm.clampMenuIdx()
	}
}

func (gm *Game) clampMenuIdx() {
	if gm.menuTab == 0 {
		if gm.menuIdx >= len(gm.groupItems()) {
			gm.menuIdx = len(gm.groupItems()) - 1
			if gm.menuIdx < 0 {
				gm.menuIdx = 0
			}
		}
	} else {
		if gm.menuIdx > 0 {
			gm.menuIdx = 0
		}
	}
}

// ---------- 渲染 ----------

func (gm *Game) renderMenu(s *engine.Screen) {
	s.Text(1, 0, "=== "+menuTitles[gm.menuTab]+" ===", engine.ColorCyan)
	s.Text(1, 1, "Q 返回", engine.ColorGray)
	switch gm.menuTab {
	case 0:
		gm.renderBag(s)
	case 1:
		gm.renderEquip(s)
	case 2:
		gm.renderMartials(s)
	case 3:
		gm.renderQuests(s)
	case 4:
		gm.renderRank(s)
	case 5:
		gm.renderHelpMenu(s)
	}
}

func (gm *Game) renderBag(s *engine.Screen) {
	items := gm.groupItems()
	if len(items) == 0 {
		s.Text(2, 3, "背包空空如也", engine.ColorGray)
		return
	}
	for i, it := range items {
		if i >= 12 {
			break
		}
		mark := " "
		if i == gm.menuIdx {
			mark = ">"
		}
		line := fmt.Sprintf("%s%d.%s x%d", mark, i+1, it.Name, it.Num)
		s.Text(2, 3+i, cut(line, 12), engine.ColorWhite)
	}
	s.Text(2, 16, "W/S 选择  E 使用", engine.ColorYellow)
}

func (gm *Game) renderEquip(s *engine.Screen) {
	p := &gm.pl
	atk, def, hp, mp := gm.equipBonus()
	s.Text(2, 3, fmt.Sprintf("兵器: %s", orEmpty(p.Weapon, "赤手空拳")), engine.ColorWhite)
	s.Text(2, 4, fmt.Sprintf("护甲: %s", orEmpty(p.Armor, "布衣粗衫")), engine.ColorWhite)
	s.Text(2, 5, fmt.Sprintf("饰品: %s", orEmpty(p.Accessory, "无")), engine.ColorWhite)
	s.Text(2, 7, "装备加成：", engine.ColorYellow)
	s.Text(2, 8, fmt.Sprintf("攻击+%d 防御+%d", atk, def), engine.ColorYellow)
	s.Text(2, 9, fmt.Sprintf("气血+%d 内力+%d", hp, mp), engine.ColorYellow)
	s.Text(2, 11, "（装备可在洛阳兵器铺购买）", engine.ColorGray)
}

func orEmpty(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

func (gm *Game) renderMartials(s *engine.Screen) {
	if len(gm.pl.Martials) == 0 {
		s.Text(2, 3, "尚未习得任何武功", engine.ColorGray)
		return
	}
	for i, m := range gm.pl.Martials {
		if i >= 12 {
			break
		}
		mm := martialDefs[m]
		line := fmt.Sprintf("%d.%s 耗内%d", i+1, m, mm.Cost)
		s.Text(2, 3+i, cut(line, 11), engine.ColorWhite)
	}
	// 描述区
	if len(gm.pl.Martials) > 0 {
		m := martialDefs[gm.pl.Martials[0]]
		desc := wrapText(m.Desc, 9)
		for j, ln := range desc {
			if j < 3 {
				s.Text(24, 4+j, ln, engine.ColorGray)
			}
		}
	}
}

func (gm *Game) renderQuests(s *engine.Screen) {
	s.Text(2, 3, "【主线】", engine.ColorYellow)
	mainDesc := ""
	switch gm.mainProg {
	case 0, 1:
		mainDesc = "前往黑风寨，取回寨主身上的密信"
	case 2:
		mainDesc = "将密信带回青州城交给师父"
	case 3:
		mainDesc = "前往幽暗古墓，寻天蚕神功"
	case 4:
		mainDesc = "持暖玉前往洛阳城见捕头"
	case 5:
		mainDesc = "天山顶决战血月教主！"
	default:
		mainDesc = "血月教覆灭，江湖太平"
	}
	lines := wrapText(mainDesc, 10)
	for i, ln := range lines {
		if i < 4 {
			s.Text(2, 5+i, ln, engine.ColorWhite)
		}
	}
	s.Text(2, 10, "【支线】", engine.ColorYellow)
	sub := 0
	for _, q := range questDefs {
		y := 12 + sub
		if y > 15 {
			break
		}
		state := "未接"
		if gm.quests[q.ID] {
			state = "完成"
		} else if gm.flags[fmt.Sprintf("quest%d_started", q.ID%100)] || gm.questStarted(q.ID) {
			state = "进行中"
		}
		s.Text(2, y, fmt.Sprintf("%s：%s", q.Name, state), engine.ColorWhite)
		sub++
	}
}

func (gm *Game) questStarted(id int) bool {
	switch id {
	case 10:
		return gm.flags["quest10_started"]
	case 11:
		return gm.flags["quest11_started"]
	case 12:
		return gm.flags["quest12_started"]
	}
	return false
}

func (gm *Game) renderRank(s *engine.Screen) {
	top := gm.sb.Top(10)
	if len(top) == 0 {
		s.Text(2, 4, "江湖榜虚位以待……", engine.ColorGray)
		return
	}
	for i, r := range top {
		if i >= 10 {
			break
		}
		line := fmt.Sprintf("%d.%s 战力%d", i+1, cut(r.Name, 4), r.Score)
		s.Text(2, 3+i, cut(line, 12), engine.ColorYellow)
	}
}

func (gm *Game) renderHelpMenu(s *engine.Screen) {
	helps := []string{
		"WASD 移动    E 交互/对话",
		"I 背包  F 武功  T 任务",
		"B 江湖榜  H 帮助",
		"1 金疮药  2 大还丹(快捷)",
		"Q 退出（自动存档）",
		"",
		"战斗：1-6 出招  D 防守",
		"X 金疮药  C 大还丹  Q 逃跑",
		"",
		"主线：黑风寨→古墓→天山",
		"支线：药铺/丐帮/镖局",
	}
	for i, h := range helps {
		s.Text(2, 3+i, h, engine.ColorWhite)
	}
}

// ---------- 奇遇渲染 ----------

func (gm *Game) renderEnc(s *engine.Screen) {
	if gm.talkI < len(gm.talk) {
		lines := wrapText(gm.talk[gm.talkI], 10)
		for i, ln := range lines {
			if infoY+i < 13 {
				s.Text(infoX, infoY+i, ln, engine.ColorYellow)
			}
		}
		s.Text(infoX, 12, "Enter 继续", engine.ColorGray)
	}
}

// ---------- 标题画面 ----------

func (gm *Game) renderIntro(s *engine.Screen) {
	art := []string{
		"   JIANGHU  ·  江  湖  行  ",
		"   武侠开放世界 RPG v1.0",
		"",
		"   血月教崛起，江湖动荡。",
		"   你，一个无名小虾米，",
		"   将一步步踏上绝世之路。",
		"",
		"   主线：黑风寨→古墓→天山",
		"   支线：药铺·丐帮·镖局",
		"   门派：少林·武当·丐帮",
		"",
		"   按任意键开始江湖之旅",
	}
	for i, line := range art {
		s.Text(6, 2+i, line, engine.ColorYellow)
	}
	s.Text(6, 17, "WASD移动 E交互 I背包 F武功 T任务", engine.ColorGray)
	s.Text(6, 18, "B江湖榜 H帮助 Q退出", engine.ColorGray)
}
