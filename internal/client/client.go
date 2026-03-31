package client

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/I-invincib1e/httli/internal/config"
)

// Response holds the HTTP response data
type Response struct {
	StatusCode int
	Status     string
	Headers    http.Header
	Body       []byte
	Duration   time.Duration
}

// addAuth adds authentication headers to the config
func addAuth(cfg *config.Config) error {
	if cfg.Headers == nil {
		cfg.Headers = make(map[string]string)
	}

	if cfg.BearerToken != "" {
		cfg.Headers["Authorization"] = "Bearer " + cfg.BearerToken
	} else if cfg.BasicAuth != "" {
		parts := strings.SplitN(cfg.BasicAuth, ":", 2)
		if len(parts) != 2 {
			return fmt.Errorf("basic auth format should be 'user:pass'")
		}
		auth := base64.StdEncoding.EncodeToString([]byte(cfg.BasicAuth))
		cfg.Headers["Authorization"] = "Basic " + auth
	}
	return nil
}

// maxBackoffDelay caps exponential backoff at 30 seconds
const maxBackoffDelay = 30 * time.Second

// ExecuteRequest executes an HTTP request based on the provided config.
// On 5xx responses, the body is always read. After exhausting retries on 5xx,
// the last response is returned (not an error), so callers can inspect the body.
func ExecuteRequest(cfg *config.Config) (*Response, error) {
	if err := addAuth(cfg); err != nil {
		return nil, err
	}

	httpClient := &http.Client{Timeout: cfg.Timeout}
	if !cfg.FollowRedirects {
		httpClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	// JSON validation
	isJsonContent := false
	for k, v := range cfg.Headers {
		if strings.ToLower(k) == "content-type" && strings.Contains(strings.ToLower(v), "application/json") {
			isJsonContent = true
			break
		}
	}

	trimmedBody := strings.TrimSpace(cfg.Body)
	if trimmedBody != "" && (isJsonContent || strings.HasPrefix(trimmedBody, "{") || strings.HasPrefix(trimmedBody, "[")) {
		if !json.Valid([]byte(trimmedBody)) {
			return nil, fmt.Errorf("invalid JSON body provided")
		}
	}

	retries := cfg.Retry
	if retries < 0 {
		retries = 0
	}
	baseDelay := time.Duration(cfg.RetryDelay) * time.Second

	var lastResp *Response
	var lastErr error

	for i := 0; i <= retries; i++ {
		var bodyReader io.Reader
		if cfg.Body != "" {
			bodyReader = strings.NewReader(cfg.Body)
		}

		req, err := http.NewRequest(cfg.Method, cfg.URL, bodyReader)
		if err != nil {
			return nil, fmt.Errorf("error creating request: %w", err)
		}

		for k, v := range cfg.Headers {
			req.Header.Set(k, v)
		}

		if cfg.Body != "" && req.Header.Get("Content-Type") == "" {
			req.Header.Set("Content-Type", "application/json")
		}

		start := time.Now()
		resp, err := httpClient.Do(req)
		duration := time.Since(start)

		if err != nil {
			// Network error — retry if we have attempts left
			lastErr = err
			if i < retries {
				delay := calcDelay(baseDelay, i, cfg.RetryBackoff)
				if cfg.Verbose {
					fmt.Fprintf(os.Stderr, "  ↻ Retry %d/%d in %s (network error: %v)\n", i+1, retries, delay, err)
				}
				time.Sleep(delay)
			}
			continue
		}

		// Always read the body (even on 5xx) — this was the critical fix
		bodyBytes, errRead := io.ReadAll(resp.Body)
		resp.Body.Close() // immediate close, not deferred in loop

		if errRead != nil {
			return nil, fmt.Errorf("error reading response body: %w", errRead)
		}

		thisResp := &Response{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
			Headers:    resp.Header,
			Body:       bodyBytes,
			Duration:   duration,
		}

		// Non-5xx: return immediately (success or 4xx)
		if resp.StatusCode < 500 {
			return thisResp, nil
		}

		// 5xx: save as lastResp and retry if attempts remain
		lastResp = thisResp
		lastErr = nil // we have a real response, not a network error

		if i < retries {
			delay := calcDelay(baseDelay, i, cfg.RetryBackoff)
			if cfg.Verbose {
				fmt.Fprintf(os.Stderr, "  ↻ Retry %d/%d in %s (HTTP %d)\n", i+1, retries, delay, resp.StatusCode)
			}
			time.Sleep(delay)
		}
	}

	// Exhausted retries
	if lastResp != nil {
		// Return the actual 5xx response so callers can inspect body/status
		return lastResp, nil
	}

	// Pure network failure (no response ever received)
	return nil, fmt.Errorf("request failed after %d retries: %v", retries, lastErr)
}

// calcDelay calculates the delay for a given retry attempt.
// With backoff enabled: delay * 2^attempt, capped at maxBackoffDelay.
// Without backoff: fixed delay.
func calcDelay(baseDelay time.Duration, attempt int, backoff bool) time.Duration {
	if !backoff {
		return baseDelay
	}
	delay := baseDelay
	for i := 0; i < attempt; i++ {
		delay *= 2
		if delay > maxBackoffDelay {
			return maxBackoffDelay
		}
	}
	return delay
}
