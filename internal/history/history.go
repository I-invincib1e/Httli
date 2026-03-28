package history

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/I-invincib1e/httli/internal/storage"
)

const maxHistory = 50

type Entry struct {
	Timestamp string `json:"timestamp"`
	Method    string `json:"method"`
	URL       string `json:"url"`
	Status    int    `json:"status"`
}

// historyPath returns the resolved path (project-local or global)
func historyPath() string {
	return storage.ResolvePath("history.json")
}

func loadHistory() ([]Entry, error) {
	data, err := os.ReadFile(historyPath())
	if err != nil {
		return nil, nil // no history yet
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

// Record appends a new entry to history (sliding window of maxHistory)
func Record(method, url string, status int) {
	entries, _ := loadHistory()
	entry := Entry{
		Timestamp: time.Now().Format(time.RFC3339),
		Method:    method,
		URL:       url,
		Status:    status,
	}
	entries = append(entries, entry)
	if len(entries) > maxHistory {
		entries = entries[len(entries)-maxHistory:]
	}
	_ = saveHistory(entries)
}

// List prints all history entries (supports --format json)
func List() {
	ListWithFormat("")
}

// ListWithFormat prints all history entries in the given format
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
		fmt.Printf("  [%d] %s %s %s → %d %s\n", i+1, statusIcon, e.Method, e.URL, e.Status, e.Timestamp)
	}
}

// Show prints details of a single history entry
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
	fmt.Printf("  Timestamp: %s\n", e.Timestamp)
	fmt.Printf("  Method:    %s\n", e.Method)
	fmt.Printf("  URL:       %s\n", e.URL)
	fmt.Printf("  Status:    %d\n", e.Status)
	return &e, nil
}

// Clear removes all history
func Clear() error {
	path := historyPath()
	if path == "" {
		return fmt.Errorf("could not determine history path")
	}
	return os.Remove(path)
}
