// Package coreui 是 core-ui 的 Go 绑定。
//
// core-ui (https://github.com/ghboke/core-ui) 是一个 Windows 桌面 UI 框架：
// 界面由一个 Vue 3 风格的单文件组件 `.uix` 声明，运行时是纯 C ABI 的
// core-ui.dll。本包在 cgo 之上做了一层薄封装，界面依然写在 `.uix` 里，
// Go 只负责加载、注入数据、注册原生回调。
//
// 最小用法：
//
//	coreui.InitWithTheme(coreui.Light)
//	page, err := coreui.LoadPageString(appUIX) // appUIX 来自 go:embed
//	win, err := page.OpenWindow(nil)
//	code := coreui.Run()
//	page.Destroy()
//	coreui.Shutdown()
//
// 构建前提：core-ui.dll 必须与 exe 同目录（或位于 PATH 里），
// 头文件与导入库位于 <模块根>/third_party/coreui/。
//
// 不想写 .uix 时可以用子包 github.com/YuFu-Onship/core-ui-go-binding/coreui/ui：它提供纯 Go 的声明式
// 界面 DSL，并在运行时自动生成 .uix 源码。
package coreui

/*
#cgo CFLAGS: -I${SRCDIR}/../third_party/coreui/include
#cgo LDFLAGS: -L${SRCDIR}/../third_party/coreui/lib -lcore-ui
#include <ui_core.h>
#include <stdlib.h>
*/
import "C"

import (
	"errors"
	"fmt"
	"runtime"
	"unsafe"
)

// Theme 是全局主题模式，与 UiThemeMode 一一对应。
type Theme int32

const (
	Dark  Theme = 0
	Light Theme = 1
)

// VersionInfo 是当前加载的 core-ui.dll 的版本信息。
type VersionInfo struct {
	Major int
	Minor int
	Patch int
	Build int
}

// String 返回 "major.minor.patch.build"。
func (v VersionInfo) String() string {
	return fmt.Sprintf("%d.%d.%d.%d", v.Major, v.Minor, v.Patch, v.Build)
}

// Compare 比较两个版本，返回 -1 / 0 / 1。
func (v VersionInfo) Compare(other VersionInfo) int {
	a := [4]int{v.Major, v.Minor, v.Patch, v.Build}
	b := [4]int{other.Major, other.Minor, other.Patch, other.Build}
	for i := range a {
		switch {
		case a[i] < b[i]:
			return -1
		case a[i] > b[i]:
			return 1
		}
	}
	return 0
}

// AtLeast 判断当前 DLL 版本是否不低于 want，适合启动时做 ABI 兼容检查。
func (v VersionInfo) AtLeast(want VersionInfo) bool {
	return v.Compare(want) >= 0
}

// DLLVersion 返回当前实际加载的 core-ui.dll 版本。
//
// 它读取的是运行时 DLL 导出的版本号，而不是 Go 模块或头文件版本，可用于
// 打包时检查用户机器上的 DLL 是否满足要求。
func DLLVersion() VersionInfo {
	var major, minor, patch C.int
	C.ui_core_version(&major, &minor, &patch)
	return VersionInfo{
		Major: int(major),
		Minor: int(minor),
		Patch: int(patch),
		Build: int(C.ui_core_version_build()),
	}
}

// Version 返回 core-ui 运行时的版本串，例如 "1.9.0.345"。
func Version() string {
	return C.GoString(C.ui_core_version_string())
}

// Init 以默认（暗色）主题初始化运行时，必须在任何其他调用之前执行。
func Init() error {
	return InitWithTheme(Dark)
}

// InitWithTheme 以指定主题初始化运行时。
//
// 初始化会把调用 goroutine 锁定到当前 OS 线程：core-ui 的窗口和消息循环
// 都绑定在线程上，goroutine 迁移线程会导致窗口收不到消息。
func InitWithTheme(t Theme) error {
	runtime.LockOSThread()
	if rc := C.ui_init_with_theme(C.UiThemeMode(t)); rc != 0 {
		return fmt.Errorf("coreui: ui_init_with_theme 失败 (rc=%d)", int32(rc))
	}
	return nil
}

// Shutdown 释放运行时资源，应在消息循环退出后调用。
func Shutdown() {
	C.ui_shutdown()
}

// Run 进入 Win32 消息循环，窗口全部关闭后返回退出码。
func Run() int {
	runtime.LockOSThread()
	return int(C.ui_run())
}

// Quit 请求消息循环退出，code 会成为 Run 的返回值。
func Quit(code int) {
	C.ui_quit(C.int(code))
}

// SetTheme 在运行时切换主题。
func SetTheme(t Theme) {
	C.ui_theme_set_mode(C.UiThemeMode(t))
}

// SetAccentHex 用 "#rrggbb" 或 "#aarrggbb" 设置主题强调色。
func SetAccentHex(hex string) error {
	cs := C.CString(hex)
	defer C.free(unsafe.Pointer(cs))
	if rc := C.ui_theme_set_accent_hex(cs); rc != 0 {
		return fmt.Errorf("coreui: 无效的强调色 %q", hex)
	}
	return nil
}

// SetQuickJSEnabled 控制 `.uix` 是否走 QuickJS 脚本运行时。
// 默认已开启，保留该开关只为兼容旧版宿主。
func SetQuickJSEnabled(enabled bool) {
	v := C.int(0)
	if enabled {
		v = 1
	}
	C.ui_page_set_quickjs_enabled(v)
}

// DefaultFont 返回主题默认字体族名。
func DefaultFont() string {
	return wideToGo(C.ui_theme_get_default_font())
}

// SetDefaultFont 设置主题默认字体族名。
func SetDefaultFont(family string) {
	w := cWide(family)
	defer freeWide(w)
	C.ui_theme_set_default_font(w)
}

// CJKFonts 返回 (拉丁字体, 中日韩字体)。
func CJKFonts() (latin, cjk string) {
	return wideToGo(C.ui_theme_get_cjk_latin_font()), wideToGo(C.ui_theme_get_cjk_cjk_font())
}

// SetCJKFont 分别设置拉丁与中日韩回退字体。
func SetCJKFont(latin, cjk string) {
	lw := cWide(latin)
	defer freeWide(lw)
	cw := cWide(cjk)
	defer freeWide(cw)
	C.ui_theme_set_cjk_font(lw, cw)
}

// TextLocale 返回文字排版使用的 locale（影响断行/字形选择）。
func TextLocale() string {
	return wideToGo(C.ui_theme_get_text_locale())
}

// SetTextLocale 设置文字排版 locale，例如 "zh-CN"。
func SetTextLocale(locale string) {
	w := cWide(locale)
	defer freeWide(w)
	C.ui_theme_set_text_locale(w)
}

// RegisterAssetDir 注册一个目录，`.uix` 里 <img src="logo.png"> 这类按名字
// 引用的资源会到该目录查找。可重复调用，先注册先匹配。
func RegisterAssetDir(dir string) {
	cs := C.CString(dir)
	defer C.free(unsafe.Pointer(cs))
	C.ui_asset_register_dir(cs)
}

// RegisterAssetBlob 把一份内存数据注册成资源名，用于把图片/样式表内嵌进 exe。
//
// 数据会被复制到 C 堆并保留到进程结束（库不复制也不释放），调用方可以自由复用 data。
func RegisterAssetBlob(name string, data []byte) {
	if len(data) == 0 {
		return
	}
	cn := C.CString(name)
	defer C.free(unsafe.Pointer(cn))
	p := C.malloc(C.size_t(len(data)))
	if p == nil {
		panic("coreui: 内存不足")
	}
	copy(unsafe.Slice((*byte)(p), len(data)), data)
	C.ui_asset_register_blob(cn, p, C.size_t(len(data)))
}

// ResetAssets 清空资源解析链。
func ResetAssets() {
	C.ui_asset_reset()
}

// AssetInvalidate 让解析失败过的同名资源重新尝试解析（异步加载完成后调用）。
func AssetInvalidate(name string) {
	cs := C.CString(name)
	defer C.free(unsafe.Pointer(cs))
	C.ui_asset_invalidate(cs)
}

// Page 对应一个已编译的 `.uix` 页面。
type Page struct {
	h C.UiPage
}

// LoadPageFile 从磁盘读取 `.uix` 文件并编译。
func LoadPageFile(path string) (*Page, error) {
	w := cWide(path)
	defer freeWide(w)
	return newPage(C.ui_page_load_file(w), path)
}

// LoadPageString 直接从字符串（例如 go:embed 的内容）编译 `.uix`。
func LoadPageString(source string) (*Page, error) {
	cs := C.CString(source)
	defer C.free(unsafe.Pointer(cs))
	return newPage(C.ui_page_load_string(cs), "内存中的 .uix")
}

func newPage(h C.UiPage, what string) (*Page, error) {
	if h == 0 {
		return nil, fmt.Errorf("coreui: 编译 %s 失败", what)
	}
	p := &Page{h: h}
	if p.Root().Handle() == 0 {
		msg := p.LastError()
		p.Destroy()
		if msg == "" {
			msg = "编译失败"
		}
		return nil, fmt.Errorf("coreui: %s", msg)
	}
	return p, nil
}

// Root 返回页面根控件。
func (p *Page) Root() Widget {
	if p == nil || p.h == 0 {
		return Widget{}
	}
	return Widget{h: C.ui_page_root(p.h)}
}

// Destroy 销毁页面及其控件树。消息循环结束后调用。
func (p *Page) Destroy() {
	if p == nil || p.h == 0 {
		return
	}
	C.ui_page_destroy(p.h)
	p.h = 0
}

// Reload 重新编译页面（磁盘加载的会重新读文件），已存在的响应式变量值按名字保留。
func (p *Page) Reload() error {
	if p == nil || p.h == 0 {
		return errors.New("coreui: 页面已销毁")
	}
	if C.ui_page_reload(p.h) == 0 {
		return fmt.Errorf("coreui: 重载失败: %s", p.LastError())
	}
	return nil
}

// LastError 返回最近一次编译或运行期错误，没有错误时返回空串。
func (p *Page) LastError() string {
	if p == nil || p.h == 0 {
		return ""
	}
	return C.GoString(C.ui_page_last_error(p.h))
}

// OpenWindow 按 `.uix` 里 <window> 标签创建并显示窗口，根控件自动挂载。
// cfg 为 nil 时完全使用标签里的设置；非 nil 时作为缺省值兜底。
func (p *Page) OpenWindow(cfg *WindowConfig) (Window, error) {
	return p.openWindow(cfg, true)
}

// PrepareWindow 与 OpenWindow 相同，但窗口保持隐藏，由调用方准备好数据后
// 自己调 Window.ShowImmediate，避免用户看到“空窗→有内容”的两段式显示。
func (p *Page) PrepareWindow(cfg *WindowConfig) (Window, error) {
	return p.openWindow(cfg, false)
}

func (p *Page) openWindow(cfg *WindowConfig, show bool) (Window, error) {
	if p == nil || p.h == 0 {
		return Window{}, errors.New("coreui: 页面已销毁")
	}
	cc, done := cfg.toC()
	defer done()

	var h C.UiWindow
	if show {
		h = C.ui_page_open_window(p.h, cc)
	} else {
		h = C.ui_page_prepare_window(p.h, cc)
	}
	if h == 0 {
		msg := p.LastError()
		if msg == "" {
			msg = "创建窗口失败"
		}
		return Window{}, fmt.Errorf("coreui: %s", msg)
	}
	return Window{h: h}, nil
}

// Menu 取得 `.uix` 里 <menu id="name"> 声明的菜单句柄，菜单归页面所有。
func (p *Page) Menu(name string) Menu {
	if p == nil || p.h == 0 {
		return Menu{}
	}
	cs := C.CString(name)
	defer C.free(unsafe.Pointer(cs))
	return Menu{h: C.ui_page_menu(p.h, cs)}
}

// SetBool 写入页面的布尔响应式变量。
func (p *Page) SetBool(name string, v bool) {
	if p == nil || p.h == 0 {
		return
	}
	cn := C.CString(name)
	defer C.free(unsafe.Pointer(cn))
	iv := C.int(0)
	if v {
		iv = 1
	}
	C.ui_page_set_bool(p.h, cn, iv)
}

// SetInt 写入页面的整数响应式变量。
func (p *Page) SetInt(name string, v int) {
	if p == nil || p.h == 0 {
		return
	}
	cn := C.CString(name)
	defer C.free(unsafe.Pointer(cn))
	C.ui_page_set_int(p.h, cn, C.int(v))
}

// SetFloat 写入页面的浮点响应式变量。
func (p *Page) SetFloat(name string, v float64) {
	if p == nil || p.h == 0 {
		return
	}
	cn := C.CString(name)
	defer C.free(unsafe.Pointer(cn))
	C.ui_page_set_float(p.h, cn, C.double(v))
}

// SetText 写入页面的字符串响应式变量。
func (p *Page) SetText(name, v string) {
	if p == nil || p.h == 0 {
		return
	}
	cn := C.CString(name)
	defer C.free(unsafe.Pointer(cn))
	wv := cWide(v)
	defer freeWide(wv)
	C.ui_page_set_text(p.h, cn, wv)
}

// SetTextList 写入字符串数组变量，供 v-for 使用；传 nil 会清空。
func (p *Page) SetTextList(name string, items []string) {
	if p == nil || p.h == 0 {
		return
	}
	cn := C.CString(name)
	defer C.free(unsafe.Pointer(cn))
	if len(items) == 0 {
		C.ui_page_set_text_list(p.h, cn, nil, 0)
		return
	}

	ptrs := make([]*C.wchar_t, len(items))
	for i, s := range items {
		ptrs[i] = cWide(s)
	}
	defer func() {
		for _, p := range ptrs {
			freeWide(p)
		}
	}()
	C.ui_page_set_text_list(p.h, cn, (**C.wchar_t)(unsafe.Pointer(&ptrs[0])), C.int(len(ptrs)))
}

// SetJSON 把任意 JSON 写成页面的响应式变量，返回 error 表示解析失败。
func (p *Page) SetJSON(name, json string) error {
	if p == nil || p.h == 0 {
		return errors.New("coreui: 页面已销毁")
	}
	cn := C.CString(name)
	defer C.free(unsafe.Pointer(cn))
	cj := C.CString(json)
	defer C.free(unsafe.Pointer(cj))
	if rc := C.ui_page_set_json(p.h, cn, cj); rc != 0 {
		return fmt.Errorf("coreui: 变量 %q 的 JSON 解析失败", name)
	}
	return nil
}

// GetJSON 读出页面变量的当前值（JSON 文本）；变量不存在时返回 error。
func (p *Page) GetJSON(name string) (string, error) {
	if p == nil || p.h == 0 {
		return "", errors.New("coreui: 页面已销毁")
	}
	cn := C.CString(name)
	defer C.free(unsafe.Pointer(cn))
	res := C.ui_page_get_json(p.h, cn)
	if res == nil {
		return "", fmt.Errorf("coreui: 变量 %q 不存在", name)
	}
	defer C.ui_page_free(unsafe.Pointer(res))
	return C.GoString(res), nil
}

// LoadTranslations 安装一组 key/value 翻译（kv 长度必须是偶数）。
func (p *Page) LoadTranslations(locale string, kv []string) {
	if p == nil || p.h == 0 || len(kv) < 2 {
		return
	}
	cl := C.CString(locale)
	defer C.free(unsafe.Pointer(cl))

	cstrs := make([]*C.char, len(kv))
	for i, s := range kv {
		cstrs[i] = C.CString(s)
	}
	defer func() {
		for _, s := range cstrs {
			C.free(unsafe.Pointer(s))
		}
	}()
	C.ui_page_load_translations(p.h, cl, (**C.char)(unsafe.Pointer(&cstrs[0])), C.int(len(cstrs)/2))
}

// LoadLanguageString 用 `.lang` 文本（UTF-8，每行 key=value）装载某个语言。
func (p *Page) LoadLanguageString(locale, content string) {
	if p == nil || p.h == 0 {
		return
	}
	cl := C.CString(locale)
	defer C.free(unsafe.Pointer(cl))
	cc := C.CString(content)
	defer C.free(unsafe.Pointer(cc))
	C.ui_page_load_language_string(p.h, cl, cc)
}

// LoadLanguageFile 从磁盘读取 `.lang` 文件装载某个语言。
func (p *Page) LoadLanguageFile(locale, path string) error {
	if p == nil || p.h == 0 {
		return errors.New("coreui: 页面已销毁")
	}
	cl := C.CString(locale)
	defer C.free(unsafe.Pointer(cl))
	w := cWide(path)
	defer freeWide(w)
	if rc := C.ui_page_load_language_file(p.h, cl, w); rc == 0 {
		return fmt.Errorf("coreui: 无法读取语言文件 %s", path)
	}
	return nil
}

// SetLocale 切换当前语言，模板里 $t() 的绑定会自动重译。
func (p *Page) SetLocale(locale string) {
	if p == nil || p.h == 0 {
		return
	}
	cl := C.CString(locale)
	defer C.free(unsafe.Pointer(cl))
	C.ui_page_set_locale(p.h, cl)
}
