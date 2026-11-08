package coreui

/*
#include <ui_core.h>
#include <stdlib.h>
*/
import "C"

// DebugImeComposition 直接给当前焦点文本控件喂一段 IME 预编辑文本，
// 用于自动化测试/调试内联输入法组合，不经过真实 IME 消息。
// caret 是预编辑文本内的 UTF-16 偏移。返回 false 表示当前焦点控件不支持 IME。
func (w Window) DebugImeComposition(text string, caret int) bool {
	if w.h == 0 {
		return false
	}
	t := cWide(text)
	defer freeWide(t)
	return C.ui_debug_ime_composition(w.h, t, C.int(caret)) == 0
}

// DebugImeCommit 直接把一段文本作为 IME 提交结果写入焦点控件。
func (w Window) DebugImeCommit(text string) bool {
	if w.h == 0 {
		return false
	}
	t := cWide(text)
	defer freeWide(t)
	return C.ui_debug_ime_commit(w.h, t) == 0
}

// DebugImeEnabled 返回窗口当前是否关联着 IME context。
// 焦点在普通控件/空白处时应为 false，焦点在文本控件时应为 true。
func (w Window) DebugImeEnabled() bool {
	if w.h == 0 {
		return false
	}
	return C.ui_debug_ime_enabled(w.h) == 1
}

// DebugImeEnd 结束一次未提交的 IME 组合（等价于取消预编辑）。
func (w Window) DebugImeEnd() bool {
	if w.h == 0 {
		return false
	}
	return C.ui_debug_ime_end(w.h) == 0
}

// DebugFocus 直接把焦点交给一个控件（自动化测试用）。
func (w Window) DebugFocus(widget Widget) bool {
	if w.h == 0 || widget.h == 0 {
		return false
	}
	return C.ui_debug_focus(w.h, widget.h) == 0
}

// DebugBlur 清空窗口内的控件焦点（自动化测试用）。
func (w Window) DebugBlur() bool {
	if w.h == 0 {
		return false
	}
	return C.ui_debug_blur(w.h) == 0
}

// DebugScrollInfo 读取 ScrollView 的布局状态：
// content height / visible height / current scrollY（均为 DIP）。
func (w Widget) DebugScrollInfo() (content, visible, scroll float64, ok bool) {
	if w.h == 0 {
		return 0, 0, 0, false
	}
	var c, v, s C.float
	if C.ui_debug_scroll_info(w.h, &c, &v, &s) != 0 {
		return 0, 0, 0, false
	}
	return float64(c), float64(v), float64(s), true
}

// DebugHandleCount 返回全局控件句柄表中的存活句柄数。
// 页面 Rebuild / 列表增删后若该值持续上涨，说明有控件树被句柄表钉住。
func DebugHandleCount() int {
	return int(C.ui_debug_handle_count())
}

// DebugClick 模拟点击一个控件（会自动换算控件中心点）。
func (w Window) DebugClick(widget Widget) bool {
	if w.h == 0 || widget.h == 0 {
		return false
	}
	return C.ui_debug_click(w.h, widget.h) == 0
}

// DebugClickAt 在窗口客户区 DIP 坐标 (x, y) 模拟一次点击。
func (w Window) DebugClickAt(x, y float32) bool {
	if w.h == 0 {
		return false
	}
	return C.ui_debug_click_at(w.h, C.float(x), C.float(y)) == 0
}

// DebugPump 处理当前线程队列里的 Win32 消息（自动化测试用）。
func DebugPump() int {
	return int(C.ui_debug_pump())
}

// DebugKey 向焦点控件模拟一次按键（自动化测试用），例如 VK_RETURN=0x0D。
func (w Window) DebugKey(vk int) bool {
	if w.h == 0 {
		return false
	}
	return C.ui_debug_key(w.h, C.int(vk)) == 0
}

// DebugDumpTree 返回窗口控件树的 JSON（自动化测试/调试用）。
func (w Window) DebugDumpTree() string {
	if w.h == 0 {
		return ""
	}
	p := C.ui_debug_dump_tree(w.h)
	if p == nil {
		return ""
	}
	defer C.ui_debug_free(p)
	return C.GoString(p)
}

// DebugScreenshot 把窗口内容保存成 PNG（自动化测试用）。
func (w Window) DebugScreenshot(path string) bool {
	if w.h == 0 {
		return false
	}
	p := cWide(path)
	defer freeWide(p)
	return C.ui_debug_screenshot(w.h, p) == 0
}
