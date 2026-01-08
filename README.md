# Heatdot

Heatdot is a cross-platform tray app that shows today's GitHub contributions for a configured profile and colors the tray dot like the GitHub heatmap.

## Features
- Tray tooltip: "Heatdot — Today: N"
- Heatmap-colored tray dot based on today's count
- Auto refresh (default every 10 minutes) plus "Refresh now"
- No API tokens or OAuth

## Config
Heatdot reads a JSON config file from the OS config directory:
- Windows: %APPDATA%/heatdot/config.json
- macOS: ~/Library/Application Support/heatdot/config.json
- Linux: ~/.config/heatdot/config.json

Example:
```json
{
  "profile_url": "https://github.com/<username>",
  "refresh_minutes": 10,
  "timezone": "local",
  "autostart": false
}
```

Notes:
- `timezone` can be `local` or any IANA timezone (e.g., `America/Los_Angeles`).
- If the config is missing, use the tray menu "Settings..." to generate a template.
- Optional: `debug` (true/false) enables verbose logging.
- Optional: `autostart` (true/false) enables Windows startup via Task Scheduler (10s delay after logon).

## Build
```bash
go build ./cmd/heatdot
```

### Cross-compile examples
```bash
# Windows
env GOOS=windows GOARCH=amd64 go build -ldflags "-H=windowsgui" -o heatdot.exe ./cmd/heatdot

# macOS
env GOOS=darwin GOARCH=arm64 go build -o heatdot ./cmd/heatdot

# Linux
env GOOS=linux GOARCH=amd64 go build -o heatdot ./cmd/heatdot
```

## Tray dependencies
- Linux: `systray` may require AppIndicator libraries (for example `libayatana-appindicator3-dev` on Debian/Ubuntu).
- macOS/Windows: standard system tray support.

## Run
```bash
./heatdot
```

## Debugging / Logs
Run with logging enabled:
```bash
./heatdot -debug
```

By default logs are written to `heatdot.log` in the same directory as `config.json`.

Optional custom log file:
```bash
./heatdot -debug -log-file /path/to/heatdot.log
```

On Windows, you can tail the log with:
```powershell
Get-Content -Path $env:APPDATA\\heatdot\\heatdot.log -Wait
```
