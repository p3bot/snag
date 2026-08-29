# `--kill-browser` / `-k`

**Status:** Complete (2025-10-30)

#### Validation Rules

**Boolean Flag:**

- No value required (presence = enabled)
- No validation errors possible

**Multiple Flags:**

- Multiple `--kill-browser` flags → **Silently ignored** (duplicate boolean)

#### Behavior

**Primary Purpose:**

- Forcefully terminate browser processes with remote debugging enabled
- Acts like `--help`, `--version`, or `--list-tabs`: Overrides all other flags except those needed for its operation
- Kills browser(s) and exits snag immediately

**Core Modes:**

1. **Kill all browsers with remote debugging (no --port):**

   ```bash
   snag --kill-browser
   ```

   - Kills all browser processes with `--remote-debugging-port` flag
   - Only targets development/debugging browsers, not regular browsing sessions
   - Reports count of killed processes or "No browser processes found"
   - Exit 0 (success, idempotent for scripting)

2. **Kill specific port:**
   ```bash
   snag --kill-browser --port 9223
   ```
   - Kills only the browser on specified port
   - Reports PID or "No browser running on port N"
   - Exit 0 (success, idempotent for scripting)

**Safety:**

- Only targets browsers with `--remote-debugging-port` flag
- Will NOT kill regular user browsing sessions
- Idempotent: Exit 0 even when nothing to kill (safe for scripting)

**Precedence Order:**

1. `--help` (highest priority, overrides everything)
2. `--version` (overrides everything below)
3. Skill flag + `--kill-browser` → usage-class error (neither runs)
4. `--doctor` (overrides everything below)
5. `--kill-browser` (overrides everything below)
6. `--list-tabs` (overrides everything below)
7. All other flags (ignored or error when `--kill-browser` is present)

#### Interaction Matrix

**Higher Priority Flags (Override):**

| Combination                     | Behavior             | Notes                      |
| ------------------------------- | -------------------- | -------------------------- |
| `--kill-browser` + `--doctor`   | `--doctor` overrides | Doctor has higher priority |

**Flags That Work WITH `--kill-browser`:**

| Combination                    | Behavior       | Notes                                        |
| ------------------------------ | -------------- | -------------------------------------------- |
| `--kill-browser` + `--port`    | Works normally | Kill only browser on specific port           |
| `--kill-browser` + `--verbose` | Works normally | Verbose logging (show PIDs, process details) |
| `--kill-browser` + `--debug`   | Works normally | Debug logging (show commands executed)       |

**Flags That ERROR With `--kill-browser` (Conflicting Operations):**

These flags represent conflicting operations and will produce an error:

| Combination                         | Behavior                       | Error Message                                 |
| ----------------------------------- | ------------------------------ | --------------------------------------------- |
| `--kill-browser` + `<url>`          | Error (conflicting operations) | Cannot use --kill-browser with URL arguments  |
| `--kill-browser` + `--url-file`     | Error (conflicting operations) | Cannot use --kill-browser with --url-file     |
| `--kill-browser` + `--list-tabs`    | Error (conflicting operations) | Cannot use --kill-browser with --list-tabs    |
| `--kill-browser` + `--all-tabs`     | Error (conflicting operations) | Cannot use --kill-browser with --all-tabs     |
| `--kill-browser` + `--tab`          | Error (conflicting operations) | Cannot use --kill-browser with --tab          |
| `--kill-browser` + `--open-browser` | Error (conflicting operations) | Cannot use --kill-browser with --open-browser |
| `--kill-browser` + skill verb       | Error (conflicting operations) | Cannot use a skill flag with --kill-browser (conflicting operations) |

**Flags That Are SILENTLY IGNORED:**

These flags are not applicable to killing browsers and are ignored:

| Combination                           | Behavior                     | Notes                            |
| ------------------------------------- | ---------------------------- | -------------------------------- |
| `--kill-browser` + `--output`         | Flag ignored, browser killed | No content to save               |
| `--kill-browser` + `--output-dir`     | Flag ignored, browser killed | No content to save               |
| `--kill-browser` + `--format`         | Flag ignored, browser killed | No content to format             |
| `--kill-browser` + `--timeout`        | Flag ignored, browser killed | No navigation performed          |
| `--kill-browser` + `--wait-for`       | Flag ignored, browser killed | No page loading                  |
| `--kill-browser` + `--close-tab`      | Flag ignored, browser killed | Entire browser is being killed   |
| `--kill-browser` + `--force-headless` | Flag ignored, browser killed | Kills existing browsers          |
| `--kill-browser` + `--user-agent`     | Flag ignored, browser killed | No navigation performed          |
| `--kill-browser` + `--user-data-dir`  | Flag ignored, browser killed | Kills existing browsers          |

**Rationale:**

- `--kill-browser` is a simple operational command like `--help` or `--version`
- Clear separation: either kill browsers OR perform other operations, not both
- Error on conflicting operations to prevent user confusion
- Ignore non-applicable flags to simplify UX

#### Examples

**Valid:**

```bash
# Kill all browsers with remote debugging enabled
snag --kill-browser
snag -k

# Kill only browser on specific port
snag --kill-browser --port 9223
snag -k --port 9223

# Kill with verbose logging
snag --kill-browser --verbose

# Kill with debug logging
snag --kill-browser --debug
```

**Invalid (Errors):**

```bash
# ERROR: Conflicting operations
snag --kill-browser https://example.com
snag --kill-browser --list-tabs
snag --kill-browser --tab 1
snag --kill-browser --all-tabs
snag --kill-browser --open-browser
snag --kill-browser --url-file urls.txt
```

**With Warnings (Flags Ignored):**

```bash
# Output flags ignored (no content)
snag --kill-browser --output file.md
snag --kill-browser --output-dir ./out
snag --kill-browser --format pdf

# Timing flags ignored (no navigation)
snag --kill-browser --timeout 60
snag --kill-browser --wait-for ".content"

# Browser config ignored (killing, not launching)
snag --kill-browser --force-headless
snag --kill-browser --user-agent "Bot/1.0"
```

**Success Cases:**

```bash
# No browsers running with remote debugging
snag --kill-browser
# → Output: "No browser processes found"
# → Exit: 0

# Browsers running with remote debugging
snag --kill-browser
# → Output: "✓ Killed 3 process(es)"
# → Exit: 0

# Specific port, browser running
snag --kill-browser --port 9223
# → Output: "✓ Killed browser process (PID 12345)"
# → Exit: 0

# Specific port, no browser
snag --kill-browser --port 9999
# → Output: "No browser running on port 9999"
# → Exit: 0
```

#### Implementation Details

**Location:**

- Flag definition: `internal/cli/root.go` (`init`)
- Flag priority chain: `internal/cli/root.go` (`runCobra`)
- Handler: `internal/cli/handlers.go` (`handleKillBrowser`)
- Kill logic: `internal/browser` (`KillBrowser`, `killBrowserOnPort`, `killAllBrowsers`)

**How it works:**

1. Check if `--kill-browser` is set (priority chain after `--help` and `--version`)
2. Validate no conflicting flags (errors on URL, tab operations, browser operations)
3. Extract `--port` flag if specified
4. Route to `handleKillBrowser(cmd)`
5. Call `bm.KillBrowser(port)` or `bm.KillBrowser(0)`
6. Exit snag immediately

**Kill strategy (without --port):**

- Detect browser type using `findBrowserPath()`
- Search processes: `ps aux | grep "$browserExe.*--remote-debugging-port"`
- Kill matched processes with `kill -9 <pid>`
- Return count of killed processes

**Kill strategy (with --port):**

- Find PID using `lsof -ti :<port>`
- If PID found: Kill with `kill -9 <pid>`
- If not found: Report "No browser running on port N" (exit 0)
- Return PID or 0

**Platform support:**

- Supported: macOS, Linux
- Not supported: Windows (requires `taskkill`, deferred to future)
- Commands used: `lsof`, `ps`, `grep`, `kill`

**Error Messages:**

- No browser found: `"No Chromium-based browser found"`
- Permission denied: `"Permission denied killing browser processes"`
- Conflicting operations: `"Cannot use --kill-browser with {flag} (conflicting operations)"`

**Warning Messages:**

None (non-applicable flags silently ignored)

---
