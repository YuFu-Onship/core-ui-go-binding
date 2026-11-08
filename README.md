# core-ui Go 绑定

把 [core-ui](https://github.com/ghboke/core-ui) 这个 Windows 桌面 UI 框架接到 Go 上，
支持两种写法：

- **`.uix` + Go**：Vue 3 风格模板，Go 负责加载页面、注入数据和注册原生回调。
- **纯 Go 声明式界面**：不写 `.uix`、不写 JavaScript，直接用 `ui.Column` /
  `ui.Row` / `ui.Button` 等 Go 代码描述界面。

> **运行前必须有 `core-ui.dll`**
>
> 本绑定通过 cgo 调用已编译好的 core-ui 动态库，不能只靠 `go build` 静态生成。
> Go 编译出的 exe 旁边必须存在 `core-ui.dll`，否则启动会失败。
> 仓库自带一份：`third_party/coreui/bin/core-ui.dll`。
>
> core-ui 本身的 C++ 构建方式请参考上游仓库；本项目专注 Go 侧的使用。

> **与 core-ui 原仓库差异较大**
>
> 除了 Go 绑定，本项目还修改了 core-ui 的部分 C++ 实现：
>
> - **窗口模式**：纯 Go 层默认改用 bitblt 交换链（`DXGI_SWAP_EFFECT_DISCARD` /
>   1 buffer / `DXGI_SCALING_STRETCH`），减少拖动、缩放窗口时的白边和旧帧；
>   flip 模式保留为可选项。
> - **窗口动画**：恢复 DWM 原生最小化 / 最大化 / 还原 / 关闭动画，修正最大化
>   后的重新布局，并同步 DWM 标题栏和边框主题色。
> - **IME 改进**：输入框内联显示预编辑文本，候选窗跟随光标定位；失焦时解除
>   IME context，避免旧位置残留黑色候选框。
> - **控件修复**：Select 下拉重选崩溃、嵌套滚动条、滚动条 gutter、暗色复选框
>   勾选颜色、列表 Rebuild 性能与句柄泄漏等。
>
> 完整的 AI 声明和改动列表见 [`AI_AND_DIFFERENCES.md`](AI_AND_DIFFERENCES.md)。

---

## 开始

环境：

- Windows 10/11 x64
- Go 1.24+
- 已编译好的 `core-ui.dll`

### 通过 `go get` 使用

在自己的 Go 模块中直接拉取：

```powershell
go get github.com/YuFu-Onship/core-ui-go-binding/coreui@latest
```

然后在代码中导入：

```go
import (
	"github.com/YuFu-Onship/core-ui-go-binding/coreui"
	"github.com/YuFu-Onship/core-ui-go-binding/coreui/ui"
)
```

`go get` 会下载 cgo 需要的头文件和导入库；运行时仍然需要
`core-ui.dll`。从模块缓存里取出 DLL：

```powershell
$mod = go list -m -f '{{.Dir}}' github.com/YuFu-Onship/core-ui-go-binding
copy "$mod\third_party\coreui\bin\core-ui.dll" .
```

启动时可以检查实际加载的 DLL 版本：

```go
v := coreui.DLLVersion()
fmt.Println("core-ui.dll:", v) // 1.9.0.345

if !v.AtLeast(coreui.VersionInfo{Major: 1, Minor: 9, Patch: 0, Build: 345}) {
	log.Fatal("core-ui.dll 版本过旧，请更新 DLL")
}
```

### 方式一：`.uix` + Go

`app.uix`：

```html
<window title="Hello" width="520" height="320" centered="true"/>

<script>
export default {
  data() {
    return { count: 0 };
  }
};
</script>

<style>
  .root { padding: 24px; gap: 12px; }
  .row  { flex-direction: row; gap: 8px; align-items: center; }
</style>

<template>
  <div class="root">
    <label>count = {{ count }}</label>
    <div class="row">
      <button @click="count++">+1</button>
      <button @click="count = 0">重置</button>
    </div>
  </div>
</template>
```

`main.go`：

```go
package main

import (
	_ "embed"
	"os"

	"github.com/YuFu-Onship/core-ui-go-binding/coreui"
)

//go:embed app.uix
var appUIX string

func main() {
	if err := coreui.InitWithTheme(coreui.Light); err != nil {
		panic(err)
	}

	page, err := coreui.LoadPageString(appUIX)
	if err != nil {
		coreui.Shutdown()
		panic(err)
	}

	if _, err := page.OpenWindow(nil); err != nil {
		page.Destroy()
		coreui.Shutdown()
		panic(err)
	}

	code := coreui.Run()
	page.Destroy()
	coreui.Shutdown()
	os.Exit(code)
}
```

### 方式二：纯 Go

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
		).Padding(24).Gap(16)
	})

	if code := app.Run(); code != 0 {
		os.Exit(code)
	}
}
```

### 编译运行

```powershell
go build -o dist/app.exe ./examples/simple

# exe 同目录必须有 core-ui.dll
copy third_party\coreui\bin\core-ui.dll dist\

dist\app.exe
```

更多示例：

| 示例 | 说明 |
|---|---|
| `examples/simple` | 最小纯 Go 窗口 |
| `examples/hello` | Go + `.uix` |
| `examples/purego` | 纯 Go 双栏列表 |
| `examples/gallery` | 完整用法画廊 |

---

## 能力与局限

直接 cgo 封装的 C 函数约 **15%**（83 / 566）；但常用窗口、页面和基础控件
场景覆盖约 **80%–90%**。原因是纯 Go 层通过生成 `.uix` 页面来使用大量控件，
不需要为每个 C 控件函数单独写 Go 包装。

| 能力 | 状态 | 粗略覆盖 |
|---|---|---|
| 初始化 / 消息循环 / 窗口 / DPI / 主题 / 页面 | ✅ 可用 | ~90% |
| `.uix` 页面、响应式状态、Watch、事件、i18n、资源 | ✅ 可用 | ~90% |
| 常用布局与控件（Text / Button / Input / TextArea / Check / Toggle / Slider / Progress / Select / Scroll / Card / Divider / Spacer） | ✅ 可用 | ~85% |
| 窗口动画、DWM 主题、IME、调试 API（截图 / 点击 / 按键 / 句柄 / 滚动信息） | ✅ 已改进或部分可用 | ~50% |
| 菜单 / ContextMenu | ⚠ 部分绑定 | ~40% |
| ImageView / 裁剪 / 分块渲染 / GhImgView / OcrImgView | ❌ 未绑定 | ~5% |
| CustomWidget / 绘图 API / Bitmap / Canvas 模式 | ❌ 未绑定 | 0% |
| `.ui` 布局文件 API | ❌ 未绑定 | 0% |

已知局限：

- `Each` / `v-for` 会创建列表中的全部 widget，不是虚拟列表；数千行建议分页。
- 列表 `Rebuild` 会重建页面，滚动位置会重置。
- `ChipsInput` 暂未接入内联 IME 预编辑。
- 只支持 Windows，且运行时必须有 `core-ui.dll`。

---

## 控件示例

### 文本、按钮与状态

```go
count := app.Int(0)

ui.Column(
	ui.Text("count = ", count).FontSize(18),
	ui.Row(
		ui.Button("+1").Primary().OnClick(func() { count.Inc() }),
		ui.Button("归零").OnClick(func() { count.Set(0) }),
	).Gap(8),
).Gap(12)
```

### 输入、布尔、数值与下拉选择

```go
name := app.Text("")
agree := app.Bool(false)
volume := app.Float(50)
choice := app.Int(0)

options := []string{"选项 A", "选项 B", "选项 C"}

ui.Column(
	ui.Input(name).Placeholder("输入名字"),
	ui.Text("你好，", name, "！"),

	ui.CheckLabeled("同意条款", agree),
	ui.Toggle("启用通知", agree),

	ui.Row(
		ui.Text("音量").Width(72),
		ui.Slider(volume, 0, 100).Grow(),
		ui.Text(volume).Width(40),
	).Gap(8).Items(ui.AlignCenter),

	ui.Row(
		ui.Text("选择").Width(72),
		ui.Select(choice, options).Width(180),
		ui.Text(choice),
	).Gap(8).Items(ui.AlignCenter),
).Gap(12)
```

### 列表与滚动

```go
todos := []string{"写代码", "测试", "发布"}

ui.Scroll(
	ui.Column(
		ui.Each(todos, func(item string, i int) *ui.Node {
			return ui.Row(
				ui.Text(item).Grow(),
				ui.Button("删除").OnClick(func() {
					todos = append(todos[:i], todos[i+1:]...)
					app.Rebuild()
				}),
			).Gap(8).Items(ui.AlignCenter)
		}),
	).Gap(6),
).Height(240)
```

### 条件渲染与主题

```go
showDetails := app.Bool(false)

ui.Column(
	ui.Button("显示/隐藏").OnClick(func() { showDetails.Toggle() }),
	ui.When(showDetails, ui.Text("这里是条件内容")),
	ui.Button("切换暗色").OnClick(func() { app.SetTheme(coreui.Dark) }),
	ui.Button("切换浅色").OnClick(func() { app.SetTheme(coreui.Light) }),
).Gap(10)
```

### 颜色

颜色参数统一使用 Go 标准库的 `color.RGBA`，主题 token 本身也是
`color.RGBA`：

```go
import "image/color"

ui.Text("主文字").Color(ui.Fg1)       // 主题 token
ui.Text("错误").Color(ui.RGB(0xd1, 0x34, 0x38)) // 自定义颜色
ui.Text("半透明").Color(ui.RGBA(0x12, 0x31, 0x75, 0x80))
ui.Column().Bg(ui.CardBg).Border(1, ui.BorderColor)
ui.Column().Bg(color.RGBA{R: 0x25, G: 0x63, B: 0xeb, A: 0xff})
```

`Color` / `Bg` / `Border*` 不再接受字符串或任意 `any` 类型；需要自定义颜色时
使用 `color.RGBA`、`ui.RGB(...)` 或 `ui.RGBA(...)`。

状态变化通常**不需要** `Rebuild`；只有列表增删、页面结构变化时才调用
`app.Rebuild()`。后台 goroutine 修改界面请使用 `app.Post(func(){ ... })`。
切换主题请使用 `app.SetTheme(...)`：它会同步刷新纯 Go 颜色 token 并自动
触发一次 Rebuild。
