# `--user-data-dir DIRECTORY`

**Status:** Complete (2025-10-23)

#### Validation Rules

**Use Cases:**

- Multiple authenticated sessions (personal vs work accounts)
- Session isolation per project/client
- Privacy (separate from personal browsing)
- Enable true multi-instance browsers with different ports

**Invalid Values:**

**Empty string:**

- Behavior: **Warning + Ignored**, use browser default profile
- Warning message: "Warning: --user-data-dir is empty, using default profile"

**Whitespace-only string:**

- Behavior: **Warning + Ignored** after trimming, use browser default profile
- All string arguments trimmed using `strings.TrimSpace()`
- Same warning as empty string

**Directory doesn't exist:**

- Behavior: **Works normally** - Snag will create the directory automatically (like `mkdir -p`)
- Verbose messages: "Creating user data directory: {path}" → "User data directory created: {path}"
- No need to create directory manually before using it
- Creates full path including parent directories if needed

**Path exists but is a file (not directory):**

- Behavior: **Error**
- Error message: "Path is not a directory: {path}"

**Permission denied:**

- Behavior: **Error**
- Error when browser tries to read/write the directory
- Error message: "Permission denied accessing user data directory: {path}"

**Invalid path characters:**

- Behavior: **Error**
- Includes null bytes, system-dependent invalid characters
- Error message: "Invalid path: {path}"

**Relative vs absolute paths:**

- Behavior: Both supported, browser handles as needed
- Relative paths resolved relative to current working directory

**Tilde expansion:**

- Behavior: **Snag expands** `~` to home directory before passing to browser
- Example: `--user-data-dir ~/browsers/snag` → `/home/user/browsers/snag`

**Multiple `--user-data-dir` flags:**

- Behavior: **Last flag wins** (no error, no warning)
- Example: `--user-data-dir dir1 --user-data-dir dir2` → Uses `dir2`

**Default value (no flag):**

- Browser uses its default profile location
- Varies by OS and browser (Chrome, Chromium, Edge, Brave)

#### Interaction Matrix

**Browser Mode Interactions:**

| Combination                                   | Behavior       | Notes                                                      |
| --------------------------------------------- | -------------- | ---------------------------------------------------------- |
| `--user-data-dir` + `--force-headless`        | Works normally | Launch headless with custom profile                        |
| `--user-data-dir` + `--open-browser` (no URL) | Works normally | Open visible browser with custom profile                   |
| `--user-data-dir` + `--open-browser` + URLs   | Works normally | Open visible browser with custom profile, navigate to URLs |
| `--user-data-dir` + `--kill-browser`          | **Flag ignored** | No browser launch needed                                   |
| `--user-data-dir` + skill verb                | **Flag ignored** | Skill modes do not launch a browser                        |

**Connecting to existing browser:**

| Scenario                                        | Behavior                 | Notes                                          |
| ----------------------------------------------- | ------------------------ | ---------------------------------------------- |
| Browser already running + `--user-data-dir`     | **Warning**, ignore flag | Cannot change profile of running browser       |
| `--user-data-dir` + `--port` + existing browser | **Warning**, ignore flag | Connection to existing browser ignores profile |

**Warning message:**

- "Warning: --user-data-dir ignored when connecting to existing browser"

**Multiple instances with same profile:**

- Behavior: **Let browser error**
- Chrome/Chromium prevents multiple instances with same profile directory
- Error from browser: "Profile directory is locked" or similar
- Documented limitation - users must use different profiles for different ports

**Profile persistence:**

- Headless browser: Profile persists between snag invocations
- Visible browser: Profile persists between snag invocations
- All browser data stored in custom directory (cookies, sessions, cache, etc.)

## Content Source Interactions

| Combination                          | Behavior                 | Notes                                         |
| ------------------------------------ | ------------------------ | --------------------------------------------- |
| `--user-data-dir` + single `<url>`   | Works normally           | Fetch URL using custom profile                |
| `--user-data-dir` + multiple `<url>` | Works normally           | Fetch all URLs using same custom profile      |
| `--user-data-dir` + `--url-file`     | Works normally           | Fetch all URLs from file using custom profile |
| `--user-data-dir` + `--tab`          | **Warning**, ignore flag | Connecting to existing browser                |
| `--user-data-dir` + `--all-tabs`     | **Warning**, ignore flag | Connecting to existing browser                |
| `--user-data-dir` + `--list-tabs`    | `--list-tabs` overrides  | `--list-tabs` overrides all other options     |
| `--user-data-dir` + `--doctor`       | **Flag ignored**         | Doctor overrides, diagnostics only            |

**Warning message for tab operations:**

- "Warning: --user-data-dir ignored when connecting to existing browser"

## Output Control Interactions

All output flags work normally with `--user-data-dir`:

| Combination                          | Behavior                                                |
| ------------------------------------ | ------------------------------------------------------- |
| `--user-data-dir` + `--output`       | Works normally - custom profile, save to file           |
| `--user-data-dir` + `--output-dir`   | Works normally - custom profile, auto-save to directory |
| `--user-data-dir` + `--format` (all) | Works normally - all formats supported                  |

## Timing & Wait Interactions

| Combination                      | Behavior                                            |
| -------------------------------- | --------------------------------------------------- |
| `--user-data-dir` + `--timeout`  | Works normally - custom profile with custom timeout |
| `--user-data-dir` + `--wait-for` | Works normally - custom profile, wait for selector  |

## Other Browser Control Flags

| Combination                        | Behavior       | Notes                                        |
| ---------------------------------- | -------------- | -------------------------------------------- |
| `--user-data-dir` + `--close-tab`  | Works normally | Custom profile, close tab after fetch        |
| `--user-data-dir` + `--user-agent` | Works normally | Custom profile WITH custom UA (both applied) |
| `--user-data-dir` + `--port`       | Works normally | See Multi-instance Support below             |

## Logging Interactions

All logging flags work normally with `--user-data-dir`:

| Combination                     | Behavior       |
| ------------------------------- | -------------- |
| `--user-data-dir` + `--verbose` | Works normally |
| `--user-data-dir` + `--debug`   | Works normally |

**Special Cases:**

| Combination                     | Behavior     | Notes                                   |
| ------------------------------- | ------------ | --------------------------------------- |
| `--user-data-dir` + `--help`    | Help wins    | Display help, ignore all other flags    |
| `--user-data-dir` + `--version` | Version wins | Display version, ignore all other flags |

**Multi-Instance Support:**

**Enable multiple browser instances:**

```bash
# Launch two separate browser instances
snag --open-browser --port 9222 --user-data-dir ~/.snag/personal
snag --open-browser --port 9223 --user-data-dir ~/.snag/work

# Fetch from different instances
snag --port 9222 --tab "gmail" -o personal-email.md
snag --port 9223 --tab "gmail" -o work-email.md
```

**Requirements:**

- Different `--port` for each instance (browsers cannot share ports)
- Different `--user-data-dir` for each instance (Chrome locks profile directories)
- Each instance has isolated: sessions, cookies, cache, authentication

**Documented Limitations:**

- Same profile on different ports → Chrome error: "Profile directory is locked"
- Same port with different profiles → Port conflict error
- Must use both different port AND different profile for true multi-instance

#### Examples

**Valid:**

```bash
# Basic usage
snag --user-data-dir ~/.snag/default https://example.com
snag --user-data-dir ./profiles/project1 https://example.com

# Tilde expansion
snag --user-data-dir ~/browsers/snag https://example.com

# Relative paths
snag --user-data-dir ./my-profile https://example.com
snag --user-data-dir ../shared-profile https://example.com

# With browser modes
snag --open-browser --user-data-dir ~/.snag/personal
snag --force-headless --user-data-dir ~/.snag/scraper https://example.com

# With output control
snag --user-data-dir ~/.snag/work --tab "dashboard" -o report.md
snag --user-data-dir ~/.snag/research --url-file urls.txt -d ./output

# With other flags
snag --user-data-dir ~/.snag/test --user-agent "Custom UA" https://example.com
snag --user-data-dir ~/.snag/slow --timeout 60 --wait-for ".loaded" https://example.com

# Multi-instance (different port + different profile)
snag --open-browser --port 9222 --user-data-dir ~/.snag/personal
snag --open-browser --port 9223 --user-data-dir ~/.snag/work

# Last flag wins
snag --user-data-dir dir1 --user-data-dir dir2 https://example.com  # Uses dir2
```

**Invalid:**

```bash
snag --user-data-dir /etc/hosts https://example.com        # ERROR: Path is file, not directory
snag --user-data-dir /root/profile https://example.com     # ERROR: Permission denied (if exists)

# Multi-instance errors (same profile on different ports)
snag --open-browser --port 9222 --user-data-dir ~/.snag/profile1
snag --open-browser --port 9223 --user-data-dir ~/.snag/profile1
# Chrome error: Profile directory is locked
```

**With Warnings:**

```bash
# Empty/whitespace
snag --user-data-dir "" https://example.com                # ⚠️ Empty, use default
snag --user-data-dir "   " https://example.com             # ⚠️ Whitespace, use default

# Connecting to existing browser
# (Browser already running on port 9222)
snag --user-data-dir ~/.snag/different --port 9222 --tab 1 # ⚠️ Flag ignored, connecting to existing
snag --user-data-dir ~/.snag/profile --all-tabs            # ⚠️ Flag ignored, connecting to existing
snag --user-data-dir ~/.snag/profile --list-tabs           # Flag ignored (list-tabs standalone)
```

## Implementation Details

**Location:**

- Flag definition: `internal/cli/root.go` (`init`)
- Path validation: `internal/validate` (`UserDataDir`)
- Tilde expansion: Before validation, using `os.UserHomeDir()` or equivalent
- Browser launch: `internal/browser` (rod launcher with `--user-data-dir` flag)

**How it works:**

1. Read flag value from CLI framework
2. Trim whitespace using `strings.TrimSpace()`
3. If empty after trim → Warn, use browser default
4. Expand tilde (`~`) to home directory
5. If directory exists → Validate it's a directory and has permissions
6. If directory doesn't exist → Create it with `os.MkdirAll()` (like `mkdir -p`)
7. Pass to rod launcher as `--user-data-dir={path}` browser flag
8. Browser loads profile from specified directory

**Validation order:**

1. Trim whitespace
2. Check if empty → Warn if empty
3. Expand tilde
4. Check if path exists:
   - If doesn't exist → Create it with `os.MkdirAll(path, 0755)` (like `mkdir -p`)
   - If exists → Validate it's a directory and check permissions

**Browser default profiles:**

- Chrome (Linux): `~/.config/google-chrome/Default`
- Chrome (macOS): `~/Library/Application Support/Google/Chrome/Default`
- Chromium (Linux): `~/.config/chromium/Default`
- Location varies by browser and OS

**Profile persistence:**

- All browser data stored in custom directory
- Persists between snag invocations
- Includes: cookies, sessions, cache, history, extensions (if any)
- Useful for authenticated sessions that persist

---
