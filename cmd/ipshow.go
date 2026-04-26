package cmd

import (
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var ipshowCmd = &cobra.Command{
	Use:   "ipshow",
	Short: "显示内网和公网 IP 地址",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🌐 内网 IP:")
		out, err := exec.Command("ipconfig", "getifaddr", "en0").Output()
		if err == nil {
			fmt.Println(" ", strings.TrimSpace(string(out)))
		} else {
			fmt.Println("  (未获取到内网 IP)")
		}

		fmt.Println("🌎 公网 IP:")
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Get("https://ifconfig.me")
		if err != nil {
			fmt.Println("  (获取公网 IP 失败:", err, ")")
			return
		}
		defer resp.Body.Close()
		buf := make([]byte, 64)
		n, _ := resp.Body.Read(buf)
		fmt.Println(" ", strings.TrimSpace(string(buf[:n])))
	},
}

func init() {
	rootCmd.AddCommand(ipshowCmd)
}
