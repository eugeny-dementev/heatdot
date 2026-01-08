package app

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/getlantern/systray"

	"heatdot/internal/config"
	"heatdot/internal/github"
	"heatdot/internal/icon"
	"heatdot/internal/util"
)

type App struct {
	logger *log.Logger

	debugOverrideSet bool
	debugOverride    bool

	cfgPath string
	cfgMu   sync.RWMutex
	cfg     config.Config

	menuToday    *systray.MenuItem
	menuRefresh  *systray.MenuItem
	menuProfile  *systray.MenuItem
	menuSettings *systray.MenuItem
	menuQuit     *systray.MenuItem

	refreshCh chan struct{}
	stopCh    chan struct{}

	refreshMu  sync.Mutex
	refreshing bool

	intervalMu      sync.Mutex
	refreshInterval time.Duration

	autostartMu      sync.Mutex
	autostartSet     bool
	autostartEnabled bool
}

type Options struct {
	Logger *log.Logger
	Debug  *bool
}

func Run() {
	RunWithOptions(Options{})
}

func RunWithOptions(opts Options) {
	app := newApp(opts)
	systray.Run(app.onReady, app.onExit)
}

func newApp(opts Options) *App {
	logger := opts.Logger
	if logger == nil {
		logger = log.Default()
	}

	app := &App{
		logger:          logger,
		refreshCh:       make(chan struct{}, 1),
		stopCh:          make(chan struct{}),
		refreshInterval: 10 * time.Minute,
	}
	if opts.Debug != nil {
		app.debugOverrideSet = true
		app.debugOverride = *opts.Debug
	}
	return app
}

func (a *App) onReady() {
	a.cfgPath = configPathFallback(a.logger)
	a.setIconError()
	systray.SetTitle("Heatdot")
	systray.SetTooltip("Heatdot — starting...")

	a.menuToday = systray.AddMenuItem("Today: --", "Today's contributions")
	a.menuToday.Disable()
	systray.AddSeparator()
	a.menuRefresh = systray.AddMenuItem("Refresh now", "Fetch latest contributions")
	a.menuProfile = systray.AddMenuItem("Open GitHub profile", "Open configured profile")
	a.menuSettings = systray.AddMenuItem("Settings...", "Open config file")

	a.menuQuit = systray.AddMenuItem("Quit", "Quit Heatdot")

	go a.menuLoop()
	go a.refreshLoop()
}

func (a *App) onExit() {
	close(a.stopCh)
}

func (a *App) menuLoop() {
	for {
		select {
		case <-a.menuRefresh.ClickedCh:
			a.requestRefresh()
		case <-a.menuProfile.ClickedCh:
			a.openProfile()
		case <-a.menuSettings.ClickedCh:
			a.openSettings()
		case <-a.menuQuit.ClickedCh:
			systray.Quit()
			return
		}
	}
}

func (a *App) refreshLoop() {
	timer := time.NewTimer(0)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			a.refreshOnce()
			timer.Reset(a.getInterval())
		case <-a.refreshCh:
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			a.refreshOnce()
			timer.Reset(a.getInterval())
		case <-a.stopCh:
			return
		}
	}
}

func (a *App) requestRefresh() {
	select {
	case a.refreshCh <- struct{}{}:
	default:
	}
}

func (a *App) refreshOnce() {
	if !a.beginRefresh() {
		return
	}
	defer a.endRefresh()

	cfg, err := config.Load(a.cfgPath)
	if err != nil {
		if err == config.ErrConfigNotFound {
			a.setNeedsConfig()
			return
		}
		a.setErrorState(err)
		return
	}
	a.setConfig(cfg)
	if strings.TrimSpace(cfg.ProfileURL) == "" {
		a.setNeedsConfig()
		return
	}
	a.debugf("fetching profile: %s", cfg.ProfileURL)

	loc := time.Local
	if cfg.Timezone != "" && strings.ToLower(cfg.Timezone) != "local" {
		loaded, err := time.LoadLocation(cfg.Timezone)
		if err != nil {
			a.logger.Printf("invalid timezone %q, using local: %v", cfg.Timezone, err)
		} else {
			loc = loaded
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	html, err := github.FetchProfileHTML(ctx, cfg.ProfileURL)
	if err != nil {
		a.setErrorState(err)
		return
	}

	today := time.Now().In(loc)
	count, err := github.ParseCountForDate(html, today)
	if err != nil {
		a.setErrorState(err)
		return
	}

	a.debugf("parsed today's count: %d", count)
	a.setSuccessState(count)
}

func (a *App) beginRefresh() bool {
	a.refreshMu.Lock()
	defer a.refreshMu.Unlock()
	if a.refreshing {
		return false
	}
	a.refreshing = true
	return true
}

func (a *App) endRefresh() {
	a.refreshMu.Lock()
	a.refreshing = false
	a.refreshMu.Unlock()
}

func (a *App) setConfig(cfg config.Config) {
	a.cfgMu.Lock()
	a.cfg = cfg
	a.cfgMu.Unlock()

	interval := time.Duration(cfg.RefreshMinutes) * time.Minute
	if interval <= 0 {
		interval = 10 * time.Minute
	}
	a.setInterval(interval)
	a.ensureAutostart(cfg.Autostart)
}

func (a *App) setInterval(interval time.Duration) {
	a.intervalMu.Lock()
	a.refreshInterval = interval
	a.intervalMu.Unlock()
}

func (a *App) getInterval() time.Duration {
	a.intervalMu.Lock()
	defer a.intervalMu.Unlock()
	return a.refreshInterval
}

func (a *App) ensureAutostart(enabled bool) {
	if runtime.GOOS != "windows" {
		return
	}

	a.autostartMu.Lock()
	if a.autostartSet && a.autostartEnabled == enabled {
		a.autostartMu.Unlock()
		return
	}
	a.autostartSet = true
	a.autostartEnabled = enabled
	a.autostartMu.Unlock()

	if err := util.EnsureAutostart("Heatdot", enabled); err != nil {
		a.logger.Printf("autostart update failed: %v", err)
	}
}

func (a *App) setSuccessState(count int) {
	title := fmt.Sprintf("Heatdot — Today: %d", count)
	systray.SetTitle(title)
	systray.SetTooltip(title)
	a.menuToday.SetTitle(fmt.Sprintf("Today: %d", count))
	a.menuProfile.Enable()

	iconBytes, err := icon.ForCount(count)
	if err != nil {
		a.logger.Printf("icon error: %v", err)
		return
	}
	systray.SetIcon(iconBytes)
}

func (a *App) setErrorState(err error) {
	a.logger.Printf("refresh error: %v", err)
	systray.SetTitle("Heatdot — error")
	systray.SetTooltip("Heatdot — error")
	a.menuToday.SetTitle("Today: --")
	a.menuProfile.Disable()
	a.setIconError()
}

func (a *App) setNeedsConfig() {
	systray.SetTitle("Heatdot")
	systray.SetTooltip("Heatdot: configure profile_url")
	a.menuToday.SetTitle("Today: --")
	a.menuProfile.Disable()
	a.setIconError()
}

func (a *App) setIconError() {
	iconBytes, err := icon.ForError()
	if err != nil {
		a.logger.Printf("icon error: %v", err)
		return
	}
	systray.SetIcon(iconBytes)
}

func (a *App) openProfile() {
	url := a.currentProfileURL()
	if strings.TrimSpace(url) == "" {
		a.openSettings()
		return
	}
	if err := util.Open(url); err != nil {
		a.logger.Printf("open profile failed: %v", err)
	}
}

func (a *App) openSettings() {
	if err := config.EnsureTemplate(a.cfgPath); err != nil {
		a.logger.Printf("settings template failed: %v", err)
		return
	}
	if err := util.Open(a.cfgPath); err != nil {
		a.logger.Printf("open settings failed: %v", err)
	}
	a.requestRefresh()
}

func (a *App) currentProfileURL() string {
	a.cfgMu.RLock()
	defer a.cfgMu.RUnlock()
	return a.cfg.ProfileURL
}

func (a *App) debugf(format string, args ...any) {
	if !a.debugEnabled() {
		return
	}
	a.logger.Printf(format, args...)
}

func (a *App) debugEnabled() bool {
	if a.debugOverrideSet {
		return a.debugOverride
	}
	a.cfgMu.RLock()
	defer a.cfgMu.RUnlock()
	return a.cfg.Debug
}

func configPathFallback(logger *log.Logger) string {
	path, err := config.ConfigPath()
	if err == nil {
		return path
	}
	logger.Printf("failed to resolve config dir, using local config.json: %v", err)
	return "config.json"
}
