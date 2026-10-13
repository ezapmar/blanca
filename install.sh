#!/bin/sh
# Install Blanca on Omarchy / Arch Linux or macOS without a Go toolchain.
# Usage: sh install.sh v0.3.8, the release to install. The README gives the line that
#        fetches this file from that release's tag and checks its SHA-256 before running it.
#
# `blanca update` runs the copy of this file built into the binary, as `sh -c`, with the
# tag of the release to install: so an older Blanca's copy must be able to install a
# newer release. Keep the names of the release files it downloads stable.
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

# Nothing is unpacked or run before its SHA-256 matches the one published with the release.
digest() {
  if command -v sha256sum >/dev/null 2>&1; then sha256sum "$1" | cut -d' ' -f1
  elif command -v shasum >/dev/null 2>&1; then shasum -a 256 "$1" | cut -d' ' -f1
  else echo "blanca: need sha256sum or shasum to verify the download" >&2; exit 1
  fi
}
verify() { # verify <file> <expected sha-256>
  [ -n "$2" ] || {
    echo "blanca: no published checksum for $(basename "$1"), refusing to install it" >&2
    echo "  releases before v0.3.7 do not carry one; install $tag from its package instead" >&2
    exit 1
  }
  got=$(digest "$1")
  [ "$got" = "$2" ] || {
    echo "blanca: $(basename "$1") does not match the checksum in the release" >&2
    echo "  expected $2" >&2
    echo "  got      $got" >&2
    exit 1
  }
}

if [ "$(uname -s)" = Darwin ]; then
  # Blanca.app into /Applications, the command line as a link into it. Fetched with
  # curl the app is not quarantined, so it opens without a Gatekeeper detour.
  apps=/Applications; [ -w "$apps" ] || apps="$HOME/Applications"
  echo "Downloading Blanca-$tag.zip"
  curl -fsSL "$base/Blanca-$tag.zip" -o "$tmp/Blanca.zip"
  want=$(curl -fsSL "$base/Blanca-$tag.zip.sha256" 2>/dev/null | cut -d' ' -f1 || true)
  verify "$tmp/Blanca.zip" "$want"
  ditto -x -k "$tmp/Blanca.zip" "$tmp"
  launchctl bootout "gui/$(id -u)/com.github.ezapmar.blanca" 2>/dev/null || true
  pkill -x blanca 2>/dev/null || true
  mkdir -p "$apps" "$HOME/.local/bin"
  rm -rf "$apps/Blanca.app"
  mv "$tmp/Blanca.app" "$apps/Blanca.app"
  ln -sf "$apps/Blanca.app/Contents/MacOS/blanca" "$HOME/.local/bin/blanca"
  open "$apps/Blanca.app" --args watch installed # it says so itself: an update from the menu has no terminal
  echo "Installed Blanca $tag to $apps and started it: look for the dog in the menu bar."
  echo "Allow Blanca under Privacy & Security > Accessibility so it can paste, and switch"
  echo "on Launch on login under Settings in its menu."
  exit 0
fi

name="blanca-$tag-linux-$arch"
echo "Downloading $name"
curl -fsSL "$base/$name.tar.gz" -o "$tmp/$name.tar.gz"
want=$(curl -fsSL "$base/SHA256SUMS" 2>/dev/null | grep " $name\.tar\.gz\$" | cut -d' ' -f1 || true)
verify "$tmp/$name.tar.gz" "$want"
tar -xzf "$tmp/$name.tar.gz" -C "$tmp"
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
