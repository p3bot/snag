# `--open-browser` / `-b`

**Status:** Complete (2025-10-23)

#### Validation Rules

**Boolean Flag:**

- No value required (presence = enabled)
- No validation errors possible

**Multiple Flags:**

- Multiple `--open-browser` flags → **Silently ignored** (duplicate boolean)

#### Behavior

**Primary Purpose:**

- Launch a persistent visible browser and exit snag (browser stays open)
- Useful for setting up authenticated sessions or manual browsing
- **Does not fetch content** - purely a "launch and exit" mode

**Core Modes:**

1. **Standalone (no URLs):**

   ```bash
   snag --open-browser
   ```

   - Opens visible browser
   - Exits snag immediately
   - Browser remains open for user interaction

2. **With URLs:**
   ```bash
   snag --open-browser https://example.com
   snag --open-browser url1 url2 url3
   snag --open-browser --url-file urls.txt
   ```
   - Opens visible browser
   - Navigates to URL(s) in separate tabs
   - Exits snag immediately (no content fetching)
   - Browser and tabs remain open for user interaction

**Key Distinction:**

- "Launch and exit" means **exit snag**, not the browser
- Browser stays open after snag exits
- No content is fetched or output

**Browser Mode Conflicts:**

- With `--force-headless`: **Error** - `"Cannot use both --force-headless and --open-browser (conflicting modes)"`
- With `--kill-browser`: **Error** - `"Cannot use --kill-browser with --open-browser (conflicting operations)"`

#### Interaction Matrix

**Content Source Interactions:**

| Combination                         | Behavior                                             | Notes                                                                                               |
| ----------------------------------- | ---------------------------------------------------- | --------------------------------------------------------------------------------------------------- |
| `--open-browser` (no URLs)          | Open browser, exit snag                              | Browser stays open                                                                                  |
| `--open-browser` + single `<url>`   | Open browser, navigate to URL, exit snag             | No fetch, browser stays open                                                                        |
| `--open-browser` + multiple `<url>` | Open browser with multiple tabs, exit snag           | One tab per URL, no fetch                                                                           |
| `--open-browser` + `--url-file`     | Open browser with multiple tabs from file, exit snag | One tab per URL, no fetch                                                                           |
| `--open-browser` + `--tab`          | **Warning**                                          | `"Warning: --tab ignored with --open-browser (no content fetching)"` - Opens browser and exits      |
| `--open-browser` + `--all-tabs`     | **Warning**                                          | `"Warning: --all-tabs ignored with --open-browser (no content fetching)"` - Opens browser and exits |
| `--open-browser` + `--list-tabs`    | `--list-tabs` overrides                              | `--list-tabs` overrides all other options                                                           |
| `--open-browser` + `--doctor`       | **Flag ignored**                                     | Doctor overrides, diagnostics only                                                                  |

**Rationale:**

- `--open-browser` is purely for launching browser, not fetching content
- Tab operations (`--tab`, `--all-tabs`) imply fetching, which conflicts with open-browser's purpose

**Output Control Interactions:**

All output/format flags are **warned and ignored** because `--open-browser` doesn't fetch content:

| Combination                       | Behavior                  | Warning Message                                                             |
| --------------------------------- | ------------------------- | --------------------------------------------------------------------------- |
| `--open-browser` + `--output`     | **Warning**, flag ignored | `"Warning: --output ignored with --open-browser (no content fetching)"`     |
| `--open-browser` + `--output-dir` | **Warning**, flag ignored | `"Warning: --output-dir ignored with --open-browser (no content fetching)"` |
| `--open-browser` + `--format`     | **Warning**, flag ignored | `"Warning: --format ignored with --open-browser (no content fetching)"`     |

**Timing Interactions:**

| Combination                     | Behavior                  | Warning Message                                                           |
| ------------------------------- | ------------------------- | ------------------------------------------------------------------------- |
| `--open-browser` + `--timeout`  | **Warning**, flag ignored | `"Warning: --timeout ignored with --open-browser (no content fetching)"`  |
| `--open-browser` + `--wait-for` | **Warning**, flag ignored | `"Warning: --wait-for ignored with --open-browser (no content fetching)"` |

**Browser Configuration Interactions:**

| Combination                                 | Behavior                  | Notes                                                                      |
| ------------------------------------------- | ------------------------- | -------------------------------------------------------------------------- |
| `--open-browser` + `--port`                 | Works normally            | Launch visible browser on custom port                                      |
| `--open-browser` + `--close-tab`            | **Warning**, flag ignored | `"Warning: --close-tab ignored with --open-browser (no content fetching)"` |
| `--open-browser` + `--user-agent` (no URLs) | Works normally            | UA set on the launched Chrome process                                      |
| `--open-browser` + `--user-agent` + URLs    | Works normally            | UA applied at launch and when opening URLs in tabs                         |
| `--open-browser` + `--user-data-dir`        | Works normally            | Launch visible browser with custom profile                                 |

If a debug browser is already running, `--user-agent` is ignored with a warning from the attach path (the running process already has its own user agent).

**Logging Interactions:**

All logging flags work normally:

| Combination                    | Behavior                                                   |
| ------------------------------ | ---------------------------------------------------------- |
| `--open-browser` + `--verbose` | Works normally (show verbose logs during browser launch)   |
| `--open-browser` + `--quiet`   | Works normally (suppress logs during browser launch)       |
| `--open-browser` + `--debug`   | Works normally (show debug/CDP logs during browser launch) |

#### Examples

**Valid:**

```bash
# Launch browser only
snag --open-browser

# Open browser with single URL (no fetch)
snag --open-browser https://example.com

# Open browser with multiple URLs in tabs (no fetch)
snag --open-browser https://example.com https://google.com

# Open browser with URLs from file (no fetch)
snag --open-browser --url-file urls.txt

# Open browser on custom port
snag --open-browser --port 9223

# Open browser with custom profile
snag --open-browser --user-data-dir ./my-profile

# Open browser with custom user agent (with or without URLs)
snag --open-browser --user-agent "CustomBot/1.0"
snag --open-browser --user-agent "CustomBot/1.0" https://example.com

# Open browser with verbose logging
snag --open-browser --verbose https://example.com
```

**Invalid (Errors):**

```bash
# ERROR: Conflicting modes
snag --open-browser --force-headless

# --open-browser ignored, lists tabs from existing browser
snag --open-browser --list-tabs
```

**With Warnings (Flags Ignored):**

```bash
# ⚠️ Output flags ignored (no fetching)
snag --open-browser https://example.com --output file.md
snag --open-browser https://example.com --output-dir ./out
snag --open-browser https://example.com --format pdf

# ⚠️ Timing flags ignored (no fetching)
snag --open-browser https://example.com --timeout 60
snag --open-browser https://example.com --wait-for ".content"

# ⚠️ Tab operations ignored (no fetching)
snag --open-browser --tab 1
snag --open-browser --all-tabs

# ⚠️ Close-tab ignored (no fetching)
snag --open-browser https://example.com --close-tab
```

#### Implementation Details

**Location:**

- Flag definition: `internal/cli/root.go` (`init`)
- Browser launch logic: `internal/browser` (`OpenBrowserOnly`, `Connect`)
- Handler: `internal/cli/root.go` (`runCobra`) and `internal/cli/handlers.go` (`handleOpenURLsInBrowser`)

**How it works:**

1. Check if `--open-browser` is set
2. Validate conflicts with `--force-headless` → Error if present
3. Warn about ignored flags (output, format, timing, tab operations, etc.)
4. Launch visible browser on specified port
5. If URLs provided: Navigate to each URL in separate tabs
6. Exit snag immediately (browser stays open)

**Error Messages:**

- Mode conflict: `"Cannot use both --force-headless and --open-browser (conflicting modes)"`

**Warning Messages:**

- Output flags: `"Warning: --output ignored with --open-browser (no content fetching)"`
- Format: `"Warning: --format ignored with --open-browser (no content fetching)"`
- Timeout: `"Warning: --timeout ignored with --open-browser (no content fetching)"`
- Wait-for: `"Warning: --wait-for ignored with --open-browser (no content fetching)"`
- Tab operations: `"Warning: --tab ignored with --open-browser (no content fetching)"`
- Close-tab: `"Warning: --close-tab ignored with --open-browser (no content fetching)"`

---
