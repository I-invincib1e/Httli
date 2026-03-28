package client

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

// ExecuteRequest executes an HTTP request based on the provided config
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

	var lastErr error

	retries := cfg.Retry
	if retries < 0 {
		retries = 0
	}
	delay := time.Duration(cfg.RetryDelay) * time.Second

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

		if err == nil && resp.StatusCode < 500 {
			defer resp.Body.Close()
			bodyBytes, errRead := io.ReadAll(resp.Body)
			if errRead != nil {
				return nil, fmt.Errorf("error reading response body: %w", errRead)
			}
			return &Response{
				StatusCode: resp.StatusCode,
				Status:     resp.Status,
				Headers:    resp.Header,
				Body:       bodyBytes,
				Duration:   duration,
			}, nil
		}

		if err != nil {
			lastErr = err
		} else {
			lastErr = fmt.Errorf("server error: %d %s", resp.StatusCode, resp.Status)
			resp.Body.Close()
		}

		if i < retries {
			time.Sleep(delay)
		}
	}

	return nil, fmt.Errorf("request failed: %v", lastErr)
}
