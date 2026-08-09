# gocligames - Go 命令行 2D 游戏引擎

纯 Go 标准库实现的跨平台终端 2D 游戏引擎，零第三方依赖。
支持 **Termux(Android) / Linux / macOS / Windows** 四大平台，
引擎代码完全一致，一处编写四处编译运行。

## 快速开始

    cd ~/hermes11/gocligames
    bash build.sh          # 一键构建三个平台的可执行文件
    ./bin/bounce           # 运行示例游戏「弹球」（本机版）
    ./bin/xiuxian          # 运行修仙地牢 RPG demo（本机版）
    ./bin/jianghu          # 运行大型武侠开放世界 RPG「江湖行」
    ./bin/jianghu --bot    # AI 自动游玩（数值平衡自测 / 自动打榜）
    ./bin/dart            # 运行快节奏反应小游戏「江湖·接暗器」
    ./bin/dart --bot      # AI 自动玩（追踪金镖/躲避毒镖，10 局统计）
    ./bin/closed          # 放置修炼小游戏「江湖·闭关」（挂机攒修为突破境界）

操作：A/D 或方向键移动挡板（W/S 同向），空格发球，P 暂停，R 重开，
E 结算（写入排行榜），Q 退出（自动保存分数）。Ctrl+C 也可退出。

## 目录结构

    gocligames/
    ├── go.mod              模块定义（module gocligames）
    ├── build.sh            三平台一键构建脚本
    ├── README.md           本文档
    ├── engine/             引擎核心（包名 engine）
    │   ├── engine.go       Game 主循环 / 生命周期 / 绘制工具(Clamp/Box/TextCentered)
    │   ├── screen.go       Screen 双缓冲字符画布 + ANSI 帧渲染
    │   ├── input.go        Input 非阻塞按键（goroutine + channel）
    │   ├── entity.go       Entity 实体（位置/速度/尺寸/碰撞盒/绘制）
    │   ├── physics.go      AABB 碰撞检测
    │   ├── score.go        Scoreboard JSON 排行榜（TopN 持久化）
    │   ├── color.go        256 色常量
    │   ├── term_linux.go   Termux/Linux 原始模式（ioctl + termios）
    │   ├── term_darwin.go  macOS 原始模式
    │   ├── term_windows.go Windows 控制台模式 + ANSI VT 启用
    │   └── term_other.go   其他平台兜底（不启用原始模式）
    ├── cmd/
    │   ├── bounce/         示例游戏「弹球」
    │   │   ├── main.go     游戏逻辑（挡板/球/碰撞/计分/结算/排行榜）
    │   │   └── bounce_test.go  单元测试
    │   ├── xiuxian/        修仙地牢 RPG demo（走地图/对话/回合制战斗/存档）
    │   │   ├── main.go     游戏逻辑（状态机/移动/战斗/存档）
    │   │   ├── world.go    世界观数据（3 层地图/怪物/物品/对话）
    │   │   └── main_test.go 单元测试
    │   ├── jianghu/        大型武侠开放世界 RPG「江湖行」（详见 cmd/jianghu/README.md）
    │       ├── main.go     状态机/移动/存档/渲染/主循环
    │       ├── world.go    6 区域地图/对象生成/传送表
    │       ├── data.go     物品/装备/武功/门派/怪物/任务数据
    │       ├── npc.go      NPC 对话/主线/支线/奇遇
    │       ├── fight.go    回合制战斗系统
    │       ├── menu.go     背包/装备/武功/任务/榜单/帮助界面
    │       ├── shop.go     商店/当铺/客栈/赌场
    │       ├── bot.go      AI 自动玩家（--bot）
    │       └── jianghu_test.go 单元测试（含地图连通性/主线通关模拟）
    │   ├── dart/          快节奏反应小游戏「江湖·接暗器」（接金镖/躲毒镖/排行榜）
    │   └── closed/        放置修炼小游戏「江湖·闭关」（选时长挂机/突破境界/修行榜）
    └── bin/                构建产物（build.sh 生成）
        ├── bounce                 Termux/Linux 本机
        ├── bounce-linux-amd64     Linux x86-64
        └── bounce-windows-amd64.exe  Windows x86-64

## 跨平台支持

| 平台 | 终端原始模式 | 按键 | ANSI 渲染 |
|------|-------------|------|-----------|
| Termux(Android) | termios ioctl ✓ | goroutine 读 stdin ✓ | ✓ |
| Linux | termios ioctl ✓ | ✓ | ✓ |
| macOS | TIOCGETA/TIOCSETA ✓ | ✓ | ✓ |
| Windows 10 1511+ | kernel32 SetConsoleMode ✓ | msvcrt 风格 scan code ✓ | VT 模式 ✓ |

- 非阻塞输入：后台 goroutine 持续读 stdin，按键经 channel 交给主循环 `Poll()`，
  跨平台行为一致，不依赖 select/epoll。
- Windows 构建只需 `GOOS=windows GOARCH=amd64 go build`（纯标准库 + syscall，
  CGO_ENABLED=0 交叉编译，无需 Windows 环境）。
- 方向键：Unix 走 ESC[A 序列，Windows 走 NUL/0xE0 scan code，两种都解析。
  游戏内以 WASD 为主控，方向键为兼容补充（Termux 手机上方向键不可靠）。

## 引擎 API

### Game - 主循环

    g := engine.NewGame(title string, w, h, fps int) *Game
    g.OnStart = func(g *engine.Game)                 // 启动钩子
    g.OnKey   = func(g *engine.Game, key string)     // 按键钩子（每帧消费所有按键）
    g.Update  = func(g *engine.Game, dt float64)     // 逻辑钩子（dt 秒）
    g.Render  = func(g *engine.Game, s *engine.Screen) // 渲染钩子
    g.OnQuit  = func(g *engine.Game)                 // 退出钩子
    g.Run()                                          // 进入主循环

按键名：小写字母（a-z）、enter、space、tab、up/down/left/right、
home、end、esc、ctrl_c（Ctrl+C 默认由引擎直接退出）。

### Screen - 画布

    s.Set(x, y int, ch rune, fg, bg int)  // 画一个字符（越界安全忽略，-1=默认色）
    s.Text(x, y int, str string, fg int)  // 画一行文字
    s.Clear()                              // 清空
    s.Width() / s.Height()                 // 尺寸
    s.Render()                             // 输出整帧（同色相邻字符自动合并转义）

### Entity - 实体

    e := engine.NewEntity(x, y float64, w, h int, ch rune, fg int) *Entity
    e.X / e.Y / e.Vx / e.Vy / e.W / e.H / e.Ch / e.Fg / e.Active
    e.Update(dt)          // 按速度位移
    e.Rect()              // AABB (x0,y0,x1,y1)
    e.CenterX() / e.CenterY()
    e.Draw(s)             // 绘制到画布

### Physics - 碰撞

    engine.AABB(ax0, ay0, ax1, ay1, bx0, by0, bx1, by1 float64) bool
    engine.Overlap(a, b *Entity) bool

### Scoreboard - 排行榜

    sb := engine.NewScoreboard("scores.json", 10)
    rank := sb.Add("PLAYER", 100, "extra info")  // 返回名次，-1=未进榜
    top := sb.Top(8)

### 绘制工具

    engine.Clamp(v, lo, hi float64) float64
    engine.Box(s, x, y, w, h int, fg int)
    engine.TextCentered(s, y int, str string, fg int)
    engine.DispWidth(str string) int   // 中英文混合宽度估算

## 写一个新游戏（3 步）

1. 子类化游戏结构体，持有实体状态：

       type MyGame struct {
           player *engine.Entity
           score  int
       }

2. 在 main 里挂钩子：

       g := engine.NewGame("我的游戏", 40, 20, 30)
       m := &MyGame{}
       g.OnStart = func(g *engine.Game) { m.reset() }
       g.OnKey = func(g *engine.Game, key string) { /* 处理按键 */ }
       g.Update = func(g *engine.Game, dt float64) { /* 逻辑 */ }
       g.Render = func(g *engine.Game, s *engine.Screen) {
           s.Clear()
           m.player.Draw(s)
           s.Text(1, 1, fmt.Sprintf("分数 %d", m.score), engine.ColorWhite)
       }
       g.Run()

3. 构建运行：`go build -o bin/mygame ./cmd/mygame && ./bin/mygame`

## 按键约定（全局统一）

    WASD 主控 | Enter 确认 | P 暂停 | R 重开 | Q 退出 | E 结算
    方向键同样解析（兼容），游戏内提示文案会写「输入法/键盘」引导手机输入法。

## 排行榜

Scoreboard 读写 JSON 文件，自动排序并截断 TopN。
示例游戏把 `bounce_scores.json` 放在当前工作目录。
测试时把排行榜路径指向临时目录即可避免污染。

## 测试

    go test ./...    # 引擎 + 示例游戏全部单元测试（不碰真实终端）

覆盖：画布越界与帧渲染、排行榜排序/截断/持久化、AABB 碰撞、
按键归一化、挡板边界、发球运动、反弹角度、漏球扣命、GameOver 自动结算、
结算防重复、渲染冒烟。

## 与 pycligames 的关系

`~/hermes11/pycligames/` 是最早用 Python 写的原型引擎（engine.py + demo.py），
功能与 gocligames 相同但只能在本机跑，性能与跨平台分发能力弱于 Go 版。
pycligames 已归档，暂不使用；gocligames 为当前主力引擎。

## 已知限制

- 渲染为整帧重绘，画布建议 < 60x30、fps <= 60。
- 中文按双宽粗略估算，界面文案尽量短（手机终端行宽有限）。
- Windows 需 Windows 10 1511+（支持 VT 序列）；老终端无颜色但能跑。
- 非 TTY 环境（管道/重定向）下 raw mode 失败会告警，按键需回车。

## 协议

MIT License（见 LICENSE）
