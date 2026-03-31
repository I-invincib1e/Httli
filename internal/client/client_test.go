package client

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/I-invincib1e/httli/internal/config"
)

func makeConfig(url string) *config.Config {
	return &config.Config{
		Method:  "GET",
		URL:     url,
		Headers: make(map[string]string),
		Timeout: 5 * time.Second,
	}
}

// --- Basic success ---

func TestExecuteRequest_200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	resp, err := ExecuteRequest(makeConfig(srv.URL))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	if string(resp.Body) != `{"ok":true}` {
		t.Errorf("unexpected body: %q", resp.Body)
	}
}

// --- 4xx returns response (not error) ---

func TestExecuteRequest_404_ReturnsResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte(`not found`))
	}))
	defer srv.Close()

	resp, err := ExecuteRequest(makeConfig(srv.URL))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 404 {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

// --- 5xx body is captured ---

func TestExecuteRequest_5xx_BodyCaptured(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(503)
		w.Write([]byte(`service unavailable body`))
	}))
	defer srv.Close()

	cfg := makeConfig(srv.URL)
	cfg.Retry = 0 // no retries

	resp, err := ExecuteRequest(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 503 {
		t.Errorf("expected 503, got %d", resp.StatusCode)
	}
	if string(resp.Body) != "service unavailable body" {
		t.Errorf("expected 5xx body to be captured, got %q", resp.Body)
	}
}

// --- Retry: retries on 5xx and returns last response ---

func TestExecuteRequest_Retry_5xx(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(500)
		w.Write([]byte(`error`))
	}))
	defer srv.Close()

	cfg := makeConfig(srv.URL)
	cfg.Retry = 2
	cfg.RetryDelay = 0 // instant for tests

	resp, err := ExecuteRequest(cfg)
	if err != nil {
		t.Fatalf("unexpected error after retries: %v", err)
	}
	if callCount != 3 { // initial + 2 retries
		t.Errorf("expected 3 total calls, got %d", callCount)
	}
	if resp.StatusCode != 500 {
		t.Errorf("expected final 500 response, got %d", resp.StatusCode)
	}
}

// --- Retry: succeeds on second attempt ---

func TestExecuteRequest_Retry_SucceedsOnSecondAttempt(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			w.WriteHeader(503)
			return
		}
		w.WriteHeader(200)
		w.Write([]byte(`ok`))
	}))
	defer srv.Close()

	cfg := makeConfig(srv.URL)
	cfg.Retry = 2
	cfg.RetryDelay = 0

	resp, err := ExecuteRequest(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200 on second attempt, got %d", resp.StatusCode)
	}
	if callCount != 2 {
		t.Errorf("expected 2 calls, got %d", callCount)
	}
}

// --- calcDelay ---

func TestCalcDelay_NoBackoff_Fixed(t *testing.T) {
	base := 2 * time.Second
	if d := calcDelay(base, 0, false); d != base {
		t.Errorf("expected fixed delay %v, got %v", base, d)
	}
	if d := calcDelay(base, 3, false); d != base {
		t.Errorf("expected fixed delay %v, got %v", base, d)
	}
}

func TestCalcDelay_Backoff_Doubles(t *testing.T) {
	base := 1 * time.Second
	if d := calcDelay(base, 0, true); d != base {
		t.Errorf("attempt 0: expected %v, got %v", base, d)
	}
	if d := calcDelay(base, 1, true); d != 2*time.Second {
		t.Errorf("attempt 1: expected 2s, got %v", d)
	}
	if d := calcDelay(base, 2, true); d != 4*time.Second {
		t.Errorf("attempt 2: expected 4s, got %v", d)
	}
}

func TestCalcDelay_Backoff_CappedAt30s(t *testing.T) {
	base := 10 * time.Second
	d := calcDelay(base, 10, true) // would be 10240s without cap
	if d != maxBackoffDelay {
		t.Errorf("expected cap at %v, got %v", maxBackoffDelay, d)
	}
}

// --- Invalid JSON body ---

func TestExecuteRequest_InvalidJSON_ReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer srv.Close()

	cfg := makeConfig(srv.URL)
	cfg.Method = "POST"
	cfg.Body = `{invalid json`

	_, err := ExecuteRequest(cfg)
	if err == nil {
		t.Error("expected error for invalid JSON body")
	}
}
