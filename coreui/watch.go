package coreui

/*
#include <ui_core.h>
#include <stdlib.h>
#include "callbacks.h"
*/
import "C"

import (
	"encoding/json"
	"unsafe"
)

// Watch 监听页面响应式变量 name 的变化。
//
// 变化来源包括 ui_page_set_* 写入、模板里的 v-model、以及页面脚本里的赋值。
// 同一页面内同名变量只保留一个回调（后注册的替换先前的）；返回值用于取消监听。
//
// fn 在 UI 线程上同步执行，参数是该变量当前值的 UTF-8 JSON。不要在回调里
// 读取同一个变量（库的注释明确禁止），写其他变量、调用 Set* 都可以。
func (p *Page) Watch(name string, fn func(json string)) (cancel func()) {
	if p == nil || p.h == 0 || fn == nil {
		return func() {}
	}
	id := registerCallback(watchCallbacks, fn)
	ud := callbackPtr(id)
	cn := C.CString(name)
	defer C.free(unsafe.Pointer(cn))
	C.coreui_page_watch(p.h, cn, ud)
	return func() {
		if p.h != 0 {
			cn := C.CString(name)
			defer C.free(unsafe.Pointer(cn))
			C.coreui_page_watch(p.h, cn, nil)
		}
		unregisterCallback(watchCallbacks, id)
		C.free(ud)
	}
}

// GetBool 读出布尔变量；不存在或类型不符时返回 false, error。
func (p *Page) GetBool(name string) (bool, error) {
	raw, err := p.GetJSON(name)
	if err != nil {
		return false, err
	}
	var v bool
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		return false, err
	}
	return v, nil
}

// GetInt 读出整数变量；不存在或类型不符时返回 0, error。
func (p *Page) GetInt(name string) (int, error) {
	raw, err := p.GetJSON(name)
	if err != nil {
		return 0, err
	}
	var v int
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		return 0, err
	}
	return v, nil
}

// GetFloat 读出浮点变量；不存在或类型不符时返回 0, error。
func (p *Page) GetFloat(name string) (float64, error) {
	raw, err := p.GetJSON(name)
	if err != nil {
		return 0, err
	}
	var v float64
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		return 0, err
	}
	return v, nil
}

// GetText 读出字符串变量；不存在或类型不符时返回 "", error。
func (p *Page) GetText(name string) (string, error) {
	raw, err := p.GetJSON(name)
	if err != nil {
		return "", err
	}
	var v string
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		return "", err
	}
	return v, nil
}
