<div align="center">

![HTTP CLI Banner](assets/hero_banner.png)

# HTTP CLI

**A zero-dependency, colorful command-line HTTP client and API workflow platform.**

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

[Getting Started](#-getting-started) •
[Features](#-features) •
[Scripting & CI](#-scripting--ci) •
[Collections](#-collections) •
[Environments](#-environments) •
[History](#-history)

</div>

---

## ⚡ What is HTTP CLI?

HTTP CLI is a fast, beautiful, and **zero-dependency** command-line HTTP client built in Go. It goes beyond simple HTTP requests to provide a complete **developer workflow platform** — think Postman, but in your terminal.

```bash
# Simple request
httli -u https://api.github.com/users/octocat

# Save, reuse, and share API workflows
httli collection save auth/login -m POST -u {{BASE_URL}}/login -d '{"user":"admin"}'
httli collection run auth/login --env prod

# Run test suites locally or in CI
httli collection run-all auth/ --fail-fast --timeout 10s
```

---

## 📦 Installation

### Build from source

```bash
git clone https://github.com/I-invincib1e/Httli.git
cd Httli
go build -o httli ./cmd/httli/main.go
```

### Add to PATH (optional)

```bash
# Linux/macOS
sudo mv httli /usr/local/bin/

# Windows (PowerShell as Admin)
Move-Item httli.exe C:\Windows\System32\
```

### Enable Shell Autocomplete

```bash
# Bash
source <(httli completion bash)

# Zsh
source <(httli completion zsh)

# PowerShell
httli completion powershell | Out-String | Invoke-Expression
```

---

## 🚀 Getting Started

### Your first request

```bash
httli -u https://jsonplaceholder.typicode.com/posts/1
```

### Using subcommands

```bash
httli request send -m GET -u https://api.github.com/users/octocat
```

### Quick reference

| Command | Description |
|---------|-------------|
| `httli -u <url>` | Quick GET request |
| `httli request send [flags]` | Full request with all options |
| `httli collection [cmd]` | Manage saved requests (`save`, `run`, `run-all`, `list`) |
| `httli history` | View request history |
| `httli rerun <n>` | Re-execute from history |
| `httli completion <shell>` | Generate autocomplete scripts |

---

## 🤖 Scripting & CI

Httli is fully equipped for headless execution, automated tests, and shell pipelines.

### JSON Output & Extraction
Output structured JSON to pass to `jq`, or use Httli's native JSON extractor to bypass external tools entirely. Httli adds a helpful `ok` boolean for quick success checking.

```bash
# Extract value natively using dot notation (supports arrays)
httli -u https://api.example.com/data --extract .items[0].id

# Or pipe structured JSON to jq
httli -u https://api.example.com/data --format json | jq -e '.ok'
```

### Piping Stdin
Pass file data or piped payload bodies directly using `@-` syntax:
```bash
echo '{"key":"value"}' | httli -m POST -u https://api.example.com -d @-
```

### Headless Controls
Perfect for scripts and CI/CD validation.

| Flag | Role |
|------|------|
| `--fail` (`-F`) | Exits with code `22` natively on HTTP 4xx/5xx responses. |
| `--silent` (`-S`) | Suppresses all CLI output entirely—only the exit code remains. |
| `--status-only` (`-s`) | Prints exactly only the response code (e.g., `200`). |
| `--timeout 5s` | Uses Go duration syntax to strictly enforce timeout limits. |

---

## 📁 Collections & Workflows

Save, organize, and replay your API requests. Collections can be scoped globally to your machine or locally to your project workspace.

### Project-Local Storage
If Httli detects a `.httli/` directory in your current working directory, it will save collections and history there instead of globally. Add this folder to your repository to share API specifications closely with your code!

### Namespace Grouping
Group related requests automatically by adding a slash prefix:
```bash
httli collection save auth/login -m POST -u {{BASE_URL}}/login
httli collection save auth/refresh -m POST -u {{BASE_URL}}/refresh
```

### Run-All Batch Execution
Run a full folder of grouped queries automatically with state chaining:
```bash
$ httli collection run-all auth/ --fail-fast --timeout 10s
```

After executing a batch, `run-all` generates a production-grade summary table outlining successful attempts, failed endpoints, and cumulative response times.

**State Chaining**: `run-all` automatically handles passing data between sequence execution via environment variables:
- `HTTLI_LAST_STATUS` (e.g. "200")
- `HTTLI_LAST_BODY_PATH` (Absolute path to a temporary file storing the raw body)
- `HTTLI_LAST_JSON` (Direct body string, available if the response is valid JSON and <32KB)

### Export & Import (Global format)
```bash
# Export
httli collection export team-api.json

# Import with conflict handling
httli collection import team-api.json              # merge (default)
httli collection import team-api.json --overwrite   # replace conflicts
```

---

## 🌍 Environments

Use `.env` files with `{{variable}}` interpolation across URLs, headers, body, and auth.

### Loading order
```
.env          → base defaults
.env.local    → local overrides
.env.<name>   → environment-specific (via --env flag)
```

### Usage
```bash
# Uses .env + .env.local + .env.prod
httli collection run auth/login --env prod
```

---

## 📊 History

Every executed request is automatically saved. The last **50 requests** are retained.

```bash
# View history
httli history

# Inspect a specific entry
httli history show 1

# Re-execute a previous request (applying local overrides)
httli rerun 1 --timeout 10s --format json
```

---

## 🔧 All Flags

| Flag | Long Form | Description |
|------|-----------|-------------|
| `-m` | `--method` | HTTP method (default: GET) |
| `-u` | `--url` | URL to request (required) |
| `-d` | `--data` | Request body (JSON string, `@-` for stdin, `@path` for file) |
| `-f` | `--file` | Read request body from file |
| `-H` | `--header` | Headers (`Key:Value,Key2:Value2`) |
| `-b` | `--bearer` | Bearer token |
| `-a` | `--auth` | Basic auth (`user:pass`) |
| `-o` | `--output` | Save response body to file |
| `-e` | `--env` | Environment name (loads `.env.<name>`) |
| `-t` | `--timeout` | Timeout duration (e.g. `5s`, `1m`) (default: `30s`) |
| `-r` | `--retry` | Number of retries on failure |
| | `--retry-delay` | Delay between retries in seconds |
| | `--dry-run` | Print request without execution |
| | `--ignore-missing-env` | Don't fail on missing `{{VAR}}` |
| | `--format` | Output format (e.g., `json`) |
| `-x` | `--extract` | Extract JSON response property (`.data.token`) |
| `-F` | `--fail` | Exit with error code 22 on HTTP 4xx/5xx |
| `-L` | `--follow` | Follow redirects |
| `-S` | `--silent` | Total silence. Exit code only. |
| `-q` | `--quiet` | Only show response body |
| `-v` | `--verbose` | Show all details |
| `-s` | `--status-only` | Show exactly only the integer status code |

---

## 🤝 Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## 📄 License

This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.
