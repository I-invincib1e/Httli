package collections

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/I-invincib1e/http-cli/internal/config"
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

var storagePath string

func init() {
	home, err := os.UserHomeDir()
	if err == nil {
		storagePath = filepath.Join(home, ".httli", "collections.json")
	}
}

// InitStorage ensures the storage directory and file exist
func InitStorage() error {
	if storagePath == "" {
		return fmt.Errorf("could not determine home directory")
	}
	dir := filepath.Dir(storagePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	if _, err := os.Stat(storagePath); os.IsNotExist(err) {
		initial := Storage{Requests: make(map[string]RequestData)}
		return saveStorage(initial)
	}
	return nil
}

func loadStorage() (Storage, error) {
	var s Storage
	data, err := os.ReadFile(storagePath)
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
	return os.WriteFile(storagePath, data, 0644)
}

func normalizeName(name string) (string, error) {
	name = strings.TrimSpace(name)
	name = strings.ToLower(name)
	if strings.Contains(name, "//") || name == "" {
		return "", fmt.Errorf("invalid collection name format: '%s'", name)
	}
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

// ListCollections lists available saved requests
func ListCollections() {
	s, err := loadStorage()
	if err != nil || len(s.Requests) == 0 {
		fmt.Println("No collections found. Use 'http-cli collection save <name> [options]' to create one.")
		return
	}
	fmt.Println("Saved Requests:")
	for name, req := range s.Requests {
		fmt.Printf("  - %s [%s %s]\n", name, req.Method, req.URL)
	}
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

