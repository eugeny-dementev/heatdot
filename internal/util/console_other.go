//go:build !windows

package util

// HideConsole is a no-op on non-Windows platforms.
func HideConsole() {}
