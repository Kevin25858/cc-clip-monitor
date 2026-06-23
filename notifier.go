package main

import (
	"fmt"
	"os/exec"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

var (
	user32Dll           = syscall.NewLazyDLL("user32.dll")
	procMessageBoxW     = user32Dll.NewProc("MessageBoxW")
	MB_OK        uintptr = 0x00000000
	MB_ICON_INFO uintptr = 0x00000040
	MB_TOPMOST   uintptr = 0x00040000
)

// showMessageBox 通过 user32.dll 弹出系统对话框，不依赖通知中心
func showMessageBox(title, body string) {
	titlePtr, _ := syscall.UTF16PtrFromString(title)
	bodyPtr, _ := syscall.UTF16PtrFromString(body)
	procMessageBoxW.Call(0, uintptr(unsafe.Pointer(bodyPtr)), uintptr(unsafe.Pointer(titlePtr)), MB_OK|MB_ICON_INFO|MB_TOPMOST)
}

func ShowToast(title, body string) {
	ps := fmt.Sprintf(`
[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null
[Windows.Data.Xml.Dom.XmlDocument, Windows.Data.Xml.Dom, ContentType = WindowsRuntime] | Out-Null
$xml = New-Object Windows.Data.Xml.Dom.XmlDocument
$xml.LoadXml('<toast><visual><binding template="ToastGeneric"><text>%s</text><text>%s</text></binding></visual></toast>')
$toast = [Windows.UI.Notifications.ToastNotification]::new($xml)
[Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier("cc-clip-monitor").Show($toast)
`, title, body)

	cmd := exec.Command("powershell", "-NoProfile", "-Command", ps)
	cmd.Run()
}

// NotifyWithFallback 先尝试 Toast，失败时用 MessageBox 弹窗
func NotifyWithFallback(title, body string) {
	done := make(chan struct{})

	// Toast 异步执行，3秒超时
	go func() {
		ShowToast(title, body)
		close(done)
	}()

	select {
	case <-done:
		// Toast 完成
	case <-time.After(3 * time.Second):
		// Toast 超时（可能通知中心被关闭），改用 MessageBox
		go showMessageBox(title, body)
	}
}

func NotifyUpload(remotePath string) {
	short := remotePath
	if idx := strings.LastIndex(remotePath, "/"); idx >= 0 {
		short = remotePath[idx+1:]
	}
	NotifyWithFallback("cc-clip — 上传成功", fmt.Sprintf("路径已复制: %s", short))
}

func NotifyFailure(err string) {
	NotifyWithFallback("cc-clip — 上传失败", err)
}
