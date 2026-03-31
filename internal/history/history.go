package history

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/I-invincib1e/httli/internal/config"
	"github.com/I-invincib1e/httli/internal/storage"
)

const maxHistory = 50

// Entry stores full request metadata so rerun can faithfully replay requests.
type Entry struct {
	Timestamp   string            `json:"timestamp"`
	Method      string            `json:"method"`
	URL         string            `json:"url"`
	Headers     map[string]string `json:"headers,omitempty"`
	Body        string            `json:"body,omitempty"`
	BearerToken string            `json:"bearer_token,omitempty"`
	BasicAuth   string            `json:"basic_auth,omitempty"`
	Status      int               `json:"status"`
	DurationMs  int64             `json:"duration_ms"`
}

// historyPath returns the resolved path (project-local or global)
func historyPath() string {
	return storage.ResolvePath("history.json")
}

func loadHistory() ([]Entry, error) {
	data, err := os.ReadFile(historyPath())
	if err != nil {
		return nil, nil // no history yet is fine
	}
	var entries []Entry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

func saveHistory(entries []Entry) error {
	path := historyPath()
	if err := storage.EnsureDir(path); err != nil {
		return err
	}
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// Record appends a new entry to history (sliding window of maxHistory).
// Now takes the full config to enable faithful rerun.
func Record(cfg *config.Config, statusCode int, durationMs int64) {
	entries, _ := loadHistory()
	entry := Entry{
		Timestamp:   time.Now().Format(time.RFC3339),
		Method:      cfg.Method,
		URL:         cfg.URL,
		Headers:     cfg.Headers,
		Body:        cfg.Body,
		BearerToken: cfg.BearerToken,
		BasicAuth:   cfg.BasicAuth,
		Status:      statusCode,
		DurationMs:  durationMs,
	}
	entries = append(entries, entry)
	if len(entries) > maxHistory {
		entries = entries[len(entries)-maxHistory:]
	}
	_ = saveHistory(entries)
}

// ListWithFormat prints all history entries in the given format.
func ListWithFormat(format string) {
	entries, err := loadHistory()
	if err != nil || len(entries) == 0 {
		if format == "json" {
			fmt.Println("[]")
		} else {
			fmt.Println("No request history found.")
		}
		return
	}

	if format == "json" {
		data, err := json.MarshalIndent(entries, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return
		}
		fmt.Println(string(data))
		return
	}

	fmt.Println("Request History:")
	for i, e := range entries {
		statusIcon := "✓"
		if e.Status >= 400 {
			statusIcon = "✗"
		}
		durationStr := ""
		if e.DurationMs > 0 {
			durationStr = fmt.Sprintf(" (%dms)", e.DurationMs)
		}
		fmt.Printf("  [%d] %s %s %s → %d%s  %s\n",
			i+1, statusIcon, e.Method, e.URL, e.Status, durationStr, e.Timestamp)
	}
}

// Show prints details of a single history entry and returns it.
func Show(index int) (*Entry, error) {
	entries, err := loadHistory()
	if err != nil || len(entries) == 0 {
		return nil, fmt.Errorf("no request history found")
	}
	if index < 1 || index > len(entries) {
		return nil, fmt.Errorf("invalid history index %d (valid: 1-%d)", index, len(entries))
	}
	e := entries[index-1]
	fmt.Printf("History Entry #%d\n", index)
	fmt.Printf("  Timestamp:  %s\n", e.Timestamp)
	fmt.Printf("  Method:     %s\n", e.Method)
	fmt.Printf("  URL:        %s\n", e.URL)
	fmt.Printf("  Status:     %d\n", e.Status)
	if e.DurationMs > 0 {
		fmt.Printf("  Duration:   %dms\n", e.DurationMs)
	}
	if len(e.Headers) > 0 {
		fmt.Printf("  Headers:\n")
		for k, v := range e.Headers {
			fmt.Printf("    %s: %s\n", k, v)
		}
	}
	if e.Body != "" {
		fmt.Printf("  Body:       %s\n", e.Body)
	}
	return &e, nil
}

// Clear removes all history.
func Clear() error {
	path := historyPath()
	if path == "" {
		return fmt.Errorf("could not determine history path")
	}
	return os.Remove(path)
}
