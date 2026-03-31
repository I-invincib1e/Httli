package cmd

import (
	"fmt"
	"os"
)

// Version is the current version string.
// Injected at build time via:
//   go build -ldflags "-X github.com/I-invincib1e/httli/cmd.Version=1.1.0" ./cmd/httli
var Version = "dev"

var VersionCmd = &Command{
	Use:   "version",
	Short: "Print the version of httli",
	Run: func(args []string) {
		fmt.Printf("httli v%s\n", Version)
		os.Exit(0)
	},
}

func init() {
	RootCmd.AddCommand(VersionCmd)
}
