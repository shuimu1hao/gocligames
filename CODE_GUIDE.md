# gocligames 代码导读教程（零基础友好版）

> 这份文档写给"水平有点差、怕看不懂"的你。
> 读完它，你不仅能看懂 gocligames 的每一行代码，
> 还能改出一个属于自己的小游戏。

---

## 目录

- [0. 这份导读怎么用](#0-这份导读怎么用)
- [1. 终端游戏引擎到底在干什么](#1-终端游戏引擎到底在干什么)
- [2. Go 语法速成（只讲本项目用到的）](#2-go-语法速成只讲本项目用到的)
- [3. 引擎总览：从 main 出发](#3-引擎总览从-main-出发)
- [4. 逐文件导读](#4-逐文件导读)
- [5. 示例游戏走读：cmd/bounce/main.go](#5-示例游戏走读cmdbouncemaingo)
- [6. 动手实践：改出你的第一个游戏](#6-动手实践改出你的第一个游戏)
- [7. 名词速查表](#7-名词速查表)

---

## 0. 这份导读怎么用

**适合谁**：会一点编程（比如写过 Python、JS、C）但没写过 Go；或者基本零基础，想看懂身边这个项目。

**推荐阅读顺序**：

1. 先看第 1、2 节，建立"游戏引擎 = 一个循环"的心智模型，扫清 Go 语法障碍；
2. 再看第 3 节，知道每个文件是干嘛的；
3. 然后按 4.1 → 4.2 → 4.3 → 4.5 → 4.6 → 4.7 的顺序读（由易到难），
   4.8 看一眼就行；
4. 最后完整走读第 5 节的示例游戏，那是所有零件拼起来的样子；
5. 去第 6 节做动手练习——**看懂的唯一标准是能改**。

**阅读技巧**：遇到不懂的语法，先看第 2 节有没有解释；没有就直接跳过，
很多细节不影响整体理解。一次读不完就分几次，文档又不会跑。

---

## 1. 终端游戏引擎到底在干什么

### 1.1 所有游戏都是一个"死循环"

不管 3A 大作还是命令行小游戏，核心都是同一个循环：

    初始化（准备好一切）
    ↓
    ┌────────────→ 读按键（玩家做了什么）
    │              ↓
    │              更新逻辑（球动了吗？撞墙了吗？分数加了吗）
    │              ↓
    │              渲染画面（把最新状态画到屏幕上）
    │              ↓
    └──── 睡一会儿（控制速度，别跑太快）── 回到读按键

这个循环每转一圈，就是"一帧"。30 帧/秒 = 每秒转 30 圈。
gocligames 引擎的 `Game.Run()` 就是这个循环的化身。

### 1.2 终端游戏只有三个秘密

1. **画字符**：终端屏幕就是一堆字符（字母、数字、符号）。画游戏 =
   往屏幕的 (x, y) 位置放字符。y 是行号，x 是列号（从 0 开始）。
2. **读按键**：玩家敲键盘，程序要能"不等回车立刻拿到按键"。
   这是引擎里最麻烦的部分（见 4.7、4.8）。
3. **控制终端**：隐藏光标、清屏、把光标移到指定位置——全靠
   往终端输出一些以 ESC（转义字符）开头的"暗号"，叫 ANSI 转义序列。

引擎就是把这三个秘密封装成好用的工具，让你专心写游戏逻辑。

---

## 2. Go 语法速成（只讲本项目用到的）

Go 的语法比 C 简洁、比 Python 严谨。下面只讲 gocligames 里出现过的语法，
每个都配"人话"解释。

### 2.1 package 和 import（包和导入）

```go
package engine   // 这个文件属于 engine 这个"包"

import (         // 导入其他包，类似 Python 的 import
    "fmt"        // 格式化输出
    "os"         // 操作系统功能（读写终端等）
)
```

- `package` 是 Go 的组织单位：一个目录 = 一个包。
- `engine/` 目录下所有 `.go` 文件第一行都写 `package engine`，
  它们共享所有代码，互相可以直接调用，不需要导入。
- `cmd/bounce/main.go` 写的是 `package main`——`main` 包是特殊包，
  程序从它的 `main()` 函数开始运行。

### 2.2 变量、常量（类型在名字后面）

```go
var old syscall.Termios   // 声明变量 old，类型是 syscall.Termios
x := 10                    // 简写：声明并赋值，类型自动推断（int）
const FPS = 30             // 常量，不能改

var v float64 = 3.14       // 浮点数（小数）
var ch rune = 'o'          // rune = 一个字符（Go 里叫"符文"）
var s string = "hello"     // 字符串
var ok bool = true         // 布尔
```

注意 Go 的类型写在**变量名后面**（跟 C 相反）。`x := 10` 是"声明+赋值"
的快捷键，最常用。

### 2.3 struct：你的第一个"类"（收纳盒）

Go 没有 class（类），但有 struct（结构体）——可以理解成一个**收纳盒**，
把相关的数据放在一起：

```go
type Entity struct {   // 定义一种叫 Entity 的结构体
    X, Y   float64     // 盒子里有：位置 X、Y
    W, H   int         // 尺寸
    Ch     rune        // 长相（用哪个字符表示）
    Active bool        // 状态（是否启用）
}
```

使用：

```go
e := Entity{X: 10, Y: 5, W: 1, H: 1, Ch: 'o', Active: true}
e.X = 20               // 修改盒子里的值
fmt.Println(e.Y)       // 读取
```

类比：`struct` 就像一个收纳盒，`X`、`Y` 是盒子里的小格子。
一个游戏里的"球"、"挡板"都是一个这样的收纳盒。

### 2.4 方法：挂在结构体上的函数（收纳盒的功能按钮）

在 Go 里，可以给结构体"绑"函数，绑定的函数叫**方法**：

```go
func (e *Entity) Update(dt float64) {   // 注意括号里多了 (e *Entity)
    e.X += e.Vx * dt                    // 移动：位置 = 位置 + 速度 × 时间
    e.Y += e.Vy * dt
}
```

`(e *Entity)` 叫**接收者**——意思是"这个方法属于 Entity"，
调用时用 `ball.Update(0.016)`，Go 会自动把 `ball` 当作 `e` 传进去。

`*Entity` 里的 `*` 是指针（见 2.6），暂时可以理解为"直接改这个实体本身"。

类比：`Entity` 是收纳盒，`Update` 是盒子上自带的按钮，
按一下（调用一次），盒子里的 X、Y 就自动更新。

### 2.5 函数也是值：回调（本项目最巧的设计）

Go 里函数可以像变量一样传来传去。看 `Game` 的定义：

```go
type Game struct {
    OnKey  func(g *Game, key string)   // 这个字段的类型是"函数"！
    Update func(g *Game, dt float64)
    Render func(g *Game, s *Screen)
}
```

`OnKey` 的类型是 `func(g *Game, key string)`——意思是：
"OnKey 这个格子里装的不是一个数，而是一个函数（一段代码）"。

怎么用？在示例游戏里：

```go
g.OnKey = b.onKey    // 把 b.onKey 这个函数塞进 Game 的 OnKey 格子
g.Run()              // 引擎每帧自动帮你调用它
```

**这就是引擎的工作方式**：引擎负责循环、读按键、渲染这些麻烦事；
你在 `OnKey`、`Update`、`Render` 里只写"这个游戏自己的逻辑"。
引擎在合适的时机帮你调用这些函数——这就是**回调（callback）**。

类比：引擎是个插座，你的函数是插头。引擎每次按键就"通电"调用你的 OnKey。

### 2.6 指针：* 符号（别怕）

C 语言的指针把人吓跑一半，Go 的指针温和多了。记住一条：

> 在 Go 里，**凡是方法（接收者）带 `*` 的，就是"直接改原盒子"；
> 不带 `*` 的是"复制一份再改"（改了不生效）。**

```go
func (e *Entity) Update(dt float64) { e.X += ... }  // 带 *：真的改了 e.X
```

所以本项目里所有"要修改对象"的方法都带 `*`。
你基本不需要手动写 `*`，调用时 Go 自动处理。
看到 `*Screen`、`*Game` 就当"指向那个东西"即可，不用深究。

### 2.7 goroutine 和 channel：并发（最难，但只用一个套路）

Go 最出名的特性。本项目只用了**一个固定套路**，看懂套路即可：

- **goroutine**：用 `go` 关键字启动一个"并发执行的小工"，代码立刻继续往下走：

```go
go in.readLoop()   // 让 readLoop 函数在后台跑，main 不等它
```

- **channel（通道）**：小工和主程序之间传东西的"管道"：

```go
ch := make(chan string, 16)   // 造一根能装 16 个字符串的管道
ch <- "a"                     // 往管道塞一个值（箭头方向 = 数据流动方向）
k := <-ch                     // 从管道取一个值
```

- **select**：多根管道之间"谁有货拿谁"：

```go
select {
case k := <-in.ch:        // 管道有按键？拿一个
    return k
default:                  // 暂时没有？
    return ""             // 返回空（调用方就知道"没按键"）
}
```

本项目套路：**一个后台小工负责一直读键盘，把按键塞进管道；
主游戏循环每帧用 select 看看管道里有没有货（按键），有就拿走，没有就继续干别的。**
这就是"非阻塞按键"的实现。4.7 会详细走读。

### 2.8 build tags：同一套代码适配多个系统

有的代码只能在特定系统跑（比如改终端设置的代码，Linux 和 Windows 完全不一样）。
Go 用文件开头的注释来"按系统挑文件"：

```go
//go:build linux       ← 这个文件只在 Linux 上编译

//go:build windows     ← 这个文件只在 Windows 上编译
```

所以 `engine/` 里有 4 个 `term_*.go` 文件，函数名都叫 `makeRaw`，
但每个系统只编译自己那个，互不冲突——这就是"一份代码跑四个平台"的关键。

### 2.9 错误处理：err

Go 没有 try/catch。函数如果可能出错，就返回一个 error 类型的值：

```go
data, err := os.ReadFile(sb.Path)   // 读文件
if err != nil {                     // err 不是 nil（nil = 空/无）说明出错
    return nil                      // 出错就提前返回
}
```

本项目里看到 `if err != nil { return ... }` 就是"出错了，别继续"。
`nil` 是 Go 的"空值"。

---

## 3. 引擎总览：从 main 出发

程序入口是 `cmd/bounce/main.go` 的 `main()`，一局游戏的完整流程：

    main() 创建 Game（发动机）+ 创建 Bounce（游戏内容）
        ↓
    g.Run() 进入引擎主循环（engine.go）
        ↓
    引擎初始化终端（term_*.go 的 makeRaw）
        ↓
    调 OnStart（你的游戏准备数据：复位、读排行榜）
        ↓
    死循环：
        读按键 → 调 OnKey（你的游戏处理按键）
             → 调 Update（你的游戏逻辑：球动、碰撞、计分）
             → 调 Render（你的游戏画画面）
             → 睡一会儿（控制帧率）
        ↓
    退出（Q / Ctrl+C / 游戏结束）
        ↓
    引擎恢复终端（显示光标、恢复按键回显）→ 程序结束

**文件地图**（按建议阅读顺序）：

| 文件 | 职责 | 难度 |
|------|------|------|
| engine/color.go | 颜色常量 | ⭐ |
| engine/entity.go | 实体（盒子） | ⭐ |
| engine/physics.go | 碰撞检测 | ⭐ |
| engine/screen.go | 画布 + ANSI 渲染 | ⭐⭐ |
| engine/score.go | 排行榜存取 | ⭐ |
| engine/engine.go | 主循环 + 回调设计 | ⭐⭐ |
| engine/input.go | 非阻塞按键 | ⭐⭐⭐ |
| engine/term_*.go | 终端原始模式 | ⭐⭐⭐（了解即可） |
| cmd/bounce/main.go | 示例游戏全部逻辑 | ⭐⭐ |

---

## 4. 逐文件导读

### 4.1 engine/color.go（最简单，30 秒读完）

```go
package engine

const (
    ColorRed     = 196
    ColorCyan    = 51
    ColorWhite   = 231
    // ...
)
```

就这么点。这些数字是 **256 色 ANSI 调色板**的编号：
终端能显示 256 种颜色，每种有个编号，196 是红、51 是青、231 是白……
游戏里给字符上色就传这些数字。`ColorGray = 245` 是灰（常用）。

**小结**：颜色 = 数字。别背，用到查。

### 4.2 engine/entity.go（实体=收纳盒 + 按钮）

```go
type Entity struct {          // 收纳盒：装一个游戏对象的数据
    X, Y   float64            // 位置（左上角坐标，浮点数才能平滑移动）
    W, H   int                // 占几格
    Vx, Vy float64            // 速度（每秒走几格）
    Ch     rune               // 用什么字符画它（球是 'o'，挡板是 '='）
    Fg     int                // 颜色
    Active bool               // 假了就既不更新也不画（比如被吃掉的球）
}

func NewEntity(x, y float64, w, h int, ch rune, fg int) *Entity {
    // 构造函数：方便地造一个装满默认值的实体
    return &Entity{X: x, Y: y, W: w, H: h, Ch: ch, Fg: fg, Active: true}
}

func (e *Entity) Update(dt float64) {   // 按钮1：移动
    if !e.Active { return }             // 假了就啥也不干
    e.X += e.Vx * dt                    // 新位置 = 旧位置 + 速度 × 时间
    e.Y += e.Vy * dt                    // （初中物理：路程 = 速度 × 时间）
}

func (e *Entity) Rect() (x0, y0, x1, y1 float64) {
    // 按钮2：算碰撞盒。返回左、上、右、下四条边
    return e.X, e.Y, e.X + float64(e.W), e.Y + float64(e.H)
}

func (e *Entity) Draw(s *Screen) {      // 按钮3：把自己画到画布上
    if !e.Active { return }
    for j := 0; j < e.H; j++ {          // 逐格画（宽几格画几格）
        for i := 0; i < e.W; i++ {
            s.Set(int(e.X)+i, int(e.Y)+j, e.Ch, e.Fg, -1)
        }
    }
}
```

重点理解：

- `float64` 位置 + `int` 格子：逻辑上位置可以是小数（平滑移动），
  真正画的时候 `int(e.X)` 取整成格子。
- `Rect()` 返回四条边，是给碰撞检测用的"看不见的盒子"。
- 一个实体就是"数据 + 三个按钮（更新/算盒/绘制）"。

### 4.3 engine/physics.go（纯函数，3 分钟）

```go
// AABB：判断两个矩形是否重叠（AABB = 轴对齐包围盒）
func AABB(ax0, ay0, ax1, ay1, bx0, by0, bx1, by1 float64) bool {
    return !(ax1 < bx0 || bx1 < ax0 || ay1 < by0 || by1 < ay0)
}
```

这是整个碰撞检测的全部！理解方法：**两个矩形不重叠的条件**是：
A 的右边在 B 的左边（ax1 < bx0），或 B 的右边在 A 的左边，或上下同理。
只要不满足任何一种"分开"，那它们就叠在一起了。

```go
func Overlap(a, b *Entity) bool {   // 方便版：直接问两个实体叠没叠
    ax0, ay0, ax1, ay1 := a.Rect()  // 取 A 的四条边
    bx0, by0, bx1, by1 := b.Rect()  // 取 B 的四条边
    return AABB(ax0, ay0, ax1, ay1, bx0, by0, bx1, by1)
}
```

**小结**：碰撞 = 四个不等式。游戏里判断"球撞到挡板了吗"就一行：
`engine.Overlap(ball, paddle)`。

### 4.4 engine/screen.go（画布与 ANSI，重点）

#### 4.4.1 数据结构：一张字符表格

```go
type Screen struct {
    W, H  int      // 画布宽高（几列 × 几行）
    cells [][]Cell // 二维表格：每行每列存一个 Cell
}
```

`[][]Cell` 是"切片（动态数组）的切片"——一张二维表格。
`Cell` 存三个东西：

```go
type Cell struct {
    Ch rune // 这个格子放什么字符
    Fg int  // 前景色（字符颜色）
    Bg int  // 背景色（格子底色）
}
```

`Screen` 就是一张"表格 + 一支笔"：`Set(x, y, ch, fg, bg)` 就是在
(x, y) 这格写下字符。**游戏里的画面 = 往表格里填字**。

```go
func (s *Screen) Set(x, y int, ch rune, fg, bg int) {
    if x < 0 || y < 0 || x >= s.W || y >= s.H {
        return          // 越界就假装没看见（防崩）
    }
    s.cells[y][x] = Cell{Ch: ch, Fg: fg, Bg: bg}
}
```

注意 `s.cells[y][x]`：先 y 后 x——因为第一维是"行"（y），第二维是"列"（x）。

#### 4.4.2 渲染：把表格变成终端能懂的命令

双缓冲的"双"是指：**先全部画到内存表格里，再一次整帧输出**，
而不是画一格输出一格（那样会闪烁、撕裂）。

```go
func (s *Screen) Frame() string {
    var b strings.Builder
    b.WriteString("\x1b[H")        // 命令①：光标移回屏幕左上角
    for y := 0; y < s.H; y++ {
        for x := 0; x < s.W; x++ {
            // 颜色变了才输出颜色命令（省流量）
            // ...
            b.WriteRune(c.Ch)      // 写字符
        }
    }
    b.WriteString("\x1b[0m\x1b[J") // 命令②：复位颜色 ③：清掉残留
    return b.String()
}
```

**这里必须讲清楚 `\x1b` 是什么**：

- `\x1b` 是 ESC 字符（ASCII 码 27）的写法。`\x` 表示"后面跟十六进制"，
  所以 `\x1b` = 十六进制 1B = 十进制 27 = 终端眼里的"命令开头"。
- 终端看到以 ESC 开头的特定序列，就知道你要干嘛：

  | 序列 | 含义 | 人话 |
  |------|------|------|
  | `\x1b[H` | 光标回家 | 把"画笔"挪回左上角 |
  | `\x1b[38;5;196m` | 设前景色 | 后面写的字用 196 号颜色 |
  | `\x1b[39m` | 恢复默认前景色 | 颜色复位 |
  | `\x1b[0m` | 全部复位 | 颜色、样式全清 |
  | `\x1b[J` | 清屏（从光标到末尾） | 擦掉旧画面残留 |
  | `\x1b[?25l` | 隐藏光标 | 让闪烁的光标别碍事 |
  | `\x1b[?25h` | 显示光标 | 退出游戏时恢复 |

- `fmt.Fprintf(&b, "\x1b[38;5;%dm", c.Fg)` 就是"我要第 c.Fg 号颜色"
  的命令拼装。颜色编号占位 `%d`，运行时填进去。

**整帧渲染的流程**：光标回家 → 逐行逐格：必要时换颜色 → 写字符 →
结尾复位。因为每帧都是从"左上角重新画"，画面就像动画片一样更新。

**小结**：Screen 帮你把"往 (x,y) 写字"翻译成终端命令。
你永远不用自己写 `\x1b` 序列，只要 `Set()`。

### 4.5 engine/score.go（排行榜，5 分钟）

```go
type ScoreRec struct {           // 一条成绩
    Name  string `json:"name"`   // 名字
    Score int    `json:"score"`  // 分数
    Extra string `json:"extra,omitempty"` // 附加信息（如 "3 hits"）
}
```

`json:"name"` 这种叫 **struct tag**：告诉 JSON 库"这个字段存盘时叫 name"。
（Go 字段名是大写开头的 `Name`，但 JSON 里用小写 `name`，靠 tag 转换。）

```go
func (sb *Scoreboard) Add(name string, score int, extra string) int {
    rec := ScoreRec{...}
    sb.Scores = append(sb.Scores, rec)        // 追加一条
    sort.SliceStable(sb.Scores, func(i, j int) bool {
        return sb.Scores[i].Score > sb.Scores[j].Score  // 按分数从高到低排
    })
    if len(sb.Scores) > sb.Limit {            // 超过 10 条？
        sb.Scores = sb.Scores[:sb.Limit]      // 切掉后面的（只留前 10）
    }
    sb.save()                                 // 写盘（JSON 文件）
    // ...找到这条记录排第几，返回名次
}
```

`append` 是 Go 的"往切片后面加"；`sort.SliceStable` 是"按我给的规则排序"
（规则 = 前一个分数大于后一个 → 降序）；`[:sb.Limit]` 是切片语法
"取前 Limit 个"。

`save()` 用 `json.MarshalIndent` 把数据变成带缩进的 JSON 文本写进文件；
`load()` 用 `os.ReadFile` + `json.Unmarshal` 读回来。**这就是排行榜
下次启动还能在的原因**。

**小结**：排行榜 = 数组 + 排序 + JSON 存盘。三个标准库函数搞定。

### 4.6 engine/engine.go（游戏心脏，重点）

先看结构：

```go
type Game struct {
    Title  string        // 标题（暂未用，可留给你的游戏显示）
    Screen *Screen       // 画布（引擎自带一个）
    Input  *Input        // 键盘（引擎自带一个）
    FPS    int           // 帧率（每秒几帧）
    Paused bool          // 暂停标志（游戏可以改它）

    running bool         // 内部：主循环是否继续（小写=私有，外部改不了）

    OnStart func(g *Game)                 // ← 你提供的钩子（回调）
    OnKey   func(g *Game, key string)     // ← 按键来了叫你
    Update  func(g *Game, dt float64)     // ← 每帧叫你更新
    Render  func(g *Game, s *Screen)      // ← 每帧叫你画画
    OnQuit  func(g *Game)                 // ← 退出时叫你
}
```

`NewGame(title, w, h, fps)` 是构造函数：造画布、造输入器。

**主角是 Run()，主循环本体**：

```go
func (g *Game) Run() {
    restore, err := makeRaw()   // ① 让终端进入"原始模式"（按键立刻生效）
    if err != nil {
        fmt.Fprintln(os.Stderr, "...")  // 失败就警告（还能跑）
    } else {
        defer restore()         // ② 保证退出时恢复终端（重要！）
    }
    defer g.Input.Close()       // ③ 保证退出时停掉读键盘的小工
    defer g.showCursor()        // ④ 保证退出时恢复光标
    g.hideCursor()              // 先隐藏光标

    if g.OnStart != nil { g.OnStart(g) }  // ⑤ 通知你：开始！
    g.running = true
    frameTime := time.Second / time.Duration(g.FPS)  // 一帧该花多久
    last := time.Now()

    for g.running {                        // ⑥ 死循环开始
        now := time.Now()
        dt := now.Sub(last).Seconds()      // 这一帧过了几秒
        last = now
        if dt > 0.1 { dt = 0.1 }           // 卡顿时限幅，防大跳跃

        for {                              // ⑦ 把队列里所有按键都处理掉
            k := g.Input.Poll()
            if k == "" { break }           // 没按键了，跳出
            if k == "ctrl_c" { g.running = false; break }  // Ctrl+C 退出
            if g.OnKey != nil { g.OnKey(g, k) }            // 叫你处理按键
        }

        if !g.Paused && g.Update != nil {  // ⑧ 没暂停才更新逻辑
            g.Update(g, dt)
        }
        if g.Render != nil {               // ⑨ 每帧都画
            g.Render(g, g.Screen)
        }
        g.Screen.Render()                  // ⑩ 输出整帧到终端

        if sleep := frameTime - time.Since(now); sleep > 0 {
            time.Sleep(sleep)              // ⑪ 睡够剩余时间，稳住帧率
        }
    }
    if g.OnQuit != nil { g.OnQuit(g) }     // ⑫ 循环结束，通知你
}
```

要点：

- **defer**：Go 的"临走前必做"清单。`defer X()` 表示"函数返回前执行 X"，
  按倒序执行。就算中间崩了、Ctrl+C 了，defer 也保证执行——
  这就是"玩完游戏终端不会被弄乱"的保障。
- **dt（delta time）**：每帧经过的秒数。所有移动都用 `速度 × dt`，
  这样不管帧率是 30 还是 60，球速都一样（时间制，不是帧制）。
- **帧率控制**：一帧该花 `1/FPS` 秒，实际花了 `time.Since(now)`，
  差多少睡多少，稳稳的 30 帧/秒。
- **钩子为 nil 也能跑**：`if g.OnKey != nil` 判断——你不填的钩子
  引擎就直接跳过，填了才调。这就是"框架"：留好接口，随你填充。

**小结**：engine.go = 一个循环 + 六个插槽（OnStart/OnKey/Update/Render/OnQuit/Paused）。

### 4.7 engine/input.go（最难但最优雅，重点）

目标：**不按回车也能立刻拿到单个按键**（游戏不能等你敲回车）。
方案：一个后台小工（goroutine）一直读键盘，把按键塞进管道，
主循环随时从管道"取货"。

#### 4.7.1 启动

```go
func NewInput() *Input {
    in := &Input{
        ch:   make(chan string, 16),   // 管道：最多缓存 16 个按键
        done: make(chan struct{}),     // 停机信号（struct{} 是空类型，只当开关用）
    }
    go in.readLoop()                   // 启动小工：后台读键盘去
    return in
}
```

#### 4.7.2 小工：读键盘循环

```go
func (in *Input) readLoop() {
    defer close(in.ch)                 // 退出时关管道（收尾）
    r := bufio.NewReader(os.Stdin)     // 拿一个"带缓冲的键盘读者"
    for {
        select {
        case <-in.done:                // 收到停机信号？
            return                     // 小工下班
        default:                       // 否则继续干活
        }
        ch, _, err := r.ReadRune()     // 读一个字符（阻塞：没按键就等着）
        if err != nil { return }       // 读不到（管道关闭等）就下班
        if ch == 0x1b {                // 是 ESC？可能是方向键，转去解析
            in.handleEscape(r)
            continue
        }
        if k := normalizeKey(ch); k != "" {  // 归一化：'A'→'a'，回车→"enter"
            in.push(k)                        // 塞进管道
        }
    }
}
```

`ReadRune` 是**阻塞**的——没按键时它就一直等。正因为小工在后台等，
主循环才能"想查就查、不查拉倒"。这是整个设计的精髓：
**阻塞的事情交给后台，前台永远不阻塞。**

#### 4.7.3 塞货与取货

```go
func (in *Input) push(k string) {
    select {
    case in.ch <- k:        // 管道有空位：塞进去
    default:                // 管道满了？
        // 丢掉（游戏卡顿时的按键丢了无所谓）
    }
}

func (in *Input) Poll() string {
    select {
    case k, ok := <-in.ch:  // 有货：拿走
        if !ok { return "" }
        return k
    default:                // 没货：返回空串
        return ""
    }
}
```

两个函数都用 `select + default` 做到**非阻塞**：
有货就交易，没货立刻返回，绝不等待。

#### 4.7.4 按键归一化

```go
func normalizeKey(r rune) string {
    switch r {
    case '\r', '\n': return "enter"   // 回车
    case ' ':        return "space"   // 空格
    case 0x03:       return "ctrl_c"  // Ctrl+C（原始模式下它是个字符）
    }
    if r >= 'A' && r <= 'Z' {
        return string(r - 'A' + 'a')  // 大写统一成小写（A→a）
    }
    if r >= 0x20 && r < 0x7f {        // 可见字符
        return string(r)              // 原样返回（如 "w"、"d"）
    }
    return ""                         // 控制字符：忽略
}
```

游戏里只认小写 + 几个名字：`"a" "d" "space" "p" "r" "q" "e"`。
`normalizeKey` 保证不管是敲 A 还是 a，游戏都收到 `"a"`。

#### 4.7.5 方向键解析（handleEscape）

方向键在终端里不是一个字符，而是一串：`ESC [ A`（上）、`ESC [ B`（下）等。
小工读到 ESC 后，再读两个字符拼出方向键：

```go
func (in *Input) handleEscape(r *bufio.Reader) {
    r1, _, err := r.ReadRune()   // 读 ESC 后的第一个字符
    ...
    if r1 == '[' || r1 == 'O' {
        r2, _, err2 := r.ReadRune()  // 读第二个字符
        switch r2 {
        case 'A': in.push("up")      // ESC [ A → 上
        case 'B': in.push("down")
        case 'C': in.push("right")
        case 'D': in.push("left")
        ...
        }
    }
    ...
}
```

Windows 的方向键是另一种暗号（`ESC` 后跟 `NUL H` 等），也有分支处理。
这就是引擎"方向键两套系统都能用"的秘密。

**小结**：input.go = 一个后台小工 + 一根管道 + 两个非阻塞函数 + 一个翻译官。

### 4.8 engine/term_*.go（终端黑魔法，了解即可）

这些文件是"让按键不用回车、立刻生效"的系统级设置，原理一句话：
**修改终端的行为参数**。用 `ioctl` 系统调用把终端从"行模式"改成"原始模式"：

- 行模式：用户敲字符，回车才交给程序（平时终端就是这样）
- 原始模式：敲一下立刻给程序（游戏要的）

```go
// term_linux.go（Linux / Termux 版）
func makeRaw() (restore func() error, err error) {
    fd := syscall.Stdin                     // 标准输入（键盘）
    var old syscall.Termios                 // 先记录当前设置（留着恢复用）
    // 读取终端当前设置（TCGETS = 一个系统暗号）
    syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), syscall.TCGETS, ...)
    raw := old
    raw.Lflag &^= syscall.ICANON | syscall.ECHO   // 关掉"行缓冲"和"回显"
    // 把新设置写回去（TCSETS）
    syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), syscall.TCSETS, ...)
    return func() error {  // 返回的 restore：退出时恢复原样
        // 把 old 写回去
    }, nil
}
```

- `ICANON` = 行缓冲开关，关掉它 → 按键立刻生效
- `ECHO` = 回显开关，关掉它 → 你敲的 w 不会显示在屏幕上（避免画面被破坏）
- 返回的 `restore` 函数让引擎退出时把终端恢复原样

Windows 版（term_windows.go）思路一样，只是调 Windows API
（kernel32.dll 的 GetConsoleMode / SetConsoleMode），
还额外打开"虚拟终端处理"开关让 Windows 认识 ANSI 颜色命令。

> 这一节不用全懂。你只需要知道：**引擎处理好了，你的游戏不用管**。

---

## 5. 示例游戏走读：cmd/bounce/main.go

现在把所有零件拼起来。这个文件就是"用引擎写的第一个游戏"。

### 5.1 开头：常量、入口

```go
package main

const (
    w, h    = 46, 22   // 画布大小（宽 46 列，高 22 行）
    gTop    = 1        // 游戏区上边界（1，因为第 0 行是边框）
    paddleY = h - 4    // 挡板在第几行（从下往上数第 4 行）
    paddleW = 6        // 挡板宽度
    fps     = 30       // 帧率
)

func main() {
    dir, _ := os.Getwd()                       // 当前目录
    scoresFile = filepath.Join(dir, "bounce_scores.json") // 排行榜文件路径

    g := engine.NewGame("弹球 (gocligames demo)", w, h, fps) // 造发动机
    b := &Bounce{}                             // 造游戏内容（自己的状态）

    g.OnStart = b.onStart   // ← 插头插进插座
    g.OnKey = b.onKey
    g.Update = b.update
    g.Render = b.render
    g.OnQuit = b.onQuit
    g.Run()                  // 启动！
}
```

`main` 只干三件事：造引擎、造游戏、把游戏的函数插进引擎的插槽，然后点火。

### 5.2 游戏状态（自己的收纳盒）

```go
type Bounce struct {
    paddle     *engine.Entity   // 挡板（实体）
    ball       *engine.Entity   // 球（实体）
    ballActive bool             // 球是否在飞（false = 等发球）
    ballSpeed  float64          // 球速
    dirX, dirY float64          // 球的方向（单位向量：x 方向 1=右，y 方向 -1=上）
    score      int              // 分数
    lives      int              // 生命
    hits       int              // 累计接球次数（连击）
    settled    bool             // 是否已结算（防止重复写排行榜）
    over       bool             // 是否游戏结束
    lastRank   int              // 本次排名（结束画面高亮用）
    hint       string           // 底部提示文字
    scores     *engine.Scoreboard  // 排行榜
}
```

**这里的关键**：`Bounce` 是"你的游戏的整个世界"，所有状态都装在里面。
引擎不关心里面是什么，只负责在合适的时机调用你的方法。

### 5.3 复位（新开一局）

```go
func (b *Bounce) reset() {
    // 挡板：放在底部中央，6 格宽，用 '=' 青色
    b.paddle = engine.NewEntity(float64(w/2-paddleW/2), paddleY, paddleW, 1, '=', engine.ColorCyan)
    // 球：放在中间偏上，1 格，用 'o' 黄色
    b.ball = engine.NewEntity(float64(w/2), float64(gTop+2), 1, 1, 'o', engine.ColorYellow)
    b.ballActive = false          // 等玩家按空格发球
    b.ballSpeed = 6.0             // 初始球速
    b.dirX, b.dirY = 0.8, -0.6    // 向右上方飞（斜着起飞）
    b.score, b.lives, b.hits = 0, 3, 0   // 分数 0，生命 3
    b.settled, b.over = false, false
    b.lastRank = -1
    b.hint = "空格发球 | AD/方向键移动 | P暂停 R重开 E结算 Q退出"
}
```

### 5.4 按键处理（onKey：把按键翻译成动作）

```go
func (b *Bounce) onKey(g *engine.Game, key string) {
    if b.settled || b.over {           // 已经结束/结算了？
        switch key {
        case "r": b.reset()            // R = 重开
        case "q": g.Quit()             // Q = 退出
        }
        return                          // 结束状态下其他键不管
    }
    switch key {
    case "p":
        g.Paused = !g.Paused           // P = 暂停（切换）
    case "r":
        b.reset()                      // R = 重开
    case "q":
        b.quitSave()                   // Q = 保存分数再退出
        g.Quit()
    case "e":
        b.settle()                     // E = 手动结算（写入排行榜）
    case "a", "w", "left":             // 左移（三键都行）
        b.paddle.X = engine.Clamp(b.paddle.X-1, 1, float64(w-1-paddleW))
    case "d", "s", "right":            // 右移
        b.paddle.X = engine.Clamp(b.paddle.X+1, 1, float64(w-1-paddleW))
    case "space":
        if !b.ballActive && !b.settled {  // 球没在飞 → 发球
            b.ballActive = true
            b.hint = "P暂停 R重开 E结算 Q退出"
        }
    }
}
```

`engine.Clamp(v, lo, hi)` 是"把 v 夹在 lo 和 hi 之间"：挡板左移最多到 1
（别出边框），右移最多到 `w-1-paddleW`（右边留一格）。

### 5.5 游戏逻辑（update：让世界动起来）

```go
func (b *Bounce) update(g *engine.Game, dt float64) {
    if g.Paused || b.settled || b.over || !b.ballActive {
        return        // 暂停/结束/没发球 → 什么都不做
    }
    bl := b.ball
    // 1) 按方向 × 速度 × 时间移动
    bl.X += b.dirX * b.ballSpeed * dt
    bl.Y += b.dirY * b.ballSpeed * dt

    // 2) 左右墙反弹
    if bl.X < 1 {                 // 碰到左墙
        bl.X = 1                  // 别出界
        b.dirX = math.Abs(b.dirX) // 方向取绝对值 = 向右
    } else if bl.X > float64(w-2) {  // 碰到右墙
        bl.X = float64(w - 2)
        b.dirX = -math.Abs(b.dirX)   // 取负 = 向左
    }
    // 3) 上墙反弹（同理）
    if bl.Y < float64(gTop) {
        bl.Y = float64(gTop)
        b.dirY = math.Abs(b.dirY)    // 向下
    }

    // 4) 挡板碰撞
    if bl.Y >= float64(paddleY) && b.dirY > 0 && engine.Overlap(bl, b.paddle) {
        bl.Y = float64(paddleY - 1)   // 贴到挡板上方
        // 关键设计：打中挡板不同位置，反弹角度不同（有技巧性）
        off := (bl.X + 0.5 - b.paddle.CenterX()) / (float64(paddleW) / 2)
        off = engine.Clamp(off, -1, 1)        // 归一化到 [-1, 1]：-1=最左端 1=最右端
        ang := off*math.Pi/4 + math.Pi/2      // 角度：中间=垂直向上，越靠边越斜
        b.dirX = math.Cos(ang)                // 方向 = 角度的 x 分量
        b.dirY = -math.Abs(math.Sin(ang))     // 向上飞（y 取负）
        b.hits++
        b.score += 10 + int(math.Abs(off)*10) // 中间接 10 分，边缘接 20 分（奖励技巧）
        if b.hits%5 == 0 && b.ballSpeed < 12.0 {
            b.ballSpeed += 0.6                // 每接 5 次球加速一次（刺激）
        }
    }

    // 5) 漏球
    if bl.Y > float64(h-2) {
        b.lives--                             // 扣一条命
        if b.lives <= 0 {
            b.over = true
            b.settle()                        // 游戏结束，自动结算
        } else {
            b.ball = engine.NewEntity(...)    // 重置球，等重新发球
            b.ballActive = false
            b.hint = "漏球！空格发球"
        }
    }
}
```

逻辑就是一句话：**球动了 → 撞墙反弹 → 撞挡板反弹+得分 → 漏了扣命**。
反弹的实现是"方向取反"（碰左墙 x 分量取正，碰右墙取负）。

### 5.6 渲染（render：把世界画出来）

```go
func (b *Bounce) render(g *engine.Game, s *engine.Screen) {
    s.Clear()                                  // 清空画布
    engine.Box(s, 0, 0, w, h, engine.ColorGray) // 画边框（+----+）
    s.Text(2, 1, fmt.Sprintf("分数 %d", b.score), engine.ColorWhite)  // 顶栏
    ...
    s.Text(2, h-2, b.hint, engine.ColorGray)   // 底部提示行
    b.paddle.Draw(s)                           // 画挡板
    if b.ballActive {
        b.ball.Draw(s)                         // 画球
    }
    // 覆盖层：暂停 / 结束
    if g.Paused {
        engine.TextCentered(s, h/2, "=== 暂停中 (P 继续) ===", engine.ColorYellow)
    } else if b.settled || b.over {
        b.renderBoard(s)                       // 结算画面（排行榜）
    }
}
```

渲染非常"直白"：先清空，再画边框、文字、实体，最后画覆盖层。
`Draw` 是 Entity 自带的方法（4.2），`Text`/`Box`/`TextCentered` 是引擎工具。

### 5.7 结算与退出（settle / quitSave）

```go
func (b *Bounce) settle() {
    if b.settled { return }        // 已经结算过？不重复
    b.settled = true
    if b.score > 0 {               // 0 分不占榜
        b.lastRank = b.scores.Add("PLAYER", b.score, fmt.Sprintf("%d hits", b.hits))
    }
}

func (b *Bounce) quitSave() {
    // Q 退出时：没结算过、有分 → 静默存一下，不让分数白打
    if !b.settled && !b.over && b.score > 0 {
        b.scores.Add("PLAYER", b.score, fmt.Sprintf("%d hits", b.hits))
    }
}
```

---

## 6. 动手实践：改出你的第一个游戏

看完不是结束，**动手改才算会**。以下任务按难度递增，全部在
`cmd/bounce/main.go` 里改，改完 `go build -o bin/bounce ./cmd/bounce` 再运行。

### 任务 1：给球换皮肤（2 分钟）

在 `reset()` 里找球的创建：

```go
b.ball = engine.NewEntity(..., 1, 1, 'o', engine.ColorYellow)
```

把 `'o'` 改成 `'●'` 或 `'*'`，把 `engine.ColorYellow` 改成
`engine.ColorRed`。重新构建运行，球变样了！

### 任务 2：挡板加长（2 分钟）

在 `const` 区找 `paddleW = 6`，改成 `paddleW = 10`。挡板更宽更好接。
（注意：`reset()` 里挡板位置是 `w/2-paddleW/2`，自动居中，不用改。）

### 任务 3：球的初始方向斜一点（3 分钟）

在 `reset()` 里找 `b.dirX, b.dirY = 0.8, -0.6`，改成 `0.95, -0.35`。
球会更平、更快到两边——难度上升。

### 任务 4：球速更快（3 分钟）

找 `b.ballSpeed = 6.0` 改成 `8.0`；把加速上限 `b.ballSpeed < 12.0`
改成 `16.0`。刺激模式。

### 任务 5：三条命变五条（2 分钟）

找 `b.score, b.lives, b.hits = 0, 3, 0` 改成 `0, 5, 0`。

### 任务 6：加分规则改一下（5 分钟）

在 update 的挡板碰撞里找 `b.score += 10 + int(math.Abs(off)*10)`。
改成 `b.score += 5 + int(math.Abs(off)*5)`——分数降一半，更难上排行榜。
或者加一条"漏球扣分"：漏球时 `b.score -= 5`（注意别扣成负数）。

### 任务 7：加一个"暂停提示"（挑战）

在 render 里暂停分支已经有覆盖层了。试试在暂停时把挡板和球也画出来
（现在 `paused` 时照样画了，其实已经支持）。再试试加个"每接 10 次
发个金色球，接住 +50 分"——这需要加一个字段、改逻辑、改渲染，
是完整的"加功能"流程。

### 任务 8：把弹球改成"打砖块"（终极挑战）

思路：加一个 `bricks []*engine.Entity` 数组，开局铺满上方几行；
update 里检测球和每个砖块的 `Overlap`，撞到就删掉那个砖、反弹、加分；
全部砖块清完 → 过关。这就是 Breakout，引擎所有能力你都用上了。

改坏了怎么办？`git init && git add -A && git commit`（先存个底），
随便折腾，改坏 `git checkout .` 恢复。**学 Go 最快的路就是改代码。**

---

## 7. 名词速查表

| 名词 | 一句话解释 |
|------|-----------|
| package | 包，一个目录一组代码；main 包是程序入口 |
| struct | 结构体，装数据的收纳盒（Go 的"类"） |
| 方法 | 挂在结构体上的函数，第一个参数是"自己" |
| 接收者 | 方法定义里 `(e *Entity)` 那个，决定方法属于谁 |
| 指针 `*` | "直接改原对象"的标记；本项目方法全带 `*` |
| 回调 | 你写的函数，引擎到点帮你调用 |
| 钩子 | 同回调；引擎留的插槽 OnStart/OnKey/Update/Render |
| goroutine | `go 函数()` 启动的后台小工（并发） |
| channel | 管道，小工和主程序之间传数据 |
| select | 多管道取货，配 default 就是非阻塞 |
| build tag | 文件开头注释，按系统挑文件编译 |
| ANSI 转义 | `\x1b[...` 开头的终端命令（清屏、上色、移光标） |
| `\x1b` | ESC 字符（ASCII 27），所有终端命令的开头 |
| dt | 一帧经过的秒数，移动都乘它（时间制） |
| FPS | 帧率，每秒画面刷新次数 |
| raw mode | 终端原始模式：按键立刻生效、不回显 |
| defer | "函数返回前必做"清单，退出清理的保障 |
| ioctl | 给终端发系统级设置命令的函数 |
| JSON | 排行榜存盘用的文本格式 |
| rune | Go 的单字符类型（存一个汉字也没问题） |
| nil | Go 的空值（类似别的语言的 null） |
| slice | 动态数组，`[]T` 表示 T 的列表 |
| append | 往 slice 尾部追加 |
| 双缓冲 | 先全部画到内存再一次性输出，防闪烁 |

---

> 最后一句真心话：引擎这种东西，读 10 遍不如改 1 遍。
> 按第 6 节从任务 1 开始，改完 5 个任务，你就真的"看懂"了。
> 有看不懂的地方，带着问题来问，比硬啃效率高十倍。
