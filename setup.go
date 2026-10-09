package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

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

var modulesRight = regexp.MustCompile(`"modules-right"\s*:\s*\[`)

// writeWaybar adds a Blanca button to the Waybar config in dir: the module, its place at
// the head of modules-right, the icon and its style. Like writeHyprSnippets it leaves a
// file that already mentions blanca alone and returns one report line per file.
func writeWaybar(dir string, icon []byte) []string {
	var out []string
	edit := func(name string, change func(string) (string, bool)) {
		p := filepath.Join(dir, name)
		b, err := os.ReadFile(p)
		if err != nil {
			out = append(out, "skip "+p+" (not found)")
			return
		}
		if strings.Contains(string(b), "blanca") {
			out = append(out, "ok   "+p+" already set up")
			return
		}
		s, ok := change(string(b))
		if !ok {
			out = append(out, "skip "+p+" (no modules-right; add custom/blanca by hand)")
			return
		}
		if err := os.WriteFile(p, []byte(s), 0o644); err != nil {
			fatal(err)
		}
		out = append(out, "add  "+p)
	}
	edit("config.jsonc", func(s string) (string, bool) {
		loc := modulesRight.FindStringIndex(s)
		if loc == nil {
			return s, false
		}
		return s[:loc[0]] + `"custom/blanca": { "format": " ", "on-click": "blanca menu", "tooltip-format": "Blanca clipboard" },
  ` + s[loc[0]:loc[1]] + `"custom/blanca", ` + s[loc[1]:], true
	})
	edit("style.css", func(s string) (string, bool) {
		os.WriteFile(filepath.Join(dir, "blanca-symbolic.svg"), icon, 0o644)
		return s + `
/* Blanca clipboard (added by blanca setup): the icon takes the bar's text colour */
#custom-blanca {
  min-width: 16px;
  margin: 0 7.5px;
  background-image: -gtk-recolor(url("blanca-symbolic.svg"));
  background-repeat: no-repeat;
  background-position: center;
  background-size: 16px 16px;
}
`, true
	})
	return out
}
