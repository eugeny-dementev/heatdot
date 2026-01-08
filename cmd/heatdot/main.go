package main

import (
	"flag"
	"io"
	"log"
	"os"
	"path/filepath"

	"heatdot/internal/app"
	"heatdot/internal/config"
)

func main() {
	logFile := flag.String("log-file", "", "Path to log file")
	debug := flag.Bool("debug", false, "Enable debug logging")
	flag.Parse()

	logger := log.New(os.Stdout, "", log.LstdFlags)
	if *debug && *logFile == "" {
		if path, err := config.ConfigPath(); err == nil {
			*logFile = filepath.Join(filepath.Dir(path), "heatdot.log")
		}
	}

	if *logFile != "" {
		if err := os.MkdirAll(filepath.Dir(*logFile), 0o755); err != nil {
			logger.Printf("failed to create log dir for %s: %v", *logFile, err)
		} else if file, err := os.OpenFile(*logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644); err != nil {
			logger.Printf("failed to open log file %s: %v", *logFile, err)
		} else {
			logger.SetOutput(io.MultiWriter(os.Stdout, file))
		}
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
