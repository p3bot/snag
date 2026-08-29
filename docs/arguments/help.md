# `--help` / `-h`

**Status:** Complete (2025-10-23)

#### Validation Rules

**Boolean Flag:**

- No value required (presence = enabled)
- No validation errors possible

**Priority Behavior:**

- Takes absolute priority over all other flags
- Displays help and exits immediately
- Exit code 0 (success)

#### Behavior

**Basic Usage:**

```bash
snag --help
snag -h
```

- Displays comprehensive help message
- Shows all available flags and usage
- Exits with code 0
- No other operations performed

**Help Message Contents:**

- Tool description and purpose
- Usage syntax
- All available flags with descriptions
- Examples
- Exit codes
- Links to documentation

#### Interaction Matrix

**With All Other Flags:**

| Combination          | Behavior             | Rationale                                 |
| -------------------- | -------------------- | ----------------------------------------- |
| `--help` alone       | Display help, exit 0 | Standard help mode                        |
| `--help <url>`       | Display help, exit 0 | Help takes priority, URL ignored          |
| `--help --version`   | Display help, exit 0 | Help takes priority over version          |
| `--version --help`   | Display help, exit 0 | Help takes priority (regardless of order) |
| `--help --doctor`    | Display help, exit 0 | Help takes priority over doctor           |
| `--doctor --help`    | Display help, exit 0 | Help takes priority (regardless of order) |
| `--help` + any flags | Display help, exit 0 | Help ignores all other flags              |

**Priority Rules:**

1. `--help` detected → Display help
2. Ignore all other flags completely
3. Exit with code 0
4. `--help` takes priority over `--version`, `--doctor`, `--kill-browser`, and skill verbs

#### Examples

**Valid (All display help and exit):**

```bash
snag --help                                         # Basic help
snag -h                                             # Short form
snag --help https://example.com                     # Help (URL ignored)
snag --help --version                               # Help (version ignored)
snag --help -o file.md --format pdf --verbose       # Help (everything ignored)
```

**No Invalid Combinations:**

- Help flag ignores all other input
- Always succeeds (exit 0)

#### Implementation Details

**Location:**

- Flag definition: Built into `github.com/spf13/cobra` framework
- Auto-generated help by CLI framework

**Processing:**

1. CLI framework checks for `--help` or `-h`
2. If present, displays auto-generated help
3. Exits with code 0
4. No custom validation code runs

**Help Display:**

- Auto-generated from flag definitions
- Includes usage patterns
- Shows all flags with descriptions
- Framework handles everything

---
