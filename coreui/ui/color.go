package ui

import (
	"fmt"
	"image/color"
)

// core-ui 主题色 token。它们对应页面 <style> 里的 CSS 变量，跟随明/暗主题
// 自动切换，比直接写 "var(--fg-3)" 更好记。
const (
	// 文本前景层级
	Fg1        = "var(--fg)"
	Fg2        = "var(--fg-2)"
	Fg3        = "var(--fg-3)"
	Fg4        = "var(--fg-4)"
	FgOnAccent = "var(--fg-on-accent)"

	// 背景层级
	Bg1        = "var(--bg)"
	Bg2        = "var(--bg-2)"
	Bg3        = "var(--bg-3)"
	Bg4        = "var(--bg-4)"
	BgElevated = "var(--bg-elevated)"

	// 窗口 / 卡片 / 侧栏 / 输入框 / 按钮
	WindowBg     = "var(--window-bg)"
	WindowBorder = "var(--window-border)"
	CardBg       = "var(--card-bg)"
	CardBorder   = "var(--card-border)"

	SidebarBg    = "var(--sidebar-bg)"
	SidebarText  = "var(--sidebar-text)"
	SidebarHover = "var(--sidebar-hover)"

	InputBg          = "var(--input-bg)"
	InputBorder      = "var(--input-border)"
	InputBorderHover = "var(--input-border-hover)"
	InputBorderFocus = "var(--input-border-focus)"

	BtnBg    = "var(--btn-bg)"
	BtnHover = "var(--btn-hover)"
	BtnPress = "var(--btn-press)"
	BtnText  = "var(--btn-text)"

	// 强调色与分隔线
	Accent         = "var(--accent)"
	AccentHover    = "var(--accent-hover)"
	AccentPress    = "var(--accent-press)"
	AccentText     = "var(--accent-text)"
	AccentSelected = "var(--accent-selected)"
	BorderColor    = "var(--border)"
	BorderSubtle   = "var(--border-subtle)"

	DisabledBg   = "var(--disabled-bg)"
	DisabledText = "var(--disabled-text)"

	White       = "#ffffff"
	Black       = "#000000"
	Transparent = "transparent"

	// 便捷别名：最常用的一级前景 / 背景。
	Fg = Fg1
	Bg = Bg1
)

// FromColor 把 Go 标准库的 color.Color（color.RGBA / color.NRGBA /
// image.Uniform 等）转成 core-ui CSS 颜色。
//
// 不透明颜色输出 "#rrggbb"，带透明度输出 "#rrggbbaa"。
func FromColor(c color.Color) string {
	if c == nil {
		return Transparent
	}
	r16, g16, b16, a16 := c.RGBA()
	if a16 == 0 {
		return Transparent
	}
	// color.Color.RGBA 返回 alpha-premultiplied 16-bit 通道；先还原成
	// 非预乘的 8-bit 分量，避免半透明颜色发暗。
	if a16 != 0xffff {
		r16 = uint32(uint64(r16) * 0xffff / uint64(a16))
		g16 = uint32(uint64(g16) * 0xffff / uint64(a16))
		b16 = uint32(uint64(b16) * 0xffff / uint64(a16))
		if r16 > 0xffff {
			r16 = 0xffff
		}
		if g16 > 0xffff {
			g16 = 0xffff
		}
		if b16 > 0xffff {
			b16 = 0xffff
		}
	}
	r, g, b := uint8(r16>>8), uint8(g16>>8), uint8(b16>>8)
	a := uint8(a16 >> 8)
	if a == 0xff {
		return fmt.Sprintf("#%02x%02x%02x", r, g, b)
	}
	return fmt.Sprintf("#%02x%02x%02x%02x", r, g, b, a)
}

// RGB 用 8-bit 分量创建不透明颜色。
func RGB(r, g, b uint8) string {
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

// RGBA 用 8-bit 分量创建带透明度颜色（a=0xff 时与 RGB 相同）。
func RGBA(r, g, b, a uint8) string {
	if a == 0xff {
		return RGB(r, g, b)
	}
	return fmt.Sprintf("#%02x%02x%02x%02x", r, g, b, a)
}

// cssColor 统一把主题 token / CSS 字符串 / color.Color 转成 CSS 文本。
func cssColor(v any) string {
	switch c := v.(type) {
	case nil:
		return Transparent
	case string:
		return c
	case color.Color:
		return FromColor(c)
	case fmt.Stringer:
		return c.String()
	default:
		panic(fmt.Sprintf("ui: 不支持的颜色类型 %T", v))
	}
}
