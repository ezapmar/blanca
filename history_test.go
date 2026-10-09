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

func TestRemote(t *testing.T) {
	isolate(t)
	add := func(s string, remote bool) {
		withHistory(func(h *History) bool { return h.Add(s, 99) })
		if err := setRemote(s, remote); err != nil {
			t.Fatal(err)
		}
	}
	add("mac", false)
	add("phone", true)
	add("other", false)
	if m := readRemote(); len(m) != 1 || !m[hash("phone")] {
		t.Fatalf("only the phone clipping is remote: %v", m)
	}
	add("phone", false) // copied again on this machine
	if m := readRemote(); len(m) != 0 {
		t.Fatalf("a local copy is not remote: %v", m)
	}
	add("gone", true)
	withHistory(func(h *History) bool { return h.Delete("gone", 0) })
	add("next", false)
	if m := readRemote(); len(m) != 0 {
		t.Fatalf("deleted clippings are forgotten: %v", m)
	}
}
