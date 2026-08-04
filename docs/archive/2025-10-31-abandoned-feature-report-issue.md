# PROJECT: Report Issue Feature

**STATUS: ⛔ ABANDONED - 2025-10-31**

**Reason:** Overly complex for minimal user benefit. Almost nobody will use a dedicated `--report-issue` flag.

**Pivot:** Instead, enhance `--doctor` output with:
1. Link to repository (already exists)
2. Link to create new issue
3. Create GitHub issue template with instructions for including `--doctor` diagnostics

---

## Original Proposal (Abandoned)

Add a `--report-issue` flag to streamline bug reporting by collecting diagnostics and opening a pre-filled GitHub issue.

### Overview

Combine the `--doctor` diagnostic functionality with seamless GitHub issue creation. Users can report bugs with one command - diagnostics are automatically collected, copied to clipboard, and a GitHub issue template opens in their browser.

## Decisions Made

1. **Flag name**: `--report-issue`
2. **Browser to use**: System default browser (not detected Chrome/Chromium) - user's default is likely authenticated for GitHub
3. **URL length problem**: Doctor output URL-encoded (~2,658 chars) exceeds browser limit (~2,048 chars)
4. **Solution**: Copy full doctor output to clipboard, open browser with simple issue template that instructs user to paste
5. **Privacy**: No warning needed - only contains username in file paths
6. **Issue title**: Pre-fill with generic "[Bug Report] Issue with snag" - user can customize
7. **No dependencies**: Don't require GitHub CLI (`gh`) or any external tools
8. **Override behavior**: Like `--doctor`, overrides all flags except `--help`

## Requirements

### Command-Line Interface

```bash
# Report an issue
snag --report-issue

# Works with custom port (includes port info in diagnostics)
snag --report-issue --port 9223
```

### Behavior Flow

1. **Collect diagnostics** - Run doctor diagnostic collection (same as `--doctor`)
2. **Copy to clipboard** - Copy full doctor output to system clipboard
3. **Build GitHub URL** - Create issue URL with pre-filled template
4. **Open browser** - Launch system default browser to that URL
5. **Confirm to user** - Display success message with instructions

### Output Example

```
Collecting diagnostic information...
✓ Diagnostics collected

Copying to clipboard...
✓ Diagnostics copied to clipboard

Opening GitHub issue in your browser...
✓ Browser opened

Instructions:
1. The issue template has been opened in your browser
2. Fill in the description and steps to reproduce
3. Paste the diagnostics (Ctrl+V / Cmd+V) in the "Diagnostic Information" section
4. Submit the issue

Your diagnostic information is ready to paste!
```

### GitHub Issue Template

**URL format:**
```
https://github.com/p3bot/snag/issues/new?title=[Bug%20Report]%20Issue%20with%20snag&body=<template>
```

**Template content (URL-encoded):**
```markdown
## Description
<!-- Please describe the issue you're experiencing -->



## Steps to Reproduce
1.
2.
3.

## Expected Behavior
<!-- What did you expect to happen? -->



## Actual Behavior
<!-- What actually happened? -->



## Diagnostic Information
<!-- ✓ Diagnostics have been copied to your clipboard - paste here (Ctrl+V / Cmd+V): -->

```



## Additional Context
<!-- Any other context, screenshots, logs, etc. -->


```

**Why this template:**
- Clear sections guide user through providing good bug report
- Empty lines give space for user input
- Clear instruction where to paste diagnostics
- Checkmark (✓) confirms clipboard action worked
- Additional Context for screenshots/extra info

## Implementation Details

### File Changes

**main.go:**
- Add `--report-issue` boolean flag definition
- Add to flag priority logic: `help > version > doctor > report-issue > ...`
- When set, ignore all other flags except `--help` and `--port`
- Route to handler

**handlers.go:**
- Add `handleReportIssue(cmd *cobra.Command) error` function
- Collect diagnostics (reuse doctor logic)
- Copy to clipboard
- Build GitHub URL
- Open browser
- Display instructions

**New file: `clipboard.go`** (or add to existing file):
- `func CopyToClipboard(text string) error` - cross-platform clipboard copy
- Uses platform-specific commands (see below)

**Reuse from doctor:**
- Diagnostic collection logic
- Doctor report formatting
- Same data structures

### Platform-Specific Clipboard

**Linux:**
```go
// Try xclip first, fall back to xsel
cmd := exec.Command("xclip", "-selection", "clipboard")
cmd.Stdin = strings.NewReader(text)
err := cmd.Run()
if err != nil {
    // Fallback to xsel
    cmd = exec.Command("xsel", "--clipboard", "--input")
    cmd.Stdin = strings.NewReader(text)
    err = cmd.Run()
}
```

**macOS:**
```go
cmd := exec.Command("pbcopy")
cmd.Stdin = strings.NewReader(text)
err := cmd.Run()
```

**Clipboard not available:**
- Print diagnostics to stdout
- Tell user to manually copy/paste
- Still open browser with template

### Platform-Specific Browser Opening

**Linux:**
```go
exec.Command("xdg-open", url).Start()
```

**macOS:**
```go
exec.Command("open", url).Start()
```

**Error handling:**
- If browser open fails, print URL for manual opening
- If clipboard fails, print diagnostics for manual copy

### GitHub URL Construction

```go
const (
    githubIssueBase = "https://github.com/p3bot/snag/issues/new"
    issueTitle = "[Bug Report] Issue with snag"
)

func buildIssueURL() string {
    template := `## Description
<!-- Please describe the issue you're experiencing -->



## Steps to Reproduce
1.
2.
3.

## Expected Behavior
<!-- What did you expect to happen? -->



## Actual Behavior
<!-- What actually happened? -->



## Diagnostic Information
<!-- ✓ Diagnostics have been copied to your clipboard - paste here (Ctrl+V / Cmd+V): -->

```



## Additional Context
<!-- Any other context, screenshots, logs, etc. -->


`

    params := url.Values{}
    params.Set("title", issueTitle)
    params.Set("body", template)

    return fmt.Sprintf("%s?%s", githubIssueBase, params.Encode())
}
```

### Implementation Steps

**Phase 1: Core functionality**
```go
func handleReportIssue(cmd *cobra.Command) error {
    port := getPort(cmd) // Check if --port specified

    // Step 1: Collect diagnostics (reuse doctor logic)
    logger.Info("Collecting diagnostic information...")
    report, err := collectDoctorInfo(port)
    if err != nil {
        logger.Warning("Failed to collect some diagnostics: %v", err)
        // Continue anyway - partial info is better than none
    }
    diagnostics := report.Format() // Get formatted string
    logger.Success("Diagnostics collected")

    // Step 2: Copy to clipboard
    logger.Info("Copying to clipboard...")
    if err := copyToClipboard(diagnostics); err != nil {
        logger.Warning("Failed to copy to clipboard: %v", err)
        logger.Info("Diagnostic information:")
        fmt.Fprintln(os.Stdout, diagnostics)
        logger.Info("Please copy the above output manually")
    } else {
        logger.Success("Diagnostics copied to clipboard")
    }

    // Step 3: Build GitHub URL
    issueURL := buildIssueURL()

    // Step 4: Open browser
    logger.Info("Opening GitHub issue in your browser...")
    if err := openBrowser(issueURL); err != nil {
        logger.Warning("Failed to open browser: %v", err)
        logger.Info("Please open this URL manually:")
        fmt.Fprintln(os.Stdout, issueURL)
    } else {
        logger.Success("Browser opened")
    }

    // Step 5: Instructions
    logger.Info("")
    logger.Info("Instructions:")
    logger.Info("1. The issue template has been opened in your browser")
    logger.Info("2. Fill in the description and steps to reproduce")
    logger.Info("3. Paste the diagnostics (Ctrl+V / Cmd+V) in the \"Diagnostic Information\" section")
    logger.Info("4. Submit the issue")
    logger.Info("")
    logger.Success("Your diagnostic information is ready to paste!")

    return nil
}
```

## Flag Interactions

| Combination | Behavior | Notes |
|-------------|----------|-------|
| `--report-issue` + `--port` | Works normally | Includes custom port in diagnostics |
| `--report-issue` + `--verbose` | Ignored | Report-issue output is already informative |
| `--report-issue` + `--quiet` | Ignored | Can't suppress instructions |
| `--report-issue` + `--debug` | Ignored | Report-issue is diagnostic by nature |
| `--report-issue` + URL | Report-issue wins | Ignore URL, collect diagnostics |
| `--report-issue` + `--list-tabs` | Report-issue wins | Ignore list-tabs |
| `--report-issue` + `--all-tabs` | Report-issue wins | Ignore all-tabs |
| `--report-issue` + `--tab` | Report-issue wins | Ignore tab |
| `--report-issue` + `--open-browser` | Report-issue wins | Ignore open-browser |
| `--report-issue` + `--url-file` | Report-issue wins | Ignore url-file |
| `--report-issue` + `--output` | Report-issue wins | Ignore output |
| `--report-issue` + `--output-dir` | Report-issue wins | Ignore output-dir |
| `--report-issue` + `--format` | Report-issue wins | Ignore format |
| `--report-issue` + `--wait-for` | Report-issue wins | Ignore wait-for |
| `--report-issue` + `--timeout` | Report-issue wins | Ignore timeout |
| `--report-issue` + `--user-agent` | Report-issue wins | Ignore user-agent |
| `--report-issue` + `--user-data-dir` | Report-issue wins | Ignore user-data-dir |
| `--report-issue` + `--close-tab` | Report-issue wins | Ignore close-tab |
| `--report-issue` + `--force-headless` | Report-issue wins | Ignore force-headless |
| `--report-issue` + `--kill-browser` | Report-issue wins | Ignore kill-browser |
| `--report-issue` + `--doctor` | Report-issue wins | Ignore doctor (same diagnostics anyway) |
| `--report-issue` + `--version` | Version wins | Show version, ignore report-issue |
| `--report-issue` + `--help` | Help wins | Show help, ignore report-issue |

**Note:** Report-issue overrides everything except help (help is always highest priority)

## Documentation Updates Required

### 1. Create `docs/arguments/report-issue.md`

Follow existing argument documentation pattern:
- Description and purpose
- Workflow explanation
- Interaction matrix (table above)
- Examples
- Troubleshooting (clipboard/browser failures)

### 2. Update `docs/design-record.md`

Add to the "Arguments" section:
```markdown
- **Issue Reporting**: [Report Issue](arguments/report-issue.md)
```

Add design decision entry:
```markdown
### DD-XX: Report Issue Flag

**Decision:** Add `--report-issue` flag for streamlined bug reporting

**Rationale:**
- Reduce friction in bug reporting process
- Ensure diagnostic information is always included
- One command does everything: collect, copy, open browser
- Users more likely to report issues with good information

**Implementation:**
- Reuses `--doctor` diagnostic collection logic
- Copies diagnostics to system clipboard
- Opens GitHub issue with pre-filled template
- No external dependencies required (no `gh` CLI)
- Falls back gracefully if clipboard/browser unavailable

**URL Length Limitation:**
- Doctor output (~2,658 chars URL-encoded) exceeds browser limit (~2,048 chars)
- Solution: Clipboard + paste instead of URL encoding full diagnostics
- Template includes clear instruction where to paste
```

### 3. Review and update ALL `docs/arguments/*.md` files

Every argument document needs interaction rules with `--report-issue`:

Files to update:
- `docs/arguments/all-tabs.md`
- `docs/arguments/close-tab.md`
- `docs/arguments/debug.md`
- `docs/arguments/doctor.md` (new - note that report-issue reuses its logic)
- `docs/arguments/force-headless.md`
- `docs/arguments/format.md`
- `docs/arguments/help.md`
- `docs/arguments/kill-browser.md` (new)
- `docs/arguments/list-tabs.md`
- `docs/arguments/open-browser.md`
- `docs/arguments/output-dir.md`
- `docs/arguments/output.md`
- `docs/arguments/port.md`
- `docs/arguments/quiet.md`
- `docs/arguments/tab.md`
- `docs/arguments/timeout.md`
- `docs/arguments/url-file.md`
- `docs/arguments/user-agent.md`
- `docs/arguments/user-data-dir.md`
- `docs/arguments/verbose.md`
- `docs/arguments/version.md`
- `docs/arguments/wait-for.md`

For most arguments, add to their interaction matrix:
```markdown
| `--<flag>` + `--report-issue` | Report-issue wins | Issue reporting overrides normal operation |
```

### 4. Update `README.md`

Add to appropriate section (after Troubleshooting/Doctor):

```markdown
### Reporting Issues

Found a bug? Report it with one command:

```bash
snag --report-issue
```

This will:
1. Collect comprehensive diagnostic information
2. Copy it to your clipboard
3. Open a GitHub issue template in your browser
4. Guide you through filing a complete bug report

Just fill in what happened and paste the diagnostics!

**Requirements:**
- Internet connection (to open GitHub)
- Clipboard support on your system (xclip/xsel on Linux, pbcopy on macOS)

If clipboard doesn't work, diagnostics will be printed for manual copy.
```

### 5. Update `AGENTS.md`

Add to appropriate section:

```bash
# Report issues with diagnostics
snag --report-issue

# This replaces manually running doctor and creating GitHub issues
# Old way (manual):
#   snag --doctor > diag.txt
#   # Open browser, create issue, paste diag.txt
# New way (automatic):
#   snag --report-issue
```

## Testing

### Manual Test Cases

```bash
# Test 1: Basic report-issue
snag --report-issue
# Expected: Diagnostics copied, browser opens, instructions shown

# Test 2: Report-issue with custom port
snag --open-browser --port 9223
snag --report-issue --port 9223
# Expected: Port 9223 status included in diagnostics

# Test 3: Report-issue overrides URL
snag --report-issue https://example.com
# Expected: Opens issue report, ignores URL

# Test 4: Help overrides report-issue
snag --report-issue --help
# Expected: Shows help, ignores report-issue

# Test 5: Version overrides report-issue
snag --report-issue --version
# Expected: Shows version, ignores report-issue

# Test 6: Clipboard failure fallback
# (Simulate by temporarily removing xclip/pbcopy)
snag --report-issue
# Expected: Prints diagnostics, still opens browser

# Test 7: Browser open failure fallback
# (Simulate by temporarily breaking xdg-open/open)
snag --report-issue
# Expected: Prints URL for manual opening

# Test 8: Full failure scenario
# (No clipboard, no browser open capability)
snag --report-issue
# Expected: Prints diagnostics and URL for manual process

# Test 9: Verify clipboard contents
snag --report-issue
# Then paste in text editor
# Expected: Full doctor output

# Test 10: Verify GitHub URL
snag --report-issue
# Check browser URL
# Expected: github.com/p3bot/snag/issues/new with template
```

### Automated Tests

Add to `cli_test.go` or new `report_issue_test.go`:
- `TestReportIssueBasic()` - runs successfully, expected output
- `TestReportIssueOverridesURL()` - report-issue wins over URL
- `TestReportIssueWithPort()` - includes port in diagnostics
- `TestReportIssueHelpWins()` - help flag overrides report-issue
- `TestBuildIssueURL()` - URL format is correct
- `TestClipboardCopy()` - clipboard function works (if testable)

### Integration Test

Verify end-to-end flow:
1. Run `snag --report-issue`
2. Verify browser opens to correct GitHub URL
3. Verify template is pre-filled
4. Paste clipboard contents
5. Verify diagnostics appear correctly

## Implementation Phases

### Phase 1: Core Diagnostic Reuse
1. Ensure `--doctor` is implemented first (dependency)
2. Extract doctor collection logic into reusable function
3. Add `--report-issue` flag to main.go

### Phase 2: Clipboard Implementation
1. Create clipboard.go with platform-specific copy functions
2. Handle Linux (xclip/xsel), macOS (pbcopy)
3. Implement fallback for clipboard failure

### Phase 3: Browser Opening
1. Create browser opening function (xdg-open/open)
2. Build GitHub issue URL with template
3. Handle browser open failures

### Phase 4: Integration & Handler
1. Implement `handleReportIssue()` in handlers.go
2. Wire up all components
3. Add user instruction output

### Phase 5: Documentation
1. Create `docs/arguments/report-issue.md`
2. Update `docs/design-record.md`
3. Review and update ALL argument docs for interactions
4. Update README.md
5. Update AGENTS.md

### Phase 6: Testing & Polish
1. Manual testing on Linux and macOS
2. Test clipboard/browser fallback scenarios
3. Automated test cases
4. Output message refinement

## Success Criteria

- [ ] CLI flag `--report-issue` defined and working
- [ ] Reuses `--doctor` diagnostic collection logic
- [ ] Copies diagnostics to system clipboard (with fallback)
- [ ] Opens GitHub issue in system default browser (with fallback)
- [ ] GitHub URL contains properly formatted template
- [ ] Clear instructions displayed to user
- [ ] Works with `--port` flag
- [ ] Overrides all flags except `--help`
- [ ] Graceful fallbacks for clipboard/browser failures
- [ ] Works on Linux and macOS
- [ ] `docs/arguments/report-issue.md` created
- [ ] `docs/design-record.md` updated with design decision
- [ ] ALL `docs/arguments/*.md` files reviewed and updated
- [ ] README.md updated with reporting section
- [ ] AGENTS.md updated with examples
- [ ] Manual testing completed on both platforms
- [ ] Automated tests passing
- [ ] No regressions in existing functionality

## Open Implementation Questions

1. **Clipboard libraries vs exec commands:**
   - Option A: Use `github.com/atotto/clipboard` Go library
   - Option B: Use platform-specific commands (xclip, pbcopy)
   - **Recommendation:** Option B (no new dependency, consistent with snag philosophy)

2. **Template format - Markdown or plain text:**
   - GitHub issues support markdown
   - Should template include markdown formatting hints?
   - **Recommendation:** Use markdown (more useful for bug reports with code blocks)

3. **Labels:**
   - Should we pre-set labels like "bug", "needs-triage"?
   - Can only set labels via URL params if user has permissions
   - **Recommendation:** Don't set labels (might fail for non-collaborators)

4. **Issue template file:**
   - GitHub supports `.github/ISSUE_TEMPLATE/bug_report.md` files
   - Should we create that file instead of URL template?
   - **Recommendation:** Both - file for web UI, URL template for --report-issue

5. **Diagnostic format in clipboard:**
   - Plain text (current doctor format)
   - Markdown formatted (with code fences)
   - **Recommendation:** Markdown with code fence:
     ```
     ```
     <doctor output>
     ```
     ```

## Notes

- Requires `--doctor` to be implemented first (dependency)
- Consider creating `.github/ISSUE_TEMPLATE/bug_report.md` in repo for consistency
- Future enhancement: `--report-issue --feature` for feature requests (different template)
- Future enhancement: Auto-detect if user has internet connection before opening browser
- Future enhancement: Option to save diagnostics to file: `--report-issue --save diagnostics.txt`
- Consider adding to docs: How to report issues without `--report-issue` (for users on unsupported platforms)
