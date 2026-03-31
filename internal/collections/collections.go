package collections

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/I-invincib1e/httli/internal/config"
	"github.com/I-invincib1e/httli/internal/storage"
)

type Storage struct {
	Requests map[string]RequestData `json:"requests"`
}

// RequestData now persists all request fields worth saving.
// All new fields use omitempty so existing collections.json files
// deserialize cleanly with zero-valued defaults.
type RequestData struct {
	Method          string            `json:"method"`
	URL             string            `json:"url"`
	Headers         map[string]string `json:"headers,omitempty"`
	Body            string            `json:"body,omitempty"`
	BearerToken     string            `json:"bearer_token,omitempty"`
	BasicAuth       string            `json:"basic_auth,omitempty"`
	TimeoutStr      string            `json:"timeout,omitempty"` // stored as "30s"
	FollowRedirects bool              `json:"follow_redirects,omitempty"`
	Retry           int               `json:"retry,omitempty"`
	RetryDelay      int               `json:"retry_delay,omitempty"`
	Description     string            `json:"description,omitempty"`
	CreatedAt       string            `json:"created_at,omitempty"`
	UpdatedAt       string            `json:"updated_at,omitempty"`
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
	if strings.Contains(name, "//") || name == "" {
		return "", fmt.Errorf("invalid collection name format: '%s'", name)
	}
	name = strings.TrimRight(name, "/")
	return name, nil
}

func exists(s Storage, name string) bool {
	_, ok := s.Requests[name]
	return ok
}

// configToRequestData converts a config.Config to a RequestData for storage.
func configToRequestData(cfg *config.Config) RequestData {
	rd := RequestData{
		Method:          cfg.Method,
		URL:             cfg.URL,
		Headers:         cfg.Headers,
		Body:            cfg.Body,
		BearerToken:     cfg.BearerToken,
		BasicAuth:       cfg.BasicAuth,
		FollowRedirects: cfg.FollowRedirects,
		Retry:           cfg.Retry,
		RetryDelay:      cfg.RetryDelay,
	}
	if cfg.Timeout > 0 {
		rd.TimeoutStr = cfg.Timeout.String()
	}
	return rd
}

// requestDataToConfig hydrates a config.Config from stored RequestData.
func requestDataToConfig(req RequestData) *config.Config {
	cfg := &config.Config{
		Method:          req.Method,
		URL:             req.URL,
		Headers:         make(map[string]string),
		Body:            req.Body,
		BearerToken:     req.BearerToken,
		BasicAuth:       req.BasicAuth,
		FollowRedirects: req.FollowRedirects,
		Retry:           req.Retry,
		RetryDelay:      req.RetryDelay,
		Timeout:         30 * time.Second, // safe default
	}
	// Parse stored timeout string (e.g. "10s", "1m")
	if req.TimeoutStr != "" {
		if d, err := time.ParseDuration(req.TimeoutStr); err == nil {
			cfg.Timeout = d
		}
	}
	for k, v := range req.Headers {
		cfg.Headers[k] = v
	}
	return cfg
}

// SaveRequest saves a parsed config into the collection, failing if it already exists.
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

	rd := configToRequestData(cfg)
	rd.CreatedAt = time.Now().Format(time.RFC3339)
	rd.UpdatedAt = rd.CreatedAt
	s.Requests[name] = rd
	return saveStorage(s)
}

// UpdateRequest updates an existing request in the collection.
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

	rd := configToRequestData(cfg)
	rd.CreatedAt = s.Requests[name].CreatedAt // preserve original creation time
	rd.Description = s.Requests[name].Description // preserve description
	rd.UpdatedAt = time.Now().Format(time.RFC3339)
	s.Requests[name] = rd
	return saveStorage(s)
}

// DescribeRequest sets or updates the description for a saved request.
func DescribeRequest(rawName, description string) error {
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

	req, ok := s.Requests[name]
	if !ok {
		return fmt.Errorf("request '%s' not found", name)
	}
	req.Description = description
	req.UpdatedAt = time.Now().Format(time.RFC3339)
	s.Requests[name] = req
	return saveStorage(s)
}

// DeleteRequest removes a request from the collection.
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

// GetRequest retrieves a saved request and builds a Config from it.
// Does NOT call InterpolateAll — callers must do that explicitly.
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

	return requestDataToConfig(req), nil
}

// ListCollections lists available saved requests, grouped by namespace.
func ListCollections() {
	s, err := loadStorage()
	if err != nil || len(s.Requests) == 0 {
		fmt.Println("No collections found. Use 'httli collection save <name> [options]' to create one.")
		return
	}

	var names []string
	for name := range s.Requests {
		names = append(names, name)
	}
	sort.Strings(names)

	groups := make(map[string][]string)
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

	for _, name := range ungrouped {
		req := s.Requests[name]
		desc := ""
		if req.Description != "" {
			desc = "  # " + req.Description
		}
		fmt.Printf("  - %s [%s %s]%s\n", name, req.Method, req.URL, desc)
	}

	var prefixes []string
	for p := range groups {
		prefixes = append(prefixes, p)
	}
	sort.Strings(prefixes)

	for _, prefix := range prefixes {
		fmt.Printf("\n  %s/\n", prefix)
		for _, name := range groups[prefix] {
			req := s.Requests[name]
			shortName := name[len(prefix)+1:]
			desc := ""
			if req.Description != "" {
				desc = "  # " + req.Description
			}
			fmt.Printf("    - %s [%s %s]%s\n", shortName, req.Method, req.URL, desc)
		}
	}

	if storage.IsProjectLocal() {
		fmt.Printf("\n  (project-local: %s)\n", storagePath())
	}
}

// ListAllNames returns all collection names (for run-all pattern matching).
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

// ExportCollections writes all collections to a file.
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
// mode:
//   "merge"     — add new entries, skip existing (default)
//   "overwrite" — replace existing entries
//   "skip"      — skip all entries that conflict
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
			case "skip":
				// Explicitly skip — do nothing
				skipped++
			default: // "merge" — add new, skip conflicts
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
