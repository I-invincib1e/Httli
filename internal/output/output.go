package output

import (
    "encoding/json"
    "fmt"
    "os"
    "strings"

    "github.com/I-invincib1e/httli/internal/client"
    "github.com/I-invincib1e/httli/internal/config"
    "github.com/I-invincib1e/httli/internal/styles"
    "github.com/charmbracelet/lipgloss"
)

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
    if cfg.Quiet || cfg.StatusOnly {
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
    statusColor := styles.StatusColor(statusCode)
    statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(statusColor)).Bold(true)
    return statusStyle.Render(statusText)
}

// DisplayResponse displays the response information
func DisplayResponse(cfg *config.Config, resp *client.Response, st *styles.Styles) error {
    if cfg.OutputFile != "" {
        if err := os.WriteFile(cfg.OutputFile, resp.Body, 0644); err != nil {
            return fmt.Errorf("error saving file: %w", err)
        }
        if !cfg.Quiet {
            fmt.Printf("%s %s\n", st.Success.Render("Response saved to:"), cfg.OutputFile)
        }
    }

    if cfg.StatusOnly {
        fmt.Printf("%s\n", renderStatus(resp.StatusCode, fmt.Sprintf("%d", resp.StatusCode)))
        return nil
    }

    if cfg.Quiet {
        if len(resp.Body) > 0 {
            fmt.Println(FormatJSON(string(resp.Body)))
        }
        return nil
    }

    fmt.Println(st.Header.Render("RESPONSE"))
    fmt.Println()
    fmt.Printf("%s %s\n", renderStatus(resp.StatusCode, "Status:"), renderStatus(resp.StatusCode, fmt.Sprintf("%d %s", resp.StatusCode, resp.Status)))
    fmt.Printf("%s %s\n", st.Key.Render("Time:"), st.Value.Render(resp.Duration.String()))

    if cfg.Verbose {
        fmt.Printf("%s %s\n", st.Key.Render("Size:"), st.Value.Render(fmt.Sprintf("%d bytes", len(resp.Body))))
    }
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

