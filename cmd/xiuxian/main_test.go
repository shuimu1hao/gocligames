package main

import (
	"os"
	"path/filepath"
	"testing"

	"gocligames/engine"
)

func testGame(t *testing.T) *Game {
	t.Helper()
	gm := &Game{savePath: filepath.Join(t.TempDir(), "save.json")}
	gm.newGame()
	return gm
}

// 地图每行必须等宽，且是纯 ASCII 地形（不含汉字对象）
func TestMapTilesUniform(t *testing.T) {
	for li, l := range layers {
		w := len(l.Tiles[0])
		for y, row := range l.Tiles {
			if len(row) != w {
				t.Fatalf("layer %d row %d width %d != %d", li, y, len(row), w)
			}
			for _, ch := range row {
				if ch > 0x7f {
					t.Fatalf("layer %d row %d has non-ascii tile %q", li, y, ch)
				}
			}
		}
	}
}

// 出生点必须是空地
func TestSpawnPointsFree(t *testing.T) {
	gm := testGame(t)
	if gm.tileAt(gm.px, gm.py) == '#' {
		t.Fatal("出生点在墙上")
	}
}

// 撞墙不动，空地能走
func TestMoveWallAndFloor(t *testing.T) {
	gm := testGame(t)
	// 出生点 (2,2) 左边是墙，直接向左撞墙不动
	gm.px, gm.py = 2, 2
	gm.tryMove(-1, 0)
	gm.tryMove(1, 0)
	if gm.px != 2 {
		t.Fatalf("撞墙应该不动，px=%d", gm.px)
	}
	gm.tryMove(-1, 0)
	if gm.px != 1 {
		t.Fatalf("空地应该能走，px=%d", gm.px)
	}
}

// 踩物品自动拾取
func TestPickupItem(t *testing.T) {
	gm := testGame(t)
	// 第 1 层 (8,8) 有一枚回元丹
	gm.px, gm.py = 8, 8
	before := gm.pl.Pills
	gm.tryMove(1, 0)
	if gm.pl.Pills != before+1 {
		t.Fatalf("拾取后丹药应 +1，got %d", gm.pl.Pills)
	}
	if !gm.opened[keyOf(0, 9, 8)] {
		t.Fatal("拾取应记录 opened")
	}
}

// 楼梯换层
func TestStairsDown(t *testing.T) {
	gm := testGame(t)
	gm.px, gm.py = 10, 10 // 第 1 层楼梯 > 在 (11,10)
	gm.tryMove(1, 0)
	if gm.layer != 1 {
		t.Fatalf("走下楼梯应到第 2 层，got %d", gm.layer)
	}
	if gm.px != startX || gm.py != startY {
		t.Fatalf("换层后应回到出生点 (%d,%d)", gm.px, gm.py)
	}
	// 存档应已写入
	if _, err := os.Stat(gm.savePath); err != nil {
		t.Fatal("换层应自动存档")
	}
}

// 打小怪：攻击到死 → 奖励 + 存档 + 对象消失
func TestFightWin(t *testing.T) {
	gm := testGame(t)
	o := gm.objAt(3, 9) // 第 1 层蛇妖
	if o == nil {
		t.Fatal("第 1 层应有蛇妖")
	}
	gm.startFight(o)
	if gm.mode != "fight" {
		t.Fatal("遭遇怪应进入战斗")
	}
	// 玩家攻 8 防 2 vs 蛇 hp22 atk5 def1：伤害 7-10，最多 4 刀，反击 3-5 打不死
	for i := 0; i < 6 && gm.mode == "fight"; i++ {
		gm.fightAttack()
	}
	if gm.mode != "map" {
		t.Fatalf("蛇妖应被打败，mode=%s", gm.mode)
	}
	if !o.Dead {
		t.Fatal("蛇妖应标记死亡")
	}
	if gm.pl.XP < 15 || gm.pl.Ling < 1 {
		t.Fatalf("应有经验/灵石奖励，xp=%d ling=%d", gm.pl.XP, gm.pl.Ling)
	}
	if !gm.dead[keyOf(0, 3, 9)] {
		t.Fatal("击杀应记录存档")
	}
}

// Boss 死亡 → bossDead 标记
func TestBossDefeat(t *testing.T) {
	gm := testGame(t)
	gm.layer = 2
	gm.spawnLayer()
	gm.pl.Atk = 200 // 秒杀
	o := gm.objAt(9, 6)
	if o == nil || o.DefIdx != 3 {
		t.Fatal("第 3 层应有 Boss")
	}
	gm.startFight(o)
	for i := 0; i < 3 && gm.mode == "fight"; i++ {
		gm.fightAttack()
	}
	if !gm.bossDead {
		t.Fatal("Boss 死后应标记 bossDead")
	}
	if gm.pl.Level < 3 {
		t.Fatalf("100 经验应升到 Lv3+，got Lv%d（经验被升级消耗）", gm.pl.Level)
	}
}

// 逃跑：多试几次至少成功一次
func TestFightRun(t *testing.T) {
	gm := testGame(t)
	ok := false
	for i := 0; i < 30 && !ok; i++ {
		o := gm.objAt(3, 9)
		gm.startFight(o)
		gm.fightRun()
		if gm.mode == "map" {
			ok = true
		}
	}
	if !ok {
		t.Fatal("逃跑 30 次应至少成功一次（70% 概率）")
	}
}

// 玩家死亡 → over
func TestPlayerDeath(t *testing.T) {
	gm := testGame(t)
	gm.pl.HP = 1
	gm.pl.Def = 0
	o := gm.objAt(3, 9)
	gm.startFight(o)
	gm.fightAttack() // 蛇没死，反击必中，HP 1 必死
	if gm.mode != "over" {
		t.Fatalf("玩家应死亡进入 over，mode=%s hp=%d", gm.mode, gm.pl.HP)
	}
}

// 存档读档
func TestSaveLoad(t *testing.T) {
	gm := testGame(t)
	gm.pl.Ling = 42
	gm.pl.Level = 3
	gm.layer = 2
	gm.dead[keyOf(1, 6, 3)] = true
	gm.save()

	gm2 := &Game{savePath: gm.savePath}
	if !gm2.load() {
		t.Fatal("读档应成功")
	}
	if gm2.pl.Ling != 42 || gm2.pl.Level != 3 || gm2.layer != 2 {
		t.Fatalf("读档状态错误: %+v", gm2.pl)
	}
	if !gm2.dead[keyOf(1, 6, 3)] {
		t.Fatal("击杀记录应读回")
	}
}

// 读档后 spawnLayer 过滤已死怪
func TestLoadFiltersDead(t *testing.T) {
	gm := testGame(t)
	gm.layer = 0
	gm.dead[keyOf(0, 3, 9)] = true
	gm.spawnLayer()
	if o := gm.objAt(3, 9); o != nil && !o.Dead {
		t.Fatal("已杀怪应保持死亡")
	}
}

// 升级逻辑
func TestLevelUp(t *testing.T) {
	gm := testGame(t)
	gm.pl.XP = 19
	gm.checkLevelUp()
	if gm.pl.Level != 1 {
		t.Fatal("未满经验不应升级")
	}
	gm.pl.XP = 20
	gm.checkLevelUp()
	if gm.pl.Level != 2 {
		t.Fatalf("满 20 经验应升 2 级，got %d", gm.pl.Level)
	}
	if gm.pl.MaxHP != 65 || gm.pl.Atk != 10 || gm.pl.Def != 3 {
		t.Fatalf("升级属性错误: %+v", gm.pl)
	}
}

// 渲染冒烟：不 panic，画布尺寸正确
func TestRenderSmoke(t *testing.T) {
	g := engine.NewGame("test", scrW, scrH, 30)
	gm := testGame(t)
	gm.render(g, g.Screen)
	if g.Screen.Width() != scrW || g.Screen.Height() != scrH {
		t.Fatal("画布尺寸错误")
	}
}
