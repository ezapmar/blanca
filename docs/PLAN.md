# Blanca plan

Port of Jumpcut (macOS, Swift) to Omarchy / Arch Linux (Hyprland, Wayland) in Go.
Goal: minimal, dependency-light, idiomatic to how Omarchy already launches TUIs.

## 1. Jumpcut, reverse engineered

Source: `Jumpcut/Jumpcut/*.swift` (0.84). Six moving parts.

**Pasteboard** (`Pasteboard.swift`). Polls `NSPasteboard.general.changeCount` every 0.5s.
On change, takes the first string item unless it is: whitespace-only (unless allowed),
a transient type (TextExpander etc.), a sensitive type (password managers, unless allowed),
> 50 000 chars (unless allowed), or marked with Jumpcut's own internal type (self-copy).
`set(text)` writes string + internal marker. `fakeCommandV()` posts a Cmd-V key event.

**ClippingStore / ClippingStack** (`Clippings.swift`). Array of strings, newest first,
capped at `rememberNum` (10..99, default 99). Persisted as XML plist `JCEngine.save` in
Application Support, rewritten on every mutation. Stack adds a cursor `position`.
Ops: add (insert at 0, truncate), deleteAt (cursor moves up if at/above), clear,
moveItemToTop, up/down (optional wraparound), move(±n) (clamped, no wrap).
New clipping: `bezelToTop=1` resets cursor to 0, else cursor shifts +1 to stay on same item.
Only adds if different from current top item.

**Bezel** (`Bezel.swift`, `Interactions.swift`). 325x325 translucent centered window showing
ONE clipping at a time plus its 1-based position. Shown by global hotkey (default Ctrl-Alt-V).
Keys while shown:
- hotkey again / Down / Right: next; Shift-hotkey / Up / Left: previous
- PgDn / PgUp: ±10 (clamped); Home: first; End: last; 1-9: jump, 0: tenth
- Return: select; Delete/Backspace: delete current (hide if empty); Esc: hide
- releasing all modifiers: select (unless `stickyBezel`)
Select = put on pasteboard; if `bezelSelectionPastes` (default true) also fake Cmd-V 0.2s later.
`moveClippingsAfterUse` moves the selected item to top.

**Menu** (`MenuManager.swift`, `StatusItem.swift`). Status-bar icon lists first `displayNum`
(default 10) clippings, shortened to first line, 40 chars + ellipsis. Click = select.
Alt menu (right/shift click) gives Copy / Paste / Delete per item. Fixed items: Clear All
(confirm), About, Preferences, Quit.

**Settings** (`Settings.swift`). UserDefaults keys and defaults:
allowWhitespaceClippings=false, askBeforeClearingClippings=true, bezelAlignment=center,
bezelSelectionPastes=true, bezelToTop=1, displayNum=10, hideStatusItem=false,
ignoreLargeClippings=true, ignoreSensitiveClippingTypes=true, launchOnStartup=false,
mainHotkey=Ctrl-Alt-V, menuSelectionPastes=true, moveClippingsAfterUse=false,
rememberNum=99, skipSave=false, stickyBezel=false, wraparoundBezel=false.

**Plumbing** dropped in the port: Sparkle updates, LaunchAtLogin, ShortcutRecorder,
Sauce keyboard-layout mapping, Accessibility prompt, Preferences window.

## 2. Mapping to Omarchy

| Jumpcut                    | Blanca                                                        |
|----------------------------|---------------------------------------------------------------|
| NSPasteboard polling       | `wl-paste --type text --watch blanca store` (Omarchy autostart)|
| internal marker type       | `$XDG_RUNTIME_DIR/blanca/placed` marker file, consumed once   |
| sensitive types            | skip if `wl-paste --list-types` has `x-kde-passwordManagerHint`|
| plist in App Support       | JSON array in `$XDG_DATA_HOME/blanca/history.json`, flock'd   |
| UserDefaults               | `$XDG_CONFIG_HOME/blanca/config.json`                         |
| Global hotkey (HotKey lib) | Hyprland `o.bind("CTRL + ALT + V", ..., "blanca bezel")`      |
| Bezel window               | TUI in a floating terminal via `omarchy-launch-tui blanca pick`|
| hotkey-again cycles        | `blanca bezel` finds running picker (pidfile) -> SIGUSR1/2    |
| release modifiers = select | not possible in a terminal; Return selects (Esc cancels)      |
| fakeCommandV               | `wl-copy` then detached `wtype -M shift -k Insert -m shift`   |
| Status-bar menu            | `blanca list` (pipe to walker/fuzzel) — bezel is the main UI  |
| Preferences window         | edit config.json                                              |

Why a TUI: zero GUI deps, pure Go static binary, cross-compiles from macOS, testable in any
terminal, and inherits the Omarchy theme/transparency for free (app-id `org.omarchy.blanca`
is already tagged as a terminal by Omarchy rules).

## 3. Commands

```
blanca store          stdin -> history (filters, dedup vs top, placed-marker)
blanca pick           TUI bezel in current terminal
blanca bezel [--up]   hotkey entry: signal running picker, else launch pick in a TUI window
blanca list [N]       first N shortened clippings, one per line (default display=10)
blanca get I          print clipping I (1-based) raw
blanca clear          empty history
blanca paste          internal: sleep 150ms, wtype paste chord
```

## 4. Config (`~/.config/blanca/config.json`, all optional)

```json
{ "remember": 99, "display": 10, "wraparound": false, "paste": true,
  "paste_mode": "shift-insert", "move_to_top": false, "allow_whitespace": false,
  "ignore_large": true, "ignore_sensitive": true, "new_to_top": true }
```
`paste_mode`: `shift-insert` (Omarchy default, works in terminals) | `ctrl-v`.

## 5. Hyprland (user `~/.config/hypr/`)

```lua
-- bindings.lua
o.bind("CTRL + ALT + V", "Blanca clipboard", "blanca bezel")
o.bind("CTRL + ALT + SHIFT + V", "Blanca clipboard (back)", "blanca bezel --up")
-- autostart.lua
o.launch_on_start("wl-paste --type text --watch blanca store")
-- hyprland.lua
o.window("org.omarchy.blanca", { float = true })
o.window("org.omarchy.blanca", { center = true })
o.window("org.omarchy.blanca", { size = { 640, 360 } })
```

## 6. Files

```
go.mod            module github.com/ezapmar/blanca; dep: golang.org/x/term
main.go           dispatch, paths, config
history.go        load/save (flock), add, delete, top, clear, shorten
store.go          store command
pick.go           TUI: raw mode, key parsing, signals, render
bezel.go          bezel/paste commands, wl-copy, terminal launch
history_test.go   stack semantics
```

## 7. Verification

- `go vet`, `go test` locally (macOS).
- `GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build` cross-compile.
- TUI smoke-tested in a macOS terminal with a seeded history.
- Real Omarchy test needed for: wl-paste watch, wtype paste, Hyprland rules, focus return.

## Decisions

- Module path: `github.com/ezapmar/blanca`, public repo (2026-10-08).
- `paste_mode` default stays `shift-insert`, matching Omarchy's own paste helper.
- Blanca keeps its own history file (`~/.local/share/blanca/history.json`) rather than
  sharing Omarchy's `clipboard-history.json`.

## Unresolved questions

1. Hyprland window rule keys (`float`, `center`, `size`) copied from Omarchy defaults;
   is `stay_focused` wanted/available in your Hyprland version?
2. Default hotkey Ctrl-Alt-V conflicts with nothing in Omarchy defaults as of today; OK?
3. Needs a real run on Omarchy: wl-paste watch, wtype paste, focus return after the bezel closes.
