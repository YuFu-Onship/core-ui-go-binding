package ui

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/YuFu-Onship/core-ui-go-binding/coreui"
)

// Options 控制窗口与全局主题。
type Options struct {
	Title     string
	Width     int
	Height    int
	MinWidth  int
	MinHeight int

	// FixedSize 为 true 时窗口不可调整大小（默认允许调整）。
	FixedSize bool
	Centered  bool
	Frameless bool

	Theme      coreui.Theme
	Background string // 根容器背景，默认 var(--bg)
	CSS        string // 额外注入的 <style> 内容

	FontFamily string // 例如 "Microsoft YaHei UI"
	Accent     string // 例如 "#2563eb"
	Locale     string // 例如 "zh-CN"

	// AllowPageWarnings 为 true 时忽略 ui_page_last_error 里的非空内容。
	// 默认把非空内容当编译错误，便于尽早发现生成器/模板问题。
	AllowPageWarnings bool

	// KeepConsole 为 true 时保留自动创建的控制台窗口。
	// 默认隐藏：双击控制台子系统 exe 时 Windows 会额外打开一个空的
	// 控制台窗口（常被误认为“经典 Aero 空白窗口”）。从已有终端启动时，
	// 控制台属于该终端，不会被隐藏。
	KeepConsole bool

	// FlipSwapChain 改回 core-ui 默认的 flip 交换链。默认（false）使用
	// Gio 风格的 bitblt（DXGI DISCARD / 1 buffer / STRETCH），对窗口拖动/
	// 缩放时的 DWM 合成更稳。高刷新/图像画布场景可显式打开 flip。
	FlipSwapChain bool
}

// App 是一个纯 Go 声明式应用的运行时。
type App struct {
	opts Options

	mu     sync.Mutex
	states []state
	view   func() *Node

	page *coreui.Page
	win  coreui.Window

	cancels  []func()
	handlers map[int]handlerSpec
	nextID   int
	eventSeq int

	langs   map[string]string
	onPages []func(*coreui.Page)

	pendingRebuild bool
	started        bool
	lastErr        error
}

// state 是四种响应式状态共用的内部接口。
type state interface {
	stateName() string
	stateJSON() string
	bindExpr() string
}

// NewApp 创建应用；用 View 注册视图，用 Run 启动消息循环。
func NewApp(opts Options) *App {
	applyPalette(opts.Theme != coreui.Light, opts.Accent)
	return &App{
		opts:     opts,
		handlers: map[int]handlerSpec{},
		langs:    map[string]string{},
	}
}

// Run 创建 App、调用 setup 注册状态与视图，然后启动消息循环。
func Run(opts Options, setup func(*App)) int {
	app := NewApp(opts)
	if setup != nil {
		setup(app)
	}
	return app.Run()
}

// OnPage 注册页面构建完成后的回调；每次首屏和 Rebuild 都会调用一次。
// 适合重新挂接翻译、资源或自定义控件。
func (a *App) OnPage(fn func(*coreui.Page)) *App {
	if fn != nil {
		a.onPages = append(a.onPages, fn)
	}
	return a
}

// LoadLanguageString 装载某个语言的 `.lang` 内容，并在之后每次 Rebuild
// 时自动重新应用（否则新页面会丢失翻译）。
func (a *App) LoadLanguageString(locale, content string) {
	a.mu.Lock()
	a.langs[locale] = content
	a.mu.Unlock()
	if a.page != nil {
		a.page.LoadLanguageString(locale, content)
	}
}

// LoadLanguageFile 从磁盘读取 `.lang` 文件并装载。
func (a *App) LoadLanguageFile(locale, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	a.LoadLanguageString(locale, string(data))
	return nil
}

// View 注册视图构建函数。该函数会被调用一次；调用 Rebuild 时再次调用。
func (a *App) View(fn func() *Node) *App {
	a.view = fn
	return a
}

// Run 启动应用，返回进程退出码。界面或编译失败时会把错误写到 stderr。
func (a *App) Run() int {
	if err := a.start(); err != nil {
		fmt.Fprintf(os.Stderr, "coreui/ui: %v\n", err)
		return 1
	}
	code := coreui.Run()
	a.shutdown()
	return code
}

// Error 返回启动或最近一次重建时的错误。
func (a *App) Error() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.lastErr
}

// Window 返回底层窗口句柄，可用于调用 coreui.Window 上的高级 API。
func (a *App) Window() coreui.Window { return a.win }

// Page 返回底层页面句柄，可用于调用 coreui.Page 上的高级 API。
func (a *App) Page() *coreui.Page { return a.page }

// Source 返回当前视图生成的 .uix 源码。用于调试生成结果，不会启动运行时。
func (a *App) Source() (string, error) {
	node, err := a.invokeView()
	if err != nil {
		return "", err
	}
	oldHandlers, oldNextID := a.handlers, a.nextID
	a.handlers = map[int]handlerSpec{}
	a.nextID = 0
	src := a.buildSource(node)
	a.handlers, a.nextID = oldHandlers, oldNextID
	return src, nil
}

// Quit 请求退出消息循环。
func (a *App) Quit() { coreui.Quit(0) }

// Post 把 fn 投递到 UI 线程；可以从任意 goroutine 调用。
func (a *App) Post(fn func()) {
	if fn == nil {
		return
	}
	if a.win.Handle() == 0 {
		fn()
		return
	}
	a.win.Post(fn)
}

// Rebuild 用当前的 Go 状态重新调用视图函数并替换页面根控件。
//
// 适合列表增删、页面结构切换等“布局本身变了”的场景。普通状态变化
// （count.Set / name.Set）会自动刷新，不需要 Rebuild。
//
// Rebuild 总是异步执行：在按钮回调里调用也不会在 QuickJS 求值过程中
// 销毁页面。连续多次调用会合并成一次。
func (a *App) Rebuild() {
	a.mu.Lock()
	if !a.started || a.pendingRebuild {
		a.mu.Unlock()
		return
	}
	a.pendingRebuild = true
	a.mu.Unlock()

	a.win.Post(func() {
		a.mu.Lock()
		a.pendingRebuild = false
		a.mu.Unlock()
		if err := a.rebuildNow(); err != nil {
			fmt.Fprintf(os.Stderr, "coreui/ui: 重建失败: %v\n", err)
			a.mu.Lock()
			a.lastErr = err
			a.mu.Unlock()
		}
	})
}

// SetTheme 在运行时切换主题。
func (a *App) SetTheme(t coreui.Theme) {
	a.opts.Theme = t
	coreui.SetTheme(t)
	applyPalette(t != coreui.Light, a.opts.Accent)
	// 纯 Go token 是具体的 color.RGBA，切换主题后需要重建一次视图。
	a.Rebuild()
}

// SetLocale 切换页面语言（配合 Page.LoadLanguageString / File 使用）。
func (a *App) SetLocale(locale string) {
	a.opts.Locale = locale
	if a.page != nil && locale != "" {
		a.page.SetLocale(locale)
	}
}

func (a *App) start() error {
	if a.view == nil {
		return fmt.Errorf("没有调用 View 注册视图")
	}
	if !a.opts.KeepConsole {
		hideOwnedConsole()
	}
	if err := coreui.InitWithTheme(a.opts.Theme); err != nil {
		return err
	}
	if a.opts.FontFamily != "" {
		coreui.SetDefaultFont(a.opts.FontFamily)
	}
	if a.opts.Accent != "" {
		if err := coreui.SetAccentHex(a.opts.Accent); err != nil {
			return err
		}
	}
	if a.opts.Locale != "" {
		coreui.SetTextLocale(a.opts.Locale)
	}

	page, err := a.render()
	if err != nil {
		coreui.Shutdown()
		return err
	}
	// 默认 bitblt；需要 flip 的调用方显式设置 FlipSwapChain。
	winCfg := &coreui.WindowConfig{CopySwapChain: !a.opts.FlipSwapChain}
	win, err := page.PrepareWindow(winCfg)
	if err != nil {
		page.Destroy()
		coreui.Shutdown()
		return err
	}
	a.page = page
	a.win = win
	a.started = true
	win.ShowImmediate()
	return nil
}

// render 构建源码、加载页面并绑定状态/事件。
func (a *App) render() (*coreui.Page, error) {
	applyPalette(a.opts.Theme != coreui.Light, a.opts.Accent)
	node, err := a.invokeView()
	if err != nil {
		return nil, err
	}

	// 每次渲染重新分配事件登记号；如果加载失败就恢复旧表，保证旧页面
	// 上的回调仍然有效。
	oldHandlers := a.handlers
	oldNextID := a.nextID
	a.handlers = map[int]handlerSpec{}
	a.nextID = 0

	src := a.buildSource(node)
	page, err := coreui.LoadPageString(src)
	if err != nil {
		a.handlers = oldHandlers
		a.nextID = oldNextID
		return nil, err
	}
	if msg := page.LastError(); msg != "" && !a.opts.AllowPageWarnings {
		page.Destroy()
		a.handlers = oldHandlers
		a.nextID = oldNextID
		return nil, fmt.Errorf("coreui 编译生成页面失败: %s", msg)
	}
	a.eventSeq = 0

	a.mu.Lock()
	langs := make(map[string]string, len(a.langs))
	for locale, content := range a.langs {
		langs[locale] = content
	}
	onPages := append([]func(*coreui.Page){}, a.onPages...)
	a.mu.Unlock()
	for locale, content := range langs {
		page.LoadLanguageString(locale, content)
	}
	if a.opts.Locale != "" {
		page.SetLocale(a.opts.Locale)
	}
	for _, fn := range onPages {
		fn(page)
	}

	oldCancels := a.cancels
	a.cancels = nil

	for _, s := range a.states {
		a.bindStatePage(s, page)
	}
	a.cancels = append(a.cancels, page.Watch("__evt", a.onEventJSON))

	if a.page != nil {
		// 热替换：窗口不重建，只换根控件。
		oldPage := a.page
		// SetRoot 已经 LayoutRoot；这里不要再 Relayout，否则每次 Rebuild
		// 会把整棵页面树完整布局两遍。
		a.win.SetRoot(page.Root())
		a.win.Invalidate()
		for _, cancel := range oldCancels {
			cancel()
		}
		oldPage.Destroy()
	}
	a.page = page
	return page, nil
}

func (a *App) rebuildNow() error {
	if !a.started {
		return nil
	}
	_, err := a.render()
	return err
}

func (a *App) invokeView() (n *Node, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("视图函数 panic: %v", r)
		}
	}()
	n = a.view()
	if n == nil {
		return nil, fmt.Errorf("视图函数返回了 nil")
	}
	return n, nil
}

func (a *App) addState(s state, setName func(string)) {
	a.mu.Lock()
	defer a.mu.Unlock()
	name := fmt.Sprintf("s%d", len(a.states)+1)
	setName(name)
	a.states = append(a.states, s)
}

func (a *App) bindStatePage(s state, p *coreui.Page) {
	switch x := s.(type) {
	case *IntState:
		x.stateBase.bindPage(p, x.syncJSON)
	case *FloatState:
		x.stateBase.bindPage(p, x.syncJSON)
	case *BoolState:
		x.stateBase.bindPage(p, x.syncJSON)
	case *TextState:
		x.stateBase.bindPage(p, x.syncJSON)
	}
}

func (a *App) addWatch(cancel func()) {
	a.cancels = append(a.cancels, cancel)
}

func (a *App) addHandler(h handlerSpec) int {
	a.nextID++
	a.handlers[a.nextID] = h
	return a.nextID
}

func (a *App) onEventJSON(raw string) {
	var ev struct {
		Seq  int               `json:"seq"`
		ID   int               `json:"id"`
		Args []json.RawMessage `json:"args"`
	}
	if err := json.Unmarshal([]byte(raw), &ev); err != nil {
		return
	}
	if ev.Seq <= a.eventSeq {
		return
	}
	a.eventSeq = ev.Seq
	h, ok := a.handlers[ev.ID]
	if !ok || h.fn == nil {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "coreui/ui: %s 事件回调 panic: %v\n", h.event, r)
		}
	}()
	h.fn(ev.Args)
}

func (a *App) shutdown() {
	for _, cancel := range a.cancels {
		cancel()
	}
	a.cancels = nil
	if a.page != nil {
		a.page.Destroy()
		a.page = nil
	}
	a.started = false
	coreui.Shutdown()
}
