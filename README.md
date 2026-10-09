<p align="center">
  <img src="assets/blanca-banner.png" alt="Blanca" width="760">
</p>

<p align="center">
  <b>Everything you copy, kept. One hotkey brings it back, one clipping at a time,<br>
  and Return pastes it where you were.</b>
</p>

<p align="center">
  <a href="https://github.com/ezapmar/blanca/releases"><img alt="release" src="https://img.shields.io/github/v/release/ezapmar/blanca?style=flat-square&labelColor=262626&color=F24B1E&label=release"></a>
  <a href="LICENSE"><img alt="licence" src="https://img.shields.io/badge/licence-MIT-0C9794?style=flat-square&labelColor=262626"></a>
  <img alt="go" src="https://img.shields.io/badge/go-static%20binary-FBA335?style=flat-square&labelColor=262626">
  <img alt="omarchy" src="https://img.shields.io/badge/omarchy-hyprland%20%2B%20wayland-F9EBDB?style=flat-square&labelColor=262626">
  <img alt="macos" src="https://img.shields.io/badge/macos-13%2B-F9EBDB?style=flat-square&labelColor=262626">
</p>

> **New in 0.2: the macOS version is here.** A menu bar app with the same bezel and
> hotkey, your clippings and the settings one click away, and a small phone beside
> whatever you copied on your iPhone. [What's new](#whats-new-in-02)

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/ezapmar/blanca/master/install.sh | sh
```

That is all of it. The script puts one static binary in `~/.local/bin`, installs the
icons, adds the keybind, autostart and window rules to your Hyprland config, reloads it
and starts the clipboard watcher. Copy something, press `Ctrl+Alt+V`.

The same line works on macOS. There it puts `Blanca.app` in Applications, starts it and
links the `blanca` command into `~/.local/bin`. macOS asks once for Accessibility
permission, which is what lets Blanca paste for you. Or take `Blanca.dmg` from the
[latest release](https://github.com/ezapmar/blanca/releases/latest) and drag the app
across; it is not notarised, so the first time macOS makes you allow it under
Privacy & Security > Open Anyway.

## What's new in 0.2

- **Blanca for macOS.** A native app with the same bezel and the same hotkey. Let go of
  the keys and it pastes, as in Jumpcut.
- **A menu in the bar.** Click the dog for your newest clippings and Clear All: the menu
  bar on macOS, a Waybar button on Omarchy.
- **Settings in that menu.** Every option is a switch there, Launch on login included.
- **Copies from your iPhone.** On macOS they land in the list with a small phone beside
  them.
- **Installers.** The one line above installs `Blanca.app` on a Mac, and each release
  also carries a `.dmg`.

## What it is

Blanca is a clipboard manager for [Omarchy](https://omarchy.org) and macOS, ported from
[Jumpcut](https://github.com/snark/jumpcut), the macOS clipboard manager Steve Cook has
kept alive since 2002. Press the hotkey and the last thing you copied appears in a small
window. Press it again, the one before that. Return pastes it. No daemon of its own, no
account, nothing leaves your machine.

It is named after my dog, an Anatolian sighthound crossed with a Russell terrier. She
has the same reaction to the word "paste".

| Key | Does |
|---|---|
| hotkey again, `↓`, `→` | older clipping |
| `Ctrl+Alt+Shift+V`, `↑`, `←` | newer clipping |
| `PgDn`, `PgUp`, `Home`, `End` | ten at a time, newest, oldest |
| `1` to `9`, `0` | jump to that position, `0` is tenth |
| `Return` | copy and paste |
| `Backspace`, `Delete` | forget this clipping |
| `Esc`, `q` | close |

---

<details>
<summary><b>Other ways to install</b></summary>

A pacman package, so `pacman` tracks it. Every release ships `blanca-bin` for
x86_64 and aarch64; take the file for your machine from the
[latest release](https://github.com/ezapmar/blanca/releases/latest):

```bash
sudo pacman -U https://github.com/ezapmar/blanca/releases/download/v0.1.4/blanca-bin-0.1.4-1-x86_64.pkg.tar.zst && blanca setup
```

The same package can be built locally from `packaging/aur` with `makepkg -si`.

From source: `go install github.com/ezapmar/blanca@latest && blanca setup`.

Runtime dependencies are `wl-clipboard`, `wtype` and `xdg-terminal-exec`, all part of
Omarchy. On plain Arch: `sudo pacman -S --needed wl-clipboard wtype xdg-terminal-exec`.

`blanca setup` appends these lines to your Hyprland config, skipping any file that
already mentions blanca. If you would rather add them yourself:

```lua
-- ~/.config/hypr/bindings.lua
o.bind("CTRL + ALT + V", "Blanca clipboard", "blanca bezel")
o.bind("CTRL + ALT + SHIFT + V", "Blanca clipboard (back)", "blanca bezel --up")

-- ~/.config/hypr/autostart.lua
o.launch_on_start("wl-paste --type text --watch blanca store")

-- ~/.config/hypr/hyprland.lua
o.window("org.omarchy.blanca", { float = true })
o.window("org.omarchy.blanca", { center = true })
o.window("org.omarchy.blanca", { size = { 640, 360 } })
```
</details>

<details>
<summary><b>How it works</b></summary>

**A watcher records.** `wl-paste --watch blanca store` hands Blanca every text copy.
It keeps the last 99 in `~/.local/share/blanca/history.json`, dropping whitespace,
anything over 50,000 characters, and anything a password manager tagged as secret.

**A bezel shows one thing.** `Ctrl+Alt+V` opens a small floating terminal with the
newest clipping and its position. Pressing the hotkey again steps back, because the
bezel is already open and the second press just signals it. A terminal means it already
has your Omarchy theme and transparency, and needs no GTK.

**Return pastes.** Blanca copies the clipping and, once the bezel has closed and focus
is back where you were, sends Shift+Insert through `wtype`, the same chord Omarchy's
own clipboard tools use, so it works in terminals too.

**The bar has a menu.** Jumpcut's other half is its menu bar icon: click it and the
newest clippings drop down, with Clear All underneath. On Omarchy that is a Waybar
button which runs `blanca menu`, a Walker list. `blanca setup` adds the button.

**Copies from your iPhone are marked.** With Handoff on, what you copy on an iPhone or
iPad lands in the list like anything else, with a small phone beside it in the menu.

**On macOS it is one process.** `blanca watch` polls the pasteboard twice a second,
owns the hotkey and the menu bar icon, and draws the bezel as a native panel, all in one
Objective-C file behind cgo. Releasing the modifier keys pastes, as in Jumpcut, and the
paste is Cmd+V.
</details>

<details>
<summary><b>Configure</b></summary>

`~/.config/blanca/config.json`. Every key is optional. These are the defaults:

```json
{
  "remember": 99,
  "display": 10,
  "wraparound": false,
  "paste": true,
  "paste_mode": "shift-insert",
  "sticky": false,
  "move_to_top": false,
  "allow_whitespace": false,
  "ignore_large": true,
  "ignore_sensitive": true
}
```

Every one of these is also under Settings in the menu: a submenu of the menu bar icon
on macOS, an entry in `blanca menu` on Omarchy. Changing one there rewrites this file.

`paste_mode` can be `ctrl-v` for the odd app that ignores Shift+Insert. `paste: false`
only copies. `move_to_top` puts a clipping back at the top after you use it. `sticky`
is for macOS: the bezel stays open when you let go of the modifiers, until Return or Esc.
</details>

<details>
<summary><b>Commands</b></summary>

```
blanca store          read one clipping from stdin into the history
blanca watch          record every text copy; on macOS also the hotkey and the bezel
blanca pick           show the bezel in the current terminal
blanca bezel [--up]   hotkey entry: advance an open bezel, or open one
blanca menu           Waybar button entry: the newest clippings as a Walker menu
blanca list [N]       print the first N clippings, shortened
blanca get N          print clipping N in full
blanca clear          forget all clippings
blanca setup          Hyprland keybind, autostart, window rules and Waybar button; on macOS a LaunchAgent
```

`list` and `get` let you build a menu, with fzf for instance:

```bash
blanca get "$(blanca list 99 | fzf | cut -f1)" | wl-copy
```
</details>

<details>
<summary><b>Where this came from</b></summary>

Jumpcut is a Swift app with a status-bar menu, a translucent bezel and twenty settings.
I read all of it before writing a line; the notes and the mapping to Linux are in
[docs/PLAN.md](docs/PLAN.md). The pasteboard poller became `wl-paste`, the plist became
JSON, the global hotkey became a Hyprland bind, the bezel became a floating terminal.

One thing did not survive. Jumpcut pastes when you release the modifier keys. A
terminal cannot see a key going up, so Blanca pastes on Return.

Jumpcut is MIT licensed and its notice is reproduced in [LICENSE](LICENSE). Jumpcut
itself traces part of its lineage to Brent Simmons' TigerLaunch. Blanca is independent
and not affiliated with the Jumpcut author or with Omarchy; the Jumpcut name and icon
belong to their owners and are not used here.
</details>

<details>
<summary><b>Honest limits</b></summary>

- Text only. Images and files are not recorded.
- Tested on Omarchy with Hyprland. Other wlroots compositors need their own keybind
  and floating rule.
- Wayland hands Blanca the text, not the window it came from, so there is no
  per-application ignore list.
- A few apps want Ctrl+V instead of Shift+Insert; that is one line in the config.
- macOS builds are not signed with a Developer ID, so the Accessibility permission
  has to be granted again after each update. The hotkey there is fixed at Ctrl+Alt+V,
  the same one Jumpcut uses, so quit Jumpcut first.
- iPhone copies reach a Mac only. That is Apple's Universal Clipboard, and Linux is not
  part of it.
- This is v0.2. Expect edges.
</details>

<details>
<summary><b>Design and contributing</b></summary>

The mark, palette and logo files are in [docs/design.md](docs/design.md).

Issues and pull requests are open. If the bezel opened in the wrong place, paste your
window rules. If a paste landed in the wrong window, say which app. Tests run offline
with `go test ./...`.
</details>

## Licence

MIT. See [LICENSE](LICENSE).

---

<p align="center">
  <img src="assets/png/blanca-64.png" alt="" width="40"><br>
  <sub>Copy twice. Lose nothing. Press the key.</sub>
</p>
