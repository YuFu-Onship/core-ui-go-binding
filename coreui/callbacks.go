package coreui

/*
#include <ui_core.h>
#include <stdlib.h>
#include "callbacks.h"
*/
import "C"

import (
	"fmt"
	"os"
	"sync"
	"unsafe"
)

// cgo 回调无法携带 Go 闭包，因此把闭包登记在表里，userdata 传登记号。
var (
	cbMu                  sync.RWMutex
	cbSeq                 uint64
	clickCallbacks        = map[uint64]func(){}
	closeCallbacks        = map[uint64]func(){}
	closeRequestCallbacks = map[uint64]func() bool{}
	watchCallbacks        = map[uint64]func(string){}
	invokeCallbacks       = map[uint64]func(){}
)

func registerCallback[T any](table map[uint64]T, fn T) uint64 {
	cbMu.Lock()
	defer cbMu.Unlock()
	cbSeq++
	table[cbSeq] = fn
	return cbSeq
}

func unregisterCallback[T any](table map[uint64]T, id uint64) {
	cbMu.Lock()
	delete(table, id)
	cbMu.Unlock()
}

func lookupCallback[T any](table map[uint64]T, userdata unsafe.Pointer) (T, bool) {
	var zero T
	if userdata == nil {
		return zero, false
	}
	id := *(*uint64)(userdata)
	cbMu.RLock()
	defer cbMu.RUnlock()
	fn, ok := table[id]
	return fn, ok
}

// callbackPtr 把登记号装进一块 C 内存传给 core-ui，作为 userdata 回传。
// 用真实指针而不是把整数强转成 unsafe.Pointer，避免触发 go vet 告警。
func callbackPtr(id uint64) unsafe.Pointer {
	p := C.malloc(C.size_t(8))
	if p == nil {
		panic("coreui: 内存不足")
	}
	*(*uint64)(p) = id
	return p
}

func guardCallback(what string) {
	if r := recover(); r != nil {
		fmt.Fprintf(os.Stderr, "coreui: %s 回调 panic: %v\n", what, r)
	}
}

//export goCoreUiClick
func goCoreUiClick(w C.UiWidget, userdata unsafe.Pointer) {
	defer guardCallback("click")
	if fn, ok := lookupCallback(clickCallbacks, userdata); ok && fn != nil {
		fn()
	}
}

//export goCoreUiWindowClose
func goCoreUiWindowClose(w C.UiWindow, userdata unsafe.Pointer) {
	defer guardCallback("close")
	if fn, ok := lookupCallback(closeCallbacks, userdata); ok && fn != nil {
		fn()
	}
}

//export goCoreUiWindowCloseRequest
func goCoreUiWindowCloseRequest(w C.UiWindow, userdata unsafe.Pointer) C.int {
	defer guardCallback("close-request")
	if fn, ok := lookupCallback(closeRequestCallbacks, userdata); ok && fn != nil {
		if fn() {
			return 1
		}
		return 0
	}
	return 1
}

//export goCoreUiPageWatch
func goCoreUiPageWatch(p C.UiPage, name *C.char, json *C.char, userdata unsafe.Pointer) {
	defer guardCallback("page-watch")
	if fn, ok := lookupCallback(watchCallbacks, userdata); ok && fn != nil {
		fn(C.GoString(json))
	}
}

//export goCoreUiInvoke
func goCoreUiInvoke(userdata unsafe.Pointer) {
	defer guardCallback("invoke")
	if fn, ok := lookupCallback(invokeCallbacks, userdata); ok && fn != nil {
		fn()
	}
}
