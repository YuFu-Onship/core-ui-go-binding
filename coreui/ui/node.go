// Package ui 在 core-ui 之上提供一套纯 Go 的声明式界面层。
//
// 它把 gpui / iced / Flutter 那种“用 Go 表达式描述界面”的写法翻译成
// core-ui 运行时认识的 .uix 源码，再交给 coreui.LoadPageString 编译。
// 使用方不需要写 .uix、不需要写 JavaScript，也不需要关心模板语法：
//
//	app := ui.NewApp(ui.Options{Title: "Counter", Width: 520, Height: 320})
//	count := app.Int(0)
//
//	app.View(func() *ui.Node {
//	    return ui.Column(
//	        ui.Text("count = ", count).FontSize(20),
//	        ui.Row(
//	            ui.Button("+1").OnClick(func() { count.Inc() }),
//	            ui.Button("reset").OnClick(func() { count.Set(0) }),
//	        ).Gap(8),
//	    ).Padding(24).Gap(12)
//	})
//
//	app.Run()
//
// 状态（IntState / FloatState / BoolState / TextState）同时存在于 Go 和
// 页面的响应式数据里：Go 端调用 Set/Inc 会立即刷新界面；输入框、开关、
// 滑块等控件上的用户操作也会回写到 Go 状态。需要改变界面结构时调用
// app.Rebuild() 重新渲染视图。
package ui

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Node 是声明式界面树上的一个节点。
type Node struct {
	tag      string
	text     []textPart
	attrs    []attrPair
	bindings []attrPair // :name="expr"
	events   []attrPair // @name="expr"
	vIf      string
	vShow    string
	vModel   string
	children []*Node
	style    []string
	classes  []string
	handlers []handlerSpec
}

type textPart struct {
	literal string
	expr    string
}

type attrPair struct {
	name  string
	value string
}

type handlerSpec struct {
	event string
	args  []string // 追加到 __emit(id, ...) 的原始 JS 表达式
	fn    func(args []json.RawMessage)
}

// binder 是能被模板插值的 Go 值，目前由四种状态类型实现。
type binder interface {
	bindExpr() string
}

// Nodes 表示一组同级节点，方便 Each 之类的辅助函数返回。
type Nodes []*Node

// El 创建一个任意标签的节点（高级用法）。
func El(tag string) *Node { return &Node{tag: tag} }

// Text 创建一个文本节点，参数可以混合字面量字符串和响应式状态：
//
//	ui.Text("Hello, ", name, "! count = ", count)
func Text(parts ...any) *Node {
	n := &Node{tag: "label"}
	n.addText(parts...)
	return n
}

// Label 是 Text 的别名。
func Label(parts ...any) *Node { return Text(parts...) }

// H1/H2/H3 是带默认字号与字重的文本便捷函数。
func H1(parts ...any) *Node { return Text(parts...).FontSize(26).Bold() }
func H2(parts ...any) *Node { return Text(parts...).FontSize(21).Bold() }
func H3(parts ...any) *Node { return Text(parts...).FontSize(17).Bold() }

// Column 创建纵向 flex 容器。字符串/状态作为子项时会自动变成 Text。
func Column(children ...any) *Node {
	n := &Node{tag: "div"}
	n.style = append(n.style, "flex-direction: column")
	n.Add(children...)
	return n
}

// Row 创建横向 flex 容器。
func Row(children ...any) *Node {
	n := &Node{tag: "div"}
	n.style = append(n.style, "flex-direction: row")
	n.Add(children...)
	return n
}

// Stack 创建一个普通 div 容器（默认就是纵向 flex）。
func Stack(children ...any) *Node {
	n := &Node{tag: "div"}
	n.Add(children...)
	return n
}

// Scroll 创建可滚动容器。
func Scroll(children ...any) *Node {
	n := &Node{tag: "ScrollView"}
	n.Add(children...)
	return n
}

// Card 创建一个带主题背景、圆角和内边距的卡片容器。
func Card(children ...any) *Node {
	n := Column(children...).Padding(16).Gap(10).Bg(CardBg).Radius(10)
	n.Border(1, CardBorder)
	return n
}

// Spacer 是占满剩余空间的弹性占位。
func Spacer() *Node {
	return (&Node{tag: "div"}).Style("flex: 1")
}

// Divider 是一条水平分隔线。
func Divider() *Node {
	return (&Node{tag: "div"}).Style("height: 1px; width: 100%").Bg(BorderColor)
}

// Button 创建一个按钮。
func Button(label string) *Node {
	n := &Node{tag: "button"}
	n.addText(label)
	return n
}

// Input 创建单行输入框，双向绑定到 TextState。
func Input(state *TextState) *Node {
	n := (&Node{tag: "input"}).Attr("type", "text")
	n.vModel = state.bindExpr()
	return n
}

// Password 创建密码输入框。
func Password(state *TextState) *Node {
	n := (&Node{tag: "input"}).Attr("type", "password")
	n.vModel = state.bindExpr()
	return n
}

// TextArea 创建多行输入框。
func TextArea(state *TextState) *Node {
	n := &Node{tag: "textarea"}
	n.vModel = state.bindExpr()
	return n
}

// Check 创建复选框，双向绑定到 BoolState。
func Check(state *BoolState) *Node {
	n := (&Node{tag: "input"}).Attr("type", "checkbox")
	n.vModel = state.bindExpr()
	return n
}

// CheckLabeled 创建“复选框 + 可点击文字”的一行。点击文字也会切换状态。
func CheckLabeled(label string, state *BoolState) *Node {
	return Row(
		Check(state),
		Text(label).OnClick(func() { state.Toggle() }),
	).Gap(6).Items(AlignCenter)
}

// Toggle 创建带文字标签的开关。
func Toggle(label string, state *BoolState) *Node {
	n := &Node{tag: "toggle"}
	n.addText(label)
	n.vModel = state.bindExpr()
	return n
}

// Slider 创建滑块，双向绑定到 FloatState。
func Slider(state *FloatState, min, max float64) *Node {
	n := (&Node{tag: "input"}).Attr("type", "range")
	n.Attr("min", num(min))
	n.Attr("max", num(max))
	n.vModel = state.bindExpr()
	return n
}

// Progress 创建进度条。value 可以是 *FloatState、*IntState、float64 或 int。
func Progress(value any) *Node {
	n := &Node{tag: "progressbar"}
	switch v := value.(type) {
	case binder:
		n.Bind("value", v.bindExpr())
	case float64:
		n.Attr("value", num(v))
	case int:
		n.Attr("value", num(float64(v)))
	default:
		panic(fmt.Sprintf("ui: Progress 不支持 %T", value))
	}
	return n
}

// Select 创建下拉选择框，双向绑定到 IntState（值为选项下标）。
func Select(state *IntState, options []string) *Node {
	n := &Node{tag: "select"}
	n.vModel = state.bindExpr()
	for _, opt := range options {
		n.children = append(n.children, &Node{tag: "option", text: []textPart{{literal: opt}}})
	}
	return n
}

// Image 创建一个图片控件，src 走 core-ui 的资源解析链。
func Image(src string) *Node {
	return (&Node{tag: "img"}).Attr("src", src)
}

// Each 把 Go 切片展开成一组节点，常用于列表：
//
//	ui.Column(ui.Each(todos, func(t Todo, i int) *ui.Node {
//	    return ui.Row(ui.Text(t.Title), ui.Button("删除").OnClick(func() { ... }))
//	})...)
func Each[T any](items []T, fn func(item T, index int) *Node) Nodes {
	out := make(Nodes, 0, len(items))
	for i, item := range items {
		if n := fn(item, i); n != nil {
			out = append(out, n)
		}
	}
	return out
}

// When 根据 BoolState 决定子树是否挂载（v-if 语义）。
func When(cond *BoolState, n *Node) *Node {
	if n != nil {
		n.vIf = cond.bindExpr()
	}
	return n
}

// Unless 是 When 的反向判断。
func Unless(cond *BoolState, n *Node) *Node {
	if n != nil {
		n.vIf = "!(" + cond.bindExpr() + ")"
	}
	return n
}

// ShowWhen 根据 BoolState 切换可见性（v-show 语义，节点始终挂载）。
func ShowWhen(cond *BoolState, n *Node) *Node {
	if n != nil {
		n.vShow = cond.bindExpr()
	}
	return n
}

// HiddenWhen 是 ShowWhen 的反向判断。
func HiddenWhen(cond *BoolState, n *Node) *Node {
	if n != nil {
		n.vShow = "!(" + cond.bindExpr() + ")"
	}
	return n
}

// WhenInt 在整数状态等于 value 时挂载子树（v-if 语义）。
func WhenInt(cond *IntState, value int, n *Node) *Node {
	if n != nil {
		n.vIf = fmt.Sprintf("%s == %d", cond.bindExpr(), value)
	}
	return n
}

// UnlessInt 是 WhenInt 的反向判断。
func UnlessInt(cond *IntState, value int, n *Node) *Node {
	if n != nil {
		n.vIf = fmt.Sprintf("%s != %d", cond.bindExpr(), value)
	}
	return n
}

// ShowWhenInt 在整数状态等于 value 时显示（v-show 语义）。
func ShowWhenInt(cond *IntState, value int, n *Node) *Node {
	if n != nil {
		n.vShow = fmt.Sprintf("%s == %d", cond.bindExpr(), value)
	}
	return n
}

// HiddenWhenInt 是 ShowWhenInt 的反向判断。
func HiddenWhenInt(cond *IntState, value int, n *Node) *Node {
	if n != nil {
		n.vShow = fmt.Sprintf("%s != %d", cond.bindExpr(), value)
	}
	return n
}

// Add 追加子节点。字符串会变成文本，状态会变成插值文本。
func (n *Node) Add(children ...any) *Node {
	if n == nil {
		return n
	}
	for _, c := range children {
		n.appendChild(c)
	}
	return n
}

func (n *Node) appendChild(v any) {
	switch x := v.(type) {
	case nil:
	case *Node:
		if x != nil {
			n.children = append(n.children, x)
		}
	case Nodes:
		n.children = append(n.children, x...)
	case []*Node:
		n.children = append(n.children, x...)
	case []any:
		for _, item := range x {
			n.appendChild(item)
		}
	case string:
		n.children = append(n.children, Text(x))
	case binder:
		n.children = append(n.children, Text(x))
	default:
		panic(fmt.Sprintf("ui: 不支持的子节点类型 %T", v))
	}
}

func (n *Node) addText(parts ...any) {
	for _, part := range parts {
		switch x := part.(type) {
		case nil:
		case string:
			n.appendLiteral(x)
		case binder:
			n.text = append(n.text, textPart{expr: x.bindExpr()})
		default:
			n.appendLiteral(fmt.Sprint(x))
		}
	}
}

func (n *Node) appendLiteral(s string) {
	if s == "" {
		return
	}
	if len(n.text) > 0 && n.text[len(n.text)-1].expr == "" {
		n.text[len(n.text)-1].literal += s
		return
	}
	n.text = append(n.text, textPart{literal: s})
}

// ---- 通用属性 ----------------------------------------------------------

// Attr 设置一个静态属性。
func (n *Node) Attr(name, value string) *Node {
	if n == nil {
		return n
	}
	n.attrs = append(n.attrs, attrPair{name: name, value: value})
	return n
}

// Bind 设置一个动态属性，expr 是页面脚本里的表达式。
func (n *Node) Bind(name, expr string) *Node {
	if n == nil {
		return n
	}
	n.bindings = append(n.bindings, attrPair{name: name, value: expr})
	return n
}

// ID 设置控件 id。
func (n *Node) ID(id string) *Node { return n.Attr("id", id) }

// Class 追加 CSS 类名。
func (n *Node) Class(names ...string) *Node {
	if n == nil {
		return n
	}
	n.classes = append(n.classes, names...)
	return n
}

// Style 追加内联 CSS 声明，例如 Style("padding: 12px", "gap: 8px")。
func (n *Node) Style(decls ...string) *Node {
	if n == nil {
		return n
	}
	n.style = append(n.style, decls...)
	return n
}

// Placeholder 设置输入框占位文字。
func (n *Node) Placeholder(s string) *Node { return n.Attr("placeholder", s) }

// Tooltip 设置悬停提示。
func (n *Node) Tooltip(s string) *Node { return n.Attr("tooltip", s) }

// Hidden 设置初始可见性。
func (n *Node) Hidden(v bool) *Node { return n.Attr("visible", boolStr(!v)) }

// Disabled 设置初始启用状态。
func (n *Node) Disabled(v bool) *Node { return n.Attr("enabled", boolStr(!v)) }

// ---- 布局 --------------------------------------------------------------

// Padding 设置四边内边距（DIP / px）。
func (n *Node) Padding(v float64) *Node {
	return n.Style("padding: " + px(v))
}

// PaddingXY 设置水平/垂直内边距。
func (n *Node) PaddingXY(x, y float64) *Node {
	return n.Style("padding: " + px(y) + " " + px(x))
}

// PaddingTRBL 按 上/右/下/左 设置内边距。
func (n *Node) PaddingTRBL(top, right, bottom, left float64) *Node {
	return n.Style(fmt.Sprintf("padding: %s %s %s %s", px(top), px(right), px(bottom), px(left)))
}

// Gap 设置 flex 子项间距。
func (n *Node) Gap(v float64) *Node { return n.Style("gap: " + px(v)) }

// Width 设置固定宽度。
func (n *Node) Width(v float64) *Node { return n.Style("width: " + px(v)) }

// Height 设置固定高度。
func (n *Node) Height(v float64) *Node { return n.Style("height: " + px(v)) }

// Size 同时设置宽高。
func (n *Node) Size(w, h float64) *Node {
	return n.Style("width: " + px(w) + "; height: " + px(h))
}

// MinWidth 设置最小宽度。
func (n *Node) MinWidth(v float64) *Node { return n.Style("min-width: " + px(v)) }

// MinHeight 设置最小高度。
func (n *Node) MinHeight(v float64) *Node { return n.Style("min-height: " + px(v)) }

// MaxWidth 设置最大宽度。
func (n *Node) MaxWidth(v float64) *Node { return n.Style("max-width: " + px(v)) }

// MaxHeight 设置最大高度。
func (n *Node) MaxHeight(v float64) *Node { return n.Style("max-height: " + px(v)) }

// FullWidth 让节点占满父容器宽度。
func (n *Node) FullWidth() *Node { return n.Style("width: 100%") }

// FullHeight 让节点占满父容器高度。
func (n *Node) FullHeight() *Node { return n.Style("height: 100%") }

// Grow 让节点占满剩余空间（flex: 1）。
func (n *Node) Grow() *Node { return n.Style("flex: 1") }

// Flex 设置 flex 权重。
func (n *Node) Flex(v float64) *Node { return n.Style("flex: " + num(v)) }

// Align 设置交叉轴对齐方式。
type Align string

const (
	AlignStart   Align = "flex-start"
	AlignCenter  Align = "center"
	AlignEnd     Align = "flex-end"
	AlignStretch Align = "stretch"
)

// Items 设置交叉轴对齐（flex align-items）。
func (n *Node) Items(a Align) *Node { return n.Style("align-items: " + string(a)) }

// Justify 设置主轴对齐（flex justify-content）。
type Justify string

const (
	JustifyStart        Justify = "flex-start"
	JustifyCenter       Justify = "center"
	JustifyEnd          Justify = "flex-end"
	JustifySpaceBetween Justify = "space-between"
	JustifySpaceAround  Justify = "space-around"
)

// Justify 设置主轴对齐。
func (n *Node) Justify(j Justify) *Node { return n.Style("justify-content: " + string(j)) }

// CenterContent 同时让子项在两条轴上居中。
func (n *Node) CenterContent() *Node {
	return n.Items(AlignCenter).Justify(JustifyCenter)
}

// ---- 视觉 --------------------------------------------------------------

// Bg 设置背景色。可以是主题 token（如 ui.Bg2）、CSS 字符串或
// color.Color（color.RGBA / color.NRGBA 等）。
func (n *Node) Bg(color any) *Node { return n.Style("background: " + cssColor(color)) }

// Color 设置文字颜色。可以是主题 token（如 ui.Fg3）、CSS 字符串或 color.Color。
func (n *Node) Color(color any) *Node { return n.Style("color: " + cssColor(color)) }

// FontSize 设置字号。
func (n *Node) FontSize(v float64) *Node { return n.Style("font-size: " + px(v)) }

// Bold 设置粗体。
func (n *Node) Bold() *Node { return n.Style("font-weight: 600") }

// Radius 设置圆角。
func (n *Node) Radius(v float64) *Node { return n.Style("border-radius: " + px(v)) }

// Opacity 设置透明度。
func (n *Node) Opacity(v float64) *Node { return n.Style("opacity: " + num(v)) }

// Border 设置边框。color 可以是主题 token、CSS 字符串或 color.Color。
func (n *Node) Border(width float64, color any) *Node {
	return n.Style(fmt.Sprintf("border: %s solid %s", px(width), cssColor(color)))
}

// BorderTop 设置上边框。
func (n *Node) BorderTop(width float64, color any) *Node {
	return n.Style(fmt.Sprintf("border-top: %s solid %s", px(width), cssColor(color)))
}

// BorderRight 设置右边框。
func (n *Node) BorderRight(width float64, color any) *Node {
	return n.Style(fmt.Sprintf("border-right: %s solid %s", px(width), cssColor(color)))
}

// BorderBottom 设置下边框。
func (n *Node) BorderBottom(width float64, color any) *Node {
	return n.Style(fmt.Sprintf("border-bottom: %s solid %s", px(width), cssColor(color)))
}

// BorderLeft 设置左边框。
func (n *Node) BorderLeft(width float64, color any) *Node {
	return n.Style(fmt.Sprintf("border-left: %s solid %s", px(width), cssColor(color)))
}

// Shadow 设置阴影。
func (n *Node) Shadow(css string) *Node { return n.Style("box-shadow: " + css) }

// TextAlign 设置文本对齐。
func (n *Node) TextAlign(v string) *Node { return n.Style("text-align: " + v) }

// Primary 把按钮设为主题强调色按钮。
func (n *Node) Primary() *Node { return n.Attr("type", "primary") }

// ---- 事件 --------------------------------------------------------------

// OnClick 注册点击回调。
func (n *Node) OnClick(fn func()) *Node {
	if n == nil {
		return n
	}
	n.handlers = append(n.handlers, handlerSpec{
		event: "click",
		fn: func([]json.RawMessage) {
			fn()
		},
	})
	return n
}

// OnDblClick 注册双击回调。
func (n *Node) OnDblClick(fn func()) *Node {
	if n == nil {
		return n
	}
	n.handlers = append(n.handlers, handlerSpec{
		event: "dblclick",
		fn: func([]json.RawMessage) {
			fn()
		},
	})
	return n
}

// OnEvent 注册任意 core-ui 模板事件（如 "change" / "submit" / "wheel"）。
// 事件自带的 $event 会作为第一个 JSON 参数传入，便于高级用法自己解码。
func (n *Node) OnEvent(event string, fn func(args []json.RawMessage)) *Node {
	if n == nil {
		return n
	}
	n.handlers = append(n.handlers, handlerSpec{
		event: event,
		args:  []string{"$event"},
		fn:    fn,
	})
	return n
}

// OnChange 是 OnEvent("change", ...) 的便捷写法。
func (n *Node) OnChange(fn func(args []json.RawMessage)) *Node {
	return n.OnEvent("change", fn)
}

// OnSubmit 注册文本控件的提交事件（TextInput / TextArea 内按 Enter）。
// 典型用法是搜索框、命令框、行内重命名。
func (n *Node) OnSubmit(fn func()) *Node {
	if n == nil {
		return n
	}
	n.handlers = append(n.handlers, handlerSpec{
		event: "submit",
		fn: func([]json.RawMessage) {
			fn()
		},
	})
	return n
}

// ---- 内部工具 ----------------------------------------------------------

func (n *Node) staticAttr(name, value string) {
	n.attrs = append(n.attrs, attrPair{name: name, value: value})
}

func px(v float64) string {
	if v == 0 {
		return "0"
	}
	return num(v) + "px"
}

func num(v float64) string {
	s := fmt.Sprintf("%.4f", v)
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	if s == "-0" {
		return "0"
	}
	return s
}

func boolStr(v bool) string {
	if v {
		return "true"
	}
	return "false"
}
