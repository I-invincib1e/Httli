# Changelog

All notable changes to Httli are documented here.

---

## [v1.1.0] — 2026-04-01

This release is a **stability and reliability overhaul**. Every critical bug from the original codebase has been fixed, the data model has been expanded, a full test suite has been added, and several developer-experience enhancements have been shipped alongside the fixes.

### 🔴 Critical Bug Fixes

- **`.gitignore` corruption** — The file contained UTF-16 null bytes from line 23 onward, causing Git to silently ignore all rules below that point. `.env` (credentials) and `test_export.json` were being tracked. Rewritten as clean UTF-8; both files removed from the index.
- **Instant timeout** — `GetRequest()` set `Timeout: 30`, which Go interprets as 30 **nanoseconds**, causing every collection request to time out instantly. Fixed to `30 * time.Second`.
- **`--fail-fast` crash** — The flag was not registered in the `FlagSet`. With `flag.ExitOnError`, any unrecognised flag calls `os.Exit(2)`, so passing `--fail-fast` to `collection run-all` silently killed the process before execution. Fixed by registering `--fail-fast` as a proper bool flag.

### 🟠 Serious Bug Fixes

- **5xx response body silently discarded** — The retry loop closed the response body before reading it on 5xx responses. Error details from the server were permanently lost. Now the body is always read before closing, and the last 5xx response is returned to the caller (not converted to a plain `error`).
- **`defer` inside retry loop** — `defer resp.Body.Close()` inside a loop stacks defers across all iterations. Replaced with immediate `resp.Body.Close()` after reading.
- **`os.Setenv` overwrites system env vars** — `LoadEnv()` unconditionally called `os.Setenv`, which could clobber `PATH`, `HOME`, or CI platform secrets. Now uses `os.LookupEnv` first and only sets variables that don't already exist.
- **Regex recompiled per call** — `Interpolate()` called `regexp.MustCompile` on every invocation. Moved to a package-level variable, compiled once at startup.
- **Double interpolation** — `ParseFlags()` called `InterpolateAll()`, and then `collection run` and `rerun` called it again on the same config. Any `{{VAR}}` that survived the first pass would error on the second. `ParseFlags()` no longer calls `InterpolateAll()`; callers invoke it explicitly after loading.
- **`collection save` drops fields** — Only `Method`, `URL`, `Headers`, and `Body` were persisted. Bearer token, basic auth, timeout, follow-redirects, retry count, and retry delay were silently discarded. The full `RequestData` struct now persists all fields.
- **`rerun` only restores method + URL** — History entries only stored minimal data. `rerun` rebuilt a bare config with no headers, body, or auth, making it useless for anything but simple GETs. History now records the full request; `rerun` restores it faithfully.
- **`--output` ignored with `--raw`** — `--raw` returned early before the `--output` file-write block ran. `--output` is now orthogonal — it saves to file regardless of which display mode is active.

### 🟡 Minor Bug Fixes

- **Duplicate status in output** — `resp.Status` already contains `"200 OK"`. It was being rendered as `"200 200 OK"`.
- **`merge` and `skip` import modes were identical** — Both skipped existing entries. `merge` now adds new entries and skips conflicts; `skip` explicitly skips all conflicting entries; `overwrite` replaces them.
- **`env list` printed empty string** — The subcommand printed a blank line. Now shows which `.env` files loaded (with ✓/✗ indicators), HTTLI chain variables, and all uppercase env vars matching `.env` file conventions.
- **Windows `.env` CRLF** — Values from `.env` files on Windows had trailing `\r` included. Stripped automatically.

### ✨ Enhancements

#### New flags
- `--retry-backoff` — Opt-in exponential backoff for retries. Delay doubles each attempt (`baseDelay × 2^n`), capped at 30 seconds. Default behaviour (fixed delay) is unchanged.
- `--fail-fast` — Now a proper registered flag, available globally (not just `run-all`).

#### New command
- `httli collection describe <name> <text>` — Annotate a saved request with a description. Descriptions appear in `collection list` output as inline comments.

#### Config system
- `ApplyOverrides(runCfg *Config)` method — Replaces a 15-line manual field-copy block that was duplicated across `collection.go`, `collection_runall.go`, and `history.go`.
- `flag.ExitOnError` → `flag.ContinueOnError` — Unknown flags now return errors instead of calling `os.Exit`, enabling proper error messages.
- Updated `PrintUsage()` — Now reflects the actual command tree.

#### Collections
- `RequestData` expanded — Persists `bearer_token`, `basic_auth`, `timeout` (as `"15s"` string), `follow_redirects`, `retry`, `retry_delay`, `description`, `created_at`, `updated_at`. All new fields use `omitempty`; existing `collections.json` files load without migration.
- `collection list` — Shows description as an inline comment next to each entry.

#### History
- `Entry` expanded — Now stores `headers`, `body`, `bearer_token`, `basic_auth`, `duration_ms`.
- `Record()` signature updated — Accepts `*config.Config` and `durationMs int64` instead of individual method/URL strings.
- History list — Displays response duration per entry.
- `history show` — Displays full request details including headers and body.

#### Output
- Response size always shown — Human-readable (`292 B`, `1.2 KB`, `2.0 MB`). Previously only shown in `--verbose` mode.
- `formatSize()` helper added.

#### Styles
- `Warning` style added (yellow, ANSI 226) — Used for retry attempt messages and future deprecation notices.

#### Version
- `Version` variable defaults to `"dev"` and is injectable via ldflags:
  ```bash
  go build -ldflags "-X github.com/I-invincib1e/httli/cmd.Version=1.1.0" -o httli ./cmd/httli
  ```

#### Environment safety
- `LoadEnv()` no longer overwrites variables already present in the shell — CI/CD platform secrets always take priority.

### 🧪 Tests Added

Zero test files existed before this release. The following were added (76 test cases total):

| File | What's tested |
|------|---------------|
| `internal/config/config_test.go` | `ParseHeaders`, `InterpolateAll`, `ApplyOverrides`, `ParseFlags` |
| `internal/config/env_test.go` | `LoadEnv` no-overwrite guarantee, CRLF stripping, comment parsing, `Interpolate` |
| `internal/client/client_test.go` | 200/4xx/5xx responses, retry count, retry success, `calcDelay`, invalid JSON |
| `internal/output/output_test.go` | `walkPath`, `splitPath`, `parseSegment`, `FormatJSON`, `formatSize` |
| `internal/collections/collections_test.go` | `normalizeName`, `configToRequestData`, timeout string round-trip |

### 📝 Documentation

- `README.md` fully rewritten — accurate to all changes in this release, includes architecture diagram, request lifecycle walkthrough, storage locations, flag system internals, and retry algorithm documentation.

---

## [v1.0.0] — Initial release

- Zero-dependency CLI HTTP client with colored output
- Collections: save, update, delete, run, run-all, export, import
- Environment variable interpolation via `.env` files
- Request history with rerun
- Shell autocomplete (bash, zsh, fish, PowerShell)
- JSON extraction via dot-notation (`--extract`)
- Structured JSON output (`--format json`)
- CI/CD flags: `--fail`, `--silent`, `--status-only`, `--dry-run`
- Bearer and basic auth
- Stdin body via `@-` syntax
- Project-local `.httli/` storage
- State chaining in `run-all` via `HTTLI_LAST_*` env vars
