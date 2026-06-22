package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type SSHHost struct {
	Name     string
	HostName string
	Port     string
	User     string
}

func ParseSSHConfig() ([]SSHHost, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(home, ".ssh", "config")
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
		return fmt.Sprintf("%s:%s", host, h.Port)
	}
	return host
}

func (h SSHHost) SCPDest() string {
	host := h.HostName
	if host == "" {
		host = h.Name
	}
	if h.Port != "" && h.Port != "22" {
		return fmt.Sprintf("-P %s", h.Port)
	}
	return ""
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
	if runtime.GOOS == "windows" {
		fmt.Print("\033[2J\033[H")
	}
}
