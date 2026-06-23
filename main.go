package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"syscall"
	"time"
)

var (
	kernel32Mod       = syscall.NewLazyDLL("kernel32.dll")
	setConsoleOutputCP = kernel32Mod.NewProc("SetConsoleOutputCP")
	setConsoleCP      = kernel32Mod.NewProc("SetConsoleCP")
)

var version = "1.1.0"

func main() {
	setConsoleOutputCP.Call(65001)
	setConsoleCP.Call(65001)

	noPaste := flag.Bool("no-paste", false, "禁用窗口切换自动粘贴")
	noCleanup := flag.Bool("no-cleanup", false, "禁用远程文件清理")
	useSCP := flag.Bool("scp", false, "使用 SCP 直连（不依赖 cc-clip）")
	pollMs := flag.Int("poll-ms", 300, "剪贴板轮询间隔（毫秒）")
	showVersion := flag.Bool("version", false, "显示版本")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "cc-clip-monitor — SSH 剪贴板图片自动上传工具\n\n")
		fmt.Fprintf(os.Stderr, "用法: cc-clip-monitor [主机名] [选项]\n\n")
		fmt.Fprintf(os.Stderr, "不指定主机名时，从 ~/.ssh/config 中选择\n\n")
		fmt.Fprintf(os.Stderr, "选项:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if *showVersion {
		fmt.Printf("cc-clip-monitor %s\n", version)
		return
	}

	var host *SSHHost
	if flag.NArg() > 0 {
		name := flag.Arg(0)
		h, err := findSSHHost(name)
		if err != nil {
			host = &SSHHost{Name: name, Port: "22"}
		} else {
			host = h
		}
	} else {
		selected, err := SelectHost()
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			fmt.Println("\n按回车键退出...")
			bufio.NewReader(os.Stdin).ReadBytes('\n')
			os.Exit(1)
		}
		host = selected
	}

	// Test SSH connection
	if err := TestSSHConnection(host); err != nil {
		fmt.Fprintf(os.Stderr, "\033[31m错误: %v\033[0m\n", err)
		fmt.Println("\n按回车键退出...")
		os.Stdin.Read(make([]byte, 1))
		os.Exit(1)
	}

	remoteDir := "~/.cache/cc-clip/uploads"
	useCCClip := !*useSCP && HasCCClip()

	clearScreen()
	fmt.Println("\033[32mcc-clip-monitor 已启动\033[0m")
	fmt.Printf("主机: %s", host.Name)
	if host.HostName != "" && host.HostName != host.Name {
		fmt.Printf(" (%s)", host.Address())
	}
	fmt.Println()
	mode := "SCP 直连"
	if useCCClip {
		mode = "cc-clip send"
	}
	fmt.Printf("模式: %s\n", mode)
	fmt.Println("复制图片 → 上传 → 点击终端 → 自动粘贴")
	fmt.Println("按 Ctrl+C 停止\n")

	monitor, err := NewClipboardMonitor()
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 剪贴板监听初始化失败: %v\n", err)
		fmt.Println("\n按回车键退出...")
		os.Stdin.Read(make([]byte, 1))
		os.Exit(1)
	}

	uploader := NewUploader(host, remoteDir, useCCClip)
	paster := NewAutoPaster()

	ticker := time.NewTicker(time.Duration(*pollMs) * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		imgPath, isNew := monitor.CheckNewImage()
		if !isNew {
			if paster.CheckAndPaste() {
				ts := time.Now().Format("15:04:05")
				fmt.Printf("\033[32m[%s] 已自动粘贴!\033[0m\n", ts)
			}
			continue
		}

		ts := time.Now().Format("15:04:05")
		fmt.Printf("\033[36m[%s] 检测到图片，上传中...\033[0m\n", ts)

		if !*noCleanup {
			go CleanupRemote(host, remoteDir)
			go CleanupLocal(monitor.tempDir)
		}

		remotePath, err := uploader.Upload(imgPath)
		if err != nil {
			fmt.Printf("\033[31m[%s] 上传失败: %v\033[0m\n", ts, err)
			NotifyFailure(err.Error())
			continue
		}

		if err := monitor.SetText(remotePath); err != nil {
			fmt.Printf("\033[31m[%s] 剪贴板写入失败: %v\033[0m\n", ts, err)
		}

		fmt.Printf("\033[32m[%s] 就绪: %s\033[0m\n", ts, remotePath)
		NotifyUpload(remotePath)

		if !*noPaste {
			paster.MarkReady()
		}
	}
}

func findSSHHost(name string) (*SSHHost, error) {
	hosts, err := ParseSSHConfig()
	if err != nil {
		return nil, err
	}
	for _, h := range hosts {
		if h.Name == name {
			return &h, nil
		}
	}
	return nil, fmt.Errorf("未找到主机: %s", name)
}
