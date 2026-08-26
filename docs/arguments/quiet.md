# `--quiet` / `-q`

**Status:** Complete (2025-10-23)

#### Validation Rules

**Boolean Flag:**

- No value required (presence = enabled)
- No validation errors possible

**Multiple Flag Conflicts:**

- Multiple `--quiet` flags → Last flag honored (standard behavior)
- `--quiet` + `--verbose` → Error: mutually exclusive
- `--quiet` + `--debug` → Error: mutually exclusive
- Only one logging level flag can be used at a time

#### Behavior

**Logging Level:**

- Suppresses all logging output to stderr except errors
- Shows only:
  - Fatal errors
  - Critical validation failures
  - Operation completion status (success/failure)
- Ideal for scripting and automation

**Basic Usage:**

```bash
snag https://example.com --quiet
```

- Outputs page content to stdout (as normal)
- Suppresses all stderr messages except errors
- Silent operation on success

#### Interaction Matrix

**Logging Level Flags (Mutually Exclusive):**

| Combination                 | Result | Error Message                                                                                                           |
| --------------------------- | ------ | ----------------------------------------------------------------------------------------------------------------------- |
| `--quiet`                   | Quiet  | (Valid - quiet mode)                                                                                                    |
| `--quiet --verbose`         | Error  | `if any flags in the group [quiet verbose debug] are set none of the others can be; [quiet verbose] were all set`       |
| `--verbose --quiet`         | Error  | `if any flags in the group [quiet verbose debug] are set none of the others can be; [quiet verbose] were all set`       |
| `--quiet --debug`           | Error  | `if any flags in the group [quiet verbose debug] are set none of the others can be; [debug quiet] were all set`         |
| `--debug --quiet`           | Error  | `if any flags in the group [quiet verbose debug] are set none of the others can be; [debug quiet] were all set`         |
| `--quiet --verbose --debug` | Error  | `if any flags in the group [quiet verbose debug] are set none of the others can be; [debug quiet verbose] were all set` |

**All Other Flags:**

- Works normally with all other flags
- Simply controls stderr logging verbosity
- No conflicts or special behaviors

**Examples with Other Flags:**

- `--quiet` + `--doctor` - Works normally (minimal logging during diagnostic operations)
- `--quiet` + `--kill-browser` - Works normally (silent on success, errors to stderr)
- `--quiet` + `--user-data-dir` - Works normally (quiet mode with custom profile)
- `--quiet` + `--user-agent` - Works normally (quiet mode with custom UA)
- `--quiet` + all browser/output/timing flags - Works normally

#### Examples

**Valid:**

```bash
snag https://example.com --quiet                    # Silent operation
snag https://example.com -q -o page.md              # Silent file save
snag --url-file urls.txt --quiet                    # Silent batch processing
snag --tab 1 --quiet                                # Silent tab fetch
snag --list-tabs --quiet                            # Silent tab listing (shows tabs only)
```

**Invalid (Mutually Exclusive):**

```bash
snag https://example.com --quiet --verbose          # Error: mutually exclusive
snag https://example.com --quiet --debug            # Error: mutually exclusive
snag https://example.com -q --verbose --debug       # Error: mutually exclusive
```

#### Implementation Details

**Location:**

- Flag definition: `internal/cli/root.go` (`init`)
- Logger initialization: `internal/cli/root.go` (`runCobra`)
- Logging level: `internal/logger`

**Processing:**

1. Cobra validates that only one logging flag is present (mutually exclusive)
2. Check if `--quiet` flag is set
3. Initialize logger with quiet level
4. All subsequent operations suppress non-error logs

**Logging Behavior:**

- Quiet output: Errors only
- Exit code 0 on success (silent)
- Exit code 1 on failure (error message shown)

---
