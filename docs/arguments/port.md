# `--port PORT` / `-p`

**Status:** Complete (2025-10-23)

#### Validation Rules

**Invalid Values:**

**Valid port range:**

- Ports 1024-65535 allowed (non-privileged ports only)
- Excludes privileged ports 1-1023 (require root/admin)
- Error immediately if outside valid range

**Validation errors:**

| Value                               | Behavior                    | Error Message                               |
| ----------------------------------- | --------------------------- | ------------------------------------------- |
| Negative (e.g., `-1`)               | Error immediately           | "Port must be between 1024 and 65535"       |
| Zero                                | Error immediately           | "Port must be between 1024 and 65535"       |
| Below 1024 (e.g., `80`, `443`)      | Error immediately           | "Port must be between 1024 and 65535"       |
| Above 65535 (e.g., `70000`)         | Error immediately           | "Port must be between 1024 and 65535"       |
| Non-integer (e.g., `9222.5`)        | Error immediately           | "Invalid port value: must be an integer"    |
| Non-numeric (e.g., `abc`)           | Error immediately           | "Invalid port value: {value}"               |
| Empty string                        | Fall back to default (9222) | No error                                    |
| Multiple `--port` flags             | Last wins                   | No error (standard CLI behavior)            |
| Port in use (by non-Chrome process) | Error at connection time    | "Failed to connect to port {port}: {error}" |

**Default behavior:**

- No `--port` specified → Use default port 9222

#### Behavior

**Port-specific connection:**

```bash
snag --port 9223 https://example.com
```

- Attempts to connect **only** to port 9223
- No auto-detection of browsers on other ports
- If browser responds on that port → Connect to it
- If no browser responds → Launch new browser on that port

**Connection strategy:**

1. Try to connect to specified port
2. If connection fails, attempt to launch on specified port
3. Launch may fail if another Chrome instance already running (Chrome locks profile)

#### Interaction Matrix

**Browser Mode Interactions:**

| Combination                     | Behavior                 | Notes                                     |
| ------------------------------- | ------------------------ | ----------------------------------------- |
| `--port 9222` (default)         | Connect/launch on 9222   | Standard behavior                         |
| `--port 9223` + no browser      | Launch on port 9223      | New instance                              |
| `--port 9223` + browser on 9223 | Connect to existing      | Reuse instance                            |
| `--port 9223` + browser on 9222 | No cross-detection       | 9222 browser not detected, try 9223       |
| `--port` + `--force-headless`   | Launch headless on port  | Works normally                            |
| `--port` + `--kill-browser`     | Kill browser on port     | Works normally                            |
| `--port` + `--doctor`           | Check custom port status | Works normally                            |
| `--port` + `--open-browser`     | Open browser on port     | Works normally                            |
| `--port` + `--list-tabs`        | `--list-tabs` overrides  | `--list-tabs` overrides all other options |
| `--port` + `--tab`              | Fetch tab from port      | Works normally                            |
| `--port` + `--all-tabs`         | Fetch all tabs from port | Works normally                            |

**Port + User Data Directory Pairing:**

| Combination                                          | Behavior                                  | Notes                                                                  |
| ---------------------------------------------------- | ----------------------------------------- | ---------------------------------------------------------------------- |
| `--port 9222` + `--user-data-dir ~/.snag/profile1`   | Launch/connect with profile1 on port 9222 | Works normally                                                         |
| Same profile, different ports                        | Attempt launch, Chrome will error         | Chrome locks user data directories - **documented limitation**         |
| Same port, different profiles                        | Port in use error                         | Cannot run multiple browsers on same port                              |
| Connecting to existing + different `--user-data-dir` | **Warn**, continue connecting             | "Warning: --user-data-dir ignored when connecting to existing browser" |
| Neither specified                                    | Port 9222 + default profile               | Current behavior preserved                                             |

**Multi-instance support:**

- Different ports + different `--user-data-dir` → Multiple isolated browser instances
- Same port → Only one instance (conflict)
- Same profile → Chrome prevents (directory lock)

**Content Source Interactions:**

All content source flags work normally with `--port`:

| Combination              | Behavior                                          |
| ------------------------ | ------------------------------------------------- |
| `--port` + `<url>`       | Works normally - fetch URL from browser on port   |
| `--port` + `--url-file`  | Works normally - use browser on port for all URLs |
| `--port` + multiple URLs | Works normally - use browser on port for all      |

**Output Control:**

All output flags work normally with `--port`:

| Combination               | Behavior                                          |
| ------------------------- | ------------------------------------------------- |
| `--port` + `--output`     | Works normally - just controls browser connection |
| `--port` + `--output-dir` | Works normally                                    |
| `--port` + `--format`     | Works normally                                    |

**Page Loading:**

All page loading flags work normally with `--port`:

| Combination               | Behavior                                        |
| ------------------------- | ----------------------------------------------- |
| `--port` + `--timeout`    | Works normally                                  |
| `--port` + `--wait-for`   | Works normally                                  |
| `--port` + `--user-agent` | Works normally - applies when launching browser |

**Browser Control:**

| Combination              | Behavior       |
| ------------------------ | -------------- |
| `--port` + `--close-tab` | Works normally |

**Logging Flags:**

- All logging flags work normally with `--port`

#### Examples

**Valid:**

```bash
snag --port 9222 https://example.com              # Default port (explicit)
snag --port 9223 https://example.com              # Custom port
snag -p 4444 https://example.com                  # Short form
snag --port 9223 --list-tabs                      # List tabs on custom port
snag --port 9223 --tab 1                          # Fetch from tab on custom port
snag --port 9223 --open-browser                   # Open browser on custom port

# Multi-instance with user-data-dir
snag --open-browser --port 9222 --user-data-dir ~/.snag/personal
snag --open-browser --port 9223 --user-data-dir ~/.snag/work

# Fetch from different instances
snag --port 9222 --tab "gmail" -o personal.md
snag --port 9223 --tab "gmail" -o work.md
snag --port 9222 --port 9223 https://example.com  # Uses 9223 (last wins)
```

**Invalid:**

```bash
snag --port -1 https://example.com                # ERROR: Negative
snag --port 0 https://example.com                 # ERROR: Zero
snag --port 80 https://example.com                # ERROR: Privileged port
snag --port 70000 https://example.com             # ERROR: Above 65535
snag --port 9222.5 https://example.com            # ERROR: Non-integer
snag --port abc https://example.com               # ERROR: Non-numeric
```

**With Warnings:**

```bash
# Browser on 9222 using profile1, connect with different profile
snag --port 9222 --user-data-dir ~/.snag/profile2 <url>
# ⚠️ Warning: --user-data-dir ignored when connecting to existing browser
```

**Documented Limitations:**

```bash
# Same profile on different ports - Chrome will error
snag --open-browser --port 9222 --user-data-dir ~/.snag/profile1
snag --open-browser --port 9223 --user-data-dir ~/.snag/profile1
# Chrome error: Profile directory is locked
```

#### Implementation Details

**Location:**

- Flag definition: `internal/cli/root.go` (`init`)
- Port validation: `internal/validate` (`Port`, range 1024-65535)
- Connection logic: `internal/browser` (`Connect`)
- Launch logic: `internal/browser` (`launchBrowser`)

**How it works:**

1. Validate port is in range 1024-65535
2. Check for multiple `--port` flags
3. When connecting:
   - Try to connect to specified port only
   - If fails, attempt to launch browser on that port
4. When launching:
   - Set `--remote-debugging-port={port}` flag in `internal/browser` (`launchBrowser`)
   - Launch browser with port configuration

**Port-specific behavior:**

- `--port` specified → Only try that specific port, no auto-detection
- No `--port` → Try default 9222, then auto-detect, then launch
- Port conflicts handled by Chrome/OS (connection failure)

**Multi-instance considerations:**

- Different ports enable multiple browser instances
- Requires different `--user-data-dir` for each instance (Chrome locks profiles)
- Each instance independent (sessions, cookies, authentication)

---
