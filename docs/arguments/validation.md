# Validation Rules and Order

**Last Updated:** 2026-09-09

This document describes the validation order and cross-cutting validation rules that apply to multiple arguments.

---

## Cross-Cutting Validation Rules

### String Argument Trimming

All string arguments are trimmed using `strings.TrimSpace()` after reading from CLI framework:

- Removes leading and trailing whitespace (spaces, tabs, newlines)
- Applied to: `--output`, `--output-dir`, `--format`, `--wait-for`, `--user-agent`, `--user-data-dir`, `--tab`, `--url-file`, `--color`, and `<url>` positional arguments
- Empty strings after trimming are handled per-argument (usually warning + ignored or error)
- Standard behavior in most CLI tools (git, docker, etc.)

### Multiple Flag Behavior

**Last Flag Wins (Standard CLI Behavior):**

- When the same flag is specified multiple times, the last value is used
- No error, no warning - silent override
- Applies to most flags:
  - **String flags**: `--output`, `--output-dir`, `--format`, `--wait-for`, `--user-agent`, `--user-data-dir`, `--tab`, `--url-file`, `--color`
  - **Integer flags**: `--timeout`, `--port`
  - **Boolean flags**: `--close-tab`, `--force-headless`, `--open-browser`, `--list-tabs`, `--all-tabs`, `--temp-profile`

**Repeatable (not last-wins):**

- `--skill-install=id` and `--skill-uninstall=id` accumulate. Each occurrence is one agent id. A comma is part of that id, not a list separator.

**Mutually Exclusive Flags:**

- **Logging flags** (`--verbose`, `--debug`) are mutually exclusive
- **Skill verbs** (`--skill`, `--skill-install`, `--skill-list`, `--skill-uninstall`) are mutually exclusive
- **Profile flags** (`--user-data-dir`, `--temp-profile`) are mutually exclusive
- Using multiple logging flags, multiple skill verbs, or both profile flags together results in an error

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
4. `--list-tabs` → List tabs, exit 0; `--port` and logging flags apply; `--user-data-dir` / `--temp-profile` warn then are ignored

---

## Validation Order

**Current implementation order in `internal/cli/root.go` (Cobra execute, then `PersistentPreRunE`, then `runCobra`):**

1. Cobra parses flags (unknown flags and `--color` with no value fail here)
2. Cobra handles `--help` → exit early (no `PersistentPreRunE`; invalid `--color` is not validated)
3. Validate `--color` (`auto`, `always`, `never`), resolve stderr colour, and install the logger (`PersistentPreRunE`)
4. Cobra validates logging flags are mutually exclusive (`--verbose`, `--debug`)
5. Cobra validates skill verbs are mutually exclusive (`--skill`, `--skill-install`, `--skill-list`, `--skill-uninstall`)
6. Cobra validates profile flags are mutually exclusive (`--user-data-dir`, `--temp-profile`)
7. Handle `--version` → exit early (wins over doctor, skill, url-file, and all other flags)
8. Handle skill flags → print/install/list/uninstall, or usage-class error with other operation modes / URL positionals
9. Handle `--doctor` → exit early
10. Handle `--kill-browser` → exit early (errors if combined with URL, tab, list, or open-browser)
11. Handle `--list-tabs` → extract `--port` and logging flags; profile flags warn then are ignored; list tabs, exit early
12. Validate flag combinations (content-source conflicts, `--force-headless` vs `--open-browser` / tabs, `-o` + `-d`, `--info` conflicts)
13. Handle `--info` → single URL or `--tab`, then exit
14. Handle `--all-tabs` → check for URL conflict (already validated), process tabs, exit
15. Handle `--tab` → check for URL conflict (already validated), fetch tab(s), exit
16. Handle `--open-browser` without URL → exit early
17. Validate URL argument required (if not in special modes above)
18. Handle `--open-browser` with URLs → open tabs, no fetch
19. Validate URL format
20. Validate format
21. Validate timeout
22. Validate port
23. Validate output path (if `-o`)
24. Validate output directory (if `-d`)
25. Execute fetch operation

**Key Patterns:**

- Early exits for standalone modes (help, version, skill, doctor, kill-browser, list-tabs, open-browser)
- Content source validation before output validation
- Mutually exclusive flag checks before individual flag validation
- Path/filesystem validation happens last (just before operation)

---

## Notes

For argument-specific validation rules, error messages, and interaction matrices, see the individual argument documentation files in this directory.
