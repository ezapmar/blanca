//go:build !mas

package main

/*
#include <stdlib.h>
void bzUpdating(const char *failed);
void bzUpdateFound(const char *tag);
*/
import "C"

import (
	"strings"
	"time"
	"unsafe"
)

// startAutoCheck is the daily look for a newer Blanca, in the one process that lives
// long enough to do it. It runs on a short tick and lets autoCheck decide when a day has
// passed, so a Mac that was asleep over the turn of the day still gets its check.
func startAutoCheck() {
	go func() {
		time.Sleep(time.Minute) // let the app settle before touching the network
		for {
			if tag := autoCheck(time.Now()); tag != "" {
				t := C.CString(tag)
				C.bzUpdateFound(t) // returns at once; the alert is the main thread's
				C.free(unsafe.Pointer(t))
			}
			time.Sleep(time.Hour)
		}
	}()
}

// goUpdateTitle is what the menu's update item says: the release the last check found,
// or the plain invitation to look. The caller frees it.
//
//export goUpdateTitle
func goUpdateTitle() *C.char {
	if tag := readUpdateState().pending(); tag != "" {
		return C.CString("Update to " + strings.TrimPrefix(tag, "v") + "…")
	}
	return C.CString("Check for Updates…")
}

// goUpdateCheck is the menu's Check for Updates, called off the main thread because it
// waits on the network. It returns the tag of a newer release, or NULL and a note to
// show instead. The caller frees both.
//
//export goUpdateCheck
func goUpdateCheck(note **C.char) *C.char {
	tag, n := checkUpdate()
	if tag == "" {
		*note = C.CString(n)
		return nil
	}
	return C.CString(tag)
}

// goUpdate installs release tag over this Blanca, which the installer stops and reopens.
//
//export goUpdate
func goUpdate(tag *C.char) {
	failed := func(why string) {
		s := C.CString(why)
		C.bzUpdating(s)
		C.free(unsafe.Pointer(s))
	}
	C.bzUpdating(nil)
	if err := startUpdate(C.GoString(tag), failed); err != nil {
		failed(err.Error())
	}
}
