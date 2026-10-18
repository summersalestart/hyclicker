package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

func WailsEmit(ctx context.Context, event string, data interface{}) {
	if ctx != nil {
		wailsRuntime.EventsEmit(ctx, event, data)
	}
}

type Config struct {
	MinCps     float64 `json:"min_cps"`
	MaxCps     float64 `json:"max_cps"`
	Button     string  `json:"button"`
	HotkeyVK   uint    `json:"hotkey_vk"`
	HotkeyMod  uint    `json:"hotkey_mod"`
	HotkeyName string  `json:"hotkey_name"`
}

// App is the bridge exposed to the frontend.
type App struct {
	ctx    context.Context
	mu     sync.Mutex
	cfg    Config
	engine *ClickEngine
	hk     *Hotkey
	cfgDir string
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = "."
	}
	a.cfgDir = filepath.Join(dir, "HyClicker")
	_ = os.MkdirAll(a.cfgDir, 0o755)
	a.cfg = a.loadConfig()
	a.engine = NewClickEngine()
	a.engine.Configure(a.cfg.MinCps, a.cfg.MaxCps, a.cfg.Button)
	a.engine.OnStateChange(func(running bool) {
		WailsEmit(ctx, "clicker-state", running)
	})
	a.hk = NewHotkey(func() {
		WailsEmit(ctx, "request-toggle", nil)
	})
	_ = a.hk.Register(a.cfg.HotkeyVK, a.cfg.HotkeyMod)
	go a.checkUpdate()
}

func (a *App) shutdown(ctx context.Context) {
	a.engine.Stop()
	a.hk.Unregister()
}

func cfgPath(dir string) string { return filepath.Join(dir, "config.json") }

func (a *App) loadConfig() Config {
	def := Config{MinCps: 8, MaxCps: 12, Button: "Left", HotkeyVK: VK_F6, HotkeyMod: 0, HotkeyName: "F6"}
	data, err := os.ReadFile(cfgPath(a.cfgDir))
	if err != nil {
		return def
	}
	var c Config
	if json.Unmarshal(data, &c) != nil {
		return def
	}
	if c.MinCps <= 0 {
		c.MinCps = def.MinCps
	}
	if c.MaxCps < c.MinCps {
		c.MaxCps = c.MinCps
	}
	return c
}

func (a *App) saveConfig() {
	data, _ := json.MarshalIndent(a.cfg, "", "  ")
	_ = os.WriteFile(cfgPath(a.cfgDir), data, 0o644)
}

// ---- exported API (called from JS) ----

func (a *App) GetConfig() Config {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.cfg
}

func (a *App) Configure(minCps, maxCps float64, button string) error {
	if minCps <= 0 || maxCps <= 0 {
		return fmt.Errorf("CPS must be greater than 0")
	}
	if minCps > maxCps {
		return fmt.Errorf("CPS (1) cannot exceed CPS (2)")
	}
	if maxCps > 50 {
		return fmt.Errorf("Max CPS capped at 50")
	}
	a.mu.Lock()
	a.cfg.MinCps, a.cfg.MaxCps, a.cfg.Button = minCps, maxCps, button
	cfgCopy := a.cfg
	a.mu.Unlock()
	a.engine.Configure(minCps, maxCps, button)
	a.mu.Lock() // re-lock only to save; saveConfig reads without external lock
	a.mu.Unlock()
	a.saveConfigLocked(&cfgCopy)
	return nil
}

func (a *App) saveConfigLocked(c *Config) {
	data, _ := json.MarshalIndent(c, "", "  ")
	_ = os.WriteFile(cfgPath(a.cfgDir), data, 0o644)
}

func (a *App) IsRunning() bool { return a.engine.IsRunning() }

func (a *App) Toggle() {
	a.mu.Lock()
	running := a.engine.IsRunning()
	cfgCopy := a.cfg
	a.mu.Unlock()
	if running {
		a.engine.Stop()
		return
	}
	if err := a.Configure(cfgCopy.MinCps, cfgCopy.MaxCps, cfgCopy.Button); err != nil {
		WailsEmit(a.ctx, "toast", err.Error())
		return
	}
	a.engine.Start()
}

// SetHotkeyFromNames registers a hotkey from friendly names, e.g.
// SetHotkeyFromNames("F6", false, false, false)
func (a *App) SetHotkeyFromNames(key string, ctrl, alt, shift bool) error {
	if key == "" {
		return fmt.Errorf("no key")
	}
	vk, ok := vkFromName(key)
	if !ok {
		return fmt.Errorf("unsupported key")
	}
	mod := 0
	if ctrl {
		mod |= MOD_CONTROL
	}
	if alt {
		mod |= MOD_ALT
	}
	if shift {
		mod |= MOD_SHIFT
	}
	if err := a.SetHotkey(vk, uint(mod), key); err != nil {
		return err
	}
	return nil
}

func (a *App) SetHotkey(vk, mod uint, name string) error {
	if err := a.hk.Register(vk, mod); err != nil {
		return fmt.Errorf("Hotkey register failed")
	}
	a.mu.Lock()
	a.cfg.HotkeyVK, a.cfg.HotkeyMod, a.cfg.HotkeyName = vk, mod, name
	a.mu.Unlock()
	a.saveConfig()
	return nil
}

func (a *App) Quit() {
	wailsRuntime.Quit(a.ctx)
}

func (a *App) Minimize() {
	wailsRuntime.WindowMinimise(a.ctx)
}

func vkFromName(key string) (uint, bool) {
	k := upperName(key)
	if v, ok := specialVKs[k]; ok {
		return v, true
	}
	if len(k) == 1 {
		c := k[0]
		if (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
			return uint(c), true
		}
	}
	return 0, false
}

var specialVKs = map[string]uint{
	"F6": 0x75, "F1": 0x70, "F2": 0x71, "F3": 0x72, "F4": 0x73, "F5": 0x74,
	"F7": 0x76, "F8": 0x77, "F9": 0x78, "F10": 0x79, "F11": 0x7A, "F12": 0x7B,
	"SPACE": 0x20, "UP": 0x26, "DOWN": 0x28, "LEFT": 0x25, "RIGHT": 0x27,
	"ESCAPE": 0x1B, "ESC": 0x1B, "TAB": 0x09, "ENTER": 0x0D,
	"INSERT": 0x2D, "DELETE": 0x2E, "DEL": 0x2E, "HOME": 0x24, "END": 0x23,
	"PAGEUP": 0x21, "PAGEDOWN": 0x22, "PGUP": 0x21, "PGDN": 0x22,
}

func upperName(s string) string {
	b := []byte(s)
	for i := range b {
		if b[i] >= 'a' && b[i] <= 'z' {
			b[i] -= 32
		}
	}
	return string(b)
}


func (a *App) OpenFileDialog() {}

type UpdateInfo struct {
	Version string `json:"version"`
}

const repoAPIURL = "https://api.github.com/repos/summersalestart/hyclicker/releases/latest"

func (a *App) checkUpdate() {
	time.Sleep(2 * time.Second) // let the window settle
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(repoAPIURL)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	var rel struct {
		TagName string `json:"tag_name"`
	}
	if json.NewDecoder(resp.Body).Decode(&rel) != nil {
		return
	}
	latest := trimV(rel.TagName)
	if isNewer(latest, AppVersion) {
		WailsEmit(a.ctx, "update-available", latest)
	}
}

func (a *App) OpenReleases() {
	_ = OpenURL("https://github.com/summersalestart/hyclicker/releases/latest")
}
