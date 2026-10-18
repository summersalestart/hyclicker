package main

import (
	"sync"
	"syscall"
	"unsafe"
)

var (
	user32          = syscall.NewLazyDLL("user32.dll")
	kernel32        = syscall.NewLazyDLL("kernel32.dll")
	shell32         = syscall.NewLazyDLL("shell32.dll")
	procGetModuleW  = kernel32.NewProc("GetModuleHandleW")
	procRegisterCls = user32.NewProc("RegisterClassW")
	procCreateWnd   = user32.NewProc("CreateWindowExW")
	procDefProc     = user32.NewProc("DefWindowProcW")
	procGetMsg      = user32.NewProc("GetMessageW")
	procDispatchMsg = user32.NewProc("DispatchMessageW")
	procRegisterHK  = user32.NewProc("RegisterHotKey")
	procUnregister  = user32.NewProc("UnregisterHotKey")
	procDestroyWnd  = user32.NewProc("DestroyWindow")
	procSendInput   = user32.NewProc("SendInput")
	procShellExec   = shell32.NewProc("ShellExecuteW")
)

func defWindowProc(hwnd uintptr, msg uint32, w, l uintptr) uintptr {
	r, _, _ := procDefProc.Call(hwnd, uintptr(msg), w, l)
	return r
}

func registerHotkey(hwnd, id uintptr, mod, vk int) bool {
	r, _, _ := procRegisterHK.Call(hwnd, id, uintptr(mod), uintptr(vk))
	return r != 0
}

func unregisterHotkey(hwnd, id uintptr) { procUnregister.Call(hwnd, id) }

func destroyWindow(hwnd uintptr) { procDestroyWnd.Call(hwnd) }

type windowClass struct {
	Style, WndProc, ClsExtra, WndExtra, Instance, Icon, Cursor, Bk, MenuName, ClassName, IconSm uintptr
}

func createMessageWindow(className string, proc func(uintptr, uint32, uintptr, uintptr) uintptr) (uintptr, error) {
	cn, err := syscall.UTF16PtrFromString(className)
	if err != nil {
		return 0, err
	}
	instance, _, _ := procGetModuleW.Call(0)
	wc := windowClass{
		WndProc:   syscall.NewCallback(proc),
		ClassName: uintptr(unsafe.Pointer(cn)),
		Instance:  instance,
	}
	atom, _, callErr := procRegisterCls.Call(uintptr(unsafe.Pointer(&wc)))
	const classAlreadyExists = 1410
	if atom == 0 && callErr != syscall.Errno(classAlreadyExists) {
		return 0, callErr
	}

	const HWND_MESSAGE = ^uintptr(2)
	hwnd, _, createErr := procCreateWnd.Call(
		0,
		uintptr(unsafe.Pointer(cn)),
		0, 0, 0, 0, 0, 0,
		HWND_MESSAGE, 0, instance, 0,
	)
	if hwnd == 0 {
		return 0, createErr
	}
	go pumpMessages(hwnd)
	return hwnd, nil
}

var pumps sync.Map

func pumpMessages(hwnd uintptr) {
	quit, _ := pumps.LoadOrStore(hwnd, make(chan struct{}))

	var msg struct {
		Hwnd, Message, WParam, LParam uintptr
		Time                          uint32
		Pt                            struct{ X, Y int32 }
	}
	for {
		select {
		case <-quit.(chan struct{}):
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

func openInBrowser(url string) error {
	u, err := syscall.UTF16PtrFromString(url)
	if err != nil {
		return err
	}
	verb, err := syscall.UTF16PtrFromString("open")
	if err != nil {
		return err
	}
	r, _, _ := procShellExec.Call(0, uintptr(unsafe.Pointer(verb)), uintptr(unsafe.Pointer(u)), 0, 0, 1)
	if r < 33 {
		return syscall.EINVAL
	}
	return nil
}
