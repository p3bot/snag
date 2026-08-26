# Validation Rules and Order

**Last Updated:** 2025-10-26

This document describes the validation order and cross-cutting validation rules that apply to multiple arguments.

---

## Cross-Cutting Validation Rules

### String Argument Trimming

All string arguments are trimmed using `strings.TrimSpace()` after reading from CLI framework:

- Removes leading and trailing whitespace (spaces, tabs, newlines)
- Applied to: `--output`, `--output-dir`, `--format`, `--wait-for`, `--user-agent`, `--user-data-dir`, `--tab`, `--url-file`, and `<url>` positional arguments
- Empty strings after trimming are handled per-argument (usually warning + ignored or error)
- Standard behavior in most CLI tools (git, docker, etc.)

### Multiple Flag Behavior

**Last Flag Wins (Standard CLI Behavior):**

- When the same flag is specified multiple times, the last value is used
- No error, no warning - silent override
- Applies to most flags:
  - **String flags**: `--output`, `--output-dir`, `--format`, `--wait-for`, `--user-agent`, `--user-data-dir`, `--tab`, `--url-file`
  - **Integer flags**: `--timeout`, `--port`
  - **Boolean flags**: `--close-tab`, `--force-headless`, `--open-browser`, `--list-tabs`, `--all-tabs`

**Mutually Exclusive Flags:**

- **Logging flags** (`--verbose`, `--quiet`, `--debug`) are mutually exclusive
- Using multiple logging flags together results in an error
- Only one logging level flag can be used at a time

**Examples:**

```bash
snag -o file1.md -o file2.md https://example.com  # Uses file2.md (last flag wins)
snag --port 9222 --port 9223 https://example.com  # Uses port 9223 (last flag wins)
snag --quiet --verbose https://example.com        # Error: mutually exclusive
```

### Priority Order for Special Flags

Certain flags override all others and exit immediately:

1. `--help` (highest priority) → Display help, exit 0
2. `--version` → Display version, exit 0
3. `--list-tabs` → List tabs, exit 0, ignore all flags except `--port` and logging flags

---

## Validation Order

**Current implementation order in `internal/cli/root.go` (`runCobra`):**

1. Cobra validates logging flags are mutually exclusive (`--quiet`, `--verbose`, `--debug`)
2. Initialize logger with selected logging level
3. Handle `--help` → exit early (handled by CLI framework)
4. Handle `--version` → exit early (handled by CLI framework)
5. Handle `--open-browser` without URL → exit early
6. Handle `--list-tabs` → extract `--port` and logging flags, ignore all others, list tabs, exit early
7. Handle `--all-tabs` → check for URL conflict, exit early
8. Handle `--tab` → check for URL conflict, exit early
9. Validate URL argument required (if not in special modes above)
10. Validate URL format
11. Validate `-o` + `-d` conflict
12. Validate format
13. Validate timeout
14. Validate port
15. Validate output path (if `-o`)
16. Validate output directory (if `-d`)
17. Execute fetch operation

**Key Patterns:**

- Early exits for standalone modes (help, version, list-tabs, open-browser)
- Content source validation before output validation
- Mutually exclusive flag checks before individual flag validation
- Path/filesystem validation happens last (just before operation)

---

## Notes

For argument-specific validation rules, error messages, and interaction matrices, see the individual argument documentation files in this directory.
