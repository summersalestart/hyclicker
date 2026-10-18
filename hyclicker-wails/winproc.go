package main

import (
	"sync"
	"syscall"
	"unsafe"
)

type WndProcFn func(hwnd uintptr, msg uint32, w, l uintptr) uintptr

type WNDCLASS struct {
	Style, WndProc, ClsExtra, WndExtra, Instance, Icon, Cursor, Bk, MenuName, ClassName, IconSm uintptr
}

func defWindowProc(hwnd uintptr, msg uint32, w, l uintptr) uintptr {
	r, _, _ := procDefWinProc.Call(hwnd, uintptr(msg), w, l)
	return r
}

func hInstance() uintptr {
	h, _, _ := procGetModuleW.Call(0)
	return h
}

// createHiddenWindow registers className with the given wndproc, creates a
// message-only window on it, and starts a background message pump.
func createHiddenWindow(className string, proc WndProcFn) (uintptr, error) {
	cn, err := syscall.UTF16PtrFromString(className)
	if err != nil {
		return 0, err
	}
	wc := WNDCLASS{WndProc: syscall.NewCallback(proc), ClassName: uintptr(unsafe.Pointer(cn)), Instance: hInstance()}
	atom, _, e := procRegisterClass.Call(uintptr(unsafe.Pointer(&wc)))
	const classAlreadyExists = 1410
	if atom == 0 && e != syscall.Errno(classAlreadyExists) {
		return 0, e
	}

	// HWND_MESSAGE parent -> invisible message-only window
	hwnd, _, e2 := procCreateWin.Call(0, uintptr(unsafe.Pointer(cn)),
		0, 0, 0, 0, 0, 0, ^uintptr(2) /* HWND_MESSAGE */, 0, hInstance(), 0)
	if hwnd == 0 {
		return 0, e2
	}
	go pumpMessages(hwnd)
	return hwnd, nil
}

func destroyWindow(hwnd uintptr) { procDestroyWin.Call(hwnd) }

func registerHotkey(hwnd uintptr, id uintptr, mod, vk int) bool {
	r, _, _ := procRegHotKey.Call(hwnd, id, uintptr(mod), uintptr(vk))
	return r != 0
}

func unregisterHotkey(hwnd uintptr, id uintptr) { procUnregHotKey.Call(hwnd, id) }

var pumps sync.Map // hwnd -> quit channel

func pumpMessages(hwnd uintptr) {
	var msg struct {
		Hwnd, Message, WParam, LParam uintptr
		Time                          uint32
		Pt                            struct{ X, Y int32 }
	}
	quit, _ := pumps.LoadOrStore(hwnd, make(chan struct{}))
	q := quit.(chan struct{})
	for {
		select {
		case <-q:
			return
		default:
		}
		r, _, _ := procGetMsg.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if r == 0 || r == ^uintptr(0) {
			return
		}
		procDispatchMsg.Call(uintptr(unsafe.Pointer(&msg)))
	}
}

func stopPump(hwnd uintptr) {
	if q, ok := pumps.Load(hwnd); ok {
		close(q.(chan struct{}))
		pumps.Delete(hwnd)
	}
}
