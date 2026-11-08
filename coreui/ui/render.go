package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/YuFu-Onship/core-ui-go-binding/coreui"
)

// buildSource 把 Go 节点树翻译成 core-ui 能编译的 .uix 源码。
func (a *App) buildSource(root *Node) string {
	var b strings.Builder

	b.WriteString(a.windowTag())
	b.WriteString("\n\n<script>\nexport default {\n  data() {\n    return {\n")
	entries := make([]string, 0, len(a.states)+1)
	for _, s := range a.states {
		entries = append(entries, fmt.Sprintf("%q: %s", s.stateName(), s.stateJSON()))
	}
	entries = append(entries, `"__evt": {"seq": 0, "id": 0, "args": []}`)
	b.WriteString("      ")
	b.WriteString(strings.Join(entries, ",\n      "))
	b.WriteString("\n    };\n  },\n  methods: {\n")
	b.WriteString("    __emit(id) {\n")
	b.WriteString("      var args = [];\n")
	b.WriteString("      for (var i = 1; i < arguments.length; i++) args.push(arguments[i]);\n")
	b.WriteString("      this.__evt = { seq: this.__evt.seq + 1, id: id, args: args };\n")
	b.WriteString("    }\n  }\n}\n</script>\n\n")

	b.WriteString("<style>\n")
	b.WriteString(".ui-root { flex-direction: column; }\n")
	if a.opts.CSS != "" {
		b.WriteString(a.opts.CSS)
		if !strings.HasSuffix(a.opts.CSS, "\n") {
			b.WriteString("\n")
		}
	}
	b.WriteString("</style>\n\n")

	bg := a.opts.Background
	if bg == "" {
		bg = "var(--bg)"
	}
	b.WriteString("<template>\n<div class=\"ui-root\" style=\"flex-direction: column; background: ")
	b.WriteString(escAttr(bg))
	b.WriteString(";\">\n")

	// 让容器型根节点默认填满窗口；用户显式设置高度时不干预。
	if root != nil && isContainerTag(root.tag) && !hasHeightStyle(root) {
		root.style = append(root.style, "flex: 1")
	}

	r := &renderer{app: a}
	r.render(root, 1)
	b.WriteString(r.b.String())
	b.WriteString("\n</div>\n</template>\n")
	return b.String()
}

func (a *App) windowTag() string {
	title := a.opts.Title
	if title == "" {
		title = "Core UI + Go"
	}
	w, h := a.opts.Width, a.opts.Height
	if w <= 0 {
		w = 900
	}
	if h <= 0 {
		h = 640
	}
	var b strings.Builder
	b.WriteString("<window")
	fmt.Fprintf(&b, ` title="%s" width="%d" height="%d"`, escAttr(title), w, h)
	if a.opts.MinWidth > 0 {
		fmt.Fprintf(&b, ` min-width="%d"`, a.opts.MinWidth)
	}
	if a.opts.MinHeight > 0 {
		fmt.Fprintf(&b, ` min-height="%d"`, a.opts.MinHeight)
	}
	fmt.Fprintf(&b, ` resizable="%s" centered="%s" frameless="%s"`,
		boolStr(!a.opts.FixedSize), boolStr(a.opts.Centered), boolStr(a.opts.Frameless))
	if a.opts.Theme == coreui.Light {
		b.WriteString(` theme="light"`)
	} else {
		b.WriteString(` theme="dark"`)
	}
	b.WriteString("/>")
	return b.String()
}

type renderer struct {
	app *App
	b   strings.Builder
}

func (r *renderer) render(n *Node, depth int) {
	if n == nil {
		return
	}
	indent := strings.Repeat("  ", depth)
	r.b.WriteString(indent)
	r.b.WriteString("<")
	r.b.WriteString(n.tag)

	if len(n.classes) > 0 {
		fmt.Fprintf(&r.b, ` class="%s"`, escAttr(strings.Join(n.classes, " ")))
	}
	if len(n.style) > 0 {
		fmt.Fprintf(&r.b, ` style="%s"`, escAttr(strings.Join(n.style, "; ")))
	}
	for _, a := range n.attrs {
		fmt.Fprintf(&r.b, ` %s="%s"`, a.name, escAttr(a.value))
	}
	for _, a := range n.bindings {
		fmt.Fprintf(&r.b, ` :%s="%s"`, a.name, escAttr(a.value))
	}
	if n.vIf != "" {
		fmt.Fprintf(&r.b, ` v-if="%s"`, escAttr(n.vIf))
	}
	if n.vShow != "" {
		fmt.Fprintf(&r.b, ` v-show="%s"`, escAttr(n.vShow))
	}
	if n.vModel != "" {
		fmt.Fprintf(&r.b, ` v-model="%s"`, escAttr(n.vModel))
	}

	events := map[string]string{}
	for _, h := range n.handlers {
		id := r.app.addHandler(h)
		expr := fmt.Sprintf("__emit(%d", id)
		for _, arg := range h.args {
			expr += ", " + arg
		}
		expr += ")"
		events[h.event] = expr
	}
	// map 遍历顺序不稳定，排一下让生成的源码可复现。
	names := make([]string, 0, len(events))
	for name := range events {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		fmt.Fprintf(&r.b, ` @%s="%s"`, name, escAttr(events[name]))
	}

	hasBody := len(n.text) > 0 || len(n.children) > 0
	if !hasBody && isVoidTag(n.tag) {
		r.b.WriteString(" />\n")
		return
	}
	r.b.WriteString(">")

	if len(n.text) > 0 {
		r.b.WriteString(renderInlineText(n.text))
	}
	if len(n.children) > 0 {
		if len(n.text) > 0 {
			r.b.WriteString("\n")
		} else {
			r.b.WriteString("\n")
		}
		for _, c := range n.children {
			r.render(c, depth+1)
		}
		r.b.WriteString(indent)
	}
	r.b.WriteString("</")
	r.b.WriteString(n.tag)
	r.b.WriteString(">\n")
}

func renderInlineText(parts []textPart) string {
	var b strings.Builder
	for _, p := range parts {
		if p.literal != "" {
			b.WriteString(escText(p.literal))
		}
		if p.expr != "" {
			b.WriteString("{{ ")
			b.WriteString(p.expr)
			b.WriteString(" }}")
		}
	}
	return b.String()
}

func isContainerTag(tag string) bool {
	switch tag {
	case "div", "section", "nav", "ScrollView", "tabs", "TabControl":
		return true
	}
	return false
}

func hasHeightStyle(n *Node) bool {
	for _, d := range n.style {
		if strings.Contains(d, "height:") {
			return true
		}
	}
	return false
}

func isVoidTag(tag string) bool {
	switch tag {
	case "input", "img", "progressbar", "separator", "br", "hr", "option":
		return true
	}
	return false
}

func escAttr(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	return s
}

func escText(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	// 字面量里的 {{ }} 不能当模板插值；实体在解析阶段会还原。
	s = strings.ReplaceAll(s, "{", "&#123;")
	s = strings.ReplaceAll(s, "}", "&#125;")
	return s
}
