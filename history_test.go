package main

import (
	"reflect"
	"testing"
	"time"
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

func TestDeleteSince(t *testing.T) {
	// Newest first. "again" was copied long ago and once more just now; "old" has no time.
	h := &History{Items: []string{"now", "again", "today", "week", "again", "old"}}
	times := map[string]int64{hash("now"): 1000, hash("again"): 990, hash("today"): 500, hash("week"): 100}
	if !h.DeleteSince(times, 900) || !reflect.DeepEqual(h.Items, []string{"today", "week", "again", "old"}) {
		t.Fatalf("last hour: %v", h.Items)
	}
	if h.DeleteSince(times, 900) {
		t.Fatal("nothing recent is left to delete")
	}
	if !h.DeleteSince(times, 0) || !reflect.DeepEqual(h.Items, []string{"again", "old"}) {
		t.Fatalf("clippings without a time stay: %v", h.Items)
	}
}

func TestTimes(t *testing.T) {
	isolate(t)
	at := time.Unix(1000, 0)
	for _, s := range []string{"a", "b"} {
		withHistory(func(h *History) bool { return h.Add(s, 99) })
		if err := setTime(s, at); err != nil {
			t.Fatal(err)
		}
		at = at.Add(time.Minute)
	}
	if m := readTimes(); len(m) != 2 || m[hash("a")] != 1000 || m[hash("b")] != 1060 {
		t.Fatalf("times: %v", m)
	}
	withHistory(func(h *History) bool { return h.Delete("a", 1) })
	withHistory(func(h *History) bool { return h.Add("c", 99) })
	setTime("c", at)
	if m := readTimes(); len(m) != 2 || m[hash("a")] != 0 {
		t.Fatalf("deleted clippings are forgotten: %v", m)
	}
}
