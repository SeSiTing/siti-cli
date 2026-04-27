package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// versionCmd mirrors `siti --version` as a subcommand,
// matching the convention of gh / kubectl / docker / git.
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "显示版本号",
	Args:  cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		fmt.Printf("siti version %s\n", rootCmd.Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
