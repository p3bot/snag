# `--version` / `-v`

**Status:** Complete (2025-10-23)

#### Validation Rules

**Boolean Flag:**

- No value required (presence = enabled)
- No validation errors possible

**Priority Behavior:**

- Displays version and exits immediately
- Exit code 0 (success)
- Lower priority than `--help`

#### Behavior

**Basic Usage:**

```bash
snag --version
snag -v
```

- Displays version number
- Exits with code 0
- No other operations performed

**Version Format:**

- Format: `snag version {version}` (e.g., `snag version 0.0.3`)
- Version set at build time via `-ldflags`

#### Interaction Matrix

**With All Other Flags:**

| Combination             | Behavior                | Rationale                                       |
| ----------------------- | ----------------------- | ----------------------------------------------- |
| `--version` alone       | Display version, exit 0 | Standard version mode                           |
| `--version <url>`       | Display version, exit 0 | Version takes priority, URL ignored             |
| `--version --help`      | Display help, exit 0    | **Help takes priority**                         |
| `--help --version`      | Display help, exit 0    | **Help takes priority**                         |
| `--version --doctor`    | Display version, exit 0 | Version takes priority over doctor              |
| `--doctor --version`    | Display version, exit 0 | Version takes priority (regardless of order)    |
| `--version --skill`     | Display version, exit 0 | Version takes priority over skill print         |
| `--version` + any flags | Display version, exit 0 | Version ignores all other flags (except --help) |

**Priority Rules:**

1. `--help` detected → Display help (higher priority)
2. Otherwise, `--version` detected → Display version
3. Ignore all other flags (including `--doctor`, `--kill-browser`, and skill verbs)
4. Exit with code 0

#### Examples

**Valid (Display version and exit):**

```bash
snag --version                                      # Basic version
snag -v                                             # Short form
snag --version https://example.com                  # Version (URL ignored)
snag --version -o file.md --format pdf              # Version (everything ignored)
```

**Help Takes Priority:**

```bash
snag --version --help                               # Shows HELP (not version)
snag --help --version                               # Shows HELP (not version)
```

**No Invalid Combinations:**

- Version flag ignores all other input (except --help)
- Always succeeds (exit 0)

#### Implementation Details

**Location:**

- Flag definition: Built into `github.com/spf13/cobra` framework
- Version variable: `internal/cli/root.go` (package-level `Version`)

**Processing:**

1. CLI framework checks for `--help` first
2. If no help, checks for `--version` or `-v`
3. If present, displays version string
4. Exits with code 0
5. No custom validation code runs

**Version String:**

- Default: `"dev"` (development builds)
- Release: Set via build flag `-ldflags "-X github.com/p3bot/snag/internal/cli.Version=0.0.3"`
- Format: Controlled by CLI framework

---
