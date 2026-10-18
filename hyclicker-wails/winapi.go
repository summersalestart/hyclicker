package main

import (
	"fmt"
	"syscall"
	"unsafe"
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	shell32  = syscall.NewLazyDLL("shell32.dll")
)

var (
	procGetModuleW    = kernel32.NewProc("GetModuleHandleW")
	procRegisterClass = user32.NewProc("RegisterClassW")
	procCreateWin     = user32.NewProc("CreateWindowExW")
	procDefWinProc    = user32.NewProc("DefWindowProcW")
	procGetMsg        = user32.NewProc("GetMessageW")
	procDispatchMsg   = user32.NewProc("DispatchMessageW")
	procRegHotKey     = user32.NewProc("RegisterHotKey")
	procUnregHotKey   = user32.NewProc("UnregisterHotKey")
	procDestroyWin    = user32.NewProc("DestroyWindow")
	procSendInput     = user32.NewProc("SendInput")
	procShellExec     = shell32.NewProc("ShellExecuteW")
)

// ---- SendInput ----

func sendInputs(inputs []INPUT) uint32 {
	if len(inputs) == 0 {
		return 0
	}
	ret, _, _ := procSendInput.Call(
		uintptr(len(inputs)),
		uintptr(unsafe.Pointer(&inputs[0])),
		unsafe.Sizeof(inputs[0]),
	)
	return uint32(ret)
}

// ---- OpenURL ----

func OpenURL(url string) error {
	u16, err := syscall.UTF16PtrFromString(url)
	if err != nil {
		return err
	}
	verb, _ := syscall.UTF16PtrFromString("open")
	r, _, _ := procShellExec.Call(0, uintptr(unsafe.Pointer(verb)),
		uintptr(unsafe.Pointer(u16)), 0, 0, 1 /* SW_SHOWNORMAL */)
	if r < 33 {
		return fmt.Errorf("open failed")
	}
	return nil
}

// ---- version compare ----

func trimV(s string) string {
	for i, c := range s {
		if c >= '0' && c <= '9' {
			return s[i:]
		}
	}
	return s
}

func isNewer(candidate, current string) bool {
	ca, cb, cc := parseVersion(candidate)
	ta, tb, tc := parseVersion(current)
	if ca != ta {
		return ca > ta
	}
	if cb != tb {
		return cb > tb
	}
	return cc > tc
}

func parseVersion(s string) (a, b, c int) {
	parts := []int{}
	num, has := 0, false
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if ch >= '0' && ch <= '9' {
			num = num*10 + int(ch-'0')
			has = true
		} else if has {
			parts = append(parts, num)
			num, has = 0, false
		}
	}
	if has {
		parts = append(parts, num)
	}
	for len(parts) < 3 {
		parts = append(parts, 0)
	}
	return parts[0], parts[1], parts[2]
}
