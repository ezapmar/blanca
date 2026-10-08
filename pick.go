package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"golang.org/x/term"
)

type key int

const (
	kNone key = iota
	kNext
	kPrev
	kPgDn
	kPgUp
	kHome
	kEnd
	kEnter
	kDelete
	kQuit
	kRedraw
	kDigit
)

type event struct {
	k key
	n int // digit for kDigit
}

// cmdPick is the bezel: one clipping at a time, driven by Jumpcut's keys.
func cmdPick(cfg Config) {
	items, err := readHistory()
	if err != nil {
		fatal(err)
	}
	if len(items) == 0 {
		return
	}
	pidfile := filepath.Join(runtimeDir(), "pick.pid")
	os.WriteFile(pidfile, []byte(strconv.Itoa(os.Getpid())), 0o600)
	defer os.Remove(pidfile)

	fd := int(os.Stdin.Fd())
	old, err := term.MakeRaw(fd)
	if err != nil {
		fatal(err)
	}
	os.Stdout.WriteString("\x1b[?1049h\x1b[?25l")
	defer func() {
		os.Stdout.WriteString("\x1b[?25h\x1b[?1049l")
		term.Restore(fd, old)
	}()

	ev := make(chan event)
	go readKeys(ev)
	sigs := make(chan os.Signal, 4)
	signal.Notify(sigs, syscall.SIGUSR1, syscall.SIGUSR2, syscall.SIGWINCH, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		for s := range sigs {
			switch s {
			case syscall.SIGUSR1:
				ev <- event{k: kNext}
			case syscall.SIGUSR2:
				ev <- event{k: kPrev}
			case syscall.SIGWINCH:
				ev <- event{k: kRedraw}
			default:
				ev <- event{k: kQuit}
			}
		}
	}()

	pos := 0
	for {
		render(items, pos, cfg.Paste)
		e := <-ev
		switch e.k {
		case kNext, kPrev, kPgDn, kPgUp, kHome, kEnd, kDigit:
			pos = move(pos, len(items), e, cfg.Wraparound)
		case kEnter:
			cur := items[pos]
			place(cfg, cur)
			if cfg.MoveToTop {
				withHistory(func(h *History) bool { return h.ToTop(cur, pos) })
			}
			return
		case kDelete:
			cur := items[pos]
			items, _ = withHistory(func(h *History) bool { return h.Delete(cur, pos) })
			if len(items) == 0 {
				return
			}
			if pos > 0 { // Jumpcut moves up after deleting the current item
				pos--
			}
			pos = min(pos, len(items)-1)
		case kQuit:
			return
		}
	}
}

// move applies one navigation event to the cursor, with Jumpcut's rules: up/down may
// wrap when enabled, page moves clamp, digits are one-based positions and 0 is tenth.
func move(pos, n int, e event, wrap bool) int {
	switch e.k {
	case kNext:
		if pos+1 < n {
			return pos + 1
		} else if wrap {
			return 0
		}
	case kPrev:
		if pos > 0 {
			return pos - 1
		} else if wrap {
			return n - 1
		}
	case kPgDn:
		return min(pos+10, n-1)
	case kPgUp:
		return max(pos-10, 0)
	case kHome:
		return 0
	case kEnd:
		return n - 1
	case kDigit:
		d := e.n
		if d == 0 {
			d = 10
		}
		return min(d, n) - 1
	}
	return pos
}

func readKeys(ev chan<- event) {
	buf := make([]byte, 32)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil {
			ev <- event{k: kQuit}
			return
		}
		ev <- parseKey(string(buf[:n]))
	}
}

func parseKey(s string) event {
	switch s {
	case "\x1b", "q", "\x03":
		return event{k: kQuit}
	case "\r", "\n":
		return event{k: kEnter}
	case "\x7f", "\x08", "\x1b[3~":
		return event{k: kDelete}
	case "\x1b[A", "\x1b[D", "\x1bOA", "\x1bOD":
		return event{k: kPrev}
	case "\x1b[B", "\x1b[C", "\x1bOB", "\x1bOC":
		return event{k: kNext}
	case "\x1b[5~":
		return event{k: kPgUp}
	case "\x1b[6~":
		return event{k: kPgDn}
	case "\x1b[H", "\x1b[1~", "\x1b[7~", "\x1bOH":
		return event{k: kHome}
	case "\x1b[F", "\x1b[4~", "\x1b[8~", "\x1bOF":
		return event{k: kEnd}
	}
	if len(s) == 1 && s[0] >= '0' && s[0] <= '9' {
		return event{k: kDigit, n: int(s[0] - '0')}
	}
	return event{}
}

func render(items []string, pos int, paste bool) {
	w, h, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || w < 12 || h < 6 {
		w, h = 80, 24
	}
	action := "copy"
	if paste {
		action = "paste"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "\x1b[H\x1b[2J\x1b[1m Blanca  %d/%d\x1b[0m\r\n\r\n", pos+1, len(items))
	lines := wrap(items[pos], w-4)
	for i, l := range lines {
		if i >= h-5 {
			b.WriteString("  \x1b[2m…\x1b[0m\r\n")
			break
		}
		b.WriteString("  " + l + "\r\n")
	}
	fmt.Fprintf(&b, "\x1b[%d;1H\x1b[2m ↑↓ move  1-9 jump  ⏎ %s  ⌫ delete  esc close\x1b[0m", h, action)
	os.Stdout.WriteString(b.String())
}

// wrap breaks text into lines of at most w runes, like Jumpcut's clipping text view.
func wrap(s string, w int) []string {
	s = strings.NewReplacer("\r", "", "\t", "    ").Replace(s)
	var out []string
	for _, line := range strings.Split(s, "\n") {
		r := []rune(line)
		for len(r) > w {
			out = append(out, string(r[:w]))
			r = r[w:]
		}
		out = append(out, string(r))
	}
	return out
}
