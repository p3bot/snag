# `<url>` Positional Argument

**Status:** Complete (2025-10-22)

## Validation Rules

**Protocol Handling:**

- Auto-add `https://` if no protocol is present
- Valid schemes: `http`, `https`, `file` only
- Invalid schemes (e.g., `ftp://`, `data:`) → Error in validation

**URL Validation:**

- Validate using Go's `url.Parse()` before passing to browser
- Must have valid URL characters only
- Localhost and private IPs are allowed (e.g., `http://localhost:3000`, `http://192.168.1.1`)
- Connection failures are handled by browser, not validation

**Error Messages:**

- Invalid URL format: `"Invalid URL format: {url}"`
- Invalid scheme: `"Invalid URL scheme '{scheme}'. Supported: http, https, file"`
- URL + `--tab`: `"Cannot use both --tab and URL arguments (mutually exclusive content sources)"`
- URL + `--all-tabs`: `"Cannot use both --all-tabs and URL arguments (mutually exclusive content sources)"`

## Multiple URLs Behavior

**With No Output Flag:**

```bash
snag https://example.com https://google.com
```

- Behavior: Auto-generate filenames in current directory
- Browser mode: Headless if no browser open
- Each URL gets separate file: `yyyy-mm-dd-hhmmss-{page-title}-{slug}.{ext}`

**With `--output FILE`:**

```bash
snag -o output.md https://example.com https://google.com
```

- Behavior: **Error** - Cannot combine multiple sources into single output file
- Error message: `"Cannot use --output with multiple content sources. Use --output-dir instead"`

**Error Messages:**

- Multiple URLs + `-o`: `"Cannot use --output with multiple content sources. Use --output-dir instead"`

**With `--output-dir DIR`:**

```bash
snag -d output/ https://example.com https://google.com
```

- Behavior: Auto-generate separate filenames in specified directory
- Browser mode: Headless if no browser open
- Each URL gets separate file in directory

## Interaction Matrix

**Content Source Conflicts:**

| Combination              | Behavior                | Rationale                                 |
| ------------------------ | ----------------------- | ----------------------------------------- |
| `<url>` + `--url-file`   | **Merge** both sources  | Allow combining CLI URLs with file URLs   |
| `<url>` + `--tab`        | **Error**               | Mutually exclusive content sources        |
| `<url>` + `--all-tabs`   | **Error**               | Mutually exclusive content sources        |
| `<url>` + `--list-tabs`  | `--list-tabs` overrides | `--list-tabs` overrides all other options |
| `<url>` + `--doctor`     | **Flag ignored**        | Doctor overrides, diagnostics only        |
| `<url>` + `--kill-browser` | **Error**             | Conflicting operations                    |

**Browser Mode:**

| Combination                  | Behavior                                                      | Notes                                   |
| ---------------------------- | ------------------------------------------------------------- | --------------------------------------- |
| `<url>` + `--open-browser`   | Navigate to URL, **do not fetch** content, leave browser open | Opens URL in tab for manual interaction |
| `<url>` + `--force-headless` | Force headless mode even if browser already open              | Override auto-detection                 |

**Output Control:**

| Combination                 | Behavior       | Notes                               |
| --------------------------- | -------------- | ----------------------------------- |
| `<url>` + `--format`        | Works normally | Apply format to fetched content     |
| `<url>` + `--timeout`       | Works normally | Apply timeout to page load          |
| `<url>` + `--wait-for`      | Works normally | Wait for selector after navigation  |
| `<url>` + `--port`          | Works normally | Use specified remote debugging port |
| `<url>` + `--user-agent`    | Works normally | Set user agent for new page         |
| `<url>` + `--user-data-dir` | Works normally | Use custom browser profile          |

**Special Behaviors:**

| Combination             | Behavior                                              | Notes                                        |
| ----------------------- | ----------------------------------------------------- | -------------------------------------------- |
| `<url>` + `--close-tab` | Close tab if browser visible; **ignored** if headless | Headless mode already closes tabs by default |

**Logging Flags:**

- `--verbose`, `--debug`: All work normally with `<url>`

## Examples

**Valid:**

```bash
snag https://example.com                           # Fetch to stdout
snag example.com                                   # Auto-adds https://
snag http://localhost:3000                         # Local development
snag file:///path/to/file.html                     # Local file
snag https://example.com -o page.md                # Save to file
snag https://example.com --open-browser            # Open URL, no fetch
```

**Invalid:**

```bash
snag ftp://example.com                             # ERROR: Invalid scheme
snag https://example.com --tab 1                   # ERROR: Conflicting sources
snag https://example.com --all-tabs                # ERROR: Conflicting sources
snag https://example.com --list-tabs               # URL ignored, lists tabs from existing browser
snag url1 url2 -o out.md                           # ERROR: Multiple URLs need -d
```
