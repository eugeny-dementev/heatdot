//go:build windows

package util

import "syscall"

// HideConsole detaches from the console window on Windows.
func HideConsole() {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	freeConsole := kernel32.NewProc("FreeConsole")
	_, _, _ = freeConsole.Call()
}
