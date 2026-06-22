package main

import (
	"os/exec"
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
	exec.Command("ssh", args...).CombinedOutput()
}
