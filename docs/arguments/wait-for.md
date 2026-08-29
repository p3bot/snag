# `--wait-for SELECTOR` / `-w`

**Status:** Complete (2025-10-23)

#### Validation Rules

**Invalid Values:**

**Empty string:**

- Behavior: **Warning + Ignored**
- Warning message: "Warning: --wait-for is empty, ignoring"
- Rationale: User explicitly provided flag but with no value - likely a mistake worth warning about
- Consistent with `--user-agent` and `--user-data-dir` empty string handling

**Whitespace-only string:**

- Behavior: **Ignored** after trimming
- All string arguments should be trimmed using `strings.TrimSpace()` after reading from CLI framework
- This is standard behavior in most CLI tools (git, docker, etc.)

**Invalid CSS selector syntax:**

- Behavior: **Error at runtime** (caught by rod's `Element()` method)
- No upfront validation - selector validation happens when rod tries to use it
- Error message from rod will indicate invalid selector

**Valid selector that never appears:**

- Behavior: **Timeout error** after `--timeout` duration expires
- Error message: "Timeout waiting for selector {selector}" with suggestion to increase timeout
- This is normal/expected behavior tested in test suite

**Multiple `--wait-for` flags:**

- Behavior: **Last wins** (standard CLI behavior, no error, no warning)

**Extremely complex selector:**

- Behavior: **Allow** if valid CSS syntax
- User's responsibility to provide working selectors

#### Interaction Matrix

**Content Source Interactions:**

**With URL arguments:**

| Combination                     | Behavior       | Notes                                       |
| ------------------------------- | -------------- | ------------------------------------------- |
| `--wait-for` + single `<url>`   | Works normally | Wait for selector after navigation          |
| `--wait-for` + multiple `<url>` | Works normally | Same selector applied to all URLs           |
| `--wait-for` + `--url-file`     | Works normally | Same selector applied to all URLs from file |

**With tab operations:**

| Combination                   | Behavior                | Notes                                                                           |
| ----------------------------- | ----------------------- | ------------------------------------------------------------------------------- |
| `--wait-for` + `--tab`        | **Works**               | Supports automation with persistent browser - wait for selector in existing tab |
| `--wait-for` + `--all-tabs`   | **Works**               | Same selector applied to all tabs before fetching                               |
| `--wait-for` + `--list-tabs`  | `--list-tabs` overrides | `--list-tabs` overrides all other options                                       |
| `--wait-for` + `--doctor`     | **Flag ignored**        | Doctor overrides, diagnostics only                                              |
| `--wait-for` + `--kill-browser` | **Flag ignored**        | No content to fetch                                                             |
| `--wait-for` + skill verb       | **Flag ignored**        | Skill modes do not fetch                                                        |

**Use case for tabs + wait-for:**

- Persistent visible browser with authenticated sessions
- Automated script runs periodically (e.g., cron job)
- Dynamic content loads asynchronously in existing tabs
- `--wait-for` ensures script waits for content before extracting
- Example: `snag --tab "dashboard" --wait-for ".data-loaded" -o daily-report.md`

**Output Control Interactions:**

All output flags work normally with `--wait-for`:

| Combination                     | Behavior                                       |
| ------------------------------- | ---------------------------------------------- |
| `--wait-for` + `--output`       | Works normally - wait before writing to file   |
| `--wait-for` + `--output-dir`   | Works normally - wait before auto-saving       |
| `--wait-for` + `--format` (all) | Works normally - wait before format conversion |

**Browser Mode Interactions:**

| Combination                              | Behavior       | Notes                                                |
| ---------------------------------------- | -------------- | ---------------------------------------------------- |
| `--wait-for` + `--force-headless`        | Works normally | Wait in headless mode                                |
| `--wait-for` + `--open-browser` (no URL) | **Warning**    | Flag ignored, no content to fetch                    |
| `--wait-for` + `--open-browser` + URL    | **Warning**    | Flag ignored, `--open-browser` doesn't fetch content |

**Warning message:**

- "Warning: --wait-for ignored with --open-browser (no content fetching)"

**Timing Interactions:**

| Combination                          | Behavior       | Notes                                                                      |
| ------------------------------------ | -------------- | -------------------------------------------------------------------------- |
| `--wait-for` + `--timeout`           | Works normally | `--timeout` applies to both navigation and selector wait                   |
| `--wait-for` + `--timeout` + `--tab` | Works normally | `--timeout` applies to `--wait-for` selector wait (warns if no --wait-for) |

**Current implementation:**

- Navigation timeout: Controlled by `--timeout` flag (default 30s)
- Selector wait timeout: Uses same `--timeout` duration
- Both use rod's timeout mechanism via `page.Timeout(duration)`

**Other Flag Interactions:**

**Compatible flags (work normally):**

- `--port` - Remote debugging port
- `--close-tab` - Close tab after fetching (works with `--tab`)
- `--verbose` / `--debug` - Logging levels
- `--user-agent` - Set user agent for new pages (ignored for existing tabs)
- `--user-data-dir` - Custom browser profile, wait for selector

#### Examples

**Valid:**

```bash
snag https://example.com --wait-for ".content"              # Basic usage
snag https://example.com --wait-for "#main-content"         # ID selector
snag https://example.com --wait-for "div > .loaded"         # Complex selector
snag https://example.com -w ".content" --timeout 60         # Custom timeout
snag url1 url2 --wait-for ".content"                        # Multiple URLs
snag --url-file urls.txt --wait-for ".content"              # URL file
snag --tab 1 --wait-for ".content" -o output.md             # Existing tab (automation)
snag --all-tabs --wait-for ".content" -d ./output           # All tabs
snag https://example.com --wait-for ".content" --format pdf # With PDF output
snag --wait-for ".content" --wait-for ".other"              # Uses ".other" (last wins)
```

**With Warnings:**

```bash
snag --open-browser https://example.com --wait-for ".content"  # ⚠️  No effect (no fetch)
snag --wait-for ""                                             # ⚠️  Warning: empty
snag --wait-for "   "                                          # ⚠️  Warning: empty (after trim)
```

## Implementation Details

**Location:**

- Flag definition: `internal/cli/root.go` (`init`)
- Wait logic: `internal/fetch` (`WaitForSelector`)
- Usage in fetch: `internal/fetch` (`Fetch`)
- Usage in tab handlers: `internal/cli/handlers.go`

**How it works:**

1. Takes CSS selector string value
2. Trim whitespace using `strings.TrimSpace()`
3. If empty string after trim, warn and skip: "Warning: --wait-for is empty, ignoring"
4. Wait for element to appear using `page.Element(selector)` with timeout
5. Wait for element to be visible using `elem.WaitVisible()`
6. Return error if selector never appears or never becomes visible
7. Timeout errors include helpful suggestion to increase `--timeout`

**Validation:**

- No upfront CSS syntax validation
- All validation happens at runtime through rod's CDP implementation
- Invalid selectors caught by rod's `Element()` method
- Timeout handled by rod's context deadline mechanism

**Timeout behavior:**

- Uses `--timeout` flag value (default 30 seconds)
- Applied to both `Element()` and `WaitVisible()` calls
- Errors include context about which operation timed out
