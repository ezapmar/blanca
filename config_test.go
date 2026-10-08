package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigDefaults(t *testing.T) {
	isolate(t)
	c := loadConfig()
	want := Config{Remember: 99, Display: 10, Paste: true, PasteMode: "shift-insert",
		IgnoreLarge: true, IgnoreSensitive: true}
	if c != want {
		t.Fatalf("defaults: %+v", c)
	}
}

func TestConfigOverrides(t *testing.T) {
	isolate(t)
	dir := filepath.Join(os.Getenv("XDG_CONFIG_HOME"), "blanca")
	os.MkdirAll(dir, 0o755)
	os.WriteFile(filepath.Join(dir, "config.json"),
		[]byte(`{"remember": 3, "paste": false, "paste_mode": "ctrl-v", "wraparound": true}`), 0o600)
	c := loadConfig()
	if c.Remember != 10 {
		t.Errorf("remember below 10 must clamp to 10, got %d", c.Remember)
	}
	if c.Paste || c.PasteMode != "ctrl-v" || !c.Wraparound {
		t.Errorf("overrides not applied: %+v", c)
	}
	if c.Display != 10 || !c.IgnoreLarge {
		t.Errorf("untouched keys must keep defaults: %+v", c)
	}
}

func TestShortenAndPaths(t *testing.T) {
	isolate(t)
	if got := shorten("  hello\nworld", 40); got != "hello" {
		t.Error(got)
	}
	if got := shorten("ééééé", 3); got != "ééé…" {
		t.Error(got)
	}
	if filepath.Base(filepath.Dir(dataPath())) != "blanca" {
		t.Error(dataPath())
	}
	if _, err := os.Stat(runtimeDir()); err != nil {
		t.Error("runtimeDir must be created:", err)
	}
}
