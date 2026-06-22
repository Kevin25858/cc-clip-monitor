package main

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type Uploader struct {
	host      *SSHHost
	remoteDir string
	useCCClip bool
}

func NewUploader(host *SSHHost, remoteDir string, useCCClip bool) *Uploader {
	return &Uploader{
		host:      host,
		remoteDir: remoteDir,
		useCCClip: useCCClip,
	}
}

func (u *Uploader) Upload(filePath string) (string, error) {
	if u.useCCClip {
		return u.uploadViaCCClip(filePath)
	}
	return u.uploadViaSCP(filePath)
}

func (u *Uploader) uploadViaCCClip(filePath string) (string, error) {
	cmd := exec.Command("cc-clip", "send", u.host.Name, filePath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("cc-clip send failed: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return strings.TrimSpace(string(out)), nil
}

func (u *Uploader) sshTarget() string {
	if u.host.User != "" && u.host.HostName != "" {
		return u.host.User + "@" + u.host.HostName
	}
	return u.host.Name
}

func (u *Uploader) sshArgs(extra ...string) []string {
	args := []string{}
	if u.host.Port != "" && u.host.Port != "22" {
		args = append(args, "-p", u.host.Port)
	}
	args = append(args, u.sshTarget())
	args = append(args, extra...)
	return args
}

func (u *Uploader) scpArgs(src, dst string) []string {
	args := []string{}
	if u.host.Port != "" && u.host.Port != "22" {
		args = append(args, "-P", u.host.Port)
	}
	args = append(args, src, dst)
	return args
}

func (u *Uploader) uploadViaSCP(filePath string) (string, error) {
	exec.Command("ssh", u.sshArgs("mkdir -p "+u.remoteDir)...).CombinedOutput()

	remoteName := fmt.Sprintf("clip-%d.png", time.Now().UnixMilli())
	remotePath := u.remoteDir + "/" + remoteName
	dst := u.sshTarget() + ":" + remotePath

	cmd := exec.Command("scp", u.scpArgs(filePath, dst)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("scp failed: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return remotePath, nil
}
