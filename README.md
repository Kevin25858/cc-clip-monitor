# cc-clip-monitor

Windows 剪贴板监听工具，作为 [cc-clip](https://github.com/ShunmeiCho/cc-clip) 的 Windows 扩展——自动将截图上传到远程服务器，让你在 Claude Code / Codex CLI 中直接粘贴图片。

## 功能

- 剪贴板图片监听（MD5 去重）
- 自动上传到远程服务器
- 远程路径回写剪贴板
- Windows Toast 通知
- 窗口切换自动粘贴
- 远程旧文件自动清理
- 支持中文 SSH 主机名
- 启动时交互选择 SSH 主机

## 功能

- 剪贴板图片监听（MD5 去重）
- 自动上传到远程服务器
- 远程路径回写剪贴板
- Windows Toast 通知
- 窗口切换自动粘贴
- 远程旧文件自动清理
- 支持中文 SSH 主机名
- 启动时交互选择 SSH 主机

## 下载

从 [Releases](https://github.com/Kevin25858/cc-clip-monitor/releases) 下载 `cc-clip-monitor.exe`，放到任意目录即可使用。

## 工作流程

1. 启动后从 `~/.ssh/config` 选择主机
2. 复制/截图图片
3. 自动上传到远程服务器
4. Toast 通知 + 远程路径写入剪贴板
5. 点击终端窗口 → 自动粘贴路径

## 依赖

- Windows 10/11
- [cc-clip](https://github.com/ShunmeiCho/cc-clip)（默认模式）或 `ssh` + `scp`
