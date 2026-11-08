// Command hello 演示用 Go + .uix 构建一个 core-ui 桌面应用。
package main

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/YuFu-Onship/core-ui-go-binding/coreui"
)

//go:embed app.uix
var appUIX string

func main() {
	if err := coreui.InitWithTheme(coreui.Light); err != nil {
		fail(err)
	}

	page, err := coreui.LoadPageString(appUIX)
	if err != nil {
		coreui.Shutdown()
		fail(err)
	}

	page.SetText("nativeMsg", "由 Go 写入 · core-ui "+coreui.Version())

	win, err := page.OpenWindow(&coreui.WindowConfig{
		Title:     "Core UI + Go",
		Resizable: true,
	})
	if err != nil {
		page.Destroy()
		coreui.Shutdown()
		fail(err)
	}

	hits := 0
	if btn := page.Root().FindByID("btnNative"); btn.Handle() != 0 {
		btn.OnClick(func() {
			hits++
			page.SetInt("nativeHits", hits)
			win.Invalidate()
		})
	}

	code := coreui.Run()
	page.Destroy()
	coreui.Shutdown()
	os.Exit(code)
}

func fail(err error) {
	fmt.Fprintf(os.Stderr, "%v\n", err)
	os.Exit(1)
}
