package main

import (
	"errors"
	"fmt"
	"sync"
)

const (
	hotkeyID = 0xB00B
	WM_HOTKEY = 0x0312
)

const (
	modAlt     = 0x0001
	modControl = 0x0002
	modShift   = 0x0004
	modWin     = 0x0008
)

type HotkeyManager struct {
	mu        sync.Mutex
	window    uintptr
	id        uintptr
	vk        uint
	mod       uint
	name      string
	registered bool
	onTrigger func()
}

func NewHotkeyManager(onTrigger func()) *HotkeyManager {
	return &HotkeyManager{onTrigger: onTrigger, id: hotkeyID}
}

var ErrHotkeyTaken = errors.New("hotkey is taken or blocked")

func (h *HotkeyManager) Register(vk, mod uint, name string) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.registered {
		unregisterHotkey(h.window, h.id)
		h.registered = false
	}
	if h.window == 0 {
		wnd, err := createMessageWindow("HyClickerHotkeyWnd", h.wndProc)
		if err != nil {
			return fmt.Errorf("hotkey setup failed: %w", err)
		}
		h.window = wnd
	}
	if !registerHotkey(h.window, h.id, int(mod), int(vk)) {
		h.recallDefaults()
		return ErrHotkeyTaken
	}
	h.vk, h.mod, h.name = vk, mod, name
	h.registered = true
	return nil
}

func (h *HotkeyManager) recallDefaults() {
	if h.vk != 0 {
		if registerHotkey(h.window, h.id, int(h.mod), int(h.vk)) {
			h.registered = true
		}
	}
}

func (h *HotkeyManager) Release() {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.window == 0 {
		return
	}
	if h.registered {
		unregisterHotkey(h.window, h.id)
		h.registered = false
	}
	stopPump(h.window)
	destroyWindow(h.window)
	h.window = 0
}

func (h *HotkeyManager) Current() (vk, mod uint, name string, ok bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if !h.registered {
		return 0, 0, "", false
	}
	return h.vk, h.mod, h.name, true
}

func (h *HotkeyManager) wndProc(hwnd uintptr, msg uint32, w, l uintptr) uintptr {
	if msg == WM_HOTKEY && w == h.id && h.onTrigger != nil {
		go h.onTrigger()
		return 0
	}
	return defWindowProc(hwnd, msg, w, l)
}

var specialVKs = map[string]uint{
	"F1": 0x70, "F2": 0x71, "F3": 0x72, "F4": 0x73, "F5": 0x74, "F6": 0x75,
	"F7": 0x76, "F8": 0x77, "F9": 0x78, "F10": 0x79, "F11": 0x7A, "F12": 0x7B,
	"SPACE": 0x20, "UP": 0x26, "DOWN": 0x28, "LEFT": 0x25, "RIGHT": 0x27,
	"ESCAPE": 0x1B, "ESC": 0x1B, "TAB": 0x09, "ENTER": 0x0D, "RETURN": 0x0D,
	"INSERT": 0x2D, "DELETE": 0x2E, "DEL": 0x2E, "INS": 0x2D,
	"HOME": 0x24, "END": 0x23, "PAGEUP": 0x21, "PAGEDOWN": 0x22,
	"PGUP": 0x21, "PGDN": 0x22, "BACKSPACE": 0x08, "BACK": 0x08,
}

func vkFromName(key string) (uint, error) {
	if key == "" {
		return 0, fmt.Errorf("no key provided")
	}
	upper := toUpperASCII(key)
	if v, ok := specialVKs[upper]; ok {
		return v, nil
	}
	if len(upper) == 1 {
		c := upper[0]
		if (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
			return uint(c), nil
		}
	}
	return 0, fmt.Errorf("unsupported key: %s", key)
}

func toUpperASCII(s string) string {
	b := []byte(s)
	for i := range b {
		if b[i] >= 'a' && b[i] <= 'z' {
			b[i] -= 32
		}
	}
	return string(b)
}
