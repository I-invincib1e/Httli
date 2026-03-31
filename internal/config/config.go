package config

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// Config holds all configuration for the HTTP request
type Config struct {
	Method           string
	URL              string
	Headers          map[string]string
	Body             string
	BodyFile         string
	Timeout          time.Duration
	FollowRedirects  bool
	BearerToken      string
	BasicAuth        string
	OutputFile       string
	Silent           bool
	Quiet            bool
	Verbose          bool
	StatusOnly       bool
	Env              string
	IgnoreMissingEnv bool
	Retry            int
	RetryDelay       int
	RetryBackoff     bool // opt-in exponential backoff
	DryRun           bool

	// Output control flags
	Format   string // "json" for structured output, "" for pretty
	Fail     bool   // exit 22 on HTTP 4xx/5xx
	FailFast bool   // stop batch execution on first error
	Raw      bool   // output raw body only, no chrome
	Extract  string // dot-notation path to extract from JSON response
}

// InterpolateAll substitutes all {{VAR}} templating in the Config using Env vars.
func (c *Config) InterpolateAll() error {
	var err error
	c.URL, err = Interpolate(c.URL, c.IgnoreMissingEnv)
	if err != nil { return err }

	c.Body, err = Interpolate(c.Body, c.IgnoreMissingEnv)
	if err != nil { return err }

	c.BearerToken, err = Interpolate(c.BearerToken, c.IgnoreMissingEnv)
	if err != nil { return err }

	c.BasicAuth, err = Interpolate(c.BasicAuth, c.IgnoreMissingEnv)
	if err != nil { return err }

	for k, v := range c.Headers {
		c.Headers[k], err = Interpolate(v, c.IgnoreMissingEnv)
		if err != nil { return err }
	}
	return nil
}

// ApplyOverrides copies runtime override flags from a parsed runCfg onto a loaded config.
// This replaces the copy-paste override block used in collection run, run-all, and rerun.
func (c *Config) ApplyOverrides(runCfg *Config) {
	c.IgnoreMissingEnv = runCfg.IgnoreMissingEnv
	c.Retry = runCfg.Retry
	c.RetryDelay = runCfg.RetryDelay
	c.RetryBackoff = runCfg.RetryBackoff
	c.DryRun = runCfg.DryRun
	c.Format = runCfg.Format
	c.Fail = runCfg.Fail
	c.FailFast = runCfg.FailFast
	c.Raw = runCfg.Raw
	c.Extract = runCfg.Extract
	c.Quiet = runCfg.Quiet
	c.StatusOnly = runCfg.StatusOnly
	c.Verbose = runCfg.Verbose
	c.Silent = runCfg.Silent

	// Only override timeout if the user explicitly set a non-default value
	if runCfg.Timeout != 30*time.Second && runCfg.Timeout != 0 {
		c.Timeout = runCfg.Timeout
	}
}

// ParseFlags parses command-line flags and returns a Config struct.
// Note: This does NOT call InterpolateAll(). Callers that need interpolation
// (e.g. collection run, rerun) must call cfg.InterpolateAll() explicitly.
// Direct requests (request send) call InterpolateAll() in the command layer.
func ParseFlags(args []string) (*Config, error) {
	var method, url, body, headers, bodyFile, bearerToken, basicAuth, outputFile, env string
	var format, extract string
	var timeout time.Duration
	var retry, retryDelay int
	var followRedirects, silent, quiet, verbose, statusOnly, ignoreMissingEnv, dryRun bool
	var fail, failFast, raw, retryBackoff bool

	fs := flag.NewFlagSet("httli", flag.ContinueOnError)
	fs.Usage = PrintUsage

	registerStr := func(short, long string, target *string, defVal, usage string) {
		fs.StringVar(target, short, defVal, usage)
		fs.StringVar(target, long, defVal, usage)
	}
	registerBool := func(short, long string, target *bool, usage string) {
		fs.BoolVar(target, short, false, usage)
		fs.BoolVar(target, long, false, usage)
	}
	registerInt := func(short, long string, target *int, defVal int, usage string) {
		fs.IntVar(target, short, defVal, usage)
		fs.IntVar(target, long, defVal, usage)
	}
	registerDuration := func(short, long string, target *time.Duration, defVal time.Duration, usage string) {
		fs.DurationVar(target, short, defVal, usage)
		fs.DurationVar(target, long, defVal, usage)
	}

	registerStr("m", "method", &method, "GET", "HTTP method (GET, POST, PUT, DELETE, PATCH)")
	registerStr("u", "url", &url, "", "URL to request")
	registerStr("d", "data", &body, "", "Request body (JSON string, or @- for stdin, or @file)")
	registerStr("f", "file", &bodyFile, "", "Read request body from file")
	registerStr("H", "header", &headers, "", "Headers (format: 'Key:Value')")
	registerStr("b", "bearer", &bearerToken, "", "Bearer token")
	registerStr("a", "auth", &basicAuth, "", "Basic auth (user:pass)")
	registerStr("o", "output", &outputFile, "", "Save response body to file")
	registerStr("e", "env", &env, "", "Environment name (loads .env.<name>)")
	registerDuration("t", "timeout", &timeout, 30*time.Second, "Request timeout (e.g. 5s, 1m)")
	registerBool("L", "follow", &followRedirects, "Follow redirects")
	registerBool("S", "silent", &silent, "Silent mode (no output at all, only exit code)")
	registerBool("q", "quiet", &quiet, "Quiet mode (output body only)")
	registerBool("v", "verbose", &verbose, "Verbose mode")
	registerBool("s", "status-only", &statusOnly, "Show only status code")
	
	// Extended flags
	fs.BoolVar(&ignoreMissingEnv, "ignore-missing-env", false, "Do not fail when {{VAR}} is missing")
	registerInt("r", "retry", &retry, 0, "Number of retries for failed network or 5xx requests")
	fs.IntVar(&retryDelay, "retry-delay", 2, "Delay in seconds between retries")
	fs.BoolVar(&retryBackoff, "retry-backoff", false, "Use exponential backoff for retries (delay doubles each attempt, capped at 30s)")
	fs.BoolVar(&dryRun, "dry-run", false, "Print request without execution")

	// Output control flags
	fs.StringVar(&format, "format", "", "Output format: 'json' for structured output")
	registerBool("F", "fail", &fail, "Exit with code 22 on HTTP 4xx/5xx responses")
	fs.BoolVar(&failFast, "fail-fast", false, "Stop batch execution on first error (for run-all)")
	fs.BoolVar(&raw, "raw", false, "Output raw response body only (no headers, colors, or formatting)")
	registerStr("x", "extract", &extract, "", "Extract a value from JSON response (dot notation: .data.token)")

	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			// --help or -h was passed; usage already printed by fs.Usage
			os.Exit(0)
		}
		return nil, err
	}

	if env == "" && Global.DefaultEnv != "" {
		env = Global.DefaultEnv
	}
	LoadEnv(env)

	// Handle @- (stdin) and @file body syntax
	finalBody := body
	if body == "@-" {
		stdinData, err := io.ReadAll(os.Stdin)
		if err != nil {
			return nil, fmt.Errorf("error reading stdin: %w", err)
		}
		finalBody = strings.TrimSpace(string(stdinData))
	} else if strings.HasPrefix(body, "@") && body != "@-" {
		filePath := body[1:]
		fileContent, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("error reading body file '%s': %w", filePath, err)
		}
		finalBody = string(fileContent)
	} else if bodyFile != "" {
		fileContent, err := os.ReadFile(bodyFile)
		if err != nil {
			return nil, fmt.Errorf("error reading file: %w", err)
		}
		finalBody = string(fileContent)
	}

	cfg := &Config{
		Method:           strings.ToUpper(method),
		URL:              url,
		Headers:          ParseHeaders(headers),
		Body:             finalBody,
		BodyFile:         bodyFile,
		Timeout:          timeout,
		FollowRedirects:  followRedirects,
		BearerToken:      bearerToken,
		BasicAuth:        basicAuth,
		OutputFile:       outputFile,
		Silent:           silent,
		Quiet:            quiet,
		Verbose:          verbose,
		StatusOnly:       statusOnly,
		Env:              env,
		IgnoreMissingEnv: ignoreMissingEnv,
		Retry:            retry,
		RetryDelay:       retryDelay,
		RetryBackoff:     retryBackoff,
		DryRun:           dryRun,
		Format:           format,
		Fail:             fail,
		FailFast:         failFast,
		Raw:              raw,
		Extract:          extract,
	}

	return cfg, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.URL == "" {
		return fmt.Errorf("URL is required")
	}
	if c.Format != "" && c.Format != "json" {
		return fmt.Errorf("unsupported format %q (supported: json)", c.Format)
	}
	return nil
}

func ParseHeaders(headersStr string) map[string]string {
	headers := make(map[string]string)
	if headersStr == "" { return headers }
	pairs := strings.Split(headersStr, ",")
	for _, pair := range pairs {
		parts := strings.SplitN(strings.TrimSpace(pair), ":", 2)
		if len(parts) == 2 {
			headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return headers
}

func PrintUsage() {
	fmt.Fprintf(os.Stderr, "Httli - A fast and colorful HTTP CLI tool\n\n")
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  httli [flags]                         Quick request mode\n")
	fmt.Fprintf(os.Stderr, "  httli request send [flags]            Send a request\n")
	fmt.Fprintf(os.Stderr, "  httli collection [command] [flags]    Manage saved requests\n")
	fmt.Fprintf(os.Stderr, "  httli history                         View request history\n")
	fmt.Fprintf(os.Stderr, "  httli rerun <index> [flags]           Re-execute from history\n")
	fmt.Fprintf(os.Stderr, "  httli env list                        Show loaded environment\n")
	fmt.Fprintf(os.Stderr, "  httli completion <shell>              Generate shell completions\n")
	fmt.Fprintf(os.Stderr, "  httli version                         Print version\n\n")
	fmt.Fprintf(os.Stderr, "Run 'httli --help' or 'httli <command> --help' for details.\n")
}
