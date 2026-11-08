//go:build windows

package ui

import (
	"syscall"
	"unsafe"
)

var (
	consoleUser32 = syscall.NewLazyDLL("user32.dll")
	consoleKernel = syscall.NewLazyDLL("kernel32.dll")

	procGetConsoleWindow      = consoleKernel.NewProc("GetConsoleWindow")
	procGetConsoleProcessList = consoleKernel.NewProc("GetConsoleProcessList")
	procShowWindow            = consoleUser32.NewProc("ShowWindow")
)

// hideOwnedConsole hides a console window that Windows created *for this
// process* (the usual double-click scenario). If the console is inherited from
// an existing terminal, GetConsoleProcessList reports more than one process and
// we leave it alone, so terminal users still see their shell and logs.
func hideOwnedConsole() {
	hwnd, _, _ := procGetConsoleWindow.Call()
	if hwnd == 0 {
		return
	}
	var pids [2]uint32
	count, _, _ := procGetConsoleProcessList.Call(
		uintptr(unsafe.Pointer(&pids[0])), uintptr(len(pids)))
	if count == 0 || count > 1 {
		return
	}
	const swHide = 0
	procShowWindow.Call(hwnd, swHide)
}
