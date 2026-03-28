package collections

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/I-invincib1e/httli/internal/config"
	"github.com/I-invincib1e/httli/internal/storage"
)

type Storage struct {
	Requests map[string]RequestData `json:"requests"`
}

type RequestData struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    string            `json:"body,omitempty"`
}

// storagePath returns the resolved path (project-local or global)
func storagePath() string {
	return storage.ResolvePath("collections.json")
}

// InitStorage ensures the storage directory and file exist
func InitStorage() error {
	path := storagePath()
	if err := storage.EnsureDir(path); err != nil {
		return err
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		initial := Storage{Requests: make(map[string]RequestData)}
		return saveStorage(initial)
	}
	return nil
}

func loadStorage() (Storage, error) {
	var s Storage
	data, err := os.ReadFile(storagePath())
	if err != nil {
		return s, err
	}
	err = json.Unmarshal(data, &s)
	if s.Requests == nil {
		s.Requests = make(map[string]RequestData)
	}
	return s, err
}

func saveStorage(s Storage) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(storagePath(), data, 0644)
}

func normalizeName(name string) (string, error) {
	name = strings.TrimSpace(name)
	name = strings.ToLower(name)
	// Allow single / for namespacing (auth/login), reject // and empty
	if strings.Contains(name, "//") || name == "" {
		return "", fmt.Errorf("invalid collection name format: '%s'", name)
	}
	// Trim trailing slash
	name = strings.TrimRight(name, "/")
	return name, nil
}

func exists(s Storage, name string) bool {
	_, ok := s.Requests[name]
	return ok
}

// SaveRequest saves a parsed config into the collection, failing if it already exists
func SaveRequest(rawName string, cfg *config.Config) error {
	name, err := normalizeName(rawName)
	if err != nil {
		return err
	}
	if err := InitStorage(); err != nil {
		return err
	}
	s, err := loadStorage()
	if err != nil {
		return err
	}

	if exists(s, name) {
		return fmt.Errorf("request '%s' already exists, use 'update' instead", name)
	}

	s.Requests[name] = RequestData{
		Method:  cfg.Method,
		URL:     cfg.URL,
		Headers: cfg.Headers,
		Body:    cfg.Body,
	}
	return saveStorage(s)
}

// UpdateRequest updates an existing request in the collection
func UpdateRequest(rawName string, cfg *config.Config) error {
	name, err := normalizeName(rawName)
	if err != nil {
		return err
	}
	if err := InitStorage(); err != nil {
		return err
	}
	s, err := loadStorage()
	if err != nil {
		return err
	}

	if !exists(s, name) {
		return fmt.Errorf("request '%s' not found, use 'save' instead", name)
	}

	s.Requests[name] = RequestData{
		Method:  cfg.Method,
		URL:     cfg.URL,
		Headers: cfg.Headers,
		Body:    cfg.Body,
	}
	return saveStorage(s)
}

// DeleteRequest removes a request from the collection
func DeleteRequest(rawName string) error {
	name, err := normalizeName(rawName)
	if err != nil {
		return err
	}
	if err := InitStorage(); err != nil {
		return err
	}
	s, err := loadStorage()
	if err != nil {
		return err
	}

	if !exists(s, name) {
		return fmt.Errorf("request '%s' not found", name)
	}

	delete(s.Requests, name)
	return saveStorage(s)
}

// GetRequest retrieves a saved request without interpolating fields
func GetRequest(rawName string) (*config.Config, error) {
	name, err := normalizeName(rawName)
	if err != nil {
		return nil, err
	}
	s, err := loadStorage()
	if err != nil {
		return nil, err
	}
	
	req, ok := s.Requests[name]
	if !ok {
		return nil, fmt.Errorf("request '%s' not found in collections", name)
	}

	cfg := &config.Config{
		Method:  req.Method,
		URL:     req.URL,
		Headers: make(map[string]string),
		Body:    req.Body,
		Timeout: 30, // Default timeout
	}
	for k, v := range req.Headers {
		cfg.Headers[k] = v
	}

	return cfg, nil
}

// ListCollections lists available saved requests, grouped by namespace
func ListCollections() {
	s, err := loadStorage()
	if err != nil || len(s.Requests) == 0 {
		fmt.Println("No collections found. Use 'httli collection save <name> [options]' to create one.")
		return
	}

	// Collect and sort names
	var names []string
	for name := range s.Requests {
		names = append(names, name)
	}
	sort.Strings(names)

	// Group by namespace prefix
	groups := make(map[string][]string) // prefix → list of full names
	var ungrouped []string

	for _, name := range names {
		if idx := strings.Index(name, "/"); idx > 0 {
			prefix := name[:idx]
			groups[prefix] = append(groups[prefix], name)
		} else {
			ungrouped = append(ungrouped, name)
		}
	}

	fmt.Println("Saved Requests:")

	// Print ungrouped first
	for _, name := range ungrouped {
		req := s.Requests[name]
		fmt.Printf("  - %s [%s %s]\n", name, req.Method, req.URL)
	}

	// Print grouped by namespace
	var prefixes []string
	for p := range groups {
		prefixes = append(prefixes, p)
	}
	sort.Strings(prefixes)

	for _, prefix := range prefixes {
		fmt.Printf("\n  %s/\n", prefix)
		for _, name := range groups[prefix] {
			req := s.Requests[name]
			shortName := name[len(prefix)+1:] // strip prefix/
			fmt.Printf("    - %s [%s %s]\n", shortName, req.Method, req.URL)
		}
	}

	if storage.IsProjectLocal() {
		fmt.Printf("\n  (project-local: %s)\n", storagePath())
	}
}

// ListAllNames returns all collection names (for run-all pattern matching)
func ListAllNames() ([]string, error) {
	s, err := loadStorage()
	if err != nil {
		return nil, err
	}
	var names []string
	for name := range s.Requests {
		names = append(names, name)
	}
	sort.Strings(names)
	return names, nil
}

// ExportCollections writes all collections to a file
func ExportCollections(path string) error {
	if err := InitStorage(); err != nil {
		return err
	}
	s, err := loadStorage()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// ImportCollections reads a JSON file and merges into current storage.
// mode: "merge" (default, skip existing), "overwrite" (replace existing), "skip" (skip all existing)
func ImportCollections(path string, mode string) error {
	if err := InitStorage(); err != nil {
		return err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("cannot read file: %w", err)
	}

	var incoming Storage
	if err := json.Unmarshal(data, &incoming); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	s, err := loadStorage()
	if err != nil {
		return err
	}

	added, skipped, overwritten := 0, 0, 0
	for name, req := range incoming.Requests {
		if exists(s, name) {
			switch mode {
			case "overwrite":
				s.Requests[name] = req
				overwritten++
			default: // "merge" and "skip" both skip existing
				skipped++
			}
		} else {
			s.Requests[name] = req
			added++
		}
	}

	if err := saveStorage(s); err != nil {
		return err
	}

	fmt.Printf("Import complete: %d added, %d skipped, %d overwritten\n", added, skipped, overwritten)
	return nil
}
