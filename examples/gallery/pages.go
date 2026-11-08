package main

import (
	"encoding/json"
	"fmt"
	"image/color"

	"github.com/YuFu-Onship/core-ui-go-binding/coreui"
	"github.com/YuFu-Onship/core-ui-go-binding/coreui/ui"
)

// ---- 公共小组件 ---------------------------------------------------------

func (g *gallery) pageView(title, subtitle string, children ...any) *ui.Node {
	return ui.Column(
		ui.Column(
			ui.Text(title).FontSize(24).Bold(),
			ui.Text(subtitle).FontSize(13).Color(ui.Fg3),
		).Gap(4),
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

func swatch(label string, c any) *ui.Node {
	value := fmt.Sprint(c)
	if cc, ok := c.(color.Color); ok {
		value = ui.FromColor(cc)
	}
	return ui.Row(
		ui.Stack().Size(18, 18).Bg(c).Radius(4).Border(1, ui.BorderColor),
		ui.Text(label).FontSize(12).Color(ui.Fg2),
		ui.Spacer(),
		ui.Text(value).FontSize(10).Color(ui.Fg3),
	).Gap(8).Items(ui.AlignCenter)
}

// ---- 总览 ---------------------------------------------------------------

func (g *gallery) overviewPage() *ui.Node {
	return g.pageView(
		"用法画廊 · 总览",
		"左侧每个页面只演示一类能力。把问题缩小到某一页，再对照 README 的检查项。",
		ui.Row(
			card("状态与事件",
				ui.Row(
					ui.Text("count = ", g.count).FontSize(18),
					ui.Button("+1").Primary().OnClick(func() { g.count.Inc() }),
					ui.Button("重置").OnClick(func() { g.count.Set(0) }),
				).Gap(8).Items(ui.AlignStart),
				kv("lastEvent =", g.lastEvent),
				kv("eventCount =", g.eventCount),
			).Grow(),
			card("快速跳转",
				ui.Text("布局：检查 Row/Column、flex、padding/gap").Color(ui.Fg2),
				ui.Text("输入：检查 v-model 回写与 password/textarea").Color(ui.Fg2),
				ui.Text("列表：检查 ScrollView 滚动条 gutter 与 Rebuild").Color(ui.Fg2),
				ui.Text("IME：检查内联 preedit / 提交 / 取消").Color(ui.Fg2),
			).Grow(),
		).Gap(16),

		card("建议的排查顺序",
			ui.Text("1. 先看「布局 Layout」是否只在该页异常。").Color(ui.Fg2),
			ui.Text("2. 输入/选择页只依赖 v-model 与 Watch，可隔离页面重建问题。").Color(ui.Fg2),
			ui.Text("3. 列表页专门用于快速点击删除、滚动、批量插入。").Color(ui.Fg2),
			ui.Text("4. 窗口页用于最大化/还原/主题切换后的布局与原生动画。").Color(ui.Fg2),
			ui.Text("5. 颜色页集中检查 token 和 image/color 转换。").Color(ui.Fg2),
		),
	)
}

// ---- 布局 ---------------------------------------------------------------

func (g *gallery) layoutPage() *ui.Node {
	box := func(text string, bg any) *ui.Node {
		return ui.Stack(ui.Text(text).Color(ui.Fg1)).Height(64).Radius(8).Bg(bg).CenterContent()
	}
	return g.pageView(
		"布局 Layout",
		"Row / Column / flex / padding / gap / 对齐 / 卡片 / 分隔线。",
		card("Row 与 flex 权重",
			ui.Row(
				box("flex: 1", ui.Bg2).Grow(),
				box("flex: 2", ui.Bg3).Flex(2),
				box("120px", ui.Bg4).Width(120),
			).Gap(10),
			ui.Text("第三个固定 120px；前两个按 1:2 分配剩余宽度。").FontSize(12).Color(ui.Fg3),
		),
		card("padding / gap / 对齐",
			ui.Stack(
				ui.Row(
					box("start", ui.Bg2).Width(90),
					box("center", ui.Bg3).Width(90),
					box("end", ui.Bg4).Width(90),
				).Gap(8).Justify(ui.JustifySpaceBetween),
			).Padding(12).Bg(ui.Bg2).Radius(10),
			ui.Row(
				ui.Button("普通按钮"),
				ui.Button("主按钮").Primary(),
				ui.Button("禁用").Disabled(true),
				ui.Spacer(),
				ui.Text("右侧由 Spacer 推开").FontSize(12).Color(ui.Fg3),
			).Gap(8).Items(ui.AlignCenter),
		),
		card("卡片与分隔线",
			ui.Row(
				ui.Card(ui.Text("Card A").Bold(), ui.Text("内边距 16 / 圆角 10").Color(ui.Fg3)).Grow(),
				ui.Card(ui.Text("Card B").Bold(), ui.Text("由 ui.Card 生成").Color(ui.Fg3)).Grow(),
				ui.Card(ui.Text("Card C").Bold(), ui.Text("主题背景 + 边框").Color(ui.Fg3)).Grow(),
			).Gap(12),
			ui.Divider(),
			ui.Text("Divider / Spacer / CenterContent 都是普通节点组合。").FontSize(12).Color(ui.Fg3),
		),
	)
}

// ---- 文本与按钮 ---------------------------------------------------------

func (g *gallery) textPage() *ui.Node {
	return g.pageView(
		"文本与按钮",
		"H1/H2/H3、文本插值、按钮变体、双击、Tooltip、禁用状态。",
		card("文本层级",
			ui.H1("H1 标题"),
			ui.H2("H2 标题"),
			ui.H3("H3 标题"),
			ui.Text("Fg1 正文").Color(ui.Fg1),
			ui.Text("Fg2 次级文字").Color(ui.Fg2),
			ui.Text("Fg3 辅助文字").Color(ui.Fg3),
			ui.Text("Fg4 最弱文字").Color(ui.Fg4),
			ui.Text("Hello, ", g.name, "！文本里可以直接插入状态。").FontSize(15),
		),
		card("按钮与事件",
			ui.Row(
				ui.Text("count = ", g.count).FontSize(18),
				ui.Button("+1").Primary().OnClick(func() {
					g.count.Inc()
					g.setEvent("click", "count="+fmt.Sprint(g.count.Get()))
				}),
				ui.Button("-1").OnClick(func() { g.count.Dec() }),
				ui.Button("归零").OnClick(func() { g.count.Set(0) }),
			).Gap(8).Items(ui.AlignCenter),
			ui.Row(
				ui.Text("双击次数 = ", g.dblClicks),
				ui.Button("双击我").OnDblClick(func() {
					g.dblClicks.Inc()
					g.setEvent("dblclick", "双击按钮")
				}),
				ui.Button("Tooltip").Tooltip("这是原生 tooltip，由 core-ui 绘制"),
				ui.Button("禁用按钮").Disabled(true),
			).Gap(10).Items(ui.AlignCenter),
			ui.Text("tooltip / disabled / primary 都是 Node 修饰器。").FontSize(12).Color(ui.Fg3),
		),
	)
}

// ---- 输入控件 -----------------------------------------------------------

func (g *gallery) inputsPage() *ui.Node {
	return g.pageView(
		"输入控件",
		"文本、密码、多行输入、复选框、开关、滑块、进度条，以及双向回写。",
		card("文本输入",
			ui.Row(
				ui.Text("名字").Width(96).Color(ui.Fg3),
				ui.Input(g.name).Placeholder("输入名字").Grow(),
			).Gap(8).Items(ui.AlignCenter),
			kv("name =", g.name),
			ui.Row(ui.Text("密码").Width(96).Color(ui.Fg3), ui.Password(g.password).Placeholder("输入密码").Grow()).Gap(8).Items(ui.AlignCenter),
			kv("密码字符数 =", g.pwdLen),
			ui.Text("多行文本").Color(ui.Fg3),
			ui.TextArea(g.notes).Placeholder("输入多行备注……").Height(96).FullWidth(),
			kv("notes 字符数 =", g.notesLen),
		),
		card("行内搜索（Enter 提交）",
			ui.Row(
				ui.Text("搜索").Width(96).Color(ui.Fg3),
				ui.Input(g.search).ID("gallerySearch").Placeholder("输入关键字，按 Enter 搜索").Grow().
					OnSubmit(g.runSearch),
				ui.Button("搜索").Primary().OnClick(g.runSearch),
			).Gap(8).Items(ui.AlignCenter),
			kv("search =", g.search),
			kv("结果 =", g.searchInfo),
			ui.Text("TextInput 聚焦时按 Enter 会触发 @submit；适合搜索框、命令框、行内重命名。").
				FontSize(12).Color(ui.Fg3),
		),
		card("布尔与数值",
			ui.Row(
				ui.CheckLabeled("同意条款", g.agree).Color(ui.Fg1),
				ui.Toggle("", g.enabled),
				ui.Text("启用"),
			).Gap(10).Items(ui.AlignCenter),
			kv("agree =", g.agree),
			kv("enabled =", g.enabled),
			ui.Row(
				ui.Text("音量").Width(96).Color(ui.Fg3),
				ui.Slider(g.volume, 0, 100).Grow(),
				ui.Text(g.volume).Width(48),
			).Gap(10).Items(ui.AlignCenter),
			ui.Row(
				ui.Text("进度").Width(96).Color(ui.Fg3),
				ui.Progress(g.volume).Grow(),
				ui.Text(g.volume).Width(48),
			).Gap(10).Items(ui.AlignCenter),
			ui.Row(
				ui.Text("固定进度").Width(96).Color(ui.Fg3),
				ui.Progress(g.progress).Grow(),
				ui.Text(g.progress).Width(48),
			).Gap(10).Items(ui.AlignCenter),
		),
	)
}

// ---- 选择控件 -----------------------------------------------------------

func (g *gallery) selectionPage() *ui.Node {
	segment := func(i int, label string) *ui.Node {
		return ui.Button(label).OnClick(func() { g.choice.Set(i) })
	}
	return g.pageView(
		"选择控件",
		"Select 下拉、分段按钮、WhenInt 页面内切换、Toggle / Check。",
		card("Select 与分段选择",
			ui.Row(
				ui.Text("Select").Width(96).Color(ui.Fg3),
				ui.Select(g.choice, g.choiceOpts).Width(220),
				ui.Text(g.choiceText).Color(ui.Fg2),
			).Gap(10).Items(ui.AlignCenter),
			kv("choice 下标 =", g.choice),
			ui.Row(
				ui.Text("分段按钮").Width(96).Color(ui.Fg3),
				segment(0, "A"), segment(1, "B"), segment(2, "C"), segment(3, "D"),
			).Gap(8).Items(ui.AlignCenter),
		),
		ui.WhenInt(g.choice, 0, card("当前选择：A", ui.Text("WhenInt(choice, 0, ...) 只在 choice == 0 时挂载。").Color(ui.Fg2))),
		ui.WhenInt(g.choice, 1, card("当前选择：B", ui.Text("这是第二个 v-if 分支，切换时节点会真正卸载/重建。").Color(ui.Fg2))),
		ui.WhenInt(g.choice, 2, card("当前选择：C", ui.Text("如果观察到状态丢失，优先检查这里。").Color(ui.Fg2))),
		ui.WhenInt(g.choice, 3, card("当前选择：D", ui.Text("分段按钮和 Select 共享同一个 IntState。").Color(ui.Fg2))),
	)
}

// ---- 列表与滚动 ---------------------------------------------------------

func (g *gallery) listsPage() *ui.Node {
	left := ui.Card(
		ui.Row(
			ui.Text("任务列表").FontSize(15).Bold().Grow(),
			ui.Text(fmt.Sprintf("%d 项", len(g.todos))).FontSize(12).Color(ui.Fg3),
		).Items(ui.AlignCenter),
		ui.Row(
			ui.Input(g.draft).Placeholder("新任务，回车或点添加").Grow(),
			ui.Button("添加").Primary().OnClick(g.addTodo),
			ui.Button("批量 +10").OnClick(func() {
				for i := 0; i < 10; i++ {
					g.todos = append(g.todos, todo{Title: fmt.Sprintf("批量任务 #%d", g.nextID)})
					g.nextID++
				}
				g.addLog("批量添加 10 个任务")
				g.app.Rebuild()
			}),
		).Gap(8).Items(ui.AlignCenter),
		ui.Scroll(
			ui.Column(
				ui.Each(g.todos, func(t todo, i int) *ui.Node {
					mark := "○"
					if t.Done {
						mark = "✓"
					}
					label := t.Title
					if t.Done {
						label += "（已完成）"
					}
					return ui.Row(
						ui.Button(mark).Width(34).OnClick(func() {
							g.todos[i].Done = !g.todos[i].Done
							g.addLog("切换任务：" + g.todos[i].Title)
							g.app.Rebuild()
						}),
						ui.Text(label).Grow(),
						ui.Button("删除").OnClick(func() {
							title := g.todos[i].Title
							g.todos = append(g.todos[:i], g.todos[i+1:]...)
							g.addLog("删除任务：" + title)
							g.app.Rebuild()
						}),
					).Gap(8).Items(ui.AlignCenter).PaddingXY(0, 6).
						BorderBottom(1, ui.BorderSubtle)
				}),
			).Gap(0).PaddingTRBL(4, 8, 8, 4),
		).Height(300).Grow(),
	).Grow()

	right := ui.Card(
		ui.Row(
			ui.Text("操作日志").FontSize(15).Bold().Grow(),
			ui.Text(fmt.Sprintf("%d 条", len(g.logs))).FontSize(12).Color(ui.Fg3),
			ui.Button("+5").OnClick(func() {
				for i := 0; i < 5; i++ {
					g.addLog(fmt.Sprintf("手动日志 #%d", g.nextID))
					g.nextID++
				}
				g.app.Rebuild()
			}),
			ui.Button("清空").OnClick(func() {
				g.logs = g.logs[:0]
				g.app.Rebuild()
			}),
		).Gap(6).Items(ui.AlignCenter),
		ui.Text("这里的两个列表各自滚动；快速连点删除用于排查 Rebuild 生命周期问题。").FontSize(12).Color(ui.Fg3),
		ui.Scroll(
			ui.Column(
				ui.Each(g.logs, func(l logEntry, i int) *ui.Node {
					return ui.Row(
						ui.Text(l.Time).Width(60).FontSize(11).Color(ui.Fg3),
						ui.Text(l.Text).Grow().FontSize(12),
					).Gap(8).Items(ui.AlignCenter).PaddingXY(0, 5).
						BorderBottom(1, ui.BorderSubtle)
				}),
			).Gap(0).PaddingTRBL(4, 8, 8, 4),
		).Height(300).Grow(),
	).Grow()

	return g.pageView(
		"列表与滚动",
		"Each + Rebuild、两个独立 ScrollView、右侧 gutter、快速点击删除。",
		ui.Row(left, right).Gap(16).Grow(),
		ui.Text("如果在这里快速点击删除会崩溃，记录最后一条日志和操作顺序，便于复现 SetRoot/旧树指针问题。").
			FontSize(12).Color(ui.Fg3),
	)
}

// ---- 条件与 Rebuild -----------------------------------------------------

func (g *gallery) conditionsPage() *ui.Node {
	return g.pageView(
		"条件与 Rebuild",
		"v-if（挂载/卸载）与 v-show（仅切 visible）的差异。",
		card("开关",
			ui.Row(
				ui.Toggle("", g.showA),
				ui.Text("showA（v-if 条件）"),
				ui.Toggle("", g.showB),
				ui.Text("showB（v-show 可见性）"),
			).Gap(12).Items(ui.AlignCenter),
			ui.Text("切换页面/ReBuild 后 count 仍然保留：状态在 Go，不在控件树里。").FontSize(12).Color(ui.Fg3),
			ui.Row(
				ui.Text("count = ", g.count).FontSize(18),
				ui.Button("+1").OnClick(func() { g.count.Inc() }),
				ui.Button("Rebuild 当前视图").OnClick(func() {
					g.addLog("手动 Rebuild")
					g.app.Rebuild()
				}),
			).Gap(8).Items(ui.AlignCenter),
		),
		ui.When(g.showA, card("v-if 分支 A",
			ui.Text("这个卡片由 When(showA, ...) 控制，关闭时会真正卸载。").Color(ui.Fg2),
		)),
		ui.Unless(g.showA, card("v-if 的 Else 分支",
			ui.Text("showA 为 false 时显示，由 Unless 生成。").Color(ui.Fg2),
		)),
		ui.ShowWhen(g.showB, card("v-show 子树",
			ui.Text("这个卡片的 widget 始终存在，只是 visible=false。").Color(ui.Fg2),
		)),
		ui.HiddenWhen(g.showB, card("v-show 的反向",
			ui.Text("showB 为 false 时这个卡片可见。").Color(ui.Fg2),
		)),
	)
}

// ---- 事件与诊断 ---------------------------------------------------------

func (g *gallery) eventsPage() *ui.Node {
	return g.pageView(
		"事件与诊断",
		"OnClick / OnDblClick / OnChange / OnEvent / 滚轮事件和原始参数。",
		card("事件计数",
			ui.Row(
				ui.Text("eventCount = ", g.eventCount).FontSize(18),
				ui.Button("click").Primary().OnClick(func() { g.setEvent("click", "") }),
				ui.Button("dblclick").OnDblClick(func() { g.setEvent("dblclick", "") }),
				ui.Button("清空事件").OnClick(func() {
					g.eventCount.Set(0)
					g.lastEvent.Set("(已清空)")
					g.lastArgs.Set("")
				}),
			).Gap(8).Items(ui.AlignCenter),
			kv("lastEvent =", g.lastEvent),
			kv("lastArgs =", g.lastArgs),
		),
		card("OnChange 原始参数",
			ui.Row(
				ui.Text("输入并失焦/回车").Width(160).Color(ui.Fg3),
				ui.Input(g.name).OnChange(func(args []json.RawMessage) {
					raw := ""
					if len(args) > 0 {
						raw = string(args[0])
					}
					g.setEvent("change", raw)
				}).Grow(),
			).Gap(8).Items(ui.AlignCenter),
			kv("name =", g.name),
		),
		card("OnEvent(\"wheel\")",
			ui.Stack(
				ui.Text("把鼠标移到这个区域并滚动滚轮").Color(ui.Fg2),
				ui.Text("事件会记录到下方日志和 lastEvent。").FontSize(12).Color(ui.Fg3),
			).CenterContent().Height(88).Bg(ui.Bg2).Radius(8).
				OnEvent("wheel", func(args []json.RawMessage) {
					raw := ""
					if len(args) > 0 {
						raw = string(args[0])
					}
					g.setEvent("wheel", raw)
				}),
		),
		card("事件日志",
			ui.Row(
				ui.Text("最近事件").Bold().Grow(),
				ui.Button("添加到日志").OnClick(func() {
					g.addLog("诊断按钮点击")
					g.app.Rebuild()
				}),
			).Items(ui.AlignCenter),
			ui.Scroll(
				ui.Column(ui.Each(g.logs, func(l logEntry, i int) *ui.Node {
					return ui.Row(
						ui.Text(l.Time).Width(60).FontSize(11).Color(ui.Fg3),
						ui.Text(l.Text).Grow().FontSize(12),
					).Gap(8).Items(ui.AlignCenter).PaddingXY(0, 4)
				})).Gap(0),
			).Height(220),
		),
	)
}

// ---- IME ----------------------------------------------------------------

func (g *gallery) imePage() *ui.Node {
	win := func() coreui.Window { return g.app.Window() }
	focusIme := func() {
		p := g.app.Page()
		if p == nil {
			return
		}
		w := win()
		if !w.DebugFocus(p.Root().FindByID("imeInput")) {
			g.imeInfo.Set("聚焦失败：找不到 imeInput")
		}
	}
	return g.pageView(
		"IME 输入法",
		"内联 preedit、提交、取消；也可以直接用系统输入法在这个页面输入。",
		card("IME 输入框",
			ui.Row(
				ui.Text("单行").Width(72).Color(ui.Fg3),
				ui.Input(g.imeText).ID("imeInput").Placeholder("点击这里，然后使用输入法或下方按钮").Grow(),
			).Gap(8).Items(ui.AlignCenter),
			kv("imeText =", g.imeText),
			ui.Row(
				ui.Button("聚焦输入框").OnClick(focusIme),
				ui.Button("模拟预编辑 ni'hao").OnClick(func() {
					if win().DebugImeComposition("ni'hao", 6) {
						g.imeInfo.Set("已注入预编辑：ni'hao（此时状态还不应回写）")
					} else {
						g.imeInfo.Set("注入失败：先点击输入框聚焦")
					}
				}),
				ui.Button("提交 你好").OnClick(func() {
					if win().DebugImeCommit("你好") {
						g.imeInfo.Set("已提交：你好")
					} else {
						g.imeInfo.Set("提交失败")
					}
				}),
				ui.Button("取消组合").OnClick(func() {
					win().DebugImeEnd()
					g.imeInfo.Set("已取消组合")
				}),
				ui.Button("失焦（模拟点空白）").OnClick(func() {
					win().DebugBlur()
					g.imeInfo.Set("已失焦：现在输入法不应再出现组合/候选窗")
				}),
				ui.Button("检查 IME 状态").OnClick(func() {
					g.imeInfo.Set(fmt.Sprintf("IME context enabled = %v", win().DebugImeEnabled()))
				}),
			).Gap(8),
			kv("imeInfo =", g.imeInfo),
		),
		card("多行 IME",
			ui.TextArea(g.notes).Placeholder("也可以在这里测试多行输入法的换行/组合").Height(100).FullWidth(),
			kv("notes 字符数 =", g.notesLen),
		),
		card("检查点",
			ui.Text("预编辑阶段：输入框内显示带下划线的拼音，Go 状态不应变化。").Color(ui.Fg2),
			ui.Text("提交阶段：状态更新一次，日志/其它绑定随后刷新。").Color(ui.Fg2),
			ui.Text("取消阶段：预编辑消失，原有选区/文本恢复。").Color(ui.Fg2),
			ui.Text("失焦阶段：点「失焦（模拟点空白）」后，IME context 应为 false，再按键不应出现黑框/候选窗。").Color(ui.Fg2),
		),
	)
}

// ---- 窗口与主题 ---------------------------------------------------------

func (g *gallery) windowPage() *ui.Node {
	return g.pageView(
		"窗口与主题",
		"主题切换、最大化/还原/最小化、标题、DPI 读取。",
		card("主题",
			ui.Row(
				ui.Button("切换明/暗主题").Primary().ID("galleryThemeToggle").OnClick(g.toggleTheme),
				ui.Text("当前主题 = ").Color(ui.Fg3),
				ui.Text(g.themeName).Bold(),
			).Gap(8).Items(ui.AlignCenter),
			ui.Text("这里切换的是 core-ui 全局主题，页面 CSS 变量会整体刷新。").FontSize(12).Color(ui.Fg3),
			ui.Text("合成模式 = "+g.compositionName()).FontSize(12).Color(ui.Fg3),
		),
		card("窗口控制",
			ui.Row(
				ui.Button("最大化").OnClick(func() { g.app.Window().Maximize() }),
				ui.Button("还原").OnClick(func() { g.app.Window().Restore() }),
				ui.Button("最小化（之后从任务栏恢复）").OnClick(func() { g.app.Window().Minimize() }),
				ui.Button("设置标题").OnClick(func() {
					g.app.Window().SetTitle("Gallery · 标题已修改")
					g.winInfo.Set("SetTitle 已调用")
				}),
				ui.Button("读取 DPI").OnClick(func() {
					g.winInfo.Set(fmt.Sprintf("DPI scale = %.2f", g.app.Window().DPIScale()))
				}),
			).Gap(8),
			kv("winInfo =", g.winInfo),
			ui.Text("最大化后检查内容是否铺满；最小化/还原依赖系统窗口动画设置。").FontSize(12).Color(ui.Fg3),
		),
	)
}

// ---- 颜色 ---------------------------------------------------------------

func (g *gallery) colorsPage() *ui.Node {
	return g.pageView(
		"颜色 Token",
		"主题变量常量、Fg/Bg 层级、以及 image/color 到 CSS 颜色的转换。",
		card("主题 token",
			ui.Row(
				ui.Column(
					swatch("Fg1", ui.Fg1),
					swatch("Fg2", ui.Fg2),
					swatch("Fg3", ui.Fg3),
					swatch("Fg4", ui.Fg4),
					swatch("FgOnAccent", ui.FgOnAccent),
				).Grow(),
				ui.Column(
					swatch("Bg1", ui.Bg1),
					swatch("Bg2", ui.Bg2),
					swatch("Bg3", ui.Bg3),
					swatch("Bg4", ui.Bg4),
					swatch("BgElevated", ui.BgElevated),
				).Grow(),
				ui.Column(
					swatch("Accent", ui.Accent),
					swatch("CardBg", ui.CardBg),
					swatch("CardBorder", ui.CardBorder),
					swatch("BorderColor", ui.BorderColor),
					swatch("InputBorder", ui.InputBorder),
				).Grow(),
			).Gap(20),
		),
		card("文本层级",
			ui.Text("Fg1：主要文本").Color(ui.Fg1).FontSize(16),
			ui.Text("Fg2：次级文本").Color(ui.Fg2).FontSize(16),
			ui.Text("Fg3：辅助/说明").Color(ui.Fg3).FontSize(16),
			ui.Text("Fg4：最弱提示").Color(ui.Fg4).FontSize(16),
			ui.Row(
				ui.Text("强调色按钮").Bold().Color(ui.FgOnAccent).PaddingTRBL(6, 12, 6, 12).Bg(ui.Accent).Radius(6),
			),
		),
		card("标准库颜色",
			ui.Row(
				ui.Column(
					swatch("color.RGBA", color.RGBA{R: 0x25, G: 0x63, B: 0xeb, A: 0xff}),
					swatch("color.NRGBA", color.NRGBA{R: 0x22, G: 0xc5, B: 0x5e, A: 0xff}),
					swatch("NRGBA alpha", color.NRGBA{R: 0x25, G: 0x63, B: 0xeb, A: 0x80}),
				).Grow(),
				ui.Column(
					swatch("ui.RGB", ui.RGB(0x25, 0x63, 0xeb)),
					swatch("ui.RGBA", ui.RGBA(0x25, 0x63, 0xeb, 0x80)),
					swatch("Transparent", ui.Transparent),
				).Grow(),
			).Gap(20),
			ui.Text("Color / Bg / Border 接受主题 token、CSS 字符串和 color.Color。").FontSize(12).Color(ui.Fg3),
		),
	)
}
