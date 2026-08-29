# `--output FILE` / `-o`

**Status:** Complete (2025-10-22)

#### Validation Rules

**Path Handling:**

- Both relative and absolute paths supported
- Paths with spaces supported (user must quote in shell: `-o "my file.md"`)
- Parent directory must exist → Error: `"Output path invalid: parent directory does not exist"`
- Permission denied → Error: `"Failed to write output file: permission denied"`
- Directory provided instead of file → Error: `"Output path is a directory, not a file"`
- Empty string → Error: `"Output file path cannot be empty"`

**File Validation:**

- File existence check before fetching (only for permission/path validation)
- File overwrite behavior: Silently overwrite (standard Unix `cp` behavior)
- Read-only existing file → Error: `"Cannot write to read-only file: {path}"`

**Multiple Output Flags:**

- Multiple `-o` flags → **Last wins** (standard CLI behavior, no error, no warning)

**Error Messages:**

- Invalid path: `"Output path invalid: parent directory does not exist"`
- Permission denied: `"Failed to write output file: permission denied"`
- Directory provided: `"Output path is a directory, not a file"`
- Empty string: `"Output file path cannot be empty"`
- Read-only file: `"Cannot write to read-only file: {path}"`

#### Behavior

**Basic Usage:**

```bash
snag https://example.com -o output.md
```

- Fetches URL content
- Writes to specified file path
- Overwrites file if it exists
- Creates file if it doesn't exist (parent directory must exist)

**Format Interactions:**

```bash
snag https://example.com -o file.md --format html
```

- **Warning message:** `"Warning: Writing HTML format to file with .md extension"`
- User gets what they requested (no error)
- Applies to all mismatched extensions

**Binary Formats:**

```bash
snag https://example.com -o output.pdf --format pdf   # Normal
snag https://example.com -o file.md --format pdf     # Warning
```

- PDF/PNG to text extension → **Warning message:** `"Warning: Writing PDF format to file with .md extension"`
- User intent honored (file written as requested)

**No Extension:**

```bash
snag https://example.com -o myfile --format markdown
```

- **Warning message:** `"Warning: Output file has no extension, expected .md for markdown format"`
- File created without extension as requested

#### Interaction Matrix

**Content Source Interactions:**

| Combination                     | Behavior       | Rationale                                                   |
| ------------------------------- | -------------- | ----------------------------------------------------------- |
| `-o file.md` + single `<url>`   | Works normally | Standard single-file output                                 |
| `-o file.md` + multiple `<url>` | **Error**      | Ambiguous: cannot concatenate multiple sources              |
| `-o file.md` + `--url-file`     | **Error**      | Ambiguous: cannot concatenate multiple sources              |
| Multiple `-o` flags             | **Last wins**  | Standard CLI behavior (e.g., `-o a.md -o b.md` uses `b.md`) |

**Error messages:**

- Multiple URLs + `-o`: `"Cannot use --output with multiple content sources. Use --output-dir instead"`
- `--url-file` + `-o`: `"Cannot use --output with multiple content sources. Use --output-dir instead"`

**Output Destination Conflicts:**

| Combination                    | Behavior  | Rationale                              |
| ------------------------------ | --------- | -------------------------------------- |
| `-o file.md` + `-d directory/` | **Error** | Mutually exclusive output destinations |

**Error message:**

- `"Cannot use both --output and --output-dir"`

**Special Operation Modes:**

| Combination                                   | Behavior                  | Notes                                     |
| --------------------------------------------- | ------------------------- | ----------------------------------------- |
| `-o file.txt` + `--list-tabs`                 | `--list-tabs` overrides   | `--list-tabs` overrides all other options |
| `-o file.md` + `--doctor`                     | **Flag ignored**          | Doctor overrides, diagnostics only        |
| `-o file.md` + `--kill-browser`               | **Flag ignored**          | No content to save                        |
| `-o file.md` + skill verb                     | **Flag ignored**          | Skill modes do not fetch                  |
| `-o file.md` + `--open-browser` (no URL)      | **Warning**, flag ignored | No content fetching                       |
| `-o file.md` + `--tab <pattern>`              | Works normally            | Fetch from tab, save to file              |
| `-o file.md` + `--tab <pattern>` (no browser) | **Error**                 | Tab requires running browser              |

**Warning messages:**

- `--open-browser` only: `"Warning: --output ignored with --open-browser (no content fetching)"`

**Error messages:**

- `--tab` no browser: `"No browser instance running with remote debugging"`

**Format Combinations:**

All format combinations work, with warnings for mismatches:

| Scenario                   | Behavior                    | Warning |
| -------------------------- | --------------------------- | ------- |
| `-o file.md --format html` | Write HTML to .md file      | ⚠️ Yes  |
| `-o file.pdf --format pdf` | Write PDF to .pdf file      | No      |
| `-o file.md --format pdf`  | Write PDF bytes to .md      | ⚠️ Yes  |
| `-o myfile` (no extension) | Write to extensionless file | ⚠️ Yes  |

**File Overwriting:**

| Scenario                | Behavior           |
| ----------------------- | ------------------ |
| File doesn't exist      | Create new file    |
| File exists (writable)  | Silently overwrite |
| File exists (read-only) | **Error**          |

**Compatible Flags:**

All these flags work normally with `-o`:

- ✅ `--format` - Apply format (with extension mismatch warnings)
- ✅ `--timeout` - Apply timeout to page load
- ✅ `--wait-for` - Wait for selector before extraction
- ✅ `--close-tab` - Close tab after fetching
- ✅ `--force-headless` - Force headless browser mode
- ✅ `--verbose` / `--debug` - Logging levels
- ✅ `--user-agent` - Custom user agent
- ✅ `--user-data-dir` - Custom browser profile
- ✅ `--port` - Remote debugging port

#### Examples

**Valid:**

```bash
snag https://example.com -o page.md                  # Basic usage
snag https://example.com -o ./output/page.md         # Relative path
snag https://example.com -o /tmp/page.md             # Absolute path
snag https://example.com -o "my file.md"             # Path with spaces
snag https://example.com -o page.html --format html  # Matching extension
snag https://example.com -o page.pdf --format pdf    # PDF format
snag --tab 1 -o content.md                           # From existing tab
snag https://example.com -o page.md --timeout 60     # With timeout
snag https://example.com -o page.md --wait-for ".content"  # With wait
snag https://example.com -o out.md -o out2.md        # Uses out2.md (last wins)
```

**Invalid:**

```bash
snag url1 url2 -o out.md                             # ERROR: Multiple URLs
snag --url-file urls.txt -o out.md                   # ERROR: Multiple sources
snag https://example.com -o file.md -d ./dir         # ERROR: -o and -d conflict
snag --list-tabs -o tabs.txt                         # --output ignored, lists tabs from existing browser
snag --tab 1 -o file.md                              # ERROR: No browser running
snag https://example.com -o /root/file.md            # ERROR: Permission denied
snag https://example.com -o /nonexistent/dir/file.md # ERROR: Parent doesn't exist
snag https://example.com -o ./existing-directory/    # ERROR: Path is directory
snag https://example.com -o ""                       # ERROR: Empty string
```

**With Warnings:**

```bash
snag --open-browser -o file.md                       # ⚠️ Warning: --output ignored (no content fetching)
snag https://example.com -o file.md --format html    # ⚠️ Extension mismatch
snag https://example.com -o file.txt --format pdf    # ⚠️ Binary to text ext
snag https://example.com -o myfile                   # ⚠️ No extension
```

#### Implementation Details

**Location:**

- Flag definition: `internal/cli/root.go` (`init`)
- Handler logic: `internal/cli/root.go` (`runCobra`)
- Path validation: `internal/validate` (`OutputPath`)
- File writing: Various handler functions

**Processing Flow:**

1. Validate output path (parent exists, not a directory, permissions)
2. Check for conflicts (`-d`, multiple URLs, `--url-file`)
3. Fetch content from source
4. Write to file (overwrite if exists)
5. Check for extension mismatch → emit warning if needed

**Warning Implementation:**

- Extension warnings logged to stderr
- Format: `"Warning: Writing {format} format to file with {ext} extension"`
- Does not prevent operation (user intent honored)
- Exit code remains 0 if operation succeeds
