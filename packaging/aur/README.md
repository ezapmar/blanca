# AUR package: blanca-bin

Installs the prebuilt release binary, icons and desktop entry. No Go toolchain needed.

Local build and install, without the AUR:

```bash
cd packaging/aur && makepkg -si
```

Publishing to the AUR (once, needs an AUR account with an SSH key):

```bash
git clone ssh://aur@aur.archlinux.org/blanca-bin.git /tmp/blanca-bin
cp PKGBUILD .SRCINFO /tmp/blanca-bin/ && cd /tmp/blanca-bin
git add PKGBUILD .SRCINFO && git commit -m "Update to 0.1.0" && git push
```

On each release: bump `pkgver`, paste the two checksums from the release's
`SHA256SUMS`, regenerate `.SRCINFO` with `makepkg --printsrcinfo > .SRCINFO`, push.
