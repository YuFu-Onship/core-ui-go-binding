# AI 声明、与 core-ui 原仓库的差异和运行说明

本文档说明三件事：

1. 本仓库中哪些代码由 AI 编写或修改；
2. 本仓库相对上游 [`ghboke/core-ui`](https://github.com/ghboke/core-ui) 做了哪些改动；
3. 如何准备 `core-ui.dll`，并运行一个最小窗口。

---

## 1. AI 编写说明

本仓库是 **core-ui 的 Go 绑定**，不是 core-ui 上游的替代实现。上游 C++ 代码的版权与许可证仍归 core-ui 原作者所有，详见仓库根目录的 `LICENSE`。

本仓库中以下内容由 **AI（DeepSeek V4.1 Flash）** 在用户需求驱动下编写或修改：

- `coreui/`：Go cgo 绑定、窗口/页面/控件 API、回调、Watch、UTF-8/UTF-16 转换、调试接口。
- `coreui/ui/`：纯 Go 声明式 UI 层（`App` / `Node` / 状态 / 渲染 / 颜色 / 测试），
  它把 Go 节点树生成 `.uix` 源码后再交给 core-ui 的 QuickJS 页面运行时。
- `examples/`：`hello`（`.uix` 示例）、`purego`（纯 Go 双栏列表示例）、
  `gallery`（完整用法画廊）、`simple`（最小纯 Go 窗口）。
- `scripts/`：core-ui 构建、Go 应用构建、窗口截图/点击等 PowerShell/Python 辅助脚本。
- `README.md`、本文件以及示例说明文档。
- `core-ui/` 中对上游 C++ 源码的补丁（见下一节的“修改文件”），用于修复
  Windows 窗口合成、IME、列表性能/内存、控件行为等问题。

上游 core-ui 的原始 C++ 实现、QuickJS 运行时、Direct2D 渲染、控件体系等并非由 AI 从零编写；
AI 只在上游代码基础上做了本仓库列出的补丁。

> 说明：AI 生成的代码已经通过本仓库的 `go build`、`go vet` 和 `coreui/ui` 测试，
> 但仍建议在目标 Windows 版本、DPI 和输入法环境下做一次人工验收。

---

## 2. 与 core-ui 原仓库的区别与改进

### 2.1 项目形态：增加 Go 绑定与纯 Go 声明式层

上游 core-ui 是 C++ / Direct2D / QuickJS 的 Windows UI 框架，页面通过 `.uix` 编写。
本仓库新增：

- `coreui` Go 包：通过 cgo 调用 `core-ui.dll` 的 C ABI，提供窗口、页面、控件、事件、状态和调试 API。
- `coreui/ui` 纯 Go 层：直接写 `ui.Column` / `ui.Row` / `ui.Button` / `ui.Each` 等 Go 代码，
  不需要 `.uix`，也不需要手写 JavaScript；运行时自动生成 core-ui 页面。
- 构建脚本：`scripts/build-core-ui.ps1` 构建 DLL 并刷新 cgo 依赖；
  `scripts/build-app.ps1` 构建 Go 应用并自动复制 DLL，默认使用 Windows GUI 子系统。
- 版本管理：`coreui.DLLVersion()` 读取运行时 DLL 的
  `major.minor.patch.build`，配合 `VersionInfo.AtLeast()` 可在启动时检查
  DLL 是否满足当前绑定要求。

### 2.2 窗口合成模式（重点）

上游默认使用 flip model 交换链。本仓库的纯 Go 层默认改用 **Gio 风格的 bitblt 交换链**：

```text
DXGI_SWAP_EFFECT_DISCARD
BufferCount = 1
DXGI_SCALING_STRETCH
```

原因是 flip + `DXGI_SCALING_NONE` 在 Windows 拖动/缩放窗口时更容易出现白边、旧帧或边缘闪烁；
bitblt 让 DWM 看到普通重定向表面，窗口背景更稳定。

- 纯 Go 层默认：bitblt。
- 需要低延迟 flip 时：`ui.Options.FlipSwapChain = true`。
- 直接使用 `coreui` 包时：`coreui.WindowConfig.CopySwapChain = true` 显式选择 bitblt，
  默认仍保持 flip 以兼容上游行为。

同时修复/改进了窗口生命周期：

- **Aero 空白窗口**：以前用控制台子系统构建时会额外出现一个空的 `ConsoleWindowClass`
  窗口，看起来像 classic Aero 空白窗。现在构建脚本默认 `-H=windowsgui`，
  纯 Go 层也会隐藏进程自己创建的 console 窗口。
- **系统边框窗口不使用 layered 淡入**：`WS_EX_LAYERED` 只用于无边框窗口的淡入/揭示，
  系统边框窗口改用普通 show，避免 DWM 重建边框时出现 ghost。
- **DWM 主题同步**：标题栏、边框、文字颜色通过 `ApplyDwmTheme()` 按当前主题设置；
  主题切换和 `Invalidate()` 时刷新。
- **最大化/还原重新布局**：修正 `WM_SIZE` 提前 return 的问题，最大化后页面根控件
  使用真实客户区尺寸重新 `LayoutRoot()`。
- **原生窗口动画**：不再强制关闭 DWM 的最小化/最大化/还原/关闭动画，是否播放遵循系统设置。
- **`SetRoot` 清理裸指针**：替换根控件时清空 hovered / pressed / focused / tooltip /
  drag / menu 等旧树裸指针，避免快速删除列表行时出现 `0xc0000005`。

### 2.3 IME（输入法）改进

上游的 IME 组合文本由 Windows 默认的黑色组合框/候选框处理。本仓库改为：

- `TextInput` / `TextArea` **内联显示预编辑文本**（带下划线），不再弹默认黑色组合框。
- 窗口处理 `WM_IME_COMPOSITION` 后返回 `TRUE`，阻止 Windows 再画一遍默认组合窗口。
- 候选窗口通过 `ImmSetCandidateWindow(CFS_EXCLUDE + rcArea)` 定位到光标附近，
  不再使用旧的固定偏移（旧实现会在高 DPI 或不同控件位置出现明显偏移）。
- 焦点离开文本控件（点击空白、按钮、切换页面）时调用
  `ImmAssociateContext(hwnd, nullptr)` 解除 IME context；重新聚焦文本控件时恢复。
  这样失焦后不会在旧光标位置残留黑色组合/候选框。
- 提供 `Window.DebugIme*`、`DebugBlur`、`DebugImeEnabled` 等调试接口，便于自动化验证。

### 2.4 控件与列表改进

- **Select / ComboBox 崩溃修复**：`onSelectionChanged` 过去在 `ForEachWidget`
  遍历中同步执行，回调里触发 `Rebuild` 会在遍历中挂载/卸载 `WhenInt` 分支，
  导致崩溃。现在回调延迟到遍历结束后执行；关闭下拉时同步重置 `hoveredIndex_`。
- **ScrollView 滚动条**：绘制顺序改为“背景 → 裁剪后的 content → 滚动条 thumb”，
  滚动条放在右侧专用 gutter（需要滚动时预留 12px），不会压住卡片和按钮；
  thumb 为中性灰，hover/drag 时只变亮，不再使用 accent 蓝。
- **嵌套滚动不再撑大外层**：外层 ScrollView 计算内容高度时，在嵌套 ScrollView
  的边界停止，不把内层列表的滚动内容算进外层高度。内层列表溢出时不会再额外
  触发页面级滚动条。
- **列表 Rebuild 性能**：
  - `ui.App.Rebuild()` 不再在 `SetRoot` 之后重复 `Relayout`；
    `SetRoot` 已经完成布局，一次 Rebuild 只布局一遍。
  - `Renderer` 缓存文本测量结果（文本 + 字号 + 字重 + 字体/宽度），
    避免列表布局反复创建 DirectWrite text layout。
  - 实测 2000 行 Rebuild 从约 950ms 降到约 110ms，100 行约 5ms。
- **Rebuild 内存泄漏修复**：`ui_page_root()` 过去每次调用都向全局 `HandleTable`
  插入一个新的 root 强引用；App 每次 `Rebuild` 都会通过 `SetRoot(page.Root())`
  取 root，导致每个旧页面的整棵控件树一直被句柄表钉住，直到进程退出。
  现在每个页面只保留一个 root handle，并在 `ui_page_destroy` 中释放；
  反复增删列表项 / Rebuild 后内存只在峰值附近波动。
  可用 `coreui.DebugHandleCount()` 检查操作前后的存活句柄数。
- **文本控件 / 列表交互**：增加 Enter 提交（`OnSubmit`）、文本双绑、列表
  `Each` + `Rebuild` 等纯 Go 用法；`examples/gallery` 覆盖这些场景。
- **颜色 API 收紧**：`Color` / `Bg` / `Border*` 只接受 `color.RGBA`；
  `ui.Fg1`、`ui.Bg2`、`ui.Accent` 等主题 token 也改为 `color.RGBA`，
  不再接受字符串或任意 `any`。`App.SetTheme` 会刷新纯 Go 颜色 token
  并自动触发一次 Rebuild。

### 2.5 受影响的 C++ 文件

本仓库对上游 `core-ui/` 的修改主要集中在：

```text
core-ui/include/ui_core.h
core-ui/src/ui/controls.cpp
core-ui/src/ui/controls.h
core-ui/src/ui/handle_table.h
core-ui/src/ui/page/page_api.cpp
core-ui/src/ui/renderer.cpp
core-ui/src/ui/renderer.h
core-ui/src/ui/ui_api.cpp
core-ui/src/ui/ui_window.cpp
core-ui/src/ui/ui_window.h
core-ui/src/ui/widget.h
```

其余上游文件保持原样。

---

## 3. 运行前提：必须提供 `core-ui.dll`

本绑定通过 cgo 调用 core-ui 的 C ABI，**必须使用动态库 `core-ui.dll`，不能只靠
`go build` 静态完成**。

原因：

- core-ui 只由 MSVC 编译（Direct2D / DirectWrite / QuickJS 等依赖）；
- Go/cgo 在 Windows 上通常使用 MinGW gcc，直接链接 C++ 目标文件容易 ABI 冲突；
- 因此采用 DLL + C ABI 边界，Go 侧通过 MinGW 兼容的导入库调用。

### 3.1 DLL 放在哪里

运行时，`core-ui.dll` 必须和 exe 在同一目录（或位于系统 DLL 搜索路径中）：

```text
dist/
  simple.exe
  core-ui.dll
```

仓库中已经提供：

```text
third_party/coreui/bin/core-ui.dll
third_party/coreui/lib/libcore-ui.dll.a   # cgo / MinGW 链接用
third_party/coreui/lib/core-ui.lib        # MSVC 链接用（可选）
third_party/coreui/include/ui_core.h      # cgo 头文件
```

`scripts/build-app.ps1` 会自动把 DLL 复制到输出目录；如果手动 `go build`，
需要自己复制：

```powershell
go build -o dist/simple.exe ./examples/simple
copy third_party\coreui\bin\core-ui.dll dist\
dist\simple.exe
```

### 3.2 重新构建 DLL（可选）

修改了 `core-ui/` 下的 C++ 源码后，需要重新构建 DLL：

```powershell
scripts/build-core-ui.ps1
```

脚本会使用 Visual Studio / CMake 编译 `core-ui.dll`，并同步更新
`third_party/coreui` 下的 DLL、导入库和头文件。构建 Go 应用：

```powershell
scripts/build-app.ps1 -Target ./examples/simple -Out dist/simple.exe
```

默认输出 GUI 子系统程序；需要控制台日志时加 `-Console`。

### 3.3 环境要求

| 项目 | 要求 |
|---|---|
| 系统 | Windows 10/11 x64 |
| Go | 1.24+ |
| C++ 构建 | Visual Studio（C++ 桌面开发工作负载）+ CMake 3.20+ |
| cgo 链接 | MSYS2 UCRT64 的 gcc（或等价 MinGW-w64 工具链） |
| DLL | 运行时必须存在，且与头文件/导入库来自同一次构建 |

---

## 4. 最小窗口示例（纯 Go）

完整代码在 [`examples/simple/main.go`](examples/simple/main.go)：

```go
package main

import (
	"os"

	"github.com/YuFu-Onship/core-ui-go-binding/coreui"
	"github.com/YuFu-Onship/core-ui-go-binding/coreui/ui"
)

func main() {
	app := ui.NewApp(ui.Options{
		Title:    "Core UI + Go",
		Width:    520,
		Height:   320,
		Centered: true,
		Theme:    coreui.Light,
	})

	count := app.Int(0)
	app.View(func() *ui.Node {
		return ui.Column(
			ui.Text("Hello, core-ui + Go").FontSize(22).Bold(),
			ui.Text("纯 Go 声明式窗口，不需要 .uix 或 JavaScript。").Color(ui.Fg2),
			ui.Row(
				ui.Text("点击次数："),
				ui.Text(count).Bold(),
				ui.Button("+1").Primary().OnClick(func() { count.Inc() }),
				ui.Button("重置").OnClick(func() { count.Set(0) }),
			).Gap(10).Items(ui.AlignCenter),
		).Padding(28).Gap(18)
	})

	if code := app.Run(); code != 0 {
		os.Exit(code)
	}
}
```

运行：

```powershell
# 1) 确认 DLL 已存在（仓库自带）
Test-Path third_party\coreui\bin\core-ui.dll

# 2) 构建并复制 DLL
scripts/build-app.ps1 -Target ./examples/simple -Out dist/simple.exe

# 3) 启动
dist\simple.exe
```

如果只想验证最小组件而不用脚本：

```powershell
go build -o dist/simple.exe ./examples/simple
copy third_party\coreui\bin\core-ui.dll dist\
.\dist\simple.exe
```

---

## 5. 目录速览

```text
core-ui-go-binding/
├── coreui/                 Go cgo 绑定
│   └── ui/                 纯 Go 声明式 UI 层
├── core-ui/                修改后的 core-ui C++ 源码（上游 1.9.0 基线）
├── examples/
│   ├── simple/             最小纯 Go 窗口（本文示例）
│   ├── hello/              Go + .uix 示例
│   ├── purego/             纯 Go 双栏列表示例
│   └── gallery/            完整用法画廊
├── third_party/coreui/     cgo 头文件、导入库、运行时 DLL
├── scripts/                构建与窗口调试脚本
├── go.mod
├── README.md
└── AI_AND_DIFFERENCES.md   本文件
```

---

## 6. 已知边界

- core-ui 的 `Each` / `v-for` 仍会创建列表中的全部 widget，不是虚拟列表；
  几千行以上建议分页或后续实现真正的虚拟列表。
- 列表 `Rebuild` 会重建页面，当前滚动位置会重置。
- `ChipsInput` 尚未接入内联 IME 预编辑。
- 原生窗口动画是否播放取决于 Windows 系统设置。
- 本绑定只支持 Windows；其他平台无法运行。
