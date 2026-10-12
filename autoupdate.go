package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
)

// checkEvery is how often Blanca looks for a release of its own accord. Once a day is
// often enough for a clipboard manager and light on GitHub.
const checkEvery = 24 * time.Hour

// updateState is what the daily check leaves behind in the data directory, so that every
// blanca invocation knows what the last one found without asking GitHub again.
type updateState struct {
	Checked int64  `json:"checked"` // unix time of the last look at GitHub
	Found   string `json:"found"`   // tag of the release it found, "v0.3.6"
	Told    string `json:"told"`    // tag the user has already been offered
}

func updatePath() string { return filepath.Join(filepath.Dir(dataPath()), "update.json") }

func readUpdateState() updateState {
	var u updateState
	if b, err := os.ReadFile(updatePath()); err == nil {
		json.Unmarshal(b, &u) // a damaged file is a check that has not happened
	}
	return u
}

func writeUpdateState(u updateState) error {
	p := updatePath()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	b, _ := json.MarshalIndent(u, "", "  ")
	return os.WriteFile(p, append(b, '\n'), 0o600)
}

// due reports whether it has been long enough since the last look at GitHub. A check
// from the future, a clock that has been set back, counts as due.
func (u updateState) due(now time.Time) bool {
	last := time.Unix(u.Checked, 0)
	return u.Checked == 0 || now.Sub(last) >= checkEvery || last.After(now)
}

// pending is the release waiting to be installed, or "" when this Blanca is current. The
// menus ask this: it is the last check's answer, never the network.
func (u updateState) pending() string {
	if newer(u.Found, version) {
		return u.Found
	}
	return ""
}

// autoCheck is the daily look for a newer Blanca, unless the user has switched it off:
// the config is read each time, so the switch takes hold without a restart. It asks
// GitHub only when a day has passed, and returns a tag only the first time it finds one: the menu goes on offering
// that release, but the user is interrupted about it once. The App Store's Blanca never
// checks, since the Store updates it.
func autoCheck(now time.Time) string {
	if appStore || !loadConfig().CheckUpdates {
		return ""
	}
	u := readUpdateState()
	if !u.due(now) {
		return ""
	}
	tag, err := latestTag()
	u.Checked = now.Unix()
	if err == nil {
		u.Found = tag
	}
	// A release already announced, or one this Blanca has caught up with, is not news.
	if u.pending() == "" || u.Found == u.Told {
		writeUpdateState(u)
		return ""
	}
	u.Told = u.Found
	writeUpdateState(u)
	return u.Found
}

// startCheck runs `blanca check` detached, for the platforms with no process of their
// own to run the check in. It is fire and forget: a check that fails is simply not
// recorded, and the next copy brings another one.
func startCheck() {
	if appStore || !loadConfig().CheckUpdates {
		return
	}
	if !readUpdateState().due(time.Now()) {
		return
	}
	self, err := os.Executable()
	if err != nil {
		return
	}
	cmd := exec.Command(self, "check")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	cmd.Start()
}

// cmdCheck is the detached `blanca check`: look once, write down what was found, say
// nothing. It is not in the usage, being nobody's business but Blanca's.
func cmdCheck() { autoCheck(time.Now()) }
