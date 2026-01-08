package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultAutostartFalse(t *testing.T) {
	cfg := Default()
	if cfg.Autostart {
		t.Fatalf("expected autostart default to be false")
	}
}

func TestLoadAutostartTrue(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	data := []byte(`{"profile_url":"https://github.com/example","refresh_minutes":5,"timezone":"local","autostart":true}`)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.Autostart {
		t.Fatalf("expected autostart to be true")
	}
}
