package main

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

// cmdStore is run by `wl-paste --type text --watch blanca store` with the clipping on stdin.
// It applies Jumpcut's filters: whitespace, size, sensitive source, and self-copies.
func cmdStore(cfg Config) {
	b, err := io.ReadAll(os.Stdin)
	if err != nil || len(b) == 0 || !utf8.Valid(b) {
		return
	}
	s := string(b)
	if !cfg.AllowWhitespace && strings.TrimSpace(s) == "" {
		return
	}
	if cfg.IgnoreLarge && utf8.RuneCountInString(s) > 50000 {
		return
	}
	if cfg.IgnoreSensitive && sensitive() {
		return
	}
	// Blanca's own wl-copy leaves a marker; consume it once and skip the echo.
	marker := filepath.Join(runtimeDir(), "placed")
	if m, err := os.ReadFile(marker); err == nil {
		os.Remove(marker)
		if string(m) == s {
			return
		}
	}
	if _, err := withHistory(func(h *History) bool { return h.Add(s, cfg.Remember) }); err != nil {
		fatal(err)
	}
}

// sensitive reports whether the current clipboard carries the password-manager hint
// (KeePassXC and friends), the Wayland counterpart of Jumpcut's ConcealedType check.
func sensitive() bool {
	out, err := exec.Command("wl-paste", "--list-types").Output()
	return err == nil && strings.Contains(string(out), "x-kde-passwordManagerHint")
}
