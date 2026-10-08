package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

// cmdSetup wires Blanca into an Omarchy Hyprland config: keybinds, autostart watcher and
// floating window rules. Idempotent: a file that already mentions blanca is left alone.
func cmdSetup() {
	for _, line := range writeHyprSnippets(filepath.Join(xdg("XDG_CONFIG_HOME", ".config"), "hypr")) {
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

// writeHyprSnippets appends Blanca's lines to each Hyprland file in dir that exists and
// does not mention blanca yet. It returns one report line per file.
func writeHyprSnippets(hypr string) []string {
	var out []string
	snippets := map[string]string{
		"bindings.lua": `o.bind("CTRL + ALT + V", "Blanca clipboard", "blanca bezel")
o.bind("CTRL + ALT + SHIFT + V", "Blanca clipboard (back)", "blanca bezel --up")`,
		"autostart.lua": `o.launch_on_start("wl-paste --type text --watch blanca store")`,
		"hyprland.lua": `o.window("org.omarchy.blanca", { float = true })
o.window("org.omarchy.blanca", { center = true })
o.window("org.omarchy.blanca", { size = { 640, 360 } })`,
	}
	for _, name := range []string{"bindings.lua", "autostart.lua", "hyprland.lua"} {
		p := filepath.Join(hypr, name)
		b, err := os.ReadFile(p)
		if err != nil {
			out = append(out, "skip "+p+" (not found; add the Blanca lines by hand)")
			continue
		}
		if strings.Contains(string(b), "blanca") {
			out = append(out, "ok   "+p+" already set up")
			continue
		}
		f, err := os.OpenFile(p, os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			fatal(err)
		}
		fmt.Fprintf(f, "\n-- Blanca clipboard (added by `blanca setup`)\n%s\n", snippets[name])
		f.Close()
		out = append(out, "add  "+p)
	}
	return out
}
