# cc-clip-monitor

Windows 剪贴板图片自动上传工具，作为 [cc-clip](https://github.com/ShunmeiCho/cc-clip) 的 Windows 扩展。

## 功能

- 剪贴板图片监听（MD5 去重）
- 自动上传到远程服务器
- 远程路径回写剪贴板
- Windows Toast 通知
- 窗口切换自动粘贴
- 远程旧文件自动清理
- 支持中文 SSH 主机名
- 启动时交互选择 SSH 主机

## 安装

```powershell
# 添加到 PATH
$env:PATH += ";D:\EXE\cc-clip-monitor"
```

## 使用

```powershell
# 交互选择主机
cc-clip-monitor

# 指定主机
cc-clip-monitor zaozhuang

# 使用 SCP 直连（不依赖 cc-clip）
cc-clip-monitor zaozhuang --scp

# 禁用自动粘贴
cc-clip-monitor zaozhuang --no-paste

# 禁用远程清理
cc-clip-monitor zaozhuang --no-cleanup
```

## 工作流程

1. 启动后从 `~/.ssh/config` 选择主机
2. 复制/截图图片
3. 自动上传到远程服务器
4. Toast 通知 + 远程路径写入剪贴板
5. 点击终端窗口 → 自动粘贴路径

## 依赖

- Windows 10/11
- `cc-clip`（默认模式）或 `ssh` + `scp`（`--scp` 模式）

## 构建

```powershell
go build -o cc-clip-monitor.exe .
```
