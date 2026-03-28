package storage

import (
	"os"
	"path/filepath"
	"sync"
)

// resolved caches the base directory so we only resolve once per process
var (
	resolvedBase string
	resolveOnce  sync.Once
)

// baseDir determines the storage root: ./.httli/ (project-local) or ~/.httli/ (global)
func baseDir() string {
	resolveOnce.Do(func() {
		// 1. Check CWD for a .httli/ directory
		localDir := filepath.Join(".", ".httli")
		if info, err := os.Stat(localDir); err == nil && info.IsDir() {
			abs, err := filepath.Abs(localDir)
			if err == nil {
				resolvedBase = abs
				return
			}
		}

		// 2. Fall back to ~/.httli/
		home, err := os.UserHomeDir()
		if err != nil {
			resolvedBase = filepath.Join(".", ".httli") // worst case
			return
		}
		resolvedBase = filepath.Join(home, ".httli")
	})
	return resolvedBase
}

// ResolvePath returns the full path for a given filename within the active storage root.
// It checks for a project-local .httli/ first, then falls back to ~/.httli/.
func ResolvePath(filename string) string {
	return filepath.Join(baseDir(), filename)
}

// EnsureDir creates the parent directory for a given path if it doesn't exist.
func EnsureDir(path string) error {
	return os.MkdirAll(filepath.Dir(path), 0755)
}

// IsProjectLocal returns true if storage is using a project-local .httli/ directory.
func IsProjectLocal() bool {
	base := baseDir()
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	return base != filepath.Join(home, ".httli")
}
