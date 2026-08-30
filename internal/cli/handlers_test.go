// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/p3bot/snag/internal/browser"
	"github.com/p3bot/snag/internal/fetch"
	"github.com/p3bot/snag/internal/logger"
)

func TestConfig_BrowserOptions(t *testing.T) {
	c := &Config{
		URL:           "https://example.com",
		Port:          9223,
		ForceHeadless: true,
		OpenBrowser:   true,
		UserAgent:     "Bot/1.0",
		UserDataDir:   "/tmp/snag-profile",
	}

	opts := c.BrowserOptions()
	if opts.Port != 9223 {
		t.Errorf("Port = %d, want 9223", opts.Port)
	}
	if !opts.ForceHeadless {
		t.Error("ForceHeadless = false, want true")
	}
	if !opts.OpenBrowser {
		t.Error("OpenBrowser = false, want true")
	}
	if opts.UserAgent != "Bot/1.0" {
		t.Errorf("UserAgent = %q, want Bot/1.0", opts.UserAgent)
	}
	if opts.UserDataDir != "/tmp/snag-profile" {
		t.Errorf("UserDataDir = %q, want /tmp/snag-profile", opts.UserDataDir)
	}
}

func TestStripURLParams(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "URL with query params",
			input:    "https://example.com/path?query=123",
			expected: "https://example.com/path",
		},
		{
			name:     "URL with hash fragment",
			input:    "https://example.com/path#section",
			expected: "https://example.com/path",
		},
		{
			name:     "URL with both query params and hash",
			input:    "https://example.com/path?query=123#section",
			expected: "https://example.com/path",
		},
		{
			name:     "URL without params or hash",
			input:    "https://example.com/path",
			expected: "https://example.com/path",
		},
		{
			name:     "URL without path",
			input:    "https://example.com",
			expected: "https://example.com",
		},
		{
			name:     "Chrome internal URL",
			input:    "chrome://newtab/",
			expected: "chrome://newtab/",
		},
		{
			name:     "Complex query params",
			input:    "https://www.ato.gov.au/about-ato/contact-us?gclsrc=aw.ds&gad_source=1&gclid=EAIaIQobChMI",
			expected: "https://www.ato.gov.au/about-ato/contact-us",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripURLParams(tt.input)
			if result != tt.expected {
				t.Errorf("stripURLParams(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFormatTabLine(t *testing.T) {
	tests := []struct {
		name      string
		index     int
		title     string
		url       string
		maxLength int
		verbose   bool
		expected  string
	}{
		{
			name:      "Normal mode - short title and URL",
			index:     1,
			title:     "Example Domain",
			url:       "https://example.com",
			maxLength: 120,
			verbose:   false,
			expected:  "  [1] https://example.com (Example Domain)",
		},
		{
			name:      "Normal mode - empty title",
			index:     2,
			title:     "",
			url:       "https://example.com",
			maxLength: 120,
			verbose:   false,
			expected:  "  [2] https://example.com",
		},
		{
			name:      "Normal mode - strips query params",
			index:     3,
			title:     "Contact Us",
			url:       "https://www.ato.gov.au/about-ato/contact-us?gclsrc=aw.ds&gad_source=1",
			maxLength: 120,
			verbose:   false,
			expected:  "  [3] https://www.ato.gov.au/about-ato/contact-us (Contact Us)",
		},
		{
			name:      "Normal mode - very long title gets truncated",
			index:     4,
			title:     "This is a very long title that should be truncated to fit within the 120 character limit when combined with the URL",
			url:       "https://example.com",
			maxLength: 120,
			verbose:   false,
			expected:  "  [4] https://example.com (This is a very long title that should be truncated to fit within the 120 character limit ...)",
		},
		{
			name:      "Normal mode - very long URL gets truncated",
			index:     5,
			title:     "Short",
			url:       "https://very-long-domain-name-that-exceeds-the-maximum-url-length-limit.com/path/to/resource",
			maxLength: 120,
			verbose:   false,
			expected:  "  [5] https://very-long-domain-name-that-exceeds-the-maximum-url-length-limit.com/p... (Short)",
		},
		{
			name:      "Verbose mode - shows full URL with query params",
			index:     6,
			title:     "YouTube Search",
			url:       "https://www.google.com/search?q=youtube&oq=tyou&gs_lcrp=EgZjaHJvbWU",
			maxLength: 120,
			verbose:   true,
			expected:  "  [6] https://www.google.com/search?q=youtube&oq=tyou&gs_lcrp=EgZjaHJvbWU - YouTube Search",
		},
		{
			name:      "Verbose mode - empty title",
			index:     7,
			title:     "",
			url:       "https://example.com",
			maxLength: 120,
			verbose:   true,
			expected:  "  [7] https://example.com",
		},
		{
			name:      "Chrome internal URL",
			index:     8,
			title:     "New Tab",
			url:       "chrome://newtab/",
			maxLength: 120,
			verbose:   false,
			expected:  "  [8] chrome://newtab/ (New Tab)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatTabLine(tt.index, tt.title, tt.url, tt.maxLength, tt.verbose)
			if result != tt.expected {
				t.Errorf("formatTabLine() =\n%q\nexpected:\n%q", result, tt.expected)
			}
			// In normal mode, verify line doesn't exceed maxLength
			if !tt.verbose && len(result) > tt.maxLength {
				t.Errorf("formatTabLine() result length %d exceeds maxLength %d", len(result), tt.maxLength)
			}
		})
	}
}

func TestFormatTabLine_Length(t *testing.T) {
	// Test that normal mode always respects the 120 character limit
	testCases := []struct {
		title string
		url   string
	}{
		{
			title: "Very long title " + strings.Repeat("a", 200),
			url:   "https://example.com",
		},
		{
			title: "Normal title",
			url:   "https://" + strings.Repeat("a", 200) + ".com",
		},
		{
			title: strings.Repeat("a", 100),
			url:   "https://" + strings.Repeat("b", 100) + ".com",
		},
	}

	for i, tc := range testCases {
		result := formatTabLine(i+1, tc.title, tc.url, 120, false)
		if len(result) > 120 {
			t.Errorf("Line %d length %d exceeds 120 chars: %q", i+1, len(result), result)
		}
	}
}

func TestDisplayTabList(t *testing.T) {
	// Create a simple string buffer to capture output
	var buf strings.Builder

	tabs := []browser.TabInfo{
		{Index: 1, URL: "https://example.com", Title: "Example Domain"},
		{Index: 2, URL: "https://github.com/user/repo?tab=readme", Title: "GitHub Repo"},
		{Index: 3, URL: "chrome://newtab/", Title: "New Tab"},
	}

	// Test normal mode (non-verbose)
	buf.Reset()
	displayTabList(tabs, &buf, false)
	output := buf.String()

	// Verify header
	if !strings.Contains(output, "Available tabs in browser (3 tabs, sorted by URL)") {
		t.Errorf("Missing or incorrect header in output")
	}

	// Verify clean URLs (no query params)
	if strings.Contains(output, "?tab=readme") {
		t.Errorf("Query params should be stripped in normal mode")
	}

	// Verify format: URL (Title)
	if !strings.Contains(output, "https://example.com (Example Domain)") {
		t.Errorf("Expected new format with URL (Title)")
	}

	// Test verbose mode
	buf.Reset()
	displayTabList(tabs, &buf, true)
	verboseOutput := buf.String()

	// Verify full URL with query params in verbose mode
	if !strings.Contains(verboseOutput, "?tab=readme") {
		t.Errorf("Query params should be shown in verbose mode")
	}

	// Test empty tabs
	buf.Reset()
	displayTabList([]browser.TabInfo{}, &buf, false)
	emptyOutput := buf.String()
	if !strings.Contains(emptyOutput, "No tabs open in browser") {
		t.Errorf("Expected 'No tabs' message for empty tab list")
	}
}

func TestDisplayTabList_LargeLists(t *testing.T) {
	// Create 100 tabs
	tabs := make([]browser.TabInfo, 100)
	for i := 0; i < 100; i++ {
		tabs[i] = browser.TabInfo{
			Index: i + 1,
			URL:   fmt.Sprintf("https://example.com/page%d", i),
			Title: fmt.Sprintf("Page %d", i),
		}
	}

	var buf strings.Builder
	displayTabList(tabs, &buf, false)
	output := buf.String()

	// Verify header shows correct count
	if !strings.Contains(output, "100 tabs") {
		t.Error("should show correct tab count")
	}

	// Verify first and last tab are included
	if !strings.Contains(output, "[1]") {
		t.Error("should show first tab")
	}
	if !strings.Contains(output, "[100]") {
		t.Error("should show last tab")
	}
}

func TestFormatTabLine_Unicode(t *testing.T) {
	tests := []struct {
		name  string
		title string
		url   string
	}{
		{
			name:  "emoji in title",
			title: "🚀 Rocket Launch 🌟",
			url:   "https://example.com",
		},
		{
			name:  "chinese characters",
			title: "中文标题",
			url:   "https://example.cn/页面",
		},
		{
			name:  "arabic text",
			title: "عنوان عربي",
			url:   "https://example.sa",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatTabLine(1, tt.title, tt.url, 120, false)

			// Should not crash with Unicode
			if len(result) == 0 {
				t.Error("should produce output for Unicode input")
			}

			// Should contain title (at least partially)
			// Note: May be truncated, but should have some content
		})
	}
}

func TestStripURLParams_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "encoded characters",
			input:    "https://example.com/path?q=%20space%20",
			expected: "https://example.com/path",
		},
		{
			name:     "multiple question marks",
			input:    "https://example.com/path?query=value?extra",
			expected: "https://example.com/path",
		},
		{
			name:     "multiple hashes",
			input:    "https://example.com/path#section#subsection",
			expected: "https://example.com/path",
		},
		{
			name:     "very long query string",
			input:    "https://example.com/path?" + strings.Repeat("a=b&", 100),
			expected: "https://example.com/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripURLParams(tt.input)
			if result != tt.expected {
				t.Errorf("stripURLParams() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestLoadURLsFromReader(t *testing.T) {
	// Setup logger for tests
	logger.SetDefault(logger.Discard())

	tests := []struct {
		name        string
		input       string
		source      string
		expected    []string
		expectError bool
	}{
		{
			name:   "single URL",
			input:  "https://example.com\n",
			source: "stdin",
			expected: []string{
				"https://example.com",
			},
			expectError: false,
		},
		{
			name: "multiple URLs",
			input: `https://example.com
https://github.com
https://golang.org
`,
			source: "stdin",
			expected: []string{
				"https://example.com",
				"https://github.com",
				"https://golang.org",
			},
			expectError: false,
		},
		{
			name: "URLs with comments",
			input: `# This is a comment
https://example.com
// Another comment
https://github.com # inline comment
https://golang.org // inline comment
`,
			source: "stdin",
			expected: []string{
				"https://example.com",
				"https://github.com",
				"https://golang.org",
			},
			expectError: false,
		},
		{
			name: "URLs with blank lines",
			input: `
https://example.com

https://github.com


https://golang.org

`,
			source: "stdin",
			expected: []string{
				"https://example.com",
				"https://github.com",
				"https://golang.org",
			},
			expectError: false,
		},
		{
			name: "auto-prepend https",
			input: `example.com
github.com/user/repo
`,
			source: "stdin",
			expected: []string{
				"https://example.com",
				"https://github.com/user/repo",
			},
			expectError: false,
		},
		{
			name: "mixed schemes",
			input: `https://example.com
http://insecure.com
example.org
`,
			source: "stdin",
			expected: []string{
				"https://example.com",
				"http://insecure.com",
				"https://example.org",
			},
			expectError: false,
		},
		{
			name:        "empty input",
			input:       "",
			source:      "stdin",
			expected:    nil,
			expectError: true, // ErrNoValidURLs
		},
		{
			name: "only comments",
			input: `# Comment 1
// Comment 2
# Comment 3
`,
			source:      "stdin",
			expected:    nil,
			expectError: true, // ErrNoValidURLs
		},
		{
			name: "only blank lines",
			input: `


`,
			source:      "stdin",
			expected:    nil,
			expectError: true, // ErrNoValidURLs
		},
		{
			name: "invalid URLs skipped",
			input: `https://example.com
not a valid url with spaces
https://github.com
`,
			source: "stdin",
			expected: []string{
				"https://example.com",
				"https://github.com",
			},
			expectError: false,
		},
		{
			name: "complex real-world example",
			input: `# My favorite sites
https://example.com
github.com/user/repo # My project

// Work URLs
https://internal.company.com
docs.company.com/api  // API docs

# End of list
`,
			source: "stdin",
			expected: []string{
				"https://example.com",
				"https://github.com/user/repo",
				"https://internal.company.com",
				"https://docs.company.com/api",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			urls, err := loadURLsFromReader(context.Background(), reader, tt.source)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				if len(urls) != len(tt.expected) {
					t.Errorf("got %d URLs, expected %d", len(urls), len(tt.expected))
				}

				for i, url := range urls {
					if i >= len(tt.expected) {
						break
					}
					if url != tt.expected[i] {
						t.Errorf("URL[%d] = %q, expected %q", i, url, tt.expected[i])
					}
				}
			}
		})
	}
}

func TestLoadURLsFromReader_Canceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := loadURLsFromReader(ctx, strings.NewReader("https://example.com\n"), "stdin")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("already-cancelled ctx: %v", err)
	}

	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()
	r, w := io.Pipe()
	t.Cleanup(func() {
		_ = w.Close()
		_ = r.Close()
	})

	done := make(chan error, 1)
	go func() {
		_, err := loadURLsFromReader(ctx, r, "stdin")
		done <- err
	}()

	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("blocked read cancelled ctx: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("loadURLsFromReader did not return after cancel")
	}
}

func TestCloseExistingTab_Nil(t *testing.T) {
	closeExistingTab(nil, nil)
}

func TestOutputPageInfo_SetsSlug(t *testing.T) {
	logger.SetDefault(logger.Discard())

	path := filepath.Join(t.TempDir(), "info.json")
	info := &fetch.PageInfo{
		Title:  "Hello & World",
		URL:    "https://example.com",
		Domain: "example.com",
	}

	if err := outputPageInfo(info, path); err != nil {
		t.Fatalf("outputPageInfo: %v", err)
	}

	if info.Slug != "hello-world" {
		t.Errorf("Slug = %q, want hello-world", info.Slug)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}

	var parsed fetch.PageInfo
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if parsed.Slug != "hello-world" {
		t.Errorf("JSON slug = %q, want hello-world", parsed.Slug)
	}
}
