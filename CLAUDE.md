# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

`siti-cli` 是一个 macOS/Linux 个人命令行工具集，用 Go + Cobra 实现，通过 Homebrew 或 goreleaser 发布。
支持 AI API 配置切换、代理管理、端口管理、网络检测等功能。

## 开发模式运行

在源码目录中直接运行（开发模式）：

```bash
go run . --help
go run . ai list
go run . proxy check
```

或先编译再运行：

```bash
go build -o siti .
./siti ai switch
```

## 版本发布流程

1. 修改 `version.go` 中的 `version` 变量（e.g. `"2.1.0"`）
2. 更新 `CHANGELOG.md`
3. 推送 main 分支，GitHub Actions（Publish on Version Bump）自动：
   - 检测版本号变更 → 调用 goreleaser → 交叉编译 → 打 tag → 创建 Release → 更新 Formula

若 tag 已存在但漏发 Release：Actions → Release from Tag → Run workflow，填写 tag 名（如 `v2.1.0`）

## 架构

### 技术栈

- **语言**: Go 1.24+
- **CLI 框架**: [cobra](https://github.com/spf13/cobra)
- **交互 UI**: [charmbracelet/huh](https://github.com/charmbracelet/huh)
- **发布**: [goreleaser](https://goreleaser.com/)

### 核心机制：EvalDirective + Shell Wrapper

部分命令（`ai switch`、`proxy on/off`）需要修改**调用方** shell 的环境变量。由于子进程无法修改父进程环境，采用以下协议：

1. Go 命令函数 `return shell.Eval(...)` — 返回一个实现了 `error` 接口的 `*EvalDirective`
2. `cmd/root.go` 的 `Execute()` 检测到该类型后，将 shell 语句打印到 stdout，以退出码 10 退出
3. Shell wrapper（通过 `siti init zsh` 安装并 source）检测到退出码 10，对 stdout 执行 `eval`，然后 `return 0`

```
用户运行: siti proxy on
→ wrapper 捕获 stdout + exit code
→ exit code=10 → eval stdout → 当前 shell 环境变量被修改 → return 0
→ exit code≠10 → 直接打印 stdout/stderr → return exit code
```

### 目录结构

```
main.go               # 入口：os.Exit(cmd.Execute(version))
version.go            # var version = "dev"（ldflags 注入）
cmd/
  root.go             # rootCmd + Execute() + EvalDirective 处理
  ai.go               # siti ai switch/list/current/test/unset
  proxy.go            # siti proxy on/off/check
  initcmd.go          # siti init zsh|bash|fish
  upgrade.go          # siti upgrade
  killports.go        # siti killports [ports/--dev/--db/...]
  brewup.go           # siti brewup
  netcheck.go         # siti netcheck
  ipshow.go           # siti ipshow
  cleanlogs.go        # siti cleanlogs
  util.go             # 公用工具函数
internal/
  shell/
    eval.go           # EvalDirective + Export/Unset/SourceIf 辅助函数
    wrapper.go        # posixWrapper / fishWrapper 模板字符串
  config/
    zshrc.go          # 解析 ~/.zshrc + ~/.zshenv，发现 AI Provider 列表
Formula/siti-cli.rb   # Homebrew Formula（由 goreleaser CI 自动更新）
completions/
  _siti               # zsh 补全（cobra 自动生成）
  siti.bash           # bash 补全（cobra 自动生成）
.goreleaser.yml       # 发布配置
```

### 新增命令规范

在 `cmd/` 下新建 `<命令名>.go`，在 `init()` 中注册到 `rootCmd`：

```go
var myCmd = &cobra.Command{
    Use:   "mycmd",
    Short: "一行描述",
    RunE: func(cmd *cobra.Command, args []string) error {
        // 普通输出
        fmt.Println("结果")
        return nil

        // 需要修改父 shell 环境变量时：
        // return shell.Eval(shell.Export("MY_VAR", "value"))
    },
}

func init() { rootCmd.AddCommand(myCmd) }
```

### AI 服务商配置

`siti ai` 从 `~/.zshrc` 和 `~/.zshenv` 读取 `*_BASE_URL` 变量自动发现服务商。
切换时通过 eval 协议修改 `ANTHROPIC_BASE_URL`、`ANTHROPIC_AUTH_TOKEN`（可选 `ANTHROPIC_MODEL`）。
跳过列表通过 `SITI_AI_SKIP` 变量（逗号分隔大写名称）配置，如 `export SITI_AI_SKIP=AZURE,OPENAI`。
