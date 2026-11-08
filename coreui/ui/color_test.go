package ui

import (
	"image/color"
	"strings"
	"testing"

	"github.com/YuFu-Onship/core-ui-go-binding/coreui"
)

func TestColorTokens(t *testing.T) {
	applyPalette(false, "")
	if Fg1 != RGB(0x24, 0x24, 0x24) || Fg4 != RGB(0x70, 0x70, 0x70) {
		t.Fatalf("Fg token mismatch: %v %v", Fg1, Fg4)
	}
	if CardBg != RGB(0xff, 0xff, 0xff) || BorderSubtle != RGB(0xe0, 0xe0, 0xe0) {
		t.Fatalf("surface token mismatch: %v %v", CardBg, BorderSubtle)
	}
	applyPalette(true, "")
	if Fg1 != White || Fg3 != RGB(0xad, 0xad, 0xad) {
		t.Fatalf("dark Fg token mismatch: %v %v", Fg1, Fg3)
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
}

func TestColorModifiersRequireRGBA(t *testing.T) {
	app := NewApp(Options{Theme: coreui.Light})
	app.View(func() *Node {
		return Column(
			Text("a").Color(Fg3),
			Text("b").Color(color.RGBA{R: 0xff, G: 0x00, B: 0x00, A: 0xff}),
			Text("c").Bg(CardBg),
		)
	})
	src, err := app.Source()
	if err != nil {
		t.Fatalf("Source: %v", err)
	}
	for _, want := range []string{
		`color: #616161`,
		`color: #ff0000`,
		`background: #ffffff`,
	} {
		if !strings.Contains(src, want) {
			t.Errorf("generated source missing %q:\n%s", want, src)
		}
	}
}
