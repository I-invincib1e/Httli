package cmd

import (
	"fmt"
	"os"

	"github.com/I-invincib1e/httli/internal/client"
	"github.com/I-invincib1e/httli/internal/config"
	"github.com/I-invincib1e/httli/internal/history"
	"github.com/I-invincib1e/httli/internal/output"
	"github.com/I-invincib1e/httli/internal/styles"
)

var RequestCmd = &Command{
	Use:   "request",
	Short: "Make HTTP requests directly",
	Long:  "Send one-time HTTP requests to any URL with full flag support.",
}

var RequestSendCmd = &Command{
	Use:   "send",
	Short: "Send a new raw request",
	Run: func(args []string) {
		cfg, err := config.ParseFlags(args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing arguments: %v\n", err)
			os.Exit(1)
		}
		if err := cfg.Validate(); err != nil {
			fmt.Fprintf(os.Stderr, "Error validating arguments: %v\n", err)
			os.Exit(1)
		}

		if err := cfg.InterpolateAll(); err != nil {
			fmt.Fprintf(os.Stderr, "Error interpolating: %v\n", err)
			os.Exit(1)
		}

		st := styles.New()

		if cfg.Verbose {
			fmt.Fprintf(os.Stderr, "Loaded env file(s) internally.\n")
		}

		output.DisplayRequest(cfg, st)

		if cfg.DryRun {
			fmt.Println("\n[Dry Run] Request fully interpolated, no network call made.")
			os.Exit(0)
		}

		resp, err := client.ExecuteRequest(cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error executing request: %v\n", err)
			os.Exit(1)
		}

		history.Record(cfg, resp.StatusCode, resp.Duration.Milliseconds())

		if err := output.DisplayResponse(cfg, resp, st); err != nil {
			fmt.Fprintf(os.Stderr, "Error displaying response: %v\n", err)
			os.Exit(1)
		}

		// --fail: exit 22 on HTTP 4xx/5xx (curl convention)
		if cfg.Fail && resp.StatusCode >= 400 {
			os.Exit(22)
		}

		os.Exit(0)
	},
}

func init() {
	RequestCmd.Run = func(args []string) {
		RequestCmd.PrintHelp()
	}
	RequestCmd.AddCommand(RequestSendCmd)
	RootCmd.AddCommand(RequestCmd)
}
