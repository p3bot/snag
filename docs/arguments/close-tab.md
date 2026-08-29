# `--close-tab` / `-c`

**Status:** Complete (2025-10-23)

#### Validation Rules

**Boolean Flag:**

- No value required (presence = enabled)
- No validation errors possible

#### Behavior

**Default Behavior (No Flag):**

- Headless mode: Tabs closed automatically after fetch
- Visible browser: Tabs remain open after fetch

**With `--close-tab`:**

- Visible browser: Close tab after fetch (primary use case)
- Headless browser: Warning issued ("--close-tab is ignored in headless mode (tabs close automatically)"), proceeds normally
- Last tab: Closes tab AND browser with message: "Closing last tab, browser will close"
- Close failure: Warning issued but fetch considered successful (content already retrieved)

**Browser Close Behavior:**

- Closing the last tab will also close the browser
- Message displayed when this occurs
- Browser processes cleanly terminated

#### Interaction Matrix

**Content Source Interactions:**

| Combination                               | Behavior                  | Notes                                                                                                                    |
| ----------------------------------------- | ------------------------- | ------------------------------------------------------------------------------------------------------------------------ |
| `--close-tab` + single `<url>`            | Works normally            | Open tab, fetch, close tab (or warn if headless)                                                                         |
| `--close-tab` + multiple `<url>`          | Works normally            | Close each tab after fetching                                                                                            |
| `--close-tab` + `--url-file`              | Works normally            | Close each tab after fetching                                                                                            |
| `--close-tab` + `--tab`                   | Works normally            | Fetch from existing tab, then close it; if last tab, close browser                                                       |
| `--close-tab` + `--all-tabs`              | Works normally            | Fetch from all tabs, close all tabs, close browser                                                                       |
| `--close-tab` + `--list-tabs`             | `--list-tabs` overrides   | `--list-tabs` overrides all other options                                                                                |
| `--close-tab` + `--doctor`                | **Flag ignored**          | Doctor overrides, diagnostics only                                                                                       |
| `--close-tab` + `--open-browser` (no URL) | **Warning**               | `"Warning: --close-tab ignored with --open-browser (no content fetching)"`, browser stays open                           |
| `--close-tab` + `--open-browser` + URL    | **Warning**, flag ignored | `"Warning: --close-tab ignored with --open-browser (no content fetching)"` - Opens browser, navigates to URL, exits snag |
| `--close-tab` + `--kill-browser`          | **Flag ignored**          | No tabs to close                                                                                                         |
| `--close-tab` + skill verb                | **Flag ignored**          | Skill modes do not fetch                                                                                                 |

**Browser Mode Interactions:**

| Combination                        | Behavior    | Notes                                                                                              |
| ---------------------------------- | ----------- | -------------------------------------------------------------------------------------------------- |
| `--close-tab` + `--force-headless` | **Warning** | `"Warning: --close-tab is ignored in headless mode (tabs close automatically)"`, proceeds normally |

**Output Control (All work normally):**

| Combination                      | Behavior                                             |
| -------------------------------- | ---------------------------------------------------- |
| `--close-tab` + `--output`       | Works - fetch content, save to file, close tab       |
| `--close-tab` + `--output-dir`   | Works - fetch content, save to directory, close tabs |
| `--close-tab` + `--format` (any) | Works - fetch in any format, close tab               |

**Other Compatible Flags (All work normally):**

| Combination                                     | Behavior                                                   |
| ----------------------------------------------- | ---------------------------------------------------------- |
| `--close-tab` + `--wait-for`                    | Works - wait for selector, fetch, close tab                |
| `--close-tab` + `--timeout`                     | Works - apply timeout to fetch, then close tab             |
| `--close-tab` + `--port`                        | Works - connect to specific port, close tabs after fetch   |
| `--close-tab` + `--user-agent`                  | Works - set user agent for new tabs, close after fetch     |
| `--close-tab` + `--user-data-dir`               | Works - use custom browser profile, close tabs after fetch |
| `--close-tab` + `--verbose`/`--debug`           | Works - logging applies to close operations                |

#### Examples

**Valid:**

```bash
snag https://example.com --close-tab              # Visible: fetch and close tab
snag https://example.com --close-tab              # Headless: warn, close anyway
snag url1 url2 url3 --close-tab                   # Close each tab after fetch
snag --url-file urls.txt --close-tab              # Close each tab after fetch
snag --tab 1 --close-tab                          # Fetch from existing tab, close it
snag --tab "dashboard" --close-tab -o report.md   # Fetch, save, close tab
snag --all-tabs --close-tab -d ./output           # Fetch all, save all, close all, close browser
```

**Invalid:**

```bash
snag --list-tabs --close-tab                      # --close-tab ignored, lists tabs from existing browser
```

**With Warnings:**

```bash
snag --close-tab --open-browser                   # ⚠️ Warning: ignored (no content fetching)
snag --force-headless https://example.com --close-tab  # ⚠️ Warning: --close-tab is ignored in headless mode (tabs close automatically)
```

**Special Cases:**

```bash
snag --tab 1 --close-tab                          # Last tab → closes browser too
snag --all-tabs --close-tab                       # Closes all tabs → closes browser
```

#### Implementation Details

**Location:**

- Flag definition: `internal/cli/root.go` (`init`)
- Close logic: `internal/browser` (`ClosePage`) and `internal/cli/handlers.go`

**How it works:**

1. After content fetch completes successfully
2. If `--close-tab` enabled and visible browser mode:
   - Attempt to close the tab via CDP
   - If last tab, message user and close browser
3. If `--close-tab` enabled and headless mode:
   - Warning message (tabs close automatically anyway)
4. If close operation fails:
   - Warning logged but operation considered successful
   - Content was already fetched before close attempted

**Error Handling:**

- Tab close failures don't fail the overall operation
- Content retrieval success is independent of tab close success
- Browser close failures logged but ignored (process cleanup)

**Parallel Processing Note:**

- Multiple URL processing strategy (sequential vs parallel) is tracked in TODO
- Affects whether tabs are closed one-by-one or in batch
- Current behavior: Sequential processing

---
