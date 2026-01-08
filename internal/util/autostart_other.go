//go:build !windows

package util

// EnsureAutostart is a no-op on non-Windows platforms.
func EnsureAutostart(appName string, enabled bool) error {
	return nil
}
