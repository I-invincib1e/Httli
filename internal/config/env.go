package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// interpolateRe is compiled once at package level (not per-call)
var interpolateRe = regexp.MustCompile(`\{\{\s*([A-Za-z0-9_]+)\s*\}\}`)

// LoadEnv reads .env, .env.local, and .env.<envName> in order.
// Existing environment variables are NOT overwritten (safe for CI/CD).
func LoadEnv(envName string) {
	// Base
	loadFile(".env")
	// Local override
	loadFile(".env.local")
	// Specific environment override
	if envName != "" {
		loadFile(".env." + envName)
	}
}

func loadFile(filename string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Strip trailing \r (Windows line endings)
		line = strings.TrimRight(line, "\r")
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			val = strings.Trim(val, `"'`)
			// Only set if not already defined in the environment
			if _, exists := os.LookupEnv(key); !exists {
				os.Setenv(key, val)
			}
		}
	}
}

// Interpolate replaces {{VAR}} patterns with env vars.
// If ignoreMissing is false, it returns an error when a variable is not found.
func Interpolate(s string, ignoreMissing bool) (string, error) {
	var extractErr error
	res := interpolateRe.ReplaceAllStringFunc(s, func(m string) string {
		match := interpolateRe.FindStringSubmatch(m)
		if len(match) > 1 {
			if val, exists := os.LookupEnv(match[1]); exists {
				return val
			} else if !ignoreMissing {
				extractErr = fmt.Errorf("environment variable '%s' not found", match[1])
			}
		}
		return m
	})
	return res, extractErr
}
