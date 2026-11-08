# core-ui Go 绑定

把 [core-ui](https://github.com/ghboke/core-ui) 这个 Windows 桌面 UI 框架接到 Go 上，
提供两种写法：

1. **`.uix` 单文件组件**：Vue 3 风格模板，Go 负责加载、注入数据、注册原生回调。
2. **纯 Go 声明式界面**（`coreui/ui`）：不写 `.uix`、不写 JavaScript，界面直接
   用 `ui.Column` / `ui.Row` / `ui.Button` 等 Go 表达式声明，运行时自动翻译成
   core-ui 模板。

两种方式共用同一个 `core-ui.dll` 运行时，逻辑、构建、打包都交给 Go 和 `go build`。

> **AI 声明**：本仓库的 Go 绑定层、纯 Go UI 层、示例、脚本、文档，以及
> `core-ui/` 中列出的 C++ 补丁，由 **AI（DeepSeek V4.1 Flash）** 在用户需求下
> 编写或修改。上游 core-ui 的原始 C++ 代码版权归原作者所有。
> 详细的 AI 范围、相对上游的差异、窗口合成 / IME / 控件改进，见
> [`AI_AND_DIFFERENCES.md`](AI_AND_DIFFERENCES.md)。

> **运行前必须准备 `core-ui.dll`**：本绑定通过 cgo 调用 core-ui 的动态库，
> 不能只靠 `go build` 静态生成。请把 `third_party/coreui/bin/core-ui.dll`
> 复制到 exe 同目录（`scripts/build-app.ps1` 会自动复制）。

```
core-ui-go-binding/
├── coreui/                      Go 绑定包 (cgo)
│   └── ui/                      ★ 纯 Go 声明式界面层
├── examples/simple/             ★ 最小纯 Go 窗口示例
├── examples/hello/              go + .uix 的完整示例
│   ├── main.go                  Go 宿主代码
│   └── app.uix                  Vue 风格单文件组件
├── examples/purego/             ★ 纯 Go 声明式示例（无 .uix / 无 JS）
├── examples/gallery/            ★ 纯 Go 用法画廊（按能力分页排查问题）
├── core-ui/                     修改后的 core-ui C++ 源码（上游 1.9.0 基线）
├── third_party/coreui/          cgo 需要的三方产物
│   ├── include/ui_core.h
│   ├── lib/libcore-ui.dll.a     MinGW 导入库（cgo 链接用）
│   ├── lib/core-ui.lib          MSVC 导入库（可选）
│   └── bin/core-ui.dll          运行时（必须和 exe 放一起）
├── AI_AND_DIFFERENCES.md        AI 声明 / 差异与改进 / 最小示例
└── scripts/
    ├── build-core-ui.ps1        编译 C++ 运行时并刷新 third_party
    ├── build-app.ps1            go build + 复制 DLL（默认 GUI 子系统）
    ├── capture-window.ps1       截窗口图（调试用，支持窗口被遮挡）
    ├── click-window.ps1         模拟点击窗口坐标（调试用）
    ├── notify.py                Python 发 Windows Toast（无第三方依赖）
    └── window-test.py           启动 gallery 并发通知，附手动测试清单
```

## 环境要求

| 项 | 说明 |
|----|------|
| Windows x64 | core-ui 只支持 Windows 10+ 与 MSVC / clang-cl 工具链 |
| Visual Studio | 需要“使用 C++ 的桌面开发”工作负载（编译 core-ui.dll） |
| MSYS2 (UCRT64) | cgo 用其 gcc；另需 `gendef` / `dlltool` 生成导入库 |
| CMake 3.20+ / Go 1.24+ | |

## 第一次构建

```powershell
# 1) 编译 core-ui.dll 并生成 cgo 需要的头文件 / 导入库
scripts/build-core-ui.ps1            # 加 -Clean 重新配置，-DynamicCRT 用 /MD

# 2) 编译 Go 应用（产物 + core-ui.dll 都在 dist/）
scripts/build-app.ps1                # 默认 examples/hello -> dist/hello.exe
scripts/build-app.ps1 -Target ./examples/hello -Out dist/hello.exe
scripts/build-app.ps1 -Target ./examples/hello -Out dist/hello-debug.exe -Console
```

`build-app.ps1` 现在默认加 `-H=windowsgui`，双击运行时不会额外弹出一个
空的 `ConsoleWindowClass`（它常被误认为“经典 Aero 空白窗口”）。需要看
stdout/stderr 调试时加 `-Console`；`coreui/ui` 也会在运行时安全地隐藏
“Windows 为本进程新建的”控制台，但从已有终端启动时不会动那个终端。

也可以直接用 go 命令，只要 exe 旁边有 `core-ui.dll`：

```powershell
go build -o dist/hello.exe ./examples/hello
copy third_party/coreui/bin/core-ui.dll dist\
dist\hello.exe
```

## 最小代码

```go
//go:embed app.uix
var appUIX string

func main() {
    coreui.InitWithTheme(coreui.Light)
    page, err := coreui.LoadPageString(appUIX)   // 也可以 LoadPageFile("app.uix")
    if err != nil { ... }

    page.SetText("nativeMsg", "来自 Go")          // 写响应式变量
    page.SetJSON("items", `[{"id":1,"label":"a"}]`)
    win, _ := page.OpenWindow(nil)                // 窗口尺寸/标题取 .uix <window>

    if btn := page.Root().FindByID("btnNative"); btn.Handle() != 0 {
        btn.OnClick(func() { wins++ })            // Go 闭包挂在原生控件上
    }

    code := coreui.Run()                          // Win32 消息循环
    page.Destroy()
    coreui.Shutdown()
    os.Exit(code)
}
```

## 纯 Go 声明式界面（`coreui/ui`）

如果不想写 `.uix`、也不想碰 JavaScript，可以直接用 `coreui/ui`。它把
gpui / iced / Flutter 那种“界面是状态的函数”的写法翻译成 core-ui 的
`.uix` 源码再加载，业务代码保持 100% Go。

```go
package main

import (
    "os"

    "github.com/YuFu-Onship/core-ui-go-binding/coreui"
    "github.com/YuFu-Onship/core-ui-go-binding/coreui/ui"
)

func main() {
    app := ui.NewApp(ui.Options{
        Title: "Counter", Width: 520, Height: 320,
        MinWidth: 360, MinHeight: 240,
        Centered: true, Theme: coreui.Light,
    })

    // 状态在 Go 里，自动同步到页面响应式数据。
    count := app.Int(0)
    name := app.Text("世界")

    app.View(func() *ui.Node {
        return ui.Column(
            ui.Text("count = ", count).FontSize(20),
            ui.Row(
                ui.Button("+1").Primary().OnClick(func() { count.Inc() }),
                ui.Button("归零").OnClick(func() { count.Set(0) }),
            ).Gap(8),
            ui.Input(name).Placeholder("输入名字"),
            ui.Text("你好，", name, "！"),
        ).Padding(24).Gap(12)
    })

    os.Exit(app.Run())
}
```

也可以写成更紧凑的形式：

```go
ui.Run(opts, func(a *ui.App) {
    count := a.Int(0)
    a.View(func() *ui.Node { return ui.Text(count) })
})
```

### 能力一览

| 类别 | API |
|------|-----|
| 布局 | `Column` / `Row` / `Stack` / `Scroll` / `Card` / `Spacer` / `Divider` |
| 控件 | `Text` / `Label` / `H1`~`H3` / `Button` / `Input` / `Password` / `TextArea` / `Check` / `Toggle` / `Slider` / `Progress` / `Select` / `Image` |
| 修饰 | `Padding*` / `Gap` / `Width` / `Height` / `Size` / `Min*` / `Max*` / `Grow` / `Flex` / `Items` / `Justify` / `Bg` / `Color` / `FontSize` / `Bold` / `Radius` / `Opacity` / `Border` / `Shadow` / `Primary` / `Tooltip` / `Hidden` / `Disabled` / `Class` / `ID` / `Style` |
| 条件 | `When` / `Unless`（v-if，挂载/卸载）、`ShowWhen` / `HiddenWhen`（v-show） |
| 列表 | `Each(slice, func(v, i) *Node)`；增删后调用 `app.Rebuild()` |
| 状态 | `app.Int` / `Float` / `Bool` / `Text`；`Get` / `Set` / `Add` / `Inc` / `Dec` / `Toggle` / `OnChange` |
| 事件 | `OnClick` / `OnDblClick` / `OnEvent` / `OnChange` / `OnSubmit`（Enter） |
| 运行时 | `Run` / `Quit` / `Rebuild` / `Post` / `SetTheme` / `SetLocale` / `OnPage` / `LoadLanguageString` / `LoadLanguageFile` |
| 调试 | `app.Source()` 查看生成的 `.uix`；`app.Window()` / `app.Page()` 拿底层句柄 |

### 颜色

`Color` / `Bg` / `Border*` 现在接受三种值：

1. 主题 token 常量：`ui.Fg1`~`ui.Fg4`、`ui.Bg1`~`ui.Bg4`、`ui.CardBg`、
   `ui.CardBorder`、`ui.Accent`、`ui.BorderColor`、`ui.InputBorder` 等；
2. 普通 CSS 字符串：`"#2563eb"`、`"rgba(37,99,235,0.5)"`；
3. Go 标准库颜色：`color.RGBA` / `color.NRGBA` / `image.Uniform` 等。

```go
ui.Text("次要文字").Color(ui.Fg3)
ui.Column(...).Bg(ui.CardBg).Border(1, ui.BorderSubtle)
ui.Text("●").Color(color.RGBA{R: 0x22, G: 0xc5, B: 0x5e, A: 0xff})
```

不确定用什么 token 时，也可以直接转换：

```go
ui.FromColor(color.NRGBA{R: 0x25, G: 0x63, B: 0xeb, A: 0x80})
ui.RGB(0x25, 0x63, 0xeb)        // #2563eb
ui.RGBA(0x25, 0x63, 0xeb, 0x80) // #2563eb80
```

### 行内搜索 / Enter 提交

`TextInput` / `TextArea` 内按 Enter 会触发 `submit` 事件。Go 侧直接用
`OnSubmit`：

```go
search := app.Text("")

ui.Row(
    ui.Input(search).Placeholder("搜索").OnSubmit(func() {
        q := strings.TrimSpace(search.Get())
        // 执行搜索，把结果写回另一个状态
    }).Grow(),
    ui.Button("搜索").OnClick(func() { /* 同一个回调也可以复用 */ }),
).Gap(8)
```

示例见 `examples/gallery` 的「输入控件 → 行内搜索」卡片。

### 双向绑定是怎么走的

- **Go → 界面**：`count.Set(5)` 直接调用 `ui_page_set_int`，页面里的
  `{{ s1 }}` / `v-model` 立即刷新，不重载页面。
- **界面 → Go**：输入框、开关、滑块、下拉用 `v-model` 写页面状态，
  `coreui.Page.Watch` 再把新值回写到 `IntState` / `FloatState` /
  `BoolState` / `TextState`，并触发 `OnChange`。
- **事件**：`OnClick` 等回调不在 C 侧直接注册，而是通过一次
  `__emit(id, ...)` → `ui_page_watch("__evt")` 的总线分发。这样即使在
  `v-for` 动态行里、或者回调中立刻修改状态，也不会破坏 QuickJS 的求值栈。
- **线程**：core-ui 的控件和消息循环绑定在 UI 线程上。后台 goroutine 想改
  界面时用 `app.Post(func(){ count.Inc() })`，不要直接调用状态 setter。

### 什么时候需要 `Rebuild`

普通状态变化**不需要** Rebuild。只有界面结构变了（列表增删、页面切换、
根据 Go 值增删节点）才调用 `app.Rebuild()`：它会重新执行 `View` 函数、
生成新页面并替换窗口根控件；所有 `IntState` 等状态值会保留。Rebuild 是
异步且可合并的，可以放心在按钮回调里调用。

约束：状态要在 `View` 之前创建（不要在 `View` 内部调用 `app.Int`），
否则每次 Rebuild 都会新建一批变量。

### 关于 DLL

纯 Go 指的是**应用源码**：界面、状态、事件、构建都不需要 `.uix` / JS。
`core-ui.dll` 仍然是 core-ui 运行时的本体（C++ / Direct2D），所以 exe
旁边依旧要带它；`coreui/ui` 只是把“写 `.uix`”这一步省掉了。如果想彻底
去掉 DLL，需要换一个纯 Go 渲染后端，而不是继续包 core-ui。

构建纯 Go 示例：

```powershell
scripts/build-app.ps1 -Target ./examples/purego -Out dist/purego.exe
dist\purego.exe

# 用法画廊：按“布局 / 输入 / 列表 / IME / 窗口 / 颜色”分页排查问题
scripts/build-app.ps1 -Target ./examples/gallery -Out dist/gallery.exe
dist\gallery.exe

# 默认就是 Gio 风格的 bitblt（DXGI DISCARD / 1 buffer / STRETCH）
dist\gallery.exe

# A/B：切回 core-ui 的 flip 交换链
$env:COREUI_FLIP_SWAPCHAIN = "1"
dist\gallery.exe
```

纯 Go 层 `ui.Options.FlipSwapChain` 默认 false，即 bitblt；设为 true 才走
flip。直接用 `coreui.WindowConfig` 时，`CopySwapChain` 仍是显式 opt-in（默认
flip），避免影响 core-ui 原有的图像/画布高刷新路径。bitblt 主要解决窗口
拖动/缩放时 flip 交换链与 DWM 合成不同步导致的白边/旧帧。

`examples/gallery/README.md` 里有每一页覆盖的能力和推荐的排查步骤。

### 测试

```powershell
$env:PATH = "$PWD\third_party\coreui\bin;$env:PATH"
go test ./coreui/ui                    # 纯渲染/状态测试，不需要加载 DLL
$env:COREUI_E2E = "1"
go test ./coreui/ui -v                 # 额外跑真实 DLL 的双向绑定/事件测试
```

## 绑定覆盖范围

- 生命周期：`Init` / `InitWithTheme` / `Run` / `Quit` / `Shutdown` / `Version`
- 主题：`SetTheme` / `SetAccentHex` / `DefaultFont` / `SetDefaultFont` /
  `CJKFonts` / `SetCJKFont` / `TextLocale` / `SetTextLocale`
- 页面：`LoadPageFile` / `LoadPageString` / `Reload` / `Root` / `LastError` /
  `OpenWindow` / `PrepareWindow` / `Menu`
- 数据：`SetBool` / `SetInt` / `SetFloat` / `SetText` / `SetTextList` / `SetJSON` /
  `GetJSON` / `GetBool` / `GetInt` / `GetFloat` / `GetText` / `Watch`
- i18n：`LoadTranslations` / `LoadLanguageString` / `LoadLanguageFile` / `SetLocale`
- 资源：`RegisterAssetDir` / `RegisterAssetBlob` / `ResetAssets` / `AssetInvalidate`
- 窗口：`Show*` / `Hide` / `Invalidate` / `Relayout` / `SetTitle` / `SetMinSize` /
  `Maximize` / `Minimize` / `Restore` / `IsMaximized` /
  `SetRoot` / `Destroy` / `DPIScale` / `DPI` / `HWND` / `Post` /
  `OnClose` / `OnCloseRequest`
- 控件：`FindByID` / `AddChild` / `SetText` / `SetVisible` / `SetEnabled` /
  `OnClick` / `Rect`

没包到的 C API（`ui_core.h` 里 400+ 个函数）可以照同款模式补：在 `coreui/*.go` 里
加一个薄封装即可，`unsafe` 的用法参考现有实现。

## 实现要点

- **为什么是 DLL 而不是静态链接**：core-ui 只能用 MSVC 编译，而 cgo 在 Windows 上
  用 MinGW gcc，两边目标文件不能混。所以走 DLL + 纯 C ABI 边界。
- **导入库**：MSVC 的 `.lib` 其实也能被 MinGW ld 直接吃，但 `build-core-ui.ps1`
  仍然用 `gendef` + `dlltool` 生成 `libcore-ui.dll.a`，这条路径更稳。
- **CRT**：默认 `/MT` 静态 CRT，所以 `core-ui.dll` 不依赖 VC 运行时，只依赖系统 DLL。
- **线程**：`Init` / `Run` 会 `runtime.LockOSThread()`。core-ui 的窗口和消息循环
  绑定在一个线程上，goroutine 迁移会引起界面失联。
- **字符串**：`.uix` 源码、变量名、资源名走 UTF-8（`C.CString`）；控件文本和文件路径
  是 `wchar_t`（UTF-16），由 `coreui/utf16.go` 负责转换。
- **回调**：`OnClick` / `OnClose` / `OnCloseRequest` / `Page.Watch` / `Window.Post`
  用“登记号 + C 侧小内存块”把 Go 闭包传给 C，回调里带 recover，避免 panic 穿过
  cgo 边界炸进程。
- **纯 Go 层的实现**：`coreui/ui` 把节点树渲染成完整的
  `<window> + <script> + <style> + <template>`，再走 `ui_page_load_string`；
  状态用 `Set*` 推送、用 `ui_page_watch` 回写。控件事件统一走
  `@click="__emit(id, ...)"` → `__evt` watcher 的总线，所以不依赖控件 id，
  在 `v-for` 动态行和 Rebuild 后都能工作。
- **Rebuild**：重新生成页面并 `ui_window_set_root` 换根，随后销毁旧页面；
  状态值会以当前 Go 值重新写进新页面，翻译表也会自动重放。换根时
  `UiWindowImpl::SetRoot` 会先清空 focused / hovered / pressed / drag /
  tooltip / active menu 等指向旧控件树的裸指针 —— 否则高频点击删除按钮时，
  旧页面销毁后下一次鼠标消息会访问已释放的 widget，表现为
  `0xc0000005` 崩溃。
- **ScrollView 滚动条**：`ScrollViewWidget::DrawTree` 原来先画背景+滚动条再画
  content，列表里的卡片/控件会把 overlay 的 thumb 盖住。现在顺序改为
  “背景 → content（带裁剪）→ 滚动条 thumb”；`DoLayout()` 在需要滚动时还会
  在右侧预留 12px gutter，让滚动条位于内容旁边而不是压在内容上。thumb 颜色
  统一为中性灰，hover/drag 时只是更亮，不再使用 accent 蓝。
  另外，外层 ScrollView 量内容高度时会**在嵌套 ScrollView 的边界停下**，
  不再把内层列表的滚动内容算进外层高度；否则内层一出现滚动条，外层页面
  也会多出一条无关的滚动条。
- **列表性能**：`ui.App.Rebuild()` 不再在 `SetRoot` 之后重复 `Relayout`
  （`SetRoot` 已经布局过），一次 Rebuild 只做一遍完整布局。`Renderer`
  还缓存了文本测量结果（文本 + 字号 + 字重 + 字体），大量列表在 resize /
  Rebuild 时不必反复创建 DirectWrite text layout。实测 2000 行的
  Rebuild 从约 950ms 降到约 110ms；100 行约 5ms。
  但 core-ui 的 `v-for` / `Each` 仍会创建列表中的全部 widget，超大列表
  （数千行）仍建议做分页或虚拟列表，而不是只靠 Rebuild 优化。
- **Rebuild 内存**：`ui_page_root()` 以前每次调用都会把 root 作为新的强引用
  插入全局 `HandleTable`，而 App 每次 `Rebuild` 都通过
  `SetRoot(page.Root())` 取一次 root，结果每个旧页面的整棵控件树都被句柄表
  永久钉住，直到进程退出。现在每个页面只保留一个 root handle，并在
  `ui_page_destroy` 中释放；列表反复增删 / Rebuild 后内存只会在峰值附近
  波动，不再随操作次数线性上涨。可用 `coreui.DebugHandleCount()` 对比操作
  前后的存活句柄数。
- **示例**：`examples/purego` 现在是左右双栏布局，两栏各自拥有独立的
  ScrollView / 滚动条 / 数据列表。
- **窗口合成**：纯 Go 层默认使用 Gio 风格的 bitblt 交换链
  （`DXGI_SWAP_EFFECT_DISCARD` / 1 buffer / `DXGI_SCALING_STRETCH`）。
  它让 DWM 看到普通重定向表面，窗口拖动/缩放时比 flip 的
  `DXGI_SCALING_NONE` 更不容易出现白边/旧帧。需要低延迟 flip 时设置
  `ui.Options.FlipSwapChain = true`；直接使用 `coreui` 时通过
  `WindowConfig.CopySwapChain` 显式开启 bitblt。
- **原生窗口动画**：创建窗口时不再强制
  `DWMWA_TRANSITIONS_FORCEDISABLED=TRUE`，DWM 的最小化 / 最大化 / 还原 /
  关闭动画恢复，是否播放仍遵循系统“在最小化和最大化时动画显示窗口”的设置。
  `WM_SIZE` 处理也修了一处提前返回：以前最大化时如果
  `ClampMaximizedRectToWorkArea()` 只挪动了一个尺寸本来就正确的窗口，会
  直接 return，导致页面根控件仍停在旧尺寸。现在无论是否 clamp 都会用真实
  客户区尺寸重新 `LayoutRoot()`，并重新请求 resize 帧。
- **IME（输入法）**：`TextInput` / `TextArea` 现在把预编辑文本直接画在输入框里
  （带下划线），窗口处理完 `WM_IME_COMPOSITION` 后返回 `TRUE`，所以 Windows
  不会再弹默认的黑色组合窗。候选窗定位参考了 Gio 的
  `ImmSetCandidateWindow(CFS_EXCLUDE + rcArea)`：用预编辑/光标矩形作为排除区，
  而不是旧的 `CFS_CANDIDATEPOS + 固定 +2px`，因此不会出现候选窗离输入框
  固定十几像素的偏移。
  另外，焦点离开文本控件（点击空白/按钮/窗口内失焦）时会
  `ImmAssociateContext(hwnd, nullptr)` 解除 IME context；无文本焦点时
  `WM_IME_*` 也会被吞掉。这样即使窗口仍是前台，按键也不会在旧光标处弹出
  Windows 默认的组合窗/候选窗；重新聚焦文本控件时再恢复 context。
  测试用的 `Window.DebugIme*` / `DebugBlur` / `DebugImeEnabled` 接口也已提供。

## 调试小工具

```powershell
# 起应用后截图 / 点按钮（示例里 "Go 原生回调" 约在窗口内 (69, 229)）
scripts/capture-window.ps1 -ProcessName hello -Out dist/shot.png
scripts/click-window.ps1   -ProcessName hello -OffsetX 69 -OffsetY 229 -Out dist/click.png
```
