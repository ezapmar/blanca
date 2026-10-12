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
// floating window rules, and a button on the bar. Idempotent: a file that already mentions
// blanca is left alone.
func cmdSetup() {
	conf := xdg("XDG_CONFIG_HOME", ".config")
	lines := writeHyprSnippets(filepath.Join(conf, "hypr"))
	// From Omarchy 4 on the bar is Quickshell's and takes plugins; before that it was
	// Waybar, which takes a module and a stylesheet.
	if _, err := exec.LookPath("omarchy-shell"); err == nil {
		plugins := filepath.Join(conf, "omarchy", "plugins")
		lines = append(lines, retireOldWidget(plugins)...)
		lines = append(lines, writeShellWidget(plugins, barIcon)...)
		lines = append(lines, enableShellWidget(filepath.Join(conf, "omarchy", "shell.json")))
	} else {
		lines = append(lines, writeWaybar(filepath.Join(conf, "waybar"), barIcon)...)
		exec.Command("pkill", "-SIGUSR2", "waybar").Run() // reload the bar
	}
	for _, line := range lines {
		fmt.Println(line)
	}
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

// retireOldWidget takes away the button 0.3.4 wrote under the un-namespaced id, now that
// the widget has a repo and an id of its own. Only Blanca's own copy is removed: one the
// user has made a checkout of is theirs, and is left for them to remove.
func retireOldWidget(plugins string) []string {
	dir := filepath.Join(plugins, oldWidgetID)
	if _, err := os.Stat(dir); err != nil {
		return nil
	}
	if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
		return []string{"skip " + dir + " (a checkout of the old id; remove it yourself)"}
	}
	exec.Command("omarchy-plugin-disable", oldWidgetID).Run() // takes it off the bar
	if err := os.RemoveAll(dir); err != nil {
		return []string{"skip " + dir + " (" + err.Error() + ")"}
	}
	return []string{"gone " + dir + " (replaced by " + widgetID + ")"}
}

// enableShellWidget asks the running Omarchy shell to take in the plugin and put its
// button on the bar. A shell.json that already names the widget is left alone: where the
// button sits, and whether it is there at all, is the user's to decide after the first run.
func enableShellWidget(shellJSON string) string {
	if b, err := os.ReadFile(shellJSON); err == nil && strings.Contains(string(b), widgetID) {
		return "ok   " + shellJSON + " already carries the button"
	}
	manual := "; put it on the bar with `omarchy plugin enable " + widgetID + "`"
	if out, err := exec.Command("omarchy-shell", "shell", "rescanPlugins").CombinedOutput(); err != nil {
		return "skip the bar button (" + reason(out, err) + manual + ")"
	}
	if out, err := exec.Command("omarchy-plugin-enable", widgetID).CombinedOutput(); err != nil {
		return "skip the bar button (" + reason(out, err) + manual + ")"
	}
	return "add  the bar button"
}

// reason is what a shelled-out command said about failing, or failing that, the error.
func reason(out []byte, err error) string {
	if s := strings.TrimSpace(string(out)); s != "" {
		return strings.ReplaceAll(s, "\n", "; ")
	}
	return err.Error()
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
// clippings, Clear, Settings, About and Check for Updates, shown in Walker. Choosing a clipping places it like the bezel.
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
	sel := dmenu("Blanca…", b.String()+"Clear\nSettings\nAbout Blanca\nCheck for Updates\n")
	switch sel {
	case "Clear":
		menuClear()
		return
	case "Settings":
		menuSettings(cfg)
		return
	case "About Blanca":
		menuAbout()
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

// menuAbout shows the version over a few words about Blanca and who made it, and offers
// its page and a letter to its maker.
func menuAbout() {
	switch dmenu("Blanca "+version, aboutText+"\n"+aboutNote+"\nWebsite\nEmail\n") {
	case "Website":
		exec.Command("xdg-open", homepage).Start()
	case "Email":
		exec.Command("xdg-open", "mailto:"+email).Start()
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

// dmenu shows lines the way Omarchy's own menus do and returns the chosen one, or "" when
// the menu was dismissed. From Omarchy 4 on that menu is the shell's own; before it, Walker.
func dmenu(prompt, lines string) string {
	cmd := exec.Command("walker", "--dmenu", "-p", prompt)
	if _, err := exec.LookPath("omarchy-menu-select"); err == nil {
		cmd = exec.Command("omarchy-menu-select", prompt)
	} else if _, err := exec.LookPath("omarchy-launch-walker"); err == nil {
		cmd = exec.Command("omarchy-launch-walker", "--dmenu", "-p", prompt)
	}
	cmd.Stdin = strings.NewReader(lines)
	out, _ := cmd.Output()
	return strings.TrimSpace(string(out))
}

func launchBezel(self string) error { return launchTUI(self, "pick") }

// launchTUI runs a command in Blanca's floating terminal.
func launchTUI(args ...string) error {
	cmd := exec.Command("xdg-terminal-exec", append([]string{"--app-id=" + appID, "-e"}, args...)...)
	// Omarchy's own launcher names the window org.omarchy.<command>, which for the blanca
	// binary is the app-id the window rules are written for. Only Omarchy 4 and later take
	// an --app-id of their own, and one passed to an older Omarchy is read as the command.
	if _, err := exec.LookPath("omarchy-launch-tui"); err == nil {
		cmd = exec.Command("omarchy-launch-tui", args...)
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
