package ui

import (
	"image"
	_ "image/png"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/YuFu-Onship/core-ui-go-binding/coreui"
)

// 这组测试会真的加载 core-ui.dll。默认跳过，运行前设置 COREUI_E2E=1，
// 并把 third_party/coreui/bin 放进 PATH。
func e2e(t *testing.T) {
	t.Helper()
	if os.Getenv("COREUI_E2E") == "" {
		t.Skip("设置 COREUI_E2E=1 并让 core-ui.dll 可被加载后运行")
	}
}

func TestStateTwoWayBinding(t *testing.T) {
	e2e(t)
	if err := coreui.InitWithTheme(coreui.Light); err != nil {
		t.Fatalf("init: %v", err)
	}
	defer coreui.Shutdown()

	app := NewApp(Options{Title: "e2e", Theme: coreui.Light})
	count := app.Int(1)
	name := app.Text("a")
	flag := app.Bool(false)
	ratio := app.Float(0.5)

	app.View(func() *Node {
		return Column(Text(count), Text(name), Check(flag), Slider(ratio, 0, 100))
	})

	page, err := app.render()
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	defer page.Destroy()

	// 模拟控件把新值写进页面响应式数据：Go 状态应通过 watch 同步。
	page.SetInt(count.stateName(), 5)
	if got := count.Get(); got != 5 {
		t.Fatalf("IntState 没有收到页面回写: %d", got)
	}
	page.SetText(name.stateName(), "hello")
	if got := name.Get(); got != "hello" {
		t.Fatalf("TextState 没有收到页面回写: %q", got)
	}
	page.SetBool(flag.stateName(), true)
	if !flag.Get() {
		t.Fatal("BoolState 没有收到页面回写")
	}
	page.SetFloat(ratio.stateName(), 42.5)
	if got := ratio.Get(); got != 42.5 {
		t.Fatalf("FloatState 没有收到页面回写: %v", got)
	}

	// 反向：Go setter 应写回页面变量。
	count.Set(9)
	if got, _ := page.GetInt(count.stateName()); got != 9 {
		t.Fatalf("setter 没有写回页面: %d", got)
	}
	name.Set("world")
	if got, _ := page.GetText(name.stateName()); got != "world" {
		t.Fatalf("setter 没有写回页面: %q", got)
	}
	flag.Set(false)
	if got, _ := page.GetBool(flag.stateName()); got {
		t.Fatal("setter 没有写回页面布尔值")
	}
	ratio.Set(3.25)
	if got, _ := page.GetFloat(ratio.stateName()); got != 3.25 {
		t.Fatalf("setter 没有写回页面浮点值: %v", got)
	}
}

func TestReRenderKeepsStateValues(t *testing.T) {
	e2e(t)
	if err := coreui.InitWithTheme(coreui.Light); err != nil {
		t.Fatalf("init: %v", err)
	}
	defer coreui.Shutdown()

	app := NewApp(Options{Title: "e2e", Theme: coreui.Light})
	count := app.Int(3)
	app.View(func() *Node { return Column(Text(count)) })

	first, err := app.render()
	if err != nil {
		t.Fatalf("first render: %v", err)
	}
	count.Set(7)

	second, err := app.render()
	if err != nil {
		t.Fatalf("second render: %v", err)
	}
	defer second.Destroy()
	if got, _ := second.GetInt(count.stateName()); got != 7 {
		t.Fatalf("重建后新页面没有拿到当前状态: %d", got)
	}
	_ = first
}

func TestInputMountKeepsState(t *testing.T) {
	e2e(t)
	if err := coreui.InitWithTheme(coreui.Light); err != nil {
		t.Fatalf("init: %v", err)
	}
	defer coreui.Shutdown()

	app := NewApp(Options{Title: "mount", Width: 520, Height: 160, Theme: coreui.Light})
	route := app.Int(0)
	search := app.Text("core")
	app.View(func() *Node {
		return Column(
			WhenInt(route, 0, Text("home")),
			WhenInt(route, 1, Input(search).ID("search")),
		).Padding(12)
	})
	if err := app.start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	defer app.shutdown()

	win := app.Window()
	route.Set(1)
	coreui.DebugPump()
	time.Sleep(150 * time.Millisecond)
	coreui.DebugPump()
	if input := app.Page().Root().FindByID("search"); input.Handle() == 0 {
		t.Fatal("v-if 没有挂载输入框")
	}
	if got := search.Get(); got != "core" {
		t.Fatalf("挂载输入框后状态被覆盖: %q", got)
	}
	win.Destroy()
}

func TestImeContextFollowsFocus(t *testing.T) {
	e2e(t)
	if err := coreui.InitWithTheme(coreui.Light); err != nil {
		t.Fatalf("init: %v", err)
	}
	defer coreui.Shutdown()

	app := NewApp(Options{Title: "ime-focus", Width: 480, Height: 160, Theme: coreui.Light})
	text := app.Text("x")
	app.View(func() *Node {
		return Column(
			Input(text).ID("input"),
			Button("button").ID("button"),
		).Padding(12)
	})
	if err := app.start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	defer app.shutdown()

	win := app.Window()
	root := app.Page().Root()
	input := root.FindByID("input")
	button := root.FindByID("button")
	if input.Handle() == 0 || button.Handle() == 0 {
		t.Fatal("找不到测试控件")
	}
	if win.DebugImeEnabled() {
		t.Fatal("没有焦点时 IME context 不应关联")
	}
	if !win.DebugFocus(input) || !win.DebugImeEnabled() {
		t.Fatal("文本输入框获得焦点后 IME context 应关联")
	}
	if !win.DebugBlur() || win.DebugImeEnabled() {
		t.Fatal("点击空白/失焦后 IME context 应解除关联")
	}
	if !win.DebugFocus(button) || win.DebugImeEnabled() {
		t.Fatal("焦点在普通按钮时 IME context 不应关联")
	}
	if !win.DebugFocus(input) || !win.DebugImeEnabled() {
		t.Fatal("再次聚焦文本输入框后 IME context 应恢复")
	}
	win.Destroy()
}

func TestNestedScrollDoesNotGrowOuter(t *testing.T) {
	e2e(t)
	if err := coreui.InitWithTheme(coreui.Light); err != nil {
		t.Fatalf("init: %v", err)
	}
	defer coreui.Shutdown()

	app := NewApp(Options{Title: "nested-scroll", Width: 420, Height: 360, Theme: coreui.Light})
	items := make([]int, 100)
	app.View(func() *Node {
		return Column(
			Scroll(
				Column(
					Scroll(Column(
						Each(items, func(_ int, _ int) *Node { return Text("row") }),
					)).ID("inner").Height(120),
					Text("footer"),
				).Gap(8),
			).ID("outer").Height(200),
		).Padding(12).Grow()
	})
	if err := app.start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	defer app.shutdown()

	root := app.Page().Root()
	inner := root.FindByID("inner")
	outer := root.FindByID("outer")
	ic, iv, _, ok := inner.DebugScrollInfo()
	if !ok || ic <= iv {
		t.Fatalf("inner should overflow: content=%v visible=%v ok=%v", ic, iv, ok)
	}
	oc, ov, _, ok := outer.DebugScrollInfo()
	if !ok {
		t.Fatal("outer scroll info unavailable")
	}
	// 外层内容 = inner 的固定高度 + footer，应该仍能放进 200px 视口。
	// 旧实现会递归量到 inner 的全部子项，导致外层也出现滚动条。
	if oc > ov+1 {
		t.Fatalf("outer content should not include inner overflow: content=%v visible=%v", oc, ov)
	}
	app.Window().Destroy()
}

// Regression: the ComboBox v-model callback used to run inside the window's
// ForEachWidget traversal. Selecting an option that mounts/unmounts v-if
// siblings invalidated the walk; clicking a later option then an earlier one
// crashed with 0xc0000005.
func TestSelectDropdownRebuild(t *testing.T) {
	e2e(t)
	if err := coreui.InitWithTheme(coreui.Light); err != nil {
		t.Fatalf("init: %v", err)
	}
	defer coreui.Shutdown()

	app := NewApp(Options{Title: "select", Width: 420, Height: 300, Theme: coreui.Light})
	choice := app.Int(0)
	app.View(func() *Node {
		return Column(
			Select(choice, []string{"A", "B", "C", "D"}).ID("sel").Width(220),
			WhenInt(choice, 0, Text("card A")),
			WhenInt(choice, 1, Text("card B")),
			WhenInt(choice, 2, Text("card C")),
			WhenInt(choice, 3, Text("card D")),
		).Padding(12).Grow()
	})
	if err := app.start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	defer app.shutdown()

	win := app.Window()
	sel := app.Page().Root().FindByID("sel")
	if sel.Handle() == 0 {
		t.Fatal("select not found")
	}
	r := sel.Rect()
	x := float32((r.Left + r.Right) / 2)
	// ComboBox dropdown: 2px gap below the control, 28px rows.
	row := func(i int) float32 { return float32(r.Bottom + 2 + float64(i*28+14)) }

	if !win.DebugClick(sel) { // open
		t.Fatal("open dropdown failed")
	}
	coreui.DebugPump()
	time.Sleep(30 * time.Millisecond)
	if !win.DebugClickAt(x, row(3)) { // later option
		t.Fatal("click item 3 failed")
	}
	coreui.DebugPump()
	time.Sleep(30 * time.Millisecond)
	if got := choice.Get(); got != 3 {
		t.Fatalf("choice = %d, want 3", got)
	}

	if !win.DebugClick(sel) { // reopen
		t.Fatal("reopen dropdown failed")
	}
	coreui.DebugPump()
	time.Sleep(30 * time.Millisecond)
	if !win.DebugClickAt(x, row(0)) { // earlier option
		t.Fatal("click item 0 failed")
	}
	coreui.DebugPump()
	time.Sleep(30 * time.Millisecond)
	if got := choice.Get(); got != 0 {
		t.Fatalf("choice = %d, want 0", got)
	}
	win.Destroy()
}

func TestThemeSwitchUpdatesBackground(t *testing.T) {
	e2e(t)
	if err := coreui.InitWithTheme(coreui.Light); err != nil {
		t.Fatalf("init: %v", err)
	}
	defer coreui.Shutdown()

	app := NewApp(Options{Title: "theme", Width: 420, Height: 260, Theme: coreui.Light})
	app.View(func() *Node {
		return Column(
			Button("toggle").ID("toggle").OnClick(func() {
				app.SetTheme(coreui.Dark)
			}),
			Text("background"),
		).Padding(12).Grow()
	})
	if err := app.start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	defer app.shutdown()

	win := app.Window()
	lightPath := filepath.Join(t.TempDir(), "light.png")
	if !win.DebugScreenshot(lightPath) {
		t.Fatal("light screenshot failed")
	}
	light := pngPixel(t, lightPath)

	toggle := app.Page().Root().FindByID("toggle")
	if toggle.Handle() == 0 || !win.DebugClick(toggle) {
		t.Fatal("theme toggle click failed")
	}
	coreui.DebugPump()
	time.Sleep(150 * time.Millisecond)
	coreui.DebugPump()

	darkPath := filepath.Join(t.TempDir(), "dark.png")
	if !win.DebugScreenshot(darkPath) {
		t.Fatal("dark screenshot failed")
	}
	dark := pngPixel(t, darkPath)

	lum := func(c [3]uint8) int { return (int(c[0]) + int(c[1]) + int(c[2])) / 3 }
	if lum(light) < 200 {
		t.Fatalf("light background should be bright, got %v", light)
	}
	if lum(dark) > 100 {
		t.Fatalf("dark background did not change: light=%v dark=%v", light, dark)
	}
	win.Destroy()
}

func pngPixel(t *testing.T, path string) [3]uint8 {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open screenshot: %v", err)
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		t.Fatalf("decode screenshot: %v", err)
	}
	b := img.Bounds()
	// Bottom-right area: inside the page root background (outside the 12px
	// padding), away from the button/text at the top.
	x, y := b.Max.X-20, b.Max.Y-20
	r, g, bl, _ := img.At(x, y).RGBA()
	return [3]uint8{uint8(r >> 8), uint8(g >> 8), uint8(bl >> 8)}
}

func TestSubmitOnEnter(t *testing.T) {
	e2e(t)
	if err := coreui.InitWithTheme(coreui.Light); err != nil {
		t.Fatalf("init: %v", err)
	}
	defer coreui.Shutdown()

	app := NewApp(Options{Title: "submit", Width: 520, Height: 160, Theme: coreui.Light})
	search := app.Text("go")
	info := app.Text("")
	app.View(func() *Node {
		return Column(
			Input(search).ID("search").OnSubmit(func() {
				info.Set("submit:" + search.Get())
			}),
		).Padding(12)
	})
	if err := app.start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	defer app.shutdown()

	win := app.Window()
	input := app.Page().Root().FindByID("search")
	if input.Handle() == 0 || !win.DebugFocus(input) {
		t.Fatal("无法聚焦搜索输入框")
	}
	if !win.DebugKey(0x0D) { // VK_RETURN
		t.Fatal("DebugKey(VK_RETURN) 失败")
	}
	if got := info.Get(); got != "submit:go" {
		t.Fatalf("Enter 没有触发 OnSubmit: %q", got)
	}
	win.Destroy()
}

func TestRapidRebuild(t *testing.T) {
	e2e(t)
	if err := coreui.InitWithTheme(coreui.Light); err != nil {
		t.Fatalf("init: %v", err)
	}
	defer coreui.Shutdown()

	app := NewApp(Options{Title: "rebuild", Width: 640, Height: 400, Theme: coreui.Light})
	draft := app.Text("")
	app.View(func() *Node {
		return Column(
			Input(draft).ID("input"),
			Button("delete").OnClick(func() {}),
		).Padding(12).Gap(8)
	})
	if err := app.start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	defer app.shutdown()

	win := app.Window()
	// 先让窗口持有焦点/命中指针，再连续换根；SetRoot 必须清理旧树指针，
	// 否则下一次鼠标/键盘消息会访问已释放的 widget。
	win.DebugFocus(app.Page().Root().FindByID("input"))
	for i := 0; i < 40; i++ {
		if err := app.rebuildNow(); err != nil {
			t.Fatalf("rebuild %d: %v", i, err)
		}
	}
	win.Destroy()
}

func TestMaximizeRelayout(t *testing.T) {
	e2e(t)
	if err := coreui.InitWithTheme(coreui.Light); err != nil {
		t.Fatalf("init: %v", err)
	}
	defer coreui.Shutdown()

	app := NewApp(Options{Title: "max", Width: 640, Height: 400, Theme: coreui.Light})
	app.View(func() *Node {
		return Column(Text("maximize layout probe")).Padding(20)
	})
	if err := app.start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	defer app.shutdown()

	win := app.Window()
	root := app.Page().Root()
	before := root.Rect()
	win.Maximize()
	coreui.DebugPump()
	after := root.Rect()
	t.Logf("before=%+v after=%+v maximized=%v", before, after, win.IsMaximized())
	if after.Width() <= before.Width()+20 || after.Height() <= before.Height()+20 {
		t.Fatalf("最大化后根控件布局没有更新: before=%+v after=%+v", before, after)
	}
	win.Restore()
	coreui.DebugPump()
	restored := root.Rect()
	t.Logf("restored=%+v", restored)
	if restored.Width() >= after.Width()-20 {
		t.Fatalf("还原后布局没有恢复: restored=%+v after=%+v", restored, after)
	}

	win.Minimize()
	coreui.DebugPump()
	win.Restore()
	coreui.DebugPump()
	fromMin := root.Rect()
	t.Logf("from-minimize=%+v", fromMin)
	if fromMin.Width() != before.Width() || fromMin.Height() != before.Height() {
		t.Fatalf("最小化/还原后布局不对: before=%+v fromMin=%+v", before, fromMin)
	}
	win.Destroy()
}

func TestScrollViewStartsWithFirstItem(t *testing.T) {
	e2e(t)
	if err := coreui.InitWithTheme(coreui.Light); err != nil {
		t.Fatalf("init: %v", err)
	}
	defer coreui.Shutdown()

	app := NewApp(Options{Title: "scroll", Width: 600, Height: 300, Theme: coreui.Light})
	app.View(func() *Node {
		items := make([]*Node, 0, 20)
		for i := 1; i <= 20; i++ {
			items = append(items, Text("row ").ID("l"+strconv.Itoa(i)).PaddingXY(0, 6))
		}
		return Column(
			Scroll(Column(items).ID("content")).ID("sv").Grow(),
		).Grow()
	})
	if err := app.start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	defer app.shutdown()

	win := app.Window()
	root := app.Page().Root()
	sv := root.FindByID("sv")
	content := root.FindByID("content")
	first := root.FindByID("l1")
	if sv.Handle() == 0 || content.Handle() == 0 || first.Handle() == 0 {
		t.Fatal("找不到 ScrollView / content / 第一行")
	}
	sr, cr, fr := sv.Rect(), content.Rect(), first.Rect()
	t.Logf("scroll=%+v content=%+v first=%+v", sr, cr, fr)
	if fr.Top < sr.Top-1 {
		t.Fatalf("第一行被滚出视口: scroll=%+v first=%+v", sr, fr)
	}
	if cr.Right > sr.Right-8 {
		t.Fatalf("内容没有给滚动条留出 gutter: scroll=%+v content=%+v", sr, cr)
	}
	win.Destroy()
}

func TestInlineImeComposition(t *testing.T) {
	e2e(t)
	if err := coreui.InitWithTheme(coreui.Light); err != nil {
		t.Fatalf("init: %v", err)
	}
	defer coreui.Shutdown()

	app := NewApp(Options{Title: "ime", Width: 480, Height: 160, Theme: coreui.Light})
	name := app.Text("")
	app.View(func() *Node {
		return Column(Input(name).ID("imeInput"))
	})
	if err := app.start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	defer app.shutdown()

	win := app.Window()
	input := app.Page().Root().FindByID("imeInput")
	if input.Handle() == 0 {
		t.Fatal("找不到输入框")
	}
	if !win.DebugFocus(input) {
		t.Fatal("DebugFocus 失败")
	}
	if !win.DebugImeComposition("ni'hao", 6) {
		t.Fatal("DebugImeComposition 失败")
	}
	if got := name.Get(); got != "" {
		t.Fatalf("预编辑阶段不应回写状态: %q", got)
	}
	if !win.DebugImeCommit("你好") {
		t.Fatal("DebugImeCommit 失败")
	}
	if got := name.Get(); got != "你好" {
		t.Fatalf("提交后状态不对: %q", got)
	}
	win.Destroy()
}

func TestEventBusDispatchesToGo(t *testing.T) {
	e2e(t)
	if err := coreui.InitWithTheme(coreui.Light); err != nil {
		t.Fatalf("init: %v", err)
	}
	defer coreui.Shutdown()

	app := NewApp(Options{Title: "e2e", Theme: coreui.Light})
	clicks := 0
	app.View(func() *Node {
		return Column(Button("click").OnClick(func() { clicks++ }))
	})
	page, err := app.render()
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	defer page.Destroy()

	// 先确认事件表本身没问题。
	app.onEventJSON(`{"seq":1,"id":1,"args":[]}`)
	if clicks != 1 {
		t.Fatalf("onEventJSON 没有调到处理函数: clicks=%d", clicks)
	}

	// 再确认“页面 -> Go”的 watcher 链路。生成器按文档顺序分配事件 id，
	// 这里第一个按钮是 1。
	if err := page.SetJSON("__evt", `{"seq":2,"id":1,"args":[]}`); err != nil {
		t.Fatalf("SetJSON: %v", err)
	}
	if clicks != 2 {
		t.Fatalf("事件没有通过页面 watcher 分发到 Go: clicks=%d", clicks)
	}
	// 相同 seq 应被忽略。
	if err := page.SetJSON("__evt", `{"seq":2,"id":1,"args":[]}`); err != nil {
		t.Fatalf("SetJSON: %v", err)
	}
	if clicks != 2 {
		t.Fatalf("重复事件不应该再次触发: clicks=%d", clicks)
	}
}
