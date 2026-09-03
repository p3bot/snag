# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Breaking

- Successful fetch progress (browser launched/connected, fetching, fetched successfully, batch counters) is silent by default; use `--verbose` to print it. Warnings, error recovery hints, and result announcements (generated filename, saved-to, browser opened) still print
- Removed `--quiet` / `-q`; passing them is an unknown-flag error
- `--info` no longer forces errors-only stderr; warnings, recovery hints, and save announcements print; fetch progress stays verbose-only

### Changed

- Layout is now `cmd/snag` plus `internal/` packages (no root `package main`)
- Install with `go install github.com/p3bot/snag/cmd/snag@latest`
- Version ldflags symbol is `github.com/p3bot/snag/internal/cli.Version`

### Fixed

- Fetch waits for `window.onload` (bounded by `--timeout`) instead of Rod `WaitStable`, so animated SPAs no longer hang and slow loads are not clipped at 3s. After load, a short HTTP-idle wait is best-effort; polling does not fail the fetch. If `load` does not fire in time, snag errors instead of printing a half page
- JavaScript `alert`/`confirm`/`prompt` dialogs are dismissed so they cannot block content extraction
- `--info` domain field no longer truncates IPv6 hostnames
- Auto-generated filenames no longer mangle IPv6 hosts when falling back to a URL slug
- `--doctor` counts tabs via Chrome's HTTP `/json/list` instead of attaching a DevTools session

## [1.1.0] - 2026-02-04

### Added

- New `--info` flag for page metadata export (JSON output with title, URL, domain, slug, timestamp)
- Repository and issue links in version output (`snag --version`)
- License, Go Report Card, Go Reference, and GitHub Release badges to README

### Changed

- Update html-to-markdown to v2.5.0 (table cell padding option, cleaner ampersand handling)
- Update cobra to v1.10.2
- Update pflag to v1.0.10
- Update golang.org/x/net to v0.47.0

### Fixed

- Page content not filling browser window in visible mode (clear Rod's default viewport emulation)
- Set 1920x1080 viewport for headless mode for better quality screenshots and PDFs
- Remove unused validation code and lint warnings

## [1.0.0] - 2025-11-01

### Added

- Initial stable release of snag
- Intelligent web page fetching using Chrome DevTools Protocol
- Five output formats: Markdown (default), HTML, Text, PDF, PNG
- Smart browser management (auto-detect existing browser or launch headless)
- Tab management capabilities:
  - List all open browser tabs
  - Fetch content from specific tabs by index or URL pattern
  - Pattern matching with exact URL, substring, and regex support
  - Fetch all tabs at once with auto-generated filenames
- Authentication support via persistent browser sessions
- Multiple URL processing in single command
- Output options:
  - Write to stdout (default for text formats)
  - Save to file with `--output`
  - Auto-generate filenames to directory with `--output-dir`
  - Auto-generated timestamped filenames for binary formats (PDF, PNG)
- Page loading controls:
  - Configurable timeout with `--timeout`
  - Wait for specific CSS selectors with `--wait-for`
  - Custom user agent support with `--user-agent`
  - Custom Chrome profile support with `--user-data-dir`
- Browser controls:
  - Force headless mode with `--force-headless`
  - Open visible browser with `--open-browser`
  - Control tab closing behavior with `--close-tab`
  - Custom remote debugging port with `--port`
  - Kill orphaned browser processes with `--kill-browser`
- Logging levels: quiet, normal, verbose, debug
- Diagnostic information with `--doctor` flag
- Cross-platform support: macOS (arm64/amd64), Linux (amd64/arm64)
- Single binary distribution with no runtime dependencies
- Comprehensive documentation and examples

[unreleased]: https://github.com/p3bot/snag/compare/v1.1.0...HEAD
[1.1.0]: https://github.com/p3bot/snag/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/p3bot/snag/releases/tag/v1.0.0
