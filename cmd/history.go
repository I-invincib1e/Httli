package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/I-invincib1e/httli/internal/config"
	"github.com/I-invincib1e/httli/internal/history"
	"github.com/I-invincib1e/httli/internal/client"
	"github.com/I-invincib1e/httli/internal/output"
	"github.com/I-invincib1e/httli/internal/styles"
)

var HistoryCmd = &Command{
	Use:   "history",
	Short: "View request history",
	Long: `View, inspect, rerun, or clear your executed request history.

The last 50 requests are saved automatically with status codes and timestamps.`,
	Run: func(args []string) {
		history.List()
		os.Exit(0)
	},
}

var HistoryShowCmd = &Command{
	Use:   "show",
	Short: "Show details of a history entry",
	Run: func(args []string) {
		if len(args) < 1 {
			fmt.Fprintf(os.Stderr, "Usage: httli history show <index>\n")
			os.Exit(1)
		}
		idx, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid index %q\n", args[0])
			os.Exit(1)
		}
		if _, err := history.Show(idx); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	},
}

var HistoryClearCmd = &Command{
	Use:   "clear",
	Short: "Clear all request history",
	Run: func(args []string) {
		if err := history.Clear(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("History cleared.")
		os.Exit(0)
	},
}

var RerunCmd = &Command{
	Use:   "rerun",
	Short: "Re-execute a request from history",
	Long:  "Re-execute a previous request by its history index number.",
	Run: func(args []string) {
		if len(args) < 1 {
			fmt.Fprintf(os.Stderr, "Usage: httli rerun <index>\n")
			os.Exit(1)
		}
		idx, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid index %q\n", args[0])
			os.Exit(1)
		}
		entry, err := history.Show(idx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		cfg := &config.Config{
			Method:  entry.Method,
			URL:     entry.URL,
			Headers: make(map[string]string),
			Timeout: 30,
		}

		st := styles.New()
		output.DisplayRequest(cfg, st)

		resp, err := client.ExecuteRequest(cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		history.Record(cfg.Method, cfg.URL, resp.StatusCode)

		if err := output.DisplayResponse(cfg, resp, st); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	},
}

func init() {
	HistoryCmd.AddCommand(HistoryShowCmd)
	HistoryCmd.AddCommand(HistoryClearCmd)
	RootCmd.AddCommand(HistoryCmd)
	RootCmd.AddCommand(RerunCmd)
}
