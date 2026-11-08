// Command simple 是最小可运行的纯 Go 窗口示例。
//
// 它不使用 .uix，也不使用 JavaScript：窗口内容全部由 coreui/ui 的
// Go 声明式节点构建。运行前请把 core-ui.dll 放到 exe 同目录。
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
