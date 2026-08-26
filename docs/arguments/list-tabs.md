# `--list-tabs` / `-l`

**Status:** Complete (2025-10-23)

#### Validation Rules

**Boolean Flag:**

- No value required (presence = enabled)
- No validation errors possible

**Multiple Flags:**

- Multiple `--list-tabs` flags → **Silently ignored** (duplicate boolean)

#### Behavior

**Primary Purpose:**

- List all open tabs in an existing browser connection
- Acts like `--help` or `--version`: Overrides all other flags except those needed for its operation
- Lists tabs and exits snag immediately

**Core Mode:**

```bash
snag --list-tabs
snag --list-tabs --port 9223
snag --list-tabs --verbose
```

- Connects to existing browser (or errors if none found)
- Displays tab list to stdout
- Exits snag immediately
- Ignores all other flags except `--port` and logging flags

**No Browser Connection:**

- Error: Connection error with helpful hint
- Message: `"No browser found. Try running 'snag --open-browser' first"`

**Precedence Order:**

1. `--help` (highest priority, overrides everything)
2. `--version` (overrides everything below)
3. `--list-tabs` (overrides everything below)
4. All other flags (ignored when `--list-tabs` is present)

#### Interaction Matrix

**Flags That Work WITH `--list-tabs`:**

| Combination                 | Behavior       | Notes                                        |
| --------------------------- | -------------- | -------------------------------------------- |
| `--list-tabs` + `--port`    | Works normally | Specify which browser instance to connect to |
| `--list-tabs` + `--verbose` | Works normally | Verbose logging during tab listing           |
| `--list-tabs` + `--quiet`   | Works normally | Quiet mode (minimal output)                  |
| `--list-tabs` + `--debug`   | Works normally | Debug/CDP logging during tab listing         |

**Higher Priority Flags (Override):**

| Combination                      | Behavior               | Notes                                                                       |
| -------------------------------- | ---------------------- | --------------------------------------------------------------------------- |
| `--list-tabs` + `--doctor`       | `--doctor` overrides   | Doctor has higher priority                                                  |
| `--list-tabs` + `--kill-browser` | **Error**              | `"Cannot use --kill-browser with --list-tabs (conflicting operations)"` |

**All Other Flags Are SILENTLY IGNORED:**

`--list-tabs` acts like `--help` and overrides all other arguments:

| Combination                        | Behavior                  | Notes                                |
| ---------------------------------- | ------------------------- | ------------------------------------ |
| `--list-tabs` + `<url>`            | URL ignored, tabs listed  | Lists tabs, ignores URL              |
| `--list-tabs` + `--url-file`       | Flag ignored, tabs listed | Lists tabs, ignores file             |
| `--list-tabs` + `--output`         | Flag ignored, tabs listed | Lists tabs to stdout only            |
| `--list-tabs` + `--output-dir`     | Flag ignored, tabs listed | Lists tabs to stdout only            |
| `--list-tabs` + `--format`         | Flag ignored, tabs listed | Lists tabs in fixed format           |
| `--list-tabs` + `--timeout`        | Flag ignored, tabs listed | Lists tabs (no navigation)           |
| `--list-tabs` + `--wait-for`       | Flag ignored, tabs listed | Lists tabs (no content fetch)        |
| `--list-tabs` + `--close-tab`      | Flag ignored, tabs listed | Lists tabs (no tab closing)          |
| `--list-tabs` + `--force-headless` | Flag ignored, tabs listed | Lists tabs from existing browser     |
| `--list-tabs` + `--open-browser`   | Flag ignored, tabs listed | Lists tabs (no browser launch)       |
| `--list-tabs` + `--tab`            | Flag ignored, tabs listed | Lists all tabs (no single tab fetch) |
| `--list-tabs` + `--all-tabs`       | Flag ignored, tabs listed | Lists tabs (no fetching)             |
| `--list-tabs` + `--user-agent`     | Flag ignored, tabs listed | Lists tabs (no navigation)           |
| `--list-tabs` + `--user-data-dir`  | Flag ignored, tabs listed | Connects to existing browser         |

**Rationale:**

- `--list-tabs` is a simple informational command like `--help` or `--version`
- Users expect it to "just work" without complex argument validation
- Simplifies UX: No need to remember which flags conflict with `--list-tabs`
- Other tools follow this pattern (e.g., `git --version <any-args>` ignores all args)

#### Examples

**Valid:**

```bash
# Basic tab listing
snag --list-tabs

# List tabs on custom port
snag --list-tabs --port 9223

# List tabs with verbose logging
snag --list-tabs --verbose

# List tabs quietly (minimal output)
snag --list-tabs --quiet

# List tabs with debug logging
snag --list-tabs --debug
```

**Silently Ignores Other Flags:**

```bash
# All other flags are ignored, tabs are listed
snag --list-tabs https://example.com
snag --list-tabs --output file.md
snag --list-tabs --format pdf --wait-for ".content"
snag --list-tabs --tab 1 --close-tab
snag --list-tabs --force-headless --user-agent "Bot/1.0"
```

**Error Cases:**

```bash
# No browser running
snag --list-tabs
# → Error: "No browser found. Try running 'snag --open-browser' first"
```

#### Implementation Details

**Location:**

- Flag definition: `internal/cli/root.go` (`init`)
- Handler: `internal/cli/handlers.go` (`handleListTabs`)
- Browser connection: `internal/browser` (`ListTabs`)

**How it works:**

1. Check if `--list-tabs` is set (should be checked early, like `--help` and `--version`)
2. Extract `--port` and logging flags (`--verbose`, `--quiet`, `--debug`)
3. Silently ignore all other flags (no warnings, no errors)
4. Connect to existing browser on specified port
5. If no browser found: Error with helpful hint
6. List tabs to stdout (index, URL, title)
7. Exit snag immediately

**Output Format:**

**Normal mode (default):**

Clean, scannable format with URL first and title in parentheses:

```
Available tabs in browser (7 tabs, sorted by URL):
  [1] chrome://newtab (New Tab)
  [2] https://example.com (Example Domain)
  [3] https://github.com/puppeteer/puppeteer/issues/7452 (Order in browser.pages() · Issue #7452)
  [4] https://ato.gov.au/about-ato/contact-us (Contact us | Australian Taxation Office)
  [5] https://bigw.com.au (BIG W | How good's that)
  [6] https://google.com/search (youtube - Google Search)
  [7] https://x.com (X. It's what's happening / X)
```

**Format Rules:**

- **Pattern**: `[N] URL (Title)`
- **URL first**: More distinctive and scannable for finding specific sites
- **Clean URLs**: Query parameters (`?...`) and hash fragments (`#...`) stripped from display
- **Truncation**: Total line length limited to 120 characters
  - Layout: `  [NNN] URL (Title)` = ~8 chars prefix + URL + title
  - URL limit: Maximum 80 characters
  - Title space: Remainder (typically 30-70 chars depending on URL length)
  - Truncation indicator: `...` added when content is truncated
- **Empty titles**: Omitted entirely (shows `[N] URL` without empty parentheses)

**Verbose mode:**

Full URLs with query parameters and hash fragments (no truncation):

```bash
snag --list-tabs --verbose
```

```
Available tabs in browser (2 tabs, sorted by URL):
  [1] https://www.ato.gov.au/about-ato/contact-us?gclsrc=aw.ds&gad_source=1&gad_campaignid=22717615122&gbraid=0AAAAAolERwRiHRb8KrrNHk7GbTWG6FA0E&gclid=EAIaIQobChMImPuotYK-kAMVJ6hmAh10QhjZEAAYASAAEgKG6_D_BwE - Contact us | Australian Taxation Office
  [2] https://www.google.com/search?gs_ssp=eJzj4tTP1TewzEouKzZg9GKvzC8tKU1KBQA_-AaN&q=youtube&oq=tyou&gs_lcrp=EgZjaHJvbWUqDwgBEC4YChiDARixAxiABDIGCAAQRRg5Mg8IARAuGAoYgwEYsQMYgAQyDwgCEAAYChiDARixAxiABDIPCAMQABgKGIMBGLEDGIAEMhUIBBAuGAoYgwEYxwEYsQMY0QMYgAQyDwgFEAAYChiDARixAxiABDIMCAYQABgKGLEDGIAEMgYIBxAFGEDSAQgyMzYxajBqN6gCB7ACAfEF13V1Q3FsVZ4&sourceid=chrome&ie=UTF-8&sei=ow_8aIvBO4uMseMPmrudyAk - youtube - Google Search
```

**Tab Ordering:**

- **Tabs are sorted by URL (primary), Title (secondary), ID (tertiary)** for predictable, stable ordering
- Chrome DevTools Protocol returns tabs in unpredictable internal order (not visual left-to-right browser order)
- Sorting ensures consistent tab indices across all operations (`--list-tabs`, `--tab`, `--tab <range>`, `--all-tabs`)
- Tab [1] = first tab alphabetically by URL, **not** first visual tab in browser window
- This design provides reproducible automation workflows and predictable tab selection

**Error Messages:**

- No browser: `"No browser found. Try running 'snag --open-browser' first"`
- Connection error: Standard browser connection error messages

**Design Note:**

- Output goes to stdout (not stderr) to enable piping: `snag --list-tabs | grep github`

---
