# snag

[![License: MPL 2.0](https://img.shields.io/badge/License-MPL_2.0-brightgreen.svg)](https://opensource.org/licenses/MPL-2.0)
[![Go Report Card](https://goreportcard.com/badge/github.com/p3bot/snag)](https://goreportcard.com/report/github.com/p3bot/snag)
[![Go Reference](https://pkg.go.dev/badge/github.com/p3bot/snag.svg)](https://pkg.go.dev/github.com/p3bot/snag)
[![GitHub Release](https://img.shields.io/github/v/release/p3bot/snag)](https://github.com/p3bot/snag/releases)

Intelligently fetch web page content using a browser engine.

## Why snag?

**Built for AI agents to consume web content efficiently.**

Modern AI agents need web content in clean, token-efficient formats. snag solves this by:

- **Markdown output** - AI models work better with markdown than HTML (70% fewer tokens)
- **Real browser rendering** - Handles JavaScript, dynamic content, lazy loading automatically
- **Authentication support** - Access private/authenticated pages through persistent browser sessions
- **Tab management** - List, select, and reuse existing browser tabs without creating new ones
- **Content archival** - Build reference libraries of web content for future AI agent use
- **Simple CLI interface** - One command, clean output, no complex automation scripts

**Perfect for:**

- AI coding assistants fetching documentation
- Building knowledge bases from authenticated sites
- Capturing dynamic web content for analysis
- Piping web content into AI processing pipelines
- Taking page screenshots for CSS/Style analysis

## Quick Start

```bash
# Install via Homebrew
brew tap p3bot/tap
brew trust p3bot/tap
brew install p3bot/tap/snag

# Fetch a page as Markdown (default format)
snag example.com

# Save to file
snag docs.example.com > docs.md
```

That's it! snag auto-detects your Chromium-based browser and handles everything else.

## Installation

### Prerequisites

snag requires a Chromium-based (Chrome) browser:

**Linux:**

```bash
# Ubuntu/Debian
sudo apt update && sudo apt install chromium-browser

# Fedora
sudo dnf check-update && sudo dnf install chromium

# Arch Linux
sudo pacman -Sy chromium

# Homebrew
brew install chromium
```

**macOS:**

```bash
# Chromium (recommended) via Homebrew
# or Chrome - download from https://www.google.com/chrome/
brew install chromium
```

**Supported browsers:** Chrome, Chromium, Microsoft Edge, Brave, other Chromium-based browsers

### Homebrew (Linux/macOS)

Note: There's a name conflict with an older deprecated tool. Use the full tap name:

```bash
brew tap p3bot/tap
brew trust p3bot/tap
brew install p3bot/tap/snag
```

### Go Install

```bash
go install github.com/p3bot/snag@latest
```

### Build from Source

```bash
git clone https://github.com/p3bot/snag.git
cd snag
go build
./snag --version
```

## Usage

### Basic Examples

```bash
# Fetch page as Markdown (default)
snag example.com
snag https://example.com

# Save to file
snag -o output.md https://example.com
snag example.com > output.md

# Get raw HTML instead
snag --format html https://example.com

# Get plain text only (strips all HTML)
snag --format text https://example.com

# Quiet mode (content only, no logs)
snag --quiet https://example.com

# Get page metadata as JSON
snag --info https://example.com

# Wait for dynamic content to load
snag --wait-for ".content-loaded" https://dynamic-site.com

# Increase timeout for slow sites
snag --timeout 60 https://slow-site.com

# Verbose logging for debugging
snag --verbose https://example.com
```

## Output Formats

snag supports 5 output formats for different use cases. Format names are case-insensitive and support aliases for convenience.

### Text Formats

**Markdown (default):**

Clean, readable text format optimized for AI agents and documentation. Uses 70% fewer tokens than HTML.

```bash
# Default format (no flag needed)
snag https://example.com

# Explicit format
snag --format md https://example.com

# Alias also works (backward compatibility)
snag --format markdown https://example.com

# Case-insensitive
snag --format MD https://example.com
snag --format Markdown https://example.com
```

**HTML:**

Raw HTML output, preserving original page structure.

```bash
# Get raw HTML
snag --format html https://example.com

# Case-insensitive
snag --format HTML https://example.com
```

**Text:**

Plain text only, strips all HTML tags and formatting.

```bash
# Extract plain text
snag --format text https://example.com

# Alias also works
snag --format txt https://example.com

# Case-insensitive
snag --format TEXT https://example.com
```

### Binary Formats (PDF, PNG)

Binary formats automatically generate filenames to prevent terminal corruption. Files are saved to the current directory unless you specify a location.

**PDF:**

Visual rendering as a PDF document using Chrome's native rendering engine.

```bash
# Auto-generates filename in current directory
snag --format pdf https://example.com
# Creates: 2025-10-22-142033-example-domain.pdf

# Specify custom filename
snag --format pdf -o report.pdf https://example.com

# Save to specific directory with auto-generated name
snag --format pdf -d ~/Downloads https://example.com
# Creates: ~/Downloads/2025-10-22-142033-example-domain.pdf

# Case-insensitive
snag --format PDF https://example.com
```

**PNG:**

Full-page screenshot as a PNG image.

```bash
# Auto-generates filename in current directory
snag --format png https://example.com
# Creates: 2025-10-22-142033-example-domain.png

# Specify custom filename
snag --format png -o screenshot.png https://example.com

# Save to specific directory with auto-generated name
snag --format png -d ~/screenshots https://example.com
# Creates: ~/screenshots/2025-10-22-142033-example-domain.png

# Case-insensitive
snag --format PNG https://example.com
```

**Why auto-generate filenames?**

Binary formats (PDF, PNG) cannot output to stdout because binary data corrupts terminal display. When you don't specify `-o` or `-d`, snag automatically generates a timestamped filename in the current directory.

**Auto-generated filename format:**

```
yyyy-mm-dd-hhmmss-{page-title-slug}.{ext}
Example: 2025-10-22-142033-github-snag-repo.png
```

### Page Info (JSON Metadata)

Get page metadata as JSON for automation scripts. Useful for extracting page titles, generating directory names, or building indexes.

```bash
# Get page metadata as JSON
snag --info https://example.com
snag -i https://example.com

# Output:
# {
#   "title": "Example Domain",
#   "url": "https://example.com/",
#   "domain": "example.com",
#   "slug": "example-domain",
#   "timestamp": "2025-02-04T14:30:22+10:00"
# }

# Save info to file
snag --info -o page-info.json https://example.com

# Get info from existing tab
snag --info --tab 1
snag -i -t "github"

# Use with jq for scripting
title=$(snag -i example.com | jq -r '.title')
slug=$(snag -i example.com | jq -r '.slug')
```

**JSON fields:**

| Field | Description |
|-------|-------------|
| title | Page title from `<title>` tag |
| url | Final URL after redirects |
| domain | Domain name (without www. prefix) |
| slug | URL-safe slug from title (for filenames) |
| timestamp | ISO 8601 timestamp of fetch |

**Notes:**

- `--info` is mutually exclusive with `--format` (always outputs JSON)
- Output is quiet by default (no log messages, only JSON)
- Use `--verbose` to see log messages alongside JSON output
- Only supports single URL or `--tab` (not multiple URLs or `--all-tabs`)

## Common Scenarios

### AI Agent Documentation Fetching

```bash
# Fetch API documentation for AI context
snag https://api.example.com/docs > api-reference.md

# Pipe directly to AI assistant
snag --quiet https://docs.python.org/3/library/os.html | your-ai-tool
```

### Building a Knowledge Base

```bash
# Save multiple pages to a reference directory
snag -o reference/golang-basics.md https://go.dev/doc/tutorial/getting-started
snag -o reference/golang-concurrency.md https://go.dev/doc/effective_go#concurrency
snag -o reference/golang-errors.md https://go.dev/blog/error-handling-and-go
```

### Fetching Dynamic Content

```bash
# Wait for JavaScript to render content
snag --wait-for "#main-content" https://single-page-app.com

# Give slow sites more time
snag --timeout 90 --wait-for ".loaded" https://heavy-site.com
```

### Working with Authenticated Tabs

```bash
# Step 1: Open browser and log in to your sites
snag --open-browser

# (Manually log in to your private sites in the browser window)

# Step 2: List tabs to see what's available
snag --list-tabs

# Example output:
#   Available tabs in browser (4 tabs, sorted by URL):
#     [1] about:blank (New Tab)
#     [2] https://app.example.com/dashboard (Dashboard)
#     [3] https://github.com/myorg/private-repo (My Private Repo)
#     [4] https://internal.company.com/docs (Internal Documentation)

# Step 3: Fetch from authenticated tabs without re-logging in
snag -t 2 -o private-repo.md
snag -t "dashboard" -o dashboard.md
snag -t "internal" -o internal-docs.md

# All fetches reuse the existing authenticated session!
```

### Working with Multiple Open Tabs

```bash
# Collect documentation from tabs you already have open
snag -t "python" > python-docs.md
snag -t "golang" > golang-docs.md
snag -t "rust" > rust-docs.md

# Use patterns to match specific tabs
snag -t "github.com/.*" > github-content.md
snag -t ".*/dashboard" > dashboard.md

# Fetch by index if you know the tab position
for i in 1 2 3 4; do
  snag -t $i -o "tab-$i.md"
done

# Process all open tabs at once
snag --all-tabs --output-dir ~/my-tabs
snag -a -d ~/reference

# Combine --all-tabs with format options
snag --all-tabs --format pdf -d ~/pdfs
snag --all-tabs --format png -d ~/screenshots
```

### Batch Processing URLs

```bash
# Process multiple URLs inline
snag -d output/ https://example.com https://github.com https://go.dev

# Process URLs from a file
snag --url-file urls.txt -d output/

# Pipe URLs from stdin
cat urls.txt | snag --url-file - -d output/

# Pipe filtered URLs
grep "^https://docs" urls.txt | snag --url-file - -d ./documentation/

# Using heredoc
snag --url-file - -d pages/ <<EOF
# My URLs
example.com
github.com/p3bot/snag
go.dev
EOF

# Process URLs from a file (shell loop alternative)
while read url; do
  filename=$(echo "$url" | sed 's/[^a-zA-Z0-9]/_/g').md
  snag --quiet -o "$filename" "$url"
done < urls.txt

# Combine multiple pages
for url in https://example.com/page1 https://example.com/page2; do
  snag --quiet "$url" >> combined.md
  echo -e "\n---\n" >> combined.md
done
```

### CI/CD Integration

```bash
# Fetch documentation in CI pipeline
snag --force-headless --timeout 30 https://docs.example.com > docs.md

# Quiet mode for clean logs
snag --quiet --force-headless https://example.com > output.md
```

## Authentication

snag makes it easy to fetch content from authenticated/private sites using persistent browser sessions.

### Method 1: Visible Browser Mode

Open a browser, authenticate manually, then snag connects to it:

```bash
# Step 1: Open browser in visible mode and log in manually
# Note: Using the --open-browser (-b) switch enables the required DevTools protocol
snag --open-browser

# Step 2: In the browser window, navigate to your site and log in
# (Leave the browser open)

# Step 3: Fetch authenticated content - snag reuses your session
snag https://private.example.com

# Step 4: Fetch more pages with the same session
snag https://private.example.com/dashboard
snag https://private.example.com/settings
```

### Method 2: Force Visible Mode

Let snag launch the browser for you:

```bash
# Open browser and navigate to page for authentication
snag --open-browser https://private.example.com

# Authenticate in the browser window that opens
# Then leave it running

# Subsequent calls reuse the session
snag https://private.example.com/other-page
```

### Method 3: Existing Chromium Session

Keep one browser session for multiple snag calls:

```bash
# Terminal 1: Start Chromium with remote debugging
chromium --remote-debugging-port=9222 --user-data-dir=/tmp/chromium-profile

# Log in to your sites manually in this browser

# Terminal 2: Use snag with the existing session
snag https://authenticated-site1.com
snag https://authenticated-site2.com
snag https://authenticated-site3.com
```

All three commands share authentication state - no repeated logins required!

### Method 4: Using Your Default Chrome Profile

You can use your existing Chrome profile with all its saved logins and cookies:

**Option A: Daily workflow - Use snag as your Chrome launcher**

If you use snag regularly, you can make it your primary way to launch Chrome:

```bash
# Close your regular Chrome first, then launch via snag:
snag --open-browser --user-data-dir ~/.config/google-chrome                       # Linux
snag --open-browser --user-data-dir ~/.config/chromium                            # Linux Chromium
snag --open-browser --user-data-dir ~/Library/Application\ Support/Google/Chrome  # macOS

# Now browse normally AND use snag for tab fetching:
snag --list-tabs
snag -t 1                                    # Fetch from any tab
snag https://example.com                     # Open new tabs
```

This gives you your full Chrome experience (bookmarks, extensions, history, passwords) PLUS snag's tab management capabilities!

**Option B: One-off fetches with your profile**

```bash
# Must close Chrome first!
snag --user-data-dir ~/.config/google-chrome \
  https://private.example.com
```

**Important caveats:**

1. **Chrome must be closed** - You cannot run both Chrome and snag with the same profile simultaneously. Chrome locks profile directories to prevent corruption.

2. **Risk of corruption** - If something goes wrong, you could corrupt your primary profile data. Consider using a separate profile for automation.

3. **Profile structure** - Chrome's `--user-data-dir` points to the parent directory containing multiple profiles (Default, Profile 1, etc.). Chrome will use the Default profile unless you specify otherwise.

**Option C: Safer alternative - Use a dedicated profile for snag**

```bash
# Create and use a dedicated profile for snag
snag --user-data-dir ~/.config/google-chrome/snag-profile \
  --open-browser

# Authenticate once in the browser window
# Profile persists between runs - no need to re-authenticate!

# Subsequent fetches reuse the same profile
snag --user-data-dir ~/.config/google-chrome/snag-profile \
  https://private.example.com
```

The dedicated profile approach gives you persistence without risking your main Chrome profile.

## Advanced Usage

### Custom User Agent

Bypass headless detection or mimic specific browsers:

```bash
# Linux Firefox user agent
snag --user-agent "Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/115.0" \
  https://example.com

# Custom bot identifier
snag --user-agent "MyBot/1.0 (+https://example.com/bot)" \
  https://api-docs.example.com
```

### Debugging Failed Fetches

```bash
# See what's happening during fetch
snag --verbose https://problematic-site.com

# Full debug output including browser messages
snag --debug https://problematic-site.com 2> debug.log

# Open browser to see what snag sees
snag --open-browser https://problematic-site.com
```

### Working with Browser Tabs

snag can list and fetch content from existing browser tabs, making it easy to reuse authenticated sessions and reduce tab clutter.

**List all open tabs:**

```bash
# See what tabs are currently open
snag --list-tabs
snag -l

# Example output:
#   Available tabs in browser (3 tabs, sorted by URL):
#     [1] https://app.example.com/dashboard (Dashboard (authenticated))
#     [2] https://docs.python.org/3/ (3.13.1 Documentation)
#     [3] https://github.com/p3bot/snag (p3bot/snag: Intelligent web content fetcher)
```

**Fetch from specific tab by index:**

```bash
# Fetch from first tab
snag --tab 1
snag -t 1

# Fetch from third tab and save to file
snag -t 3 -o docs.md

# Get HTML instead of Markdown
snag -t 2 --format html

# Get as PDF or PNG
snag -t 3 --format pdf -o docs.pdf
snag -t 1 --format png -o screenshot.png
```

**Fetch from tab by URL pattern:**

```bash
# Exact URL match (case-insensitive)
snag -t "https://docs.python.org/3/"
snag -t "GITHUB.COM/p3bot/snag"

# Contains/substring match (processes ALL matching tabs if multiple)
snag -t "dashboard"      # Outputs to stdout if 1 match, auto-saves all if multiple
snag -t "python"         # Fetches all tabs containing "python"
snag -t "github" -d ./   # Saves all github tabs to current directory

# Regex pattern match (processes ALL matching tabs if multiple)
snag -t "https://.*\.com"        # All .com URLs
snag -t ".*/dashboard"           # All dashboard URLs
snag -t "(github|gitlab)\.com"   # All github or gitlab tabs
```

**Pattern matching behavior:**

- Tries in order: exact URL match → contains match → regex match
- **Single match**: Outputs to stdout (or to file with `-o`)
- **Multiple matches**: Auto-saves all tabs with generated filenames (use `-d` for custom directory)

**Why use tabs?**

- Reuse authenticated sessions without re-logging in
- Fetch from multiple pages without creating new tabs
- Quick access to content you already have open

**Tab closing behavior:**

```bash
# Close tab after fetching (default in headless mode)
snag --close-tab https://example.com

# Keep tab open (default in visible mode)
snag https://example.com
```

### Custom Remote Debugging Port

```bash
# Use different port if 9222 is busy
snag --port 9223 https://example.com

# Connect to Chromium running on custom port
chromium --remote-debugging-port=9223 &
snag --port 9223 https://example.com
```

## CLI Reference

### Core Arguments

```
<url>                      URL to fetch (required, unless using --list-tabs or --tab)
-v, --version              Display version information
-h, --help                 Show help message and exit
```

### Tab Operations

```
-l, --list-tabs            List all open tabs in the browser
-t, --tab <PATTERN>        Fetch from existing tab by index (1, 2, 3...) or URL pattern
                           Patterns can be:
                             - Index number: 1, 2, 3 (tab position)
                             - Exact URL: https://example.com (case-insensitive)
                             - Substring: dashboard, github, docs (contains match)
                             - Regex: https://.*\.com, .*/dashboard, (github|gitlab)\.com
-a, --all-tabs             Process all open browser tabs (saves with auto-generated filenames)
                           Requires --output-dir or saves to current directory
```

**Note:** Tabs are sorted alphabetically by URL (primary), then Title (secondary), then ID (tertiary) for predictable ordering. Chrome DevTools Protocol doesn't guarantee visual left-to-right tab order, so snag sorts tabs to ensure consistent, reproducible results. Tab [1] = first tab alphabetically by URL, not the first visual tab in your browser.

### Output Control

```
-o, --output <file>        Save output to file instead of stdout
-d, --output-dir <dir>     Save files with auto-generated names to directory
-f, --format <FORMAT>      Output format: md (default) | html | text | pdf | png
                           Format aliases: markdown→md, txt→text
                           Case-insensitive: MD, MARKDOWN, Html, PDF, etc.
-i, --info                 Output page metadata as JSON (title, URL, domain, slug, timestamp)
                           Mutually exclusive with --format (always outputs JSON)
                           Output is quiet by default (no log messages)
```

### Page Loading

```
--timeout <seconds>        Page load timeout in seconds (default: 30)
-w, --wait-for <selector>  Wait for CSS selector before extracting content
```

### Browser Control

```
-p, --port <port>          Chromium remote debugging port (default: 9222)
-c, --close-tab            Close the browser tab after fetching content
--force-headless           Force headless mode even if Chromium is running
-b, --open-browser         Open Chromium browser in visible state (no URL required)
-k, --kill-browser         Kill browser processes with remote debugging enabled
```

### Logging/Debugging

```
--verbose                  Enable verbose logging output
-q, --quiet                Suppress all output except errors and content
--debug                    Enable debug output with CDP messages
```

### Request Control

```
--user-agent <string>      Custom user agent string (bypass headless detection)
```

## Troubleshooting

### Browser Issues

**"Browser not found" error**

snag cannot locate Chrome/Chromium on your system.

Solutions:

- Install Chromium: `brew install chromium`
- Install Chrome from https://www.google.com/chrome/
- Ensure Chromium/Chrome is in your system PATH

**"Failed to connect to existing browser"**

Cannot connect to running browser instance.

Solutions:

- Ensure Chromium/Chrome is launched with `--remote-debugging-port=9222`
- Try different port: `snag --port 9223 https://example.com`
- Kill existing Chromium/Chrome processes and let snag launch a new instance

**"Stuck or lingering browser processes"**

Browser processes with remote debugging enabled remain after snag exits.

Solutions:

- Kill all debugging browsers: `snag --kill-browser` or `snag -k`
- Kill specific port only: `snag --kill-browser --port 9223`
- Note: Only kills browsers with `--remote-debugging-port` enabled (development browsers), never regular browsing sessions
- Safe for scripting: exits with code 0 even if no browsers found (idempotent)

### Diagnostic Information

Get comprehensive diagnostic information about your snag environment:

```bash
# Run diagnostics
snag --doctor

# Check specific port
snag --doctor --port 9223
```

This displays:

- snag and Go versions (with update check)
- Detected browser and version
- Browser connection status and tab counts
- Profile locations for all common browsers
- Environment variables
- Working directory

**Use this when:**

- Troubleshooting issues
- Reporting bugs (include doctor output)
- Checking if browser is running
- Finding profile paths
- Verifying snag installation

### Authentication Issues

**"Authentication required" error**

Page requires login but snag cannot authenticate.

Solutions:

- Open browser with `snag --open-browser`, log in, then run snag again
- Use `--list-tabs` to find authenticated tabs, then `--tab` to fetch from them
- Browser session persists authentication across snag calls

### Tab Issues

**"No Chrome instance running" when using --list-tabs or --tab**

Tab features require an existing browser with remote debugging enabled.

Solutions:

- Open browser first: `snag --open-browser`
- Or manually start Chrome/Chromium: `chromium --remote-debugging-port=9222`
- Then run `snag --list-tabs` to verify connection

**"Tab index out of range" or "No tab matches pattern"**

Cannot find the specified tab.

Solutions:

- Run `snag --list-tabs` to see available tabs and their indexes
- Tab indexes are 1-based (first tab is 1, not 0)
- **Tabs are sorted by URL**, not visual browser order - tab [1] is first alphabetically by URL
- For patterns, try simpler matches: `snag -t "example"` instead of complex regex
- Remember: pattern matching is case-insensitive

**Pattern not matching expected tab**

Your pattern matches a different tab than expected.

Solutions:

- Use `--list-tabs` to see exact URLs of open tabs
- Be more specific with your pattern: use full URL instead of substring
- Remember: multiple matching tabs will all be processed and auto-saved (not just first match)
- For single specific tab: use exact URL pattern or tab index: `snag -t 3`

### Timeout Issues

**"Page load timeout" error**

Page takes too long to load.

Solutions:

- Increase timeout: `snag --timeout 60 https://example.com`
- Use `--wait-for` for specific element: `snag --wait-for ".content" https://example.com`
- Check network connectivity
- Try `--verbose` to see what's happening

**Page loads but content is missing**

Dynamic content hasn't appeared yet.

Solutions:

- Use `--wait-for` with selector: `snag --wait-for "#main-content" https://example.com`
- Increase timeout to allow for slow loading
- Inspect page with `--format html` to see raw output

### Output Issues

**Output is empty or incomplete**

Fetched page but content is missing.

Solutions:

- Try `--format html` to see raw HTML
- Try `--format text` to see plain text extraction
- Use `--verbose` to check if page loaded correctly
- Page may require authentication (see authentication section)
- Content may be loaded dynamically (use `--wait-for`)

**Markdown formatting looks wrong**

Converted Markdown has formatting issues.

Solutions:

- Use `--format html` to get raw HTML instead
- Use `--format text` for plain text only (no formatting)
- Some complex HTML structures may not convert perfectly to Markdown
- Report specific issues at https://github.com/p3bot/snag/issues

### Platform-Specific Issues

**Linux: "No DISPLAY environment variable"**

Running in headless environment without display.

Solutions:

- Headless mode should work automatically
- Ensure Xvfb is installed: `sudo apt install xvfb`
- Use `--force-headless` explicitly

**macOS: "Chromium.app cannot be opened"**

macOS security blocking Chromium/Chrome launch.

Solutions:

- Open Chromium manually first: `open -a Chromium` or `open -a "Google Chrome"`
- Check System Preferences > Security & Privacy
- Allow the browser in privacy settings

**macOS: Browser processes remain after closing window**

On macOS, closing a Chrome/Chromium window doesn't quit the application - processes continue running in the background.

This is normal macOS behavior. To fully quit:

- Press **Cmd+Q** in the browser window
- Right-click Chrome icon in Dock → Quit
- Or: `pkill -f "Chrome.*remote-debugging-port"`

### Getting Help

Still having issues?

1. Run with `--debug` flag for detailed logs
2. Check existing issues: https://github.com/p3bot/snag/issues
3. Create new issue with:
   - snag version: `snag --version`
   - Operating system and version
   - Full command you ran
   - Complete error message
   - Output from `--debug` flag

## How It Works

### Smart Browser Management

1. **Session Detection**: Auto-detects existing Chromium-based browser instance with remote debugging enabled
2. **Mode Selection**:
   - If Chromium browser is running → Connect to existing session (preserves auth/cookies)
   - If no browser found → Launch headless mode
   - Use `--open-browser` to open visible browser for authentication
3. **Tab Management**:
   - List tabs with `--list-tabs` to see what's currently open
   - Fetch from specific tabs using `--tab` (by index or URL pattern)
   - Tabs stay open in visible mode, close in headless mode (or use `--close-tab`)
   - Reuse authenticated sessions without creating new tabs

### Output Routing

- **stdout**: Content only (HTML/Markdown) - enables piping to other tools
- **stderr**: All logs, warnings, errors, progress indicators

This design makes snag perfect for shell pipelines and AI agent integration.

### Technology

- **Language**: Go 1.25.3
- **Browser Control**: Chrome DevTools Protocol via [go-rod/rod](https://github.com/go-rod/rod)
- **HTML Conversion**: [html-to-markdown/v2](https://github.com/JohannesKaufmann/html-to-markdown)
- **CLI Framework**: [cobra](https://github.com/spf13/cobra)

## Contributing

Contributions welcome! Please:

1. Check existing issues: https://github.com/p3bot/snag/issues
2. Create issue for bugs or feature requests
3. Submit pull requests against `main` branch

### Reporting Issues

Include:

- snag version: `snag --version`
- Operating system and version
- Full command and error message
- Output from `--debug` flag

## License

`snag` is licensed under the [Mozilla Public License 2.0](LICENSE).

### Third-Party Licenses

This project uses the following open-source libraries:

- [go-rod/rod](https://github.com/go-rod/rod) - MIT License
- [cobra](https://github.com/spf13/cobra) - Apache 2.0 License
- [html-to-markdown](https://github.com/JohannesKaufmann/html-to-markdown) - MIT License

See the [LICENSES](LICENSES/) directory for full license texts.

## Author

Grant Carthew <grant@carthew.net>
