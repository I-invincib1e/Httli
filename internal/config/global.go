package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type GlobalConfig struct {
	DefaultEnv string `json:"default_env"`
}

var Global GlobalConfig

// LoadGlobalConfig loads `~/.httli/config.json`. Soft failure (doesn't panic)
func LoadGlobalConfig() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	path := filepath.Join(home, ".httli", "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return err // it's clean to ignore
	}
	return json.Unmarshal(data, &Global)
}
