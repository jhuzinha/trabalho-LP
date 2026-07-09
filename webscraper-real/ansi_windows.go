//go:build windows

package main

import (
	"os"
	"syscall"
	"unsafe"
)

// enableWindowsANSI habilita cores ANSI no console do Windows.
func enableWindowsANSI() {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getMode := kernel32.NewProc("GetConsoleMode")
	setMode := kernel32.NewProc("SetConsoleMode")

	handle := syscall.Handle(os.Stdout.Fd())
	var mode uint32
	getMode.Call(uintptr(handle), uintptr(unsafe.Pointer(&mode)))
	setMode.Call(uintptr(handle), uintptr(mode|0x0004))
}
