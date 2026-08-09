// 江湖行 - 武侠开放世界 RPG（gocligames 引擎）
// 数据定义：物品 / 装备 / 武功 / 门派 / 怪物 / 任务
package main

import "gocligames/engine"

// ===== 物品 =====
type ItemKind string

const (
	ItemGold   ItemKind = "gold"   // 金疮药
	ItemGreat  ItemKind = "great"  // 大还丹
	ItemMana   ItemKind = "mana"   // 内力丹
	ItemHerb   ItemKind = "herb"   // 草药（支线）
	ItemWine   ItemKind = "wine"   // 女儿红（支线）
	ItemLetter ItemKind = "letter" // 密信（主线）
	ItemJade   ItemKind = "jade"   // 暖玉（主线）
	ItemToken  ItemKind = "token"  // 血月令牌（主线）
	ItemScroll ItemKind = "scroll" // 天蚕神功（主线）
)

// ItemDef 物品定义（背包/商店数据表）。
type ItemDef struct {
	Kind  ItemKind
	Name  string
	Desc  string
	Price int
	Sell  int
	Quest bool
	HP    int
	MP    int
}

var itemDefs = map[ItemKind]*ItemDef{
	ItemGold:   {ItemGold, "金疮药", "疗伤圣药，恢复60点气血", 30, 15, false, 60, 0},
	ItemGreat:  {ItemGreat, "大还丹", "少林秘药，恢复150点气血", 90, 45, false, 150, 0},
	ItemMana:   {ItemMana, "内力丹", "回复40点内力", 50, 25, false, 0, 40},
	ItemHerb:   {ItemHerb, "千年灵芝", "药铺掌柜重金收购", 60, 30, false, 0, 0},
	ItemWine:   {ItemWine, "女儿红", "洛阳陈酿，丐帮帮主的最爱", 40, 20, false, 0, 0},
	ItemLetter: {ItemLetter, "密信", "血月教密信，字迹诡谲", 0, 0, true, 0, 0},
	ItemJade:   {ItemJade, "暖玉", "古墓中发现的奇异暖玉", 0, 0, true, 0, 0},
	ItemToken:  {ItemToken, "血月令牌", "刻着血色弯月的令牌", 0, 0, true, 0, 0},
	ItemScroll: {ItemScroll, "天蚕神功残卷", "传说中的绝世心法", 0, 0, true, 0, 0},
}

// ===== 装备 =====
type EquipSlot int

const (
	SlotWeapon EquipSlot = iota
	SlotArmor
	SlotAccessory
)

func (e EquipSlot) String() string {
	switch e {
	case SlotWeapon:
		return "兵器"
	case SlotArmor:
		return "护甲"
	default:
		return "饰品"
	}
}

// EquipDef 装备定义（穿戴属性加成）。
type EquipDef struct {
	Name  string
	Slot  EquipSlot
	Atk   int
	Def   int
	HP    int
	MP    int
	Price int
	Sell  int
	Drop  bool
}

var equipDefs = []EquipDef{
	{"木剑", SlotWeapon, 2, 0, 0, 0, 20, 10, false},
	{"铁剑", SlotWeapon, 5, 0, 0, 0, 120, 60, true},
	{"青锋剑", SlotWeapon, 10, 0, 0, 0, 400, 200, true},
	{"龙泉宝剑", SlotWeapon, 16, 0, 0, 0, 1000, 500, true},
	{"屠龙刀", SlotWeapon, 22, 0, 0, 0, 2200, 1100, true},
	{"布衣", SlotArmor, 0, 2, 0, 0, 30, 15, false},
	{"皮甲", SlotArmor, 0, 5, 0, 0, 150, 75, true},
	{"铁甲", SlotArmor, 0, 9, 0, 0, 450, 225, true},
	{"金丝软甲", SlotArmor, 0, 14, 0, 0, 1100, 550, true},
	{"玄铁重甲", SlotArmor, 0, 20, 0, 0, 2400, 1200, true},
	{"护身符", SlotAccessory, 0, 1, 20, 0, 100, 50, true},
	{"玉镯", SlotAccessory, 0, 2, 40, 0, 300, 150, true},
	{"玉佩", SlotAccessory, 0, 3, 60, 10, 800, 400, true},
	{"夜明珠", SlotAccessory, 0, 4, 100, 20, 1800, 900, true},
}

func findEquip(name string) *EquipDef {
	for i := range equipDefs {
		if equipDefs[i].Name == name {
			return &equipDefs[i]
		}
	}
	return nil
}

// ===== 武功 =====
type Martial struct {
	Name   string
	Cost   int
	Base   int
	Rand   int
	Acc    int
	Kind   string // normal strong combo heal poison
	Heal   int
	Poison int
	Combo  int
	Desc   string
}

var martialDefs = map[string]Martial{
	"江湖把式":  {"江湖把式", 0, 6, 4, 90, "normal", 0, 0, 0, "市井拳脚，无门无派"},
	"基础剑法":  {"基础剑法", 3, 8, 4, 90, "normal", 0, 0, 0, "江湖通用入门剑法"},
	"基础内功":  {"基础内功", 5, 0, 0, 100, "heal", 20, 0, 0, "运功调息，恢复20点气血"},
	"少林罗汉拳": {"少林罗汉拳", 6, 14, 6, 92, "normal", 0, 0, 0, "少林入门拳法，刚猛直接"},
	"少林金刚掌": {"少林金刚掌", 12, 22, 8, 88, "strong", 0, 0, 0, "金刚怒目，掌力排山倒海"},
	"易筋经":   {"易筋经", 10, 0, 0, 100, "heal", 60, 0, 0, "少林至高内功，疗伤奇效"},
	"武当绵掌":  {"武当绵掌", 6, 13, 5, 95, "normal", 0, 0, 0, "以柔克刚，连绵不绝"},
	"太极剑法":  {"太极剑法", 12, 20, 8, 90, "combo", 0, 0, 2, "借力打力，一剑化三剑"},
	"武当九阳功": {"武当九阳功", 10, 0, 0, 100, "heal", 55, 0, 0, "纯阳真气，生生不息"},
	"打狗棒法":  {"打狗棒法", 6, 15, 6, 92, "combo", 0, 0, 2, "丐帮绝学，棍影重重"},
	"降龙十八掌": {"降龙十八掌", 14, 26, 10, 85, "strong", 0, 0, 0, "天下至刚至阳的掌法"},
	"逍遥游身法": {"逍遥游身法", 8, 0, 0, 100, "heal", 45, 0, 0, "逍遥派心法，游刃有余"},
	"五毒掌":   {"五毒掌", 8, 10, 5, 85, "poison", 0, 6, 0, "掌风带毒，中者浑身溃烂"},
	"血月大法":  {"血月大法", 15, 30, 12, 80, "strong", 0, 0, 0, "血月教禁术，吸取生人精血"},
	"天蚕神功":  {"天蚕神功", 12, 34, 14, 95, "combo", 0, 0, 3, "丝雨绵绵，一剑三影"},
}

// ===== 门派 =====
type Sect struct {
	Name     string
	Master   string
	MasterCh rune
	Fg       int
	Martials []string
	NeedLv   int
}

var sects = []Sect{
	{"少林寺", "方丈玄慈", '方', engine.ColorYellow, []string{"少林罗汉拳", "少林金刚掌", "易筋经"}, 1},
	{"武当派", "掌门冲虚", '掌', engine.ColorCyan, []string{"武当绵掌", "太极剑法", "武当九阳功"}, 1},
	{"丐帮", "帮主洪七公", '帮', engine.ColorOrange, []string{"打狗棒法", "降龙十八掌", "逍遥游身法"}, 1},
}

// ===== 怪物 =====
type MonsterDef struct {
	Name      string
	Ch        rune
	Fg        int
	HP, Atk   int
	Def       int
	XP        int
	GoldMin   int
	GoldMax   int
	DropEquip string
	DropRate  float64
	Boss      bool
	Poison    int
}

var monsterDefs = []MonsterDef{
	{"野狗", '犬', 94, 14, 4, 0, 8, 1, 3, "", 0, false, 0},
	{"野狼", '狼', engine.ColorGray, 22, 6, 1, 14, 2, 5, "", 0, false, 0},
	{"山贼", '贼', engine.ColorRed, 30, 7, 2, 20, 3, 8, "铁剑", 0.15, false, 0},
	{"毒蛇", '蛇', engine.ColorGreen, 18, 8, 0, 18, 2, 6, "", 0, false, 6},
	{"黑风寨兵", '兵', engine.ColorOrange, 52, 11, 4, 40, 8, 16, "皮甲", 0.12, false, 0},
	{"黑风寨主", '寨', engine.ColorRed, 140, 15, 6, 120, 30, 50, "青锋剑", 1.0, true, 0},
	{"古墓尸卫", '尸', engine.ColorGray, 60, 13, 5, 45, 10, 20, "铁甲", 0.10, false, 0},
	{"尸王", '王', engine.ColorMagenta, 200, 20, 8, 200, 50, 80, "龙泉宝剑", 1.0, true, 0},
	{"血月教徒", '教', engine.ColorMagenta, 70, 15, 5, 60, 12, 24, "护身符", 0.15, false, 0},
	{"血月护法", '护', engine.ColorRed, 120, 20, 8, 120, 25, 45, "金丝软甲", 0.5, false, 0},
	{"血月教主", '尊', engine.ColorMagenta, 320, 26, 10, 400, 100, 200, "屠龙刀", 1.0, true, 0},
	{"血月余孽", '余', engine.ColorMagenta, 90, 20, 6, 120, 30, 60, "夜明珠", 0.3, false, 0},
}

// ===== 任务 =====
type QuestDef struct {
	ID   int
	Name string
	Desc string
}

// 支线任务定义（主线用 mainProgress 记录）
var questDefs = []QuestDef{
	{10, "药铺急单", "帮青州城药铺掌柜采一株千年灵芝"},
	{11, "洛阳乞酒", "给丐帮帮主带一坛女儿红"},
	{12, "古墓疑云", "取回古墓中的暖玉交给洛阳城捕头"},
}

// ===== 商店 =====
type ShopDef struct {
	Name string
	Desc string
	// 商品：物品 kind 或装备名
	Items  []ItemKind
	Equips []string
}

var shopDefs = map[string]ShopDef{
	"qingzhou_shop": {"青州药铺", "掌柜：客官，疗伤灵药应有尽有", []ItemKind{ItemGold, ItemGreat, ItemMana}, nil},
	"luoyang_shop":  {"洛阳兵器铺", "铁匠：好兵器配好汉，看看？", []ItemKind{ItemGold, ItemGreat, ItemMana}, []string{"木剑", "铁剑", "布衣", "皮甲", "护身符"}},
	"luoyang_pawn":  {"洛阳当铺", "朝奉：死当活当，童叟无欺", []ItemKind{}, nil},
}
