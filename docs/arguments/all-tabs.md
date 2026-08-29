# `--all-tabs` / `-a`

**Status:** Complete (2025-10-23)

#### Validation Rules

**Flag Type:**

- Boolean flag (no value required)
- Multiple `--all-tabs` flags → Silently ignored (duplicate boolean)

**Browser Connection:**

- Requires existing browser connection with remote debugging enabled
- Connection failure → Error: `"No browser found. Try running 'snag --open-browser' first"`
- Zero tabs open → Error (browser cannot have zero tabs)

**Output Requirements:**

- No output flags (neither `-o` nor `-d`) → Default to current directory (`.`)
- Auto-generates filename for each tab
- Cannot use `-o` (single file output) with `--all-tabs` (multiple outputs)

**Error Messages:**

- No existing browser: `"No browser found. Try running 'snag --open-browser' first"`
- Zero tabs (if encountered): `"No tabs found in browser"`

#### Behavior

**Primary Purpose:**

- Fetch content from ALL open browser tabs
- Process tabs sequentially in browser tab order (left to right)
- Save each tab to separate file with auto-generated filename
- Continue processing all tabs even if individual tabs fail

**Tab Processing:**

- **Order**: Browser's tab order (same as `--list-tabs` display order)
- **Error handling**: Continue on error (process all tabs)
- **Exit code**: 0 if ALL succeed, 1 if ANY fail
- **Summary**: `"Batch complete: X succeeded, Y failed"` (verbose only)
- **Progress logging**: `[n/N] Processing: URL` is verbose; skip warnings, errors, and `Saved to` still print at default

**Special Tab Handling:**

- Non-fetchable URLs (`chrome://`, `about:`, `devtools://`, etc.):
  - Skip with warning: `"Skipping tab: {URL} (not fetchable)"`
  - Does not count as failure
  - Continue to next tab

**Output Behavior:**

- Always saves to files (never stdout)
- No `-d` flag → Auto-generated filenames in current directory (`.`)
- With `-d ./dir` → Auto-generated filenames in specified directory
- Filename format: `yyyy-mm-dd-hhmmss-{page-title}-{slug}.{ext}`
- Filename conflicts: Append `-1`, `-2`, etc. (existing conflict resolution)

**Close-Tab Behavior:**

- With `--close-tab`: Close each tab immediately after fetching
- Last tab closure → Browser closes automatically
- Message: `"Closing last tab, browser will close"`

#### Interaction Matrix

**Content Source Conflicts (All ERROR - Mutually Exclusive):**

| Combination                       | Behavior                | Error Message                                                                         |
| --------------------------------- | ----------------------- | ------------------------------------------------------------------------------------- |
| `--all-tabs` + `<url>` (single)   | **Error**               | `"Cannot use both --all-tabs and URL arguments (mutually exclusive content sources)"` |
| `--all-tabs` + `<url>` (multiple) | **Error**               | `"Cannot use both --all-tabs and URL arguments (mutually exclusive content sources)"` |
| `--all-tabs` + `--url-file`       | **Error**               | `"Cannot use both --all-tabs and --url-file (mutually exclusive content sources)"`    |
| `--all-tabs` + `--tab`            | **Error**               | `"Cannot use both --tab and --all-tabs (mutually exclusive content sources)"`         |
| `--all-tabs` + `--list-tabs`      | `--list-tabs` overrides | `--list-tabs` overrides all other options (no error)                                  |
| `--all-tabs` + `--doctor`         | **Flag ignored**        | Doctor overrides, diagnostics only                                                    |

**Browser Mode Conflicts (All ERROR):**

| Combination                       | Behavior                  | Error Message                                                                                     |
| --------------------------------- | ------------------------- | ------------------------------------------------------------------------------------------------- |
| `--all-tabs` + `--force-headless` | **Error**                 | `"Cannot use --force-headless with --all-tabs (--all-tabs requires existing browser connection)"` |
| `--all-tabs` + `--kill-browser`   | **Error**                 | `"Cannot use --kill-browser with --all-tabs (conflicting operations)"`                            |
| `--all-tabs` + skill verb         | **Error**                 | `"Cannot use a skill flag with --all-tabs (conflicting operations)"`                              |
| `--all-tabs` + `--open-browser`   | **Warning**, flag ignored | `"Warning: --all-tabs ignored with --open-browser (no content fetching)"`                         |

**Output Control:**

| Combination                     | Behavior       | Notes                                                                           |
| ------------------------------- | -------------- | ------------------------------------------------------------------------------- |
| `--all-tabs` (no output flags)  | Works normally | Auto-save to current directory (`.`) with auto-generated filenames              |
| `--all-tabs` + `--output-dir`   | Works normally | Save all tabs to specified directory with auto-generated filenames              |
| `--all-tabs` + `--output`       | **Error**      | `"Cannot use --output with multiple content sources. Use --output-dir instead"` |
| `--all-tabs` + `--format` (all) | Works normally | All formats supported (md/html/text/pdf/png), applied to all tabs               |

**Timing & Selector:**

| Combination                                  | Behavior               | Notes                                                                      |
| -------------------------------------------- | ---------------------- | -------------------------------------------------------------------------- |
| `--all-tabs` + `--timeout` (no `--wait-for`) | Works with **warning** | `"Warning: --timeout is ignored without --wait-for when using --all-tabs"` |
| `--all-tabs` + `--timeout` + `--wait-for`    | Works normally         | Timeout applies to selector wait for each tab (30s per tab, not total)     |
| `--all-tabs` + `--wait-for`                  | Works normally         | Same selector applied to all tabs before fetching                          |

**Browser Configuration:**

| Combination                      | Behavior             | Notes                                                                                           |
| -------------------------------- | -------------------- | ----------------------------------------------------------------------------------------------- |
| `--all-tabs` + `--port`          | Works normally       | Connect to browser on specific port                                                             |
| `--all-tabs` + `--close-tab`     | Works normally       | Close each tab after fetching; last tab closes browser                                          |
| `--all-tabs` + `--user-agent`    | **Warning**, ignored | `"Warning: --user-agent is ignored with --all-tabs (cannot change existing tabs' user agents)"` |
| `--all-tabs` + `--user-data-dir` | **Warning**, ignored | `"Warning: --user-data-dir ignored when connecting to existing browser"`                        |

**Logging Flags (All Work Normally):**

| Combination                | Behavior                                           |
| -------------------------- | -------------------------------------------------- |
| `--all-tabs` + `--verbose` | Works normally - verbose logging of tab processing |
| `--all-tabs` + `--debug`   | Works normally - debug logging of all operations   |

#### Examples

**Valid:**

```bash
# Fetch all tabs to current directory
snag --all-tabs                                 # Auto-save all tabs to ./

# Fetch all tabs to specific directory
snag --all-tabs -d ./tabs                       # Save all tabs to ./tabs/
snag -a -d ./output                             # Short form

# With output format
snag --all-tabs --format html                   # All tabs as HTML
snag --all-tabs --format pdf -d ./pdfs          # All tabs as PDF

# With selectors and timeouts
snag --all-tabs --wait-for ".content"           # Wait for selector in each tab
snag --all-tabs --wait-for ".loaded" --timeout 60  # Custom timeout per tab

# Close tabs after fetching
snag --all-tabs --close-tab                     # Fetch all, close all, browser closes
snag --all-tabs -c -d ./backup                  # Backup all tabs then close

# With custom port
snag --all-tabs --port 9223 -d ./tabs           # Connect to browser on port 9223

# With logging
snag --all-tabs --verbose                       # Verbose progress logging
```

**Invalid (Errors):**

```bash
snag --all-tabs https://example.com             # ERROR: Cannot mix with URL
snag --all-tabs url1 url2                       # ERROR: Cannot mix with URLs
snag --all-tabs --url-file urls.txt             # ERROR: Cannot mix with --url-file
snag --all-tabs --tab 1                         # ERROR: Cannot mix with --tab
snag --all-tabs -o output.md                    # ERROR: Use -d for multiple outputs
snag --all-tabs --force-headless                # ERROR: Requires existing browser
```

**With Warnings:**

```bash
snag --all-tabs --open-browser                  # ⚠️  Warning: --all-tabs ignored (no content fetching)
snag --all-tabs --timeout 30                    # ⚠️  Warning: --timeout is ignored without --wait-for when using --all-tabs
snag --all-tabs --user-agent "Bot/1.0"          # ⚠️  Warning: --user-agent is ignored with --all-tabs (cannot change existing tabs' user agents)
snag --all-tabs                                 # ⚠️  Skipping tab: chrome://settings (not fetchable)
```

**Overridden (No Error):**

```bash
snag --list-tabs --all-tabs                     # --list-tabs overrides, lists all tabs
```

#### Implementation Details

**Location:**

- Flag definition: `internal/cli/root.go` (`init`)
- Handler: `internal/cli/handlers.go` (`handleAllTabs`)
- Tab listing: `internal/browser` (`ListTabs`)
- Tab fetching: Reuse existing fetch logic per tab

**How it works:**

1. Validate mutually exclusive flags (URL, --url-file, --tab, --open-browser, --force-headless)
2. Check output flags (error if `-o`, default to `.` if no `-d`)
3. Connect to existing browser (error if none found)
4. Get list of all tabs using `ListTabs()`
5. For each tab in browser order:
   - Check if URL is fetchable (skip `chrome://`, `about:`, `devtools://`, etc.)
   - Log: `"Fetching tab {index}/{total}: {URL}"`
   - Fetch content from tab (apply `--wait-for`, `--timeout`, `--format`)
   - Generate filename with timestamp and title slug
   - Check for filename conflicts (append `-1`, `-2`, etc.)
   - Save to file in output directory
   - If `--close-tab`: Close tab after saving
   - On error: Log error, increment failure count, continue to next tab
6. Log summary: `"Fetched X of Y tabs successfully"`
7. Exit 0 if all succeeded, exit 1 if any failed

**Error Handling Strategy:**

- **Continue on error**: Process all tabs even if individual tabs fail
- **Log errors**: Display error message for each failed tab as it occurs
- **Track failures**: Count total successes and failures
- **Final summary**: Display count of successful vs failed fetches
- **Exit code**: 0 only if ALL tabs succeeded, 1 if ANY tab failed

**Special URL Detection:**

```go
// Pseudocode for non-fetchable URL detection
func isNonFetchableURL(url string) bool {
    nonFetchablePrefixes := []string{
        "chrome://",
        "about:",
        "devtools://",
        "chrome-extension://",
        "edge://",
        "brave://",
    }

    for _, prefix := range nonFetchablePrefixes {
        if strings.HasPrefix(strings.ToLower(url), prefix) {
            return true
        }
    }
    return false
}
```

**Filename Generation:**

- Format: `yyyy-mm-dd-hhmmss-{page-title}-{slug}.{ext}`
- Extension based on `--format` flag
- Title slug: Lowercase, alphanumeric + hyphens, max length
- Conflict resolution: Append `-1`, `-2`, etc. if file exists
- See `design-record.md` for complete filename generation specification

**Close-Tab Behavior:**

- Close each tab immediately after successful fetch
- If tab fetch fails, still attempt to close (if `--close-tab` set)
- Last tab closure automatically closes browser
- Log message when last tab is being closed

**Performance Considerations:**

- Sequential processing (not parallel) to avoid browser resource contention
- Each tab operation is independent (failure doesn't affect others)
- Timeout applies per tab (not cumulative across all tabs)

**Design Notes:**

- Similar to `--tab` but processes all tabs instead of one
- Reuses existing fetch, format, and output logic
- Per-tab progress and the batch summary are verbose; default stderr still shows skip warnings, errors, and `Saved to`
- Continue-on-error ensures maximum data recovery even with partial failures
