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

// platformSettings: "Launch on login" is the watcher line in Hyprland's autostart.lua.
// The running watcher is left as it is.
func platformSettings(c *Config) []setting {
	hypr := filepath.Join(xdg("XDG_CONFIG_HOME", ".config"), "hypr")
	login := setting{name: "Launch on login",
		on: func() bool { return autostarts(hypr) },
		pick: func(int) {
			if err := setAutostart(hypr, !autostarts(hypr)); err != nil {
				fmt.Fprintln(os.Stderr, "blanca:", err)
			}
		}}
	return []setting{login, {name: "Paste with Ctrl+V",
		on: func() bool { return c.PasteMode == "ctrl-v" },
		pick: func(int) {
			if c.PasteMode == "ctrl-v" {
				c.PasteMode = "shift-insert"
			} else {
				c.PasteMode = "ctrl-v"
			}
		}}}
}

// cmdMenu is the Waybar button's click: the first `display`
// clippings, Clear, Settings and Check for Updates, shown in Walker. Choosing a clipping places it like the bezel.
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
	sel := dmenu("Blanca…", b.String()+"Clear\nSettings\nCheck for Updates\n")
	switch sel {
	case "Clear":
		menuClear()
		return
	case "Settings":
		menuSettings(cfg)
		return
	case "Check for Updates":
		menuUpdate()
		return
	}
	var i int
	if fmt.Sscanf(sel, "%d", &i); i >= 1 && i <= len(items) {
		use(cfg, items[i-1], i-1)
	}
}

// menuClear offers how far back to clear and asks before doing it.
func menuClear() {
	var names []string
	for _, o := range clearOptions {
		names = append(names, o.name)
	}
	i := slices.Index(names, dmenu("Clear…", strings.Join(names, "\n")+"\n"))
	if i < 0 || dmenu("Clear clippings: "+strings.ToLower(names[i])+"?", "Cancel\nClear\n") != "Clear" {
		return
	}
	if err := clearSpan(clearOptions[i].span); err != nil {
		fatal(err)
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

// menuUpdate looks for a newer release and offers it. The update itself runs in a
// terminal, `blanca update`, where pacman can ask for a password and the installer can
// be read.
func menuUpdate() {
	tag, note := checkUpdate()
	if tag == "" {
		dmenu(note, "OK\n")
		return
	}
	if dmenu("Blanca "+strings.TrimPrefix(tag, "v")+" is available", "Update\nLater\n") != "Update" {
		return
	}
	self, err := os.Executable()
	if err == nil {
		err = launchTUI("sh", "-c", `"$0" update; printf '\nReturn closes this. '; read _`, self)
	}
	if err != nil {
		fatal(err)
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

func launchBezel(self string) error { return launchTUI(self, "pick") }

// launchTUI runs a command in Blanca's floating terminal.
func launchTUI(args ...string) error {
	cmd := exec.Command("xdg-terminal-exec", append([]string{"--app-id=" + appID, "-e"}, args...)...)
	if _, err := exec.LookPath("omarchy-launch-tui"); err == nil {
		cmd = exec.Command("omarchy-launch-tui", append([]string{"--app-id=" + appID}, args...)...)
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
// (KeePassXC and friends), the Wayland counterpart of macOS's ConcealedType.
func sensitive() bool {
	out, err := exec.Command("wl-paste", "--list-types").Output()
	return err == nil && strings.Contains(string(out), "x-kde-passwordManagerHint")
}
