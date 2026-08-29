# Validation Rules and Order

**Last Updated:** 2026-08-29

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

**Repeatable (not last-wins):**

- `--skill-install=id` and `--skill-uninstall=id` accumulate. Each occurrence is one agent id. A comma is part of that id, not a list separator.

**Mutually Exclusive Flags:**

- **Logging flags** (`--verbose`, `--debug`) are mutually exclusive
- **Skill verbs** (`--skill`, `--skill-install`, `--skill-list`, `--skill-uninstall`) are mutually exclusive
- Using multiple logging flags or multiple skill verbs together results in an error

**Examples:**

```bash
snag -o file1.md -o file2.md https://example.com  # Uses file2.md (last flag wins)
snag --port 9222 --port 9223 https://example.com  # Uses port 9223 (last flag wins)
snag --verbose --debug https://example.com        # Error: mutually exclusive
snag --skill-install=grok --skill-install=claude-code  # Both ids apply
snag --skill --skill-list                         # Error: mutually exclusive skill verbs
```

### Priority Order for Special Flags

Certain flags override all others and exit immediately:

1. `--help` (highest priority) → Display help, exit 0
2. `--version` → Display version, exit 0
3. Skill flags (`--skill`, `--skill-install`, `--skill-list`, `--skill-uninstall`) → skill mode, or usage-class error if combined with other operation modes
4. `--list-tabs` → List tabs, exit 0, ignore all flags except `--port` and logging flags

---

## Validation Order

**Current implementation order in `internal/cli/root.go` (`runCobra`):**

1. Cobra validates logging flags are mutually exclusive (`--verbose`, `--debug`)
2. Cobra validates skill verbs are mutually exclusive (`--skill`, `--skill-install`, `--skill-list`, `--skill-uninstall`)
3. Initialize logger with selected logging level
4. Handle `--help` → exit early (handled by CLI framework)
5. Handle `--version` → exit early (wins over doctor, skill, url-file, and all other flags)
6. Handle skill flags → print/install/list/uninstall, or usage-class error with other operation modes / URL positionals
7. Handle `--doctor` → exit early
8. Handle `--open-browser` without URL → exit early
9. Handle `--list-tabs` → extract `--port` and logging flags, ignore all others, list tabs, exit early
10. Handle `--all-tabs` → check for URL conflict, exit early
11. Handle `--tab` → check for URL conflict, exit early
12. Validate URL argument required (if not in special modes above)
13. Validate URL format
14. Validate `-o` + `-d` conflict
15. Validate format
16. Validate timeout
17. Validate port
18. Validate output path (if `-o`)
19. Validate output directory (if `-d`)
20. Execute fetch operation

**Key Patterns:**

- Early exits for standalone modes (help, version, skill, doctor, list-tabs, open-browser)
- Content source validation before output validation
- Mutually exclusive flag checks before individual flag validation
- Path/filesystem validation happens last (just before operation)

---

## Notes

For argument-specific validation rules, error messages, and interaction matrices, see the individual argument documentation files in this directory.
