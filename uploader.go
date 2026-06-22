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

func (u *Uploader) scpArgs(src, dst string) []string {
	args := []string{}
	if u.host.Port != "" && u.host.Port != "22" {
		args = append(args, "-P", u.host.Port)
	}
	args = append(args, src, dst)
	return args
}

func (u *Uploader) uploadViaSCP(filePath string) (string, error) {
	remoteName := fmt.Sprintf("clip-%d.png", time.Now().UnixMilli())
	remotePath := u.remoteDir + "/" + remoteName

	// mkdir -p remoteDir first
	mkdirTarget := u.host.sshTarget()
	mkdirArgs := []string{}
	if u.host.Port != "" && u.host.Port != "22" {
		mkdirArgs = append(mkdirArgs, "-p", u.host.Port)
	}
	mkdirArgs = append(mkdirArgs, mkdirTarget, "mkdir -p "+u.remoteDir)
	mkdirCmd := exec.Command("ssh", mkdirArgs...)
	u.withPassword(mkdirCmd)
	mkdirCmd.CombinedOutput()

	dst := u.host.sshTarget() + ":" + remotePath
	cmd := exec.Command("scp", u.scpArgs(filePath, dst)...)
	u.withPassword(cmd)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("scp failed: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return remotePath, nil
}

// withPassword wires stdin so ssh/scp can read a password prompt.
// If no password is set, stdin is left untouched (key-based auth).
func (u *Uploader) withPassword(cmd *exec.Cmd) {
	if u.host.Password != "" {
		cmd.Stdin = strings.NewReader(u.host.Password + "\n")
	}
}
