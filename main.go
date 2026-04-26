package main

import (
	"os"

	"github.com/SeSiTing/siti-cli/cmd"
)

func main() {
	os.Exit(cmd.Execute(version))
}
