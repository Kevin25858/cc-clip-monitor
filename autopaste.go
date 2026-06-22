package main

import (
	"time"
	"unsafe"

	"syscall"
)

var (
	getForegroundWindow = user32.NewProc("GetForegroundWindow")
	keybdEvent          = user32.NewProc("keybd_event")
)

const (
	VK_CONTROL        = 0x11
	VK_V              = 0x56
	KEYEVENTF_KEYUP   = 0x0002
)

type AutoPaster struct {
	uploadHwnd uintptr
	ready      bool
}

func NewAutoPaster() *AutoPaster {
	return &AutoPaster{}
}

func (ap *AutoPaster) MarkReady() {
	ap.uploadHwnd, _, _ = getForegroundWindow.Call()
	ap.ready = true
}

func (ap *AutoPaster) CheckAndPaste() bool {
	if !ap.ready {
		return false
	}
	curHwnd, _, _ := getForegroundWindow.Call()
	if curHwnd != 0 && curHwnd != ap.uploadHwnd {
		time.Sleep(200 * time.Millisecond)
		simulateCtrlV()
		ap.ready = false
		return true
	}
	return false
}

func simulateCtrlV() {
	keybdEvent.Call(VK_CONTROL, 0, 0, 0)
	keybdEvent.Call(VK_V, 0, 0, 0)
	keybdEvent.Call(VK_V, 0, KEYEVENTF_KEYUP, 0)
	keybdEvent.Call(VK_CONTROL, 0, KEYEVENTF_KEYUP, 0)
}

func writeClipboardText(text string) error {
	r1, _, _ := openClipboard.Call(0)
	if r1 == 0 {
		return syscall.GetLastError()
	}
	defer closeClipboard.Call()

	emptyProc := user32.NewProc("EmptyClipboard")
	emptyProc.Call()

	setClipboardData := user32.NewProc("SetClipboardData")
	gAlloc := kernel32.NewProc("GlobalAlloc")
	gLock := kernel32.NewProc("GlobalLock")
	gUnlock := kernel32.NewProc("GlobalUnlock")

	n := len(text) + 1
	h, _, _ := gAlloc.Call(0x0042, uintptr(n))
	if h == 0 {
		return syscall.GetLastError()
	}
	ptr, _, _ := gLock.Call(h)
	if ptr == 0 {
		return syscall.GetLastError()
	}

	buf := unsafe.Slice((*byte)(unsafe.Pointer(ptr)), n)
	copy(buf, []byte(text))
	buf[n-1] = 0

	gUnlock.Call(h)
	setClipboardData.Call(uintptr(CF_TEXT), h)
	return nil
}
