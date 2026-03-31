package output

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/I-invincib1e/httli/internal/client"
	"github.com/I-invincib1e/httli/internal/config"
	"github.com/I-invincib1e/httli/internal/styles"
	"github.com/charmbracelet/lipgloss"
)

// JSONResponse is the structured output for --format json
type JSONResponse struct {
	OK         bool                `json:"ok"`
	Status     int                 `json:"status"`
	StatusText string              `json:"status_text"`
	DurationMs int64               `json:"duration_ms"`
	Headers    map[string][]string `json:"headers"`
	Body       interface{}         `json:"body"`
}

// maskHeaderValue masks sensitive auth headers unless in verbose mode
func maskHeaderValue(k, v string, verbose bool) string {
	if verbose || (k != "Authorization" && strings.ToLower(k) != "authorization") {
		return v
	}
	if strings.HasPrefix(v, "Bearer ") || strings.HasPrefix(v, "Basic ") {
		return strings.SplitN(v, " ", 2)[0] + " ***"
	}
	return v
}

// DisplayRequest displays the request information
func DisplayRequest(cfg *config.Config, st *styles.Styles) {
	if cfg.Silent {
		return
	}
	// Suppress request display for JSON/raw/extract/quiet/status-only modes
	if cfg.Quiet || cfg.StatusOnly || cfg.Format == "json" || cfg.Raw || cfg.Extract != "" {
		return
	}

	fmt.Println()
	fmt.Println(st.Header.Render("REQUEST"))
	fmt.Println()
	fmt.Printf("%s %s\n", st.Method.Render(cfg.Method), st.URL.Render(cfg.URL))
	fmt.Println()

	if len(cfg.Headers) > 0 {
		fmt.Println(st.Key.Render("Headers:"))
		for k, v := range cfg.Headers {
			fmt.Printf("  %s: %s\n", st.Key.Render(k), st.Value.Render(maskHeaderValue(k, v, cfg.Verbose)))
		}
		fmt.Println()
	}

	if cfg.Body != "" && cfg.Verbose {
		fmt.Println(st.Key.Render("Body:"))
		prettyBody := FormatJSON(cfg.Body)
		fmt.Println(st.Body.Render(prettyBody))
		fmt.Println()
	}
}

// renderStatus renders status code with appropriate color
func renderStatus(statusCode int, statusText string) string {
	color := styles.StatusColor(statusCode)
	if color == "" {
		return statusText // non-TTY: plain text
	}
	statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Bold(true)
	return statusStyle.Render(statusText)
}

// DisplayResponse decides which output mode to use based on config flags.
// --output is orthogonal: it always saves to file regardless of display mode.
func DisplayResponse(cfg *config.Config, resp *client.Response, st *styles.Styles) error {
	if cfg.Silent {
		return nil
	}

	// --output: always save body to file, independent of display mode
	if cfg.OutputFile != "" {
		if err := os.WriteFile(cfg.OutputFile, resp.Body, 0644); err != nil {
			return fmt.Errorf("error saving file: %w", err)
		}
		if !cfg.Quiet && !cfg.StatusOnly && !cfg.Raw && cfg.Format != "json" && cfg.Extract == "" {
			fmt.Printf("%s %s\n", st.Success.Render("Response saved to:"), cfg.OutputFile)
		}
	}

	// --extract: print just the extracted value
	if cfg.Extract != "" {
		return displayExtract(cfg, resp)
	}

	// --format json: structured JSON output
	if cfg.Format == "json" {
		return displayResponseJSON(cfg, resp)
	}

	// --raw: raw body only
	if cfg.Raw {
		fmt.Print(string(resp.Body))
		return nil
	}

	// --status-only: just the code
	if cfg.StatusOnly {
		fmt.Printf("%s\n", renderStatus(resp.StatusCode, fmt.Sprintf("%d", resp.StatusCode)))
		return nil
	}

	// --quiet: just body
	if cfg.Quiet {
		if len(resp.Body) > 0 {
			fmt.Println(FormatJSON(string(resp.Body)))
		}
		return nil
	}

	// Full pretty output
	fmt.Println(st.Header.Render("RESPONSE"))
	fmt.Println()
	// resp.Status is already "200 OK" — don't add StatusCode again
	fmt.Printf("%s %s\n", renderStatus(resp.StatusCode, "Status:"), renderStatus(resp.StatusCode, resp.Status))
	fmt.Printf("%s %s\n", st.Key.Render("Time:"), st.Value.Render(resp.Duration.String()))
	fmt.Printf("%s %s\n", st.Key.Render("Size:"), st.Value.Render(formatSize(len(resp.Body))))
	fmt.Println()

	if len(resp.Headers) > 0 && cfg.Verbose {
		fmt.Println(st.Key.Render("Headers:"))
		for k, v := range resp.Headers {
			fmt.Printf("  %s: %s\n", st.Key.Render(k), st.Value.Render(strings.Join(v, ", ")))
		}
		fmt.Println()
	}

	if len(resp.Body) > 0 {
		fmt.Println(st.Key.Render("Body:"))
		fmt.Println(st.Body.Render(FormatJSON(string(resp.Body))))
	}
	fmt.Println()

	return nil
}

// displayResponseJSON outputs a structured JSON representation of the response
func displayResponseJSON(cfg *config.Config, resp *client.Response) error {
	headers := make(map[string][]string)
	for k, v := range resp.Headers {
		headers[k] = v
	}

	// Try to parse body as JSON for nested output; fall back to string
	var bodyField interface{}
	var jsonObj interface{}
	if err := json.Unmarshal(resp.Body, &jsonObj); err == nil {
		bodyField = jsonObj
	} else {
		bodyField = string(resp.Body)
	}

	output := JSONResponse{
		OK:         resp.StatusCode < 400,
		Status:     resp.StatusCode,
		StatusText: resp.Status,
		DurationMs: resp.Duration.Milliseconds(),
		Headers:    headers,
		Body:       bodyField,
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON output: %w", err)
	}
	fmt.Println(string(data))

	// Also save to file if --output is set
	if cfg.OutputFile != "" {
		if err := os.WriteFile(cfg.OutputFile, resp.Body, 0644); err != nil {
			return fmt.Errorf("error saving file: %w", err)
		}
	}

	return nil
}

// displayExtract extracts a value from JSON response body using dot notation
// Supports: .data.token, .items[0].id, .count
func displayExtract(cfg *config.Config, resp *client.Response) error {
	var root interface{}
	if err := json.Unmarshal(resp.Body, &root); err != nil {
		return fmt.Errorf("--extract requires JSON response body, got: %s", err)
	}

	path := strings.TrimPrefix(cfg.Extract, ".")

	result, err := walkPath(root, path)
	if err != nil {
		return fmt.Errorf("extract %q: %w", cfg.Extract, err)
	}

	// Print the extracted value
	switch v := result.(type) {
	case string:
		fmt.Println(v)
	case float64:
		// Print integers without decimal point
		if v == float64(int64(v)) {
			fmt.Println(int64(v))
		} else {
			fmt.Println(v)
		}
	case bool:
		fmt.Println(v)
	case nil:
		fmt.Println("null")
	default:
		// Complex types: re-marshal as JSON
		data, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return fmt.Errorf("error formatting extracted value: %w", err)
		}
		fmt.Println(string(data))
	}
	return nil
}

// walkPath navigates a parsed JSON structure using dot notation with array index support
// Examples: "data.token", "items[0].id", "users[2].name"
func walkPath(obj interface{}, path string) (interface{}, error) {
	if path == "" {
		return obj, nil
	}

	segments := splitPath(path)
	current := obj

	for _, seg := range segments {
		// Check for array index: "items[0]"
		name, idx, hasIdx := parseSegment(seg)

		if name != "" {
			m, ok := current.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("expected object at %q, got %T", name, current)
			}
			val, exists := m[name]
			if !exists {
				return nil, fmt.Errorf("key %q not found", name)
			}
			current = val
		}

		if hasIdx {
			arr, ok := current.([]interface{})
			if !ok {
				return nil, fmt.Errorf("expected array at %q, got %T", seg, current)
			}
			if idx < 0 || idx >= len(arr) {
				return nil, fmt.Errorf("index %d out of range (length %d)", idx, len(arr))
			}
			current = arr[idx]
		}
	}

	return current, nil
}

// splitPath splits "data.items[0].id" into ["data", "items[0]", "id"]
func splitPath(path string) []string {
	var segments []string
	current := ""
	for i := 0; i < len(path); i++ {
		if path[i] == '.' {
			if current != "" {
				segments = append(segments, current)
				current = ""
			}
		} else {
			current += string(path[i])
		}
	}
	if current != "" {
		segments = append(segments, current)
	}
	return segments
}

// parseSegment parses "items[0]" → ("items", 0, true) or "data" → ("data", 0, false)
func parseSegment(seg string) (string, int, bool) {
	bracketStart := strings.Index(seg, "[")
	if bracketStart == -1 {
		return seg, 0, false
	}
	bracketEnd := strings.Index(seg, "]")
	if bracketEnd == -1 || bracketEnd <= bracketStart+1 {
		return seg, 0, false
	}

	name := seg[:bracketStart]
	idxStr := seg[bracketStart+1 : bracketEnd]
	idx, err := strconv.Atoi(idxStr)
	if err != nil {
		return seg, 0, false
	}
	return name, idx, true
}

// FormatJSON attempts to format a JSON string with indentation
func FormatJSON(jsonStr string) string {
	var jsonObj interface{}
	if err := json.Unmarshal([]byte(jsonStr), &jsonObj); err != nil {
		return jsonStr // Return as-is if not valid JSON
	}

	prettyJSON, err := json.MarshalIndent(jsonObj, "", "  ")
	if err != nil {
		return jsonStr
	}

	return string(prettyJSON)
}

// formatSize returns a human-readable byte size string.
func formatSize(bytes int) string {
	switch {
	case bytes >= 1024*1024:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(1024*1024))
	case bytes >= 1024:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(1024))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
