package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// releases is a stand-in for GitHub that answers with tag and counts the asking.
func releases(t *testing.T, tag string) *int {
	t.Helper()
	n := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n++
		w.Write([]byte(`{"tag_name": "` + tag + `"}`))
	}))
	t.Cleanup(srv.Close)
	old := latestURL
	latestURL = srv.URL
	t.Cleanup(func() { latestURL = old })
	return &n
}

func TestDue(t *testing.T) {
	now := time.Date(2026, 10, 12, 9, 0, 0, 0, time.UTC)
	for _, c := range []struct {
		why     string
		checked time.Time
		want    bool
	}{
		{"never checked", time.Time{}, true},
		{"an hour ago", now.Add(-time.Hour), false},
		{"a day ago", now.Add(-checkEvery), true},
		{"a week ago", now.Add(-7 * 24 * time.Hour), true},
		{"the clock went back", now.Add(time.Hour), true},
	} {
		var u updateState
		if !c.checked.IsZero() {
			u.Checked = c.checked.Unix()
		}
		if got := u.due(now); got != c.want {
			t.Errorf("%s: due = %v, want %v", c.why, got, c.want)
		}
	}
}

func TestPending(t *testing.T) {
	old := version
	version = "0.3.5"
	defer func() { version = old }()
	for _, c := range []struct {
		found, want string
	}{
		{"v0.3.6", "v0.3.6"},
		{"v0.3.5", ""}, // caught up with
		{"v0.3.4", ""},
		{"", ""},
	} {
		if got := (updateState{Found: c.found}).pending(); got != c.want {
			t.Errorf("pending with found %q = %q, want %q", c.found, got, c.want)
		}
	}
}

func TestAutoCheck(t *testing.T) {
	isolate(t)
	old := version
	version = "0.3.5"
	defer func() { version = old }()
	asked := releases(t, "v0.3.6")
	now := time.Date(2026, 10, 12, 9, 0, 0, 0, time.UTC)

	if got := autoCheck(now); got != "v0.3.6" {
		t.Fatalf("first check = %q, want v0.3.6", got)
	}
	if u := readUpdateState(); u.Found != "v0.3.6" || u.Told != "v0.3.6" || u.Checked != now.Unix() {
		t.Fatalf("after the first check: %+v", u)
	}
	// Within the day it does not ask again, and the menu still offers the release.
	if got := autoCheck(now.Add(time.Hour)); got != "" {
		t.Errorf("an hour later = %q, want nothing", got)
	}
	if *asked != 1 {
		t.Errorf("asked GitHub %d times, want 1", *asked)
	}
	if got := readUpdateState().pending(); got != "v0.3.6" {
		t.Errorf("pending = %q, want v0.3.6", got)
	}
	// A day later it asks again, but the same release is no longer news.
	if got := autoCheck(now.Add(checkEvery)); got != "" {
		t.Errorf("a day later = %q, want nothing: already offered", got)
	}
	if *asked != 2 {
		t.Errorf("asked GitHub %d times, want 2", *asked)
	}
}

// A release that arrives after the last one was offered is news of its own.
func TestAutoCheckNextRelease(t *testing.T) {
	isolate(t)
	old := version
	version = "0.3.5"
	defer func() { version = old }()
	now := time.Date(2026, 10, 12, 9, 0, 0, 0, time.UTC)

	releases(t, "v0.3.6")
	if got := autoCheck(now); got != "v0.3.6" {
		t.Fatalf("first check = %q, want v0.3.6", got)
	}
	releases(t, "v0.3.7")
	if got := autoCheck(now.Add(checkEvery)); got != "v0.3.7" {
		t.Errorf("after 0.3.7 is out = %q, want v0.3.7", got)
	}
}

// GitHub being unreachable costs a day's check, not what the last one found.
func TestAutoCheckOffline(t *testing.T) {
	isolate(t)
	old := version
	version = "0.3.5"
	defer func() { version = old }()
	now := time.Date(2026, 10, 12, 9, 0, 0, 0, time.UTC)

	releases(t, "v0.3.6")
	autoCheck(now)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusInternalServerError)
	}))
	defer srv.Close()
	latestURL = srv.URL
	if got := autoCheck(now.Add(checkEvery)); got != "" {
		t.Errorf("failed check = %q, want nothing", got)
	}
	u := readUpdateState()
	if u.Found != "v0.3.6" {
		t.Errorf("a failed check forgot the release: %+v", u)
	}
	if u.Checked != now.Add(checkEvery).Unix() {
		t.Errorf("a failed check is still a check: %+v", u)
	}
}

// Switched off in Settings, Blanca does not go near GitHub.
func TestAutoCheckSwitchedOff(t *testing.T) {
	isolate(t)
	old := version
	version = "0.3.5"
	defer func() { version = old }()
	asked := releases(t, "v0.3.6")

	c := loadConfig()
	c.CheckUpdates = false
	if err := saveConfig(c); err != nil {
		t.Fatal(err)
	}
	if got := autoCheck(time.Now()); got != "" {
		t.Errorf("switched off, check = %q, want nothing", got)
	}
	if *asked != 0 {
		t.Errorf("switched off, asked GitHub %d times", *asked)
	}
	// And the switch takes hold again without a restart.
	c.CheckUpdates = true
	if err := saveConfig(c); err != nil {
		t.Fatal(err)
	}
	if got := autoCheck(time.Now()); got != "v0.3.6" {
		t.Errorf("switched back on, check = %q, want v0.3.6", got)
	}
}
