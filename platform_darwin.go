package main

/*
#cgo CFLAGS: -fobjc-arc
#cgo LDFLAGS: -framework Cocoa -framework Carbon -framework ServiceManagement
#include <stdbool.h>
#include <stdlib.h>
void bzRun(bool paste, const void *icon, int iconLen);
void bzMenuAdd(const char *title, bool remote);
void bzClearAdd(const char *title);
void bzSetting(const char *title, bool on, int tag, bool sub);
void bzShow(const char *text, const char *title);
void bzHide(void);
void bzCopy(const char *text);
void bzClear(void);
void bzPaste(void);
bool bzSensitive(void);
bool bzLogin(void);
bool bzSetLogin(bool on);
*/
import "C"

import (
	_ "embed"
	"errors"
	"fmt"
	"html"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"unsafe"
)

const label = "com.github.ezapmar.blanca"

//go:embed assets/png/blanca-symbolic-36.png
var menuIcon []byte

func init() { runtime.LockOSThread() } // AppKit wants the main thread

// bz is the bezel and menu state of the running `blanca watch`. Only the main thread touches it.
var bz struct {
	cfg   Config
	items []string
	pos   int
	shown bool
	menu  []string // clippings listed in the open menu
}

// cmdWatch is Blanca's one process on macOS: it polls the pasteboard, owns the
// Ctrl+Alt+V hotkey, the menu bar item and the bezel. It never returns.
func cmdWatch(cfg Config) {
	bz.cfg = cfg
	if _, err := os.Stat(agentPath()); err == nil && inApp() && bool(C.bzSetLogin(true)) {
		os.Remove(agentPath()) // a LaunchAgent from before Blanca.app could be a login item
	}
	C.bzRun(C.bool(cfg.Paste), unsafe.Pointer(&menuIcon[0]), C.int(len(menuIcon)))
}

// goMenu fills the opening menu with the first `display` clippings, as Jumpcut does.
//
//export goMenu
func goMenu() {
	items, _ := readHistory()
	bz.menu = items[:min(bz.cfg.Display, len(items))]
	remote := readRemote()
	for _, s := range bz.menu {
		t := C.CString(shorten(s, 40))
		C.bzMenuAdd(t, C.bool(remote[hash(s)]))
		C.free(unsafe.Pointer(t))
	}
}

//export goMenuPick
func goMenuPick(i C.int) {
	if int(i) < len(bz.menu) {
		use(bz.cfg, bz.menu[i], int(i))
	}
}

// goClearMenu fills the Clear submenu with how far back to clear.
//
//export goClearMenu
func goClearMenu() {
	for _, o := range clearOptions {
		t := C.CString(o.name)
		C.bzClearAdd(t)
		C.free(unsafe.Pointer(t))
	}
}

//export goMenuClear
func goMenuClear(i C.int) {
	if err := clearSpan(clearOptions[i].span); err != nil {
		fmt.Fprintln(os.Stderr, "blanca:", err)
	}
}

// goSettings fills the Settings submenu: a ticked line per switch, and for a number a
// submenu of its values. A line's tag is 100 times the setting plus the value's index.
//
//export goSettings
func goSettings() {
	add := func(title string, on bool, tag int, sub bool) {
		t := C.CString(title)
		C.bzSetting(t, C.bool(on), C.int(tag), C.bool(sub))
		C.free(unsafe.Pointer(t))
	}
	for i, s := range settings(&bz.cfg) {
		if s.on != nil {
			add(s.name, s.on(), i*100, false)
			continue
		}
		add(s.name, false, -1, false)
		for j, n := range s.nums {
			add(strconv.Itoa(n), n == *s.num, i*100+j, true)
		}
	}
}

//export goSet
func goSet(tag C.int) {
	if tag < 0 {
		return
	}
	settings(&bz.cfg)[tag/100].pick(int(tag % 100))
	if err := saveConfig(bz.cfg); err != nil {
		fmt.Fprintln(os.Stderr, "blanca:", err)
	}
}

// goUpdateCheck is the menu's Check for Updates, called off the main thread because it
// waits on the network. It returns the tag of a newer release, or NULL and a note to
// show instead. The caller frees both.
//
//export goUpdateCheck
func goUpdateCheck(note **C.char) *C.char {
	tag, n := checkUpdate()
	if tag == "" {
		*note = C.CString(n)
		return nil
	}
	return C.CString(tag)
}

// goUpdate installs release tag over this Blanca, which the installer stops and reopens.
//
//export goUpdate
func goUpdate(tag *C.char) {
	if err := startUpdate(C.GoString(tag)); err != nil {
		fmt.Fprintln(os.Stderr, "blanca:", err)
	}
}

// goRelease is every modifier key going up, which selects, as in Jumpcut.
//
//export goRelease
func goRelease() {
	if !bz.cfg.Sticky {
		bezelEvent(event{k: kEnter})
	}
}

// goClip records a copy. remote is set when it came from another device, an iPhone say,
// over Universal Clipboard.
//
//export goClip
func goClip(s *C.char, remote C.bool) {
	text := C.GoString(s)
	added, err := store(bz.cfg, text, sensitive)
	// macOS can announce a remote copy twice, the marker only on the second, which store
	// drops as a repeat of the top clipping. Mark it all the same.
	if err == nil && (added || bool(remote)) {
		err = setRemote(text, bool(remote))
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "blanca:", err)
	}
}

//export goHotkey
func goHotkey(up C.bool) {
	e := event{k: kNext}
	if up {
		e.k = kPrev
	}
	if !bz.shown {
		items, err := readHistory()
		if err != nil || len(items) == 0 {
			return
		}
		bz.items, bz.pos, bz.shown, e = items, 0, true, event{}
	}
	bezelEvent(e)
}

//export goKey
func goKey(code, ch C.int) { bezelEvent(macKey(int(code), rune(ch))) }

func bezelEvent(e event) {
	if !bz.shown {
		return
	}
	var done bool
	if bz.items, bz.pos, done = act(bz.cfg, bz.items, bz.pos, e); done {
		bz.shown = false
		C.bzHide()
		return
	}
	text := C.CString(bz.items[bz.pos])
	title := C.CString(fmt.Sprintf("Blanca  %d/%d", bz.pos+1, len(bz.items)))
	C.bzShow(text, title)
	C.free(unsafe.Pointer(text))
	C.free(unsafe.Pointer(title))
}

// macKey maps a macOS virtual key code, or failing that its character, to a bezel event.
func macKey(code int, char rune) event {
	switch code {
	case 125, 124: // down, right
		return event{k: kNext}
	case 126, 123: // up, left
		return event{k: kPrev}
	case 121:
		return event{k: kPgDn}
	case 116:
		return event{k: kPgUp}
	case 115:
		return event{k: kHome}
	case 119:
		return event{k: kEnd}
	case 36, 76: // return, enter
		return event{k: kEnter}
	case 51, 117: // backspace, delete
		return event{k: kDelete}
	case 53: // escape
		return event{k: kQuit}
	}
	if char >= '0' && char <= '9' {
		return event{k: kDigit, n: int(char - '0')}
	}
	return event{}
}

func agentPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Library", "LaunchAgents", label+".plist")
}

// writeAgent installs the LaunchAgent that runs this binary's `watch` at login and
// restarts it after a crash, but not after Quit.
func writeAgent() error {
	self, err := os.Executable()
	if err != nil {
		return err
	}
	os.MkdirAll(filepath.Dir(agentPath()), 0o755)
	return os.WriteFile(agentPath(), []byte(fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key><string>%s</string>
	<key>ProgramArguments</key><array><string>%s</string><string>watch</string></array>
	<key>RunAtLoad</key><true/>
	<key>KeepAlive</key><dict><key>SuccessfulExit</key><false/></dict>
	<key>ProcessType</key><string>Interactive</string>
</dict>
</plist>
`, label, html.EscapeString(self))), 0o644)
}

// inApp reports whether this is the binary inside Blanca.app.
func inApp() bool {
	self, _ := os.Executable()
	return strings.Contains(self, ".app/Contents/MacOS/")
}

// platformSettings: "Launch on login" makes Blanca.app one of the Open at Login items in
// System Settings. The bare binary cannot be one, so there the switch writes or removes
// the LaunchAgent file. Either way the running Blanca is left as it is.
func platformSettings(c *Config) []setting {
	login := func() bool { _, err := os.Stat(agentPath()); return err == nil }
	pick := func(int) {
		if login() {
			os.Remove(agentPath())
		} else if err := writeAgent(); err != nil {
			fmt.Fprintln(os.Stderr, "blanca:", err)
		}
	}
	if inApp() {
		login = func() bool { return bool(C.bzLogin()) }
		pick = func(int) { C.bzSetLogin(C.bool(!login())) }
	}
	return []setting{
		{name: "Launch on login", on: login, pick: pick},
		toggle("Sticky bezel", &c.Sticky),
	}
}

// cmdSetup installs the LaunchAgent and starts it now.
func cmdSetup() {
	p := agentPath()
	if err := writeAgent(); err != nil {
		fatal(err)
	}
	fmt.Println("add  " + p)
	gui := "gui/" + strconv.Itoa(os.Getuid())
	exec.Command("launchctl", "bootout", gui+"/"+label).Run() // replace a running watcher
	if out, err := exec.Command("launchctl", "bootstrap", gui, p).CombinedOutput(); err != nil {
		fatal(fmt.Errorf("launchctl bootstrap: %s", out))
	}
	fmt.Println("ok   clipboard watcher started")
	fmt.Println("done Allow blanca under Privacy & Security > Accessibility so it can paste,")
	fmt.Println("     then press Ctrl+Alt+V after copying something.")
}

func cmdMenu(Config) { fatal(errors.New("on macOS the menu is Blanca's menu bar icon")) }

func launchBezel(string) error {
	return errors.New("on macOS the bezel lives in `blanca watch`; run `blanca setup`")
}

func copyText(text string) error {
	s := C.CString(text)
	defer C.free(unsafe.Pointer(s))
	C.bzCopy(s)
	return nil
}

func clearClipboard() { C.bzClear() }

func sendPaste(Config) { C.bzPaste() }

// sensitive reports whether the pasteboard is marked as coming from a password manager.
func sensitive() bool { return bool(C.bzSensitive()) }
