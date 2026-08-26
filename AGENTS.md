# AGENTS.md

> **A README for AI coding agents** - This document provides context and instructions for AI coding agents working on snag. For human contributors, see [README.md](README.md).

## Project Overview

`snag` is a CLI tool that intelligently fetches web page content using Chrome/Chromium via the Chrome DevTools Protocol (CDP). Built for AI agents to consume web content efficiently.

**Key Features:**

- Auto-detect and connect to existing Chrome instances
- Launch headless or visible browser modes
- Handle authenticated sessions gracefully
- Multiple output formats: Markdown (default), HTML, Text, PDF, PNG
- Multiple URL support: Process multiple URLs in a single command
- Tab management: List, select, and fetch from existing browser tabs
- Pattern matching: Select tabs by index, exact URL, substring, or regex
- Output directory support: Auto-generate filenames with timestamps
- Single binary distribution, no runtime dependencies

**Technology Stack:**

- Language: Go 1.25.3
- CLI Framework: github.com/spf13/cobra v1.10.1
- Browser Control: github.com/go-rod/rod v0.116.2 (Chrome DevTools Protocol)
- HTML to Markdown: github.com/JohannesKaufmann/html-to-markdown/v2 v2.4.0
- HTML to Text: github.com/k3a/html2text v1.2.1

## Setup Commands

```bash
# Clone repository
git clone https://github.com/p3bot/snag.git
cd snag

# Install dependencies
go mod download

# Install
go install github.com/p3bot/snag/cmd/snag@latest

# Build binary
go build -o snag ./cmd/snag

# Run snag
./snag --version
./snag --help
```

## Build and Test Commands

```bash
# Build for current platform
go build -o snag ./cmd/snag

# Build with version info
go build -ldflags "-X github.com/p3bot/snag/internal/cli.Version=1.0.0" -o snag ./cmd/snag

# Run all tests (integration tests with real browser)
go test -v ./...

# Run specific test
go test -v ./... -run TestFetchPage

# Run with coverage
go test -v -cover ./...

# Test basic content fetching
snag https://example.com                # Fetch page as Markdown
snag --format html https://example.com  # Fetch as HTML
snag --format text https://example.com  # Fetch as plain text
snag --format pdf https://example.com   # Save as PDF (auto-generates filename)
snag --format png https://example.com   # Save as PNG screenshot (auto-generates filename)
snag -o output.md https://example.com   # Save to file

# Test multiple URL fetching
snag https://example.com https://google.com               # Fetch multiple URLs
snag -d output/ https://example.com https://google.com    # Save multiple with auto-generated names
snag -o combined.md https://example.com https://google.com # Combine into single file

# Test tab management features (Phase 2)
snag --open-browser                     # Open persistent browser (with DevTools enabled)
snag --list-tabs                        # List all open tabs
snag https://example.com                # Fetch URL (creates new tab)
snag --list-tabs                        # List tabs again (should show example.com)

# Tab selection by index (1-based)
snag --tab 1                            # Fetch from first tab
snag -t 2                               # Fetch from second tab (short form)

# Tab selection by pattern
snag -t "example.com"                   # Exact URL match (case-insensitive)
snag -t "example"                       # Substring/contains match
snag -t "https://.*\.com"               # Regex pattern match

# Tab features with output options
snag -t 1 --format html                 # Fetch tab 1 as HTML
snag -t "github" -o repo.md             # Fetch tab matching "github", save to file
snag -t 1 --wait-for ".content"         # Wait for selector in existing tab

# Cross-platform builds
GOOS=darwin GOARCH=arm64 go build -o snag-darwin-arm64 ./cmd/snag
GOOS=darwin GOARCH=amd64 go build -o snag-darwin-amd64 ./cmd/snag
GOOS=linux GOARCH=amd64 go build -o snag-linux-amd64 ./cmd/snag
GOOS=linux GOARCH=arm64 go build -o snag-linux-arm64 ./cmd/snag

# Code quality checks
go vet ./...
gofmt -l .

# Clean build artifacts
rm -f snag snag-*
```

## Code Style Guidelines

**Go Conventions:**

- Follow standard Go formatting: use `gofmt` or `goimports`
- Use Go 1.25.3+ features and idioms
- Keep functions focused and small
- Use descriptive variable names

**Project-Specific Patterns:**

- Layout: `cmd/snag/main.go` entry point, Cobra in `internal/cli`, domain work in sibling `internal/` packages
- Custom logger for CLI output (`internal/logger`)
- Sentinel errors live in the package that owns the failure
- Exit codes: 0 (success), 1 (any error), 130 (SIGINT), 143 (SIGTERM)
- Output routing (critical for piping):
  - stdout: Content only (HTML/Markdown/Text or binary formats)
  - stderr: All logs, warnings, errors, progress indicators

**Naming Conventions:**

- Exported constants use PascalCase: `format.Markdown`, `format.HTML`
- Sentinel errors: `browser.ErrBrowserNotFound`, `fetch.ErrPageLoadTimeout`, etc.
- Functions: Use descriptive verbs - `validate.URL()`, `fetchPage()`, `convertToMarkdown()`

**Error Handling:**

- Use sentinel errors defined in the owning package
- Wrap errors with context: `fmt.Errorf("failed to navigate to %s: %w", url, err)`
- Clear, actionable error messages via logger
- Never panic for expected errors

**Logging:**

- Use custom Logger with 4 levels (quiet, normal, verbose, debug)
- Log to stderr only (stdout reserved for content)
- Success messages: `logger.Success("Connected to existing Chrome instance")`
- Errors: `logger.Error("Failed to connect")`
- Info messages: `logger.Info("Fetching https://example.com...")`
- Verbose: `logger.Verbose("Navigating to URL...")`
- Debug: `logger.Debug("CDP message: ...")`

**License Headers:**

- Add MPL 2.0 header to all Go source files when creating them
- Apply this header to every new `.go` file created
- See `LICENSE` file Exhibit A (lines 355-367) for reference

```go
// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package <name>
```

## Development Workflow

**Branch Management:**

- Main branch: `main`
- Development branch pattern: `feature/description` or `fix/description`
- Always work on feature branches, PR to main

**Commit Message Format:**

Follow conventional commits style:

```
feat(browser): add support for custom user agents
fix(fetch): resolve WebSocket URL from HTTP endpoint
docs: reorganise project documentation
chore(deps): update rod to v0.116.2
```

Types: `feat`, `fix`, `docs`, `chore`, `test`, `refactor`

**Pull Request Guidelines:**

- Title format: `[component] Brief description`
- Run tests and build before committing
- Ensure code is formatted with `gofmt`
- Update documentation if changing CLI interface
- Keep PRs focused on single feature/fix

## Testing Instructions

**Strategy:**

- **Unit tests**: Pure functions without mocking (validate, format, browser detection)
- **Integration tests**: Real Chrome/Chromium browser (no mocking browser interactions)
- Tab tests may show minor isolation issues (not functional bugs)

**Test Suite:**

- **170 test functions** across 10 test files
- Unit tests live beside each `internal/` package
- Integration tests (browser-dependent) live in `internal/cli/cli_test.go`

**Running Tests:**

```bash
go test -v ./...                     # All tests
go test -v ./internal/validate       # Unit tests only (validation)
go test -v ./internal/browser -run TestDetectBrowser    # Unit tests only (browser detection)
go test -v ./internal/cli -run TestBrowser              # Integration tests (requires browser)
go test -v -cover ./...              # With coverage
```

**Requirements:**

- Chrome, Chromium, Edge, or Brave installed (for integration tests)
- Unit tests run without browser dependency

## Security Considerations

**Browser Security:**

- `internal/browser` contains security comments about browser handling
- Never bypass certificate validation in production
- User agent customization available via `--user-agent` flag to bypass headless detection
- Close tabs in headless mode by default (configurable with `--close-tab`)

**Input Validation:**

- URL validation via `internal/validate`
- Format validation (md, html, text, pdf, png)
- Timeout bounds checking (positive integers)
- Port validation (1-65535 range)

**Authentication:**

- Visible browser mode persists user sessions for authenticated sites
- No credential storage - relies on browser's session management
- Users authenticate manually in visible browser mode

**Output Handling:**

- Content written to stdout (can be piped)
- Logs written to stderr (prevents content leakage)
- File output uses standard file permissions

## Architecture

**File Structure:**

- `cmd/snag/main.go` - Thin entry point (`cli.Execute`, exit codes)
- `internal/cli` - Cobra command, flags, handlers, signals
- `internal/browser` - Rod launch, connect, tab list/select
- `internal/fetch` - Page navigation and content extraction
- `internal/format` - HTML to md/html/text/pdf/png
- `internal/output` - Filename generation and conflict resolution
- `internal/validate` - URL and input validation
- `internal/logger` - Four-level stderr logger
- `internal/doctor` - `--doctor` diagnostics
- `testdata/` - Shared HTML and URL fixtures at the module root

**Key Tab Code Locations:**

```
internal/browser/browser.go    # ListTabs, GetTabByIndex, GetTabsByPattern
internal/cli/handlers.go       # handleListTabs, handleTabFetch
```

**Browser Modes:**

1. Connect to existing Chrome (auto-detect)
2. Launch headless (if none found)
3. Open only (`--open-browser`)

**Tab Management (Phase 2):**

Phase 2 adds efficient tab management capabilities to work with existing browser tabs:

- **List Tabs** (`--list-tabs`, `-l`): Display all open tabs with index, URL, and title
- **Select by Index** (`--tab <index>`, `-t <index>`): Fetch content from specific tab (1-based indexing)
- **Select by Pattern** (`--tab <pattern>`, `-t <pattern>`): Match tabs by URL (exact/substring/regex)

**Tab Selection Examples:**

```bash
# List all open tabs
$ snag --list-tabs
Available tabs in browser (7 tabs, sorted by URL):
  [1] chrome://newtab (New Tab)
  [2] https://developer.mozilla.org/ (MDN Web Docs)
  [3] https://docs.python.org/3/ (Python 3 Documentation)
  [4] https://example.com (Example Domain)
  [5] https://github.com/puppeteer/puppeteer/issues/7452 (Order in browser.pages() · Issue #7452)
  [6] https://google.com/search (youtube - Google Search)
  [7] https://x.com (X. It's what's happening / X)

# Fetch from tab by index (1-based)
snag --tab 1        # First tab
snag -t 3           # Third tab

# Fetch by exact URL (case-insensitive)
snag -t "https://github.com/p3bot/snag"
snag -t "EXAMPLE.COM"  # Case-insensitive match

# Fetch by substring/contains (single match - stdout, multiple matches - auto-save all)
snag -t "github"       # All tabs containing "github" (auto-saves if multiple matches)
snag -t "dashboard"    # All tabs containing "dashboard" (auto-saves if multiple matches)

# Fetch by regex pattern
snag -t "https://.*\.com"              # Regex: https:// + anything + .com
snag -t ".*/dashboard"                 # Regex: any URL ending with /dashboard
snag -t "(github|gitlab)\.com"         # Regex: github.com or gitlab.com

# With output options
snag -t 1 --format html                # HTML format
snag -t "docs" -o reference.md         # Save to file
snag -t 2 --wait-for ".loaded"         # Wait for selector
```

**Pattern Matching Rules (Progressive Fallthrough):**

1. **Integer** → Tab index (1-based, e.g., "1" = first tab)
2. **Exact match** → Case-insensitive URL match (`strings.EqualFold`) - returns ALL exact matches
3. **Contains match** → Case-insensitive substring (`strings.Contains`) - returns ALL containing matches
4. **Regex match** → Case-insensitive regex (`(?i)` flag) - returns ALL regex matches
5. **Error** → No match found (`ErrNoTabMatch`)

**Multiple Matches Behavior:**

- Single match: Outputs to stdout (or to file with `-o`)
- Multiple matches: Auto-saves all with generated filenames (like `--all-tabs`), no confirmation prompt
- Processes in same sort order as `--list-tabs` (alphabetically by URL)
- Cannot use `--output` with multiple matches (error: use `--output-dir`)

**Key Implementation Details:**

- Tab features require existing browser connection (won't auto-launch)
- `--tab` and `<url>` argument are mutually exclusive
- Tab indexes are 1-based for user display (converted internally to 0-based)
- All pattern matching is case-insensitive
- Performance optimized: Single-pass page.Info() caching in GetTabsByPattern
- Tab listing output goes to stdout (enables piping)

**Tab Ordering:**

- **Tabs are sorted by URL (primary), Title (secondary), ID (tertiary)** for predictable, stable ordering
- Chrome DevTools Protocol returns tabs in unpredictable internal order (not visual left-to-right)
- Sorting ensures consistent tab indices across all operations (`--list-tabs`, `--tab`, `--tab <range>`, `--all-tabs`)
- Tab [1] = first tab alphabetically by URL, not first visual tab in browser
- This design choice provides reproducible automation workflows

**Use Cases:**

1. **Authenticated Sessions**: Fetch from authenticated tabs without re-authentication
2. **Reduce Tab Clutter**: Reuse existing tabs instead of creating new ones
3. **Quick Access**: List tabs to find content quickly
4. **Pattern Workflows**: Match tabs by URL patterns for automation
5. **Batch Processing**: Fetch all tabs matching a pattern (e.g., all GitHub repos, all docs pages)

## Dependencies

**Direct Dependencies:**

- `github.com/spf13/cobra` v1.10.1 - CLI framework
- `github.com/go-rod/rod` v0.116.2 - Chrome DevTools Protocol
- `github.com/JohannesKaufmann/html-to-markdown/v2` v2.4.0 - HTML to Markdown conversion
- `github.com/k3a/html2text` v1.2.1 - HTML to plain text conversion

**Runtime Requirements:**

- Chromium-based browser (Chrome, Chromium, Edge, Brave)
- macOS (arm64/amd64) or Linux (amd64/arm64)

## Troubleshooting

**Quick fixes:**

- Browser not found: Install Chromium-based browser (Firefox NOT supported)
- Connection refused: Try different port `--port 9223`
- Timeout: Increase with `--timeout 60`
- Auth required: Use `--open-browser`, authenticate manually in visible browser
- Empty output: Try `--format html` or `--wait-for <selector>`
- Tab errors: Run `snag --list-tabs` first to see available tabs
- Stuck browser processes: Use `--kill-browser` to force-kill debugging browsers

**Debug logging:**

```bash
snag --verbose <url>   # Verbose logs to stderr
snag --debug <url>     # CDP message logs
```

**Get diagnostic information:**

```bash
snag --doctor          # Comprehensive environment diagnostics
```

Include this output when reporting issues.

## Design Philosophy

**Core Principles:**

- Single binary, no config files
- Passive observer (fetch content, don't automate)
- Unix philosophy: stdout = content, stderr = logs
- Clear, actionable error messages

**Non-Goals:**

- Browser automation (use Puppeteer/Playwright instead)
- Web scraping framework
- JavaScript execution/testing

## Important Implementation Notes

**Critical Bug Fix - Remote Debugging Port:**

- ALWAYS explicitly set `--remote-debugging-port` flag (`internal/browser`)
- Rod's launcher won't set it for default port 9222, causing random port selection
- Test with both default and custom ports

**Tab Indexing:**

- User-facing: 1-based (tabs [1], [2], [3]...)
- Internal: 0-based (converted in TabInfo struct and GetTabByIndex)

**Performance - GetTabsByPattern():**

- Caches `page.Info()` results in a single pass (`internal/browser/browser.go`)
- Reduces network calls from 3N to N (3x improvement for 10 tabs)
- Do not modify pattern matching without preserving this optimization

## Release Process

Follow the comprehensive guide in `.ai/tasks/release-process.md`.

**Quick reference**:

```bash
export VERSION="0.0.4"
go test -v ./...                    # Run tests first
# Then follow .ai/tasks/release-process.md for full steps
```

## License

Mozilla Public License 2.0

Third-party licenses in `LICENSES/` directory.

## Current Development Status

**Recently Completed:**

- ✅ Multiple URL support (commit: 7dddb31)
- ✅ Text format support (plain text extraction)
- ✅ PDF format support (Chrome PDF rendering)
- ✅ PNG format support (full-page screenshots)
- ✅ Output directory support with auto-generated filenames
- ✅ Format refactoring (md/html/text/pdf/png)
- ✅ Argument documentation cross-review - All 3 phases complete (2025-10-24)
  - Resolved 7 critical contradictions across 21 argument files
  - Standardized 7 warning message inconsistencies
  - Updated validation.md and README.md compatibility matrices
- ✅ Cobra help template improvements (commit: 50ba8b5)
- ✅ Validation refactoring with flagChanged parameter (commit: 52a4d24)
- ✅ Markdown converter instance reuse optimization (commit: edc37f9)

**Known Limitations:**

- Parallel processing strategy undefined for multiple URL/tab operations

## Additional Resources

- **README.md**: User-facing documentation and usage examples
- **PROJECT.md**: Current project status and work tracking
- **.ai/design/design-records/dr-001-snag-design.md**: Comprehensive design decisions and rationale (33 design decisions documented)
- **.ai/tasks/**: Task documents (code-review, release-process, etc.)
- **docs/arguments/**: Complete argument compatibility matrix
- **docs/testing.md**: Testing documentation and strategies
- **Repository**: https://github.com/p3bot/snag
- **Issues**: https://github.com/p3bot/snag/issues
