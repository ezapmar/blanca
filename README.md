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

> **New in 0.3: the Mac app opens like any other.** No Gatekeeper detour, and it
> updates itself from the menu.
> [What's new](#whats-new-in-03)

## Install

**Mac.** Download [Blanca.dmg](https://github.com/ezapmar/blanca/releases/latest/download/Blanca.dmg),
open it and drag Blanca to Applications. It needs macOS 13 or later, and asks once for
Accessibility permission, which is what lets it paste for you.

**Omarchy.** Download the [pacman package](https://github.com/ezapmar/blanca/releases/latest/download/blanca-x86_64.pkg.tar.zst)
and install it with `sudo pacman -U blanca-x86_64.pkg.tar.zst && blanca setup`.

Or paste this in a terminal, on either system:

```bash
curl -fsSL https://raw.githubusercontent.com/ezapmar/blanca/master/install.sh | sh
```

Then copy something and press `Ctrl+Alt+V`. On a Mac that is Control + Option + V.

To update, on either system: `blanca update`, or Check for Updates in the menu.
Other ways to install are in the fold further down.

## What's new in 0.3

- **No Open Anyway on a Mac.** The app and the `.dmg` open like any other.
- **Updates.** `blanca update`, or Check for Updates in the menu, installs the newest
  release.
- **Clear by age.** Clear the last hour, 24 hours, month, or all of it.

Since 0.2:

- **Blanca for macOS.** A native app with the same bezel and the same hotkey. Let go of
  the keys and it pastes.
- **A menu in the bar.** Click the dog for your newest clippings and Clear All: the menu
  bar on macOS, a Waybar button on Omarchy.
- **Settings in that menu.** Every option is a switch there, Launch on login included.
- **Copies from your iPhone.** On macOS they land in the list with a small phone beside
  them.
- **Installers.** The one line above installs `Blanca.app` on a Mac, and each release
  also carries a `.dmg`.

## What it is

Blanca is a clipboard manager for [Omarchy](https://omarchy.org) and macOS. Press the
hotkey and the last thing you copied appears in a small window. Press it again, the one before that. Return pastes it. No daemon of its own, no
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

On a Mac keyboard `Ctrl+Alt+V` is Control (⌃) + Option (⌥) + V. Alt is the Option key,
and it is Control, not Command. Hold Control and Option, tap `V` until the clipping you
want is showing, and let go: it pastes.

---

<details>
<summary><b>Other ways to install</b></summary>

On a Mac the terminal command puts `Blanca.app` in Applications and links the `blanca`
command into `~/.local/bin`.

A pacman package, so `pacman` tracks it. Every release ships `blanca-bin` for
x86_64 and aarch64; take the file for your machine from the
[latest release](https://github.com/ezapmar/blanca/releases/latest):

```bash
sudo pacman -U https://github.com/ezapmar/blanca/releases/download/v0.3.3/blanca-bin-0.3.3-1-x86_64.pkg.tar.zst && blanca setup
```

The same package can be built locally from `packaging/aur` with `makepkg -si`.

From source: `go install github.com/ezapmar/blanca@latest && blanca setup`.

Or build it from a clone. You need Go 1.26 or newer. On Omarchy:

```bash
git clone https://github.com/ezapmar/blanca && cd blanca
go build -o ~/.local/bin/blanca . && blanca setup
```

On macOS, with the Xcode command line tools installed (`xcode-select --install`),
`scripts/macapp` builds a universal `dist/Blanca.app` and a `dist/Blanca.dmg` beside it:

```bash
git clone https://github.com/ezapmar/blanca && cd blanca
scripts/macapp 0.3.0
cp -R dist/Blanca.app /Applications/ && open /Applications/Blanca.app
```

The app is signed ad hoc and targets macOS 13, the first release with the login item
API it uses. A build of your own is a new app as far as macOS is concerned, so it asks
for Accessibility again.

`scripts/macapp --store` builds the Mac App Store's Blanca into `dist/store`: the same
app in the App Sandbox, without Check for Updates, since the Store updates it. The
script's header lists the signing identities it takes.

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

**The bar has a menu.** Click the icon in the bar and the
newest clippings drop down, with Clear underneath. On Omarchy that is a Waybar
button which runs `blanca menu`, a Walker list. `blanca setup` adds the button.

**Clear goes back as far as you say.** The last hour, the last 24 hours, the last month,
or all of it. Blanca notes when each clipping was copied, as a hash beside the history.
Clippings from before it kept times have none, and only All forgets those.

**Copies from your iPhone are marked.** With Handoff on, what you copy on an iPhone or
iPad lands in the list like anything else, with a small phone beside it in the menu.

**On macOS it is one process.** `blanca watch` polls the pasteboard twice a second,
owns the hotkey and the menu bar icon, and draws the bezel as a native panel, all in one
Objective-C file behind cgo. Releasing the modifier keys pastes, and the paste is Cmd+V.
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
blanca clear [hour|day|month]
                      forget all clippings, or those of the last hour, 24 hours or month
blanca setup          Hyprland keybind, autostart, window rules and Waybar button; on macOS a LaunchAgent
blanca update         install the newest release, if there is one
```

`list` and `get` let you build a menu, with fzf for instance:

```bash
blanca get "$(blanca list 99 | fzf | cut -f1)" | wl-copy
```
</details>

<details>
<summary><b>Credits</b></summary>

Blanca is derived from [Jumpcut](https://github.com/snark/jumpcut) by Steve Cook, which
is MIT licensed; its notice is reproduced in [LICENSE](LICENSE). Blanca is independent
and not affiliated with its author or with Omarchy. The design notes are in
[docs/PLAN.md](docs/PLAN.md).
</details>

<details>
<summary><b>Honest limits</b></summary>

- Text only. Images and files are not recorded.
- Tested on Omarchy with Hyprland. Other wlroots compositors need their own keybind
  and floating rule.
- Wayland hands Blanca the text, not the window it came from, so there is no
  per-application ignore list.
- A few apps want Ctrl+V instead of Shift+Insert; that is one line in the config.
- Coming from 0.2 on a Mac, allow Blanca under Accessibility once more: macOS
  treats 0.3 as a new app. The hotkey there is fixed at
  Ctrl+Alt+V, so quit any other app that uses it first.
- iPhone copies reach a Mac only. That is Apple's Universal Clipboard, and Linux is not
  part of it.
- This is v0.3. Expect edges.
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
