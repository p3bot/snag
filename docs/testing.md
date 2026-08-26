# Testing Guide for Snag

This document describes the testing infrastructure and practices for the Snag project.

---

## Table of Contents

1. [Test Overview](#test-overview)
2. [Test Files](#test-files)
3. [Running Tests](#running-tests)
4. [Coverage Reports](#coverage-reports)
5. [Interactive Testing](#interactive-testing)
6. [Writing Tests](#writing-tests)
7. [CI/CD Integration](#cicd-integration)

---

## Test Overview

Snag has a comprehensive test suite covering unit tests, integration tests, and interactive manual tests.

### Test Statistics

- **Total Test Functions**: 170
- **Test Files**: 10 (`*_test.go`)
- **Interactive Test Cases**: 73 (CSV-driven)
- **Test Code Lines**: ~3,100 lines
- **Production Code Lines**: ~2,500 lines
- **Test Coverage Ratio**: 1.24:1

### Test Categories

| Category              | Description                                  | Test Files                                                                                                       |
| --------------------- | -------------------------------------------- | ---------------------------------------------------------------------------------------------------------------- |
| **Unit Tests**        | Fast, isolated tests of individual functions | `internal/{output,validate,format,logger,browser,cli,fetch,doctor}/*_test.go` |
| **Integration Tests** | End-to-end tests with browser                | `internal/cli/cli_test.go`                                       |
| **Interactive Tests** | Manual verification via script               | `test-interactive` + `test-interactive.csv`                                                                      |

---

## Test Files

### 1. `internal/cli/cli_test.go` (62 tests, 1,600 lines)

**Purpose**: Integration tests for CLI functionality and browser interactions

**Key Test Functions**:

- `TestCLI_*` - Command-line flag validation
- `TestBrowser_*` - Browser integration (URL fetching, formats)
- `TestTab_*` - Tab operations (list, fetch by index/pattern)
- `TestBatch_*` - Batch operations (--all-tabs)

**Test Infrastructure**:

- `TestMain()` - Test setup/teardown with signal handling
- `runSnag()` - Helper to run snag binary and capture output
- `startTestServer()` - Local HTTP server for test pages
- `isBrowserAvailable()` - Browser detection for skipping tests

**Example Tests**:

```go
TestBrowser_PDFFormat()      // Tests PDF generation
TestBrowser_TextFormat()     // Tests plain text extraction
TestBrowser_OutputDir()      // Tests --output-dir flag
TestTab_FetchByIndex()       // Tests --tab <n> flag
TestCLI_InvalidFormat()      // Tests format validation
```

**Test Patterns**:

- Browser-required tests: Skipped if no browser available
- Output verification: Checks stdout, stderr, and files
- Assertion helpers: `assertContains()`, `assertNoError()`, `assertExitCode()`

### 2. `internal/output/output_test.go` (10 tests)

**Purpose**: Tests filename generation, slugification, and conflict resolution

**Coverage**:

- `TestSlugifyTitle()` - 12 test cases for slug generation
- `TestGenerateURLSlug()` - URL fallback slugs, including IPv6 hosts
- `TestGetFileExtension()` - 7 test cases for format mapping
- `TestGenerateFilename()` - 8 test cases for complete filename generation
- `TestResolveConflict()` - File conflict resolution
- `TestSlugifyTitle_Truncation()` - Edge cases for truncation

**Example Slugification**:

```
"Example Domain"              → "example-domain"
"GitHub - Project Page"       → "github-project-page"
"!!!Test???"                  → "test"
"Very Long Title..."          → "very-long-title" (80 char max)
```

### 3. `internal/validate/validate_test.go` (26 tests)

**Purpose**: Tests input validation and security checks

**Coverage**:

- `TestValidateURL_*` - URL validation (3 tests)
- `TestValidateFormat_*` - Format validation (2 tests)
- `TestNormalizeFormat()` - Format normalization (18 cases)
- `TestValidateTimeout_*` - Timeout validation (2 tests)
- `TestValidatePort_*` - Port validation (2 tests)
- `TestValidateOutputPath_*` - Output path validation (3 tests)
- `TestValidateDirectory_*` - Directory validation (4 tests)
- `TestIsNonFetchableURL()` - Browser-internal URL detection (18 cases)
- `TestCheckExtensionMismatch()` - File extension validation (16 cases)
- `TestValidateWaitFor()` - CSS selector validation (10 cases)
- `TestValidateWaitFor_Injection()` - Documents that selectors are not sanitised
- `TestValidateUserAgent()` - User agent sanitisation (11 cases)
- `TestValidateUserAgent_SecuritySanitization()` - HTTP header injection prevention (4 cases)
- `TestValidateUserAgent_ExtremeLength()` - Long user-agent strings
- `TestValidateURL_IDNHomograph()` - IDN / punycode URLs allowed
- `TestValidateURL_ExtremelyLong()` - Long URL paths

**Security Tests**:

```go
// Newlines in --user-agent are replaced with spaces
TestValidateUserAgent_SecuritySanitization()
  - "MyBot/1.0\r\nX-Injected: malicious" → no CR/LF in result
  - "MyBot/1.0\nX-Injected: malicious"   → no CR/LF in result
```

### 4. `internal/format/format_test.go` (24 tests)

**Purpose**: Tests format conversion (Markdown, HTML, Text, PDF, PNG)

**Coverage**:

- `TestConvertToMarkdown_*` - Markdown conversion (7 tests)
  - Headings, links, tables, strikethrough, lists, code blocks
- `TestExtractPlainText_*` - Plain text extraction (8 tests)
  - Headings, links, formatting, scripts, lists
- `TestProcessContent_*` - Page interface (markdown file, PDF bytes)

**Format Verification**:

```go
// HTML → Markdown
"<h1>Title</h1>" → "# Title"

// HTML → Plain Text (no markup)
"<strong>bold</strong>" → "bold"
"<a href='...'>link</a>" → "link text + URL"
```

### 5. `internal/logger/logger_test.go` (9 tests)

**Purpose**: Tests logging functionality and utility functions

**Coverage**:

- `TestLogger_Success()` - Success messages
- `TestLogger_Error()` - Error messages
- `TestLogger_Info()` - Info messages
- `TestLogger_Debug()` - Debug messages
- `TestLogger_QuietMode()` - Quiet mode (no output)
- `TestLogger_VerboseMode()` - Verbose mode
- `TestLogger_StderrOnly()` - Stderr-only output
- `TestShouldUseColor()` - NO_COLOR environment variable handling (3 cases)
- `TestNewLogger()` - Logger constructor for all log levels (4 cases)

### 6. `internal/browser/browser_test.go` (6 tests)

**Purpose**: Tests browser detection and utility functions

**Coverage**:

- `TestDetectBrowserName()` - Browser name detection (46 cases)
  - Chrome, Chromium, Ungoogled-Chromium, Edge, Brave
  - Opera, Vivaldi, Arc, Yandex, Thorium, Slimjet, Cent
  - Extension stripping (.exe, .app)
  - Case-insensitive matching
  - Fallback behavior for unknown browsers
- `TestDetectBrowserName_OrderOfPrecedence()` - Detection priority (2 cases)
- `TestDetectBrowserName_ExtensionStripping()` - Cross-platform path handling (2 cases)
- `TestDetectBrowserName_FallbackBehavior()` - Unknown browser handling (3 cases)
- `internal/browser/page_test.go` (2 tests) - nil page and element guards

### 7. `internal/cli/handlers_test.go` (9 tests)

**Purpose**: Tests handler utility functions

**Coverage**:

- `TestStripURLParams()` - URL parameter stripping (8 cases)
- `TestFormatTabLine()` - Tab list formatting (8 cases)
- `TestFormatTabLine_Length()` - Line length verification
- `TestDisplayTabList()` - Tab list display with sorting
- `TestOutputPageInfo_SetsSlug()` - `--info` JSON slug filled at output time

### 8. `internal/fetch/info_test.go` (5 tests)

**Purpose**: Tests page metadata extraction (domain, JSON shape, IPv6 hosts)

### 9. `internal/doctor/doctor_test.go` (17 tests)

**Purpose**: Tests `--doctor` report formatting and collection

---

## Running Tests

### Quick Test Commands

```bash
# Run all tests (unit + integration)
go test ./...

# Run all tests with verbose output
go test -v ./...

# Run only unit tests (fast, no browser)
go test -short ./...

# Run specific test function
go test ./... -run TestSlugifyTitle

# Run tests matching pattern
go test ./... -run "Test(Validate|Output)"

# Run with race detection
go test -race ./...

# Run with parallel execution
go test ./... -parallel 4
```

### Unit Tests Only (Fast)

```bash
# Run only unit tests (exclude browser integration)
go test ./... -run "^Test(Validate|Normalize|Slugify|Generate|Resolve|Extract|Output|Format|Logger)"

# Approximate runtime: ~1 second
```

### Integration Tests Only

```bash
# Run only browser integration tests
go test ./... -run "^Test(Browser|Tab|Batch)"

# Approximate runtime: ~3-5 minutes (browser startup overhead)
```

### Test Output

```
$ go test -v ./...
=== RUN   TestSlugifyTitle
--- PASS: TestSlugifyTitle (0.00s)
=== RUN   TestValidateFormat_Valid
--- PASS: TestValidateFormat_Valid (0.00s)
=== RUN   TestBrowser_PDFFormat
--- PASS: TestBrowser_PDFFormat (3.97s)
...
PASS
ok  	github.com/p3bot/snag/internal/cli	12.345s
```

---

## Coverage Reports

### Generate Coverage Report

```bash
# Generate coverage profile
go test -coverprofile=coverage.out ./...

# View coverage summary
go test -cover ./...

# Output:
# PASS
# coverage: 78.5% of statements
# ok      github.com/p3bot/snag/internal/cli    1.234s
```

### Detailed Coverage by Function

```bash
# Generate coverage profile
go test -coverprofile=coverage.out ./...

# View function-level coverage
go tool cover -func=coverage.out

# Output:
# output.go:32:    SlugifyTitle         100.0%
# output.go:58:    GenerateURLSlug       85.7%
# validate.go:143: validateFormat       100.0%
# ...
# total:           (statements)          78.5%
```

### HTML Coverage Report (Recommended)

```bash
# Generate and open HTML coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Opens browser with interactive visualization:
# - Green highlighting = code covered by tests
# - Red highlighting = code NOT covered by tests
# - Gray = not trackable (comments, declarations)
```

### Coverage Statistics

**Overall Coverage**: 19.7% of statements

This includes integration code (handlers, main) which requires browser. Core business logic has much higher coverage.

```bash
# Generate current coverage
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | grep total
# total: (statements) 19.7%
```

### Coverage by Module

| Module              | Coverage | Functions   | Priority  | Notes                               |
| ------------------- | -------- | ----------- | --------- | ----------------------------------- |
| `internal/output`   | **~94%** | 5/5 tested  | ✅ High   | Core filename logic                 |
| `internal/validate` | **~96%** | 5/5 tested  | ✅ High   | Security & input validation         |
| `internal/logger`   | **~53%** | 5/7 tested  | ⚠️ Medium | Logging utilities                   |
| `internal/format`   | **~60%** | 2/11 tested | ⚠️ Medium | Format conversion (unit tests only) |
| `internal/cli`      | mixed    | helpers unit-tested; handlers via integration | ⏸️ Low    | Browser integration in `cli_test.go` |
| `cmd/snag`          | **0%**   | 0/1 tested  | ⏸️ Low    | Thin entry (not critical)           |

**Key Coverage Details**:

- `SlugifyTitle()`: 100%
- `GenerateURLSlug()`: 85.7%
- `GenerateFilename()`: 100%
- `ResolveConflict()`: 84.2%
- `validateFormat()`: 100%
- `validateDirectory()`: 94.7%
- `validate.OutputPath()`: existing file vs missing/read-only directory
- `validate.UserAgent()`: trim and newline sanitisation
- `convertToMarkdown()`: 80.0%
- `extractPlainText()`: 100%

**Note**: Integration code (handlers, I/O functions) requires browser and is tested via `cli_test.go` integration tests, not unit test coverage.

---

## Interactive Testing

Snag includes an interactive testing script for manual verification of features.

### Overview

**Script**: `test-interactive`
**Test Data**: `test-interactive.csv`
**Test Cases**: 73 interactive scenarios
**Dependencies**: `go`, `fzf` (fuzzy finder)

### Running Interactive Tests

```bash
# Make script executable (first time only)
chmod +x test-interactive

# Run interactive test suite
./test-interactive
```

### Interactive Test Workflow

1. **Build**: Script builds `snag` binary
2. **Setup**: Creates temporary test directory
3. **Selection**: Use `fzf` to select test sections
   - Press `Tab` to multi-select
   - Press `Enter` to confirm
4. **Execution**: Tests run one-by-one
   - Command shown before execution
   - Output displayed
   - Press any key to continue
5. **Verification**: Automatic or manual verification
6. **Summary**: Pass/fail statistics
7. **Cleanup**: Option to preserve or delete test directory

### Test Sections (CSV)

The `test-interactive.csv` file organizes tests into sections:

```csv
section,description,command,verify
Basic Text Fetching,Markdown to stdout,./snag https://example.com,stdout
Basic Text Fetching,HTML to stdout,./snag -f html https://example.com,stdout
Binary Output,PDF generation,./snag -f pdf https://example.com,ls
Output Directory,Auto-generated filename,./snag -d ./output https://example.com,ls
Tab Operations,List all tabs,./snag --list-tabs,stdout
Batch Operations,All tabs as PDFs,./snag -a -f pdf -d ./pdfs,ls
Error Conditions,Invalid format,./snag -f json https://example.com,error
```

### Verification Types

| Verify Type | Description                          | Example                       |
| ----------- | ------------------------------------ | ----------------------------- |
| `stdout`    | Check command succeeded, show output | Text format tests             |
| `error`     | Expect error exit code               | Invalid input tests           |
| `file:path` | Check file created, display content  | Specific filename tests       |
| `open:path` | Check file created, open in viewer   | PDF/PNG viewer tests          |
| `ls`        | List directory, auto-open new files  | Auto-generated filename tests |

### Example Session

```bash
$ ./test-interactive

═══════════════════════════════════════════════════════════════════
Snag Interactive Testing Suite
═══════════════════════════════════════════════════════════════════

Building Snag
─────────────
✔ Built snag binary successfully

Test Environment Setup
──────────────────────
Created temporary test directory: /tmp/snag-test-ABC123
✔ Working directory: /tmp/snag-test-ABC123

Loading Test Definitions
─────────────────────────
✔ Loaded 73 tests from CSV

Section Selection
─────────────────
Select test section(s) to run:
> All
  Basic Text Fetching
  Binary Output
  Output Directory
  Tab Operations
  ...

✔ Will run 73 tests

Starting Test Execution
───────────────────────
Total tests to run: 73

Test 1/73: Markdown to stdout
─────────────────────────────
Section: Basic Text Fetching
Command: ./snag https://example.com

[output displayed]

✔ Command succeeded (output shown above)

Press any key to continue...

[... continues through all tests ...]

Test Summary
────────────
Total tests: 73
✔ Passed: 71
✖ Failed: 2

Cleanup
───────
Test directory: /tmp/snag-test-ABC123
Keep test directory? [Y/n]:
```

### Adding New Tests

Edit `test-interactive.csv` to add new test cases:

```csv
section,description,command,verify
My Section,Test description,./snag [options] <url>,stdout
```

**Tips**:

- Keep descriptions concise but descriptive
- Use `./snag` (relative path) for commands
- Choose appropriate verify type
- Group related tests in same section

---

## Writing Tests

### Test File Structure

```go
package output

import (
    "testing"

    "github.com/p3bot/snag/internal/logger"
)

func init() {
    logger.SetDefault(logger.Discard())
}

// Test function naming: TestFunctionName or TestFunctionName_Scenario
func TestSlugifyTitle(t *testing.T) {
    // Table-driven tests
    tests := []struct {
        input    string
        maxLen   int
        expected string
        desc     string
    }{
        {"Example Title", 80, "example-title", "basic slugification"},
        {"Special !@# Chars", 80, "special-chars", "special characters"},
    }

    for _, tt := range tests {
        result := SlugifyTitle(tt.input, tt.maxLen)
        if result != tt.expected {
            t.Errorf("SlugifyTitle(%q, %d) [%s] = %q, expected %q",
                tt.input, tt.maxLen, tt.desc, result, tt.expected)
        }
    }
}
```

### Best Practices

#### 1. Use Table-Driven Tests

```go
// Good: Table-driven test
func TestValidateFormat(t *testing.T) {
    tests := []struct {
        format   string
        expected bool
    }{
        {"md", true},
        {"html", true},
        {"json", false},
    }

    for _, tt := range tests {
        err := validateFormat(tt.format)
        if (err == nil) != tt.expected {
            t.Errorf("validateFormat(%q) unexpected result", tt.format)
        }
    }
}
```

#### 2. Use Descriptive Test Names

```go
// Good: Descriptive test names
TestSlugifyTitle_WithSpecialCharacters
TestValidateDirectory_NonExistent
TestResolveConflict_MultipleConflicts

// Avoid: Vague names
TestSlugify
TestValidate
TestConflict
```

#### 3. Test Error Cases

```go
func TestResolveConflict_NonexistentDirectory(t *testing.T) {
    _, err := ResolveConflict("/nonexistent/dir", "test.md")
    if err != nil {
        // This should succeed (no error for nonexistent dir check)
        t.Fatalf("unexpected error: %v", err)
    }
}
```

#### 4. Use t.Fatalf for Critical Failures

```go
// Use t.Fatalf when test cannot continue
content, err := os.ReadFile(filepath)
if err != nil {
    t.Fatalf("failed to read file: %v", err)
}

// Use t.Errorf for non-critical failures
if len(files) != 1 {
    t.Errorf("expected 1 file, got %d", len(files))
}
```

#### 5. Clean Up Test Resources

```go
func TestOutputDir(t *testing.T) {
    // Use t.TempDir() - auto-cleanup
    tmpDir := t.TempDir()

    // Or manual cleanup with t.Cleanup()
    file, _ := os.CreateTemp("", "test-*")
    t.Cleanup(func() {
        os.Remove(file.Name())
    })
}
```

#### 6. Skip Tests When Appropriate

```go
func TestBrowser_Integration(t *testing.T) {
    if !isBrowserAvailable() {
        t.Skip("Browser not available, skipping test")
    }
    // ... test code ...
}
```

### Test Helpers

Common assertion helpers in `cli_test.go`:

```go
// Assert helpers
assertNoError(t, err)                       // err must be nil
assertError(t, err)                         // err must be non-nil
assertExitCode(t, err, expectedCode)        // check exit code
assertContains(t, output, substr)           // check substring
assertNotContains(t, output, substr)        // check NOT contains
```

---

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: "1.25"

      - name: Run unit tests
        run: go test -short ./... -v

      - name: Run all tests with coverage
        run: go test -coverprofile=coverage.out ./... -v

      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out
```

### Local Pre-commit Hook

Create `.git/hooks/pre-commit`:

```bash
#!/bin/bash
# Run tests before commit

echo "Running unit tests..."
go test -short ./...

if [ $? -ne 0 ]; then
    echo "Tests failed. Commit aborted."
    exit 1
fi

echo "Tests passed!"
```

```bash
chmod +x .git/hooks/pre-commit
```

---

## Troubleshooting

### Tests Timing Out

```bash
# Increase timeout for slow tests
go test -timeout 30m ./...
```

### Browser Tests Failing

```bash
# Check if browser is available
./snag --list-tabs

# Skip browser tests
go test -short ./...

# Run only unit tests
go test ./... -run "^Test(Validate|Output|Format|Logger)"
```

### Coverage Not Updating

```bash
# Clean test cache
go clean -testcache

# Re-run with coverage
go test -coverprofile=coverage.out ./...
```

### Verbose Test Output

```bash
# Show all test output
go test -v ./...

# Show only failing tests
go test -v ./... 2>&1 | grep -E "(FAIL|RUN.*FAIL)"
```

---

## Test Maintenance

### Adding New Tests Checklist

- [ ] Write table-driven test when possible
- [ ] Test happy path AND error cases
- [ ] Use descriptive test name
- [ ] Add test documentation comment
- [ ] Update this document if needed
- [ ] Run `go test ./...` to verify
- [ ] Run `go test -cover ./...` to check coverage

### Test Review Checklist

- [ ] Tests are deterministic (no random failures)
- [ ] Tests clean up resources (temp files, directories)
- [ ] Browser tests skip if browser unavailable
- [ ] Error messages are clear and actionable
- [ ] Coverage increases or stays same

---

## Summary

Snag has a robust testing infrastructure:

✅ **100 automated test functions**
✅ **73 interactive test scenarios**
✅ **~1.14:1 test-to-code ratio**
✅ **Unit + integration + manual testing**
✅ **Coverage reporting available**
✅ **Interactive test script for manual verification**

**Quick Commands**:

```bash
go test ./...                    # Run all tests
go test -short ./...             # Unit tests only
go test -cover ./...             # With coverage
go tool cover -html=coverage.out # View coverage
./test-interactive               # Manual testing
```

For questions or issues with testing, see the [main README](../README.md) or open an issue.

---

**Last Updated**: 2025-10-22
**Version**: 1.0.0
