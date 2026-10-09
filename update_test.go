package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewer(t *testing.T) {
	for _, c := range []struct {
		tag, current string
		want         bool
	}{
		{"v0.2.1", "0.2.0", true},
		{"v0.10.0", "0.9.9", true},
		{"v1.0", "0.9.9", true},
		{"v0.2.0", "0.2.0", false},
		{"v0.2", "0.2.0", false},
		{"v0.2.0", "0.3.0", false},
		{"v0.2.0", "dev", true},
		{"nightly", "0.2.0", false},
	} {
		if got := newer(c.tag, c.current); got != c.want {
			t.Errorf("newer(%q, %q) = %v, want %v", c.tag, c.current, got, c.want)
		}
	}
}

func TestLatestTag(t *testing.T) {
	body := `{"tag_name": "v0.3.1", "name": "v0.3.1"}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(body))
	}))
	defer srv.Close()
	old := latestURL
	latestURL = srv.URL
	defer func() { latestURL = old }()

	if tag, err := latestTag(); err != nil || tag != "v0.3.1" {
		t.Fatalf("latestTag = %q, %v", tag, err)
	}
	// A tag that is not a version must never reach the shell.
	body = `{"tag_name": "v1; rm -rf ~"}`
	if tag, err := latestTag(); err == nil {
		t.Fatalf("odd tag accepted: %q", tag)
	}
}

func TestUpdateCommand(t *testing.T) {
	got := updateCommand("v0.3.1", "/usr/bin/blanca", "linux", "arm64")
	if got != "sudo pacman -U https://github.com/ezapmar/blanca/releases/download/v0.3.1/blanca-bin-0.3.1-1-aarch64.pkg.tar.zst" {
		t.Errorf("pacman install: %s", got)
	}
	for _, self := range []string{"/home/me/.local/bin/blanca", "/Applications/Blanca.app/Contents/MacOS/blanca"} {
		goos := "linux"
		if strings.Contains(self, ".app/") {
			goos = "darwin"
		}
		got := updateCommand("v0.3.1", self, goos, "amd64")
		if got != "curl -fsSL https://raw.githubusercontent.com/ezapmar/blanca/v0.3.1/install.sh | sh -s v0.3.1" {
			t.Errorf("%s: %s", self, got)
		}
	}
}
