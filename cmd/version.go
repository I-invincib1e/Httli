package cmd

import (
	"fmt"
	"os"
)

var Version = "1.0.0"

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
