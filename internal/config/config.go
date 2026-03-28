package config

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

// Config holds all configuration for the HTTP request
type Config struct {
	Method           string
	URL              string
	Headers          map[string]string
	Body             string
	BodyFile         string
	Timeout          int
	FollowRedirects  bool
	BearerToken      string
	BasicAuth        string
	OutputFile       string
	Quiet            bool
	Verbose          bool
	StatusOnly       bool
	Env              string
	IgnoreMissingEnv bool
	Retry            int
	RetryDelay       int
	DryRun           bool
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

// ParseFlags parses command-line flags and returns a Config struct
func ParseFlags(args []string) (*Config, error) {
	var method, url, body, headers, bodyFile, bearerToken, basicAuth, outputFile, env string
	var timeout, retry, retryDelay int
	var followRedirects, quiet, verbose, statusOnly, ignoreMissingEnv, dryRun bool

	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
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

	registerStr("m", "method", &method, "GET", "HTTP method (GET, POST, PUT, DELETE, PATCH)")
	registerStr("u", "url", &url, "", "URL to request")
	registerStr("d", "data", &body, "", "Request body (JSON string)")
	registerStr("f", "file", &bodyFile, "", "Read request body from file")
	registerStr("H", "header", &headers, "", "Headers (format: 'Key:Value')")
	registerStr("b", "bearer", &bearerToken, "", "Bearer token")
	registerStr("a", "auth", &basicAuth, "", "Basic auth (user:pass)")
	registerStr("o", "output", &outputFile, "", "Save response body to file")
	registerStr("e", "env", &env, "", "Environment name (loads .env.<name>)")
	registerInt("t", "timeout", &timeout, 30, "Request timeout in seconds")
	registerBool("L", "follow", &followRedirects, "Follow redirects")
	registerBool("q", "quiet", &quiet, "Quiet mode")
	registerBool("v", "verbose", &verbose, "Verbose mode")
	registerBool("s", "status-only", &statusOnly, "Show only status code")
	
	// New flags
	fs.BoolVar(&ignoreMissingEnv, "ignore-missing-env", false, "Do not fail when {{VAR}} is missing")
	registerInt("r", "retry", &retry, 0, "Number of retries for failed network or 5xx requests")
	fs.IntVar(&retryDelay, "retry-delay", 2, "Delay in seconds between retries")
	fs.BoolVar(&dryRun, "dry-run", false, "Print request without execution")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	if env == "" && Global.DefaultEnv != "" {
		env = Global.DefaultEnv
	}
	LoadEnv(env)

	finalBody := body
	if bodyFile != "" {
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
		Quiet:            quiet,
		Verbose:          verbose,
		StatusOnly:       statusOnly,
		Env:              env,
		IgnoreMissingEnv: ignoreMissingEnv,
		Retry:            retry,
		RetryDelay:       retryDelay,
		DryRun:           dryRun,
	}

	if err := cfg.InterpolateAll(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.URL == "" {
		return fmt.Errorf("URL is required")
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
	fmt.Fprintf(os.Stderr, "HTTP CLI - Developer workflow tool\n\n")
	fmt.Fprintf(os.Stderr, "Usage: %s [options] | %s [subcommand]\n\n", os.Args[0], os.Args[0])
	fmt.Fprintf(os.Stderr, "Subcommands:\n")
	fmt.Fprintf(os.Stderr, "  save <name> [options]    Save a request (fails if exists)\n")
	fmt.Fprintf(os.Stderr, "  update <name> [options]  Update an existing request\n")
	fmt.Fprintf(os.Stderr, "  delete <name>            Delete a saved request\n")
	fmt.Fprintf(os.Stderr, "  show <name>              Display a saved request\n")
	fmt.Fprintf(os.Stderr, "  run <name>               Run a saved request\n")
	fmt.Fprintf(os.Stderr, "  collection list          List all saved requests\n\n")
	fmt.Fprintf(os.Stderr, "Run with -h for flag details.\n")
}
