package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const appID = "org.omarchy.blanca"

// cmdBezel is the hotkey entry point. Pressing the hotkey while the bezel is open
// advances it (Shift goes back), exactly like Jumpcut; otherwise a bezel is opened
// in a floating terminal the Omarchy way.
func cmdBezel(up bool) {
	pidfile := filepath.Join(runtimeDir(), "pick.pid")
	if b, err := os.ReadFile(pidfile); err == nil {
		pid, _ := strconv.Atoi(strings.TrimSpace(string(b)))
		sig := syscall.SIGUSR1
		if up {
			sig = syscall.SIGUSR2
		}
		if pid > 0 && syscall.Kill(pid, sig) == nil {
			return
		}
		os.Remove(pidfile) // stale
	}
	self, err := os.Executable()
	if err != nil {
		fatal(err)
	}
	var cmd *exec.Cmd
	if _, err := exec.LookPath("omarchy-launch-tui"); err == nil {
		cmd = exec.Command("omarchy-launch-tui", "--app-id="+appID, self, "pick")
	} else {
		cmd = exec.Command("xdg-terminal-exec", "--app-id="+appID, "-e", self, "pick")
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		fatal(err)
	}
}

// place copies text to the clipboard and, if configured, pastes it once the bezel
// has closed and focus has returned to the previous window.
func place(cfg Config, text string) {
	os.WriteFile(filepath.Join(runtimeDir(), "placed"), []byte(text), 0o600)
	tool := "wl-copy"
	if _, err := exec.LookPath(tool); err != nil {
		tool = "pbcopy" // development on macOS
	}
	cp := exec.Command(tool)
	cp.Stdin = strings.NewReader(text)
	cp.SysProcAttr = &syscall.SysProcAttr{Setsid: true} // survive the terminal closing
	if err := cp.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "blanca: copy:", err)
		return
	}
	if !cfg.Paste {
		return
	}
	self, _ := os.Executable()
	p := exec.Command(self, "paste")
	p.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	p.Start()
}

// cmdPaste is Jumpcut's fakeCommandV: wait for focus to settle, then send the paste chord.
func cmdPaste(cfg Config) {
	time.Sleep(200 * time.Millisecond)
	args := []string{"-M", "shift", "-k", "Insert", "-m", "shift"}
	if cfg.PasteMode == "ctrl-v" {
		args = []string{"-M", "ctrl", "-k", "v", "-m", "ctrl"}
	}
	exec.Command("wtype", args...).Run()
}
