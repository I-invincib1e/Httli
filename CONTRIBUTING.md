# Contributing to HTTP CLI

Thank you for your interest in contributing! This document provides guidelines and information for contributors.

## 🚀 Quick Start

1. **Fork** the repository
2. **Clone** your fork:
   ```bash
   git clone https://github.com/YOUR-USERNAME/Httli.git
   cd Httli
   ```
3. **Build** the project:
   ```bash
   go build -o http-cli ./cmd/http-cli/main.go
   ```
4. **Create a branch** for your feature:
   ```bash
   git checkout -b feature/your-feature-name
   ```

## 📁 Project Structure

```
cmd/            → CLI command definitions (routing, flags)
internal/       → Business logic (NEVER import from outside)
  client/       → HTTP execution engine
  config/       → Flag parsing, env loading, global config
  collections/  → Collection storage (JSON)
  history/      → Request history engine
  output/       → Terminal output renderer
  styles/       → Lipgloss style definitions
docs/           → How-to guides
assets/         → Images and visual assets
```

## 🧪 Development Workflow

### Building
```bash
go build -o http-cli ./cmd/http-cli/main.go
```

### Testing your changes
```bash
# Run a simple request
./http-cli -u https://jsonplaceholder.typicode.com/posts/1

# Test collection workflow
./http-cli collection save test/ping -u https://httpbin.org/get
./http-cli collection run test/ping
./http-cli collection delete test/ping

# Test history
./http-cli history
```

### Code Style
- Follow standard Go conventions (`gofmt`, `go vet`)
- Keep functions small and focused
- Add comments for exported functions
- Use the existing helper patterns (e.g., `normalizeName()`, `exitError()`)

## 🎯 What to Contribute

### Good First Issues
- Add more status code descriptions in output
- Improve error messages
- Add unit tests for `internal/` packages

### Feature Ideas
- Parallel request execution (`--parallel 5`)
- Interactive TUI mode
- Request chaining (use response from one request in the next)
- OpenAPI/Swagger import
- Response time statistics

### Documentation
- Fix typos or unclear instructions
- Add examples for specific use cases
- Translate documentation

## 📝 Commit Messages

Use clear, descriptive commit messages:

```
feat: add parallel request execution
fix: handle empty response body gracefully
docs: update environment variable guide
refactor: extract auth logic into helper
```

## 🔄 Pull Request Process

1. Ensure your code builds cleanly: `go build ./...`
2. Test your changes manually
3. Update documentation if you changed behavior
4. Open a PR with a clear description of what changed and why
5. Link any related issues

## 🏗️ Architecture Guidelines

### Zero Dependencies Policy
This project intentionally avoids external dependencies. Before adding any:
- Consider if the Go standard library can achieve the same goal
- If truly necessary, discuss in an issue first

### Adding New Commands
1. Create a new file in `cmd/` (e.g., `cmd/mycommand.go`)
2. Define your `Command` struct with `Use`, `Short`, `Long`, and `Run`
3. Register it in an `init()` function via `RootCmd.AddCommand()`
4. Add business logic in `internal/`

### Adding New Flags
1. Add the field to `Config` struct in `internal/config/config.go`
2. Register the flag in `ParseFlags()` using the helper functions
3. Use the flag value in the appropriate execution path

## 📄 License

By contributing, you agree that your contributions will be licensed under the MIT License.
