//go:build !(darwin && mas)

package main

// appStore is set in the build for the Mac App Store, `-tags mas`. That Blanca runs in
// the App Sandbox and the Store updates it, so it leaves out Check for Updates, the move
// off the disk image and the LaunchAgent.
const appStore = false

// aboutWhat opens the menu's About.
const aboutWhat = "A clipboard manager for Omarchy Linux and macOS."
