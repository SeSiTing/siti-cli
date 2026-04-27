package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var ipCmd = &cobra.Command{
	Use:   "ip",
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
		fmt.Println(" ", lookupPublicIP())
	},
}

// lookupPublicIP tries multiple endpoints; returns the first plain-text IP it gets.
// Reading the full body (not just 64 bytes) avoids truncating HTML error pages.
func lookupPublicIP() string {
	endpoints := []string{
		"https://api.ipify.org",
		"https://ifconfig.me/ip",
		"https://ipinfo.io/ip",
	}
	client := &http.Client{Timeout: 5 * time.Second}
	for _, url := range endpoints {
		resp, err := client.Get(url)
		if err != nil {
			continue
		}
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 64))
		resp.Body.Close()
		ip := strings.TrimSpace(string(body))
		if ip != "" && !strings.ContainsAny(ip, "<> ") {
			return ip
		}
	}
	return "(获取失败)"
}

func init() {
	rootCmd.AddCommand(ipCmd)
}
