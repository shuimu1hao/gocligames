// 江湖行 - 商店 / 客栈 / 赌场
package main

import (
	"fmt"
	"math/rand"

	"gocligames/engine"
)

// ---------- 商店 ----------

func (gm *Game) shopDef() *ShopDef {
	d := shopDefs[gm.shopID]
	return &d
}

func (gm *Game) onKeyShop(key string) {
	d := gm.shopDef()
	switch key {
	case "q":
		gm.mode = "map"
	case "1", "2", "3", "4", "5", "6", "7", "8", "9":
		idx := int(key[0] - '1')
		if gm.shopID == "luoyang_pawn" {
			gm.pawnSell(idx)
			return
		}
		if idx < len(d.Items) {
			gm.buyItem(d.Items[idx])
			return
		}
		if idx-len(d.Items) < len(d.Equips) {
			gm.buyEquip(d.Equips[idx-len(d.Items)])
		}
	}
}

func (gm *Game) buyItem(k ItemKind) {
	d := itemDefs[k]
	if gm.pl.Money < d.Price {
		gm.push("银两不足！")
		return
	}
	gm.pl.Money -= d.Price
	gm.addItem(k)
	gm.push("购入" + d.Name + "（-" + fmt.Sprint(d.Price) + "两）")
	gm.save()
}

func (gm *Game) buyEquip(name string) {
	e := findEquip(name)
	if e == nil {
		return
	}
	if gm.pl.Money < e.Price {
		gm.push("银两不足！")
		return
	}
	gm.pl.Money -= e.Price
	gm.equipByName(name)
	gm.push("购入并穿戴" + e.Name + "（-" + fmt.Sprint(e.Price) + "两）")
	gm.save()
}

func (gm *Game) equipByName(name string) {
	e := findEquip(name)
	if e == nil {
		return
	}
	switch e.Slot {
	case SlotWeapon:
		gm.pl.Weapon = name
	case SlotArmor:
		gm.pl.Armor = name
	case SlotAccessory:
		gm.pl.Accessory = name
	}
}

// 当铺：卖物品（第1-6项）和装备（第7-9项：已穿戴的）
func (gm *Game) pawnSell(idx int) {
	// 先物品
	items := gm.groupItems()
	if idx < len(items) && idx < 6 {
		it := items[idx]
		d := itemDefs[it.Kind]
		if d.Quest {
			gm.push("任务道具不能卖")
			return
		}
		gm.removeItem(it.Kind)
		gm.pl.Money += d.Sell
		gm.push("当出" + d.Name + " +" + fmt.Sprint(d.Sell) + "两")
		gm.save()
		return
	}
	// 后装备
	equips := []string{gm.pl.Weapon, gm.pl.Armor, gm.pl.Accessory}
	ei := idx - 6
	if ei >= 0 && ei < len(equips) && equips[ei] != "" {
		name := equips[ei]
		e := findEquip(name)
		switch e.Slot {
		case SlotWeapon:
			gm.pl.Weapon = ""
		case SlotArmor:
			gm.pl.Armor = ""
		case SlotAccessory:
			gm.pl.Accessory = ""
		}
		gm.pl.Money += e.Sell
		gm.push("当出" + name + " +" + fmt.Sprint(e.Sell) + "两")
		gm.save()
		return
	}
	gm.push("没有这个货物")
}

func (gm *Game) renderShop(s *engine.Screen) {
	d := gm.shopDef()
	s.Text(1, 0, "=== "+d.Name+" ===", engine.ColorCyan)
	s.Text(1, 1, "Q 离开", engine.ColorGray)
	s.Text(1, 2, cut(d.Desc, 20), engine.ColorYellow)
	s.Text(1, 3, fmt.Sprintf("银两：%d 两", gm.pl.Money), engine.ColorWhite)
	y := 5
	if gm.shopID == "luoyang_pawn" {
		gm.renderPawnList(s, y)
		return
	}
	for i, k := range d.Items {
		it := itemDefs[k]
		line := fmt.Sprintf("%d. %s %d两", i+1, it.Name, it.Price)
		s.Text(2, y+i, cut(line, 13), engine.ColorWhite)
	}
	for i, en := range d.Equips {
		e := findEquip(en)
		line := fmt.Sprintf("%d. %s(攻%d防%d) %d两", len(d.Items)+i+1, e.Name, e.Atk, e.Def, e.Price)
		s.Text(2, y+len(d.Items)+i, cut(line, 14), engine.ColorYellow)
	}
	s.Text(2, 15, "数字键购买", engine.ColorGray)
}

func (gm *Game) renderPawnList(s *engine.Screen, y int) {
	items := gm.groupItems()
	s.Text(2, y, "【物品】", engine.ColorYellow)
	yy := y + 1
	for i, it := range items {
		if i >= 6 {
			break
		}
		d := itemDefs[it.Kind]
		line := fmt.Sprintf("%d. %s x%d 当%d两", i+1, it.Name, it.Num, d.Sell)
		s.Text(2, yy+i, cut(line, 14), engine.ColorWhite)
	}
	yy += 6
	s.Text(2, yy, "【装备】", engine.ColorYellow)
	equips := []string{gm.pl.Weapon, gm.pl.Armor, gm.pl.Accessory}
	for i, en := range equips {
		if en == "" {
			continue
		}
		e := findEquip(en)
		line := fmt.Sprintf("%d. %s 当%d两", 7+i, e.Name, e.Sell)
		s.Text(2, yy+1+i, cut(line, 14), engine.ColorYellow)
	}
	s.Text(2, 15, "数字键出售", engine.ColorGray)
}

// ---------- 客栈 ----------

func (gm *Game) onKeyInn(key string) {
	switch key {
	case "1", "e", "enter":
		if gm.pl.Money < gm.innCost {
			gm.push("银两不足，住不起店")
			gm.mode = "map"
			return
		}
		gm.pl.Money -= gm.innCost
		gm.pl.HP = gm.pl.MaxHP
		gm.pl.MP = gm.pl.MaxMP
		gm.pl.Poisoned = 0
		gm.respawnWorld()
		gm.push("在" + gm.innName + "休息一晚，神清气爽（-" + fmt.Sprint(gm.innCost) + "两）")
		gm.save()
		gm.mode = "map"
	case "q":
		gm.mode = "map"
	}
}

func (gm *Game) renderInn(s *engine.Screen) {
	s.Text(1, 0, "=== "+gm.innName+" ===", engine.ColorCyan)
	s.Text(1, 2, "客栈小二：客官请！", engine.ColorYellow)
	s.Text(1, 3, fmt.Sprintf("住宿 %d 两，恢复全部气血内力", gm.innCost), engine.ColorWhite)
	s.Text(1, 5, "1 住店   Q 离开", engine.ColorGray)
}

// ---------- 赌场 ----------

func (gm *Game) onKeyCasino(key string) {
	switch key {
	case "1", "2":
		bet := 20
		if gm.casinoState == 1 {
			bet = 50
		}
		if gm.pl.Money < bet {
			gm.push("银两不足！")
			gm.mode = "map"
			return
		}
		a := 1 + rand.Intn(6)
		b := 1 + rand.Intn(6)
		c := 1 + rand.Intn(6)
		sum := a + b + c
		big := sum >= 11
		wantBig := key == "1"
		won := big == wantBig
		gm.pl.Money -= bet
		if won {
			gm.pl.Money += bet * 2
			gm.push(fmt.Sprintf("骰子 %d+%d+%d=%d 你押对了！+%d两", a, b, c, sum, bet))
		} else {
			gm.push(fmt.Sprintf("骰子 %d+%d+%d=%d 押错了 -%d两", a, b, c, sum, bet))
		}
		gm.save()
		gm.mode = "map"
	case "3", "q":
		gm.mode = "map"
	}
}

func (gm *Game) renderCasino(s *engine.Screen) {
	s.Text(1, 0, "=== 洛阳赌坊 ===", engine.ColorCyan)
	s.Text(1, 2, "赌场老板：买定离手！", engine.ColorYellow)
	s.Text(1, 3, "规则：3颗骰子 11+ 为大，10- 为小", engine.ColorWhite)
	s.Text(1, 4, fmt.Sprintf("你现有 %d 两", gm.pl.Money), engine.ColorWhite)
	if gm.casinoState == 0 {
		s.Text(1, 6, "1 押大(20两)  2 押小(20两)", engine.ColorGray)
		s.Text(1, 7, "Q 离开", engine.ColorGray)
	} else {
		s.Text(1, 6, "1 押大(50两)  2 押小(50两)", engine.ColorGray)
		s.Text(1, 7, "Q 离开", engine.ColorGray)
	}
}
