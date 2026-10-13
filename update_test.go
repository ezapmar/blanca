package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
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
	if !slices.Equal(got, []string{"sh", "-c", pacmanInstaller, "sh", "v0.3.1", "aarch64"}) {
		t.Errorf("pacman install: %q", got)
	}
	for _, self := range []string{"/home/me/.local/bin/blanca", "/Applications/Blanca.app/Contents/MacOS/blanca"} {
		goos := "linux"
		if strings.Contains(self, ".app/") {
			goos = "darwin"
		}
		got := updateCommand("v0.3.1", self, goos, "amd64")
		if !slices.Equal(got, []string{"sh", "-c", installer, "sh", "v0.3.1"}) {
			t.Errorf("%s: %q", self, got)
		}
	}
}

// An update must never run a script it fetched: the scripts are this binary's own, and
// the only things they download are the release's files and their checksums.
func TestUpdateRunsNoDownloadedScript(t *testing.T) {
	for _, self := range []string{"/usr/bin/blanca", "/home/me/.local/bin/blanca"} {
		script := updateCommand("v0.3.1", self, "linux", "amd64")[2]
		if script == "" {
			t.Fatalf("%s: no script", self)
		}
		for _, bad := range []string{"| sh", "|sh", "install.sh", "raw.githubusercontent.com"} {
			for _, line := range strings.Split(script, "\n") {
				if !strings.HasPrefix(strings.TrimSpace(line), "#") && strings.Contains(line, bad) {
					t.Errorf("%s: update script has %q: %s", self, bad, line)
				}
			}
		}
	}
}

// pacmanInstaller must stop before pacman when the package is not the one the release
// published. curl, sha256sum and sudo are stand-ins on PATH.
func TestPacmanInstallerVerifies(t *testing.T) {
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("no sh")
	}
	for _, c := range []struct{ name, sums, want string }{
		{"match", "abc  blanca-bin-0.3.1-1-x86_64.pkg.tar.zst\n", "pacman -U"},
		{"mismatch", "def  blanca-bin-0.3.1-1-x86_64.pkg.tar.zst\n", ""},
		{"absent", "abc  blanca-v0.3.1-linux-amd64.tar.gz\n", ""},
	} {
		bin := t.TempDir()
		ran := filepath.Join(bin, "ran")
		stub := func(name, body string) {
			if err := os.WriteFile(filepath.Join(bin, name), []byte("#!/bin/sh\n"+body), 0o755); err != nil {
				t.Fatal(err)
			}
		}
		// curl -fsSL <url> -o <file>: the sums for SHA256SUMS, anything for the package.
		stub("curl", `case "$2" in *SHA256SUMS) printf '%s' "$SUMS" > "$4";; *) echo pkg > "$4";; esac`)
		stub("sha256sum", `echo "abc  $1"`)
		stub("sudo", `echo "$1 $2" > "$RAN"`)
		cmd := exec.Command("/bin/sh", "-c", pacmanInstaller, "sh", "v0.3.1", "x86_64")
		cmd.Env = []string{"PATH=" + bin + ":/usr/bin:/bin", "SUMS=" + c.sums, "RAN=" + ran}
		out, err := cmd.CombinedOutput()
		b, _ := os.ReadFile(ran)
		if got := strings.TrimSpace(string(b)); got != c.want {
			t.Errorf("%s: pacman ran %q, want %q\n%s", c.name, got, c.want, out)
		}
		if (err == nil) != (c.want != "") {
			t.Errorf("%s: exit %v\n%s", c.name, err, out)
		}
	}
}
