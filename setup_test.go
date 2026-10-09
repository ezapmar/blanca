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

func TestWriteWaybar(t *testing.T) {
	dir := t.TempDir()
	cfg := "{\n  // bar\n  \"modules-right\": [\n    \"network\"\n  ],\n  \"network\": {}\n}\n"
	os.WriteFile(filepath.Join(dir, "config.jsonc"), []byte(cfg), 0o644)
	os.WriteFile(filepath.Join(dir, "style.css"), []byte("* { color: red; }\n"), 0o644)
	out := writeWaybar(dir, []byte("<svg/>"))
	if len(out) != 2 || !strings.HasPrefix(out[0], "add ") || !strings.HasPrefix(out[1], "add ") {
		t.Fatalf("first run: %v", out)
	}
	b, _ := os.ReadFile(filepath.Join(dir, "config.jsonc"))
	want := "{\n  // bar\n  \"custom/blanca\": { \"format\": \" \", \"on-click\": \"blanca menu\", \"tooltip-format\": \"Blanca clipboard\" },\n" +
		"  \"modules-right\": [\"custom/blanca\", \n    \"network\"\n  ],\n  \"network\": {}\n}\n"
	if string(b) != want {
		t.Fatalf("config:\n%s", b)
	}
	css, _ := os.ReadFile(filepath.Join(dir, "style.css"))
	if !strings.HasPrefix(string(css), "* { color: red; }\n") || !strings.Contains(string(css), "#custom-blanca") {
		t.Fatalf("style not appended: %s", css)
	}
	if icon, _ := os.ReadFile(filepath.Join(dir, "blanca-symbolic.svg")); string(icon) != "<svg/>" {
		t.Fatalf("icon not written: %s", icon)
	}
	// Second run must change nothing; a config without modules-right is left alone.
	if out = writeWaybar(dir, nil); !strings.HasPrefix(out[0], "ok ") || !strings.HasPrefix(out[1], "ok ") {
		t.Fatalf("second run must be a no-op: %v", out)
	}
	os.WriteFile(filepath.Join(dir, "config.jsonc"), []byte("{}\n"), 0o644)
	os.Remove(filepath.Join(dir, "style.css"))
	if out = writeWaybar(dir, nil); !strings.HasPrefix(out[0], "skip ") || !strings.HasPrefix(out[1], "skip ") {
		t.Fatalf("unknown config: %v", out)
	}
}

func TestSetAutostart(t *testing.T) {
	hypr := t.TempDir()
	p := filepath.Join(hypr, "autostart.lua")
	os.WriteFile(p, []byte("-- user config\no.launch_on_start(\"mako\")\n"), 0o644)
	if autostarts(hypr) {
		t.Fatal("not set up yet")
	}
	writeHyprSnippets(hypr)
	set, _ := os.ReadFile(p)
	if !autostarts(hypr) {
		t.Fatalf("setup must turn it on: %s", set)
	}
	if err := setAutostart(hypr, false); err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(p); string(b) != "-- user config\no.launch_on_start(\"mako\")\n" {
		t.Fatalf("off must leave only the user's lines: %q", b)
	}
	setAutostart(hypr, true)
	if b, _ := os.ReadFile(p); string(b) != string(set) || !autostarts(hypr) {
		t.Fatalf("on must restore what setup wrote: %q", b)
	}
	if setAutostart(t.TempDir(), true) == nil {
		t.Fatal("no autostart.lua is an error")
	}
}
