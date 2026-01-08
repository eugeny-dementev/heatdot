//go:build windows

package util

import (
	"fmt"
	"os"

	"golang.org/x/sys/windows/registry"
)

const runKeyPath = `Software\Microsoft\Windows\CurrentVersion\Run`

// EnsureAutostart sets or clears the Windows Run key for the current user.
func EnsureAutostart(appName string, enabled bool) error {
	key, err := registry.OpenKey(registry.CURRENT_USER, runKeyPath, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("open Run key: %w", err)
	}
	defer key.Close()

	if !enabled {
		if err := key.DeleteValue(appName); err != nil && err != registry.ErrNotExist {
			return fmt.Errorf("delete Run value: %w", err)
		}
		return nil
	}

	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve executable: %w", err)
	}
	value := fmt.Sprintf("%q", exePath)
	if err := key.SetStringValue(appName, value); err != nil {
		return fmt.Errorf("set Run value: %w", err)
	}
	return nil
}
