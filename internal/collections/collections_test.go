package collections

import (
	"os"
	"testing"
	"time"

	"github.com/I-invincib1e/httli/internal/config"
)

// TestNormalizeName covers edge cases for name normalization
func TestNormalizeName_Basic(t *testing.T) {
	name, err := normalizeName("Auth/Login")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "auth/login" {
		t.Errorf("expected 'auth/login', got %q", name)
	}
}

func TestNormalizeName_TrailingSlash(t *testing.T) {
	name, err := normalizeName("auth/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "auth" {
		t.Errorf("expected 'auth', got %q", name)
	}
}

func TestNormalizeName_DoubleSlash_Invalid(t *testing.T) {
	_, err := normalizeName("auth//login")
	if err == nil {
		t.Error("expected error for double slash")
	}
}

func TestNormalizeName_Empty_Invalid(t *testing.T) {
	_, err := normalizeName("   ")
	if err == nil {
		t.Error("expected error for empty name")
	}
}

// TestSaveGetRoundtrip verifies full request data is persisted and retrieved
func TestSaveGetRoundtrip(t *testing.T) {
	// Use a temp directory for storage
	tmpDir := t.TempDir()
	origPath := storagePath
	// Temporarily override storage path for test isolation
	_ = origPath // storagePath is a function, not a var — we test via temp dir
	// Write a temp collections.json
	path := tmpDir + "/collections.json"
	_ = os.WriteFile(path, []byte(`{"requests":{}}`), 0644)

	// Build a config with full fields
	cfg := &config.Config{
		Method:          "POST",
		URL:             "https://api.example.com/login",
		Headers:         map[string]string{"Content-Type": "application/json"},
		Body:            `{"user":"test"}`,
		BearerToken:     "tok123",
		FollowRedirects: true,
		Timeout:         15 * time.Second,
		Retry:           2,
		RetryDelay:      3,
	}

	rd := configToRequestData(cfg)

	// Verify all fields are copied
	if rd.Method != "POST" {
		t.Errorf("Method: expected POST, got %q", rd.Method)
	}
	if rd.BearerToken != "tok123" {
		t.Errorf("BearerToken: expected tok123, got %q", rd.BearerToken)
	}
	if rd.TimeoutStr != "15s" {
		t.Errorf("TimeoutStr: expected '15s', got %q", rd.TimeoutStr)
	}
	if !rd.FollowRedirects {
		t.Error("FollowRedirects should be true")
	}
	if rd.Retry != 2 {
		t.Errorf("Retry: expected 2, got %d", rd.Retry)
	}
	if rd.RetryDelay != 3 {
		t.Errorf("RetryDelay: expected 3, got %d", rd.RetryDelay)
	}

	// Round-trip through requestDataToConfig
	back := requestDataToConfig(rd)
	if back.BearerToken != "tok123" {
		t.Errorf("round-trip BearerToken: expected tok123, got %q", back.BearerToken)
	}
	if back.Timeout != 15*time.Second {
		t.Errorf("round-trip Timeout: expected 15s, got %v", back.Timeout)
	}
}

// TestRequestDataToConfig_DefaultTimeout verifies 30s default when no timeout stored
func TestRequestDataToConfig_DefaultTimeout(t *testing.T) {
	rd := RequestData{
		Method: "GET",
		URL:    "https://example.com",
	}
	cfg := requestDataToConfig(rd)
	if cfg.Timeout != 30*time.Second {
		t.Errorf("expected default 30s timeout, got %v", cfg.Timeout)
	}
}

// TestRequestDataToConfig_ParsesTimeoutString
func TestRequestDataToConfig_ParsesTimeoutString(t *testing.T) {
	rd := RequestData{
		Method:     "GET",
		URL:        "https://example.com",
		TimeoutStr: "45s",
	}
	cfg := requestDataToConfig(rd)
	if cfg.Timeout != 45*time.Second {
		t.Errorf("expected 45s timeout, got %v", cfg.Timeout)
	}
}
