package main

import (
	"flag"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"heatdot/internal/app"
	"heatdot/internal/config"
	"heatdot/internal/util"
)

func main() {
	util.HideConsole()

	logFile := flag.String("log-file", "", "Path to log file")
	debug := flag.Bool("debug", false, "Enable debug logging")
	installAutostart := flag.Bool("install-autostart-task", false, "Install autostart task and exit")
	removeAutostart := flag.Bool("remove-autostart-task", false, "Remove autostart task and exit")
	flag.Parse()

	logger := log.New(os.Stdout, "", log.LstdFlags)

	logPath := *logFile
	if logPath == "" {
		if cfgPath, err := config.ConfigPath(); err == nil {
			logPath = filepath.Join(filepath.Dir(cfgPath), "heatdot.log")
		} else {
			logger.Printf("failed to resolve config dir: %v", err)
		}
	}

	if logPath != "" {
		if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
			logger.Printf("failed to create log dir for %s: %v", logPath, err)
		} else if file, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644); err != nil {
			logger.Printf("failed to open log file %s: %v", logPath, err)
		} else {
			if runtime.GOOS == "windows" {
				logger.SetOutput(file)
			} else {
				logger.SetOutput(io.MultiWriter(os.Stdout, file))
			}
		}
	}

	if *installAutostart || *removeAutostart {
		exePath, err := os.Executable()
		if err != nil {
			logger.Printf("failed to resolve executable: %v", err)
			os.Exit(1)
		}
		enabled := *installAutostart && !*removeAutostart
		if err := util.EnsureAutostartWithPathNoElevate("Heatdot", enabled, exePath); err != nil {
			logger.Printf("autostart task update failed: %v", err)
			os.Exit(1)
		}
		return
	}

	var debugOverride *bool
	if *debug {
		debugOverride = debug
	}

	app.RunWithOptions(app.Options{
		Logger: logger,
		Debug:  debugOverride,
	})
}
