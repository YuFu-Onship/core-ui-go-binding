package ui

import (
	"image/color"
	"strings"
	"testing"
)

func TestColorTokens(t *testing.T) {
	if Fg1 != "var(--fg)" || Fg4 != "var(--fg-4)" {
		t.Fatalf("Fg token 不对: %q %q", Fg1, Fg4)
	}
	if CardBg != "var(--card-bg)" || BorderSubtle != "var(--border-subtle)" {
		t.Fatalf("背景/边框 token 不对: %q %q", CardBg, BorderSubtle)
	}
}

func TestColorFromStandardLibrary(t *testing.T) {
	cases := []struct {
		name string
		in   color.Color
		want string
	}{
		{"RGBA opaque", color.RGBA{R: 0x25, G: 0x63, B: 0xeb, A: 0xff}, "#2563eb"},
		{"NRGBA", color.NRGBA{R: 0x25, G: 0x63, B: 0xeb, A: 0xff}, "#2563eb"},
		{"NRGBA alpha", color.NRGBA{R: 0x25, G: 0x63, B: 0xeb, A: 0x80}, "#2563eb80"},
		{"transparent", color.RGBA{}, "transparent"},
	}
	for _, tc := range cases {
		if got := FromColor(tc.in); got != tc.want {
			t.Errorf("%s: FromColor=%q want %q", tc.name, got, tc.want)
		}
	}
	if got := RGB(0x25, 0x63, 0xeb); got != "#2563eb" {
		t.Errorf("RGB=%q", got)
	}
	if got := RGBA(0x25, 0x63, 0xeb, 0x80); got != "#2563eb80" {
		t.Errorf("RGBA=%q", got)
	}
}

func TestColorModifiersAcceptAnyColor(t *testing.T) {
	app := NewApp(Options{})
	app.View(func() *Node {
		return Column(
			Text("a").Color(Fg3),
			Text("b").Color(color.NRGBA{R: 0xff, G: 0x00, B: 0x00, A: 0xff}),
			Text("c").Bg(CardBg),
		)
	})
	src, err := app.Source()
	if err != nil {
		t.Fatalf("Source: %v", err)
	}
	for _, want := range []string{
		`color: var(--fg-3)`,
		`color: #ff0000`,
		`background: var(--card-bg)`,
	} {
		if !strings.Contains(src, want) {
			t.Errorf("源码缺少 %q:\n%s", want, src)
		}
	}
}
