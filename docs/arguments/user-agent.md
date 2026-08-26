# `--user-agent STRING`

**Status:** Complete (2025-10-23)

#### Validation Rules

Set a custom user agent string for browser requests. This flag allows you to customize how the browser identifies itself to web servers, useful for bypassing headless detection or testing different user agent scenarios.

**Invalid Values:**

**Empty string:**

- Behavior: **Warning + Ignored**
- Warning message: "Warning: --user-agent is empty, using default user agent"
- Rationale: Empty UA provides no value, better to warn user and use browser default

**Whitespace-only string:**

- Behavior: **Trimmed, then warning + ignored**
- All string arguments should be trimmed using `strings.TrimSpace()` after reading from CLI framework
- After trimming to empty, same warning as empty string case
- This is standard behavior in most CLI tools (git, docker, etc.)

**Extremely long string:**

- Behavior: **Allow** (no artificial limit)
- Pass through to browser/CDP - let them handle header size limits
- Rationale: Browsers typically have ~8KB header limits; Rod will error if truly problematic
- No need for snag to impose artificial restrictions

**Multiple `--user-agent` flags:**

- Behavior: **Last wins** (standard CLI behavior, no error, no warning)
- Example: `snag --user-agent "First" --user-agent "Second" https://example.com` → Uses "Second"

**Special characters and Unicode:**

- Behavior: **Allow** (pass through as-is after trimming)
- HTTP headers handle encoding automatically
- Real-world use case: Some bots use emoji for identification (e.g., "Mozilla/5.0 🤖 Bot")
- Example: `snag --user-agent "Mozilla/5.0 🤖 Bot" https://example.com` → Works

**Newlines in user agent:**

- Behavior: **Strip/replace with space character**
- Silently sanitize: `strings.ReplaceAll(ua, "\n", " ")` and `strings.ReplaceAll(ua, "\r", " ")`
- Rationale: Newlines break HTTP protocol (header injection risk)
- Better UX than erroring; prevents security issues

#### Interaction Matrix

**Content Source Interactions:**

**With URL arguments:**

| Combination                       | Behavior       | Notes                                 |
| --------------------------------- | -------------- | ------------------------------------- |
| `--user-agent` + single `<url>`   | Works normally | UA applied to new page navigation     |
| `--user-agent` + multiple `<url>` | Works normally | Same UA applied to all URLs           |
| `--user-agent` + `--url-file`     | Works normally | Same UA applied to all URLs from file |

**With tab operations:**

| Combination                    | Behavior                | Notes                                             |
| ------------------------------ | ----------------------- | ------------------------------------------------- |
| `--user-agent` + `--tab`       | **Warning + Ignored**   | Tab already open with its own UA; can't change it |
| `--user-agent` + `--all-tabs`  | **Warning + Ignored**   | Tabs already open with their own UAs              |
| `--user-agent` + `--list-tabs` | `--list-tabs` overrides | `--list-tabs` overrides all other options         |
| `--user-agent` + `--doctor`    | **Flag ignored**        | Doctor overrides, diagnostics only                |

**Warning messages for tabs:**

- `--tab`: "Warning: --user-agent is ignored with --tab (cannot change existing tab's user agent)"
- `--all-tabs`: "Warning: --user-agent is ignored with --all-tabs (cannot change existing tabs' user agents)"

**Rationale for tab behavior:**

- Once a page/tab is loaded, its user agent cannot be changed
- The UA is set during initial navigation/page creation
- Existing tabs have already completed navigation with their original UA

## Browser Mode Interactions

| Combination                                | Behavior              | Notes                                            |
| ------------------------------------------ | --------------------- | ------------------------------------------------ |
| `--user-agent` + `--force-headless`        | Works normally | UA applied in headless mode                                         |
| `--user-agent` + `--open-browser` (no URL) | Works normally | UA set on the launched Chrome process                               |
| `--user-agent` + `--open-browser` + URL    | Works normally | UA applied at launch and when opening URLs in tabs                  |
| `--user-agent` + `--user-data-dir`         | Works normally | UA for this session; profile for persistent data                    |
| `--user-agent` + `--kill-browser`          | **Flag ignored** | No navigation performed                                           |

If a debug browser is already running, `--open-browser` attaches and warns that `--user-agent` is ignored (the running process already has its own user agent).

## Output Control Interactions

All output flags work normally with `--user-agent`:

| Combination                       | Behavior                                                        |
| --------------------------------- | --------------------------------------------------------------- |
| `--user-agent` + `--output`       | Works normally - UA used during fetch, content saved to file    |
| `--user-agent` + `--output-dir`   | Works normally - UA used during fetch, auto-save to directory   |
| `--user-agent` + `--format` (all) | Works normally - UA used during fetch, format applied to output |

## Timing and Page Control Interactions

| Combination                   | Behavior                                                     |
| ----------------------------- | ------------------------------------------------------------ |
| `--user-agent` + `--timeout`  | Works normally - UA set, then navigate with timeout          |
| `--user-agent` + `--wait-for` | Works normally - UA set, then navigate and wait for selector |

## Browser Configuration Interactions

| Combination                    | Behavior                                                          |
| ------------------------------ | ----------------------------------------------------------------- |
| `--user-agent` + `--port`      | Works normally - Connect to browser on port, set UA for new pages |
| `--user-agent` + `--close-tab` | Works normally - UA used during fetch, tab closed after           |

## Logging Interactions

All logging flags work normally with `--user-agent`:

| Combination                  | Behavior                                                   |
| ---------------------------- | ---------------------------------------------------------- |
| `--user-agent` + `--verbose` | Works normally - Verbose logs may show UA being set        |
| `--user-agent` + `--quiet`   | Works normally - Suppresses non-error messages             |
| `--user-agent` + `--debug`   | Works normally - Debug logs show CDP messages including UA |

## Help and Version Interactions

| Combination                  | Behavior                                            |
| ---------------------------- | --------------------------------------------------- |
| `--user-agent` + `--help`    | `--help` overrides everything, exits immediately    |
| `--user-agent` + `--version` | `--version` overrides everything, exits immediately |

#### Examples

**Valid usage:**

```bash
# Basic custom user agent
snag --user-agent "CustomBot/1.0" https://example.com

# Bypass headless detection
snag --user-agent "Mozilla/5.0 (Windows NT 10.0; Win64; x64)" https://example.com

# With multiple URLs
snag --user-agent "CustomBot/1.0" url1 url2 url3

# With URL file
snag --user-agent "CustomBot/1.0" --url-file urls.txt

# With output options
snag --user-agent "CustomBot/1.0" https://example.com -o output.md
snag --user-agent "CustomBot/1.0" https://example.com -d ./results

# With format options
snag --user-agent "CustomBot/1.0" https://example.com --format html
snag --user-agent "CustomBot/1.0" https://example.com --format pdf

# With wait-for and timeout
snag --user-agent "CustomBot/1.0" https://example.com --wait-for ".content" --timeout 60

# Open browser with custom UA (with or without URLs)
snag --user-agent "CustomBot/1.0" --open-browser
snag --user-agent "CustomBot/1.0" --open-browser https://example.com

# Special characters and Unicode (allowed)
snag --user-agent "Mozilla/5.0 🤖 Bot" https://example.com

# Multiple flags - last wins
snag --user-agent "First" --user-agent "Second" https://example.com  # Uses "Second"
```

**With warnings (flag ignored):**

```bash
# Empty or whitespace-only
snag --user-agent "" https://example.com          # ⚠️  Warning: empty UA
snag --user-agent "   " https://example.com       # ⚠️  Warning: empty after trim

# Existing tabs (can't change UA)
snag --user-agent "CustomBot/1.0" --tab 1         # ⚠️  Warning: tab already has UA
snag --user-agent "CustomBot/1.0" --all-tabs -d output  # ⚠️  Warning: tabs already have UAs
```

**Silently ignored:**

```bash
# List-tabs standalone mode
snag --user-agent "CustomBot/1.0" --list-tabs     # Silently ignored
```

**Sanitized (newlines stripped):**

```bash
# Newlines replaced with spaces (security + protocol compliance)
snag --user-agent "UA\nwith\nnewlines" https://example.com  # → "UA with newlines"
```

## Use Cases

**1. Bypass headless detection:**

```bash
# Some sites block headless browsers - use realistic desktop UA
snag --user-agent "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36" https://protected-site.com
```

**2. Mobile testing:**

```bash
# Test mobile-specific content
snag --user-agent "Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)" https://responsive-site.com
```

**3. Bot identification:**

```bash
# Identify your bot to site owners
snag --user-agent "MyCompanyBot/1.0 (+https://example.com/bot-info)" https://target-site.com
```

**4. Search engine simulation:**

```bash
# Test how your site appears to search engines
snag --user-agent "Googlebot/2.1 (+http://www.google.com/bot.html)" https://my-site.com
```

#### Implementation Details

**Processing order:**

1. Read flag value from CLI framework
2. Trim whitespace: `strings.TrimSpace(userAgent)`
3. Check if empty after trim → warn and ignore
4. Sanitize newlines: replace `\n` and `\r` with space
5. Check content source mode (URL vs tab)
6. If tab mode → warn and ignore
7. If URL mode → apply to browser/page during navigation

**When user agent is applied:**

- New page creation: Set before navigation
- Browser launch: Applied to all new pages created in that browser instance
- Existing tabs: Cannot be applied (warn and ignore)

**Location:**

- Flag definition: `internal/cli/root.go` (`init`)
- Validation/sanitization: `internal/validate` (`UserAgent`)
- Application: `internal/browser` (launcher `--user-agent` flag on launch; ignored when attaching to an existing browser)

**Validation rules:**

1. Trim whitespace
2. If empty → warn and ignore
3. Sanitize newlines (replace with space)
4. Check for tab mode → warn if tab-based operation
5. Pass through to Rod/browser (no length limits)

**Security considerations:**

- Newline stripping prevents HTTP header injection attacks
- No arbitrary restrictions on content (user responsibility)
- Warnings for ignored flags prevent user confusion
