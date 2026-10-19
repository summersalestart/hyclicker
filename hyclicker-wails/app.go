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

	wails "github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx       context.Context
	quietExit bool
	mu        sync.Mutex
	cfg       Config
	engine    *ClickEngine
	hk        *HotkeyManager
}

func NewApp() *App { return &App{} }

func (a *App) Startup(ctx context.Context) {
	if err := acquireSingleInstance(); err != nil {
		focusExistingInstance()
		a.quietExit = true
		return
	}
	a.ctx = ctx
	a.quietExit = false
	cfg := readConfig(appConfigPath())
	a.cfg = cfg

	a.engine = NewClickEngine()
	a.engine.OnStateChange(func(running bool) {
		a.Emit("clicker-state", running)
	})
	a.engine.SetConfig(cfg.MinCps, cfg.MaxCps, normalizeButtonE(cfg.Button), cfg.Rocket)

	a.hk = NewHotkeyManager(func() { a.Emit("request-toggle", nil) })
	a.bindHotkeyOrFallback(cfg)
	go a.pollForUpdate()
}

func (a *App) DomReady(ctx context.Context) {
	if a.quietExit {
		wails.Quit(ctx)
	}
}

func (a *App) Shutdown(context.Context) {
	if a.engine != nil {
		a.engine.Stop()
	}
	if a.hk != nil {
		a.hk.Release()
	}
}

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
	if maxCps > maxAllowedCPS {
		return fmt.Errorf("Max CPS capped at %v", maxAllowedCPS)
	}
	if _, err := normalizeButton(button); err != nil {
		return err
	}

	a.mu.Lock()
	a.cfg.MinCps, a.cfg.MaxCps, a.cfg.Button = minCps, maxCps, button
	cfg := a.cfg
	a.mu.Unlock()

	a.engine.SetConfig(cfg.MinCps, cfg.MaxCps, normalizeButtonE(cfg.Button), cfg.Rocket)
	writeConfig(appConfigPath(), cfg)
	return nil
}

func (a *App) IsRunning() bool { return a.engine.Running() }

func (a *App) SetRocket(on bool) bool {
	a.mu.Lock()
	a.cfg.Rocket = on
	cfg := a.cfg
	a.mu.Unlock()
	a.engine.SetConfig(cfg.MinCps, cfg.MaxCps, normalizeButtonE(cfg.Button), cfg.Rocket)
	writeConfig(appConfigPath(), cfg)
	a.Emit("rocket-state", on)
	return on
}

func (a *App) Toggle() {
	if a.engine.Running() {
		a.engine.Stop()
		return
	}
	cfg := a.GetConfig()
	if err := a.Configure(cfg.MinCps, cfg.MaxCps, cfg.Button); err != nil {
		a.Emit("toast", err.Error())
		return
	}
	if err := a.engine.Start(); err != nil {
		a.Emit("toast", err.Error())
	}
}

func logHotkey(action string, err error) {
	f, ferr := os.OpenFile(filepath.Join(os.Getenv("APPDATA"), "HyClicker", "log.txt"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if ferr != nil {
		return
	}
	defer f.Close()
	if err != nil {
		fmt.Fprintf(f, "%s %s %v\n", time.Now().Format("15:04:05"), action, err)
	} else {
		fmt.Fprintf(f, "%s %s OK\n", time.Now().Format("15:04:05"), action)
	}
}

func (a *App) bindHotkeyOrFallback(cfg Config) {
	err := a.hk.Register(cfg.HotkeyVK, cfg.HotkeyMod, cfg.HotkeyName)
	logHotkey("startup("+cfg.HotkeyName+")", err)
	if err == nil {
		return
	}
	if _, _, _, ok := a.hk.Current(); ok {
		return
	}
	for _, alt := range fallbackKeys(cfg.HotkeyName) {
		vk, err := vkFromName(alt)
		if err != nil {
			continue
		}
		if err := a.hk.Register(vk, 0, alt); err == nil {
			logHotkey("fallback("+alt+")", nil)
			a.persistHotkey(vk, 0, alt)
			a.Emit("hotkey-changed", alt)
			return
		}
		logHotkey("fallback("+alt+")", err)
	}
	a.Emit("toast", fmt.Sprintf("Hotkey register failed: %v", err))
}

func (a *App) persistHotkey(vk, mod uint, name string) {
	a.mu.Lock()
	a.cfg.HotkeyVK, a.cfg.HotkeyMod, a.cfg.HotkeyName = vk, mod, name
	cfg := a.cfg
	a.mu.Unlock()
	writeConfig(appConfigPath(), cfg)
}

func fallbackKeys(taken string) []string {
	all := []string{"F6", "F7", "F8", "F9", "F10"}
	used := toUpperASCII(taken)
	out := make([]string, 0, len(all))
	for _, c := range all {
		if toUpperASCII(c) != used {
			out = append(out, c)
		}
	}
	return out
}

func (a *App) BindHotkey(key string, ctrl, alt, shift bool) error {
	vk, err := vkFromName(key)
	if err != nil {
		return err
	}
	mod := 0
	if ctrl {
		mod |= modControl
	}
	if alt {
		mod |= modAlt
	}
	if shift {
		mod |= modShift
	}
	if err := a.hk.Register(vk, uint(mod), key); err != nil {
		logHotkey("bind("+key+")", err)
		return err
	}
	logHotkey("bind("+key+")", nil)

	a.mu.Lock()
	a.cfg.HotkeyVK, a.cfg.HotkeyMod, a.cfg.HotkeyName = vk, uint(mod), key
	cfg := a.cfg
	a.mu.Unlock()
	writeConfig(appConfigPath(), cfg)
	return nil
}

func (a *App) MinimizeWindow() {
	if a.ctx != nil {
		wails.WindowMinimise(a.ctx)
	}
}

func (a *App) CloseApp() {
	if a.ctx != nil {
		wails.Quit(a.ctx)
	}
}

func (a *App) OpenReleases() {
	if err := openInBrowser(releasesPageURL); err != nil {
		a.Emit("toast", "Could not open browser")
	}
}

func (a *App) Emit(event string, data interface{}) {
	if a.ctx != nil {
		wails.EventsEmit(a.ctx, event, data)
	}
}

func (a *App) pollForUpdate() {
	time.Sleep(2 * time.Second)
	latest, err := fetchLatestRelease()
	if err != nil || latest == "" {
		return
	}
	if isNewerVersion(latest, appVersion) {
		a.Emit("update-available", latest)
	}
}

func fetchLatestRelease() (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(releaseAPIURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("release check failed: %s", resp.Status)
	}
	var body struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", err
	}
	return trimVersionPrefix(body.TagName), nil
}
