package coreui

/*
#include <ui_core.h>
#include <stdlib.h>
#include "callbacks.h"
*/
import "C"

import (
	"errors"
	"unsafe"
)

// WindowConfig 对应 UiWindowConfig。零值即“全部用 `.uix` <window> 标签里的值”。
//
// 单位：Width / Height 是 DIP（1/96 英寸，等同 CSS px），X / Y 是屏幕物理像素。
type WindowConfig struct {
	Title          string
	Width          int
	Height         int
	SystemFrame    bool // true = 系统标题栏，false = 无边框自绘标题栏
	Resizable      bool
	AcceptFiles    bool
	X              int
	Y              int
	ToolWindow     bool
	SkipAnimation  bool
	StartMaximized bool
	LayeredCanvas  bool
	// CopySwapChain 使用 DXGI bitblt 交换链（Gio 风格的 DISCARD / 1 buffer /
	// STRETCH），而不是默认的 flip 模型。用于兼容拖动/缩放时对 DWM 合成更
	// 敏感的窗口。
	CopySwapChain bool
	Owner         Window
}

func (c *WindowConfig) toC() (*C.UiWindowConfig, func()) {
	if c == nil {
		return nil, func() {}
	}
	var cc C.UiWindowConfig
	cc.width = C.int(c.Width)
	cc.height = C.int(c.Height)
	cc.resizable = boolToC(c.Resizable)
	cc.system_frame = boolToC(c.SystemFrame)
	cc.accept_files = boolToC(c.AcceptFiles)
	cc.x = C.int(c.X)
	cc.y = C.int(c.Y)
	cc.tool_window = boolToC(c.ToolWindow)
	cc.skip_animation = boolToC(c.SkipAnimation)
	cc.start_maximized = boolToC(c.StartMaximized)
	cc.layered_canvas = boolToC(c.LayeredCanvas)
	cc.copy_swap_chain = boolToC(c.CopySwapChain)
	cc.owner = c.Owner.h

	title := (*C.wchar_t)(nil)
	if c.Title != "" {
		title = cWide(c.Title)
		cc.title = title
	}
	return &cc, func() { freeWide(title) }
}

// Window 是一个 core-ui 窗口句柄。
type Window struct {
	h C.UiWindow
}

// Handle 返回底层句柄数值，0 表示无效。
func (w Window) Handle() uint64 {
	return uint64(w.h)
}

// Show 显示窗口（带开场动画）。
func (w Window) Show() {
	if w.h != 0 {
		C.ui_window_show(w.h)
	}
}

// ShowImmediate 立即显示，跳过开场动画。
func (w Window) ShowImmediate() {
	if w.h != 0 {
		C.ui_window_show_immediate(w.h)
	}
}

// ShowImmediateNoActivate 显示窗口但不抢键盘焦点。
func (w Window) ShowImmediateNoActivate() {
	if w.h != 0 {
		C.ui_window_show_immediate_no_activate(w.h)
	}
}

// Hide 隐藏窗口。
func (w Window) Hide() {
	if w.h != 0 {
		C.ui_window_hide(w.h)
	}
}

// Invalidate 请求重绘。
func (w Window) Invalidate() {
	if w.h != 0 {
		C.ui_window_invalidate(w.h)
	}
}

// Relayout 强制重新布局（可见性等影响布局的属性变化后调用）。
func (w Window) Relayout() {
	if w.h != 0 {
		C.ui_window_relayout(w.h)
	}
}

// SetTitle 修改窗口标题。
func (w Window) SetTitle(title string) {
	if w.h == 0 {
		return
	}
	t := cWide(title)
	defer freeWide(t)
	C.ui_window_set_title(w.h, t)
}

// SetResizable 修改可调整大小标志。
func (w Window) SetResizable(v bool) {
	if w.h != 0 {
		C.ui_window_set_resizable(w.h, boolToC(v))
	}
}

// SetMinSize 设置最小尺寸，单位 DIP。
func (w Window) SetMinSize(widthDip, heightDip int) {
	if w.h != 0 {
		C.ui_window_set_min_size(w.h, C.int(widthDip), C.int(heightDip))
	}
}

// Maximize 最大化窗口（走系统标题栏的同一路径）。
func (w Window) Maximize() {
	if w.h != 0 {
		C.ui_window_maximize(w.h)
	}
}

// Minimize 最小化窗口。
func (w Window) Minimize() {
	if w.h != 0 {
		C.ui_window_minimize(w.h)
	}
}

// Restore 从最大化/最小化恢复窗口。
func (w Window) Restore() {
	if w.h != 0 {
		C.ui_window_restore(w.h)
	}
}

// IsMaximized 返回窗口当前是否处于最大化状态。
func (w Window) IsMaximized() bool {
	return w.h != 0 && C.ui_window_is_maximized(w.h) != 0
}

// SetRoot 替换窗口根控件。
func (w Window) SetRoot(root Widget) {
	if w.h != 0 {
		C.ui_window_set_root(w.h, root.h)
	}
}

// Destroy 销毁窗口。
func (w Window) Destroy() {
	if w.h != 0 {
		C.ui_window_destroy(w.h)
	}
}

// DPIScale 返回当前 DPI 缩放比：100% → 1.0，150% → 1.5。
func (w Window) DPIScale() float64 {
	if w.h == 0 {
		return 1
	}
	return float64(C.ui_window_dpi_scale(w.h))
}

// DPI 返回当前显示器 DPI（96 的倍数）。
func (w Window) DPI() int {
	if w.h == 0 {
		return 96
	}
	return int(C.ui_window_dpi(w.h))
}

// HWND 返回底层 Win32 窗口句柄（unsafe.Pointer 形式）。
func (w Window) HWND() unsafe.Pointer {
	if w.h == 0 {
		return nil
	}
	return C.ui_window_hwnd(w.h)
}

// Post 把 fn 投递到窗口的 UI 线程执行，可以从任意 goroutine 调用。
//
// 如果当前已经在 UI 线程上，fn 不会在本次调用内重入，而是等当前消息处理完
// 再执行 —— 这一点对“在事件回调里重建页面”很关键。
func (w Window) Post(fn func()) {
	if w.h == 0 || fn == nil {
		return
	}
	id := registerCallback(invokeCallbacks, fn)
	ud := callbackPtr(id)
	go func() {
		defer unregisterCallback(invokeCallbacks, id)
		defer C.free(ud)
		C.coreui_invoke_sync(w.h, ud)
	}()
}

// OnClose 注册关闭前通知回调，回调不能阻止关闭。
func (w Window) OnClose(fn func()) error {
	if w.h == 0 {
		return errors.New("coreui: 窗口无效")
	}
	id := registerCallback(closeCallbacks, fn)
	C.ui_window_on_close(w.h, (C.UiWindowCloseCallback)(C.goCoreUiWindowClose), callbackPtr(id))
	return nil
}

// OnCloseRequest 注册关闭请求回调：返回 true 继续默认关闭，返回 false 取消。
func (w Window) OnCloseRequest(fn func() bool) error {
	if w.h == 0 {
		return errors.New("coreui: 窗口无效")
	}
	id := registerCallback(closeRequestCallbacks, fn)
	C.ui_window_on_close_request(w.h, (C.UiWindowCloseRequestCallback)(C.goCoreUiWindowCloseRequest), callbackPtr(id))
	return nil
}

// Rect 是控件在窗口内的布局矩形，单位 DIP。
type Rect struct {
	Left   float64
	Top    float64
	Right  float64
	Bottom float64
}

// Width 返回矩形宽度。
func (r Rect) Width() float64 { return r.Right - r.Left }

// Height 返回矩形高度。
func (r Rect) Height() float64 { return r.Bottom - r.Top }

// Widget 是一个 core-ui 控件句柄。
type Widget struct {
	h C.UiWidget
}

// Handle 返回底层句柄数值，0 表示无效。
func (w Widget) Handle() uint64 {
	return uint64(w.h)
}

// Rect 返回控件当前的布局矩形（DIP）。
func (w Widget) Rect() Rect {
	if w.h == 0 {
		return Rect{}
	}
	r := C.ui_widget_get_rect(w.h)
	return Rect{
		Left:   float64(r.left),
		Top:    float64(r.top),
		Right:  float64(r.right),
		Bottom: float64(r.bottom),
	}
}

// FindByID 按 `.uix` 里的 id 属性查找子控件，找不到时返回零值 Widget。
func (w Widget) FindByID(id string) Widget {
	if w.h == 0 {
		return Widget{}
	}
	cs := C.CString(id)
	defer C.free(unsafe.Pointer(cs))
	return Widget{h: C.ui_widget_find_by_id(w.h, cs)}
}

// AddChild 把一个控件挂到当前控件下。
func (w Widget) AddChild(child Widget) {
	if w.h != 0 && child.h != 0 {
		C.ui_widget_add_child(w.h, child.h)
	}
}

// SetText 设置标签/按钮等控件的文本。
func (w Widget) SetText(text string) {
	if w.h == 0 {
		return
	}
	t := cWide(text)
	defer freeWide(t)
	C.ui_label_set_text(w.h, t)
}

// SetVisible 设置可见性；影响布局时随后调用窗口的 Relayout。
func (w Widget) SetVisible(v bool) {
	if w.h != 0 {
		C.ui_widget_set_visible(w.h, boolToC(v))
	}
}

// SetEnabled 设置启用状态。
func (w Widget) SetEnabled(v bool) {
	if w.h != 0 {
		C.ui_widget_set_enabled(w.h, boolToC(v))
	}
}

// OnClick 注册鼠标点击回调（原生控件，与 `.uix` 里的 @click 相互独立）。
func (w Widget) OnClick(fn func()) error {
	if w.h == 0 {
		return errors.New("coreui: 控件无效")
	}
	id := registerCallback(clickCallbacks, fn)
	C.ui_widget_on_click(w.h, (C.UiClickCallback)(C.goCoreUiClick), callbackPtr(id))
	return nil
}

// Menu 是一个菜单句柄，归页面所有。
type Menu struct {
	h C.UiMenu
}

// Handle 返回底层句柄数值，0 表示无效。
func (m Menu) Handle() uint64 {
	return uint64(m.h)
}

// Show 在窗口坐标 (x, y) 处弹出菜单。
func (m Menu) Show(w Window, x, y float32) {
	if m.h == 0 || w.h == 0 {
		return
	}
	C.ui_menu_show(w.h, m.h, C.float(x), C.float(y))
}

func boolToC(v bool) C.int {
	if v {
		return 1
	}
	return 0
}
