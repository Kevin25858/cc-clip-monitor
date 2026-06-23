package main

import (
	"fmt"
	"strings"
	"syscall"
	"unsafe"
)

var (
	user32Dll           = syscall.NewLazyDLL("user32.dll")
	procMessageBoxW     = user32Dll.NewProc("MessageBoxW")
	MB_OK        uintptr = 0x00000000
	MB_ICON_INFO uintptr = 0x00000040
	MB_TOPMOST   uintptr = 0x00040000
)

func ShowMessage(title, body string) {
	titlePtr, _ := syscall.UTF16PtrFromString(title)
	bodyPtr, _ := syscall.UTF16PtrFromString(body)
	procMessageBoxW.Call(0, uintptr(unsafe.Pointer(bodyPtr)), uintptr(unsafe.Pointer(titlePtr)), MB_OK|MB_ICON_INFO|MB_TOPMOST)
}

func NotifyUpload(remotePath string) {
	short := remotePath
	if idx := strings.LastIndex(remotePath, "/"); idx >= 0 {
		short = remotePath[idx+1:]
	}
	ShowMessage("cc-clip — 上传成功", fmt.Sprintf("路径已复制: %s", short))
}

func NotifyFailure(err string) {
	ShowMessage("cc-clip — 上传失败", err)
}
