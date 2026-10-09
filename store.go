package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"
)

// cmdStore is run by `wl-paste --type text --watch blanca store` with the clipping on stdin.
// It applies the filters: whitespace, size, sensitive source, and self-copies.
func cmdStore(cfg Config) {
	b, err := io.ReadAll(os.Stdin)
	if err != nil || !utf8.Valid(b) {
		return
	}
	if _, err := store(cfg, string(b), sensitive); err != nil {
		fatal(err)
	}
}

// store applies the filters and records s. It reports whether s was added.
func store(cfg Config, s string, sensitive func() bool) (bool, error) {
	if s == "" {
		return false, nil
	}
	if !cfg.AllowWhitespace && strings.TrimSpace(s) == "" {
		return false, nil
	}
	if cfg.IgnoreLarge && utf8.RuneCountInString(s) > 50000 {
		return false, nil
	}
	if cfg.IgnoreSensitive && sensitive() {
		return false, nil
	}
	// Blanca's own wl-copy leaves a marker; consume it once and skip the echo.
	marker := filepath.Join(runtimeDir(), "placed")
	if m, err := os.ReadFile(marker); err == nil {
		os.Remove(marker)
		if string(m) == s {
			return false, nil
		}
	}
	added := false
	_, err := withHistory(func(h *History) bool { added = h.Add(s, cfg.Remember); return added })
	if err == nil && added {
		err = setTime(s, time.Now())
	}
	return added, err
}
