# `--temp-profile`

**Status:** Complete (2026-09-09)

#### Validation Rules

**Boolean Flag:**

- No value required (presence = enabled)
- Mutually exclusive with `--user-data-dir` (usage error, exit 1)

**Purpose:**

- Launch with a unique directory under `os.TempDir()` (Rod's ephemeral profile)
- Headless close removes that directory
- Visible `--open-browser` leaves the process and the directory (same as the old unspecified visible launch)
- Isolation tool for a second instance, `--force-headless` on another port, or a throwaway session
- When a non-ephemeral launch fails and Chrome's `SingletonLock` is still in the profile dir, snag reports `launch profile is in use` and suggests `snag --temp-profile --force-headless --port 9223 <url>` (keeps a non-default `--port`)

**Default (flag absent):**

- Persistent snag profile: Linux `$XDG_STATE_HOME/snag/chrome` when that variable is non-empty and absolute, otherwise `$HOME/.local/state/snag/chrome`; macOS `$HOME/Library/Application Support/snag/chrome`
- Relative or empty `XDG_STATE_HOME` is ignored (Base Directory spec)

#### Interaction Matrix

**Profile flags:**

| Combination | Behavior | Notes |
| ----------- | -------- | ----- |
| `--temp-profile` + `--user-data-dir` | **Error** (exit 1) | Mutually exclusive |
| `--temp-profile` + existing CDP browser | **Warning**, ignore flag | Cannot change the running browser's profile |
| `--temp-profile` + `--force-headless` | Works normally | Headless ephemeral launch; close deletes the dir |
| `--temp-profile` + `--open-browser` | Works normally | Visible ephemeral launch; process and dir left running |
| `--temp-profile` + `--kill-browser` | **Flag ignored** | Kill does not wipe profiles |
| `--temp-profile` + `--doctor` | Works normally | Doctor reports ephemeral launches; exit 0; no mkdir |
| `--temp-profile` + skill verb | **Flag ignored** | Skill modes do not launch a browser |
| `--temp-profile` + `--list-tabs` | **Warning**, ignored | Connects to existing browser |

**Warning message when connecting:**

- `"Warning: --temp-profile ignored when connecting to existing browser"` (CLI tab/list paths)
- `"Warning: --temp-profile ignored (browser already running with its own profile)"` (launch/connect path)

#### Examples

```bash
# Isolated headless fetch (profile deleted on close)
snag --temp-profile https://example.com

# Second instance while the default profile is in use
snag --open-browser
snag --temp-profile --force-headless --port 9223 https://example.com

# ERROR
snag --temp-profile --user-data-dir /tmp/snag https://example.com
```

## Implementation Details

**Location:**

- Flag definition: `internal/cli/root.go` (`init`, Cobra mutually exclusive with `--user-data-dir`)
- Launch and cleanup: `internal/browser` (`applyLaunchProfile`, `abandonLauncher`, `Close`)
- Doctor: `internal/doctor` (`ResolveLaunchProfile`, stat only)

**How it works:**

1. Cobra rejects `--temp-profile` with `--user-data-dir`
2. Launch does not set snag's XDG/darwin path; Rod uses a unique temp `user-data-dir`
3. Headless `Close()` and failed-launch teardown call `Cleanup()` (delete the temp dir)
4. Persistent XDG/darwin profile is never deleted
