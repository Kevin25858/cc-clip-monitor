package main

import (
	"os/exec"
	"strings"
)

func CleanupRemote(host *SSHHost, remoteDir string) {
	args := []string{}
	if host.Port != "" && host.Port != "22" {
		args = append(args, "-p", host.Port)
	}
	target := host.Name
	if host.User != "" && host.HostName != "" {
		target = host.User + "@" + host.HostName
	}
	args = append(args, target, "rm -f "+remoteDir+"/* 2>/dev/null")
	cmd := exec.Command("ssh", args...)
	if host.Password != "" {
		cmd.Stdin = strings.NewReader(host.Password + "\n")
	}
	cmd.CombinedOutput()
}
