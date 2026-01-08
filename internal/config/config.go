package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

var ErrConfigNotFound = errors.New("config not found")

// Config holds Heatdot settings.
type Config struct {
	ProfileURL     string `json:"profile_url"`
	RefreshMinutes int    `json:"refresh_minutes"`
	Timezone       string `json:"timezone"`
	Debug          bool   `json:"debug"`
}

func Default() Config {
	return Config{
		ProfileURL:     "",
		RefreshMinutes: 10,
		Timezone:       "local",
		Debug:          false,
	}
}

func ConfigPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "heatdot", "config.json"), nil
}

func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Default(), ErrConfigNotFound
		}
		return Default(), err
	}

	cfg := Default()
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Default(), err
	}
	applyDefaults(&cfg)
	return cfg, nil
}

func Save(path string, cfg Config) error {
	applyDefaults(&cfg)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}

func EnsureTemplate(path string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}

	cfg := Default()
	cfg.ProfileURL = "https://github.com/<username>"
	return Save(path, cfg)
}

func applyDefaults(cfg *Config) {
	if cfg.RefreshMinutes <= 0 {
		cfg.RefreshMinutes = Default().RefreshMinutes
	}
	if cfg.Timezone == "" {
		cfg.Timezone = Default().Timezone
	}
}
