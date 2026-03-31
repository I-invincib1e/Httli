package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/I-invincib1e/httli/internal/client"
	"github.com/I-invincib1e/httli/internal/collections"
	"github.com/I-invincib1e/httli/internal/config"
	"github.com/I-invincib1e/httli/internal/history"
	"github.com/I-invincib1e/httli/internal/styles"
)


var CollectionRunAllCmd = &Command{
	Use:   "run-all",
	Short: "Run multiple saved requests sequentially",
	Long: `Run all saved requests matching an optional prefix (e.g. 'auth/').
Supports passing state between requests via HTTLI_LAST_STATUS and HTTLI_LAST_BODY_PATH.

Examples:
  httli col run-all              # runs everything
  httli col run-all auth         # runs auth/*
  httli col run-all --fail-fast  # stop on first error`,
	Run: func(args []string) {
		// Basic parsing to separate prefix logic from flag logic
		prefix := ""
		if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
			prefix = args[0]
			args = args[1:]
		}

		// Parse the rest of flags (--fail-fast is now a proper registered flag)
		runCfg, err := config.ParseFlags(args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		allNames, err := collections.ListAllNames()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if len(allNames) == 0 {
			fmt.Println("No collections found to run.")
			os.Exit(0)
		}

		// Filter by prefix
		var targets []string
		for _, name := range allNames {
			if prefix == "" || strings.HasPrefix(name, prefix) {
				targets = append(targets, name)
			}
		}

		if len(targets) == 0 {
			fmt.Printf("No collections found matching prefix '%s'\n", prefix)
			os.Exit(0)
		}

		type runResult struct {
			Name       string `json:"name"`
			Method     string `json:"method"`
			URL        string `json:"url"`
			StatusCode int    `json:"status_code"`
			DurationMs int64  `json:"duration_ms"`
			Error      string `json:"error,omitempty"`
		}
		var results []runResult

		st := styles.New()
		if runCfg.Format != "json" {
			fmt.Printf("Running %d requests...\n\n", len(targets))
		}

		tmpBodyPath := os.TempDir() + "/httli_last_body.json"

		for i, name := range targets {
			if runCfg.Format != "json" {
				fmt.Printf("[%d/%d] %s...\n", i+1, len(targets), name)
			}

			// Clean any previous temp state
			os.Remove(tmpBodyPath)

			cfg, err := collections.GetRequest(name)
			if err != nil {
				res := runResult{Name: name, Error: err.Error()}
				results = append(results, res)
				if runCfg.FailFast { break }
				continue
			}

			// Merge overriding configs
			cfg.ApplyOverrides(runCfg)

			// Interpolate AFTER env has been updated by previous runs!
			if err := cfg.InterpolateAll(); err != nil {
				res := runResult{Name: name, Error: err.Error()}
				results = append(results, res)
				if runCfg.FailFast { break }
				continue
			}

			start := time.Now()
			resp, err := client.ExecuteRequest(cfg)
			duration := time.Since(start).Milliseconds()

			if err != nil {
				res := runResult{Name: name, Method: cfg.Method, URL: cfg.URL, Error: err.Error()}
				results = append(results, res)
				if runCfg.FailFast { break }
				continue
			}

			history.Record(cfg.Method, cfg.URL, resp.StatusCode)

			res := runResult{
				Name:       name,
				Method:     cfg.Method,
				URL:        cfg.URL,
				StatusCode: resp.StatusCode,
				DurationMs: duration,
			}
			results = append(results, res)

			// --- Setup Chaining Env Vars ---
			os.Setenv("HTTLI_LAST_STATUS", fmt.Sprintf("%d", resp.StatusCode))
			
			// Always write to temp file
			if err := os.WriteFile(tmpBodyPath, resp.Body, 0644); err == nil {
				os.Setenv("HTTLI_LAST_BODY_PATH", tmpBodyPath)
			}

			// If JSON and small (<32KB), expose directly in HTTLI_LAST_JSON
			if len(resp.Body) > 0 && len(resp.Body) <= 32768 {
				if json.Valid(resp.Body) {
					os.Setenv("HTTLI_LAST_JSON", string(resp.Body))
				}
			} else {
				os.Unsetenv("HTTLI_LAST_JSON")
			}

			if runCfg.FailFast && resp.StatusCode >= 400 {
				if runCfg.Format != "json" {
					fmt.Printf("  -> Stopped due to --fail-fast (HTTP %d)\n", resp.StatusCode)
				}
				break
			}
		}

		// Print Summary
		if runCfg.Format == "json" {
			data, _ := json.MarshalIndent(results, "", "  ")
			fmt.Println(string(data))
		} else {
			var successful, failed int
			var totalTime int64

			fmt.Println("\nRun Summary:")
			fmt.Println("------------------------------------------------")
			for _, r := range results {
				totalTime += r.DurationMs
				if r.Error != "" {
					failed++
					fmt.Printf("%s\t%s\tERROR: %s\n", st.Error.Render("FAIL"), r.Name, r.Error)
					continue
				}
				
				statusLabel := st.Success.Render(fmt.Sprintf("%d", r.StatusCode))
				if r.StatusCode >= 400 {
					failed++
					statusLabel = st.Error.Render(fmt.Sprintf("%d", r.StatusCode))
				} else {
					successful++
				}
				fmt.Printf("%s\t%s\t%dms\n", statusLabel, r.Name, r.DurationMs)
			}
			fmt.Println("------------------------------------------------")
			fmt.Printf("Summary:\n")
			fmt.Printf("  Total:   %d\n", len(results))
			fmt.Printf("  Success: %s\n", st.Success.Render(fmt.Sprintf("%d", successful)))
			if failed > 0 {
				fmt.Printf("  Failed:  %s\n", st.Error.Render(fmt.Sprintf("%d", failed)))
			} else {
				fmt.Printf("  Failed:  %d\n", failed)
			}
			fmt.Printf("  Time:    %dms\n", totalTime)
		}

		// Clean up temp file
		os.Remove(tmpBodyPath)

		// Apply final failure state
		if runCfg.Fail {
			for _, r := range results {
				if r.Error != "" || r.StatusCode >= 400 {
					os.Exit(22)
				}
			}
		}

		os.Exit(0)
	},
}
