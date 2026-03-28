# Environments Guide

HTTP CLI supports environment variables through `.env` files and `{{variable}}` interpolation, similar to tools like dotenv and Postman environments.

## Creating Environment Files

### Base defaults (`.env`)
```env
BASE_URL=https://dev.api.example.com
API_TOKEN=dev-token-123
API_VERSION=v2
```

### Local overrides (`.env.local`)
```env
API_TOKEN=my-personal-dev-token
DEBUG=true
```

### Environment-specific (`.env.prod`, `.env.staging`)
```env
# .env.prod
BASE_URL=https://api.example.com
API_TOKEN=prod-token-abc
```

## Loading Order

Environment files are loaded in this strict order (later files override earlier ones):

```
1. .env              → base defaults
2. .env.local        → personal local overrides
3. .env.<envName>    → environment-specific (when --env is used)
```

### Example
With `--env prod`, the loader reads:
1. `.env` → sets `BASE_URL=https://dev.api.example.com`
2. `.env.local` → overrides `API_TOKEN` with your personal token
3. `.env.prod` → overrides `BASE_URL=https://api.example.com`

## Using Variables

Variables use `{{VAR_NAME}}` syntax and work in:
- **URLs**: `-u {{BASE_URL}}/users`
- **Headers**: `-H "Authorization:Bearer {{API_TOKEN}}"`
- **Body**: `-d '{"key":"{{SECRET}}"}'`
- **Auth flags**: `-b {{API_TOKEN}}`

### Examples
```bash
# Direct request
http-cli -u '{{BASE_URL}}/users' -b '{{API_TOKEN}}'

# With specific environment
http-cli -u '{{BASE_URL}}/users' --env prod

# Saved request with env
http-cli collection run auth/login --env staging
```

## Strict Variable Checking

By default, HTTP CLI **fails fast** if a referenced variable is not found:

```
Error: environment variable 'BASE_URL' not found
```

This prevents silent bugs from unresolved variables hitting your API.

### Bypass strict checking
```bash
http-cli -u '{{BASE_URL}}/data' --ignore-missing-env
```

## Global Default Environment

Instead of always typing `--env dev`, set a global default:

Create `~/.httli/config.json`:
```json
{
  "default_env": "dev"
}
```

Now HTTP CLI automatically loads `.env.dev` on every command unless overridden with `--env`.

## Best Practices

1. **Never commit `.env.local`** — Add it to `.gitignore`
2. **Commit `.env` and `.env.example`** — Share defaults with your team
3. **Use `--dry-run` to verify interpolation** before hitting the network
4. **Use descriptive variable names** — `AUTH_API_BASE_URL` over `URL`

## Gitignore Template
```gitignore
.env.local
.env.*.local
```
