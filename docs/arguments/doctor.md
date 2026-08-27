# `--doctor`

**Status:** Complete (2025-10-30)

#### Validation Rules

**Boolean Flag:**

- No value required (presence = enabled)
- No validation errors possible

**Priority Behavior:**

- Displays diagnostic information and exits immediately
- Exit code 0 (always success)
- Lower priority than `--help` and `--version`
- Higher priority than `--kill-browser` and all content operations

#### Behavior

**Basic Usage:**

```bash
snag --doctor
snag --doctor --port 9223
snag --doctor --verbose
```

**Primary Purpose:**

- Display comprehensive diagnostic information about snag's environment
- Help users troubleshoot issues
- Provide maintainers with essential debug info for issue reports
- Similar to `brew doctor`, `npm doctor`, etc.

**Diagnostic Information Displayed:**

1. **Version Information:**
   - snag version (current and latest from GitHub)
   - Go version
   - OS/Architecture

2. **Working Directory:**
   - Current working directory

3. **Browser Detection:**
   - Detected browser name (Chrome, Chromium, Brave, Edge, etc.)
   - Browser executable path
   - Browser version (raw output from `--version`)

4. **Connection Status:**
   - Default port 9222 status (running/not running, tab count)
   - Custom port status if `--port` specified

5. **Profile Location:**
   - Profile path for detected browser
   - Existence check (✓/✗)

6. **Environment Variables:**
   - `CHROME_PATH` (if set)
   - `CHROMIUM_PATH` (if set)

**Exit Behavior:**

- Always exits with code 0 (success)
- Never fails even if some checks fail (partial information still useful)
- Diagnostic mode complete → snag exits (no other operations performed)

#### Interaction Matrix

**Flags That Override `--doctor` (Higher Priority):**

| Combination            | Behavior             | Rationale                    |
| ---------------------- | -------------------- | ---------------------------- |
| `--doctor` + `--help`  | Display help, exit 0 | Help has highest priority    |
| `--help` + `--doctor`  | Display help, exit 0 | Help has highest priority    |
| `--doctor` + `--version` | Display version, exit 0 | Version has higher priority |

**Flags That Work WITH `--doctor`:**

| Combination               | Behavior       | Notes                                          |
| ------------------------- | -------------- | ---------------------------------------------- |
| `--doctor` + `--port`     | Works normally | Check both default (9222) and custom port      |
| `--doctor` + `--verbose`  | Works normally | Verbose logging during diagnostic operations   |
| `--doctor` + `--debug`    | Works normally | Debug logging during diagnostic operations     |

**All Other Flags Are SILENTLY IGNORED:**

`--doctor` acts like `--help` and overrides all other arguments:

| Combination                       | Behavior                       | Notes                                    |
| --------------------------------- | ------------------------------ | ---------------------------------------- |
| `--doctor` + `<url>`              | URL ignored, doctor runs       | Diagnostics, ignores URL                 |
| `--doctor` + `--url-file`         | Flag ignored, doctor runs      | Diagnostics, ignores file                |
| `--doctor` + `--output`           | Flag ignored, doctor runs      | Diagnostics to stdout only               |
| `--doctor` + `--output-dir`       | Flag ignored, doctor runs      | Diagnostics to stdout only               |
| `--doctor` + `--format`           | Flag ignored, doctor runs      | Diagnostics in fixed format              |
| `--doctor` + `--timeout`          | Flag ignored, doctor runs      | No navigation performed                  |
| `--doctor` + `--wait-for`         | Flag ignored, doctor runs      | No content fetching                      |
| `--doctor` + `--close-tab`        | Flag ignored, doctor runs      | No tab operations                        |
| `--doctor` + `--force-headless`   | Flag ignored, doctor runs      | Connects to existing browser             |
| `--doctor` + `--open-browser`     | Flag ignored, doctor runs      | No browser launch                        |
| `--doctor` + `--kill-browser`     | Flag ignored, doctor runs      | Doctor has higher priority               |
| `--doctor` + `--tab`              | Flag ignored, doctor runs      | No content fetching                      |
| `--doctor` + `--all-tabs`         | Flag ignored, doctor runs      | No content fetching                      |
| `--doctor` + `--list-tabs`        | Flag ignored, doctor runs      | Doctor has higher priority               |
| `--doctor` + `--user-agent`       | Flag ignored, doctor runs      | No navigation performed                  |
| `--doctor` + `--user-data-dir`    | Flag ignored, doctor runs      | Connects to existing browser             |

**Priority Rules:**

1. `--help` detected → Display help (highest priority)
2. `--version` detected → Display version
3. `--doctor` detected → Display diagnostics
4. Ignore all other flags (except `--port` and logging flags which enhance diagnostics)
5. Exit with code 0

**Rationale:**

- `--doctor` is a simple informational command like `--help` or `--version`
- Users expect it to "just work" without complex argument validation
- Simplifies UX: No need to remember which flags conflict with `--doctor`
- Other tools follow this pattern (e.g., `brew doctor <any-args>` ignores all args)

#### Examples

**Valid:**

```bash
# Basic diagnostics
snag --doctor

# Check specific port
snag --doctor --port 9223

# With verbose logging
snag --doctor --verbose

# With debug logging
snag --doctor --debug
```

**Silently Ignores Other Flags:**

```bash
# All other flags are ignored, diagnostics displayed
snag --doctor https://example.com
snag --doctor --output file.md
snag --doctor --format pdf --wait-for ".content"
snag --doctor --tab 1 --close-tab
snag --doctor --force-headless --user-agent "Bot/1.0"
snag --doctor --kill-browser
snag --doctor --list-tabs
```

**Higher Priority Flags Win:**

```bash
# Shows HELP (not diagnostics)
snag --doctor --help
snag --help --doctor

# Shows VERSION (not diagnostics)
snag --doctor --version
```

**Output Example:**

```
snag Doctor Report
==================
https://github.com/p3bot/snag

Version Information
───────────────────
  snag version:    0.0.4
  Latest version:  0.0.5 (update available)
  Go version:      go1.25.3
  OS/Arch:         darwin/arm64

Working Directory
─────────────────
  /Users/username/projects/myproject

Browser Detection
─────────────────
  Detected:        Chrome
  Path:            /Applications/Google Chrome.app/Contents/MacOS/Google Chrome
  Version:         131.0.6778.85

Connection Status
─────────────────
  Port 9222:       ✓ Running (7 tabs open)

Profile Location
─────────────────
  Chrome:          ✓ ~/Library/Application Support/Google/Chrome

Environment Variables
─────────────────────
  CHROME_PATH:     (not set)
  CHROMIUM_PATH:   (not set)
```

**With Custom Port:**

```bash
snag --doctor --port 9223
```

```
Connection Status
─────────────────
  Port 9222:       ✓ Running (7 tabs open)
  Port 9223:       ✗ Not running
```

**No Browser Detected:**

```
Browser Detection
─────────────────
  Detected:        ✗ No Chromium-based browser found
  Path:            (none)
  Version:         (none)

Connection Status
─────────────────
  Port 9222:       ✗ Not running

Profile Location
─────────────────
  (no browser detected)
```

**Update Check Failed:**

```
Version Information
───────────────────
  snag version:    0.0.4
  Latest version:  (unable to check - network error)
  Go version:      go1.25.3
  OS/Arch:         darwin/arm64
```

#### Implementation Details

**Location:**

- Flag definition: `internal/cli/root.go` (`init`)
- Flag priority chain: `internal/cli/root.go` (`runCobra`)
- Handler: `internal/cli/handlers.go` (`handleDoctor`)
- Doctor logic: `internal/doctor/doctor.go` (`DoctorReport`, `CollectDoctorInfo`, `Print`)

**How it works:**

1. Check if `--doctor` is set (priority chain after `--help` and `--version`)
2. Extract `--port` and logging flags (`--verbose`, `--debug`)
3. Silently ignore all other flags (no warnings, no errors)
4. Collect diagnostic information:
   - Version info (snag, Go, OS/Arch)
   - GitHub latest release (10 second timeout)
   - Working directory
   - Browser detection (path, version)
   - Profile path and existence
   - Connection status (default + custom port if specified)
   - Environment variables
5. Format and print report to stdout
6. Exit snag immediately with code 0

**Data Collection:**

- **Version**: From build ldflags (`-X github.com/p3bot/snag/internal/cli.Version=...`)
- **Go Version**: `runtime.Version()`
- **OS/Arch**: `runtime.GOOS`, `runtime.GOARCH`
- **Working Dir**: `os.Getwd()`
- **Browser Path**: `findBrowserPath()` (reuses existing detection)
- **Browser Version**: Execute `<browser> --version`, return raw output
- **Profile Path**: OS-specific paths from `browserDetectionRule` struct
- **Connection**: Try connecting to port, count tabs if successful
- **Latest Version**: GitHub API query with 10s timeout
- **Env Vars**: `os.Getenv()` for `CHROME_PATH`, `CHROMIUM_PATH`

**Output Format:**

- Human-readable text with sections
- Unicode box drawing characters for section dividers
- Checkmarks (✓) for success/exists
- X marks (✗) for failure/not found
- Aligned labels and values
- Output to stdout (enables piping/redirection)
- Logs to stderr (verbose/debug mode)

**Error Handling:**

- Never fails (always exit 0)
- Partial information still displayed if some checks fail
- Network errors for GitHub check → Display "unable to check"
- No browser found → Display clear message with alternatives
- Connection failures → Display "Not running"

**Use Cases:**

- Troubleshooting environment issues
- Verifying browser installation and version
- Checking if browser is running on expected port
- Finding browser profile locations
- Generating debug info for issue reports
- Checking for snag updates

**Design Note:**

- Output goes to stdout (not stderr) to enable redirection: `snag --doctor > diagnostics.txt`
- Verbose/debug logs go to stderr during collection
- Read-only operation (never modifies state)
- Idempotent (safe to run repeatedly)

---
