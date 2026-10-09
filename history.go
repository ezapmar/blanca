package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

// History is the clipping list, newest first. It is a JSON array of strings on disk.
type History struct{ Items []string }

// withHistory runs fn on the locked, freshly loaded history and saves it if fn returns true.
// It returns the resulting items so callers can refresh their view.
func withHistory(fn func(h *History) bool) ([]string, error) {
	path := dataPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	lock, err := os.OpenFile(path+".lock", os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	defer lock.Close() // releases the flock
	if err := syscall.Flock(int(lock.Fd()), syscall.LOCK_EX); err != nil {
		return nil, err
	}
	h := &History{}
	if b, err := os.ReadFile(path); err == nil {
		json.Unmarshal(b, &h.Items)
	}
	if fn(h) {
		b, _ := json.MarshalIndent(h.Items, "", " ")
		if err := os.WriteFile(path+".tmp", b, 0o600); err != nil {
			return nil, err
		}
		if err := os.Rename(path+".tmp", path); err != nil {
			return nil, err
		}
	}
	return h.Items, nil
}

func readHistory() ([]string, error) {
	return withHistory(func(*History) bool { return false })
}

// Add puts s on top unless it already is; keeps at most max items.
func (h *History) Add(s string, max int) bool {
	if len(h.Items) > 0 && h.Items[0] == s {
		return false
	}
	h.Items = append([]string{s}, h.Items...)
	if len(h.Items) > max {
		h.Items = h.Items[:max]
	}
	return true
}

// find locates s, trying index hint first; the watcher may have shifted things meanwhile.
func (h *History) find(s string, hint int) int {
	if hint >= 0 && hint < len(h.Items) && h.Items[hint] == s {
		return hint
	}
	for i, it := range h.Items {
		if it == s {
			return i
		}
	}
	return -1
}

func (h *History) Delete(s string, hint int) bool {
	i := h.find(s, hint)
	if i < 0 {
		return false
	}
	h.Items = append(h.Items[:i:i], h.Items[i+1:]...)
	return true
}

func (h *History) ToTop(s string, hint int) bool {
	i := h.find(s, hint)
	if i <= 0 {
		return false
	}
	h.Items = append([]string{s}, append(h.Items[:i:i], h.Items[i+1:]...)...)
	return true
}

// shorten renders a clipping as Jumpcut's menu did: trimmed, first line, n runes + ellipsis.
func shorten(s string, n int) string {
	s = strings.TrimSpace(s)
	if i := strings.IndexAny(s, "\r\n"); i >= 0 {
		s = s[:i]
	}
	if r := []rune(s); len(r) > n {
		return string(r[:n]) + "…"
	}
	return s
}

// Remote clippings are the ones that arrived from another device over Universal
// Clipboard. remote.json beside the history lists their hashes, so the history stays a
// plain list of strings and a deleted clipping leaves no text behind.
func remotePath() string { return filepath.Join(filepath.Dir(dataPath()), "remote.json") }

func hash(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:8])
}

func readRemote() map[string]bool {
	var hashes []string
	if b, err := os.ReadFile(remotePath()); err == nil {
		json.Unmarshal(b, &hashes)
	}
	m := map[string]bool{}
	for _, h := range hashes {
		m[h] = true
	}
	return m
}

// setRemote records where the newly added clipping s came from, and forgets clippings
// that have left the history.
func setRemote(s string, remote bool) error {
	old := readRemote()
	if !remote && len(old) == 0 {
		return nil
	}
	items, err := readHistory()
	if err != nil {
		return err
	}
	keep := map[string]bool{}
	for _, it := range items {
		if h := hash(it); old[h] {
			keep[h] = true
		}
	}
	if delete(keep, hash(s)); remote {
		keep[hash(s)] = true
	}
	hashes := make([]string, 0, len(keep))
	for h := range keep {
		hashes = append(hashes, h)
	}
	b, _ := json.Marshal(hashes)
	return os.WriteFile(remotePath(), b, 0o600)
}
