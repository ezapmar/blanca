package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const repo = "ezapmar/blanca"

// latestURL answers with the newest release. A variable so the tests can point it at
// a server of their own.
var latestURL = "https://api.github.com/repos/" + repo + "/releases/latest"

// latestTag asks GitHub for the tag of the newest release, "v0.2.0".
func latestTag() (string, error) {
	c := http.Client{Timeout: 15 * time.Second}
	resp, err := c.Get(latestURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%s: %s", latestURL, resp.Status)
	}
	var r struct {
		Tag string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return "", err
	}
	if _, ok := parseVersion(r.Tag); !ok {
		return "", fmt.Errorf("latest release has an odd tag %q", r.Tag)
	}
	return r.Tag, nil
}

// parseVersion turns "v0.2.0" or "0.2.0" into its numbers. Nothing but digits and dots
// gets through, which is what lets a tag go into a shell command.
func parseVersion(s string) ([]int, bool) {
	var v []int
	for _, p := range strings.Split(strings.TrimPrefix(s, "v"), ".") {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 {
			return nil, false
		}
		v = append(v, n)
	}
	return v, true
}

// newer reports whether the release tag is ahead of the running version. A build that
// is not a release, "dev", is behind every release.
func newer(tag, current string) bool {
	t, ok := parseVersion(tag)
	if !ok {
		return false
	}
	c, ok := parseVersion(current)
	if !ok {
		return true
	}
	for i := range max(len(t), len(c)) {
		var a, b int
		if i < len(t) {
			a = t[i]
		}
		if i < len(c) {
			b = c[i]
		}
		if a != b {
			return a > b
		}
	}
	return false
}

// updateCommand is the shell line that installs release tag over the blanca at self.
// A binary under /usr came from the pacman package, so pacman replaces it; any other
// goes the way it came, through install.sh.
func updateCommand(tag, self, goos, goarch string) string {
	if goos == "linux" && strings.HasPrefix(self, "/usr/") {
		arch := map[string]string{"amd64": "x86_64", "arm64": "aarch64"}[goarch]
		ver := strings.TrimPrefix(tag, "v")
		return fmt.Sprintf("sudo pacman -U https://github.com/%s/releases/download/%s/blanca-bin-%s-1-%s.pkg.tar.zst",
			repo, tag, ver, arch)
	}
	return fmt.Sprintf("curl -fsSL https://raw.githubusercontent.com/%s/%s/install.sh | sh -s %s", repo, tag, tag)
}

// checkUpdate is the menu's Check for Updates. It returns the tag of a release ahead of
// this one, or no tag and a line to show instead: that this is the latest, or why the
// check failed.
func checkUpdate() (tag, note string) {
	tag, err := latestTag()
	if err != nil {
		return "", "Could not check for updates: " + err.Error()
	}
	if !newer(tag, version) {
		return "", "Blanca " + version + " is the latest"
	}
	return tag, ""
}

// startUpdate runs the installer for tag in a session of its own, so that it outlives
// the blanca it replaces.
func startUpdate(tag string) error {
	self, err := os.Executable()
	if err != nil {
		return err
	}
	cmd := exec.Command("/bin/sh", "-c", updateCommand(tag, self, runtime.GOOS, runtime.GOARCH))
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	return cmd.Start()
}

// cmdUpdate installs the newest release if it is ahead of this one.
func cmdUpdate() {
	tag, err := latestTag()
	if err != nil {
		fatal(err)
	}
	if !newer(tag, version) {
		fmt.Printf("blanca %s is the latest\n", version)
		return
	}
	self, err := os.Executable()
	if err != nil {
		fatal(err)
	}
	cmd := updateCommand(tag, self, runtime.GOOS, runtime.GOARCH)
	fmt.Printf("blanca %s -> %s\n%s\n", version, strings.TrimPrefix(tag, "v"), cmd)
	// Become the shell rather than start one: on macOS the installer stops every running
	// blanca, and this would be one of them.
	fatal(syscall.Exec("/bin/sh", []string{"sh", "-c", cmd}, os.Environ()))
}
