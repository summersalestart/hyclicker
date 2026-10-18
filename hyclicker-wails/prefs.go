package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	MinCps     float64 `json:"min_cps"`
	MaxCps     float64 `json:"max_cps"`
	Button     string  `json:"button"`
	HotkeyVK   uint    `json:"hotkey_vk"`
	HotkeyMod  uint    `json:"hotkey_mod"`
	HotkeyName string  `json:"hotkey_name"`
}

const maxAllowedCPS = 50.0

var defaultConfig = Config{
	MinCps: 8, MaxCps: 12, Button: "Left",
	HotkeyVK: 0x75, HotkeyMod: 0, HotkeyName: "F6",
}

func (c Config) normalized() Config {
	if c.MinCps <= 0 {
		c.MinCps = defaultConfig.MinCps
	}
	if c.MaxCps < c.MinCps {
		c.MaxCps = c.MinCps
	}
	switch c.Button {
	case "Right", "Middle":
	default:
		c.Button = "Left"
	}
	if c.HotkeyVK == 0 {
		c.HotkeyVK = defaultConfig.HotkeyVK
		c.HotkeyMod = defaultConfig.HotkeyMod
		c.HotkeyName = defaultConfig.HotkeyName
	}
	return c
}

func readConfig(path string) Config {
	raw, err := os.ReadFile(path)
	if err != nil {
		return defaultConfig
	}
	var c Config
	if err := json.Unmarshal(raw, &c); err != nil {
		return defaultConfig
	}
	return c.normalized()
}

func writeConfig(path string, c Config) bool {
	c = c.normalized()
	raw, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return false
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return false
	}
	return os.WriteFile(path, raw, 0o644) == nil
}

func configFilePath() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "HyClicker", "config.json"), nil
}

func normalizeButton(name string) (string, error) {
	switch name {
	case "Left", "Right", "Middle":
		return name, nil
	default:
		return "", fmt.Errorf("unknown button: %s", name)
	}
}

func normalizeButtonE(name string) MouseButton {
	switch name {
	case "Right":
		return ButtonRight
	case "Middle":
		return ButtonMiddle
	default:
		return ButtonLeft
	}
}

func trimVersionPrefix(tag string) string {
	if i := strings.IndexFunc(tag, func(r rune) bool { return r >= '0' && r <= '9' }); i >= 0 {
		return tag[i:]
	}
	return tag
}

func isNewerVersion(candidate, current string) bool {
	c := versionSegments(candidate)
	t := versionSegments(current)
	for i := 0; i < 3; i++ {
		if c[i] != t[i] {
			return c[i] > t[i]
		}
	}
	return false
}

func versionSegments(s string) [3]int {
	var seg [3]int
	idx, num, has := 0, 0, false
	for i := 0; i < len(s) && idx < 3; i++ {
		ch := s[i]
		if ch >= '0' && ch <= '9' {
			num = num*10 + int(ch-'0')
			has = true
		} else if has {
			seg[idx], idx, num, has = num, idx+1, 0, false
		}
	}
	if has && idx < 3 {
		seg[idx] = num
	}
	return seg
}

func contextDone(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

var errNoStorePath = errors.New("config path unavailable")
