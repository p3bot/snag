# Contributing to snag

Thank you for your interest in contributing to snag! This document provides guidelines and instructions for contributing.

## Quick Links

- [Issues](https://github.com/p3bot/snag/issues) - Report bugs or request features
- [Pull Requests](https://github.com/p3bot/snag/pulls) - Submit code changes
- [AGENTS.md](AGENTS.md) - Detailed technical documentation for AI agents and developers

## Ways to Contribute

- Report bugs
- Suggest new features or improvements
- Improve documentation
- Submit pull requests with bug fixes or features
- Help answer questions in issues

## Reporting Bugs

When reporting bugs, please:

1. Check [existing issues](https://github.com/p3bot/snag/issues) first
2. Use the bug report template when creating a new issue
3. Include the output from `snag --doctor` (automatically collects diagnostics)
4. Provide:
   - snag version: `snag --version`
   - Operating system and version
   - Full command that triggered the bug
   - Complete error message or unexpected output
   - Output from `--debug` flag (if applicable)

## Development Setup

### Prerequisites

- Go 1.25.3 or later
- Chromium-based browser (Chrome, Chromium, Edge, Brave)
- Git

### Getting Started

```bash
# Clone the repository
git clone https://github.com/p3bot/snag.git
cd snag

# Install dependencies
go mod download

# Build the project
go build -o snag ./cmd/snag

# Verify installation
./snag --version
```

### Project Structure

snag uses `cmd/snag` plus `internal/` packages:

- `cmd/snag/main.go` - Thin entry point (run CLI, map errors to exit codes)
- `internal/cli` - Cobra command, flags, handlers, signals
- `internal/browser` - Browser and tab management (rod)
- `internal/fetch` - Page navigation and content extraction
- `internal/format` - Content conversion (HTML to Markdown/Text, PDF, PNG)
- `internal/logger` - Custom logger (4 levels, stderr only)
- `internal/validate` - Input validation
- `internal/output` - Filename generation and conflict resolution
- `internal/doctor` - Diagnostic information collection

Tests live beside the code they exercise. Shared fixtures are in module-root `testdata/`.

## Running Tests

```bash
# Run all tests
go test -v ./...

# Run specific package
go test -v ./internal/validate

# Run with coverage
go test -v -cover ./...

# Run browser integration tests (requires Chrome/Chromium)
go test -v ./internal/cli -run TestBrowser
```

**Test requirements:**
- Unit tests: No browser required
- Integration tests: Chrome/Chromium must be installed

## Code Style

### Go Conventions

- Follow standard Go formatting: `gofmt` or `goimports`
- Use Go 1.25+ features and idioms
- Keep functions focused and small
- Use descriptive variable names

### Project-Specific Patterns

**Output Routing (Critical for piping):**
- `stdout`: Content only (HTML/Markdown/Text or binary formats)
- `stderr`: All logs, warnings, errors, progress indicators

**Naming Conventions:**
- Exported constants: `format.Markdown`, `format.HTML`
- Sentinel errors: `browser.ErrBrowserNotFound`, `fetch.ErrPageLoadTimeout`
- Functions: Descriptive verbs - `validate.URL()`, `fetchPage()`, `convertToMarkdown()`

**Error Handling:**
- Use sentinel errors defined in the owning package
- Wrap errors with context: `fmt.Errorf("failed to navigate to %s: %w", url, err)`
- Clear, actionable error messages via logger
- Never panic for expected errors

**Logging:**
- Use custom Logger with 4 levels (quiet, normal, verbose, debug). Quiet is internal (errors-only); there is no `--quiet` flag
- User-facing logging flags: `--verbose` and `--debug` (mutually exclusive)
- Default (normal): warnings, errors, and result announcements. Connection, fetch, and batch progress is silent
- Log to stderr only (stdout reserved for content)
- `logger.Success()` - Result announcements
- `logger.Error()` - Errors
- `logger.Info()` - Info messages (recovery hints, filenames)
- `logger.Verbose()` - Fetch/connection/batch progress and extra detail
- `logger.Debug()` - CDP/debug messages

**License Headers:**

Add MPL 2.0 header to all new `.go` files:

```go
// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package <name>
```

### Code Quality

Before submitting:

```bash
# Format code
gofmt -w .

# Vet code
go vet ./...

# Run tests
go test -v ./...
```

## Branch Management

- Main branch: `main`
- Feature branches: `feature/description`
- Bug fix branches: `fix/description`
- Always work on feature branches, PR to `main`

## Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/) style:

```
feat(browser): add support for custom user agents
fix(fetch): resolve WebSocket URL from HTTP endpoint
docs: reorganize project documentation
chore(deps): update rod to v0.116.2
test(validate): add URL validation test cases
refactor(logger): simplify error formatting
```

**Types:**
- `feat` - New feature
- `fix` - Bug fix
- `docs` - Documentation changes
- `chore` - Maintenance tasks, dependencies
- `test` - Test additions or changes
- `refactor` - Code restructuring without behavior change

**Format:**
```
<type>(<scope>): <short description>

[optional body]

[optional footer]
```

## Pull Request Process

1. **Create an issue first** (for significant changes)
2. **Fork the repository** and create a feature branch
3. **Make your changes:**
   - Write clear, focused commits
   - Follow code style guidelines
   - Add/update tests as needed
   - Update documentation if changing CLI interface
4. **Test thoroughly:**
   - Run all tests: `go test -v ./...`
   - Build successfully: `go build -o snag ./cmd/snag`
   - Test manually with `./snag`
5. **Submit pull request:**
   - Title: `[component] Brief description`
   - Reference related issue(s)
   - Describe what changed and why
   - Keep PRs focused on a single feature/fix
6. **Respond to review feedback**

### Pull Request Guidelines

- Run tests and build before submitting
- Ensure code is formatted with `gofmt`
- Update relevant documentation
- Keep PRs focused on a single feature or fix
- Add tests for new functionality
- Ensure all tests pass

## Documentation

When changing functionality:

- Update `README.md` for user-facing changes
- Update `AGENTS.md` for technical implementation details
- Update `docs/arguments/*.md` if adding/changing CLI flags
- Add design decisions to `.agents/design/design-records/` for architectural changes

## Testing Philosophy

- **Unit tests**: Pure functions without mocking (validate, format, browser detection)
- **Integration tests**: Real Chrome/Chromium browser (no mocking browser interactions)
- Focus on behavior, not implementation details
- Test error cases and edge conditions

## Questions or Need Help?

- Open an issue for questions
- Check [AGENTS.md](AGENTS.md) for detailed technical documentation
- Review existing code and tests for examples

## License

By contributing to snag, you agree that your contributions will be licensed under the [Mozilla Public License 2.0](LICENSE).

## Code of Conduct

Be respectful, professional, and constructive in all interactions. We welcome contributors of all experience levels.
