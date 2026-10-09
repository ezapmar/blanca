package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
)

// setting is one line of the settings menu: a switch, or a number picked from nums.
// It edits the Config it was made for; saveConfig makes the change last.
type setting struct {
	name string
	on   func() bool // set for a switch
	pick func(i int) // flips a switch; takes nums[i] for a number
	num  *int
	nums []int
}

// value is how the setting reads in a list: on, off or its number.
func (s setting) value() string {
	if s.on == nil {
		return strconv.Itoa(*s.num)
	} else if s.on() {
		return "on"
	}
	return "off"
}

func toggle(name string, b *bool) setting {
	return setting{name: name, on: func() bool { return *b }, pick: func(int) { *b = !*b }}
}

func number(name string, n *int, nums ...int) setting {
	return setting{name: name, num: n, nums: nums, pick: func(i int) { *n = nums[i] }}
}

// settings lists what the menu offers.
func settings(c *Config) []setting {
	return append(platformSettings(c),
		toggle("Selection pastes", &c.Paste),
		toggle("Wraparound bezel", &c.Wraparound),
		toggle("Move clippings to top after use", &c.MoveToTop),
		toggle("Allow whitespace clippings", &c.AllowWhitespace),
		toggle("Ignore large clippings", &c.IgnoreLarge),
		toggle("Ignore confidential clipping types", &c.IgnoreSensitive),
		number("Clippings to remember", &c.Remember, 10, 25, 50, 99, 200),
		number("Clippings in menu", &c.Display, 5, 10, 15, 20, 25))
}

func saveConfig(c Config) error {
	p := configPath()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	b, _ := json.MarshalIndent(c, "", "  ")
	return os.WriteFile(p, append(b, '\n'), 0o644)
}
