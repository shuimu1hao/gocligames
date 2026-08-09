// 江湖行 - NPC 对话 / 任务 / 奇遇
package main

import (
	"fmt"
	"math/rand"
)

// ---------- 对话入口 ----------

func (gm *Game) startTalk(name string) {
	switch name {
	case "师父李长风":
		gm.talkMaster()
	case "药铺掌柜":
		gm.talkDoctor()
	case "客栈小二":
		gm.talkInn()
	case "武馆教头":
		gm.talkWuguan()
	case "镖局镖头":
		gm.talkBiaotou()
	case "铁匠":
		gm.talkSmith()
	case "赌场老板":
		gm.talkGamble()
	case "当铺朝奉":
		gm.talkPawn()
	case "丐帮弟子":
		gm.talkGaizi()
	case "酒馆老板":
		gm.talkJiuguan()
	case "捕头":
		gm.talkButou()
	case "方丈玄慈":
		gm.talkSect(0)
	case "掌门冲虚":
		gm.talkSect(1)
	case "帮主洪七公":
		gm.talkSect(2)
	default:
		gm.talk = []string{name + "：施主有礼。"}
		gm.talkI = 0
		gm.mode = "talk"
	}
}

func (gm *Game) beginTalk(lines []string, end func()) {
	gm.talk = lines
	gm.talkI = 0
	gm.onTalkEnd = end
	gm.mode = "talk"
}

// ---------- 师父（主线） ----------

func (gm *Game) talkMaster() {
	switch {
	case gm.mainProg == 0:
		gm.beginTalk([]string{
			"李长风：徒儿，你来得正好。",
			"李长风：血月教近日与黑风寨贼寇勾结，蠢蠢欲动。",
			"李长风：你去黑风寨打探消息，若得密信，速速带回！",
			"（任务：前往黑风寨，取回寨主身上的密信）",
		}, func() {
			gm.addItem(ItemGold)
			gm.mainProg = 1
			gm.push("师父赠你金疮药×1，主线开启：黑风寨取密信")
			gm.save()
		})
	case gm.mainProg == 1:
		gm.beginTalk([]string{
			"李长风：黑风寨就在中原大地西侧，小心寨兵。",
			"李长风：若寻得密信，速速带回！",
		}, nil)
	case gm.mainProg == 2:
		if gm.countItem(ItemLetter) > 0 {
			gm.beginTalk([]string{
				"李长风：你带回了密信！（展信细看）",
				"李长风：血月教要在古墓寻找天蚕神功……",
				"李长风：绝不能让魔教得逞！去古墓一探究竟！",
				"（任务：前往幽暗古墓，寻天蚕神功）",
			}, func() {
				gm.removeItem(ItemLetter)
				gm.mainProg = 3
				gm.push("主线推进：前往古墓")
				gm.save()
			})
		} else {
			gm.beginTalk([]string{"李长风：密信呢？快去黑风寨寻来！"}, nil)
		}
	case gm.mainProg == 3:
		gm.beginTalk([]string{"李长风：古墓凶险，万事小心。"}, nil)
	case gm.mainProg == 4:
		gm.beginTalk([]string{
			"李长风：你得了天蚕神功？！",
			"李长风：那暖玉……去洛阳城找捕头，他知晓天山之事。",
			"（任务：前往洛阳城见捕头）",
		}, nil)
	case gm.mainProg == 5:
		gm.beginTalk([]string{
			"李长风：血月教主就在天山，去吧徒儿！",
			"李长风：此行凶多吉少，但江湖需要你。",
			"（任务：天山顶决战血月教主）",
		}, nil)
	default:
		gm.beginTalk([]string{"李长风：江湖有你，老夫放心了。"}, nil)
	}
}

// ---------- 药铺掌柜（支线10：采灵芝） ----------

func (gm *Game) talkDoctor() {
	started := gm.flags["quest10_started"]
	if gm.quests[10] {
		gm.beginTalk([]string{"药铺掌柜：多谢侠士！灵药已入柜，客官可随意采买。"}, nil)
		return
	}
	if gm.countItem(ItemHerb) > 0 {
		gm.beginTalk([]string{
			"药铺掌柜：千年灵芝！正是我缺的那味药！",
			"药铺掌柜：这是 200 两谢银，外加两粒大还丹，请收下！",
		}, func() {
			gm.removeItem(ItemHerb)
			gm.pl.Money += 200
			gm.addItem(ItemGreat)
			gm.addItem(ItemGreat)
			gm.quests[10] = true
			gm.push("支线完成：药铺急单")
			gm.save()
		})
		return
	}
	if !started {
		gm.beginTalk([]string{
			"药铺掌柜：哎哟，最近药材紧俏得很。",
			"药铺掌柜：若是侠士能采一株千年灵芝给我，必有重谢！",
			"（支线：山贼与黑风寨兵身上或可寻得灵芝）",
		}, func() {
			gm.flags["quest10_started"] = true
			gm.push("支线领取：药铺急单")
			gm.save()
		})
	} else {
		gm.beginTalk([]string{"药铺掌柜：灵芝可曾寻到？山贼身上常有。"}, nil)
	}
}

// ---------- 客栈 ----------

func (gm *Game) talkInn() {
	who := ""
	for _, o := range gm.objs {
		if o.Kind == SpawnNPC && o.NPC == "客栈小二" && gm.adjacent(o) {
			if gm.region == "qingzhou" {
				who = "青州客栈"
			} else {
				who = "洛阳客栈"
			}
		}
	}
	if who == "" {
		who = "客栈"
	}
	gm.beginTalk([]string{
		"客栈小二：客官住店吗？20 两一晚，包您睡得舒坦。",
		"（休息可恢复全部气血内力并自动存档）",
	}, func() {
		gm.innName = who
		gm.mode = "inn"
		gm.innCost = 20
	})
}

// ---------- 武馆教头 ----------

func (gm *Game) talkWuguan() {
	if !gm.hasMartial("基础剑法") {
		gm.beginTalk([]string{
			"武馆教头：小兄弟底子不错，教你几路基础剑法如何？",
			"武馆教头：看好了！",
		}, func() {
			gm.learnMartial("基础剑法")
			gm.push("学会武功：基础剑法")
			gm.save()
		})
		return
	}
	gm.beginTalk([]string{"武馆教头：江湖险恶，勤练武功才是正道。"}, nil)
}

// ---------- 镖局镖头（支线12：押镖） ----------

func (gm *Game) talkBiaotou() {
	if gm.quests[12] {
		gm.beginTalk([]string{"镖局镖头：上回多谢侠士，回头客好说！"}, nil)
		return
	}
	if gm.flags["quest12_done_fight"] {
		gm.beginTalk([]string{
			"镖局镖头：好身手！那群山贼被你打得落花流水！",
			"镖局镖头：这是 120 两镖银，聊表谢意！",
		}, func() {
			gm.pl.Money += 120
			gm.quests[12] = true
			gm.push("支线完成：镖局押镖 +120两")
			gm.save()
		})
		return
	}
	gm.beginTalk([]string{
		"镖局镖头：正要押一镖货去洛阳，路上有山贼出没……",
		"镖局镖头：侠士若能护送，必有厚报！",
		"（支线：护送镖车，击退拦路山贼）",
	}, func() {
		gm.flags["quest12_started"] = true
		gm.startFight(2) // 山贼
	})
}

// ---------- 铁匠（商店） ----------

func (gm *Game) talkSmith() {
	gm.beginTalk([]string{
		"铁匠：好兵器配好汉，客官看看？",
		"（打开洛阳兵器铺）",
	}, func() {
		gm.shopID = "luoyang_shop"
		gm.mode = "shop"
	})
}

// ---------- 赌场老板 ----------

func (gm *Game) talkGamble() {
	gm.beginTalk([]string{
		"赌场老板：客官，来玩两把骰子？押大押小，一赔一！",
		"（3 颗骰子 11 点以上为大，10 点以下为小）",
	}, func() {
		gm.casinoState = 0
		gm.mode = "casino"
	})
}

// ---------- 当铺朝奉 ----------

func (gm *Game) talkPawn() {
	gm.beginTalk([]string{
		"当铺朝奉：死当活当，童叟无欺。",
		"（打开当铺，可出售物品与装备）",
	}, func() {
		gm.shopID = "luoyang_pawn"
		gm.mode = "shop"
	})
}

// ---------- 丐帮弟子（支线11：乞酒） ----------

func (gm *Game) talkGaizi() {
	if gm.quests[11] {
		gm.beginTalk([]string{"丐帮弟子：多谢侠士的美酒，帮主他老人家高兴得很！"}, nil)
		return
	}
	if gm.countItem(ItemWine) > 0 {
		gm.beginTalk([]string{
			"丐帮弟子：女儿红？！帮主他老人家就好这一口！",
			"丐帮弟子：这 300 两银票您收好，还有两粒内力丹！",
		}, func() {
			gm.removeItem(ItemWine)
			gm.pl.Money += 300
			gm.addItem(ItemMana)
			gm.addItem(ItemMana)
			gm.quests[11] = true
			gm.push("支线完成：洛阳乞酒 +300两")
			gm.save()
		})
		return
	}
	gm.beginTalk([]string{
		"丐帮弟子：帮主最近闷闷不乐，就惦记洛阳酒馆的女儿红。",
		"丐帮弟子：侠士若肯带一坛来，帮主必有重谢！",
		"（支线：去酒馆老板处买一坛女儿红）",
	}, func() {
		gm.flags["quest11_started"] = true
		gm.push("支线领取：洛阳乞酒")
		gm.save()
	})
}

// ---------- 酒馆老板 ----------

func (gm *Game) talkJiuguan() {
	if gm.countItem(ItemWine) > 0 {
		gm.beginTalk([]string{"酒馆老板：客官，女儿红已经买过了，慢用！"}, nil)
		return
	}
	gm.beginTalk([]string{
		"酒馆老板：本店招牌女儿红，40 两一坛，要来一坛吗？",
	}, func() {
		if gm.pl.Money >= 40 {
			gm.pl.Money -= 40
			gm.addItem(ItemWine)
			gm.push("买下女儿红一坛（-40两）")
			gm.save()
		} else {
			gm.push("银两不足，买不起女儿红")
		}
	})
}

// ---------- 捕头（主线衔接） ----------

func (gm *Game) talkButou() {
	switch {
	case gm.mainProg == 4:
		if gm.countItem(ItemJade) > 0 {
			gm.beginTalk([]string{
				"捕头：这暖玉！是血月教主的信物！",
				"捕头：教主藏身天山顶，凭此玉可寻其踪迹。",
				"捕头：官府不便出手，侠士保重！",
				"（任务：前往天山顶决战血月教主）",
			}, func() {
				gm.removeItem(ItemJade)
				gm.mainProg = 5
				gm.push("主线推进：决战天山")
				gm.save()
			})
		} else {
			gm.beginTalk([]string{"捕头：古墓中的暖玉是破案关键，侠士可曾寻得？"}, nil)
		}
	case gm.mainProg == 3:
		gm.beginTalk([]string{
			"捕头：听说古墓有异宝出世，侠士小心为上。",
		}, nil)
	case gm.mainProg == 5:
		gm.beginTalk([]string{"捕头：天山凶险，血月教高手如云，万万小心！"}, nil)
	case gm.mainProg >= 6:
		gm.beginTalk([]string{"捕头：血月教覆灭，天下太平！侠士之名，当传遍江湖！"}, nil)
	default:
		gm.beginTalk([]string{"捕头：这位侠士面生得很，可是初到中原？"}, nil)
	}
}

// ---------- 门派（拜师） ----------

func (gm *Game) talkSect(idx int) {
	st := sects[idx]
	if gm.pl.Sect != "" && gm.pl.Sect != st.Name {
		gm.beginTalk([]string{st.Master + "：你已另投他门，贫僧/贫道不便再授。"}, nil)
		return
	}
	if gm.pl.Sect == st.Name {
		gm.beginTalk([]string{st.Master + "：勤加修炼，莫负师门。"}, nil)
		return
	}
	gm.beginTalk([]string{
		st.Master + "：贫" + sectTitle(st.Name) + "见你骨骼清奇，可愿拜入" + st.Name + "门下？",
		"（拜师可习得" + st.Name + "绝学）",
	}, func() {
		gm.pl.Sect = st.Name
		for _, m := range st.Martials {
			gm.learnMartial(m)
		}
		gm.push("拜入" + st.Name + "！学会师门武功")
		gm.save()
	})
}

func sectTitle(name string) string {
	switch name {
	case "少林寺":
		return "衲"
	case "武当派":
		return "道"
	default:
		return "丐"
	}
}

// ---------- 奇遇 ----------

func (gm *Game) triggerEncounter(key string) {
	r := rand.Intn(100)
	gm.encID = key
	switch {
	case r < 35:
		gm.encOpts = []string{"收下秘籍", "婉言谢绝"}
		gm.beginTalk([]string{
			"路遇一名老乞丐：小友，老夫看你筋骨奇佳……",
			"老乞丐塞给你一卷泛黄的秘籍（五毒掌）",
		}, func() {
			gm.flags["enc_"+key] = true
			gm.learnMartial("五毒掌")
			gm.push("学会武功：五毒掌")
			gm.save()
		})
	case r < 55:
		gm.encOpts = nil
		gm.beginTalk([]string{
			"草丛中窜出一只野狼，拦住了你的去路！",
		}, func() {
			gm.flags["enc_"+key] = true
			gm.startFight(1)
		})
	case r < 75:
		gm.encOpts = nil
		gm.beginTalk([]string{
			"你在一棵老树下发现一个包袱，里面有些银两。",
		}, func() {
			gm.flags["enc_"+key] = true
			gm.pl.Money += 30 + rand.Intn(40)
			gm.push("拾获银两！")
			gm.save()
		})
	default:
		gm.encOpts = nil
		gm.beginTalk([]string{
			"山中雾起，一名黑衣人一闪而过，你隐约听见：血月教……",
			"（得到血月教活动的线索）",
		}, func() {
			gm.flags["enc_"+key] = true
			gm.addXP(20)
			gm.push("经历江湖历练 +20经验")
			gm.save()
		})
	}
	gm.mode = "talk"
}

func (gm *Game) onKeyEnc(key string) {
	// 奇遇复用 talk 模式，encOpts 预留
	gm.onKey(gm.g, key)
}

func (gm *Game) hasMartial(name string) bool {
	for _, m := range gm.pl.Martials {
		if m == name {
			return true
		}
	}
	return false
}

// ---------- 战斗胜利特殊逻辑 ----------

// onBossKilled 战斗胜利时调用（在 fight.go 的 fightWin 里）
func (gm *Game) onBossKilled(mi int) {
	switch mi {
	case 5: // 黑风寨主
		if gm.countItem(ItemLetter) == 0 {
			gm.addItem(ItemLetter)
			gm.mainProg = 2
			gm.push("从寨主身上搜出密信！（主线推进）")
		}
	case 7: // 尸王
		gm.push("尸王倒下，古墓深处传来异响……")
	case 10: // 血月教主
		gm.won = true
		gm.mainProg = 6
		gm.recordRank()
		gm.push("血月教主伏诛！天下太平！")
		gm.save()
	}
}

// ---------- 江湖榜 ----------

func (gm *Game) recordRank() {
	extra := fmt.Sprintf("%s Lv%d %s", gm.pl.Sect, gm.pl.Level, realmNames[gm.pl.Realm])
	score := gm.power()
	gm.rankIn = gm.sb.Add(gm.pl.Name, score, extra)
	gm.push(fmt.Sprintf("战力 %d，名列江湖榜第 %d 位", score, gm.rankIn+1))
}
