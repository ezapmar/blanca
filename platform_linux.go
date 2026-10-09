package main

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"syscall"
)

const appID = "org.omarchy.blanca"

//go:embed assets/blanca-symbolic.svg
var barIcon []byte

// cmdSetup wires Blanca into an Omarchy config: Hyprland keybinds, autostart watcher and
// floating window rules, and a Waybar button. Idempotent: a file that already mentions
// blanca is left alone.
func cmdSetup() {
	conf := xdg("XDG_CONFIG_HOME", ".config")
	for _, line := range append(writeHyprSnippets(filepath.Join(conf, "hypr")), writeWaybar(filepath.Join(conf, "waybar"), barIcon)...) {
		fmt.Println(line)
	}
	exec.Command("pkill", "-SIGUSR2", "waybar").Run() // reload the bar
	if err := exec.Command("hyprctl", "reload").Run(); err == nil {
		fmt.Println("ok   hyprctl reload")
	}
	if exec.Command("pgrep", "-f", "wl-paste.*blanca store").Run() != nil {
		w := exec.Command("wl-paste", "--type", "text", "--watch", "blanca", "store")
		w.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
		if w.Start() == nil {
			fmt.Println("ok   clipboard watcher started")
		}
	}
	fmt.Println("done Press Ctrl+Alt+V after copying something.")
}

// cmdWatch hands the watching to wl-paste, which runs `blanca store` on every text copy.
func cmdWatch(Config) {
	self, err := os.Executable()
	if err != nil {
		fatal(err)
	}
	wl, err := exec.LookPath("wl-paste")
	if err != nil {
		fatal(err)
	}
	fatal(syscall.Exec(wl, []string{wl, "--type", "text", "--watch", self, "store"}, os.Environ()))
}

func platformSettings(c *Config) []setting {
	return []setting{{name: "Paste with Ctrl+V",
		on: func() bool { return c.PasteMode == "ctrl-v" },
		pick: func(int) {
			if c.PasteMode == "ctrl-v" {
				c.PasteMode = "shift-insert"
			} else {
				c.PasteMode = "ctrl-v"
			}
		}}}
}

// cmdMenu is the Waybar button's click, after Jumpcut's status menu: the first `display`
// clippings and Clear All, shown in Walker. Choosing a clipping places it like the bezel.
func cmdMenu(cfg Config) {
	items, err := readHistory()
	if err != nil {
		fatal(err)
	}
	items = items[:min(cfg.Display, len(items))]
	var b strings.Builder
	for i, s := range items {
		fmt.Fprintf(&b, "%d  %s\n", i+1, shorten(s, 40))
	}
	sel := dmenu("Blanca…", b.String()+"Clear All\nSettings\n")
	switch sel {
	case "Clear All":
		if dmenu("Clear all clippings?", "Cancel\nClear\n") == "Clear" { // Jumpcut asks first
			clearAll()
		}
		return
	case "Settings":
		menuSettings(cfg)
		return
	}
	var i int
	if fmt.Sscanf(sel, "%d", &i); i >= 1 && i <= len(items) {
		use(cfg, items[i-1], i-1)
	}
}

// menuSettings lists the settings with their values. Choosing a switch flips it, choosing
// a number offers its values; the list then comes back until it is dismissed.
func menuSettings(cfg Config) {
	for {
		list := settings(&cfg)
		lines := make([]string, len(list))
		for i, s := range list {
			lines[i] = s.name + ": " + s.value()
		}
		i := slices.Index(lines, dmenu("Settings…", strings.Join(lines, "\n")+"\n"))
		if i < 0 {
			return
		}
		j := 0
		if s := list[i]; s.on == nil {
			nums := make([]string, len(s.nums))
			for k, n := range s.nums {
				nums[k] = strconv.Itoa(n)
			}
			if j = slices.Index(nums, dmenu(s.name, strings.Join(nums, "\n")+"\n")); j < 0 {
				continue
			}
		}
		list[i].pick(j)
		if err := saveConfig(cfg); err != nil {
			fatal(err)
		}
	}
}

// dmenu shows lines in Walker the way Omarchy's own menus do and returns the chosen one.
func dmenu(prompt, lines string) string {
	walker := "walker"
	if _, err := exec.LookPath("omarchy-launch-walker"); err == nil {
		walker = "omarchy-launch-walker"
	}
	cmd := exec.Command(walker, "--dmenu", "-p", prompt)
	cmd.Stdin = strings.NewReader(lines)
	out, _ := cmd.Output()
	return strings.TrimSpace(string(out))
}

func launchBezel(self string) error {
	var cmd *exec.Cmd
	if _, err := exec.LookPath("omarchy-launch-tui"); err == nil {
		cmd = exec.Command("omarchy-launch-tui", "--app-id="+appID, self, "pick")
	} else {
		cmd = exec.Command("xdg-terminal-exec", "--app-id="+appID, "-e", self, "pick")
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	return cmd.Start()
}

func copyText(text string) error {
	cp := exec.Command("wl-copy")
	cp.Stdin = strings.NewReader(text)
	cp.SysProcAttr = &syscall.SysProcAttr{Setsid: true} // survive the terminal closing
	return cp.Run()
}

func clearClipboard() { exec.Command("wl-copy", "--clear").Run() }

func sendPaste(cfg Config) {
	args := []string{"-M", "shift", "-k", "Insert", "-m", "shift"}
	if cfg.PasteMode == "ctrl-v" {
		args = []string{"-M", "ctrl", "-k", "v", "-m", "ctrl"}
	}
	exec.Command("wtype", args...).Run()
}

// sensitive reports whether the current clipboard carries the password-manager hint
// (KeePassXC and friends), the Wayland counterpart of Jumpcut's ConcealedType check.
func sensitive() bool {
	out, err := exec.Command("wl-paste", "--list-types").Output()
	return err == nil && strings.Contains(string(out), "x-kde-passwordManagerHint")
}
