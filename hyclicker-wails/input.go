package main

import (
	"fmt"
	"sync"
)

const (
	VK_F6       = 0x75
	MOD_ALT     = 0x0001
	MOD_CONTROL = 0x0002
	MOD_SHIFT   = 0x0004
	MOD_WIN     = 0x0008
	WM_HOTKEY   = 0x0312
)

type Hotkey struct {
	mu         sync.Mutex
	hotkeyID   uintptr
	window     uintptr
	currentVK  uint
	currentMod uint
	registered bool
	onTrigger  func()
	className  string
}

func NewHotkey(onTrigger func()) *Hotkey {
	return &Hotkey{onTrigger: onTrigger, hotkeyID: 0xB00B, className: "HyClickerHotkeyWnd"}
}

func (h *Hotkey) Register(vk, mod uint) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.registered {
		unregisterHotkey(h.window, h.hotkeyID)
		h.registered = false
	}
	if h.window == 0 {
		wnd, err := createHiddenWindow(h.className, h.wndProc)
		if err != nil {
			return err
		}
		h.window = wnd
	}
	if !registerHotkey(h.window, h.hotkeyID, int(mod), int(vk)) {
		return fmt.Errorf("hotkey is taken or blocked")
	}
	h.currentVK, h.currentMod = vk, mod
	h.registered = true
	return nil
}

func (h *Hotkey) Unregister() {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.registered {
		unregisterHotkey(h.window, h.hotkeyID)
		h.registered = false
	}
	if h.window != 0 {
		stopPump(h.window)
		destroyWindow(h.window)
		h.window = 0
	}
}

func (h *Hotkey) wndProc(hwnd uintptr, msg uint32, w, l uintptr) uintptr {
	if msg == WM_HOTKEY && w == h.hotkeyID && h.onTrigger != nil {
		go h.onTrigger()
		return 0
	}
	return defWindowProc(hwnd, msg, w, l)
}

// ---- friendly key name -> VK (implemented in app.go) ----
