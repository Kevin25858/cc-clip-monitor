# AGENTS.md

## What this is

Windows-only clipboard image monitor that auto-uploads screenshots to a remote Linux server via SSH. Companion tool to [cc-clip](https://github.com/ShunmeiCho/cc-clip). Single Go binary, zero external dependencies.

## Build & run

```powershell
go build -o cc-clip-monitor.exe .   # produces ~3.6MB binary
.\cc-clip-monitor.exe               # interactive host selection
.\cc-clip-monitor.exe <host>        # direct mode
.\run.bat                           # launcher with pause on exit
```

No tests exist. Verify by running the binary and copying an image to clipboard.

## Architecture (all files are `package main`)

| File | Responsibility |
|------|---------------|
| `main.go` | Entry, flag parsing, main ticker loop, host resolution |
| `clipboard.go` | Windows clipboard read via raw syscall (OpenClipboard/GetClipboardData/GlobalLock), DIB→PNG conversion, MD5 dedup |
| `autopaste.go` | Window focus detection (GetForegroundWindow), Ctrl+V simulation (keybd_event), clipboard text write (SetClipboardData) |
| `uploader.go` | Two upload modes: `cc-clip send <alias>` or direct SCP using resolved HostName:Port |
| `sshconfig.go` | Parses `~/.ssh/config`, interactive host selection, SSHHost struct |
| `cleaner.go` | `ssh <host> "rm -f ..."` to delete old uploads |
| `notifier.go` | Toast notifications via PowerShell subprocess (not native Go) |

## Key design decisions

- **Zero dependencies**: all Windows API calls use `syscall.NewLazyDLL` — do NOT add `golang.org/x/sys` or other external modules (network may be blocked during `go get`)
- **Chinese host names**: SCP mode resolves real HostName/Port/User from SSH config to bypass encoding issues; cc-clip mode passes the alias directly (cc-clip handles it internally)
- **Console UTF-8**: `SetConsoleOutputCP(65001)` + `SetConsoleCP(65001)` at startup — required for Chinese UI strings
- **Toast via PowerShell**: `notifier.go` shells out to `powershell -NoProfile -Command` with Windows Runtime XML — no native Go toast library
- **Clipboard polling**: 300ms default ticker, MD5 hash of raw DIB data for dedup

## Windows API usage pattern

All Windows API calls follow this pattern:
```go
var user32 = syscall.NewLazyDLL("user32.dll")
var someFunc = user32.NewProc("SomeFunction")
r1, _, _ := someProc.Call(arg1, arg2)
```

Clipboard functions are in `clipboard.go`, keyboard/focus in `autopaste.go`. `kernel32` and `user32` DLL handles are package-level vars shared across files.

## SSH config parsing

`sshconfig.go` reads `~/.ssh/config` line-by-line. Only extracts: `Host`, `HostName`, `Port`, `User`. Wildcard hosts (`*`) are skipped. Port defaults to `"22"`.

## Conventions

- All user-facing strings are in Chinese (Simplified)
- ANSI color codes used for terminal output (`\033[32m` green, `\033[36m` cyan, `\033[31m` red)
- Error messages include both Chinese label and underlying error
- Remote upload dir: `~/.cache/cc-clip/uploads`
