package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func isolate(t *testing.T) {
	t.Helper()
	d := t.TempDir()
	t.Setenv("XDG_DATA_HOME", d)
	t.Setenv("XDG_CONFIG_HOME", d)
	t.Setenv("XDG_RUNTIME_DIR", d)
}

func no() bool  { return false }
func yes() bool { return true }

func TestStoreFilters(t *testing.T) {
	isolate(t)
	cfg := loadConfig()
	cases := []struct {
		name string
		text string
		cfg  Config
		sens func() bool
		want bool
	}{
		{"plain", "hello", cfg, no, true},
		{"empty", "", cfg, no, false},
		{"whitespace", " \n\t", cfg, no, false},
		{"whitespace allowed", " \n\t", Config{AllowWhitespace: true, Remember: 10}, no, true},
		{"large", strings.Repeat("x", 50001), cfg, no, false},
		{"large allowed", strings.Repeat("x", 50001), Config{Remember: 10}, no, true},
		{"sensitive", "hunter2", cfg, yes, false},
		{"sensitive allowed", "hunter2", Config{Remember: 10}, yes, true},
	}
	for _, c := range cases {
		isolate(t)
		got, err := store(c.cfg, c.text, c.sens)
		if err != nil || got != c.want {
			t.Errorf("%s: added=%v err=%v, want %v", c.name, got, err, c.want)
		}
	}
}

func TestStoreDedupAndCap(t *testing.T) {
	isolate(t)
	cfg := Config{Remember: 10}
	for _, s := range []string{"a", "a", "b"} {
		store(cfg, s, no)
	}
	items, _ := readHistory()
	if len(items) != 2 || items[0] != "b" || items[1] != "a" {
		t.Fatalf("dedup: %v", items)
	}
	for i := 0; i < 20; i++ {
		store(cfg, strings.Repeat("z", i+1), no)
	}
	items, _ = readHistory()
	if len(items) != 10 || items[0] != strings.Repeat("z", 20) {
		t.Fatalf("cap: %d items, top %q", len(items), items[0])
	}
}

func TestStoreSelfCopyMarker(t *testing.T) {
	isolate(t)
	cfg := Config{Remember: 10}
	store(cfg, "old", no)
	marker := filepath.Join(runtimeDir(), "placed")
	os.WriteFile(marker, []byte("old"), 0o600)
	if added, _ := store(cfg, "old", no); added {
		t.Fatal("echo of our own copy must not be stored")
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatal("marker must be consumed")
	}
	// A different text consumes the marker too, but is stored.
	os.WriteFile(marker, []byte("old"), 0o600)
	if added, _ := store(cfg, "new", no); !added {
		t.Fatal("new text must be stored")
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatal("marker must be consumed even when text differs")
	}
}

func TestHistoryPersistsAndSurvivesGarbage(t *testing.T) {
	isolate(t)
	store(Config{Remember: 10}, "kept", no)
	if b, _ := os.ReadFile(dataPath()); !strings.Contains(string(b), `"kept"`) {
		t.Fatalf("not persisted as JSON: %s", b)
	}
	os.WriteFile(dataPath(), []byte("not json"), 0o600)
	items, err := readHistory()
	if err != nil || len(items) != 0 {
		t.Fatalf("corrupt file should read as empty, got %v %v", items, err)
	}
}
