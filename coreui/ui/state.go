package ui

import (
	"encoding/json"
	"strconv"

	"github.com/YuFu-Onship/core-ui-go-binding/coreui"
)

// stateBase 保存四种状态共用的页面绑定信息。
type stateBase struct {
	app  *App
	name string
	page *coreui.Page
}

func (b *stateBase) bindExpr() string  { return b.name }
func (b *stateBase) stateName() string { return b.name }

// bindPage 把状态挂到新页面上，并注册“页面 -> Go”的同步回调。
func (b *stateBase) bindPage(p *coreui.Page, sync func(string)) {
	b.page = p
	if p == nil {
		return
	}
	b.app.addWatch(p.Watch(b.name, sync))
}

func (b *stateBase) unbind() { b.page = nil }

// IntState 是同时存在于 Go 与页面响应式数据里的整数。
type IntState struct {
	stateBase
	v         int
	listeners []func(int)
}

// Int 创建一个整数状态。
func (a *App) Int(initial int) *IntState {
	s := &IntState{v: initial}
	s.app = a
	a.addState(s, func(name string) { s.name = name })
	return s
}

// Get 返回当前值。
func (s *IntState) Get() int { return s.v }

// Set 写入新值；如果有页面绑定会立即刷新界面。
func (s *IntState) Set(v int) {
	if s.v == v {
		return
	}
	s.v = v
	if s.page != nil {
		s.page.SetInt(s.name, v)
	}
	for _, fn := range s.listeners {
		fn(v)
	}
}

// Add 在原值上加 delta。
func (s *IntState) Add(delta int) { s.Set(s.v + delta) }

// Inc 自增 1。
func (s *IntState) Inc() { s.Set(s.v + 1) }

// Dec 自减 1。
func (s *IntState) Dec() { s.Set(s.v - 1) }

// OnChange 注册 Go 侧变更回调（包括用户操作控件导致的回写）。
func (s *IntState) OnChange(fn func(int)) *IntState {
	s.listeners = append(s.listeners, fn)
	return s
}

func (s *IntState) stateJSON() string { return strconv.Itoa(s.v) }

func (s *IntState) syncJSON(raw string) {
	var v int
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		return
	}
	if v == s.v {
		return
	}
	s.v = v
	for _, fn := range s.listeners {
		fn(v)
	}
}

// FloatState 是同时存在于 Go 与页面响应式数据里的浮点数。
type FloatState struct {
	stateBase
	v         float64
	listeners []func(float64)
}

// Float 创建一个浮点状态。
func (a *App) Float(initial float64) *FloatState {
	s := &FloatState{v: initial}
	s.app = a
	a.addState(s, func(name string) { s.name = name })
	return s
}

// Get 返回当前值。
func (s *FloatState) Get() float64 { return s.v }

// Set 写入新值并刷新界面。
func (s *FloatState) Set(v float64) {
	if s.v == v {
		return
	}
	s.v = v
	if s.page != nil {
		s.page.SetFloat(s.name, v)
	}
	for _, fn := range s.listeners {
		fn(v)
	}
}

// Add 在原值上加 delta。
func (s *FloatState) Add(delta float64) { s.Set(s.v + delta) }

// OnChange 注册 Go 侧变更回调。
func (s *FloatState) OnChange(fn func(float64)) *FloatState {
	s.listeners = append(s.listeners, fn)
	return s
}

func (s *FloatState) stateJSON() string { return strconv.FormatFloat(s.v, 'f', -1, 64) }

func (s *FloatState) syncJSON(raw string) {
	var v float64
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		return
	}
	if v == s.v {
		return
	}
	s.v = v
	for _, fn := range s.listeners {
		fn(v)
	}
}

// BoolState 是同时存在于 Go 与页面响应式数据里的布尔值。
type BoolState struct {
	stateBase
	v         bool
	listeners []func(bool)
}

// Bool 创建一个布尔状态。
func (a *App) Bool(initial bool) *BoolState {
	s := &BoolState{v: initial}
	s.app = a
	a.addState(s, func(name string) { s.name = name })
	return s
}

// Get 返回当前值。
func (s *BoolState) Get() bool { return s.v }

// Set 写入新值并刷新界面。
func (s *BoolState) Set(v bool) {
	if s.v == v {
		return
	}
	s.v = v
	if s.page != nil {
		s.page.SetBool(s.name, v)
	}
	for _, fn := range s.listeners {
		fn(v)
	}
}

// Toggle 翻转布尔值。
func (s *BoolState) Toggle() { s.Set(!s.v) }

// OnChange 注册 Go 侧变更回调。
func (s *BoolState) OnChange(fn func(bool)) *BoolState {
	s.listeners = append(s.listeners, fn)
	return s
}

func (s *BoolState) stateJSON() string {
	if s.v {
		return "true"
	}
	return "false"
}

func (s *BoolState) syncJSON(raw string) {
	var v bool
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		return
	}
	if v == s.v {
		return
	}
	s.v = v
	for _, fn := range s.listeners {
		fn(v)
	}
}

// TextState 是同时存在于 Go 与页面响应式数据里的字符串。
type TextState struct {
	stateBase
	v         string
	listeners []func(string)
}

// Text 创建一个字符串状态。
func (a *App) Text(initial string) *TextState {
	s := &TextState{v: initial}
	s.app = a
	a.addState(s, func(name string) { s.name = name })
	return s
}

// Get 返回当前值。
func (s *TextState) Get() string { return s.v }

// Set 写入新值并刷新界面。
func (s *TextState) Set(v string) {
	if s.v == v {
		return
	}
	s.v = v
	if s.page != nil {
		s.page.SetText(s.name, v)
	}
	for _, fn := range s.listeners {
		fn(v)
	}
}

// OnChange 注册 Go 侧变更回调。
func (s *TextState) OnChange(fn func(string)) *TextState {
	s.listeners = append(s.listeners, fn)
	return s
}

func (s *TextState) stateJSON() string {
	b, _ := json.Marshal(s.v)
	return string(b)
}

func (s *TextState) syncJSON(raw string) {
	var v string
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		return
	}
	if v == s.v {
		return
	}
	s.v = v
	for _, fn := range s.listeners {
		fn(v)
	}
}
