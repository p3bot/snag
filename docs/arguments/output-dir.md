# `--output-dir DIRECTORY` / `-d`

**Status:** Complete (2025-10-22)

#### Validation Rules

**Path Handling:**

- Both relative and absolute paths supported
- Paths with spaces supported (user must quote in shell: `-d "my output dir"`)
- Path validation using Go's `os.Stat()` and `fileInfo.IsDir()`
- Empty string → Use current working directory (pwd)
- Relative paths resolved relative to current directory

**Directory Validation:**

- Directory must exist → Error: `"Output directory does not exist: {path}"`
- Path exists but is a file (not directory) → Error: `"Output directory path is a file, not a directory: {path}"`
- Permission denied (no write access) → Error: `"Cannot write to output directory: permission denied"`

**Multiple Directory Flags:**

- Multiple `-d` flags → **Last wins** (standard CLI behavior, no error, no warning)

**Error Messages:**

- Directory doesn't exist: `"Output directory does not exist: {path}"`
- Path is file: `"Output directory path is a file, not a directory: {path}"`
- Permission denied: `"Cannot write to output directory: permission denied"`
- Empty after trim: Uses current directory (no error)

#### Behavior

**Basic Usage:**

```bash
snag https://example.com -d ./output
```

- Fetches URL content
- Generates filename automatically: `yyyy-mm-dd-hhmmss-{page-title}-{slug}.{ext}`
- Writes to `./output/{generated-filename}`
- Creates file if it doesn't exist

**Filename Generation:**

```bash
snag https://example.com -d ./docs
# Creates: ./docs/2025-10-22-214530-example-domain.md
```

- Format: `yyyy-mm-dd-hhmmss-{page-title}-{slug}.{ext}`
- Extension matches `--format` flag
- Page title used for slug (sanitized, lowercased, hyphens)

**Collision Handling:**

```bash
snag https://example.com -d ./output  # file.md
snag https://example.com -d ./output  # file-1.md (if collision)
snag https://example.com -d ./output  # file-2.md (if collision)
```

- Append counter suffix: `-1`, `-2`, `-3`, etc.
- Ensures no file overwriting

**Empty String Behavior:**

```bash
snag https://example.com -d ""
# Uses current working directory (pwd)
```

#### Interaction Matrix

**Output Destination Conflicts:**

| Combination                    | Behavior      | Rationale                                                   |
| ------------------------------ | ------------- | ----------------------------------------------------------- |
| `-d directory/` + `-o file.md` | **Error**     | Mutually exclusive output destinations                      |
| Multiple `-d` flags            | **Last wins** | Standard CLI behavior (e.g., `-d dir1 -d dir2` uses `dir2`) |

**Error messages:**

- `-d` + `-o`: `"Cannot use both --output and --output-dir"`

**Content Source Interactions:**

| Combination                    | Behavior       | Notes                                      |
| ------------------------------ | -------------- | ------------------------------------------ |
| `-d ./dir` + single `<url>`    | Works normally | Auto-generated filename in directory       |
| `-d ./dir` + multiple `<url>`  | Works normally | Each URL gets separate auto-generated file |
| `-d ./dir` + `--url-file`      | Works normally | Each URL from file gets separate file      |
| `-d ./dir` + `--tab <pattern>` | Works normally | Tab content saved with auto-generated name |
| `-d ./dir` + `--all-tabs`      | Works normally | Each tab gets separate auto-generated file |

**Special Operation Modes:**

| Combination                                     | Behavior                  | Notes                                                                                                                     |
| ----------------------------------------------- | ------------------------- | ------------------------------------------------------------------------------------------------------------------------- |
| `-d ./dir` + `--list-tabs`                      | `--list-tabs` overrides   | `--list-tabs` overrides all other options                                                                                 |
| `-d ./dir` + `--doctor`                         | **Flag ignored**          | Doctor overrides, diagnostics only                                                                                        |
| `-d ./dir` + `--kill-browser`                   | **Flag ignored**          | No content to save                                                                                                        |
| `-d ./dir` + `--open-browser` (no URL/url-file) | **Warning**, flag ignored | `"Warning: --output-dir ignored with --open-browser (no content fetching)"`                                               |
| `-d ./dir` + `--open-browser` + `<url>`         | **Warning**, flag ignored | `"Warning: --output-dir ignored with --open-browser (no content fetching)"` - Opens browser, navigates to URL, exits snag |

**Format Combinations:**

All format combinations work normally:

| Scenario                 | Behavior                | Output Extension |
| ------------------------ | ----------------------- | ---------------- |
| `-d ./dir --format md`   | Auto-save as Markdown   | `.md`            |
| `-d ./dir --format html` | Auto-save as HTML       | `.html`          |
| `-d ./dir --format text` | Auto-save as plain text | `.txt`           |
| `-d ./dir --format pdf`  | Auto-save as PDF        | `.pdf`           |
| `-d ./dir --format png`  | Auto-save as PNG        | `.png`           |

**Compatible Flags:**

All these flags work normally with `-d`:

- ✅ `--format` - Apply format (extension matches format)
- ✅ `--timeout` - Apply timeout to page load
- ✅ `--wait-for` - Wait for selector before extraction
- ✅ `--close-tab` - Close tab after fetching
- ✅ `--force-headless` - Force headless browser mode
- ✅ `--verbose` / `--quiet` / `--debug` - Logging levels
- ✅ `--user-agent` - Custom user agent (applies to all new pages)
- ✅ `--user-data-dir` - Custom browser profile
- ✅ `--port` - Remote debugging port

#### Examples

**Valid:**

```bash
snag https://example.com -d ./output                 # Basic usage
snag https://example.com -d ./docs/pages             # Nested directory
snag https://example.com -d /tmp/snag                # Absolute path
snag https://example.com -d "my output dir"          # Path with spaces
snag https://example.com -d ""                       # Current directory
snag https://example.com -d ./output --format html   # HTML format
snag https://example.com -d ./output --format pdf    # PDF format
snag url1 url2 url3 -d ./batch                       # Multiple URLs
snag --url-file urls.txt -d ./results                # URL file
snag --tab 1 -d ./tabs                               # From existing tab
snag --all-tabs -d ./all                             # All tabs
snag https://example.com -d ./out --timeout 60       # With timeout
snag https://example.com -d ./out --wait-for ".content"  # With wait
snag https://example.com -d ./dir1 -d ./dir2         # Uses ./dir2 (last wins)
```

**Invalid:**

```bash
snag https://example.com -d /nonexistent/dir         # ERROR: Directory doesn't exist
snag https://example.com -d ./existing-file.txt      # ERROR: Path is file, not directory
snag https://example.com -d /root/restricted         # ERROR: Permission denied
snag https://example.com -d ./dir -o file.md         # ERROR: -d and -o conflict
snag --list-tabs -d ./tabs                           # --output-dir ignored, lists tabs from existing browser
snag --open-browser -d ./output                      # OK but -d ignored (nothing to fetch)
```

#### Implementation Details

**Location:**

- Flag definition: `internal/cli/root.go` (`init`)
- Handler logic: `internal/cli/handlers.go`
- Path validation: `internal/validate` (`Directory`)
- Filename generation: `internal/output` (`GenerateFilename`)

**Processing Flow:**

1. Validate output directory path (exists, is directory, writable)
2. Check for conflicts (`-o`, multiple `-d`)
3. Fetch content from source(s)
4. Generate filename(s) with timestamp and page title
5. Check for collisions → append counter if needed
6. Write to directory

**Filename Generation:**

- Function: `generateOutputFilename(pageTitle, format, outputDir string)`
- Timestamp: Single timestamp for batch operations
- Title slug: Sanitized, lowercase, hyphens for spaces
- Extension: Matches format (`.md`, `.html`, `.txt`, `.pdf`, `.png`)
- Collision resolution: Append `-1`, `-2`, etc.

**Special Behaviors:**

- Empty string for directory → Resolves to current working directory
- Relative paths → Resolved relative to current directory
- `--open-browser` without URL/url-file → Directory flag ignored (nothing to save)
