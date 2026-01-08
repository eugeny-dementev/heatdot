//go:build windows

package util

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"golang.org/x/sys/windows/registry"
)

const (
	runKeyPath         = `Software\Microsoft\Windows\CurrentVersion\Run`
	autostartDelay     = "0000:10"
	installTaskArg     = "-install-autostart-task"
	removeTaskArg      = "-remove-autostart-task"
	schtasksNotFound   = "cannot find the file specified"
	schtasksNoTask     = "the system cannot find the file specified"
	schtasksNoTaskAlt  = "the specified task name does not exist"
	schtasksNoTaskAlt2 = "cannot find the specified task"
	schtasksDenied     = "access is denied"
)

// EnsureAutostart sets or clears the Windows Run key for the current user.
func EnsureAutostart(appName string, enabled bool) error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve executable: %w", err)
	}
	return EnsureAutostartWithPath(appName, enabled, exePath)
}

// EnsureAutostartWithPath sets or clears autostart for the current user.
func EnsureAutostartWithPath(appName string, enabled bool, exePath string) error {
	return ensureAutostartWithPath(appName, enabled, exePath, true)
}

// EnsureAutostartWithPathNoElevate sets or clears autostart without elevation.
func EnsureAutostartWithPathNoElevate(appName string, enabled bool, exePath string) error {
	return ensureAutostartWithPath(appName, enabled, exePath, false)
}

func ensureAutostartWithPath(appName string, enabled bool, exePath string, allowElevate bool) error {
	if err := deleteRunKeyValue(appName); err != nil {
		return err
	}
	output, err := ensureScheduledTask(appName, enabled, exePath)
	if err != nil && allowElevate && isAccessDenied(output) {
		if elevErr := requestElevationForTask(appName, enabled, exePath); elevErr != nil {
			return fmt.Errorf("autostart task failed: %w (%s); elevation failed: %v", err, strings.TrimSpace(output), elevErr)
		}
		return nil
	}
	if err != nil {
		return fmt.Errorf("autostart task failed: %w (%s)", err, strings.TrimSpace(output))
	}
	return nil
}

func deleteRunKeyValue(appName string) error {
	key, err := registry.OpenKey(registry.CURRENT_USER, runKeyPath, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("open Run key: %w", err)
	}
	defer key.Close()

	if err := key.DeleteValue(appName); err != nil && err != registry.ErrNotExist {
		return fmt.Errorf("delete Run value: %w", err)
	}
	return nil
}

func ensureScheduledTask(appName string, enabled bool, exePath string) (string, error) {
	if !enabled {
		output, err := runSchtasks("/Delete", "/TN", appName, "/F")
		if err != nil && !isTaskMissing(output) {
			return output, err
		}
		return output, nil
	}

	queryOutput, queryErr := runSchtasks("/Query", "/TN", appName, "/V", "/FO", "LIST")
	if queryErr == nil && taskMatches(queryOutput, exePath) {
		return queryOutput, nil
	}
	if queryErr != nil && !isTaskMissing(queryOutput) {
		return queryOutput, queryErr
	}
	if queryErr == nil {
		changeOutput, changeErr := runSchtasks("/Change", "/TN", appName, "/TR", fmt.Sprintf("%q", exePath), "/F")
		if changeErr == nil {
			return changeOutput, nil
		}
		return changeOutput, changeErr
	}

	taskRun := fmt.Sprintf("%q", exePath)
	output, err := runSchtasks(
		"/Create",
		"/TN", appName,
		"/TR", taskRun,
		"/SC", "ONLOGON",
		"/DELAY", autostartDelay,
		"/RL", "LIMITED",
		"/F",
	)
	if err != nil {
		return output, err
	}
	return output, nil
}

func runSchtasks(args ...string) (string, error) {
	cmd := exec.Command("schtasks", args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func isTaskMissing(output string) bool {
	normalized := strings.ToLower(output)
	return strings.Contains(normalized, schtasksNotFound) ||
		strings.Contains(normalized, schtasksNoTask) ||
		strings.Contains(normalized, schtasksNoTaskAlt) ||
		strings.Contains(normalized, schtasksNoTaskAlt2)
}

func isAccessDenied(output string) bool {
	return strings.Contains(strings.ToLower(output), schtasksDenied)
}

func requestElevationForTask(appName string, enabled bool, exePath string) error {
	taskArg := installTaskArg
	if !enabled {
		taskArg = removeTaskArg
	}
	command := fmt.Sprintf("Start-Process -FilePath %q -ArgumentList %q -Verb RunAs -WindowStyle Hidden", exePath, taskArg)
	cmd := exec.Command("powershell", "-NoProfile", "-WindowStyle", "Hidden", "-Command", command)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd.Run()
}

func taskMatches(output string, exePath string) bool {
	normalizedExe := strings.ToLower(filepath.Clean(exePath))
	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(strings.ToLower(trimmed), "task to run:") {
			continue
		}
		value := strings.TrimSpace(trimmed[len("task to run:"):])
		value = strings.Trim(value, "\"")
		value = strings.TrimSpace(value)
		value = strings.Trim(value, "\"")
		if strings.EqualFold(filepath.Clean(value), normalizedExe) {
			return true
		}
	}
	return strings.Contains(strings.ToLower(output), normalizedExe)
}
