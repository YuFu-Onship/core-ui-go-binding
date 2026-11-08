package coreui

/*
#include <ui_core.h>
#include <stdlib.h>
*/
import "C"

import (
	"unicode/utf16"
	"unsafe"
)

// cWide 把 Go 字符串转成 NUL 结尾的 wchar_t 缓冲（C 堆分配），
// 调用方必须用 freeWide 释放。
func cWide(s string) *C.wchar_t {
	units := utf16.Encode([]rune(s))
	size := int(unsafe.Sizeof(C.wchar_t(0)))
	p := C.calloc(C.size_t(len(units)+1), C.size_t(size))
	if p == nil {
		panic("coreui: 内存不足")
	}
	if size == 2 {
		dst := unsafe.Slice((*uint16)(p), len(units)+1)
		copy(dst, units)
	} else {
		dst := unsafe.Slice((*uint32)(p), len(units)+1)
		for i, u := range units {
			dst[i] = uint32(u)
		}
	}
	return (*C.wchar_t)(p)
}

// freeWide 释放 cWide 分配的内存。
func freeWide(p *C.wchar_t) {
	if p != nil {
		C.free(unsafe.Pointer(p))
	}
}

// wideToGo 把 NUL 结尾的 wchar_t 串转成 Go 字符串。
func wideToGo(p *C.wchar_t) string {
	if p == nil {
		return ""
	}
	if unsafe.Sizeof(C.wchar_t(0)) == 2 {
		src := unsafe.Slice((*uint16)(unsafe.Pointer(p)), 1<<20)
		n := 0
		for n < len(src) && src[n] != 0 {
			n++
		}
		return string(utf16.Decode(src[:n]))
	}
	src := unsafe.Slice((*uint32)(unsafe.Pointer(p)), 1<<20)
	n := 0
	for n < len(src) && src[n] != 0 {
		n++
	}
	units := make([]uint16, 0, n)
	for _, r := range src[:n] {
		units = utf16.AppendRune(units, rune(r))
	}
	return string(utf16.Decode(units))
}
