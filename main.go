// Blanca is a clipboard manager for Omarchy Linux and macOS, derived from Jumpcut.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var version = "dev" // set by -ldflags at release time

// Config mirrors Jumpcut's preferences; every field is optional in config.json.
type Config struct {
	Remember        int    `json:"remember"`         // clippings kept (min 10)
	Display         int    `json:"display"`          // clippings shown by `list`
	Wraparound      bool   `json:"wraparound"`       // bezel wraps at the ends
	Paste           bool   `json:"paste"`            // selecting also pastes
	Sticky          bool   `json:"sticky"`           // macOS: releasing the modifiers does not select
	PasteMode       string `json:"paste_mode"`       // "shift-insert" | "ctrl-v"
	MoveToTop       bool   `json:"move_to_top"`      // move clipping to top after use
	AllowWhitespace bool   `json:"allow_whitespace"` // keep whitespace-only clippings
	IgnoreLarge     bool   `json:"ignore_large"`     // skip clippings > 50000 chars
	IgnoreSensitive bool   `json:"ignore_sensitive"` // skip password-manager clippings
}

func loadConfig() Config {
	c := Config{Remember: 99, Display: 10, Paste: true, PasteMode: "shift-insert",
		IgnoreLarge: true, IgnoreSensitive: true}
	p := configPath()
	if b, err := os.ReadFile(p); err == nil {
		if err := json.Unmarshal(b, &c); err != nil {
			fatal(fmt.Errorf("%s: %w", p, err))
		}
	}
	c.Remember = max(c.Remember, 10)
	return c
}

func configPath() string {
	return filepath.Join(xdg("XDG_CONFIG_HOME", ".config"), "blanca", "config.json")
}

func xdg(env, fallback string) string {
	if v := os.Getenv(env); v != "" {
		return v
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, fallback)
}

func dataPath() string {
	return filepath.Join(xdg("XDG_DATA_HOME", ".local/share"), "blanca", "history.json")
}

func runtimeDir() string {
	d := os.Getenv("XDG_RUNTIME_DIR")
	if d == "" {
		d = os.TempDir()
	}
	d = filepath.Join(d, "blanca")
	os.MkdirAll(d, 0o700)
	return d
}

// clearAll forgets every clipping, the one on the clipboard included.
func clearAll() error {
	_, err := withHistory(func(h *History) bool { h.Items = nil; return true })
	os.Remove(remotePath())
	clearClipboard()
	return err
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "blanca:", err)
	os.Exit(1)
}

func usage() {
	fmt.Fprint(os.Stderr, `usage: blanca <command>

  store         read a clipping from stdin into the history (for wl-paste --watch)
  watch         record every text copy; on macOS also the hotkey and the bezel
  pick          show the bezel in this terminal
  bezel [--up]  hotkey entry: advance an open bezel, or open one
  menu          Waybar button entry: the newest clippings and the settings as a Walker menu
  list [N]      print the first N clippings, shortened
  get N         print clipping N in full
  clear         forget all clippings
  setup         Hyprland keybind, autostart, window rules and Waybar button; on macOS a LaunchAgent
  update        install the newest release, if there is one
  version
`)
	os.Exit(2)
}

func main() {
	if len(os.Args) < 2 {
		self, _ := os.Executable()
		if !strings.Contains(self, ".app/Contents/MacOS/") {
			usage()
		}
		os.Args = append(os.Args, "watch") // Blanca.app opened from Finder
	}
	cfg := loadConfig()
	arg := ""
	if len(os.Args) > 2 {
		arg = os.Args[2]
	}
	switch os.Args[1] {
	case "store":
		cmdStore(cfg)
	case "watch":
		cmdWatch(cfg)
	case "pick":
		cmdPick(cfg)
	case "bezel":
		cmdBezel(arg == "--up")
	case "menu":
		cmdMenu(cfg)
	case "paste":
		cmdPaste(cfg)
	case "list":
		n := cfg.Display
		if arg != "" {
			n, _ = strconv.Atoi(arg)
		}
		items, err := readHistory()
		if err != nil {
			fatal(err)
		}
		for i, s := range items[:min(n, len(items))] {
			fmt.Printf("%d\t%s\n", i+1, shorten(s, 40))
		}
	case "get":
		i, _ := strconv.Atoi(arg)
		items, err := readHistory()
		if err != nil {
			fatal(err)
		}
		if i < 1 || i > len(items) {
			os.Exit(1)
		}
		fmt.Print(items[i-1])
	case "clear":
		if err := clearAll(); err != nil {
			fatal(err)
		}
	case "setup":
		cmdSetup()
	case "update":
		cmdUpdate()
	case "version", "--version":
		fmt.Println("blanca", version)
	default:
		usage()
	}
}
