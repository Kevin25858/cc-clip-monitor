package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type SSHHost struct {
	Name     string
	HostName string
	Port     string
	User     string
	Password string // only for this session
}

func ParseSSHConfig() ([]SSHHost, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configPath := home + "/.ssh/config"
	f, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var hosts []SSHHost
	var current *SSHHost

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, " ", 2)
		if len(parts) < 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		switch strings.ToLower(key) {
		case "host":
			if current != nil {
				hosts = append(hosts, *current)
			}
			name := strings.Fields(val)
			if len(name) > 0 && !strings.Contains(name[0], "*") {
				current = &SSHHost{Name: name[0], Port: "22"}
			} else {
				current = nil
			}
		case "hostname":
			if current != nil {
				current.HostName = val
			}
		case "port":
			if current != nil {
				current.Port = val
			}
		case "user":
			if current != nil {
				current.User = val
			}
		}
	}
	if current != nil {
		hosts = append(hosts, *current)
	}
	return hosts, nil
}

func (h SSHHost) Address() string {
	host := h.HostName
	if host == "" {
		host = h.Name
	}
	if h.Port != "" && h.Port != "22" {
		return host + ":" + h.Port
	}
	return host
}

func (h SSHHost) SCPDest() string {
	host := h.HostName
	if host == "" {
		host = h.Name
	}
	if h.Port != "" && h.Port != "22" {
		return "-P " + h.Port
	}
	return ""
}

func (h SSHHost) sshTarget() string {
	if h.User != "" && h.HostName != "" {
		return h.User + "@" + h.HostName
	}
	return h.Name
}

// sshArgs builds the argument list for ssh, optionally with a command to run.
func (h SSHHost) sshArgs(extra ...string) []string {
	args := []string{}
	if h.Port != "" && h.Port != "22" {
		args = append(args, "-p", h.Port)
	}
	args = append(args, h.sshTarget())
	args = append(args, extra...)
	return args
}

// TestSSHConnection tests if the host can be connected via SSH.
// If key-based auth works, returns immediately.
// Otherwise prompts for password and verifies it.
func TestSSHConnection(host *SSHHost) error {
	fmt.Printf("\033[36m正在连接 %s ...\033[0m\n", host.Address())

	// Try with default (key-based) first: run a harmless command
	args := host.sshArgs("echo ok")
	cmd := exec.Command("ssh", args...)
	out, err := cmd.CombinedOutput()
	if err == nil && strings.TrimSpace(string(out)) == "ok" {
		fmt.Printf("\033[32m✓ 密钥认证成功\033[0m\n")
		return nil
	}

	// Key auth failed — prompt for password
	fmt.Printf("\033[33m需要输入密码\033[0m\n")
	for attempt := 0; attempt < 3; attempt++ {
		fmt.Print("密码: ")
		pw, err := readPassword()
		if err != nil {
			return fmt.Errorf("读取密码失败: %w", err)
		}
		if pw == "" {
			fmt.Println("\033[31m密码不能为空\033[0m")
			continue
		}

		// Verify password by running a test command
		args = host.sshArgs("echo ok")
		cmd = exec.Command("ssh", args...)
		cmd.Stdin = strings.NewReader(pw + "\n")
		out, err = cmd.CombinedOutput()
		if err == nil && strings.TrimSpace(string(out)) == "ok" {
			fmt.Printf("\033[32m✓ 密码认证成功\033[0m\n")
			host.Password = pw
			return nil
		}

		fmt.Printf("\033[31m密码错误，还剩 %d 次机会\033[0m\n", 2-attempt)
	}

	return fmt.Errorf("无法连接到 %s，认证失败", host.Address())
}

// readPassword reads a line from stdin without echo (best-effort).
func readPassword() (string, error) {
	b, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(b), nil
}

func SelectHost() (*SSHHost, error) {
	hosts, err := ParseSSHConfig()
	if err != nil {
		return nil, fmt.Errorf("无法读取 ~/.ssh/config: %w", err)
	}
	if len(hosts) == 0 {
		return nil, fmt.Errorf("~/.ssh/config 中未找到主机")
	}

	fmt.Println("\n可用 SSH 主机:\n")
	for i, h := range hosts {
		detail := h.HostName
		if h.User != "" {
			detail = h.User + "@" + detail
		}
		if h.Port != "" && h.Port != "22" {
			detail += ":" + h.Port
		}
		fmt.Printf("  [%d] %-20s (%s)\n", i+1, h.Name, detail)
	}
	fmt.Printf("\n选择主机 [1-%d]: ", len(hosts))

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	var idx int
	_, err = fmt.Sscanf(input, "%d", &idx)
	if err != nil || idx < 1 || idx > len(hosts) {
		return nil, fmt.Errorf("无效选择: %s", input)
	}

	selected := &hosts[idx-1]
	return selected, nil
}

func clearScreen() {
	fmt.Print("\033[2J\033[H")
}
