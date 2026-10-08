package main

import "testing"

func TestMove(t *testing.T) {
	const n = 12
	cases := []struct {
		name string
		pos  int
		e    event
		wrap bool
		want int
	}{
		{"next", 0, event{k: kNext}, false, 1},
		{"next at end stays", n - 1, event{k: kNext}, false, n - 1},
		{"next at end wraps", n - 1, event{k: kNext}, true, 0},
		{"prev", 5, event{k: kPrev}, false, 4},
		{"prev at start stays", 0, event{k: kPrev}, false, 0},
		{"prev at start wraps", 0, event{k: kPrev}, true, n - 1},
		{"pgdn clamps", 5, event{k: kPgDn}, true, n - 1},
		{"pgup clamps", 5, event{k: kPgUp}, true, 0},
		{"pgdn", 0, event{k: kPgDn}, false, 10},
		{"home", 7, event{k: kHome}, false, 0},
		{"end", 7, event{k: kEnd}, false, n - 1},
		{"digit 3", 0, event{k: kDigit, n: 3}, false, 2},
		{"digit 0 is tenth", 0, event{k: kDigit, n: 0}, false, 9},
		{"digit beyond count clamps", 0, event{k: kDigit, n: 9}, false, 4},
		{"redraw keeps", 4, event{k: kRedraw}, false, 4},
	}
	for _, c := range cases {
		size := n
		if c.name == "digit beyond count clamps" {
			size = 5
		}
		if got := move(c.pos, size, c.e, c.wrap); got != c.want {
			t.Errorf("%s: got %d want %d", c.name, got, c.want)
		}
	}
}

func TestParseKey(t *testing.T) {
	cases := map[string]key{
		"\x1b": kQuit, "q": kQuit, "\x03": kQuit,
		"\r": kEnter, "\n": kEnter,
		"\x7f": kDelete, "\x08": kDelete, "\x1b[3~": kDelete,
		"\x1b[A": kPrev, "\x1b[D": kPrev, "\x1bOA": kPrev,
		"\x1b[B": kNext, "\x1b[C": kNext, "\x1bOB": kNext,
		"\x1b[5~": kPgUp, "\x1b[6~": kPgDn,
		"\x1b[H": kHome, "\x1b[1~": kHome, "\x1bOH": kHome,
		"\x1b[F": kEnd, "\x1b[4~": kEnd, "\x1bOF": kEnd,
		"x": kNone, "\x1b[Z": kNone,
	}
	for in, want := range cases {
		if got := parseKey(in).k; got != want {
			t.Errorf("%q: got %v want %v", in, got, want)
		}
	}
	for d := 0; d <= 9; d++ {
		e := parseKey(string(rune('0' + d)))
		if e.k != kDigit || e.n != d {
			t.Errorf("digit %d parsed as %+v", d, e)
		}
	}
}

func TestWrap(t *testing.T) {
	cases := []struct {
		in   string
		w    int
		want []string
	}{
		{"abcdef", 4, []string{"abcd", "ef"}},
		{"a\r\nb", 10, []string{"a", "b"}},
		{"\tx", 10, []string{"    x"}},
		{"", 10, []string{""}},
		{"ééééé", 2, []string{"éé", "éé", "é"}},
	}
	for _, c := range cases {
		got := wrap(c.in, c.w)
		if len(got) != len(c.want) {
			t.Errorf("%q: %q", c.in, got)
			continue
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Errorf("%q: %q", c.in, got)
			}
		}
	}
}
