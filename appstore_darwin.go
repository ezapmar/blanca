//go:build mas

package main

// #cgo CFLAGS: -DMAS
import "C"

const appStore = true

// The Store's Blanca says nothing of other systems: its reviewers turn that down.
const aboutWhat = "A clipboard manager for the Mac."
