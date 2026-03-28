# Collections Guide

Collections let you save, organize, run, export, and share your API requests — like Postman, but in the terminal.

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

### Naming Conventions
- Use `/` to organize logically: `auth/login`, `users/create`, `orders/list`
- Names are **automatically normalized** to lowercase
- Invalid names like `auth//login` or empty names are rejected

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

## Listing & Inspecting

```bash
# List all saved requests
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

## Storage

Collections are stored at `~/.httli/collections.json` in human-readable, indented JSON format.
