package cmd

import (
	"fmt"

	"github.com/SeSiTing/siti-cli/internal/shell"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:       "init [zsh|bash|fish]",
	Short:     "输出 shell wrapper 配置（添加到 shell 配置文件）",
	Args:      cobra.MaximumNArgs(1),
	ValidArgs: []string{"zsh", "bash", "sh", "fish"},
	Long: `输出 shell wrapper 函数，使 'siti ai switch' 和 'siti proxy on/off'
能够修改当前终端的环境变量。

用法:
  # 查看内容
  siti init zsh

  # 添加到 zshrc（推荐）
  echo 'eval "$(siti init zsh)"' >> ~/.zshrc
  source ~/.zshrc`,
	RunE: func(cmd *cobra.Command, args []string) error {
		shellType := "zsh"
		if len(args) > 0 {
			shellType = args[0]
		}

		switch shellType {
		case "zsh", "bash", "sh", "fish":
			fmt.Println(shell.WrapperFor(shellType))
		default:
			return fmt.Errorf("不支持的 shell 类型: %s\n支持: zsh, bash, fish", shellType)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
