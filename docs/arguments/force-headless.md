# `--force-headless`

**Status:** Complete (2025-10-23)

#### Validation Rules

**Boolean Flag:**

- No value required (presence = enabled)
- No validation errors possible

**Multiple Flags:**

- Multiple `--force-headless` flags → **Silently ignored** (duplicate boolean)

#### Behavior

**Primary Purpose:**

- Override auto-detection to force launching a headless browser
- Useful for automation that requires consistent headless behavior

**When No Existing Browser:**

- Flag is **silently ignored** (headless is already default behavior)
- Browser launches in headless mode on default or specified port

**When Existing Browser Running:**

- Default port (9222): Launch fails (port in use). If the default snag profile is also locked, snag reports `launch profile is in use`
- Custom port via `--port`: Launches a new headless browser on that port, still using the default snag profile unless `--temp-profile` or `--user-data-dir` is set. A visible snag browser holding that profile causes `launch profile is in use`

**Browser Mode Conflicts:**

- With `--open-browser`: **Error** - `"Cannot use both --force-headless and --open-browser (conflicting modes)"`

#### Interaction Matrix

**Content Source Interactions:**

| Combination                           | Behavior                 | Notes                                                                                             |
| ------------------------------------- | ------------------------ | ------------------------------------------------------------------------------------------------- |
| `--force-headless` + single `<url>`   | **Silently ignore** flag | Headless is default, not needed                                                                   |
| `--force-headless` + multiple `<url>` | **Silently ignore** flag | Headless is default, not needed                                                                   |
| `--force-headless` + `--url-file`     | **Silently ignore** flag | Headless is default, not needed                                                                   |
| `--force-headless` + `--tab`          | **Error**                | `"Cannot use --force-headless with --tab (--tab requires existing browser connection)"`           |
| `--force-headless` + `--all-tabs`     | **Error**                | `"Cannot use --force-headless with --all-tabs (--all-tabs requires existing browser connection)"` |
| `--force-headless` + `--list-tabs`    | `--list-tabs` overrides  | `--list-tabs` overrides all other options                                                         |
| `--force-headless` + `--doctor`       | **Flag ignored**         | Doctor overrides, diagnostics only                                                                |
| `--force-headless` + `--kill-browser` | **Flag ignored**         | No browser launch needed                                                                          |
| `--force-headless` + skill verb       | **Flag ignored**         | Skill modes do not launch a browser                                                               |

**Rationale for Tab Errors:**

- `--force-headless` implies launching a new browser
- Tab operations (`--tab`, `--all-tabs`) require existing browser with tabs
- These are fundamentally incompatible operations

**Browser Mode Interactions:**

| Combination                            | Behavior       | Notes                                            |
| -------------------------------------- | -------------- | ------------------------------------------------ |
| `--force-headless` + `--open-browser`  | **Error**      | Conflicting modes (open-browser implies visible) |
| `--force-headless` + `--user-data-dir` | Works normally | Launch headless with custom profile              |
| `--force-headless` + `--temp-profile`  | Works normally | Launch headless with an ephemeral profile        |

**Other Flag Interactions:**

| Combination                                          | Behavior       | Notes                                                                           |
| ---------------------------------------------------- | -------------- | ------------------------------------------------------------------------------- |
| `--force-headless` + `--close-tab`                   | **Warning**    | `"Warning: --close-tab is ignored in headless mode (tabs close automatically)"` |
| `--force-headless` + `--port`                        | Works normally | Launch headless on specified port                                               |
| `--force-headless` + `--output` / `--output-dir`     | Works normally | Output control unaffected by browser mode                                       |
| `--force-headless` + `--format` (any)                | Works normally | Format conversion unaffected by browser mode                                    |
| `--force-headless` + `--timeout`                     | Works normally | Navigation timeout applies                                                      |
| `--force-headless` + `--wait-for`                    | Works normally | Selector wait applies                                                           |
| `--force-headless` + `--user-agent`                  | Works normally | User agent set for headless browser                                             |
| `--force-headless` + `--verbose`/`--debug`           | Works normally | Logging levels apply                                                            |

#### Examples

**Valid:**

```bash
# Force headless when browser might be open (silently ignored if none open)
snag --force-headless https://example.com

# Force headless with custom port (default profile is a singleton; isolate it)
snag --force-headless --port 9223 --temp-profile https://example.com
snag --force-headless --port 9223 --user-data-dir /tmp/snag-profile https://example.com

# Force headless with custom profile
snag --force-headless --user-data-dir /tmp/snag-profile https://example.com

# Force headless with output options
snag --force-headless https://example.com -o output.md
snag --force-headless https://example.com --format pdf
```

**Invalid (Errors):**

```bash
# ERROR: Conflicting modes
snag --force-headless --open-browser

# ERROR: Tab operations require existing browser
snag --force-headless --tab 1
snag --force-headless --all-tabs
snag --force-headless --list-tabs                    # --force-headless ignored, lists tabs from existing browser
```

**With Warnings:**

```bash
# ⚠️ Warning: redundant in headless mode
snag --force-headless --close-tab https://example.com
```

**Silently Ignored (Not Needed):**

```bash
# Headless is default behavior when no browser open
snag --force-headless https://example.com          # Flag ignored (no browser)
snag --force-headless url1 url2                    # Flag ignored (no browser)
snag --force-headless --url-file urls.txt          # Flag ignored (no browser)

# Multiple flags (duplicate boolean)
snag --force-headless --force-headless https://example.com
```

#### Implementation Details

**Location:**

- Flag definition: `internal/cli/root.go` (`init`)
- Browser launch logic: `internal/browser` (mode detection and launch)

**How it works:**

1. Check if `--force-headless` is set
2. If set with conflicting flags (`--open-browser`, tab operations) → Error
3. If set with `--close-tab` → Warning (redundant)
4. If no existing browser running → Silently ignore (default is headless)
5. If existing browser on default port → Let connection fail (port conflict)
6. If existing browser + custom `--port` → Launch new headless on custom port using the same resolved profile as any other unspecified launch (XDG/darwin default unless `--temp-profile` or `--user-data-dir` is set). If Chrome still holds `SingletonLock` after launch fails, snag reports `launch profile is in use` and suggests `snag --temp-profile --force-headless --port <port> <url>`.

**Error Messages:**

- Tab operation conflicts: `"Cannot use --force-headless with --tab (--tab requires existing browser connection)"`
- Profile already locked: `"Launch profile is in use by another Chrome process"` with `Try: snag --temp-profile --force-headless --port 9223 <url>` (uses the requested port when it is not 9222)

**Warning Messages:**

- With `--close-tab`: `"Warning: --close-tab is ignored in headless mode (tabs close automatically)"`

---
