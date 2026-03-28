package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/I-invincib1e/httli/internal/client"
	"github.com/I-invincib1e/httli/internal/collections"
	"github.com/I-invincib1e/httli/internal/config"
	"github.com/I-invincib1e/httli/internal/history"
	"github.com/I-invincib1e/httli/internal/output"
	"github.com/I-invincib1e/httli/internal/styles"
)

var CollectionCmd = &Command{
	Use:     "collection",
	Short:   "Manage saved requests",
	Long:    "Save, run, list, show, update, delete, export, and import API request collections.",
	Aliases: []string{"col"},
}

var CollectionSaveCmd = &Command{
	Use:   "save",
	Short: "Save a new request (fails if exists)",
	Run: func(args []string) {
		if len(args) < 1 {
			fmt.Fprintf(os.Stderr, "Error: missing name\nUsage: httli collection save <name> [options]\n")
			os.Exit(1)
		}
		name := args[0]
		cfg, err := config.ParseFlags(args[1:])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if err := collections.SaveRequest(name, cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Successfully saved request '%s'\n", name)
		os.Exit(0)
	},
}

var CollectionUpdateCmd = &Command{
	Use:   "update",
	Short: "Update an existing request",
	Run: func(args []string) {
		if len(args) < 1 {
			fmt.Fprintf(os.Stderr, "Error: missing name\nUsage: httli collection update <name> [options]\n")
			os.Exit(1)
		}
		name := args[0]
		cfg, err := config.ParseFlags(args[1:])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if err := collections.UpdateRequest(name, cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Successfully updated request '%s'\n", name)
		os.Exit(0)
	},
}

var CollectionDeleteCmd = &Command{
	Use:   "delete",
	Short: "Delete a saved request",
	Run: func(args []string) {
		if len(args) < 1 {
			fmt.Fprintf(os.Stderr, "Error: missing name\nUsage: httli collection delete <name>\n")
			os.Exit(1)
		}
		name := args[0]
		if err := collections.DeleteRequest(name); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Successfully deleted request '%s'\n", name)
		os.Exit(0)
	},
}

var CollectionShowCmd = &Command{
	Use:   "show",
	Short: "Display visually formatted saved request",
	Run: func(args []string) {
		if len(args) < 1 {
			fmt.Fprintf(os.Stderr, "Error: missing name\nUsage: httli collection show <name>\n")
			os.Exit(1)
		}
		name := args[0]
		cfg, err := collections.GetRequest(name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		st := styles.New()
		output.DisplayRequest(cfg, st)
		os.Exit(0)
	},
}

var CollectionListCmd = &Command{
	Use:   "list",
	Short: "List all saved requests",
	Run: func(args []string) {
		collections.ListCollections()
		os.Exit(0)
	},
}

var CollectionRunCmd = &Command{
	Use:   "run",
	Short: "Execute a saved collection request",
	Run: func(args []string) {
		if len(args) < 1 {
			fmt.Fprintf(os.Stderr, "Error: missing name\nUsage: httli collection run <name> [flags]\n")
			os.Exit(1)
		}
		name := args[0]

		runCfg, err := config.ParseFlags(args[1:])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		cfg, err := collections.GetRequest(name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		cfg.IgnoreMissingEnv = runCfg.IgnoreMissingEnv
		cfg.Retry = runCfg.Retry
		cfg.RetryDelay = runCfg.RetryDelay
		cfg.DryRun = runCfg.DryRun
		cfg.Format = runCfg.Format
		cfg.Fail = runCfg.Fail
		cfg.Raw = runCfg.Raw
		cfg.Extract = runCfg.Extract
		cfg.Quiet = runCfg.Quiet
		cfg.StatusOnly = runCfg.StatusOnly
		cfg.Verbose = runCfg.Verbose
		
		if runCfg.Timeout != 30*time.Second && runCfg.Timeout != 0 {
			cfg.Timeout = runCfg.Timeout
		}

		if err := cfg.InterpolateAll(); err != nil {
			fmt.Fprintf(os.Stderr, "Error interpolating: %v\n", err)
			os.Exit(1)
		}

		if runCfg.Verbose {
			fmt.Fprintf(os.Stderr, "Loaded env file(s) intentionally.\n")
		}

		st := styles.New()
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

		history.Record(cfg.Method, cfg.URL, resp.StatusCode)

		if err := output.DisplayResponse(cfg, resp, st); err != nil {
			fmt.Fprintf(os.Stderr, "Error displaying response: %v\n", err)
			os.Exit(1)
		}

		if cfg.Fail && resp.StatusCode >= 400 {
			os.Exit(22)
		}

		os.Exit(0)
	},
}

var CollectionExportCmd = &Command{
	Use:   "export",
	Short: "Export collections to a JSON file",
	Run: func(args []string) {
		if len(args) < 1 {
			fmt.Fprintf(os.Stderr, "Usage: httli collection export <file.json>\n")
			os.Exit(1)
		}
		if err := collections.ExportCollections(args[0]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Collections exported to %s\n", args[0])
		os.Exit(0)
	},
}

var CollectionImportCmd = &Command{
	Use:   "import",
	Short: "Import collections from a JSON file",
	Long: `Import API collections from a shared JSON file.

Conflict modes:
  --merge      (default) Add new, skip existing
  --overwrite  Replace existing entries
  --skip       Skip all conflicts`,
	Run: func(args []string) {
		if len(args) < 1 {
			fmt.Fprintf(os.Stderr, "Usage: httli collection import <file.json> [--overwrite|--skip]\n")
			os.Exit(1)
		}
		mode := "merge"
		for _, a := range args[1:] {
			switch a {
			case "--overwrite":
				mode = "overwrite"
			case "--skip":
				mode = "skip"
			}
		}
		if err := collections.ImportCollections(args[0], mode); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	},
}

func init() {
	CollectionCmd.Run = func(args []string) {
		CollectionCmd.PrintHelp()
	}
	CollectionCmd.AddCommand(CollectionSaveCmd)
	CollectionCmd.AddCommand(CollectionUpdateCmd)
	CollectionCmd.AddCommand(CollectionDeleteCmd)
	CollectionCmd.AddCommand(CollectionShowCmd)
	CollectionCmd.AddCommand(CollectionListCmd)
	CollectionCmd.AddCommand(CollectionRunCmd)
	CollectionCmd.AddCommand(CollectionRunAllCmd) // Defined in collection_runall.go
	CollectionCmd.AddCommand(CollectionExportCmd)
	CollectionCmd.AddCommand(CollectionImportCmd)
	RootCmd.AddCommand(CollectionCmd)
}
