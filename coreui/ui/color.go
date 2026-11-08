package ui

import (
	"fmt"
	"image/color"
	"strconv"
)

// Theme tokens are plain color.RGBA values. They are refreshed from
// Options.Theme / App.SetTheme before each render, so theme switching still
// works when the app is rebuilt.
var (
	// 文本前景层级
	Fg1        color.RGBA
	Fg2        color.RGBA
	Fg3        color.RGBA
	Fg4        color.RGBA
	FgOnAccent color.RGBA

	// 背景层级
	Bg1        color.RGBA
	Bg2        color.RGBA
	Bg3        color.RGBA
	Bg4        color.RGBA
	BgElevated color.RGBA

	// 窗口 / 卡片 / 边线 / 输入框 / 按钮
	WindowBg     color.RGBA
	WindowBorder color.RGBA
	CardBg       color.RGBA
	CardBorder   color.RGBA

	SidebarBg    color.RGBA
	SidebarText  color.RGBA
	SidebarHover color.RGBA

	InputBg          color.RGBA
	InputBorder      color.RGBA
	InputBorderHover color.RGBA
	InputBorderFocus color.RGBA

	BtnBg    color.RGBA
	BtnHover color.RGBA
	BtnPress color.RGBA
	BtnText  color.RGBA

	// 强调色 / 边框
	Accent         color.RGBA
	AccentHover    color.RGBA
	AccentPress    color.RGBA
	AccentText     color.RGBA
	AccentSelected color.RGBA
	BorderColor    color.RGBA
	BorderSubtle   color.RGBA

	DisabledBg   color.RGBA
	DisabledText color.RGBA

	White       = color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
	Black       = color.RGBA{R: 0x00, G: 0x00, B: 0x00, A: 0xff}
	Transparent = color.RGBA{}

	// 最常用的前景 / 背景
	Fg = Fg1
	Bg = Bg1
)

func init() {
	applyPalette(false, "")
}

type palette struct {
	Fg1, Fg2, Fg3, Fg4, FgOnAccent color.RGBA
	Bg1, Bg2, Bg3, Bg4, BgElevated color.RGBA

	WindowBg, WindowBorder color.RGBA
	CardBg, CardBorder     color.RGBA

	SidebarBg, SidebarText, SidebarHover color.RGBA

	InputBg, InputBorder, InputBorderHover, InputBorderFocus color.RGBA
	BtnBg, BtnHover, BtnPress, BtnText                       color.RGBA

	Accent, AccentHover, AccentPress, AccentText, AccentSelected color.RGBA
	BorderColor, BorderSubtle                                    color.RGBA
	DisabledBg, DisabledText                                     color.RGBA
}

func rgb(r, g, b uint8) color.RGBA {
	return color.RGBA{R: r, G: g, B: b, A: 0xff}
}

func rgba(r, g, b, a uint8) color.RGBA {
	return color.RGBA{R: r, G: g, B: b, A: a}
}

func applyPalette(dark bool, accentHex string) {
	var p palette
	if dark {
		p = palette{
			Fg1: rgb(0xff, 0xff, 0xff), Fg2: rgb(0xd6, 0xd6, 0xd6),
			Fg3: rgb(0xad, 0xad, 0xad), Fg4: rgb(0x99, 0x99, 0x99),
			FgOnAccent: rgb(0xff, 0xff, 0xff),

			Bg1: rgb(0x29, 0x29, 0x29), Bg2: rgb(0x1f, 0x1f, 0x1f),
			Bg3: rgb(0x14, 0x14, 0x14), Bg4: rgb(0x0a, 0x0a, 0x0a),
			BgElevated: rgb(0x24, 0x24, 0x24),

			WindowBg: rgb(0x29, 0x29, 0x29), WindowBorder: rgb(0x47, 0x47, 0x47),
			CardBg: rgb(0x33, 0x33, 0x33), CardBorder: rgb(0x42, 0x42, 0x42),

			SidebarBg: rgb(0x1f, 0x1f, 0x1f), SidebarText: rgb(0xd6, 0xd6, 0xd6),
			SidebarHover: rgba(0xff, 0xff, 0xff, 0x14),

			InputBg: rgb(0x14, 0x14, 0x14), InputBorder: rgb(0x42, 0x42, 0x42),
			InputBorderHover: rgb(0x4d, 0x4d, 0x4d),

			BtnBg: Transparent, BtnHover: rgba(0xff, 0xff, 0xff, 0x14),
			BtnPress: rgba(0xff, 0xff, 0xff, 0x0a), BtnText: rgb(0xff, 0xff, 0xff),

			BorderColor: rgb(0x42, 0x42, 0x42), BorderSubtle: rgb(0x38, 0x38, 0x38),
			DisabledBg: rgb(0x14, 0x14, 0x14), DisabledText: rgb(0x5c, 0x5c, 0x5c),
		}
	} else {
		p = palette{
			Fg1: rgb(0x24, 0x24, 0x24), Fg2: rgb(0x42, 0x42, 0x42),
			Fg3: rgb(0x61, 0x61, 0x61), Fg4: rgb(0x70, 0x70, 0x70),
			FgOnAccent: rgb(0xff, 0xff, 0xff),

			Bg1: rgb(0xff, 0xff, 0xff), Bg2: rgb(0xfa, 0xfa, 0xfa),
			Bg3: rgb(0xf5, 0xf5, 0xf5), Bg4: rgb(0xf0, 0xf0, 0xf0),
			BgElevated: rgb(0xfc, 0xfc, 0xfc),

			WindowBg: rgb(0xff, 0xff, 0xff), WindowBorder: rgb(0xd1, 0xd1, 0xd1),
			CardBg: rgb(0xff, 0xff, 0xff), CardBorder: rgb(0xd1, 0xd1, 0xd1),

			SidebarBg: rgb(0xfa, 0xfa, 0xfa), SidebarText: rgb(0x42, 0x42, 0x42),
			SidebarHover: rgb(0xf5, 0xf5, 0xf5),

			InputBg: rgb(0xff, 0xff, 0xff), InputBorder: rgb(0xd1, 0xd1, 0xd1),
			InputBorderHover: rgb(0xc7, 0xc7, 0xc7),

			BtnBg: Transparent, BtnHover: rgb(0xf5, 0xf5, 0xf5),
			BtnPress: rgb(0xe0, 0xe0, 0xe0), BtnText: rgb(0x24, 0x24, 0x24),

			BorderColor: rgb(0xd1, 0xd1, 0xd1), BorderSubtle: rgb(0xe0, 0xe0, 0xe0),
			DisabledBg: rgb(0xf0, 0xf0, 0xf0), DisabledText: rgb(0xbd, 0xbd, 0xbd),
		}
	}

	base := defaultAccent(dark)
	if accentHex != "" {
		if c, ok := parseHexColor(accentHex); ok {
			base = c
		}
	}
	p.Accent = base
	p.AccentHover = shift(base, 20)
	p.AccentPress = shift(base, -31)
	p.AccentSelected = base
	p.AccentText = contrastOn(base)
	p.InputBorderFocus = base

	Fg1, Fg2, Fg3, Fg4, FgOnAccent = p.Fg1, p.Fg2, p.Fg3, p.Fg4, p.FgOnAccent
	Bg1, Bg2, Bg3, Bg4, BgElevated = p.Bg1, p.Bg2, p.Bg3, p.Bg4, p.BgElevated
	WindowBg, WindowBorder = p.WindowBg, p.WindowBorder
	CardBg, CardBorder = p.CardBg, p.CardBorder
	SidebarBg, SidebarText, SidebarHover = p.SidebarBg, p.SidebarText, p.SidebarHover
	InputBg, InputBorder, InputBorderHover, InputBorderFocus =
		p.InputBg, p.InputBorder, p.InputBorderHover, p.InputBorderFocus
	BtnBg, BtnHover, BtnPress, BtnText = p.BtnBg, p.BtnHover, p.BtnPress, p.BtnText
	Accent, AccentHover, AccentPress, AccentText, AccentSelected =
		p.Accent, p.AccentHover, p.AccentPress, p.AccentText, p.AccentSelected
	BorderColor, BorderSubtle = p.BorderColor, p.BorderSubtle
	DisabledBg, DisabledText = p.DisabledBg, p.DisabledText
	Fg, Bg = Fg1, Bg1
}

func defaultAccent(dark bool) color.RGBA {
	if dark {
		return rgb(0x11, 0x5e, 0xa3)
	}
	return rgb(0x39, 0x71, 0xec)
}

func shift(c color.RGBA, delta int) color.RGBA {
	clamp := func(v int) uint8 {
		if v < 0 {
			return 0
		}
		if v > 255 {
			return 255
		}
		return uint8(v)
	}
	return color.RGBA{
		R: clamp(int(c.R) + delta),
		G: clamp(int(c.G) + delta),
		B: clamp(int(c.B) + delta),
		A: c.A,
	}
}

func contrastOn(c color.RGBA) color.RGBA {
	y := 0.2126*float64(c.R) + 0.7152*float64(c.G) + 0.0722*float64(c.B)
	if y > 153 {
		return Black
	}
	return White
}

func parseHexColor(s string) (color.RGBA, bool) {
	if len(s) == 0 || s[0] != '#' {
		return color.RGBA{}, false
	}
	hex := s[1:]
	parse := func(v string) (uint8, bool) {
		n, err := strconv.ParseUint(v, 16, 8)
		return uint8(n), err == nil
	}
	switch len(hex) {
	case 3:
		r, ok1 := parse(string([]byte{hex[0], hex[0]}))
		g, ok2 := parse(string([]byte{hex[1], hex[1]}))
		b, ok3 := parse(string([]byte{hex[2], hex[2]}))
		return color.RGBA{R: r, G: g, B: b, A: 0xff}, ok1 && ok2 && ok3
	case 6:
		r, ok1 := parse(hex[0:2])
		g, ok2 := parse(hex[2:4])
		b, ok3 := parse(hex[4:6])
		return color.RGBA{R: r, G: g, B: b, A: 0xff}, ok1 && ok2 && ok3
	case 8:
		// 与 core-ui 的 ui_theme_set_accent_hex 一致：#RRGGBBAA
		r, ok1 := parse(hex[0:2])
		g, ok2 := parse(hex[2:4])
		b, ok3 := parse(hex[4:6])
		a, ok4 := parse(hex[6:8])
		return color.RGBA{R: r, G: g, B: b, A: a}, ok1 && ok2 && ok3 && ok4
	default:
		return color.RGBA{}, false
	}
}

// FromColor 把 Go 标准库 color.Color 转成 core-ui 可用的 CSS 颜色字符串。
func FromColor(c color.Color) string {
	if c == nil {
		return "transparent"
	}
	r16, g16, b16, a16 := c.RGBA()
	if a16 == 0 {
		return "transparent"
	}
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

// RGB 返回不透明的 color.RGBA。
func RGB(r, g, b uint8) color.RGBA {
	return color.RGBA{R: r, G: g, B: b, A: 0xff}
}

// RGBA 返回 color.RGBA（注意 Go 的 RGBA 是 alpha-premultiplied）。
func RGBA(r, g, b, a uint8) color.RGBA {
	return color.RGBA{R: r, G: g, B: b, A: a}
}

// cssColor 把 color.RGBA 转成 CSS 文本。
func cssColor(c color.RGBA) string {
	return FromColor(c)
}
