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

// cmdBezel is the hotkey entry point. Pressing the hotkey while the bezel is open
// advances it (Shift goes back); otherwise a bezel is opened
// in a floating terminal the Omarchy way. On macOS `blanca watch` owns the hotkey.
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
	if err := launchBezel(self); err != nil {
		fatal(err)
	}
}

// place copies text to the clipboard and, if configured, pastes it once the bezel
// has closed and focus has returned to the previous window.
func place(cfg Config, text string) {
	os.WriteFile(filepath.Join(runtimeDir(), "placed"), []byte(text), 0o600)
	if err := copyText(text); err != nil {
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

// cmdPaste is the paste after a selection: wait for focus to settle, then send the paste chord.
func cmdPaste(cfg Config) {
	time.Sleep(200 * time.Millisecond)
	sendPaste(cfg)
}
