package main

import (
	"reflect"
	"testing"
)

func TestHistory(t *testing.T) {
	h := &History{}
	for _, s := range []string{"a", "b", "b", "c"} {
		h.Add(s, 10)
	}
	if want := []string{"c", "b", "a"}; !reflect.DeepEqual(h.Items, want) {
		t.Fatalf("add/dedup: %v", h.Items)
	}
	h.Add("d", 3)
	if want := []string{"d", "c", "b"}; !reflect.DeepEqual(h.Items, want) {
		t.Fatalf("cap: %v", h.Items)
	}
	if !h.ToTop("b", 2) || h.Items[0] != "b" || len(h.Items) != 3 {
		t.Fatalf("totop: %v", h.Items)
	}
	if h.ToTop("b", 0) {
		t.Fatal("totop of top should be a no-op")
	}
	if !h.Delete("c", 0) || !reflect.DeepEqual(h.Items, []string{"b", "d"}) { // wrong hint, found by value
		t.Fatalf("delete: %v", h.Items)
	}
	if h.Delete("zzz", 0) {
		t.Fatal("delete missing")
	}
}

func TestShorten(t *testing.T) {
	if got := shorten("  hello\nworld", 40); got != "hello" {
		t.Fatal(got)
	}
	if got := shorten("ééééé", 3); got != "ééé…" {
		t.Fatal(got)
	}
}

func TestWrapAndKeys(t *testing.T) {
	if got := wrap("abcdef\tx", 4); !reflect.DeepEqual(got, []string{"abcd", "ef  ", "  x"}) {
		t.Fatal(got)
	}
	if parseKey("\x1b[B").k != kNext || parseKey("\x1b").k != kQuit || parseKey("7").n != 7 {
		t.Fatal("keys")
	}
}
