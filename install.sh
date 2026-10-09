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
os=linux
if [ "$(uname -s)" = Darwin ]; then os=darwin arch=universal; fi
tag=${1:-$(curl -fsSL "https://api.github.com/repos/$repo/releases/latest" | grep -m1 '"tag_name"' | cut -d'"' -f4)}
[ -n "$tag" ] || { echo "blanca: could not determine latest release" >&2; exit 1; }
tmp=$(mktemp -d); trap 'rm -rf "$tmp"' EXIT
name="blanca-$tag-$os-$arch"
echo "Downloading $name"
curl -fsSL "https://github.com/$repo/releases/download/$tag/$name.tar.gz" | tar -xz -C "$tmp"
src="$tmp/$name"
mkdir -p "$HOME/.local/bin"
install -m755 "$src/blanca" "$HOME/.local/bin/blanca"
echo "Installed blanca $tag to ~/.local/bin"
if [ "$os" = linux ]; then
  data=${XDG_DATA_HOME:-$HOME/.local/share}
  install -Dm644 "$src/assets/blanca.svg" "$data/icons/hicolor/scalable/apps/blanca.svg"
  install -Dm644 "$src/assets/blanca-tile.svg" "$data/icons/hicolor/scalable/apps/blanca-tile.svg"
  install -Dm644 "$src/assets/blanca-symbolic.svg" "$data/icons/hicolor/symbolic/apps/blanca-symbolic.svg"
  install -Dm644 "$src/assets/blanca.desktop" "$data/applications/blanca.desktop"
  for dep in wl-paste wl-copy wtype xdg-terminal-exec; do
    command -v "$dep" >/dev/null 2>&1 || echo "Missing: $dep (sudo pacman -S --needed wl-clipboard wtype xdg-terminal-exec)"
  done
fi
"$HOME/.local/bin/blanca" setup
