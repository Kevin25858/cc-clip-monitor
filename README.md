# cc-clip-monitor

[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

[cc-clip](https://github.com/ShunmeiCho/cc-clip) 的 Windows 扩展——监听剪贴板截图，自动通过 SSH 上传到远程服务器，让你在终端里通过 SSH 运行的 Claude Code、Codex CLI、opencode 等工具中直接粘贴图片。

## 功能

- **剪贴板监听** — 自动检测截图，MD5 去重避免重复上传
- **自动上传** — 通过 `cc-clip send` 或 SCP 上传到远程服务器
- **路径回写** — 上传后将远程路径写入剪贴板
- **自动粘贴** — 切换到终端窗口时自动 Ctrl+V 粘贴路径
- **Toast 通知** — Windows 桌面通知提示上传结果
- **远程清理** — 自动删除旧上传文件
- **SSH 主机** — 从 `~/.ssh/config` 交互选择，支持中文主机名

## 下载

从 [Releases](https://github.com/Kevin25858/cc-clip-monitor/releases) 下载 `cc-clip-monitor.exe`，放到任意目录即可使用。

## 工作流程

1. 双击运行 `cc-clip-monitor.exe`，从 SSH 主机列表中选择目标服务器
2. 在任意应用中截图或复制图片到剪贴板
3. 工具自动检测 → 上传到远程 `~/.cache/cc-clip/uploads` 目录
4. 弹出 Toast 通知，远程路径已写入剪贴板
5. 切换到 Claude Code / Codex CLI 窗口，路径自动粘贴

## 致谢

本项目基于 [cc-clip](https://github.com/ShunmeiCho/cc-clip)（by [@ShunmeiCho](https://github.com/ShunmeiCho)），为其 Windows 平台的扩展实现。感谢原作者的优秀工作。
