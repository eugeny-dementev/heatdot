//go:build !windows

package util

// EnsureAutostart is a no-op on non-Windows platforms.
func EnsureAutostart(appName string, enabled bool) error {
	return nil
}

// EnsureAutostartWithPath is a no-op on non-Windows platforms.
func EnsureAutostartWithPath(appName string, enabled bool, exePath string) error {
	return nil
}

// EnsureAutostartWithPathNoElevate is a no-op on non-Windows platforms.
func EnsureAutostartWithPathNoElevate(appName string, enabled bool, exePath string) error {
	return nil
}
