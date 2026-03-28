# Collections Guide

Collections let you save, organize, run, export, and share your API requests — like Postman, but in the terminal.

## Local vs Global Storage

By default, collections are stored globally at `~/.httli/collections.json`. 

**Project-Local Collections:** If Httli detects a `.httli/` directory in your current working directory, it will automatically use `{current_dir}/.httli/collections.json`. 
This allows you to commit your `.httli/collections.json` to version control and share your API suite directly alongside your codebase!

## Saving Requests

```bash
httli collection save <name> [flags]
```

The `save` command **fails if the request already exists** to prevent accidental overwrites.

```bash
# Save a GET request
httli collection save users/list -m GET -u https://api.example.com/users

# Save a POST request with body
httli collection save users/create \
  -m POST \
  -u https://api.example.com/users \
  -d '{"name":"John","email":"john@example.com"}'

# Save with authentication
httli collection save admin/dashboard \
  -u https://api.example.com/admin \
  -b '{{API_TOKEN}}'
```

### Naming Conventions & Namespaces
- Use `/` to organize logically into namespaces: `auth/login`, `users/create`, `orders/list`
- When you use `collection list`, requests are automatically grouped by these folders.
- Names are automatically normalized to lowercase.
- Invalid names like `auth//login` or empty names are rejected.

## Updating Requests

```bash
httli collection update <name> [flags]
```

The `update` command **fails if the request doesn't exist** — use `save` first.

```bash
httli collection update users/create -d '{"name":"Jane","role":"admin"}'
```

## Running Saved Requests

```bash
httli collection run <name> [flags]
```

You can override runtime behavior:
```bash
# Run with a specific environment
httli collection run auth/login --env prod

# Run with retries
httli collection run health/check --retry 3 --retry-delay 2

# Dry run (no network call)
httli collection run auth/login --dry-run
```

## Batch Execution with `run-all`

Run a full folder of grouped queries automatically.

```bash
# Run everything in the auth/ folder
httli collection run-all auth/ --fail-fast --timeout 5s
```

After executing a batch, `run-all` generates a production-grade summary table outlining successful attempts, failed endpoints, and cumulative response times.

### State Chaining
`run-all` automatically handles passing data between sequences via environment variables:
- `HTTLI_LAST_STATUS` (e.g. "200")
- `HTTLI_LAST_BODY_PATH` (Absolute path to a temporary file storing the raw body)
- `HTTLI_LAST_JSON` (Direct body string, available if the response is valid JSON and <32KB)

## Listing & Inspecting

```bash
# List all saved requests (grouped logically)
httli collection list

# Show full details of a request (formatted output)
httli collection show auth/login
```

## Deleting Requests

```bash
httli collection delete auth/login
```

## Export & Import

### Exporting
```bash
httli collection export my-api-collection.json
```

This creates a shareable JSON file containing all your saved requests.

### Importing
```bash
# Default: merge (add new, skip existing)
httli collection import team-collection.json

# Overwrite existing conflicts
httli collection import team-collection.json --overwrite

# Skip all conflicts
httli collection import team-collection.json --skip
```

### Team Workflow Example
```bash
# Developer A exports their collection
httli collection export api-v2.json
git add api-v2.json && git commit -m "Add API collection"

# Developer B imports it
git pull
httli collection import api-v2.json
httli collection run auth/login --env dev
```
