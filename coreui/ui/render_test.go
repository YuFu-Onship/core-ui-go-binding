package ui

import (
	"strings"
	"testing"

	"github.com/YuFu-Onship/core-ui-go-binding/coreui"
)

func buildTestApp() *App {
	app := NewApp(Options{Title: "Test", Width: 640, Height: 480, Theme: coreui.Light})
	app.View(func() *Node { return Column() })
	return app
}

func TestSourceContainsWindowAndBindings(t *testing.T) {
	app := NewApp(Options{
		Title:    "Counter",
		Width:    520,
		Height:   360,
		Theme:    coreui.Light,
		Centered: true,
	})
	count := app.Int(3)
	name := app.Text("世界")
	show := app.Bool(false)

	app.View(func() *Node {
		return Column(
			Text("count = ", count).FontSize(20),
			Input(name).Placeholder("name"),
			Check(show),
			Row(
				Button("inc").Primary().OnClick(func() { count.Inc() }),
				When(show, Text("details")),
			).Gap(8),
		).Padding(12).Gap(6)
	})

	src, err := app.Source()
	if err != nil {
		t.Fatalf("Source: %v", err)
	}

	for _, want := range []string{
		`<window title="Counter" width="520" height="360"`,
		`resizable="true" centered="true" frameless="false" theme="light"/>`,
		`"s1": 3`,
		`"s2": "世界"`,
		`"s3": false`,
		`count = {{ s1 }}`,
		`<input type="text" placeholder="name" v-model="s2" />`,
		`<input type="checkbox" v-model="s3" />`,
		`type="primary" @click="__emit(`,
		`v-if="s3"`,
		`padding: 12px`,
		`gap: 6px`,
		`flex: 1`, // 视图根节点自动填满窗口
	} {
		if !strings.Contains(src, want) {
			t.Errorf("生成的源码缺少 %q\n---\n%s", want, src)
		}
	}
}

func TestSourceEscapesTextAndAttributes(t *testing.T) {
	app := buildTestApp()
	app.View(func() *Node {
		return Column(Text(`a < b & "c" {{x}}`).Attr("tooltip", `x < y & "z"`))
	})
	src, err := app.Source()
	if err != nil {
		t.Fatalf("Source: %v", err)
	}
	if !strings.Contains(src, `a &lt; b &amp; "c" &#123;&#123;x&#125;&#125;`) {
		t.Errorf("文本/属性没有正确转义:\n%s", src)
	}
	if !strings.Contains(src, `tooltip="x &lt; y &amp; &quot;z&quot;"`) {
		t.Errorf("tooltip 没有正确转义:\n%s", src)
	}
}

func TestStateSettersAndListeners(t *testing.T) {
	app := NewApp(Options{})
	i := app.Int(1)
	f := app.Float(1.5)
	b := app.Bool(false)
	s := app.Text("a")

	var gotInt int
	var gotBool bool
	var gotText string
	i.OnChange(func(v int) { gotInt = v })
	b.OnChange(func(v bool) { gotBool = v })
	s.OnChange(func(v string) { gotText = v })

	i.Inc()
	i.Add(3)
	f.Add(0.5)
	b.Toggle()
	s.Set("b")

	if i.Get() != 5 || f.Get() != 2.0 || !b.Get() || s.Get() != "b" {
		t.Fatalf("状态值不对: i=%d f=%v b=%v s=%q", i.Get(), f.Get(), b.Get(), s.Get())
	}
	if gotInt != 5 || !gotBool || gotText != "b" {
		t.Fatalf("变更回调没有触发: int=%d bool=%v text=%q", gotInt, gotBool, gotText)
	}
}

func TestEachAndWhen(t *testing.T) {
	app := NewApp(Options{})
	show := app.Bool(true)
	app.View(func() *Node {
		return Column(
			Each([]string{"a", "b"}, func(s string, i int) *Node {
				return Text(s)
			}),
			When(show, Text("hidden")),
			Unless(show, Text("shown")),
		)
	})
	src, err := app.Source()
	if err != nil {
		t.Fatalf("Source: %v", err)
	}
	if strings.Count(src, "<label>a</label>") != 1 || strings.Count(src, "<label>b</label>") != 1 {
		t.Errorf("Each 没有展开:\n%s", src)
	}
	if !strings.Contains(src, `v-if="s1"`) || !strings.Contains(src, `v-if="!(s1)"`) {
		t.Errorf("When/Unless 生成错误:\n%s", src)
	}
}

func TestWhenIntAndFullWidth(t *testing.T) {
	app := NewApp(Options{})
	page := app.Int(2)
	app.View(func() *Node {
		return Column(
			WhenInt(page, 1, Text("one")),
			UnlessInt(page, 1, Text("not-one")),
			ShowWhenInt(page, 1, Text("shown")),
			HiddenWhenInt(page, 1, Text("hidden")),
			TextArea(app.Text("")).FullWidth(),
		)
	})
	src, err := app.Source()
	if err != nil {
		t.Fatalf("Source: %v", err)
	}
	for _, want := range []string{
		`v-if="s1 == 1"`,
		`v-if="s1 != 1"`,
		`v-show="s1 == 1"`,
		`v-show="s1 != 1"`,
		`width: 100%`,
	} {
		if !strings.Contains(src, want) {
			t.Errorf("源码缺少 %q:\n%s", want, src)
		}
	}
}

func TestFloatFormatting(t *testing.T) {
	if got := num(1.25); got != "1.25" {
		t.Errorf("num(1.25)=%q", got)
	}
	if got := px(0); got != "0" {
		t.Errorf("px(0)=%q", got)
	}
	if got := px(12); got != "12px" {
		t.Errorf("px(12)=%q", got)
	}
}
