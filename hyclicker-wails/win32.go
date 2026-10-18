package main

import (
	"syscall"
	"unsafe"
)

var (
	user32            = syscall.NewLazyDLL("user32.dll")
	kernel32          = syscall.NewLazyDLL("kernel32.dll")
	shell32           = syscall.NewLazyDLL("shell32.dll")
	procGetModuleW    = kernel32.NewProc("GetModuleHandleW")
	procRegisterCls   = user32.NewProc("RegisterClassW")
	procCreateWnd     = user32.NewProc("CreateWindowExW")
	procDefProc       = user32.NewProc("DefWindowProcW")
	procGetMsg        = user32.NewProc("GetMessageW")
	procDispatchMsg   = user32.NewProc("DispatchMessageW")
	procRegisterHK    = user32.NewProc("RegisterHotKey")
	procUnregister    = user32.NewProc("UnregisterHotKey")
	procDestroyWnd    = user32.NewProc("DestroyWindow")
	procSendInput     = user32.NewProc("SendInput")
	procShellExec     = shell32.NewProc("ShellExecuteW")
	procCreateMutex   = kernel32.NewProc("CreateMutexW")
	procFindWindow    = user32.NewProc("FindWindowW")
	procSetForeground = user32.NewProc("SetForegroundWindow")
	procGetCurrentId  = kernel32.NewProc("GetCurrentThreadId")
	procPostThreadMsg = user32.NewProc("PostThreadMessageW")
	procCreateEvent   = kernel32.NewProc("CreateEventW")
	procSetEvent      = kernel32.NewProc("SetEvent")
	procMsgWait       = user32.NewProc("MsgWaitForMultipleObjects")
	procPeekMsg       = user32.NewProc("PeekMessageW")
)

func defWindowProc(hwnd uintptr, msg uint32, w, l uintptr) uintptr {
	r, _, _ := procDefProc.Call(hwnd, uintptr(msg), w, l)
	return r
}

func registerHotkey(hwnd, id uintptr, mod, vk int) (bool, uint32) {
	r, _, err := procRegisterHK.Call(hwnd, id, uintptr(mod), uintptr(vk))
	return r != 0, uint32(err.(syscall.Errno))
}

func unregisterHotkey(hwnd, id uintptr) { procUnregister.Call(hwnd, id) }

func destroyWindow(hwnd uintptr) { procDestroyWnd.Call(hwnd) }

type windowClass struct {
	Style     uint32
	WndProc   uintptr
	ClsExtra  int32
	WndExtra  int32
	Instance  uintptr
	Icon      uintptr
	Cursor    uintptr
	Bk        uintptr
	MenuName  uintptr
	ClassName uintptr
	IconSm    uintptr
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
	return hwnd, nil
}

func pumpLoop(hwnd, event uintptr, cmd <-chan func(), done chan<- struct{}) {
	defer func() {
		procDestroyWnd.Call(hwnd)
		if done != nil {
			close(done)
		}
	}()
	var msg struct {
		Hwnd, Message, WParam, LParam uintptr
		Time                          uint32
		Pt                            struct {
			X, Y int32
		}
	}
	for {
		procMsgWait.Call(1, uintptr(unsafe.Pointer(&event)), 0, 0xFFFFFFFF, 0x04FF)
		drainCommands(cmd)
		for {
			r, _, _ := procPeekMsg.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0, 1)
			if r == 0 {
				break
			}
			if msg.Message == 0x0012 {
				return
			}
			procDispatchMsg.Call(uintptr(unsafe.Pointer(&msg)))
		}
	}
}

func drainCommands(cmd <-chan func()) {
	for {
		select {
		case f := <-cmd:
			f()
		default:
			return
		}
	}
}

func postQuit(threadID uint32) {
	if threadID != 0 {
		procPostThreadMsg.Call(uintptr(threadID), 0x0012, 0, 0)
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
