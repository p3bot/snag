# snag - Project Documentation

**Status**: Design Phase
**Language**: Go
**Type**: CLI Tool
**Started**: 2025-10-16
**Current Phase**: Technical Design & Architecture Planning

---

## Overview

`snag` is a CLI tool for intelligently fetching web page content using a browser engine, with smart session management and format conversion capabilities. It's designed to replace the current `get-webpage` Bash+Node.js implementation with a single, cross-platform Go binary.

**Core Value Proposition:**

- **Single binary** - No runtime dependencies (Node.js, npm, bash_modules)
- **Smart session management** - Auto-detect existing Chrome, preserves authentication
- **Format conversion** - Clean HTML to Markdown output
- **Authentication handling** - Detects auth requirements and launches visible browser
- **Cross-platform** - macOS, Linux, Windows support from one codebase

---

## Project Structure

```
snag/
├── PROJECT.md                  # This file - project documentation
├── NOTES.md                    # Development notes and minor items
├── .gitignore                  # Git ignore patterns
├── docs/
│   └── design.md              # Detailed design document (see below)
├── reference/                  # Reference documentation (gitignored)
│   ├── INDEX.csv              # Reference material catalog
│   ├── html-to-markdown/      # Go library for HTML→Markdown
│   ├── puppeteer/             # Puppeteer docs for CDP patterns
│   ├── cobra/                 # CLI framework reference
│   ├── chromedp/              # CDP library option #1
│   ├── rod/                   # CDP library option #2
│   └── get-webpage/           # Current Bash+Node.js implementation
│       ├── README.md          # Implementation analysis
│       ├── get-webpage        # Main Bash script
│       ├── bash_modules/      # Terminal utilities
│       └── lib/chromium/      # Puppeteer core logic
└── [future: cmd/, internal/, pkg/]
```

---

## Documentation

### Primary Documents

**Design Document**: `docs/design.md`

- Complete feature specification
- CLI arguments specification (16 MVP arguments)
- Architecture diagrams
- Technology stack rationale
- Distribution strategy
- Success criteria and MVP scope
- Migration plan from current implementation
- Design decisions (completed and pending)

**Development Notes**: `NOTES.md`

- Minor items and considerations
- Testing requirements
- Position-independent argument parsing notes
- Future considerations

**Reference Implementation**: `reference/get-webpage/`

- Current working Bash+Node.js code
- Battle-tested patterns for:
  - Browser session detection
  - Authentication handling
  - Tab management
  - URL normalization
  - Error handling and exit codes

### Reference Materials

All reference materials are cataloged in `reference/INDEX.csv`:

- **html-to-markdown**: Go library for HTML conversion
- **puppeteer**: CDP patterns and implementation reference
- **cobra**: CLI framework for Go
- **chromedp**: Primary CDP library candidate
- **rod**: Alternative CDP library candidate
- **get-webpage**: Current implementation to port

---

## Current Phase: Design Completion

### Design Decisions Completed ✅

#### 1. **CLI Arguments & Argument Parsing** ✅

- **Decision Made**: 16 arguments for MVP (position independent)
- **Key Arguments**:
  - Core: `<url>`, `--version`, `-h/--help`
  - Output: `-o/--output`, `--format markdown|html`
  - Page Loading: `-t/--timeout`, `-w/--wait-for`
  - Browser: `-p/--port`, `-c/--close-tab`, `-fh/--force-headless`, `-fv/--force-visible`, `-ob/--open-browser`
  - Logging: `-v/--verbose`, `-q/--quiet`, `--debug`
  - Request: `--user-agent`
- **Changes from get-webpage**:
  - ✅ Added: `--version`, `--quiet`, `--user-agent`
  - ✅ Replaced `--html` with `--format` for extensibility
  - ✅ Position independent (flags before or after URL both work)
- **Rationale**:
  - Preserves all original `get-webpage` functionality
  - Adds essential modern CLI features (version, quiet mode, user-agent)
  - Extensible format flag supports future formats (text, pdf)
  - Position independence is Go CLI framework default
- **Complete Specification**: `docs/design.md:89-137`
- **Status**: ✅ Complete

#### 2. **Output Formats** ✅

- **Decision Made**: MVP supports `markdown` (default) and `html` only
- **Post-MVP Formats**: `text`, `pdf` (separate enhancement projects)
- **Rationale**:
  - Keeps MVP scope focused on core use cases
  - Extensible `--format` flag design accommodates future formats
  - Text extraction and PDF rendering add significant complexity
- **Status**: ✅ Complete

#### 3. **Platform Support** ✅

- **Decision Made**: MVP targets macOS and Linux only; Windows is post-MVP
- **MVP Platforms**:
  - darwin/arm64 (macOS Apple Silicon)
  - darwin/amd64 (macOS Intel)
  - linux/amd64 (Linux 64-bit)
  - linux/arm64 (Linux ARM - Raspberry Pi, servers)
- **Post-MVP**: Windows support (requires Windows-specific path handling)
- **Rationale**:
  - Primary development/use on macOS and Linux
  - Windows adds complexity (path conventions, file handling)
  - Can add later without breaking existing users
- **Status**: ✅ Complete

#### 4. **Config File Support** ✅

- **Decision Made**: No config file support for MVP; post-MVP feature
- **Post-MVP Consideration**: `.snagrc` file for default preferences
- **Rationale**:
  - CLI flags are sufficient for core functionality
  - Most users will use defaults (30s timeout, markdown format, auto-detect Chromium)
  - Power users can use shell aliases
  - Adds complexity (file parsing, precedence rules)
- **Status**: ✅ Complete

#### 5. **HTML→Markdown Conversion** ✅

- **Decision Made**: Embed `github.com/JohannesKaufmann/html-to-markdown/v2` Go library
- **Library**: html-to-markdown v2 (MIT license)
- **Rationale**:
  - Current `html2markdown` CLI is a wrapper around this exact library
  - Proven output quality (already using it)
  - No external dependencies (single binary)
  - Simple API, well-maintained
  - Supports CommonMark, tables, strikethrough
- **Reference**: `reference/html-to-markdown/`
- **Status**: ✅ Complete

#### 6. **License Attribution** ✅

- **Decision Made**: Use `LICENSES/` directory for third-party license attribution
- **Approach**:
  - Create `LICENSES/` directory in repository
  - Include each dependency's license as separate file
  - Complies with MIT license requirements
  - Visible in GitHub, included in source releases
- **Post-MVP**: Consider `snag --licenses` command
- **Automation**: Use `go-licenses` tool during build/release
- **Status**: ✅ Complete

#### 7. **CLI Framework Choice** ✅

- **Decision Made**: Use `github.com/urfave/cli` for CLI framework
- **Library**: urfave/cli v2 (MIT license)
- **Rationale**:
  - Smaller binary size compared to Cobra (important for single-binary tool)
  - Simpler, less boilerplate code
  - Better dynamic bash autocompletion (can autocomplete argument values)
  - Still supports subcommands for Phase 2 (--list-tabs, --tab)
  - Declarative, clean API
  - Widely used (23.6k GitHub stars), well-maintained
  - Less globals-heavy than Cobra's architectural pattern
- **Alternatives Considered**:
  - Cobra: More feature-rich but larger binaries, more dependencies
  - Coral: Cobra fork with fewer dependencies but less mature (431 stars)
  - Standard library flag package: Too basic, no subcommand support
- **Reference**: `reference/urfave-cli/`
- **Status**: ✅ Complete

#### 8. **CDP Library Choice** ✅

- **Decision Made**: Use `github.com/go-rod/rod` for Chrome DevTools Protocol
- **Library**: rod (MIT license)
- **Rationale**:
  - Simpler, more intuitive API compared to chromedp
  - Better resource efficiency (uses less CPU/memory)
  - More stable architecture with consistent CDP versioning
  - Auto-wait elements feature reduces error handling complexity
  - Chained context design for intuitive timeout/cancel
  - Debugging-friendly with auto input tracing
  - API more closely resembles Puppeteer (easier porting)
  - Perfect fit for passive content fetching use case
- **Alternatives Considered**:
  - chromedp: Faster raw speed but steeper learning curve, more resource usage
  - Direct CDP: Too low-level, too much work
- **Reference**: `reference/rod/`
- **Status**: ✅ Complete

#### 9. **Chrome/Chromium Discovery** ✅

- **Decision Made**: Use rod's built-in `launcher.LookPath()` for browser discovery
- **Approach**:
  - Three-tier strategy: connect to existing → find system browser → launch
  - No auto-download, no config file, no environment variable needed
  - rod's comprehensive path detection handles everything
- **Supported Browsers** (Chromium-based only):
  - Google Chrome, Chromium, Microsoft Edge, Brave, Chrome Canary
  - Firefox NOT supported (CDP deprecated in Firefox, moved to WebDriver BiDi)
- **Platform Coverage**:
  - macOS: `/Applications/*.app`, `/usr/bin/*`
  - Linux: `/usr/bin/*`, system PATH
- **Rationale**:
  - rod's `LookPath()` is comprehensive and battle-tested
  - Cross-platform support built-in
  - Zero maintenance - rod team keeps paths updated
  - Clear error message if no browser found
- **Reference**: `reference/rod/lib/launcher/browser.go:202-251`
- **Status**: ✅ Complete

#### 10. **Logging & Output Strategy** ✅

- **Decision Made**: Simple custom logger with colored output, no external dependencies
- **Output Routing**:
  - stdout: Content only (HTML/Markdown) - enables piping
  - stderr: All logs, warnings, errors, progress indicators
- **Log Levels**: Quiet (fatal only), Normal (default), Verbose, Debug
- **Color Support**: Auto-detect TTY + NO_COLOR environment variable
- **Emojis**: Use by default (✓ ⚠ ✗) - modern terminals support UTF-8
- **Format**: Clean, no timestamps, emoji indicators
- **Why Custom Logger**:
  - Standard `log` too basic (timestamps, no levels)
  - `log/slog` too verbose/structured for CLI
  - Custom: ~100 lines, exactly what we need, zero deps
- **Reference**: `reference/get-webpage/bash_modules/terminal.sh`
- **Status**: ✅ Complete

#### 11. **Error Handling & Exit Codes** ✅

- **Decision Made**: Simple exit codes (0/1) with clear error messages
- **Exit Codes**:
  - 0: Success (content fetched and output)
  - 1: Any error (network, browser, auth, timeout, validation, conversion)
- **Sentinel Errors**: For internal logic/testing (not exit codes)
- **Error Wrapping**: Use `fmt.Errorf("%w")` for context
- **Error Messages**: Clear + actionable (explain problem + suggest fix)
- **Rationale**:
  - Modern CLI best practice: keep it simple (gh, kubectl use 0/1)
  - Multiple exit codes hard to document/discover
  - Error messages more useful than exit code numbers
  - Most scripts just check `$? != 0`
- **Status**: ✅ Complete

#### 12. **Project Structure** ✅

- **Decision Made**: Flat structure at repository root, refactor later if needed
- **Structure**: main.go, browser.go, fetch.go, convert.go, logger.go, errors.go at root
- **Build**: `go build -o snag` (simple!)
- **Rationale**:
  - Simple for focused single-binary CLI
  - No over-engineering for MVP (<2000 lines)
  - Simpler Homebrew formula
  - Easy to refactor to internal/ later if complexity grows
  - Go philosophy: "start simple, refactor as needed"
- **Distribution**: Perfect for Homebrew, direct downloads, go install
- **Status**: ✅ Complete

#### 13. **Testing Strategy** ✅

- **Decision Made**: Integration tests with real Chrome/Chromium browser
- **Test Approach**: Real browser via rod, test fixtures in testdata/, local HTTP server
- **Coverage**: Normal fetch, auth detection, timeouts, error conditions, CLI flags
- **Rationale**:
  - Blackbox testing matches user experience
  - Real browser catches integration issues early
  - No complex mocking needed
  - GitHub Actions has Chrome pre-installed
- **Status**: ✅ Complete

### Design Decisions Complete ✅

**All 13 design decisions finalized!** Ready for implementation phase.

---

## Technology Stack

### Confirmed Decisions

- **Language**: Go 1.21+
- **Distribution**: Homebrew + direct binary download
- **Target Platforms**: darwin/arm64, darwin/amd64, linux/amd64, linux/arm64 ✅
- **Browser Engine**: Chromium (Chrome also supported) via CDP
- **CLI Arguments**: 16 MVP arguments (position independent) ✅
- **CLI Framework**: urfave/cli v2 ✅
- **CDP Library**: rod ✅
- **Browser Discovery**: rod's launcher.LookPath() (Chromium-based browsers only) ✅
- **Logging Strategy**: Custom logger with colors, emojis, 4 levels (quiet/normal/verbose/debug) ✅
- **Error Handling**: Simple exit codes (0/1), sentinel errors, clear messages ✅
- **Project Structure**: Flat structure at root (main.go, browser.go, etc.) ✅
- **Testing Strategy**: Integration tests with real browser ✅
- **Output Formats**: markdown (default), html ✅
- **HTML Conversion**: Embed html-to-markdown/v2 Go library ✅
- **License Attribution**: LICENSES/ directory ✅
- **Config File**: Post-MVP (MVP uses flags only) ✅

### All Design Decisions Complete! ✅

---

## Reference Implementation Analysis

The current `get-webpage` tool (Bash + Node.js + Puppeteer) provides proven patterns:

### Key Features to Port

1. **Smart Browser Detection**

   - Check for running Chrome on debug port
   - Auto-select connect/headless/visible mode
   - Port: `reference/get-webpage/get-webpage:282-300`

2. **Authentication Detection**

   - HTTP status codes (401, 403)
   - Page content analysis (login forms, OAuth redirects)
   - Port: `reference/get-webpage/lib/chromium/fetch-html.js:194-261`

3. **Three Operation Modes**

   - Connect: Attach to existing Chrome (preserves auth)
   - Headless: Launch background Chrome (fast, clean)
   - Visible: Launch visible Chrome (user authentication)
   - Port: `reference/get-webpage/lib/chromium/fetch-html.js:41-84`

4. **Tab Management**

   - Headless: Always close tabs
   - Connect/Visible: Keep tabs open unless `--close-tab`
   - Port: `reference/get-webpage/lib/chromium/fetch-html.js:158-173`

5. **URL Normalization**
   - Auto-add `https://` for domains
   - Auto-add `http://` for localhost/IPs
   - Support bare domains and paths
   - Port: `reference/get-webpage/get-webpage:230-253`

### Architecture to Replicate

```
┌─────────────────────────────────────────────┐
│           CLI Argument Parsing              │
│  Parse flags, validate inputs               │
└─────────────┬───────────────────────────────┘
              │
┌─────────────▼───────────────────────────────┐
│         Browser Session Manager             │
│  - Detect existing Chrome                   │
│  - Select mode (connect/headless/visible)   │
│  - Launch or connect to browser             │
└─────────────┬───────────────────────────────┘
              │
┌─────────────▼───────────────────────────────┐
│           Page Fetcher                      │
│  - Navigate to URL                          │
│  - Wait for page load                       │
│  - Detect authentication                    │
│  - Extract HTML content                     │
└─────────────┬───────────────────────────────┘
              │
┌─────────────▼───────────────────────────────┐
│         Content Converter                   │
│  - HTML → Markdown conversion               │
│  - Output formatting                        │
│  - Write to file or stdout                  │
└─────────────────────────────────────────────┘
```

---

## Next Steps

### Design Phase ✅ COMPLETE

1. ✅ Create `.gitignore`
2. ✅ Clone reference repositories
3. ✅ Copy current implementation to reference/
4. ✅ Create PROJECT.md (this document)
5. ✅ Create NOTES.md for minor items
6. ✅ **Design CLI arguments** (16 MVP arguments defined)
7. ✅ **Design output formats** (markdown + html for MVP)
8. ✅ Update `docs/design.md` with CLI argument decisions
9. ✅ **Make all 13 design decisions** (100% complete)
10. ✅ Finalize `docs/design.md` with all decisions
11. ⏳ Create architectural diagrams if needed (optional)

### Implementation Phase (Next)

1. Initialize Go module (`go mod init github.com/p3bot/snag`)
2. Create flat project structure (main.go, browser.go, fetch.go, convert.go, logger.go, errors.go)
3. Implement CLI framework with urfave/cli and 16 MVP arguments
4. Implement custom logger with 4 levels and color support
5. Implement browser detection and connection using rod
6. Implement page fetch logic with authentication detection
7. Implement HTML to Markdown conversion using html-to-markdown/v2
8. Add error handling with sentinel errors and clear messages
9. Write integration tests with real browser
10. Create testdata/ fixtures for testing
11. Create README.md with usage examples
12. Set up GitHub Actions for multi-platform builds
13. Create Homebrew tap and formula
14. Tag v1.0.0 release

---

## Success Criteria (MVP)

From `docs/design.md:317-334`:

**MVP Complete When:**

- [ ] Fetch URL and output Markdown to stdout
- [ ] Detect and connect to existing Chrome instance
- [ ] Launch headless Chrome when needed
- [ ] Detect authentication requirements
- [ ] Launch visible Chrome for auth flows
- [ ] Save output to file with `-o` flag
- [ ] Support `--format html` for raw HTML output
- [ ] Support `--format markdown` (default) for Markdown output
- [ ] Implement `--version` flag
- [ ] Implement `--quiet` mode (suppress logs, show only content/errors)
- [ ] Implement `--user-agent` custom user agent support
- [ ] Position-independent argument parsing
- [ ] All 16 MVP arguments functional
- [ ] Homebrew formula working
- [ ] Basic documentation (README, --help)
- [ ] Test suite (unit + integration tests)

**Quality Gates:**

- [ ] Cross-platform builds (macOS arm64/amd64, Linux amd64)
- [ ] Unit tests for core functions
- [ ] Integration test with real websites
- [ ] Error handling for common failures
- [ ] Clean logging with `--verbose` and `--quiet` flags
- [ ] Exit codes match get-webpage convention (0, 1, 2, 3)

---

## Links & Resources

### Internal Documentation

- **Design Document**: `docs/design.md`
- **Reference Implementation**: `reference/get-webpage/README.md`
- **Reference Index**: `reference/INDEX.csv`

### External Resources

- **Current Tool**: `~/bin/scripts/get-webpage`
- **Homebrew Tap**: (future) `p3bot/tap/snag`
- **Repository**: (future) `github.com/p3bot/snag`

### Libraries Under Consideration

- chromedp: https://github.com/chromedp/chromedp
- rod: https://github.com/go-rod/rod
- html-to-markdown: https://github.com/JohannesKaufmann/html-to-markdown
- cobra: https://github.com/spf13/cobra
- urfave/cli: https://github.com/urfave/cli

---

## Session Summary & Key Learnings

### 2025-10-16: CLI Arguments Design Session

**Completed Work:**

1. ✅ Analyzed existing `get-webpage` arguments (13 arguments)
2. ✅ Designed new CLI argument structure (16 MVP arguments)
3. ✅ Made key design decisions:
   - Position-independent argument parsing (Go standard)
   - Replaced `--html` with extensible `--format markdown|html`
   - Added `--version`, `--quiet`, `--user-agent` as essential features
   - Deferred advanced formats (`text`, `pdf`) to post-MVP
4. ✅ Updated `docs/design.md` with complete argument specification
5. ✅ Created `NOTES.md` for tracking minor items
6. ✅ Defined 8 post-MVP feature projects

**Key Insights:**

- Position independence comes naturally with Go CLI frameworks (cobra, urfave/cli)
- Format extensibility important: `--format` supports future formats better than flags
- User agent customization essential for bypassing headless browser detection
- Test suite required but not originally in design document
- MVP scope well-defined: 2 formats (markdown, html), advanced formats post-MVP

**Post-MVP Features Identified:**

1. Text format extraction (plain text only)
2. PDF export (Chrome native rendering)
3. Screenshot capture
4. JavaScript control (--no-js)
5. Cookie management
6. Advanced headers (repeatable --header flag)
7. Redirect control
8. Proxy support

**Remaining Design Decisions:** 8

- CLI framework choice (cobra vs urfave/cli vs flag)
- CDP library (chromedp vs rod)
- HTML→Markdown conversion approach
- Chrome/Chromium discovery strategy
- Logging & output strategy
- Error handling & exit codes
- Configuration strategy
- Platform priorities
- Project structure
- Testing strategy

---

## Notes

- This project replaces the current `get-webpage` tool entirely
- The new tool will be distributed via Homebrew for easy installation
- Focus is on simplicity and reliability - not a full browser automation framework
- Philosophy: "passive observer" for content retrieval, not active automation
- All reference materials are git-ignored to keep repository clean
- See `NOTES.md` for minor items and testing considerations

---

**Last Updated**: 2025-10-16
**Phase**: Design Complete ✅
**Next Milestone**: Begin implementation - initialize Go module and set up project structure
**Progress**: 100% design phase complete - Ready to code!
