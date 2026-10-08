# Blanca

Blanca is a clipboard manager for [Omarchy](https://omarchy.org) Linux. It
keeps a history of the text you have cut or copied, so you can get back to a
snippet, a URL, or that sentence you had _just right_ five minutes ago, even
after you have copied something else.

Blanca is a port of [Jumpcut](https://github.com/snark/jumpcut), the
long-running macOS clipboard manager by Steve Cook, rebuilt in Go for Wayland
and Hyprland. It is named after my dog.

## How it works

Jumpcut's signature is the **bezel**: press a hotkey and one clipping appears,
front and center. Press the hotkey again to step back through history, hit
Return to paste. Blanca keeps that exact interaction, Omarchy-style:

- `wl-paste --watch blanca store` records every text copy into
  `~/.local/share/blanca/history.json`.
- `Ctrl+Alt+V` runs `blanca bezel`, which opens the bezel as a small floating
  terminal (app id `org.omarchy.blanca`, so it picks up your Omarchy theme and
  transparency). Pressing the hotkey again while it is open advances it;
  `Ctrl+Alt+Shift+V` goes back.
- Return copies the clipping and pastes it into the window you came from with
  `wtype` (Shift+Insert, like Omarchy's own clipboard tools). Esc closes.

Keys in the bezel, same as Jumpcut: `↑ ↓ ← →` move, `PgUp PgDn` ±10,
`Home End`, `1`–`9` jump (`0` is tenth), `Return` select, `Backspace`/`Delete`
remove the clipping, `Esc` or `q` close.

## Install (Arch / Omarchy)

No Go toolchain needed. Blanca is a single static binary; the installer fetches
the latest release, puts it in `~/.local/bin`, and installs the icons and
desktop entry:

```bash
curl -fsSL https://raw.githubusercontent.com/ezapmar/blanca/master/install.sh | sh
```

Runtime dependencies are `wl-clipboard`, `wtype`, and `xdg-terminal-exec`, all
already part of Omarchy. On plain Arch:

```bash
sudo pacman -S --needed wl-clipboard wtype
```

To build from source instead: `go install github.com/ezapmar/blanca@latest`.

Then add to your Hyprland config:

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

Reload with `hyprctl reload` and start the watcher once by hand (or log out
and in):

```bash
setsid wl-paste --type text --watch blanca store &
```

## Commands

```
blanca store          read one clipping from stdin into the history
blanca pick           show the bezel in the current terminal
blanca bezel [--up]   hotkey entry: advance an open bezel, or open one
blanca list [N]       print the first N clippings, shortened (default 10)
blanca get N          print clipping N in full
blanca clear          forget all clippings
```

`list` and `get` let you build a menu with walker, fuzzel, or fzf if you want
Jumpcut's status-bar menu too:

```bash
blanca get "$(blanca list 99 | fzf | cut -f1)" | wl-copy
```

## Configuration

`~/.config/blanca/config.json`, every key optional. Defaults shown:

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

`paste_mode` can be `ctrl-v` for apps that do not accept Shift+Insert.
`ignore_sensitive` skips clippings tagged `x-kde-passwordManagerHint` by
password managers such as KeePassXC.

## Icon

The logo lives in `assets/`: `blanca.svg` is the app icon (white face, black
line art, transparent background, so it works on dark and light themes),
`blanca-symbolic.svg` is the single-colour toolbar/tray glyph for 16 to 32 px,
`blanca-tile.svg` is the same art on a rounded graphite tile for the launcher,
and `png/` holds rasterised sizes. To register Blanca in the app launcher:

```bash
install -Dm644 assets/blanca.svg ~/.local/share/icons/hicolor/scalable/apps/blanca.svg
install -Dm644 assets/blanca-tile.svg ~/.local/share/icons/hicolor/scalable/apps/blanca-tile.svg
install -Dm644 assets/blanca-symbolic.svg ~/.local/share/icons/hicolor/symbolic/apps/blanca-symbolic.svg
install -Dm644 assets/blanca.desktop ~/.local/share/applications/blanca.desktop
```

## Differences from Jumpcut

- No "release the modifier to paste": terminals do not see key-up events, so
  Return selects. Everything else in the bezel behaves the same.
- No status-bar menu, Sparkle updates, or preferences window.
- The bezel cursor starts at the newest clipping each time it opens.

See [docs/PLAN.md](docs/PLAN.md) for the reverse-engineering notes and the
full mapping from Jumpcut to Blanca.

## Credits

Blanca is derived from **Jumpcut** by Steve Cook, released under the MIT
License. Jumpcut's original copyright and permission notice are reproduced
in [LICENSE](LICENSE). Jumpcut itself traces part of its lineage to Brent
Simmons' TigerLaunch.

Blanca uses [golang.org/x/term](https://pkg.go.dev/golang.org/x/term) and,
at runtime, [wl-clipboard](https://github.com/bugaevc/wl-clipboard) and
[wtype](https://github.com/atx/wtype). It leans on Omarchy's TUI launcher for
its floating window.

Blanca is an independent project and is not affiliated with or endorsed by
the Jumpcut author or by Omarchy. The Jumpcut name and icon belong to their
respective owners and are not used here.

## License

MIT. See [LICENSE](LICENSE).
