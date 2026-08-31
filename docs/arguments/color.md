# `--color MODE`

**Status:** Complete (2026-08-30)

#### Validation Rules

**Allowed values:**

- `auto` (default), `always`, `never`
- Case-insensitive: `AUTO`, `Always` are valid
- Trimmed: leading and trailing whitespace removed
- Invalid value → Error: `invalid --color value "{value}" (allowed: auto, always, never)`
- Empty string after trim → same invalid-value error
- `--color` with no value → parse error from CLI framework

**Multiple `--color` flags:**

- Last wins (standard CLI behaviour, no error, no warning)

**Error messages:**

- Invalid value names the bad value and the allowed set (`auto`, `always`, `never`)
- Printed as `Error: …` on stderr, exit 1
- No `Code:`, `Fix:`, or ValidValues envelope

#### Behaviour

**Scope:**

- Colours stderr diagnostics only: logger prefixes and messages, and the `Error:` prefix
- stdout is never coloured by snag, including page content and `--list-tabs`, including `--color=always`

**Precedence (first match wins):**

1. `--color=never` → off. Overrides `NO_COLOR`, `FORCE_COLOR`, TTY, and `TERM=dumb`
2. `--color=always` → on. Overrides `NO_COLOR`, a non-TTY stderr, and `TERM=dumb`
3. `--color=auto` (default):
   - `NO_COLOR` present and non-empty → off. Wins over `FORCE_COLOR`. Empty `NO_COLOR=` is unset
   - `FORCE_COLOR` or `CLICOLOR_FORCE` truthy (`1`/`true`/`yes`, or any non-empty non-falsy value) → on. Overrides a non-TTY stderr and `TERM=dumb`
   - else `TERM=dumb` → off
   - else stderr is a TTY → on
   - else off

`CI` is not a colour signal. `FORCE_COLOR=0` / `false` / `no` are falsy.

When colour is off, prefixes stay as plain `✓` / `⚠` / `✗`. There is no `--emoji` flag.

**Basic usage:**

```bash
snag --color auto --verbose https://example.com
snag --color always --verbose https://example.com 2> log.txt
snag --color never --verbose https://example.com
NO_COLOR=1 snag --color always --verbose https://example.com
TERM=dumb FORCE_COLOR=1 snag --verbose --color auto https://example.com
```

#### Interaction Matrix

**All other flags:**

- Works with all other flags
- Does not change stdout content or `--list-tabs` listing
- Does not conflict with `--verbose` / `--debug`

**Examples with other flags:**

- `--color` + `--verbose` / `--debug` — colours those stderr diagnostics when the detector is on
- `--color` + `--doctor` / `--kill-browser` / skill verbs — works
- `--color=always` + `--list-tabs` — tab list on stdout stays uncoloured; errors on stderr may be coloured

#### Examples

**Valid:**

```bash
snag --color auto https://example.com
snag --color always --verbose https://example.com
snag --color never --verbose https://example.com
snag --color ALWAYS --verbose https://example.com   # value is case-insensitive
```

**Invalid:**

```bash
snag --color rainbow                                # ERROR: allowed auto, always, never
snag --color                                        # parse error: flag needs an argument
```

#### Implementation Details

**Location:**

- Flag definition and validation: `internal/cli/root.go` (`init`, `PersistentPreRunE`)
- Detector: `internal/logger` (`UseColor`, `ResolveColor`)
- Applied on the default logger in `PersistentPreRunE`; `Error:` prefix via `cli.FormatError` from the parsed `--color` value (auto if that value is invalid)

**Processing:**

1. Trim and lowercase the flag value
2. Reject values other than `auto`, `always`, `never` (exit 1); do not apply the invalid mode
3. Detect stderr TTY (not stdout)
4. Apply the precedence table above once per invocation
5. Install a logger with that colour bool in `PersistentPreRunE`; the `Error:` prefix uses the same resolved mode, including parse errors when `--color` was parsed
