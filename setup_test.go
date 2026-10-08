package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteHyprSnippets(t *testing.T) {
	hypr := t.TempDir()
	for _, f := range []string{"bindings.lua", "autostart.lua"} {
		os.WriteFile(filepath.Join(hypr, f), []byte("-- user config\n"), 0o644)
	}
	out := writeHyprSnippets(hypr)
	if len(out) != 3 || !strings.HasPrefix(out[0], "add ") || !strings.HasPrefix(out[1], "add ") || !strings.HasPrefix(out[2], "skip ") {
		t.Fatalf("first run: %v", out)
	}
	b, _ := os.ReadFile(filepath.Join(hypr, "bindings.lua"))
	if !strings.HasPrefix(string(b), "-- user config\n") || !strings.Contains(string(b), `"blanca bezel"`) {
		t.Fatalf("bindings not appended: %s", b)
	}
	b, _ = os.ReadFile(filepath.Join(hypr, "autostart.lua"))
	if !strings.Contains(string(b), "wl-paste --type text --watch blanca store") {
		t.Fatalf("autostart not appended: %s", b)
	}
	// Second run must change nothing.
	before, _ := os.ReadFile(filepath.Join(hypr, "bindings.lua"))
	out = writeHyprSnippets(hypr)
	after, _ := os.ReadFile(filepath.Join(hypr, "bindings.lua"))
	if !strings.HasPrefix(out[0], "ok ") || string(before) != string(after) {
		t.Fatalf("second run must be a no-op: %v", out)
	}
}
