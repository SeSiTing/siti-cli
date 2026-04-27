package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// evalKey is the context key for the per-invocation eval buffer.
type evalKey struct{}

// evalBuffer accumulates shell statements queued by Eval().
type evalBuffer struct{ lines []string }

// Eval queues one or more shell statement strings to be evaluated by the
// parent shell wrapper (exit-code-10 protocol).
//
// Call this from any RunE instead of returning an error-typed directive.
// The command should return nil afterwards; Execute() will detect the buffer
// and exit with code 10 after printing the lines to stdout.
//
// Example:
//
//	cmd.Eval(c, shell.Export("http_proxy", proxyURL))
//	return nil
func Eval(c *cobra.Command, lines ...string) {
	if buf, ok := c.Context().Value(evalKey{}).(*evalBuffer); ok {
		buf.lines = append(buf.lines, lines...)
	}
}

var rootCmd = &cobra.Command{
	Use:   "siti",
	Short: "个人 CLI 工具集",
	Long:  "siti — 个人命令行助手，支持 AI 切换、代理管理、端口清理等便捷操作。",
}

// Execute runs the root command and returns an exit code.
// os.Exit is called exactly once, in main.go.
func Execute(ver string) int {
	rootCmd.Version = ver

	buf := &evalBuffer{}
	ctx := context.WithValue(context.Background(), evalKey{}, buf)

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		// Real errors are already printed by Cobra.
		return 1
	}

	if len(buf.lines) > 0 {
		// When SITI_EVAL_FILE is set (by the shell wrapper), write shell
		// statements there so child processes can inherit the real TTY on stdout.
		// When not set (no wrapper / direct invocation), fall back to stdout.
		if evalFile := os.Getenv("SITI_EVAL_FILE"); evalFile != "" {
			data := []byte{}
			for _, line := range buf.lines {
				data = append(data, line...)
				data = append(data, '\n')
			}
			if err := os.WriteFile(evalFile, data, 0o600); err != nil {
				fmt.Fprintf(os.Stderr, "✗ write eval file: %v\n", err)
				return 1
			}
		} else {
			for _, line := range buf.lines {
				fmt.Println(line)
			}
		}
		return 10
	}
	return 0
}

func init() {
	rootCmd.SilenceErrors = true // handled in Execute()
	rootCmd.SilenceUsage = true  // don't dump usage on every error
	rootCmd.CompletionOptions.DisableDefaultCmd = false
}

// printErr writes a formatted message to stderr (human-visible output).
func printErr(format string, a ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", a...)
}
