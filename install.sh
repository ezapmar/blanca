#!/bin/sh
# Install Blanca on Omarchy / Arch Linux or macOS without a Go toolchain.
# Usage: curl -fsSL https://raw.githubusercontent.com/ezapmar/blanca/master/install.sh | sh
#        (optionally: sh install.sh v0.1.0 to pin a version)
set -eu
repo=ezapmar/blanca
case "$(uname -m)" in
  x86_64) arch=amd64 ;;
  aarch64|arm64) arch=arm64 ;;
  *) echo "blanca: unsupported architecture $(uname -m)" >&2; exit 1 ;;
esac
tag=${1:-$(curl -fsSL "https://api.github.com/repos/$repo/releases/latest" | grep -m1 '"tag_name"' | cut -d'"' -f4)}
[ -n "$tag" ] || { echo "blanca: could not determine latest release" >&2; exit 1; }
tmp=$(mktemp -d); trap 'rm -rf "$tmp"' EXIT
base="https://github.com/$repo/releases/download/$tag"

if [ "$(uname -s)" = Darwin ]; then
  # Blanca.app into /Applications, the command line as a link into it. Fetched with
  # curl the app is not quarantined, so it opens without a Gatekeeper detour.
  apps=/Applications; [ -w "$apps" ] || apps="$HOME/Applications"
  echo "Downloading Blanca-$tag.zip"
  curl -fsSL "$base/Blanca-$tag.zip" -o "$tmp/Blanca.zip"
  ditto -x -k "$tmp/Blanca.zip" "$tmp"
  launchctl bootout "gui/$(id -u)/com.github.ezapmar.blanca" 2>/dev/null || true
  pkill -x blanca 2>/dev/null || true
  mkdir -p "$apps" "$HOME/.local/bin"
  rm -rf "$apps/Blanca.app"
  mv "$tmp/Blanca.app" "$apps/Blanca.app"
  ln -sf "$apps/Blanca.app/Contents/MacOS/blanca" "$HOME/.local/bin/blanca"
  open "$apps/Blanca.app"
  echo "Installed Blanca $tag to $apps and started it: look for the dog in the menu bar."
  echo "Allow Blanca under Privacy & Security > Accessibility so it can paste (again after"
  echo "each update), and switch on Launch on login under Settings in its menu."
  exit 0
fi

name="blanca-$tag-linux-$arch"
echo "Downloading $name"
curl -fsSL "$base/$name.tar.gz" | tar -xz -C "$tmp"
src="$tmp/$name"
mkdir -p "$HOME/.local/bin"
install -m755 "$src/blanca" "$HOME/.local/bin/blanca"
echo "Installed blanca $tag to ~/.local/bin"
data=${XDG_DATA_HOME:-$HOME/.local/share}
install -Dm644 "$src/assets/blanca.svg" "$data/icons/hicolor/scalable/apps/blanca.svg"
install -Dm644 "$src/assets/blanca-tile.svg" "$data/icons/hicolor/scalable/apps/blanca-tile.svg"
install -Dm644 "$src/assets/blanca-symbolic.svg" "$data/icons/hicolor/symbolic/apps/blanca-symbolic.svg"
install -Dm644 "$src/assets/blanca.desktop" "$data/applications/blanca.desktop"
for dep in wl-paste wl-copy wtype xdg-terminal-exec; do
  command -v "$dep" >/dev/null 2>&1 || echo "Missing: $dep (sudo pacman -S --needed wl-clipboard wtype xdg-terminal-exec)"
done
"$HOME/.local/bin/blanca" setup
