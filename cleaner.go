package main

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"os/exec"
)

const cleanupAge = 15 // minutes

// CleanupRemote 删除远程 uploads 目录中超过 cleanupAge 分钟的文件
func CleanupRemote(host *SSHHost, remoteDir string) {
	args := []string{}
	if host.Port != "" && host.Port != "22" {
		args = append(args, "-p", host.Port)
	}
	target := host.Name
	if host.User != "" && host.HostName != "" {
		target = host.User + "@" + host.HostName
	}
	// find 只删除修改时间超过 15 分钟的 png 文件
	args = append(args, target, "find "+remoteDir+" -name '*.png' -mmin +"+itoa(cleanupAge)+" -delete 2>/dev/null")
	cmd := exec.Command("ssh", args...)
	if host.Password != "" {
		cmd.Stdin = strings.NewReader(host.Password + "\n")
	}
	cmd.CombinedOutput()
}

// CleanupLocal 删除本地临时目录中超过 cleanupAge 分钟的图片文件
func CleanupLocal(tempDir string) {
	entries, err := os.ReadDir(tempDir)
	if err != nil {
		return
	}
	cutoff := time.Now().Add(-time.Duration(cleanupAge) * time.Minute)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) && strings.HasSuffix(entry.Name(), ".png") {
			os.Remove(filepath.Join(tempDir, entry.Name()))
		}
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
