# `--format FORMAT` / `-f`

**Status:** Complete (2025-10-22)

#### Validation Rules

**Format Names:**

- Valid formats: `md`, `html`, `text`, `pdf`, `png`
- Format aliases: `markdown` accepted as alias for `md`
- Case-insensitive matching: `HTML`, `Html`, `html` all valid
- Invalid format name → Error: `"Invalid format '{format}'. Supported: md, html, text, pdf, png"`
- Empty string → Error: `"Format cannot be empty"`
- Typos/close matches → Error (no fuzzy matching)

**Multiple Format Flags:**

- Multiple `--format` flags → **Last wins** (standard CLI behavior, no error, no warning)

**Format Without Value:**

- `--format` with no value → Parse error from CLI framework

**Error Messages:**

- Invalid format: `"Invalid format '{format}'. Supported: md, html, text, pdf, png"`
- Empty format: `"Format cannot be empty"`

#### Behavior

**Basic Usage:**

```bash
snag https://example.com --format html
```

- Fetches URL content
- Converts to specified format
- Outputs according to output rules (stdout or file)

**Text Formats (md, html, text):**

```bash
snag https://example.com --format markdown       # Markdown (alias)
snag https://example.com --format html           # HTML
snag https://example.com --format text           # Plain text
```

- Output to stdout by default
- Can be redirected with `-o` or `-d`
- Suitable for piping

**Binary Formats (pdf, png):**

```bash
snag https://example.com --format pdf            # PDF
snag https://example.com --format png            # Screenshot PNG
```

- **Never output to stdout** (would corrupt terminal)
- Auto-generate filename in current directory if no `-o` or `-d`
- Always saved to file

**Extension Mismatch Warning:**

```bash
snag https://example.com --format html -o page.md
```

- **Warning message:** `"Warning: Writing HTML format to file with .md extension"`
- User intent honored (file written as requested)
- Exit code 0 (warning, not error)

**Default Format:**

- No `--format` flag → Defaults to `md` (Markdown)

#### Interaction Matrix

**Content Source Interactions:**

| Combination                   | Behavior       | Notes                                 |
| ----------------------------- | -------------- | ------------------------------------- |
| `--format` + single `<url>`   | Works normally | Apply format to fetched content       |
| `--format` + multiple `<url>` | Works normally | Apply same format to all URLs         |
| `--format` + `--url-file`     | Works normally | Apply same format to all URLs in file |
| `--format` + `--tab`          | Works normally | Apply format to tab content           |
| `--format` + `--all-tabs`     | Works normally | Apply same format to all tabs         |

**Output Destination Interactions:**

| Combination                               | Behavior                             | Notes                              |
| ----------------------------------------- | ------------------------------------ | ---------------------------------- |
| `--format text` (no output flag)          | Output to stdout                     | Standard behavior for text formats |
| `--format html` (no output flag)          | Output to stdout                     | Standard behavior for text formats |
| `--format md` (no output flag)            | Output to stdout                     | Standard behavior for text formats |
| `--format pdf` (no output flag)           | Auto-save to current dir             | Binary formats never go to stdout  |
| `--format png` (no output flag)           | Auto-save to current dir             | Binary formats never go to stdout  |
| `--format` + `-o file.md` (matching ext)  | Write to file                        | Normal operation                   |
| `--format html` + `-o file.md` (mismatch) | Write to file with **warning**       | Extension mismatch warning shown   |
| `--format pdf` + `-o file.pdf`            | Write to file                        | Normal operation                   |
| `--format` + `-d directory/`              | Auto-generate with correct extension | Extension matches format           |

**Extension Mapping with `-d`:**

| Format            | Auto-generated Extension |
| ----------------- | ------------------------ |
| `md` / `markdown` | `.md`                    |
| `html`            | `.html`                  |
| `text`            | `.txt`                   |
| `pdf`             | `.pdf`                   |
| `png`             | `.png`                   |

**Special Operation Modes:**

| Combination                            | Behavior                  | Notes                                                                   |
| -------------------------------------- | ------------------------- | ----------------------------------------------------------------------- |
| `--format` + `--doctor`                | **Flag ignored**          | Doctor overrides, diagnostics in fixed format                           |
| `--format` + `--list-tabs`             | `--list-tabs` overrides   | `--list-tabs` overrides all other options                               |
| `--format` + `--open-browser` (no URL) | **Warning**, flag ignored | `"Warning: --format ignored with --open-browser (no content fetching)"` |
| `--format` + `--kill-browser`          | **Flag ignored**          | No content to save                                                      |

**Browser Mode Interactions:**

All work normally:

| Combination                         | Behavior                          |
| ----------------------------------- | --------------------------------- |
| `--format` + `--force-headless`     | Works normally                    |
| `--format` + `--open-browser` + URL | Works normally (current behavior) |

**Page Loading Interactions:**

All work normally:

| Combination                    | Behavior                                       |
| ------------------------------ | ---------------------------------------------- |
| `--format` + `--timeout`       | Works normally - timeout applies to page load  |
| `--format` + `--wait-for`      | Works normally - wait before format conversion |
| `--format` + `--user-agent`    | Works normally - UA set for new pages          |
| `--format` + `--user-data-dir` | Works normally - use custom browser profile    |
| `--format` + `--port`          | Works normally - use specified port            |
| `--format` + `--close-tab`     | Works normally - close after fetching          |

**Logging Flags:**

All work normally:

- `--verbose`: Format output, verbose logs to stderr
- `--quiet`: Format output, suppress logs (errors only)
- `--debug`: Format output, debug logs to stderr

#### Examples

**Valid:**

```bash
snag https://example.com --format html              # HTML to stdout
snag https://example.com --format markdown          # Markdown alias
snag https://example.com --format text              # Plain text
snag https://example.com --format pdf               # PDF auto-saved
snag https://example.com --format png               # PNG screenshot auto-saved
snag https://example.com -f HTML                    # Case-insensitive
snag https://example.com --format html -o page.html # HTML to file
snag https://example.com --format pdf -o doc.pdf    # PDF to file
snag https://example.com --format html -d ./output  # Auto-generated .html
snag url1 url2 --format pdf -d ./docs               # Batch PDF generation
snag --url-file urls.txt --format text              # Text format for all
snag --tab 1 --format html                          # HTML from tab
snag --all-tabs --format pdf -d ./tabs              # All tabs as PDFs
snag https://example.com --format html --format pdf # Uses pdf (last wins)
```

**Invalid:**

```bash
snag https://example.com --format invalid           # ERROR: Invalid format
snag https://example.com --format ""                # ERROR: Empty format
snag https://example.com --format                   # ERROR: Missing value
snag https://example.com --format markdwon          # ERROR: Typo (no fuzzy match)
```

**With Warnings:**

```bash
snag https://example.com --format html -o page.md   # ⚠️ Extension mismatch
snag https://example.com --format pdf -o doc.txt    # ⚠️ Binary to text extension
snag https://example.com --format text -o file      # ⚠️ No extension
```

#### Implementation Details

**Location:**

- Flag definition: `internal/cli/root.go` (`init`)
- Format validation: `internal/validate` (`Format`, `NormalizeFormat`)
- Format conversion: `internal/format` (`ProcessContent`)
- Extension mapping: Output generation functions

**Format Conversion Flow:**

1. Validate format name (case-insensitive, check against valid list)
2. Fetch page content (HTML)
3. Convert based on format:
   - `md` / `markdown`: HTML → Markdown (html-to-markdown library)
   - `html`: Raw HTML (pass-through)
   - `text`: HTML → Plain text (html2text library)
   - `pdf`: Chrome PDF rendering via CDP
   - `png`: Full-page screenshot via CDP
4. Route output based on format type and output flags

**Binary Format Handling:**

- PDF/PNG use Chrome's native rendering capabilities
- Never sent to stdout (would corrupt terminal)
- Auto-generate filename if no output destination specified
- Filename pattern: `yyyy-mm-dd-hhmmss-{page-title}-{slug}.{ext}`

**Extension Mismatch Detection:**

- Check output file extension against format
- Emit warning to stderr if mismatch detected
- Does not prevent operation (user intent honored)
- Warning format: `"Warning: Writing {format} format to file with {ext} extension"`

**Format Aliases:**

- `markdown` → `md` (implemented in validation)
- Case normalization happens during validation
