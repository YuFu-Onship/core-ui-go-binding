// Command gallery 是 core-ui Go 绑定 / 纯 Go 声明式层的用法画廊。
package main

import (
	"fmt"
	"os"
	"strings"
	"sync"
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
	pageEvents
	pageColors
	pageSettings
)

var pageTitles = []string{
	"总览",
	"布局",
	"文本与按钮",
	"输入控件",
	"选择与条件",
	"列表与滚动",
	"事件与诊断",
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

	page *ui.IntState

	count      *ui.IntState
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

	search     *ui.TextState
	searchInfo *ui.TextState
	searchData []string

	showA *ui.BoolState
	showB *ui.BoolState

	draft  *ui.TextState
	todos  []todo
	logs   []logEntry
	nextID int

	eventCount *ui.IntState
	lastEvent  *ui.TextState
	lastArgs   *ui.TextState
	diag       *ui.TextState

	dark      bool
	themeName *ui.TextState
	winInfo   *ui.TextState

	clock     *ui.TextState
	clockOnce sync.Once

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
		searchInfo: app.Text(""),
		searchData: []string{
			"core-ui 窗口层",
			"core-ui IME 内联组合",
			"coreui Go 声明式绑定",
			"Go 布局 Layout",
			"Go 列表与滚动",
			"Go 事件总线",
			"Windows 原生动画",
			"Direct2D 渲染",
			"QuickJS 响应式页面",
		},
		showA:      app.Bool(true),
		showB:      app.Bool(false),
		draft:      app.Text(""),
		eventCount: app.Int(0),
		lastEvent:  app.Text("(无事件)"),
		lastArgs:   app.Text(""),
		diag:       app.Text(""),
		themeName:  app.Text("浅色"),
		winInfo:    app.Text(""),
		clock:      app.Text(time.Now().Format("2006-01-02 15:04:05")),
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
		// 选择与条件页使用普通 Go if/else，需要重新执行 View。
		g.app.Rebuild()
	})
	g.showA.OnChange(func(bool) { g.app.Rebuild() })
	g.showB.OnChange(func(bool) { g.app.Rebuild() })

	g.todos = []todo{
		{Title: "检查 Row / Column 嵌套", Done: true},
		{Title: "检查 padding / gap", Done: true},
		{Title: "检查 flex: 1 分配"},
		{Title: "检查滚动条 gutter", Done: true},
		{Title: "检查深色主题"},
		{Title: "检查输入控件回写", Done: true},
		{Title: "检查 v-if 挂载/卸载"},
		{Title: "检查 v-show 可见性"},
		{Title: "检查事件参数"},
		{Title: "检查最大化布局", Done: true},
		{Title: "检查最小化/还原"},
	}
	g.nextID = len(g.todos) + 1

	app.OnPage(func(*coreui.Page) { g.startClock() })
	return g
}

func (g *gallery) startClock() {
	g.clockOnce.Do(func() {
		go func() {
			ticker := time.NewTicker(time.Second)
			defer ticker.Stop()
			for range ticker.C {
				g.app.Post(func() {
					g.clock.Set(time.Now().Format("2006-01-02 15:04:05"))
				})
			}
		}()
	})
}

func (g *gallery) addLog(text string) {
	g.logs = append(g.logs, logEntry{
		Time: time.Now().Format("15:04:05"),
		Text: text,
	})
	if g.diag != nil {
		g.diag.Set(text)
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
	g.searchInfo.Set(fmt.Sprintf("「%s」匹配 %d 条", q, matches))
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
			ui.Column(g.navItems()...).Gap(1),
			ui.Spacer(),
			g.navButton(pageSettings, "设置"),
		).Width(200).Bg(ui.SidebarBg).Padding(12)

		pages := ui.Column(
			ui.WhenInt(g.page, pageOverview, g.overviewPage()),
			ui.WhenInt(g.page, pageLayout, g.layoutPage()),
			ui.WhenInt(g.page, pageText, g.textPage()),
			ui.WhenInt(g.page, pageInputs, g.inputsPage()),
			ui.WhenInt(g.page, pageSelection, g.selectionConditionsPage()),
			ui.WhenInt(g.page, pageLists, g.listsPage()),
			ui.WhenInt(g.page, pageEvents, g.eventsPage()),
			ui.WhenInt(g.page, pageColors, g.colorsPage()),
			ui.WhenInt(g.page, pageSettings, g.settingsPage()),
		).Gap(0)

		return ui.Row(
			nav,
			ui.Stack().Width(1).Bg(ui.BorderColor),
			ui.Scroll(pages).Grow(),
		).Grow()
	})
}

func (g *gallery) navItems() []any {
	items := make([]any, 0, len(pageTitles))
	for i, title := range pageTitles {
		items = append(items, g.navButton(i, title))
	}
	return items
}

func (g *gallery) navButton(i int, title string) *ui.Node {
	return ui.Row(
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
		OnClick(func() { g.page.Set(i) })
}

func (g *gallery) compositionName() string {
	if g.copySwap {
		return "bitblt / DISCARD"
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
