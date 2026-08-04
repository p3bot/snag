# `--tab PATTERN` / `-t`

**Status:** Complete (2025-10-23)

#### Validation Rules

**Pattern Types:**

- Integer: Tab index (1-based, e.g., "1" = first tab)
- Range: Tab index range (1-based, e.g., "1-3" = first three tabs)
- String: URL pattern for matching (exact/substring/regex)
- Empty/whitespace: **Error + list tabs**

**Multiple Flags:**

- Multiple `--tab` flags → **Last wins** (standard CLI behavior, no error, no warning)

**Pattern Matching Priority (Progressive Fallthrough):**

1. **Range** → Tab index range (format: `N-M` where N and M are positive integers)
2. **Integer** → Tab index (1-based)
3. **Exact match** → Case-insensitive URL match (`strings.EqualFold`)
4. **Contains match** → Case-insensitive substring (`strings.Contains`)
5. **Regex match** → Case-insensitive regex (`(?i)` flag)
6. **No match** → Error + list tabs

**Range Validation:**

- Range format: `N-M` where both N and M are positive integers >= 1
- Start must be <= end
- Both indices must exist in browser tabs
- Reverse ranges (e.g., `3-1`) → **Error**

**Error Messages:**

- Empty/whitespace pattern: `"Tab pattern cannot be empty"` + list available tabs
- No matching tabs: `"No tab matches pattern '{pattern}'"` + list available tabs
- Index out of range: `"Tab index {n} out of range (only {count} tabs open)"` + list available tabs
- Invalid regex: `"Invalid regex pattern: {error}"`
- Invalid range format: `"Invalid range format: use N-M (e.g., 1-3)"`
- Invalid range: start > end: `"Invalid range: start must be <= end (got {start}-{end})"`
- Range start < 1: `"Tab range must start from 1"`
- Range exceeds tabs: `"Tab index {n} out of range in range {start}-{end} (only {count} tabs open)"`

#### Behavior

**Primary Purpose:**

- Fetch content from an existing browser tab without creating a new one
- Requires existing browser connection (won't auto-launch)
- Mutually exclusive with all other content sources

**Pattern Matching Examples:**

```bash
# By index (1-based)
snag --tab 1                                    # First tab
snag -t 3                                       # Third tab

# By range (1-based)
snag --tab 1-3                                  # First three tabs (1, 2, 3)
snag -t 4-6                                     # Tabs 4 through 6
snag -t 1-1                                     # Single tab (same as --tab 1)

# By exact URL (case-insensitive)
snag -t "https://github.com/p3bot/snag"  # Exact match
snag -t "EXAMPLE.COM"                           # Case-insensitive

# By substring/contains
snag -t "github"                                # Contains "github"
snag -t "dashboard"                             # Contains "dashboard"

# By regex pattern
snag -t "https://.*\.com"                       # Regex: https:// + anything + .com
snag -t ".*/dashboard"                          # Regex: any URL ending with /dashboard
snag -t "(github|gitlab)\.com"                  # Regex: github.com or gitlab.com
```

**Range Behavior:**

- Ranges fetch multiple tabs (like `--all-tabs` but limited to range)
- Auto-saves to current directory with generated filenames (never outputs to stdout)
- Cannot use `-o` flag (use `-d` for custom directory)
- Processes tabs sequentially in order (1→2→3)
- Fails fast: stops at first tab that doesn't exist

**Multiple Matches:**

- **Single match**: Fetch and output to stdout (or to file with `-o`)
- **Multiple matches**: Fetch all matching tabs, auto-save with generated filenames (like `--all-tabs`)
  - No confirmation prompt - auto-proceed with good logging
  - Process in same sort order as `--list-tabs` (alphabetically by URL)
  - Cannot use `--output` flag (error), use `--output-dir` instead

**No Browser Connection:**

- Error: `"No browser found. Try running 'snag --open-browser' first"`
- No special error handling (same as other browser operations)

#### Interaction Matrix

**Content Source Conflicts (All ERROR - Mutually Exclusive):**

| Combination                  | Behavior                | Error Message                                                                    |
| ---------------------------- | ----------------------- | -------------------------------------------------------------------------------- |
| `--tab` + `<url>` (single)   | **Error**               | `"Cannot use both --tab and URL arguments (mutually exclusive content sources)"` |
| `--tab` + `<url>` (multiple) | **Error**               | `"Cannot use both --tab and URL arguments (mutually exclusive content sources)"` |
| `--tab` + `--url-file`       | **Error**               | `"Cannot use both --tab and --url-file (mutually exclusive content sources)"`    |
| `--tab` + `--all-tabs`       | **Error**               | `"Cannot use both --tab and --all-tabs (mutually exclusive content sources)"`    |
| `--tab` + `--list-tabs`      | `--list-tabs` overrides | `--list-tabs` overrides all other options (no error)                             |
| `--tab` + `--doctor`         | **Flag ignored**        | Doctor overrides, diagnostics only                                               |

**Browser Mode Conflicts (All ERROR):**

| Combination                  | Behavior                  | Error Message                                                                           |
| ---------------------------- | ------------------------- | --------------------------------------------------------------------------------------- |
| `--tab` + `--force-headless` | **Error**                 | `"Cannot use --force-headless with --tab (--tab requires existing browser connection)"` |
| `--tab` + `--kill-browser`   | **Error**                 | `"Cannot use --kill-browser with --tab (conflicting operations)"`                       |
| `--tab` + `--open-browser`   | **Warning**, flag ignored | `"Warning: --tab ignored with --open-browser (no content fetching)"`                    |

**Output Control:**

| Combination                                | Behavior       | Notes                                                                |
| ------------------------------------------ | -------------- | -------------------------------------------------------------------- |
| `--tab N` (single match) + `--output`      | Works normally | Fetch from tab, save to file                                         |
| `--tab N-M` (range) + `--output`           | **Error**      | `"Cannot use --output with multiple tabs. Use --output-dir instead"` |
| `--tab PATTERN` (multi-match) + `--output` | **Error**      | `"Cannot use --output with multiple tabs. Use --output-dir instead"` |
| `--tab` + `--output-dir`                   | Works normally | Fetch from tab(s), auto-generate filename(s)                         |
| `--tab` + `--format` (all)                 | Works normally | All formats supported (md/html/text/pdf/png)                         |
| `--tab N` (single match) + no output       | Works normally | Output to stdout                                                     |
| `--tab N-M` (range) + no output            | Auto-save      | Auto-generates filenames in current directory                        |
| `--tab PATTERN` (multi-match) + no output  | Auto-save      | Auto-generates filenames in current directory                        |

**Timing & Selector:**

| Combination                             | Behavior               | Notes                                                                 |
| --------------------------------------- | ---------------------- | --------------------------------------------------------------------- |
| `--tab` + `--timeout` (no `--wait-for`) | Works with **warning** | `"Warning: --timeout is ignored without --wait-for when using --tab"` |
| `--tab` + `--timeout` + `--wait-for`    | Works normally         | Timeout applies to selector wait                                      |
| `--tab` + `--wait-for`                  | Works normally         | Wait for selector in existing tab (automation use case)               |

**Browser Configuration:**

| Combination                 | Behavior             | Notes                                                                                     |
| --------------------------- | -------------------- | ----------------------------------------------------------------------------------------- |
| `--tab` + `--port`          | Works normally       | Connect to browser on specific port                                                       |
| `--tab` + `--close-tab`     | Works normally       | Close tab after fetching (covered in Task 9)                                              |
| `--tab` + `--user-agent`    | **Warning**, ignored | `"Warning: --user-agent is ignored with --tab (cannot change existing tab's user agent)"` |
| `--tab` + `--user-data-dir` | **Warning**, ignored | `"Warning: --user-data-dir ignored when connecting to existing browser"`                  |

**Logging Flags (All Work Normally):**

| Combination           | Behavior                                                    |
| --------------------- | ----------------------------------------------------------- |
| `--tab` + `--verbose` | Works normally - verbose logging of tab selection and fetch |
| `--tab` + `--quiet`   | Works normally - suppress non-error messages                |
| `--tab` + `--debug`   | Works normally - debug logging of tab operations            |

#### Examples

**Valid:**

```bash
# Fetch from tab by index
snag --tab 1                                    # First tab
snag -t 2 -o output.md                          # Second tab, save to file

# Fetch from tab range
snag --tab 1-3                                  # First three tabs (auto-save to current dir)
snag -t 4-6 -d ./output/                        # Tabs 4-6, save to specific directory
snag -t 1-5 --format pdf                        # Tabs 1-5 as PDFs (auto-save)
snag -t 2-2                                     # Single tab using range syntax

# Fetch by URL pattern (single match - outputs to stdout)
snag -t "github.com" --format html              # Match substring (if only one github tab)
snag -t "https://example.com" -o page.md        # Exact URL match (single tab)

# Fetch by URL pattern (multiple matches - auto-saves all)
snag -t "github"                                # Fetches all tabs containing "github"
snag -t "github" -d ./repos/                    # Saves all github tabs to directory
snag -t "https://.*\.com" --format pdf          # All .com tabs as PDFs

# With selectors and timeouts (automation)
snag -t "dashboard" --wait-for ".loaded"        # Wait for selector in existing tab
snag -t 1 --wait-for ".content" --timeout 60    # Custom timeout for selector

# With output options
snag -t "docs" -o reference.md                  # Save to specific file
snag -t 2 --format pdf                          # Save as PDF (auto-generates filename)
snag -t 1 -d ./output/                          # Auto-generate filename in directory

# Close tab after fetching
snag -t 1 --close-tab                           # Fetch and close tab
```

**Invalid (Errors):**

```bash
snag --tab ""                                   # ERROR: Empty pattern + list tabs
snag --tab "   "                                # ERROR: Whitespace only + list tabs
snag --tab 99                                   # ERROR: Index out of range + list tabs
snag --tab "nonexistent"                        # ERROR: No match + list tabs
snag --tab "([invalid"                          # ERROR: Invalid regex
snag --tab 1 --tab 2                            # Uses tab 2 (last wins)
snag --tab 1 https://example.com                # ERROR: Mutually exclusive with URL
snag --tab "github" -o output.md                # ERROR: Cannot use --output with multiple matches (if multiple github tabs)
snag --tab 1 --url-file urls.txt                # ERROR: Mutually exclusive with --url-file
snag --tab 1 --all-tabs                         # ERROR: Mutually exclusive with --all-tabs
snag --tab 1 --force-headless                   # ERROR: Tab requires existing browser

# Range errors
snag --tab 3-1                                  # ERROR: Invalid range (start > end)
snag --tab 0-3                                  # ERROR: Tab range must start from 1
snag --tab 1-100                                # ERROR: Tab index out of range (fails at first missing tab)
snag --tab 1-                                   # ERROR: Invalid range format
snag --tab -3                                   # ERROR: Invalid range format
snag --tab 1-3 -o output.md                     # ERROR: Cannot use --output with multiple tabs
```

**With Warnings:**

```bash
snag --tab 1 --open-browser                     # ⚠️  Warning: --tab ignored (no content fetching)
snag --tab 1 --timeout 30                       # ⚠️  Warning: --timeout is ignored without --wait-for when using --tab
snag --tab 1 --user-agent "Bot/1.0"             # ⚠️  Warning: --user-agent is ignored with --tab (cannot change existing tab's user agent)
```

**Overridden (No Error):**

```bash
snag --list-tabs --tab 1                        # --list-tabs overrides, lists all tabs
```

#### Implementation Details

**Location:**

- Flag definition: `main.go:init()`
- Handler: `handlers.go:handleTabFetch()`
- Tab selection: `browser.go:GetTabByIndex()`, `browser.go:GetTabByPattern()`
- Pattern matching: Progressive fallthrough in `GetTabByPattern()`

**How it works:**

1. Validate pattern is not empty/whitespace
2. Check for mutually exclusive flags (URL, --url-file, --all-tabs, --open-browser, --force-headless)
3. Connect to existing browser (error if none found)
4. Parse pattern:
   - If range (N-M): Validate range, fetch tabs sequentially, handle like multiple tabs
   - If integer: Convert to tab index (1-based → 0-based)
   - If string: Try exact match → substring → regex (collect ALL matches)
5. If no match: Error and list available tabs
6. Determine single vs. multiple matches:
   - Single match: Fetch and output to stdout or file
   - Multiple matches: Batch process all matches (auto-save with generated filenames)
7. Fetch content from selected tab(s)
8. Apply output options (--format, --output, --output-dir)
   - Single tab: Can use `-o` or stdout
   - Multiple tabs (range/multi-match): Must auto-save, error if `-o` specified
9. Close tab(s) if `--close-tab` is set

**Pattern Matching Algorithm:**

```go
// Pseudocode
func GetTabsByPattern(pattern string) ([]*Tab, error) {
    // 1. Try integer (tab index)
    if isInteger(pattern) {
        tab := GetTabByIndex(index)
        return []*Tab{tab}, nil
    }

    // 2. Cache page.Info() for all tabs (single pass, optimization)
    tabInfos := cacheAllTabInfo()
    var matches []*Tab

    // 3. Try exact match (case-insensitive) - collect ALL matches
    for tab, info := range tabInfos {
        if strings.EqualFold(info.URL, pattern) {
            matches = append(matches, tab)
        }
    }
    if len(matches) > 0 {
        return matches, nil
    }

    // 4. Try substring/contains match (case-insensitive) - collect ALL matches
    for tab, info := range tabInfos {
        if strings.Contains(strings.ToLower(info.URL), strings.ToLower(pattern)) {
            matches = append(matches, tab)
        }
    }
    if len(matches) > 0 {
        return matches, nil
    }

    // 5. Try regex match (case-insensitive) - collect ALL matches
    regex := regexp.MustCompile("(?i)" + pattern)
    for tab, info := range tabInfos {
        if regex.MatchString(info.URL) {
            matches = append(matches, tab)
        }
    }
    if len(matches) > 0 {
        return matches, nil
    }

    // 6. No match found
    return nil, ErrNoTabMatch
}
```

**Performance Optimization:**

- Single-pass `page.Info()` caching in `browser.go:GetTabsByPattern()`
- Reduces network calls from 3N to N (3x improvement for 10 tabs)
- Do not modify pattern matching without preserving this optimization

**Design Note:**

- User-facing indexes are 1-based (natural for humans)
- Internal indexes are 0-based (natural for Go)
- Conversion happens in `TabInfo` struct and `GetTabByIndex()`
