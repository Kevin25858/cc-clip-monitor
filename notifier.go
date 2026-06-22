package main

import (
	"fmt"
	"os/exec"
	"strings"
)

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

func NotifyUpload(remotePath string) {
	short := remotePath
	if idx := strings.LastIndex(remotePath, "/"); idx >= 0 {
		short = remotePath[idx+1:]
	}
	ShowToast("cc-clip — 上传成功", fmt.Sprintf("路径已复制: %s", short))
}

func NotifyFailure(err string) {
	ShowToast("cc-clip — 上传失败", err)
}
