package main

import (
	"bytes"
	_ "embed"
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

// installer is the install.sh this blanca was built with. An update runs this copy and
// never one fetched at update time: it is as trusted as the binary it sits in, and the
// first thing it does with a download is check it against the release's checksum.
//
//go:embed install.sh
var installer string

// pacmanInstaller fetches the release's pacman package, $1 the tag and $2 the pacman
// architecture, and hands it to pacman only once it matches the release's SHA256SUMS.
const pacmanInstaller = `set -eu
pkg="blanca-bin-${1#v}-1-$2.pkg.tar.zst"
base="https://github.com/` + repo + `/releases/download/$1"
tmp=$(mktemp -d); trap 'rm -rf "$tmp"' EXIT
echo "Downloading $pkg"
curl -fsSL "$base/$pkg" -o "$tmp/$pkg"
curl -fsSL "$base/SHA256SUMS" -o "$tmp/SHA256SUMS"
want=$(grep " $pkg\$" "$tmp/SHA256SUMS" | cut -d' ' -f1 || true)
got=$(sha256sum "$tmp/$pkg" | cut -d' ' -f1)
[ -n "$want" ] && [ "$got" = "$want" ] || {
  echo "blanca: $pkg does not match the checksum in the release" >&2
  echo "  expected ${want:-one in SHA256SUMS, which has none}" >&2
  echo "  got      $got" >&2
  exit 1
}
sudo pacman -U "$tmp/$pkg"
`

// updateCommand is the sh command line that installs release tag over the blanca at
// self. A binary under /usr came from the pacman package, so pacman replaces it; any
// other goes the way it came, through the installer. Either way the script is this
// binary's own, and nothing it downloads is unpacked or run before its checksum matches.
func updateCommand(tag, self, goos, goarch string) []string {
	if goos == "linux" && strings.HasPrefix(self, "/usr/") {
		arch := map[string]string{"amd64": "x86_64", "arm64": "aarch64"}[goarch]
		return []string{"sh", "-c", pacmanInstaller, "sh", tag, arch}
	}
	return []string{"sh", "-c", installer, "sh", tag}
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
// the blanca it replaces. The installer stops this blanca on its way, so one that is
// still here when the installer ends was not updated: failed is then told why.
func startUpdate(tag string, failed func(why string)) error {
	self, err := os.Executable()
	if err != nil {
		return err
	}
	var out bytes.Buffer
	cmd := exec.Command("/bin/sh", updateCommand(tag, self, runtime.GOOS, runtime.GOARCH)[1:]...)
	cmd.Stderr = &out
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		return err
	}
	go func() {
		cmd.Wait()
		failed(strings.TrimSpace(out.String()))
	}()
	return nil
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
	fmt.Printf("blanca %s -> %s\n", version, strings.TrimPrefix(tag, "v"))
	// Become the shell rather than start one: on macOS the installer stops every running
	// blanca, and this would be one of them.
	fatal(syscall.Exec("/bin/sh", updateCommand(tag, self, runtime.GOOS, runtime.GOARCH), os.Environ()))
}
