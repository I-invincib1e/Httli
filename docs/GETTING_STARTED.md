# Getting Started with HTTP CLI

This guide walks you through installing, configuring, and making your first API requests with HTTP CLI.

## Installation

### Prerequisites
- [Go 1.21+](https://go.dev/dl/) installed
- Git installed

### Build from source
```bash
git clone https://github.com/I-invincib1e/Httli.git
cd Httli
go build -o http-cli ./cmd/http-cli/main.go
```

### Verify installation
```bash
./http-cli --help
```

You should see the full command tree with `request`, `collection`, `env`, `history`, and `completion` commands.

## Your First Request

```bash
./http-cli -u https://jsonplaceholder.typicode.com/posts/1
```

This makes a GET request and displays the response with beautiful color-coded output:
- Purple section headers
- Green 200 status code
- Pretty-printed JSON body

## POST Request with JSON

```bash
./http-cli -m POST \
  -u https://jsonplaceholder.typicode.com/posts \
  -d '{"title":"Hello","body":"World","userId":1}'
```

## Using Authentication

```bash
# Bearer token
./http-cli -u https://api.example.com/me -b "your-jwt-token"

# Basic auth
./http-cli -u https://api.example.com/me -a "admin:password123"
```

## Save and Reuse Requests

```bash
# Save a request
./http-cli collection save my-api/users -m GET -u https://jsonplaceholder.typicode.com/users

# Run it anytime
./http-cli collection run my-api/users
```

## Set Up Environments

Create a `.env` file in your project directory:
```
BASE_URL=https://jsonplaceholder.typicode.com
```

Now use variables in your requests:
```bash
./http-cli -u '{{BASE_URL}}/posts/1'
```

## Enable Shell Autocomplete

```bash
# Bash
source <(./http-cli completion bash)

# PowerShell
./http-cli completion powershell | Out-String | Invoke-Expression
```

## Next Steps

- 📁 [Collections Guide](COLLECTIONS.md) — Save, organize, and share API workflows
- 🌍 [Environments Guide](ENVIRONMENTS.md) — Master variable interpolation
- 🤝 [Contributing](../CONTRIBUTING.md) — Help improve HTTP CLI
