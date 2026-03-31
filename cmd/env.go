package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/I-invincib1e/httli/internal/config"
)

var EnvCmd = &Command{
	Use:   "env",
	Short: "Manage local environments",
}

var EnvListCmd = &Command{
	Use:   "list",
	Short: "List loaded variables (from .env, .env.local, etc.)",
	Run: func(args []string) {
		// Parse to get the env name flag and trigger LoadEnv
		cfg, err := config.ParseFlags(args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		envName := cfg.Env
		if envName != "" {
			fmt.Printf("Active Environment: %s\n\n", envName)
		} else {
			fmt.Printf("Active Environment: (default)\n\n")
		}

		// Collect env vars that httli is likely responsible for (from .env files)
		// We show all HTTLI_* vars plus any that were set by LoadEnv
		// Since Go doesn't track which vars came from files, we display the known
		// HTTLI chain variables and any user-defined ones from the env files
		fmt.Println("Loaded files (in order):")
		printIfExists(".env")
		printIfExists(".env.local")
		if envName != "" {
			printIfExists(".env." + envName)
		}

		// Show HTTLI chain vars if set
		chainVars := []string{"HTTLI_LAST_STATUS", "HTTLI_LAST_BODY_PATH", "HTTLI_LAST_JSON"}
		hasChain := false
		for _, k := range chainVars {
			if v := os.Getenv(k); v != "" {
				if !hasChain {
					fmt.Println("\nChaining Variables:")
					hasChain = true
				}
				display := v
				if len(display) > 60 {
					display = display[:57] + "..."
				}
				fmt.Printf("  %s = %s\n", k, display)
			}
		}

		// Show all env vars that look like they could be from a .env file
		// (all-caps with underscores, not system vars)
		fmt.Println("\nAll Environment Variables (use --verbose to see all):")
		var keys []string
		for _, pair := range os.Environ() {
			parts := strings.SplitN(pair, "=", 2)
			if len(parts) == 2 {
				k := parts[0]
				// Only show ALL_CAPS_WITH_UNDERSCORES style vars (typical .env pattern)
				if isEnvFileStyle(k) && !strings.HasPrefix(k, "HTTLI_") {
					keys = append(keys, k)
				}
			}
		}
		sort.Strings(keys)
		if len(keys) == 0 {
			fmt.Println("  (none found matching .env file patterns)")
		}
		for _, k := range keys {
			v := os.Getenv(k)
			if len(v) > 60 {
				v = v[:57] + "..."
			}
			fmt.Printf("  %s = %s\n", k, v)
		}

		os.Exit(0)
	},
}

func printIfExists(filename string) {
	if _, err := os.Stat(filename); err == nil {
		fmt.Printf("  ✓ %s\n", filename)
	} else {
		fmt.Printf("  ✗ %s (not found)\n", filename)
	}
}

// isEnvFileStyle returns true if the key looks like a .env-style variable
// (uppercase letters, digits, underscores only, doesn't start with _)
func isEnvFileStyle(key string) bool {
	if len(key) == 0 || key[0] == '_' {
		return false
	}
	for _, c := range key {
		if !((c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_') {
			return false
		}
	}
	return true
}

func init() {
	EnvCmd.Run = func(args []string) { EnvCmd.PrintHelp() }
	EnvCmd.AddCommand(EnvListCmd)
	RootCmd.AddCommand(EnvCmd)
}
