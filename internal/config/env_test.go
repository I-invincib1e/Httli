package config

import (
	"os"
	"testing"
)

// --- LoadEnv: doesn't overwrite existing vars ---

func TestLoadEnv_DoesNotOverwriteExisting(t *testing.T) {
	// Set a variable before loading
	os.Setenv("EXISTING_VAR", "original")
	defer os.Unsetenv("EXISTING_VAR")

	// Write a temp .env file that tries to override it
	f, err := os.CreateTemp("", ".env_test_*")
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString("EXISTING_VAR=overwritten\n")
	f.Close()
	defer os.Remove(f.Name())

	// Manually call loadFile with our temp file
	loadFile(f.Name())

	if got := os.Getenv("EXISTING_VAR"); got != "original" {
		t.Errorf("existing var was overwritten: expected 'original', got %q", got)
	}
}

func TestLoadEnv_SetsNewVar(t *testing.T) {
	os.Unsetenv("NEW_ENV_TEST_VAR_XYZ")
	defer os.Unsetenv("NEW_ENV_TEST_VAR_XYZ")

	f, err := os.CreateTemp("", ".env_test_*")
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString("NEW_ENV_TEST_VAR_XYZ=hello\n")
	f.Close()
	defer os.Remove(f.Name())

	loadFile(f.Name())

	if got := os.Getenv("NEW_ENV_TEST_VAR_XYZ"); got != "hello" {
		t.Errorf("expected 'hello', got %q", got)
	}
}

func TestLoadEnv_StripsCarriageReturn(t *testing.T) {
	os.Unsetenv("CRLF_TEST_VAR")
	defer os.Unsetenv("CRLF_TEST_VAR")

	f, err := os.CreateTemp("", ".env_test_*")
	if err != nil {
		t.Fatal(err)
	}
	// Windows-style CRLF
	f.WriteString("CRLF_TEST_VAR=value\r\n")
	f.Close()
	defer os.Remove(f.Name())

	loadFile(f.Name())

	got := os.Getenv("CRLF_TEST_VAR")
	if got != "value" {
		t.Errorf("expected 'value' without \\r, got %q", got)
	}
}

func TestLoadEnv_IgnoresComments(t *testing.T) {
	os.Unsetenv("COMMENT_VAR")
	defer os.Unsetenv("COMMENT_VAR")

	f, err := os.CreateTemp("", ".env_test_*")
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString("# this is a comment\nCOMMENT_VAR=set\n")
	f.Close()
	defer os.Remove(f.Name())

	loadFile(f.Name())

	if got := os.Getenv("COMMENT_VAR"); got != "set" {
		t.Errorf("expected 'set', got %q", got)
	}
}

// --- Interpolate ---

func TestInterpolate_Simple(t *testing.T) {
	os.Setenv("INTERP_TEST", "world")
	defer os.Unsetenv("INTERP_TEST")

	got, err := Interpolate("hello {{INTERP_TEST}}", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "hello world" {
		t.Errorf("expected 'hello world', got %q", got)
	}
}

func TestInterpolate_MultipleVars(t *testing.T) {
	os.Setenv("INTERP_A", "foo")
	os.Setenv("INTERP_B", "bar")
	defer os.Unsetenv("INTERP_A")
	defer os.Unsetenv("INTERP_B")

	got, err := Interpolate("{{INTERP_A}}-{{INTERP_B}}", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "foo-bar" {
		t.Errorf("expected 'foo-bar', got %q", got)
	}
}

func TestInterpolate_MissingStrict_ReturnsError(t *testing.T) {
	os.Unsetenv("NOT_SET_VAR_ZZZ")
	_, err := Interpolate("{{NOT_SET_VAR_ZZZ}}", false)
	if err == nil {
		t.Error("expected error for missing var in strict mode")
	}
}

func TestInterpolate_MissingIgnore_ReturnsPlaceholder(t *testing.T) {
	os.Unsetenv("NOT_SET_VAR_ZZZ")
	got, err := Interpolate("{{NOT_SET_VAR_ZZZ}}", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "{{NOT_SET_VAR_ZZZ}}" {
		t.Errorf("expected placeholder unchanged, got %q", got)
	}
}

func TestInterpolate_SpacesInBraces(t *testing.T) {
	os.Setenv("INTERP_SPACE", "trimmed")
	defer os.Unsetenv("INTERP_SPACE")

	got, err := Interpolate("{{ INTERP_SPACE }}", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "trimmed" {
		t.Errorf("expected 'trimmed', got %q", got)
	}
}

func TestInterpolate_NoPlaceholders(t *testing.T) {
	got, err := Interpolate("plain string", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "plain string" {
		t.Errorf("expected unchanged, got %q", got)
	}
}
