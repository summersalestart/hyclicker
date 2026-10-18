package main

import (
	"os"
	"path/filepath"
)

const (
	appVersion      = "1.2.1"
	releaseAPIURL   = "https://api.github.com/repos/summersalestart/hyclicker/releases/latest"
	releasesPageURL = "https://github.com/summersalestart/hyclicker/releases/latest"
)

func appConfigPath() string {
	base, err := os.UserConfigDir()
	if err != nil {
		return fallbackConfigPath()
	}
	dir := filepath.Join(base, "HyClicker")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fallbackConfigPath()
	}
	return filepath.Join(dir, "config.json")
}

func fallbackConfigPath() string {
	exe, err := os.Executable()
	if err != nil {
		return "hyclicker.json"
	}
	return filepath.Join(filepath.Dir(exe), "hyclicker.json")
}
