package cmd

import (
	"fmt"
	"os"

	"github.com/I-invincib1e/http-cli/internal/config"
)

var EnvCmd = &Command{
	Use:   "env",
	Short: "Manage local environments",
	Run: func(args []string) {
		fmt.Fprintf(os.Stderr, "Use 'http-cli env --help' to see subcommands.\n")
	},
}

var EnvListCmd = &Command{
	Use:   "list",
	Short: "List loaded variables (from .env, .env.local, etc)",
	Run: func(args []string) {
		// Just parse empty flags to trigger LoadEnv
		cfg, _ := config.ParseFlags([]string{})
		
		fmt.Printf("Active Environment: %s\n", cfg.Env)
		fmt.Println("Variables are loaded automatically per command.")
		os.Exit(0)
	},
}

func init() {
	EnvCmd.AddCommand(EnvListCmd)
	RootCmd.AddCommand(EnvCmd)
}
