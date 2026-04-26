package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var cleanlogsCmd = &cobra.Command{
	Use:   "cleanlogs",
	Short: "清理当前目录下所有 *.log 文件",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		var deleted int
		err = filepath.Walk(cwd, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // skip inaccessible files
			}
			if !info.IsDir() && filepath.Ext(path) == ".log" {
				if removeErr := os.Remove(path); removeErr == nil {
					deleted++
				}
			}
			return nil
		})
		if err != nil {
			return err
		}

		fmt.Printf("🧹 已清理 %d 个日志文件 (*.log)\n", deleted)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(cleanlogsCmd)
}
