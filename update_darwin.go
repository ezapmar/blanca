//go:build !mas

package main

/*
#include <stdlib.h>
void bzUpdating(const char *failed);
*/
import "C"

import "unsafe"

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
