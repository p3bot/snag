# `--verbose`

**Status:** Complete (2025-10-23)

#### Validation Rules

**Boolean Flag:**

- No value required (presence = enabled)
- No validation errors possible

**Multiple Flag Conflicts:**

- Multiple `--verbose` flags → Last flag honored (standard behavior)
- `--verbose` + `--debug` → Error: mutually exclusive
- Only one logging level flag can be used at a time

#### Behavior

**Logging Level:**

- Enables verbose logging output to stderr
- Restores fetch progress that is silent by default:
  - Browser launched/connected
  - Fetching URL / fetched successfully
  - Batch counters
  - Navigation, conversion, and other operational detail
- Does not affect stdout content output

**Basic Usage:**

```bash
snag https://example.com --verbose
```

- Outputs page content to stdout (as normal)
- Logs verbose messages to stderr, including:
  - "Chrome launched in headless mode" (or connected to an existing browser)
  - "Fetching https://example.com..."
  - "Navigating to https://example.com (timeout: 30s)..."
  - "Fetched successfully"

**Tab Listing Behavior:**

When used with `--list-tabs`, verbose mode changes the output format:

```bash
snag --list-tabs --verbose
```

- **Normal mode**: Shows clean URLs without query parameters: `[N] URL (Title)`
- **Verbose mode**: Shows full URLs with all query parameters and hash fragments: `[N] full-url - Title`
- No truncation in verbose mode (shows complete URLs and titles)

#### Interaction Matrix

**Logging Level Flags (Mutually Exclusive):**

| Combination         | Result  | Error Message                                                                                                     |
| ------------------- | ------- | ----------------------------------------------------------------------------------------------------------------- |
| `--verbose`         | Verbose | (Valid - verbose mode)                                                                                            |
| `--verbose --debug` | Error   | `if any flags in the group [verbose debug] are set none of the others can be; [debug verbose] were all set`       |
| `--debug --verbose` | Error   | `if any flags in the group [verbose debug] are set none of the others can be; [debug verbose] were all set`       |

**All Other Flags:**

- Works normally with all other flags
- Simply controls stderr logging verbosity
- No conflicts or special behaviors

**Examples with Other Flags:**

- `--verbose` + `--doctor` - Works normally (verbose logs during diagnostic operations)
- `--verbose` + `--kill-browser` - Works normally (verbose logs showing PIDs and process details)
- `--verbose` + `--user-data-dir` - Works normally (verbose logs with custom profile)
- `--verbose` + `--user-agent` - Works normally (verbose logs with custom UA)
- `--verbose` + all browser/output/timing flags - Works normally

#### Examples

**Valid:**

```bash
snag https://example.com --verbose                  # Verbose logging
snag https://example.com --verbose -o page.md       # Verbose + file output
snag --url-file urls.txt --verbose                  # Verbose batch processing
snag --tab 1 --verbose                              # Verbose tab fetch
snag --list-tabs --verbose                          # Verbose tab listing
```

**Invalid (Mutually Exclusive):**

```bash
snag https://example.com --verbose --debug          # Error: mutually exclusive
```

#### Implementation Details

**Location:**

- Flag definition: `internal/cli/root.go` (`init`)
- Logger initialization: `internal/cli/root.go` (`runCobra`)
- Logging level: `internal/logger`

**Processing:**

1. Cobra validates that `--verbose` and `--debug` are not both present (mutually exclusive)
2. Check if `--verbose` flag is set
3. Initialize logger with verbose level
4. All subsequent operations use verbose logging

**Logging Behavior:**

- Default: warnings, errors, and result announcements; connection, fetch, and batch progress is silent
- Verbose: those default lines plus connection, fetch, and batch progress, and operational detail
- Logs go to stderr (stdout reserved for content)

---
