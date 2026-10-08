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
</p>

<p align="center">
  <a href="#install">Install</a> &nbsp;·&nbsp;
  <a href="#the-first-minute">First minute</a> &nbsp;·&nbsp;
  <a href="#configure">Configure</a> &nbsp;·&nbsp;
  <a href="#where-this-came-from">Jumpcut</a> &nbsp;·&nbsp;
  <a href="docs/design.md">Design</a>
</p>

Blanca is a clipboard manager for [Omarchy](https://omarchy.org). It is a port of
[Jumpcut](https://github.com/snark/jumpcut), the macOS clipboard manager Steve Cook has
kept alive since 2002, rebuilt in Go for Hyprland and Wayland. One static binary, no
daemon of its own, no account, nothing leaves your machine.

It is named after my dog. She is an Anatolian sighthound crossed with a Russell terrier,
and she has the same reaction to the word "paste".

---

## The problem

I copy a command from a terminal. Then I copy an error message to search for it. Now
the command is gone, and I go back to the first window to copy it again. Twenty times a
day, for twenty years, on every machine I have owned.

Jumpcut fixed this on my Mac in 2004 and I never thought about it again. Press
Ctrl+Alt+V, the last thing you copied appears in a small window. Press it again, the one
before that. Let go, it is pasted. That is the whole product, and I have not found a
better one.

Then I moved to Omarchy. Omarchy has a clipboard panel, and it is fine. A list you pick
from with a mouse is a different tool, though. Jumpcut is a keyboard reflex, and after
twenty years I wanted the reflex back.

## The solution

The same reflex, done the Omarchy way.

**A watcher that records.** `wl-paste --watch blanca store` runs from Hyprland's
autostart and hands Blanca every text copy. Blanca keeps the last 99 in a plain JSON
file, drops whitespace, anything over 50,000 characters, and anything a password
manager has tagged as secret.

**A bezel that shows one thing.** `Ctrl+Alt+V` opens a small floating terminal in the
middle of the screen with the newest clipping and its position, `1/42`. Press the hotkey
again to step back. Arrow keys, Home, End and the digits do what you expect. Backspace
deletes the clipping. Esc closes. Because it is a terminal, it already has your Omarchy
theme and transparency.

**Return pastes.** Blanca copies the clipping and, once the bezel has closed and focus is
back where you were, sends Shift+Insert through `wtype`. The same chord Omarchy's own
clipboard tools use, so it works in terminals too.

That is it. There is no tray icon and no preferences window. There is a JSON file with
nine keys.

## The first minute

Copy three things. Press `Ctrl+Alt+V`. You see the third. Press it again, the second.
Press Return. It is pasted where your cursor was, and the bezel is gone.

Keys in the bezel, the same as Jumpcut's:

| Key | Does |
|---|---|
| hotkey again, `↓`, `→` | next (older) clipping |
| `Ctrl+Alt+Shift+V`, `↑`, `←` | previous (newer) clipping |
| `PgDn`, `PgUp` | ten at a time |
| `Home`, `End` | newest, oldest |
| `1` to `9`, `0` | jump to that position, `0` is tenth |
| `Return` | copy and paste |
| `Backspace`, `Delete` | forget this clipping |
| `Esc`, `q` | close |

## Install

No Go toolchain needed. Blanca is a single static binary; the installer fetches the
latest release, puts it in `~/.local/bin`, and installs the icons and desktop entry:

```bash
curl -fsSL https://raw.githubusercontent.com/ezapmar/blanca/master/install.sh | sh
```

Yes, curl into sh. The script is forty lines and I would read it first.

Or build an Arch package from the release binary, which puts it under `/usr` and lets
`pacman` track it:

```bash
git clone https://github.com/ezapmar/blanca && cd blanca/packaging/aur && makepkg -si
```

Runtime dependencies are `wl-clipboard`, `wtype` and `xdg-terminal-exec`, all part of
Omarchy already. On plain Arch: `sudo pacman -S --needed wl-clipboard wtype`. To build
from source instead: `go install github.com/ezapmar/blanca@latest`.

Then three lines in your Hyprland config. Omarchy's config is Lua, so:

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

Reload, and start the watcher once by hand so you do not have to log out:

```bash
hyprctl reload && setsid wl-paste --type text --watch blanca store &
```

## Configure

`~/.config/blanca/config.json`. Every key is optional. These are the defaults:

```json
{
  "remember": 99,
  "display": 10,
  "wraparound": false,
  "paste": true,
  "paste_mode": "shift-insert",
  "move_to_top": false,
  "allow_whitespace": false,
  "ignore_large": true,
  "ignore_sensitive": true
}
```

`paste_mode` can be `ctrl-v` for the odd app that ignores Shift+Insert. `paste: false`
only copies and leaves the pasting to you. `move_to_top` puts a clipping back at the
top after you use it, which Jumpcut users either love or switch off in the first hour.

## Commands

```
blanca store          read one clipping from stdin into the history
blanca pick           show the bezel in the current terminal
blanca bezel [--up]   hotkey entry: advance an open bezel, or open one
blanca list [N]       print the first N clippings, shortened
blanca get N          print clipping N in full
blanca clear          forget all clippings
```

`list` and `get` are there so you can build a menu if you miss Jumpcut's status-bar
list. With fzf, for instance:

```bash
blanca get "$(blanca list 99 | fzf | cut -f1)" | wl-copy
```

## Where this came from

Jumpcut is a Swift app with a status-bar menu, a translucent bezel and twenty settings.
I read all of it before writing a line, and the notes are in [docs/PLAN.md](docs/PLAN.md):
what each file does, which behaviours matter, and what each one became on Linux. Most
of the mapping is boring in the good way. The pasteboard poller became `wl-paste`. The
plist became JSON. The global hotkey became a Hyprland bind. The bezel window became a
floating terminal, because Omarchy already knows how to launch and style those, and a
terminal needs no GTK and no compositor protocols.

One thing did not survive. Jumpcut pastes when you release the modifier keys. A
terminal cannot see a key going up, so Blanca pastes on Return instead. Everything else
in the bezel behaves the same.

Jumpcut is MIT licensed and its notice is reproduced in [LICENSE](LICENSE). Jumpcut
itself traces part of its lineage to Brent Simmons' TigerLaunch. Blanca is independent
and is not affiliated with the Jumpcut author or with Omarchy; the Jumpcut name and icon
belong to their owners and are not used here.

## Honest limits

- Text only. Images and files are not recorded.
- Tested on Omarchy with Hyprland. Other wlroots compositors should work with their own
  keybind and floating rule, but I have not tried them.
- Wayland hands Blanca the text, not the window it came from, so there is no
  per-application ignore list.
- Shift+Insert is what `wtype` sends by default. A few apps want Ctrl+V; that is one
  line in the config.
- The release is v0.1.0 and the wiring to a real desktop is a few days old. Expect edges.

## Contributing

Issues and pull requests are open. If the bezel opened in the wrong place, paste your
window rules. If a paste landed in the wrong window, say which app. Both are more useful
than a feature request. Tests run offline:

```bash
go test ./...
```

## Licence

MIT. See [LICENSE](LICENSE).

---

<p align="center">
  <img src="assets/png/blanca-64.png" alt="" width="40"><br>
  <sub>Copy twice. Lose nothing. Press the key.</sub>
</p>
