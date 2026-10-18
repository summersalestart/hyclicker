package main

import (
	"unsafe"
)

type MOUSEINPUT struct {
	X, Y      int32
	MouseData uint32
	Flags     uint32
	Time      uint32
	ExtraInfo uintptr
}

type INPUT struct {
	Type  uint32
	_     [4]byte
	Mouse MOUSEINPUT
}

func sendInput(inputs []INPUT) {
	if len(inputs) == 0 {
		return
	}
	procSendInput.Call(
		uintptr(len(inputs)),
		uintptr(unsafe.Pointer(&inputs[0])),
		unsafe.Sizeof(inputs[0]),
	)
}

func clickInput(down, up uint32) []INPUT {
	return []INPUT{
		{Mouse: MOUSEINPUT{Flags: down}},
		{Mouse: MOUSEINPUT{Flags: up}},
	}
}
