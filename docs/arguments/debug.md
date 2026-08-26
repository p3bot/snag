# `--debug`

**Status:** Complete (2025-10-23)

#### Validation Rules

**Boolean Flag:**

- No value required (presence = enabled)
- No validation errors possible

**Multiple Flag Conflicts:**

- Multiple `--debug` flags → Last flag honored (standard behavior)
- `--debug` + `--verbose` → Error: mutually exclusive
- `--debug` + `--quiet` → Error: mutually exclusive
- Only one logging level flag can be used at a time

#### Behavior

**Logging Level:**

- Enables maximum logging output to stderr
- Shows all verbose information plus:
  - Chrome DevTools Protocol (CDP) messages
  - Browser connection debugging
  - Internal state information
  - Detailed error traces
- For troubleshooting and development

**Basic Usage:**

```bash
snag https://example.com --debug
```

- Outputs page content to stdout (as normal)
- Logs extensive debug information to stderr
- Includes CDP protocol messages

#### Interaction Matrix

**Logging Level Flags (Mutually Exclusive):**

| Combination                 | Result | Error Message                                                                                                           |
| --------------------------- | ------ | ----------------------------------------------------------------------------------------------------------------------- |
| `--debug`                   | Debug  | (Valid - debug mode)                                                                                                    |
| `--debug --verbose`         | Error  | `if any flags in the group [quiet verbose debug] are set none of the others can be; [debug verbose] were all set`       |
| `--verbose --debug`         | Error  | `if any flags in the group [quiet verbose debug] are set none of the others can be; [debug verbose] were all set`       |
| `--debug --quiet`           | Error  | `if any flags in the group [quiet verbose debug] are set none of the others can be; [debug quiet] were all set`         |
| `--quiet --debug`           | Error  | `if any flags in the group [quiet verbose debug] are set none of the others can be; [debug quiet] were all set`         |
| `--debug --quiet --verbose` | Error  | `if any flags in the group [quiet verbose debug] are set none of the others can be; [debug quiet verbose] were all set` |

**All Other Flags:**

- Works normally with all other flags
- Simply controls stderr logging verbosity
- No conflicts or special behaviors

**Examples with Other Flags:**

- `--debug` + `--doctor` - Works normally (debug logs during diagnostic operations)
- `--debug` + `--kill-browser` - Works normally (debug logs showing commands executed)
- `--debug` + `--user-data-dir` - Works normally (debug logs with custom profile)
- `--debug` + `--user-agent` - Works normally (debug logs with custom UA)
- `--debug` + all browser/output/timing flags - Works normally

#### Examples

**Valid:**

```bash
snag https://example.com --debug                    # Debug logging
snag https://example.com --debug -o page.md         # Debug + file output
snag --url-file urls.txt --debug                    # Debug batch processing
snag --tab 1 --debug                                # Debug tab fetch
snag --list-tabs --debug                            # Debug tab listing
```

**Invalid (Mutually Exclusive):**

```bash
snag https://example.com --debug --verbose          # Error: mutually exclusive
snag https://example.com --debug --quiet            # Error: mutually exclusive
snag https://example.com --debug --verbose --quiet  # Error: mutually exclusive
```

#### Implementation Details

**Location:**

- Flag definition: `internal/cli/root.go` (`init`)
- Logger initialization: `internal/cli/root.go` (`runCobra`)
- Logging level: `internal/logger`

**Processing:**

1. Cobra validates that only one logging flag is present (mutually exclusive)
2. Check if `--debug` flag is set
3. Initialize logger with debug level
4. All subsequent operations use debug logging
5. CDP messages logged via rod's debug capabilities

**Logging Behavior:**

- Debug output: Everything (verbose + CDP messages + internals)
- Extremely detailed for troubleshooting
- Logs go to stderr (stdout reserved for content)

---
