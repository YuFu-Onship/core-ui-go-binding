package main

import (
	"encoding/json"
	"fmt"
	"image/color"

	"github.com/YuFu-Onship/core-ui-go-binding/coreui"
	"github.com/YuFu-Onship/core-ui-go-binding/coreui/ui"
)

func (g *gallery) pageView(title string, children ...any) *ui.Node {
	return ui.Column(
		ui.Text(title).FontSize(24).Bold(),
		ui.Column(children...).Gap(16),
	).Padding(24).Gap(18)
}

func card(title string, children ...any) *ui.Node {
	return ui.Card(
		ui.Text(title).FontSize(15).Bold(),
		ui.Column(children...).Gap(10),
	).Gap(10)
}

func kv(label string, value any) *ui.Node {
	return ui.Row(
		ui.Text(label).Width(132).Color(ui.Fg3),
		ui.Text(value).Grow(),
	).Gap(8).Items(ui.AlignCenter)
}

func swatch(label string, c color.RGBA) *ui.Node {
	return ui.Row(
		ui.Stack().Size(18, 18).Bg(c).Radius(4).Border(1, ui.BorderColor),
		ui.Text(label).FontSize(12).Color(ui.Fg2),
		ui.Spacer(),
		ui.Text(ui.FromColor(c)).FontSize(10).Color(ui.Fg3),
	).Gap(8).Items(ui.AlignCenter)
}

// ---- 总览 ---------------------------------------------------------------

func (g *gallery) overviewPage() *ui.Node {
	return g.pageView("总览",
		ui.Column(
			ui.Text(g.clock).FontSize(46).Bold(),
		).CenterContent().Padding(32).Grow(),
	)
}

// ---- 布局 ---------------------------------------------------------------

func (g *gallery) layoutPage() *ui.Node {
	box := func(text string, bg color.RGBA) *ui.Node {
		return ui.Stack(ui.Text(text)).Height(64).Radius(8).Bg(bg).CenterContent()
	}
	return g.pageView("布局",
		card("Flex",
			ui.Row(
				box("flex: 1", ui.Bg2).Grow(),
				box("flex: 2", ui.Bg3).Flex(2),
				box("120px", ui.Bg4).Width(120),
			).Gap(10),
		),
		card("对齐",
			ui.Row(
				box("start", ui.Bg2).Width(90),
				box("center", ui.Bg3).Width(90),
				box("end", ui.Bg4).Width(90),
			).Gap(8).Justify(ui.JustifySpaceBetween),
		),
		card("卡片",
			ui.Row(
				ui.Card(ui.Text("Card A").Bold(), ui.Text("12px").Color(ui.Fg3)).Grow(),
				ui.Card(ui.Text("Card B").Bold(), ui.Text("gutter").Color(ui.Fg3)).Grow(),
				ui.Card(ui.Text("Card C").Bold(), ui.Text("theme").Color(ui.Fg3)).Grow(),
			).Gap(12),
		),
	)
}

// ---- 文本与按钮 ---------------------------------------------------------

func (g *gallery) textPage() *ui.Node {
	return g.pageView("文本与按钮",
		card("文本",
			ui.Text("Hello, ", g.name).FontSize(20),
			kv("name =", g.name),
		),
		card("按钮",
			ui.Row(
				ui.Button("+1").Primary().OnClick(func() { g.count.Inc() }),
				ui.Button("重置").OnClick(func() { g.count.Set(0) }),
				ui.Button("禁用").Disabled(true),
			).Gap(8),
			kv("count =", g.count),
		),
	)
}

// ---- 输入控件 -----------------------------------------------------------

func (g *gallery) inputsPage() *ui.Node {
	return g.pageView("输入控件",
		card("输入",
			ui.Row(
				ui.Text("文本").Width(72).Color(ui.Fg3),
				ui.Input(g.name).Grow(),
			).Gap(8).Items(ui.AlignCenter),
			ui.Row(
				ui.Text("密码").Width(72).Color(ui.Fg3),
				ui.Password(g.password).Grow(),
			).Gap(8).Items(ui.AlignCenter),
			kv("密码字符 =", g.pwdLen),
			ui.TextArea(g.notes).Placeholder("多行文本").Height(96).FullWidth(),
		),
		card("布尔与数值",
			ui.Row(
				ui.CheckLabeled("启用", g.agree),
				ui.Toggle("通知", g.enabled),
			).Gap(18).Items(ui.AlignCenter),
			ui.Row(
				ui.Text("音量").Width(72).Color(ui.Fg3),
				ui.Slider(g.volume, 0, 100).Grow(),
				ui.Text(g.volume).Width(48),
			).Gap(8).Items(ui.AlignCenter),
			ui.Progress(g.volume).Grow(),
		),
		card("搜索",
			ui.Row(
				ui.Input(g.search).Placeholder("搜索").OnSubmit(g.runSearch).Grow(),
				ui.Button("搜索").Primary().OnClick(g.runSearch),
			).Gap(8),
			kv("结果 =", g.searchInfo),
		),
	)
}

// ---- 选择与条件 ---------------------------------------------------------

func (g *gallery) selectionConditionsPage() *ui.Node {
	segment := func(i int, label string) *ui.Node {
		return ui.Button(label).OnClick(func() { g.choice.Set(i) })
	}

	// 条件演示使用普通 Go if/else，不在 gallery 里依赖 When /
	// WhenInt / Unless / ShowWhen / HiddenWhen。状态变化后显式 Rebuild，
	// 避免这些特殊控件在反复挂载/卸载时引入额外的生命周期问题。
	children := []any{
		card("选择",
			ui.Row(
				ui.Text("Select").Width(72).Color(ui.Fg3),
				ui.Select(g.choice, g.choiceOpts).Width(220),
				ui.Text(g.choiceText),
			).Gap(8).Items(ui.AlignCenter),
			ui.Row(
				ui.Text("分段").Width(72).Color(ui.Fg3),
				segment(0, "A"), segment(1, "B"), segment(2, "C"), segment(3, "D"),
			).Gap(8).Items(ui.AlignCenter),
		),
		// 条件与 Rebuild 放在所有条件演示卡片之前。
		card("条件与 Rebuild",
			ui.Row(
				ui.Toggle("showA", g.showA),
				ui.Toggle("showB", g.showB),
				ui.Button("+1").OnClick(func() { g.count.Inc() }),
				ui.Button("Rebuild").OnClick(func() { g.app.Rebuild() }),
				ui.Text(g.count),
			).Gap(12).Items(ui.AlignCenter),
		),
	}

	choice := g.choice.Get()
	if choice == 0 {
		children = append(children, card("当前选择 A", ui.Text("if choice == 0")))
	} else if choice == 1 {
		children = append(children, card("当前选择 B", ui.Text("if choice == 1")))
	} else if choice == 2 {
		children = append(children, card("当前选择 C", ui.Text("if choice == 2")))
	} else {
		children = append(children, card("当前选择 D", ui.Text("if choice == 3")))
	}

	if g.showA.Get() {
		children = append(children, card("v-if A", ui.Text("showA = true")))
	} else {
		children = append(children, card("v-else", ui.Text("showA = false")))
	}
	if g.showB.Get() {
		children = append(children, card("v-show", ui.Text("showB = true")))
	} else {
		children = append(children, card("v-show 反向", ui.Text("showB = false")))
	}

	return g.pageView("选择与条件", children...)
}

// ---- 列表与滚动 ---------------------------------------------------------

func (g *gallery) listsPage() *ui.Node {
	return g.pageView("列表与滚动",
		card("任务列表",
			ui.Row(
				ui.Text(fmt.Sprintf("%d 项", len(g.todos))).Grow(),
				ui.Button("+5").OnClick(func() {
					for i := 0; i < 5; i++ {
						g.todos = append(g.todos, todo{Title: fmt.Sprintf("任务 %d", g.nextID)})
						g.nextID++
					}
					g.app.Rebuild()
				}),
			).Items(ui.AlignCenter),
			ui.Row(
				ui.Input(g.draft).Placeholder("新任务").Grow(),
				ui.Button("添加").Primary().OnClick(g.addTodo),
			).Gap(8),
			ui.Scroll(
				ui.Column(
					ui.Each(g.todos, func(t todo, i int) *ui.Node {
						return ui.Row(
							ui.Text(t.Title).Grow(),
							ui.Button("删除").OnClick(func() {
								g.todos = append(g.todos[:i], g.todos[i+1:]...)
								g.app.Rebuild()
							}),
						).Gap(8).Items(ui.AlignCenter).PaddingXY(0, 6).
							BorderBottom(1, ui.BorderSubtle)
					}),
				).Gap(0).PaddingTRBL(4, 8, 8, 4),
			).Height(420).Grow(),
		),
	)
}

// ---- 事件与诊断 ---------------------------------------------------------

func (g *gallery) eventsPage() *ui.Node {
	return g.pageView("事件与诊断",
		card("事件",
			ui.Row(
				ui.Text("count = ", g.eventCount).FontSize(18),
				ui.Button("click").Primary().OnClick(func() { g.setEvent("click", "") }),
				ui.Button("dblclick").OnDblClick(func() { g.setEvent("dblclick", "") }),
				ui.Button("重置").OnClick(func() {
					g.eventCount.Set(0)
					g.lastEvent.Set("(无事件)")
					g.lastArgs.Set("")
				}),
			).Gap(8).Items(ui.AlignCenter),
			kv("lastEvent =", g.lastEvent),
			kv("lastArgs =", g.lastArgs),
		),
		card("OnChange",
			ui.Row(
				ui.Text("输入").Width(72).Color(ui.Fg3),
				ui.Input(g.name).OnChange(func(args []json.RawMessage) {
					raw := ""
					if len(args) > 0 {
						raw = string(args[0])
					}
					g.setEvent("change", raw)
				}).Grow(),
			).Gap(8).Items(ui.AlignCenter),
		),
		card("Wheel",
			ui.Stack(
				ui.Text("滚轮").Color(ui.Fg2),
			).CenterContent().Height(88).Bg(ui.Bg2).Radius(8).
				OnEvent("wheel", func(args []json.RawMessage) {
					raw := ""
					if len(args) > 0 {
						raw = string(args[0])
					}
					g.setEvent("wheel", raw)
				}),
		),
	)
}

// ---- 颜色 ---------------------------------------------------------------

func (g *gallery) colorsPage() *ui.Node {
	return g.pageView("颜色 Token",
		card("主题 token",
			ui.Row(
				ui.Column(
					swatch("Fg1", ui.Fg1),
					swatch("Fg2", ui.Fg2),
					swatch("Fg3", ui.Fg3),
					swatch("Fg4", ui.Fg4),
				).Grow(),
				ui.Column(
					swatch("Bg1", ui.Bg1),
					swatch("Bg2", ui.Bg2),
					swatch("Bg3", ui.Bg3),
					swatch("Bg4", ui.Bg4),
				).Grow(),
				ui.Column(
					swatch("Accent", ui.Accent),
					swatch("CardBg", ui.CardBg),
					swatch("CardBorder", ui.CardBorder),
					swatch("Border", ui.BorderColor),
				).Grow(),
			).Gap(20),
		),
		card("文本",
			ui.Text("Fg1").Color(ui.Fg1).FontSize(16),
			ui.Text("Fg2").Color(ui.Fg2).FontSize(16),
			ui.Text("Fg3").Color(ui.Fg3).FontSize(16),
			ui.Text("Fg4").Color(ui.Fg4).FontSize(16),
		),
		card("自定义",
			ui.Row(
				ui.Column(
					swatch("color.RGBA", color.RGBA{R: 0x25, G: 0x63, B: 0xeb, A: 0xff}),
					swatch("RGBA alpha", ui.RGBA(0x12, 0x31, 0x75, 0x80)),
					swatch("Transparent", ui.Transparent),
				).Grow(),
				ui.Column(
					swatch("ui.RGB", ui.RGB(0x22, 0xc5, 0x5e)),
					swatch("ui.Accent", ui.Accent),
					swatch("ui.Fg1", ui.Fg1),
				).Grow(),
			).Gap(20),
		),
	)
}

// ---- 设置 ---------------------------------------------------------------

func (g *gallery) settingsPage() *ui.Node {
	return g.pageView("设置",
		card("主题",
			ui.Row(
				ui.Button("切换明暗").Primary().ID("galleryThemeToggle").OnClick(g.toggleTheme),
				ui.Text("当前 = ").Color(ui.Fg3),
				ui.Text(g.themeName).Bold(),
			).Gap(8).Items(ui.AlignCenter),
		),
		card("窗口",
			ui.Row(
				ui.Button("最大化").OnClick(func() { g.app.Window().Maximize() }),
				ui.Button("还原").OnClick(func() { g.app.Window().Restore() }),
				ui.Button("最小化").OnClick(func() { g.app.Window().Minimize() }),
				ui.Button("设置标题").OnClick(func() {
					g.app.Window().SetTitle("Gallery - 标题已修改")
					g.winInfo.Set("SetTitle 已调用")
				}),
				ui.Button("读取 DPI").OnClick(func() {
					g.winInfo.Set(fmt.Sprintf("DPI scale = %.2f", g.app.Window().DPIScale()))
				}),
			).Gap(8),
			kv("状态 =", g.winInfo),
		),
		card("信息",
			kv("DLL 版本 =", coreui.DLLVersion()),
			kv("core-ui =", coreui.Version()),
			kv("合成模式 =", g.compositionName()),
		),
	)
}
