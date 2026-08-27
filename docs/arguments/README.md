# Argument Handling Reference

**Purpose:** Complete specification of all argument/flag combinations and their interactions.

**Status:** All arguments analyzed and documented ✅ | All inconsistencies resolved ✅

**Last Updated:** 2025-10-26

---

## Individual Argument Documentation

### Completed Arguments ✅

- [**`<url>`** - Positional URL argument](./url.md)
- [**`--url-file FILE`** - Read URLs from file](./url-file.md)
- [**`--output FILE`** / **`-o`** - Save to specific file](./output.md)
- [**`--output-dir DIRECTORY`** / **`-d`** - Save with auto-generated name](./output-dir.md)
- [**`--format FORMAT`** / **`-f`** - Output format selection](./format.md)
- [**`--timeout SECONDS`** - Page load timeout](./timeout.md)
- [**`--wait-for SELECTOR`** / **`-w`** - Wait for CSS selector](./wait-for.md)
- [**`--port PORT`** / **`-p`** - Remote debugging port](./port.md)
- [**`--close-tab`** / **`-c`** - Close tab after fetching](./close-tab.md)
- [**`--force-headless`** - Force headless mode](./force-headless.md)
- [**`--open-browser`** / **`-b`** - Open visible browser](./open-browser.md)
- [**`--list-tabs`** / **`-l`** - List all open tabs](./list-tabs.md)
- [**`--tab PATTERN`** / **`-t`** - Fetch from existing tab](./tab.md)
- [**`--verbose`** - Verbose logging](./verbose.md)
- [**`--debug`** - Debug logging](./debug.md)
- [**`--help`** / **`-h`** - Show help](./help.md)
- [**`--version`** / **`-v`** - Show version](./version.md)
- [**`--all-tabs`** / **`-a`** - Process all open tabs](./all-tabs.md)
- [**`--user-agent STRING`** - Custom user agent](./user-agent.md)
- [**`--user-data-dir DIRECTORY`** - Custom browser profile](./user-data-dir.md)

### Advanced Topics

- [**Validation Rules**](./validation.md) - Validation order and cross-cutting rules

---

## All Arguments and Flags

### Positional Arguments

| Argument | Type   | Description                           |
| -------- | ------ | ------------------------------------- |
| `<url>`  | String | URL to fetch (supports multiple URLs) |

### Output Control Flags

| Flag           | Aliases | Type   | Default | Description                                       |
| -------------- | ------- | ------ | ------- | ------------------------------------------------- |
| `--output`     | `-o`    | String | -       | Save to specific file path                        |
| `--output-dir` | `-d`    | String | -       | Save with auto-generated name to directory        |
| `--format`     | `-f`    | String | `md`    | Output format: `md`, `html`, `text`, `pdf`, `png` |

### Page Loading Flags

| Flag         | Aliases | Type   | Default | Description                             |
| ------------ | ------- | ------ | ------- | --------------------------------------- |
| `--timeout`  | -       | Int    | `30`    | Page load timeout in seconds            |
| `--wait-for` | `-w`    | String | -       | Wait for CSS selector before extraction |

### Browser Control Flags

| Flag               | Aliases | Type   | Default | Description                                |
| ------------------ | ------- | ------ | ------- | ------------------------------------------ |
| `--port`           | `-p`    | Int    | `9222`  | Chrome remote debugging port               |
| `--close-tab`      | `-c`    | Bool   | `false` | Close browser tab after fetching           |
| `--force-headless` | -       | Bool   | `false` | Force headless mode                        |
| `--open-browser`   | `-b`    | Bool   | `false` | Open browser in visible state              |
| `--list-tabs`      | `-l`    | Bool   | `false` | List all open tabs                         |
| `--tab`            | `-t`    | String | -       | Fetch from existing tab (index or pattern) |
| `--all-tabs`       | `-a`    | Bool   | `false` | Process all open tabs                      |
| `--user-data-dir`  | -       | String | -       | Custom browser profile directory           |

### URL Input Flags

| Flag         | Aliases | Type   | Default | Description                        |
| ------------ | ------- | ------ | ------- | ---------------------------------- |
| `--url-file` | -       | String | -       | Read URLs from file (one per line) |

### Logging Flags

| Flag        | Aliases | Type | Default | Description                                    |
| ----------- | ------- | ---- | ------- | ---------------------------------------------- |
| `--verbose` | -       | Bool | `false` | Enable verbose logging (fetch/connection/batch) |
| `--debug`   | -       | Bool | `false` | Enable debug output                            |

### Request Control Flags

| Flag           | Aliases | Type   | Default | Description              |
| -------------- | ------- | ------ | ------- | ------------------------ |
| `--user-agent` | -       | String | -       | Custom user agent string |

---

## Quick Reference Matrix

### Mode Selection (Mutually Exclusive Groups)

These determine the primary operation mode:

| Combination                | Behavior                           |
| -------------------------- | ---------------------------------- |
| No flags, no URL           | ❌ ERROR: URL required             |
| `<url>` only               | Fetch URL to stdout                |
| `--list-tabs` (any args)   | List tabs, exit, ignore other args |
| `--tab <pattern>`          | Fetch from existing tab            |
| `--tab` + `<url>`          | ❌ ERROR: cannot mix               |
| `--all-tabs`               | Fetch all tabs to files            |
| `--all-tabs` + `<url>`     | ❌ ERROR: cannot mix               |
| `--open-browser` only      | Open browser, keep open, no fetch  |
| `--open-browser` + `<url>` | Open browser, navigate, NO fetch   |
| `<url> <url> <url>`        | Batch fetch multiple URLs          |
| `--url-file urls.txt`      | Fetch URLs from file               |

### Output Destination (Mutually Exclusive)

| Combination                        | Behavior                               |
| ---------------------------------- | -------------------------------------- |
| No output flags                    | Content to stdout                      |
| `-o file.md`                       | Content to specific file               |
| `-d ./dir`                         | Content to auto-generated file in dir  |
| `-o` + `-d`                        | ❌ ERROR: cannot use both              |
| Multiple URLs + `-o`               | ❌ ERROR: use `-d` instead             |
| Multiple URLs + `-d`               | ✅ Each URL gets auto-generated name   |
| Multiple URLs, no output flags     | ✅ Auto-save to current dir            |
| Binary format (PDF/PNG), no output | Auto-generates filename in current dir |

### Browser Mode (Mutually Exclusive)

| Combination                           | Behavior                       |
| ------------------------------------- | ------------------------------ |
| No browser flags                      | Auto-detect or launch headless |
| `--force-headless`                    | Always headless                |
| `--open-browser`                      | Open visible browser           |
| `--open-browser` + `--force-headless` | ❌ ERROR                       |

### Logging Level (Mutually Exclusive)

| Combination             | Result   |
| ----------------------- | -------- |
| No logging flags        | Normal (progress silent; warnings, errors, and result announcements print) |
| `--verbose`             | Verbose  |
| `--debug`               | Debug    |
| `--debug` + `--verbose` | ❌ ERROR |

---

## Mutually Exclusive Combinations

### Group 1: Operation Mode

**Only ONE of these can be active:**

1. Fetch URL: `<url>`
2. Fetch multiple URLs: `<url> <url>` or `--url-file`
3. Fetch from tab: `--tab <pattern>`
4. Fetch all tabs: `--all-tabs`
5. List tabs: `--list-tabs`
6. Open browser only: `--open-browser` (no URL)

**Conflict Matrix:**

|                             | URL     | Multi-URL | --tab   | --all-tabs | --list-tabs | --open-browser (no URL) |
| --------------------------- | ------- | --------- | ------- | ---------- | ----------- | ----------------------- |
| **URL**                     | ✅      | N/A       | ❌      | ❌         | ❌          | N/A                     |
| **Multi-URL**               | N/A     | ✅        | ❌      | ❌         | ❌          | N/A                     |
| **--tab**                   | ❌      | ❌        | ✅      | ❌         | ❌          | N/A                     |
| **--all-tabs**              | ❌      | ❌        | ❌      | ✅         | ❌          | N/A                     |
| **--list-tabs**             | Ignores | Ignores   | Ignores | Ignores    | ✅          | Ignores                 |
| **--open-browser (no URL)** | N/A     | N/A       | N/A     | N/A        | N/A         | ✅                      |

**Error Messages:**

- `--tab` + URL: `"Cannot use both --tab and URL arguments (mutually exclusive content sources)"`
- `--all-tabs` + URL: `"Cannot use both --all-tabs and URL arguments (mutually exclusive content sources)"`

### Group 2: Output Destination

**Only ONE of these can be active:**

1. Stdout (default, no flags)
2. Specific file: `-o file.md`
3. Auto-generated in directory: `-d ./dir`

**Conflict:**

- `-o` + `-d`: ❌ ERROR: `"Cannot use both --output and --output-dir"`

### Group 3: Browser Mode

**Only ONE of these can be active:**

1. Auto-detect (default, no flags)
2. Force headless: `--force-headless`
3. Open visible: `--open-browser`

---

## Mode-Based Behavior

### Mode 1: Fetch Single URL (Current)

**Invocation:** `snag <url> [flags]`

**Compatible Flags:**

- ✅ `-o, -d` - Output control
- ✅ `--format` - Format selection
- ✅ `--timeout` - Load timeout
- ✅ `--wait-for` - Wait for selector
- ✅ `--port` - Remote debugging port
- ✅ `--close-tab` - Close after fetch
- ✅ `--force-headless` - Browser mode
- ✅ `--open-browser` - Open in visible browser
- ✅ `--user-agent` - Custom UA
- ✅ Logging flags

**Incompatible Flags:**

- ❌ `--tab` - Conflicts with URL
- ❌ `--all-tabs` - Conflicts with URL

**Output Behavior:**

- No `-o` or `-d`: → stdout (unless PDF/PNG)
- `-o file.md`: → file
- `-d ./dir`: → auto-generated filename in dir
- PDF/PNG without output flag: → auto-generated filename in current dir

### Mode 2: Fetch Multiple URLs

**Invocation:** `snag <url1> <url2> <url3>` or `snag --url-file urls.txt [<url4> ...]`

**Compatible Flags:**

- ✅ `-d` - Output directory (default: current dir)
- ✅ `--format` - Applied to all URLs
- ✅ `--timeout` - Applied to each URL
- ✅ `--wait-for` - Applied to each page
- ✅ `--close-tab` - Close each tab after fetching
- ✅ `--port` - Remote debugging port
- ✅ `--force-headless` - Browser mode
- ✅ `--open-browser` - Opens all URLs in tabs, NO fetch
- ✅ `--user-agent` - Applied to all
- ✅ Logging flags

**Incompatible Flags:**

- ❌ `-o` - Ambiguous for multiple outputs
- ❌ `--tab` - Conflicts with URLs
- ❌ `--all-tabs` - Conflicts with URLs

**Output Behavior:**

- Always saves to files (never stdout)
- No `-d`: → auto-generated names in current dir (`.`)
- `-d ./dir`: → auto-generated names in specified dir

**Error Behavior:**

- Continue-on-error (process all URLs)
- Summary: "X succeeded, Y failed"
- Exit 0 if all succeed, exit 1 if any fail

### Mode 3: Fetch from Tab

**Invocation:** `snag --tab <pattern>`

See [tab.md](./tab.md) for complete details.

**Compatible Flags:**

- ✅ `-o, -d` - Output control
- ✅ `--format` - Format selection
- ✅ `--timeout` - Applies to `--wait-for` if present (warns if no --wait-for)
- ✅ `--wait-for` - Wait for selector (supports automation with persistent browser)
- ✅ `--port` - Remote debugging port
- ✅ `--user-agent` - Warned and ignored (tab already open with its own user agent)
- ✅ `--close-tab` - Honored (closes the tab after fetching)
- ✅ Logging flags

**Incompatible Flags:**

- ❌ `<url>` - Conflicts with tab (mutually exclusive content sources)
- ❌ `--all-tabs` - Use one or the other (mutually exclusive)
- ❌ `--force-headless` - Error (tab requires existing browser)
- ⚠️ `--open-browser` - Warning, flag ignored (no content fetching)

### Mode 4: Fetch All Tabs

**Invocation:** `snag --all-tabs`

**Compatible Flags:**

- ✅ `-d` - Output directory (optional, defaults to current dir)
- ✅ `--format` - Applied to all tabs
- ✅ `--timeout` - Applies to `--wait-for` if present (warns if no --wait-for)
- ✅ `--wait-for` - Wait for same selector in each tab before fetching
- ✅ `--port` - Remote debugging port
- ✅ `--close-tab` - Close each tab after fetching; last tab closes browser
- ✅ `--user-data-dir` - Custom browser profile
- ✅ Logging flags

**Incompatible Flags:**

- ❌ `<url>` - Conflicts with all-tabs (mutually exclusive content sources)
- ❌ `-o` - Multiple outputs (use `-d`)
- ❌ `--tab` - Use one or the other (mutually exclusive)
- ❌ `--force-headless` - Error (tabs require existing browser)
- ⚠️ `--open-browser` - Warning, flag ignored (no content fetching)
- ⚠️ `--user-agent` - Warning, ignored (tabs already open with their own user agents)

### Mode 5: List Tabs

**Invocation:** `snag --list-tabs`

See [list-tabs.md](./list-tabs.md) for complete details.

**Compatible Flags:**

- ✅ `--port` - Remote debugging port
- ✅ Logging flags

**Incompatible Flags:**

- All other flags are silently ignored (standalone mode like `--help`)

### Mode 6: Open Browser Only

**Invocation:** `snag --open-browser` (no URL)

See [open-browser.md](./open-browser.md) for complete details.

**Compatible Flags:**

- ✅ `--port` - Remote debugging port
- ✅ `--user-data-dir` - Custom browser profile
- ✅ Logging flags

**Incompatible Flags:**

- Most flags are irrelevant (no fetching)

---

## Output Routing Rules

### Stdout vs File Output

| Scenario                      | Output Destination                     | Notes                       |
| ----------------------------- | -------------------------------------- | --------------------------- |
| Single URL, no flags          | stdout                                 | Default behavior            |
| Single URL, `-o file.md`      | `file.md`                              | Specified file              |
| Single URL, `-d ./dir`        | `./dir/{auto-generated}.md`            | Auto-generated name         |
| Single URL, PDF/PNG, no flags | `./{auto-generated}.pdf`               | Binary formats never stdout |
| Multiple URLs, no flags       | `./{auto-generated}.md` (each)         | Batch auto-save             |
| Multiple URLs, `-d ./dir`     | `./dir/{auto-generated}.md` (each)     | Custom directory            |
| `--tab`, no flags             | stdout                                 | Same as single URL          |
| `--tab`, `-o file.md`         | `file.md`                              | Specified file              |
| `--all-tabs`                  | `-d` or `./{auto-generated}.md` (each) | Always files                |
| `--list-tabs`                 | stdout (tab list only)                 | Informational output        |

### Filename Generation Format

**Pattern:** `yyyy-mm-dd-hhmmss-{page-title}-{slug}.{ext}`

**Examples:**

- `2025-10-22-124752-example-domain.md`
- `2025-10-22-124753-go-programming-language.html`
- `2025-10-22-124754-github-p3bot-snag.pdf`

**Extension Mapping:**

- `md` → `.md`
- `html` → `.html`
- `text` → `.txt`
- `pdf` → `.pdf`
- `png` → `.png`

**Conflict Resolution:**

- If file exists: append `-1`, `-2`, etc.
- Example: `file.md` → `file-1.md` → `file-2.md`

---

## Examples by Use Case

### Basic Fetching

```bash
snag https://example.com                    # Stdout
snag https://example.com -o page.md         # To file
snag https://example.com -d ./docs          # Auto-generated name
snag https://example.com --format html      # HTML output
snag https://example.com --format pdf       # PDF (auto-saves)
```

### Batch Processing

```bash
snag url1 url2 url3                         # Multiple URLs
snag --url-file urls.txt                    # From file
snag --url-file urls.txt url4 url5          # Combined
snag --url-file urls.txt -d ./results       # Custom directory
```

### Tab Management

```bash
snag --list-tabs                            # List tabs
snag --tab 1                                # Fetch tab 1
snag --tab "github"                         # Pattern match
snag --all-tabs -d ./tabs                   # All tabs
```

### Browser Control

```bash
snag --open-browser                         # Open browser only
snag --open-browser https://example.com     # Open + fetch (current)
snag --force-headless https://example.com   # Force headless
```

### Advanced Options

```bash
snag https://example.com --wait-for ".content"  # Wait for selector
snag https://example.com --timeout 60           # Custom timeout
snag https://example.com --user-agent "Custom"  # Custom UA
snag https://example.com --port 9223            # Custom port
```

---

## Notes

1. **Exit Codes:**

   - `0`: Success (all operations succeeded)
   - `1`: Error (any operation failed)
   - `130`: SIGINT (Ctrl+C)
   - `143`: SIGTERM

2. **Binary Formats:**

   - PDF and PNG formats always save to file (never stdout)
   - Auto-generated filenames used if no output flag specified

3. **Browser Connection:**

   - Tab features require existing browser with remote debugging enabled
   - Use `snag --open-browser` to start persistent browser

4. **Pattern Matching:**
   - Tab patterns support: index (1-based), exact URL, substring, regex
   - Case-insensitive matching for all pattern types
