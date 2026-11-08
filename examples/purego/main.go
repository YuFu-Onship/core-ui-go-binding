// Command purego 演示完全用 Go 声明界面，不写 .uix / JavaScript。
//
// 界面是左右两栏列表：两栏各自拥有独立的 ScrollView 和滚动条，
// 滚动条位于内容区域右侧的专用 gutter 内，不再覆盖卡片/控件。
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

type todo struct {
	Title string
	Done  bool
}

type logEntry struct {
	Time string
	Text string
}

func main() {
	app := ui.NewApp(ui.Options{
		Title:         "Core UI · 纯 Go 双栏列表",
		Width:         1080,
		Height:        760,
		MinWidth:      760,
		MinHeight:     520,
		Centered:      true,
		Theme:         coreui.Light,
		Accent:        "#2563eb",
		Locale:        "zh-CN",
		FlipSwapChain: os.Getenv("COREUI_FLIP_SWAPCHAIN") == "1",
	})

	dark := false
	themeName := app.Text("浅色")

	// 左栏：任务列表。
	draft := app.Text("")
	todos := []todo{
		{Title: "梳理 core-ui 窗口层", Done: true},
		{Title: "修复 IME 内联组合", Done: true},
		{Title: "验证最大化后的布局", Done: true},
		{Title: "把滚动条移到内容右侧", Done: true},
		{Title: "重写纯 Go 双栏示例", Done: false},
		{Title: "补充 Window.Maximize 绑定", Done: true},
		{Title: "检查 DWM 原生窗口动画", Done: false},
		{Title: "整理 DebugIme 测试接口", Done: true},
		{Title: "给滚动条做灰色激活态", Done: true},
		{Title: "录制一段演示动画", Done: false},
		{Title: "补 README 的 IME 说明", Done: true},
		{Title: "检查深色主题对比度", Done: false},
		{Title: "验证 125% DPI 布局", Done: false},
		{Title: "给列表加上空状态", Done: false},
		{Title: "整理示例目录结构", Done: false},
		{Title: "准备发布说明", Done: false},
	}
	nextID := len(todos) + 1

	// 右栏：独立的操作日志列表。
	logs := make([]logEntry, 0, 24)
	addLog := func(text string) {
		logs = append(logs, logEntry{Time: time.Now().Format("15:04:05"), Text: text})
	}
	for _, s := range []string{
		"应用启动，加载 Light 主题",
		"注册窗口和页面根控件",
		"启用纯 Go 声明式视图",
		"初始化两个独立 ScrollView",
		"左栏任务数据装载完成",
		"右栏日志缓冲装载完成",
		"滚动条 gutter 预留 12px",
		"滚动条颜色切换为中性灰",
		"最大化后重新 LayoutRoot",
		"恢复原生 DWM 过渡动画",
		"IME 预编辑改为内联绘制",
		"候选窗使用 CFS_EXCLUDE 定位",
		"焦点切换时取消 IME 组合",
		"列表滚动条绘制在内容之后",
		"左栏支持新增 / 完成 / 删除",
		"右栏支持新增 / 清空日志",
		"两栏共享同一个 Go 状态树",
		"结构变化后调用 app.Rebuild()",
		"等待下一次用户操作",
	} {
		addLog(s)
	}

	addTodo := func() {
		title := strings.TrimSpace(draft.Get())
		if title == "" {
			return
		}
		todos = append(todos, todo{Title: title})
		nextID++
		draft.Set("")
		addLog("新增任务：" + title)
		app.Rebuild()
	}

	btn_add := ui.Button("add").Primary().OnClick(addTodo).Color(ui.Fg1)

	app.View(func() *ui.Node {
		// ---- 左栏 ----
		leftPanel := ui.Column(
			ui.Row(
				ui.Text("任务列表").FontSize(17).Bold().Grow(),
				ui.Text(fmt.Sprintf("%d 项", len(todos))).FontSize(12),
			).Items(ui.AlignCenter).PaddingTRBL(18, 18, 8, 18),

			ui.Row(
				ui.Input(draft).Placeholder("输入新任务，回车或点添加").Grow(),
				btn_add,
			).Gap(8).Items(ui.AlignCenter).PaddingTRBL(0, 18, 8, 18),

			ui.Scroll(
				ui.Column(
					ui.Each(todos, func(t todo, i int) *ui.Node {
						mark := "○"
						if t.Done {
							mark = "✓"
						}
						label := t.Title
						textColor := ui.Fg1
						if t.Done {
							label += "（已完成）"
							textColor = ui.Fg3
						}
						return ui.Row(
							ui.Button(mark).Width(34).OnClick(func() {
								todos[i].Done = !todos[i].Done
								if todos[i].Done {
									addLog("完成任务：" + todos[i].Title)
								} else {
									addLog("重新打开：" + todos[i].Title)
								}
								app.Rebuild()
							}),
							ui.Text(label).Grow().Color(textColor),
							ui.Button("删除").OnClick(func() {
								title := todos[i].Title
								todos = append(todos[:i], todos[i+1:]...)
								addLog("删除任务：" + title)
								app.Rebuild()
							}),
						).Gap(8).Items(ui.AlignCenter).PaddingXY(0, 7).
							BorderBottom(1, ui.BorderSubtle)
					}),
				).Gap(0).PaddingTRBL(4, 10, 8, 18),
			).Grow(),
		).Grow().Bg(ui.CardBg)

		// ---- 右栏 ----
		rightPanel := ui.Column(
			ui.Row(
				ui.Text("活动日志").FontSize(17).Bold().Grow(),
				ui.Text(fmt.Sprintf("%d 条", len(logs))).FontSize(12).Color(ui.Fg3),
				ui.Button("清空").OnClick(func() {
					logs = logs[:0]
					app.Rebuild()
				}),
			).Gap(8).Items(ui.AlignCenter).PaddingTRBL(18, 18, 8, 18),

			ui.Row(
				ui.Button("添加示例日志").OnClick(func() {
					addLog(fmt.Sprintf("手动记录 #%d", nextID))
					nextID++
					app.Rebuild()
				}),
				ui.Spacer(),
				ui.Text("每一栏独立滚动").FontSize(12).Color(ui.Fg3),
			).Items(ui.AlignCenter).PaddingTRBL(0, 18, 8, 18),

			ui.Scroll(
				ui.Column(
					ui.Each(logs, func(l logEntry, i int) *ui.Node {
						return ui.Row(
							ui.Text(l.Time).Width(62).FontSize(12).Color(ui.Fg3),
							ui.Text(l.Text).Grow().FontSize(13),
						).Gap(8).Items(ui.AlignCenter).PaddingXY(0, 6).
							BorderBottom(1, ui.BorderSubtle)
					}),
				).Gap(0).PaddingTRBL(4, 10, 8, 18),
			).Grow(),
		).Grow().Bg(ui.CardBg)

		return ui.Column(
			// 顶栏：自己带 padding，让两栏和滚动条可以贴到窗口左右边缘。
			ui.Row(
				ui.Column(
					ui.Text("CORE UI × GO").FontSize(11).Color(ui.Fg3),
					ui.Text("纯 Go 双栏列表").FontSize(24).Bold(),
				).Gap(2).Grow(),
				// color.RGBA 会由 ui 层自动转成 "#rrggbb" / "#rrggbbaa"。
				ui.Text("●").FontSize(12).Color(color.RGBA{R: 0x22, G: 0xc5, B: 0x5e, A: 0xff}),
				ui.Text("运行中").FontSize(12).Color(ui.Fg3),
				ui.Text(themeName).FontSize(12).Color(ui.Fg3),
				ui.Button("切换主题").OnClick(func() {
					dark = !dark
					if dark {
						app.SetTheme(coreui.Dark)
						themeName.Set("深色")
					} else {
						app.SetTheme(coreui.Light)
						themeName.Set("浅色")
					}
				}),
			).Items(ui.AlignCenter).Gap(12).PaddingTRBL(18, 24, 12, 24),

			// 两栏主体：Row 撑满剩余高度，中间 1px 分隔线。
			ui.Row(
				leftPanel,
				ui.Stack().Width(1).Bg(ui.BorderColor),
				rightPanel,
			).Grow(),

			// 底栏
			ui.Row(
				ui.Text("core-ui ", coreui.Version()).FontSize(11).Color(ui.Fg3),
				ui.Spacer(),
				ui.Button("退出").OnClick(func() { app.Quit() }),
			).Items(ui.AlignCenter).PaddingTRBL(10, 24, 10, 24),
		)
	})

	os.Exit(app.Run())
}
