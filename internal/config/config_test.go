package config

import (
	"os"
	"testing"
	"time"
)

// --- ParseHeaders ---

func TestParseHeaders_Basic(t *testing.T) {
	h := ParseHeaders("Content-Type:application/json,Authorization:Bearer tok")
	if h["Content-Type"] != "application/json" {
		t.Errorf("expected 'application/json', got %q", h["Content-Type"])
	}
	if h["Authorization"] != "Bearer tok" {
		t.Errorf("expected 'Bearer tok', got %q", h["Authorization"])
	}
}

func TestParseHeaders_Empty(t *testing.T) {
	h := ParseHeaders("")
	if len(h) != 0 {
		t.Errorf("expected empty headers, got %v", h)
	}
}

func TestParseHeaders_MissingValue(t *testing.T) {
	h := ParseHeaders("Key:")
	if h["Key"] != "" {
		t.Errorf("expected empty value, got %q", h["Key"])
	}
}

// --- InterpolateAll ---

func TestInterpolateAll_HappyPath(t *testing.T) {
	os.Setenv("TEST_URL", "https://example.com")
	defer os.Unsetenv("TEST_URL")

	cfg := &Config{
		URL:              "{{TEST_URL}}/api",
		IgnoreMissingEnv: false,
	}
	if err := cfg.InterpolateAll(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.URL != "https://example.com/api" {
		t.Errorf("expected interpolated URL, got %q", cfg.URL)
	}
}

func TestInterpolateAll_MissingVar_Strict(t *testing.T) {
	os.Unsetenv("MISSING_VAR_XYZ")
	cfg := &Config{
		URL:              "{{MISSING_VAR_XYZ}}",
		IgnoreMissingEnv: false,
	}
	err := cfg.InterpolateAll()
	if err == nil {
		t.Error("expected error for missing var in strict mode")
	}
}

func TestInterpolateAll_MissingVar_IgnoreMissing(t *testing.T) {
	os.Unsetenv("MISSING_VAR_XYZ")
	cfg := &Config{
		URL:              "{{MISSING_VAR_XYZ}}",
		IgnoreMissingEnv: true,
	}
	if err := cfg.InterpolateAll(); err != nil {
		t.Fatalf("unexpected error in ignore-missing mode: %v", err)
	}
	// Should leave placeholder unchanged
	if cfg.URL != "{{MISSING_VAR_XYZ}}" {
		t.Errorf("expected placeholder to remain, got %q", cfg.URL)
	}
}

func TestInterpolateAll_Headers(t *testing.T) {
	os.Setenv("MY_TOKEN", "secret")
	defer os.Unsetenv("MY_TOKEN")

	cfg := &Config{
		URL: "https://example.com",
		Headers: map[string]string{
			"Authorization": "Bearer {{MY_TOKEN}}",
		},
	}
	if err := cfg.InterpolateAll(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Headers["Authorization"] != "Bearer secret" {
		t.Errorf("expected 'Bearer secret', got %q", cfg.Headers["Authorization"])
	}
}

// --- ApplyOverrides ---

func TestApplyOverrides_TimeoutDefault_NotOverridden(t *testing.T) {
	base := &Config{Timeout: 10 * time.Second}
	run := &Config{Timeout: 30 * time.Second} // default — should not override
	base.ApplyOverrides(run)
	if base.Timeout != 10*time.Second {
		t.Errorf("expected base timeout 10s preserved, got %v", base.Timeout)
	}
}

func TestApplyOverrides_TimeoutCustom_Overrides(t *testing.T) {
	base := &Config{Timeout: 10 * time.Second}
	run := &Config{Timeout: 5 * time.Second}
	base.ApplyOverrides(run)
	if base.Timeout != 5*time.Second {
		t.Errorf("expected timeout overridden to 5s, got %v", base.Timeout)
	}
}

func TestApplyOverrides_Flags(t *testing.T) {
	base := &Config{}
	run := &Config{
		Verbose:    true,
		Quiet:      true,
		StatusOnly: true,
		Fail:       true,
		FailFast:   true,
		RetryBackoff: true,
		Retry:      3,
		RetryDelay: 5,
	}
	base.ApplyOverrides(run)
	if !base.Verbose || !base.Quiet || !base.StatusOnly {
		t.Error("bool flags not applied by ApplyOverrides")
	}
	if !base.Fail || !base.FailFast || !base.RetryBackoff {
		t.Error("fail/failfast/backoff flags not applied by ApplyOverrides")
	}
	if base.Retry != 3 || base.RetryDelay != 5 {
		t.Error("int fields not applied by ApplyOverrides")
	}
}

// --- ParseFlags ---

func TestParseFlags_Defaults(t *testing.T) {
	cfg, err := ParseFlags([]string{"-u", "https://example.com"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Method != "GET" {
		t.Errorf("expected default method GET, got %q", cfg.Method)
	}
	if cfg.Timeout != 30*time.Second {
		t.Errorf("expected default timeout 30s, got %v", cfg.Timeout)
	}
	if cfg.URL != "https://example.com" {
		t.Errorf("expected URL, got %q", cfg.URL)
	}
}

func TestParseFlags_FailFastRegistered(t *testing.T) {
	// This would have caused os.Exit with the old ExitOnError
	cfg, err := ParseFlags([]string{"--fail-fast"})
	if err != nil {
		t.Fatalf("--fail-fast caused error: %v", err)
	}
	if !cfg.FailFast {
		t.Error("expected FailFast=true")
	}
}

func TestParseFlags_RetryBackoffRegistered(t *testing.T) {
	cfg, err := ParseFlags([]string{"--retry-backoff", "--retry", "3"})
	if err != nil {
		t.Fatalf("--retry-backoff caused error: %v", err)
	}
	if !cfg.RetryBackoff {
		t.Error("expected RetryBackoff=true")
	}
	if cfg.Retry != 3 {
		t.Errorf("expected Retry=3, got %d", cfg.Retry)
	}
}

func TestParseFlags_MethodUppercased(t *testing.T) {
	cfg, err := ParseFlags([]string{"-m", "post", "-u", "https://example.com"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Method != "POST" {
		t.Errorf("expected method uppercased to POST, got %q", cfg.Method)
	}
}
