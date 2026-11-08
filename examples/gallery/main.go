// Command gallery 是 core-ui Go 绑定 / 纯 Go 声明式层的用法画廊。
//
// 左侧是按主题拆分的页面，每一页只演示一类能力，便于定位问题：
// 布局、文本、输入控件、选择控件、列表滚动、条件渲染、事件、IME、窗口/主题、
// 颜色 token 等。具体测试方式和排查建议见同目录 README.md。
package main

import (
	"fmt"
	"image/color"
	"os"
	"strings"
	"time"

	"github.com/YuFu-Onship/core-ui-go-binding/coreui"
	"github.com/YuFu-Onship/core-ui-go-binding/coreui/ui"
)

const (
	pageOverview = iota
	pageLayout
	pageText
	pageInputs
	pageSelection
	pageLists
	pageConditions
	pageEvents
	pageIME
	pageWindow
	pageColors
)

var pageTitles = []string{
	"总览",
	"布局 Layout",
	"文本与按钮",
	"输入控件",
	"选择控件",
	"列表与滚动",
	"条件与 Rebuild",
	"事件与诊断",
	"IME 输入法",
	"窗口与主题",
	"颜色 Token",
}

type todo struct {
	Title string
	Done  bool
}

type logEntry struct {
	Time string
	Text string
}

type gallery struct {
	app *ui.App

	// 页面路由
	page *ui.IntState

	// 通用 / 文本
	count      *ui.IntState
	dblClicks  *ui.IntState
	name       *ui.TextState
	password   *ui.TextState
	pwdLen     *ui.IntState
	notes      *ui.TextState
	notesLen   *ui.IntState
	agree      *ui.BoolState
	enabled    *ui.BoolState
	volume     *ui.FloatState
	progress   *ui.FloatState
	choice     *ui.IntState
	choiceText *ui.TextState
	choiceOpts []string

	// 行内搜索
	search     *ui.TextState
	searchInfo *ui.TextState
	searchData []string

	// 条件
	showA *ui.BoolState
	showB *ui.BoolState

	// 列表
	draft  *ui.TextState
	todos  []todo
	logs   []logEntry
	nextID int

	// 事件 / 诊断
	eventCount *ui.IntState
	lastEvent  *ui.TextState
	lastArgs   *ui.TextState
	diag       *ui.TextState

	// IME
	imeText *ui.TextState
	imeInfo *ui.TextState

	// 窗口 / 主题
	dark      bool
	themeName *ui.TextState
	winInfo   *ui.TextState

	// 窗口合成模式（默认 Gio 风格 bitblt；COREUI_FLIP_SWAPCHAIN=1 时为 flip）
	copySwap bool
}

func main() {
	flipSwap := os.Getenv("COREUI_FLIP_SWAPCHAIN") == "1"
	app := ui.NewApp(ui.Options{
		Title:         "Core UI Go Gallery",
		Width:         1180,
		Height:        820,
		MinWidth:      900,
		MinHeight:     600,
		Centered:      true,
		Theme:         coreui.Light,
		Accent:        "#2563eb",
		Locale:        "zh-CN",
		FlipSwapChain: flipSwap,
	})

	g := newGallery(app)
	g.copySwap = !flipSwap
	g.setupView()

	os.Exit(app.Run())
}

func newGallery(app *ui.App) *gallery {
	g := &gallery{
		app:        app,
		page:       app.Int(pageOverview),
		count:      app.Int(0),
		dblClicks:  app.Int(0),
		name:       app.Text("World"),
		password:   app.Text(""),
		pwdLen:     app.Int(0),
		notes:      app.Text(""),
		notesLen:   app.Int(0),
		agree:      app.Bool(true),
		enabled:    app.Bool(true),
		volume:     app.Float(35),
		progress:   app.Float(62),
		choice:     app.Int(0),
		choiceText: app.Text("选项 A"),
		choiceOpts: []string{"选项 A", "选项 B", "选项 C", "选项 D"},
		search:     app.Text("core"),
		searchInfo: app.Text("输入关键字后按 Enter，或点右侧搜索按钮。"),
		searchData: []string{
			"core-ui 窗口层",
			"core-ui IME 内联组合",
			"coreui Go 声明式绑定",
			"Go 布局 Layout",
			"Go 列表与滚动",
			"Go 事件总线",
			"Rust gpui 风格声明式 UI",
			"Flutter 风格 Widget 组合",
			"Windows 原生动画",
			"Direct2D 渲染",
			"QuickJS 响应式页面",
		},
		showA:      app.Bool(true),
		showB:      app.Bool(false),
		draft:      app.Text(""),
		eventCount: app.Int(0),
		lastEvent:  app.Text("(还没有事件)"),
		lastArgs:   app.Text(""),
		diag:       app.Text("等待操作"),
		imeText:    app.Text(""),
		imeInfo:    app.Text("点击输入框后，可用下面的按钮模拟 IME 组合。"),
		themeName:  app.Text("浅色"),
		winInfo:    app.Text("窗口 API 演示"),
	}

	g.password.OnChange(func(s string) {
		g.pwdLen.Set(len([]rune(s)))
	})
	g.notes.OnChange(func(s string) {
		g.notesLen.Set(len([]rune(s)))
	})
	g.choice.OnChange(func(i int) {
		if i >= 0 && i < len(g.choiceOpts) {
			g.choiceText.Set(g.choiceOpts[i])
		}
	})

	g.todos = []todo{
		{Title: "检查 Row / Column 嵌套", Done: true},
		{Title: "检查 padding / gap", Done: true},
		{Title: "检查 flex: 1 分配", Done: false},
		{Title: "检查滚动条 gutter", Done: true},
		{Title: "检查深色主题", Done: false},
		{Title: "检查输入控件回写", Done: true},
		{Title: "检查 v-if 挂载/卸载", Done: false},
		{Title: "检查 v-show 可见性", Done: false},
		{Title: "检查事件参数", Done: false},
		{Title: "检查 IME 内联 preedit", Done: false},
		{Title: "检查最大化布局", Done: true},
		{Title: "检查最小化/还原", Done: false},
	}
	g.nextID = len(g.todos) + 1

	for _, s := range []string{
		"gallery 启动",
		"加载 Light 主题",
		"注册页面路由",
		"初始化输入状态",
		"初始化列表数据",
		"注册 IME / 窗口调试按钮",
		"页面状态全部保存在 Go",
		"列表结构变化使用 Rebuild",
		"滚动条使用右侧 gutter",
		"主题 token 跟随明暗模式",
		"输入控件通过 Watch 回写",
		"事件通过 __evt 总线分发",
		"窗口最大化会重新 LayoutRoot",
		"原生动画遵循系统设置",
		"等待用户操作",
		"提示：列表页可快速连点删除",
		"提示：IME 页可模拟预编辑",
		"提示：颜色页可检查 token 转换",
	} {
		g.addLog(s)
	}
	return g
}

func (g *gallery) addLog(text string) {
	g.logs = append(g.logs, logEntry{
		Time: time.Now().Format("15:04:05"),
		Text: text,
	})
	if g.diag != nil {
		g.diag.Set("最近：" + text)
	}
}

func (g *gallery) setEvent(name string, args string) {
	g.lastEvent.Set(name)
	g.lastArgs.Set(args)
	g.eventCount.Inc()
	g.addLog("事件：" + name)
}

func (g *gallery) runSearch() {
	q := strings.TrimSpace(g.search.Get())
	if q == "" {
		g.searchInfo.Set("请输入关键字")
		return
	}
	needle := strings.ToLower(q)
	matches := 0
	for _, s := range g.searchData {
		if strings.Contains(strings.ToLower(s), needle) {
			matches++
		}
	}
	g.searchInfo.Set(fmt.Sprintf("搜索「%s」：模拟匹配 %d 条", q, matches))
	g.addLog("搜索：" + q)
}

func (g *gallery) addTodo() {
	title := strings.TrimSpace(g.draft.Get())
	if title == "" {
		return
	}
	g.todos = append(g.todos, todo{Title: title})
	g.nextID++
	g.draft.Set("")
	g.addLog("新增任务：" + title)
	g.app.Rebuild()
}

func (g *gallery) setupView() {
	g.app.View(func() *ui.Node {
		nav := ui.Column(
			ui.Column(
				ui.Text("GALLERY").FontSize(10).Color(ui.Fg3),
				ui.Text("Core UI Go").FontSize(18).Bold(),
			).Gap(2).PaddingTRBL(4, 8, 12, 8),
			ui.Column(g.navItems()...).Gap(1),
			ui.Spacer(),
			ui.Text("点击左侧切换页面").FontSize(11).Color(ui.Fg3).PaddingTRBL(8, 8, 2, 8),
		).Width(210).Bg(ui.SidebarBg).PaddingTRBL(8, 8, 8, 8)

		pages := ui.Column(
			ui.WhenInt(g.page, pageOverview, g.overviewPage()),
			ui.WhenInt(g.page, pageLayout, g.layoutPage()),
			ui.WhenInt(g.page, pageText, g.textPage()),
			ui.WhenInt(g.page, pageInputs, g.inputsPage()),
			ui.WhenInt(g.page, pageSelection, g.selectionPage()),
			ui.WhenInt(g.page, pageLists, g.listsPage()),
			ui.WhenInt(g.page, pageConditions, g.conditionsPage()),
			ui.WhenInt(g.page, pageEvents, g.eventsPage()),
			ui.WhenInt(g.page, pageIME, g.imePage()),
			ui.WhenInt(g.page, pageWindow, g.windowPage()),
			ui.WhenInt(g.page, pageColors, g.colorsPage()),
		).Gap(0)

		return ui.Column(
			ui.Row(
				ui.Column(
					ui.Text("CORE UI × GO").FontSize(11).Color(ui.Fg3),
					ui.Text("纯 Go 声明式用法画廊").FontSize(24).Bold(),
				).Gap(2).Grow(),
				ui.Text("●").FontSize(12).Color(color.RGBA{R: 0x22, G: 0xc5, B: 0x5e, A: 0xff}),
				ui.Text(g.themeName).FontSize(12).Color(ui.Fg3),
				ui.Button("切换主题").OnClick(g.toggleTheme),
			).Items(ui.AlignCenter).Gap(12).PaddingTRBL(18, 24, 12, 24),

			ui.Row(
				nav,
				ui.Stack().Width(1).Bg(ui.BorderColor),
				ui.Scroll(pages).Grow(),
			).Grow(),

			ui.Row(
				ui.Text("core-ui ", coreui.Version()).FontSize(11).Color(ui.Fg3),
				ui.Spacer(),
				ui.Text(g.diag).FontSize(11).Color(ui.Fg3),
				ui.Button("退出").OnClick(func() { g.app.Quit() }),
			).Items(ui.AlignCenter).Gap(12).PaddingTRBL(10, 24, 10, 24),
		)
	})
}

func (g *gallery) navItems() []any {
	items := make([]any, 0, len(pageTitles))
	for i, title := range pageTitles {
		i, title := i, title
		items = append(items, ui.Row(
			ui.Stack(
				ui.WhenInt(g.page, i, ui.Stack().Width(3).Height(16).Radius(2).Bg(ui.Accent)),
				ui.UnlessInt(g.page, i, ui.Stack().Width(3).Height(16)),
			).Width(3),
			ui.Stack(
				ui.WhenInt(g.page, i, ui.Text(title).Color(ui.Fg1).Bold()),
				ui.UnlessInt(g.page, i, ui.Text(title).Color(ui.Fg3)),
			).Grow(),
		).Gap(9).Items(ui.AlignCenter).PaddingTRBL(7, 10, 7, 6).
			Style("cursor: pointer").
			OnClick(func() { g.page.Set(i) }))
	}
	return items
}

func (g *gallery) compositionName() string {
	if g.copySwap {
		return "bitblt / DISCARD（Gio 风格）"
	}
	return "flip / FLIP_SEQUENTIAL"
}

func (g *gallery) toggleTheme() {
	g.dark = !g.dark
	if g.dark {
		g.app.SetTheme(coreui.Dark)
		g.themeName.Set("深色")
	} else {
		g.app.SetTheme(coreui.Light)
		g.themeName.Set("浅色")
	}
	g.addLog("切换主题：" + g.themeName.Get())
}
