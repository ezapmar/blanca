package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// The lines Blanca adds to Hyprland's autostart.lua, under hyprNote.
const (
	hyprNote     = "-- Blanca clipboard (added by `blanca setup`)"
	autostartLua = `o.launch_on_start("wl-paste --type text --watch blanca store")`
)

// writeHyprSnippets appends Blanca's lines to each Hyprland file in dir that exists and
// does not mention blanca yet. It returns one report line per file.
func writeHyprSnippets(hypr string) []string {
	var out []string
	snippets := map[string]string{
		"bindings.lua": `o.bind("CTRL + ALT + V", "Blanca clipboard", "blanca bezel")
o.bind("CTRL + ALT + SHIFT + V", "Blanca clipboard (back)", "blanca bezel --up")`,
		"autostart.lua": autostartLua,
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
		fmt.Fprintf(f, "\n%s\n%s\n", hyprNote, snippets[name])
		f.Close()
		out = append(out, "add  "+p)
	}
	return out
}

// autostarts reports whether Hyprland starts the clipboard watcher at login.
func autostarts(hypr string) bool {
	b, _ := os.ReadFile(filepath.Join(hypr, "autostart.lua"))
	return strings.Contains(string(b), "blanca store")
}

// setAutostart is the "Launch on login" switch on Omarchy: it takes Blanca's lines out
// of autostart.lua and, when on, puts them back at the end.
func setAutostart(hypr string, on bool) error {
	p := filepath.Join(hypr, "autostart.lua")
	b, err := os.ReadFile(p)
	if err != nil {
		return err
	}
	var keep []string
	for _, l := range strings.Split(string(b), "\n") {
		if !strings.Contains(l, "blanca") && l != hyprNote {
			keep = append(keep, l)
		}
	}
	s := strings.TrimRight(strings.Join(keep, "\n"), "\n") + "\n"
	if on {
		s += "\n" + hyprNote + "\n" + autostartLua + "\n"
	}
	return os.WriteFile(p, []byte(s), 0o644)
}

// Blanca's button on the Omarchy bar from Omarchy 4 on, where the bar is Quickshell's and
// its buttons are shell plugins living in ~/.config/omarchy/plugins/<id>.
const widgetID = "blanca.clipboard"

const widgetManifest = `{
  "schemaVersion": 1,
  "id": "` + widgetID + `",
  "name": "Blanca clipboard",
  "version": "1.0.0",
  "author": "Tunca Üçer",
  "description": "Everything you copy, kept; the button opens the clippings",
  "kinds": ["bar-widget"],
  "entryPoints": { "barWidget": "BarWidget.qml" },
  "barWidget": {
    "displayName": "Blanca clipboard",
    "description": "Opens the Blanca clipboard menu",
    "category": "Utilities",
    "allowMultiple": false,
    "defaultSection": "right"
  }
}
`

// The widget itself: the dog in the bar's own text colour, a click opens `blanca menu`.
// MultiEffect paints the silhouette, so the icon follows the theme the way a bar glyph does.
const widgetQML = `import QtQuick
import QtQuick.Effects

Item {
  id: root

  // Filled in by the bar when it loads the widget.
  property var bar
  property string moduleName: "` + widgetID + `"
  property var settings

  readonly property real iconSize: 16

  implicitWidth: 28
  implicitHeight: bar ? bar.barSize : 26

  Image {
    id: icon
    anchors.centerIn: parent
    width: root.iconSize
    height: root.iconSize
    // Decoded at twice the drawn size so the silhouette stays crisp on a HiDPI screen.
    sourceSize.width: root.iconSize * 2
    sourceSize.height: root.iconSize * 2
    source: Qt.resolvedUrl("blanca-symbolic.svg")
    fillMode: Image.PreserveAspectFit
    // Kept as a hidden layer so the effect can sample it as a texture.
    visible: false
    layer.enabled: true
  }

  MultiEffect {
    anchors.fill: icon
    source: icon
    colorization: 1.0
    colorizationColor: root.bar ? root.bar.foreground : "white"
  }

  MouseArea {
    anchors.fill: parent
    hoverEnabled: true
    onEntered: if (root.bar) root.bar.showTooltip(root, "Blanca clipboard")
    onExited: if (root.bar) root.bar.hideTooltip(root)
    onClicked: if (root.bar) root.bar.run("blanca menu")
  }
}
`

// writeShellWidget writes the bar-widget plugin into the Omarchy plugins directory: the
// manifest, the widget and the icon. The files are Blanca's own, so they are rewritten
// rather than left alone; where the button sits on the bar is the user's and is decided
// by enableShellWidget. Qt's SVG renderer has no currentColor, so the icon is written
// with a plain fill and recoloured by the widget.
func writeShellWidget(plugins string, icon []byte) []string {
	dir := filepath.Join(plugins, widgetID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return []string{"skip " + dir + " (" + err.Error() + ")"}
	}
	files := map[string][]byte{
		"manifest.json":       []byte(widgetManifest),
		"BarWidget.qml":       []byte(widgetQML),
		"blanca-symbolic.svg": []byte(strings.ReplaceAll(string(icon), "currentColor", "#ffffff")),
	}
	var out []string
	for _, name := range []string{"manifest.json", "BarWidget.qml", "blanca-symbolic.svg"} {
		p := filepath.Join(dir, name)
		old, err := os.ReadFile(p)
		if err == nil && string(old) == string(files[name]) {
			out = append(out, "ok   "+p+" already set up")
			continue
		}
		if err := os.WriteFile(p, files[name], 0o644); err != nil {
			fatal(err)
		}
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
