// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package validate

import (
	"os"
	"strings"
	"testing"

	"github.com/p3bot/snag/internal/format"
	"github.com/p3bot/snag/internal/logger"
)

func init() {
	logger.SetDefault(logger.Discard())
}

func TestValidateURL_Valid(t *testing.T) {
	tests := []string{
		"https://example.com",
		"http://example.com",
		"https://example.com/path",
		"https://example.com/path?query=value",
		"https://subdomain.example.com",
		"https://example.com:8080",
	}

	for _, url := range tests {
		_, err := URL(url)
		if err != nil {
			t.Errorf("expected valid URL %q to pass, got error: %v", url, err)
		}
	}
}

func TestValidateURL_Invalid(t *testing.T) {
	tests := []struct {
		url  string
		desc string
	}{
		{"ftp://example.com", "invalid scheme"},
		{"javascript:alert(1)", "javascript scheme"},
		{"://example.com", "malformed URL"},
	}

	for _, tt := range tests {
		_, err := URL(tt.url)
		if err == nil {
			t.Errorf("expected invalid URL %q (%s) to fail validation", tt.url, tt.desc)
		}
	}
}

func TestValidateURL_MissingScheme(t *testing.T) {
	// URL actually adds https:// if no scheme is present
	tests := []string{
		"example.com",
		"www.example.com",
		"example.com/path",
	}

	for _, url := range tests {
		normalized, err := URL(url)
		if err != nil {
			t.Errorf("expected URL without scheme %q to be normalized, got error: %v", url, err)
		}
		if !strings.HasPrefix(normalized, "https://") {
			t.Errorf("expected normalized URL to start with https://, got: %s", normalized)
		}
	}
}

func TestValidateFormat_Valid(t *testing.T) {
	// Test with normalized format values (as they would be after NormalizeFormat)
	validFormats := []string{
		format.Markdown, // "md"
		format.HTML,     // "html"
		format.Text,     // "text"
		format.PDF,      // "pdf"
		format.PNG,      // "png"
	}

	for _, format := range validFormats {
		err := Format(format)
		if err != nil {
			t.Errorf("expected valid format %q to pass validation, got error: %v", format, err)
		}
	}
}

func TestValidateFormat_Invalid(t *testing.T) {
	// Test with truly invalid formats (not supported by snag)
	// Note: Format expects already-normalized input
	invalidFormats := []string{
		"json",
		"xml",
		"yaml",
		"txt", // Should be normalized to "text" before validation
		"",
		"invalid",
		"markdown", // Should be normalized to "md" before validation
	}

	for _, format := range invalidFormats {
		err := Format(format)
		if err == nil {
			t.Errorf("expected invalid format %q to fail validation", format)
		}
	}
}

func TestNormalizeFormat(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Case normalization
		{"MD", "md"},
		{"HTML", "html"},
		{"Text", "text"},
		{"PDF", "pdf"},
		{"PNG", "png"},
		// Aliases
		{"markdown", "md"},
		{"Markdown", "md"},
		{"MARKDOWN", "md"},
		{"txt", "text"},
		{"TXT", "text"},
		{"Txt", "text"},
		// Already normalized
		{"md", "md"},
		{"html", "html"},
		{"text", "text"},
		{"pdf", "pdf"},
		{"png", "png"},
		// Invalid formats (returned as-is after lowercase conversion)
		{"json", "json"},
		{"xml", "xml"},
	}

	for _, tt := range tests {
		result := NormalizeFormat(tt.input)
		if result != tt.expected {
			t.Errorf("NormalizeFormat(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}

func TestValidateTimeout_Valid(t *testing.T) {
	validTimeouts := []int{1, 30, 60, 120, 3600}

	for _, timeout := range validTimeouts {
		err := Timeout(timeout)
		if err != nil {
			t.Errorf("expected valid timeout %d to pass validation, got error: %v", timeout, err)
		}
	}
}

func TestValidateTimeout_Invalid(t *testing.T) {
	invalidTimeouts := []int{-1, 0, -100}

	for _, timeout := range invalidTimeouts {
		err := Timeout(timeout)
		if err == nil {
			t.Errorf("expected invalid timeout %d to fail validation", timeout)
		}
	}
}

func TestValidatePort_Valid(t *testing.T) {
	validPorts := []int{1024, 8080, 9222, 65535}

	for _, port := range validPorts {
		err := Port(port)
		if err != nil {
			t.Errorf("expected valid port %d to pass validation, got error: %v", port, err)
		}
	}
}

func TestValidatePort_Invalid(t *testing.T) {
	invalidPorts := []int{-1, 0, 1, 80, 443, 1023, -100, 65536, 99999}

	for _, port := range invalidPorts {
		err := Port(port)
		if err == nil {
			t.Errorf("expected invalid port %d to fail validation", port)
		}
	}
}

func TestValidateOutputPath_Valid(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Test with valid writable path
	validPath := tmpDir + "/output.md"
	err := OutputPath(validPath)
	if err != nil {
		t.Errorf("expected valid writable path %q to pass validation, got error: %v", validPath, err)
	}
}

func TestValidateOutputPath_NonexistentDirectory(t *testing.T) {
	// Test with path to non-existent directory
	invalidPath := "/nonexistent/directory/output.md"
	err := OutputPath(invalidPath)
	if err == nil {
		t.Errorf("expected path with non-existent directory %q to fail validation", invalidPath)
	}
}

func TestValidateOutputPath_ReadOnlyDirectory(t *testing.T) {
	// Create a temporary directory and make it read-only
	tmpDir := t.TempDir()
	readOnlyDir := tmpDir + "/readonly"

	err := os.Mkdir(readOnlyDir, 0755)
	if err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	// Make directory read-only (no write permission)
	err = os.Chmod(readOnlyDir, 0555)
	if err != nil {
		t.Fatalf("failed to make directory read-only: %v", err)
	}

	// Ensure cleanup restores permissions so TempDir can clean up
	t.Cleanup(func() {
		os.Chmod(readOnlyDir, 0755)
	})

	// Test with path to read-only directory
	invalidPath := readOnlyDir + "/output.md"
	err = OutputPath(invalidPath)
	if err == nil {
		t.Errorf("expected path to read-only directory %q to fail validation", invalidPath)
	}
}

// Phase 3 validator tests

func TestValidateDirectory_Valid(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Should pass validation for existing writable directory
	err := Directory(tmpDir)
	if err != nil {
		t.Errorf("expected valid directory %q to pass validation, got error: %v", tmpDir, err)
	}
}

func TestValidateDirectory_NonExistent(t *testing.T) {
	// Test with non-existent directory
	invalidDir := "/nonexistent/test/directory"
	err := Directory(invalidDir)
	if err == nil {
		t.Errorf("expected non-existent directory %q to fail validation", invalidDir)
	}
}

func TestValidateDirectory_NotADirectory(t *testing.T) {
	// Create a temporary file (not a directory)
	tmpFile, err := os.CreateTemp("", "snag-test-*")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpFilePath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpFilePath)

	// Should fail because it's a file, not a directory
	err = Directory(tmpFilePath)
	if err == nil {
		t.Errorf("expected file path %q to fail directory validation", tmpFilePath)
	}
}

func TestValidateDirectory_ReadOnly(t *testing.T) {
	// Create a temporary directory and make it read-only
	tmpDir := t.TempDir()
	readOnlyDir := tmpDir + "/readonly"

	err := os.Mkdir(readOnlyDir, 0755)
	if err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	// Make directory read-only (no write permission)
	err = os.Chmod(readOnlyDir, 0555)
	if err != nil {
		t.Fatalf("failed to make directory read-only: %v", err)
	}

	// Ensure cleanup restores permissions so TempDir can clean up
	t.Cleanup(func() {
		os.Chmod(readOnlyDir, 0755)
	})

	// Should fail because directory is not writable
	err = Directory(readOnlyDir)
	if err == nil {
		t.Errorf("expected read-only directory %q to fail validation", readOnlyDir)
	}
}

func TestIsNonFetchableURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		// Non-fetchable URLs (browser-internal)
		{"chrome newtab", "chrome://newtab", true},
		{"chrome settings", "chrome://settings", true},
		{"chrome uppercase", "CHROME://FLAGS", true},
		{"about blank", "about:blank", true},
		{"about preferences", "about:preferences", true},
		{"devtools", "devtools://devtools/bundled/inspector.html", true},
		{"chrome extension", "chrome-extension://abcdefg", true},
		{"edge internal", "edge://settings", true},
		{"brave internal", "brave://settings", true},

		// Fetchable URLs
		{"http example", "http://example.com", false},
		{"https example", "https://example.com", false},
		{"file url", "file:///path/to/file.html", false},
		{"domain only", "example.com", false},
		{"subdomain", "https://sub.example.com", false},
		{"path and query", "https://example.com/path?query=value", false},
		{"localhost", "http://localhost:8080", false},
		{"ip address", "http://192.168.1.1", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsNonFetchableURL(tt.url)
			if result != tt.expected {
				t.Errorf("IsNonFetchableURL(%q) = %v, expected %v", tt.url, result, tt.expected)
			}
		})
	}
}

func TestCheckExtensionMismatch(t *testing.T) {
	tests := []struct {
		name       string
		outputFile string
		format     string
		expected   bool // true = mismatch detected
	}{
		// Matching extensions
		{"md matches", "output.md", "md", false},
		{"html matches", "output.html", "html", false},
		{"txt matches", "output.txt", "text", false},
		{"pdf matches", "output.pdf", "pdf", false},
		{"png matches", "output.png", "png", false},

		// Case insensitivity
		{"uppercase ext", "output.MD", "md", false},
		{"mixed case", "output.Html", "html", false},

		// Mismatches
		{"md vs html", "output.md", "html", true},
		{"html vs text", "output.html", "text", true},
		{"md vs pdf", "output.md", "pdf", true},
		{"txt vs png", "output.txt", "png", true},

		// No extension
		{"no extension md", "output", "md", true},
		{"no extension html", "output", "html", true},

		// Empty output file (should not mismatch)
		{"empty output", "", "md", false},

		// Path with extension
		{"path with ext", "/path/to/output.md", "md", false},
		{"path mismatch", "/path/to/output.html", "md", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CheckExtensionMismatch(tt.outputFile, tt.format)
			if result != tt.expected {
				t.Errorf("CheckExtensionMismatch(%q, %q) = %v, expected %v",
					tt.outputFile, tt.format, result, tt.expected)
			}
		})
	}
}

func TestValidateWaitFor(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"valid selector", ".content", ".content"},
		{"valid id selector", "#main", "#main"},
		{"valid complex selector", "div.content > p", "div.content > p"},
		{"with whitespace", "  .content  ", ".content"},
		{"with tabs", "\t.content\t", ".content"},
		{"empty string", "", ""},
		{"only whitespace", "   ", ""},
		{"attribute selector", "[data-test='value']", "[data-test='value']"},
		{"pseudo selector", "div:first-child", "div:first-child"},
		{"multiple classes", ".class1.class2", ".class1.class2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WaitFor(tt.input, true)
			if result != tt.expected {
				t.Errorf("WaitFor(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestValidateUserAgent(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"valid chrome ua", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/120.0.0.0", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/120.0.0.0"},
		{"valid firefox ua", "Mozilla/5.0 (X11; Linux x86_64; rv:120.0) Gecko/20100101 Firefox/120.0", "Mozilla/5.0 (X11; Linux x86_64; rv:120.0) Gecko/20100101 Firefox/120.0"},
		{"simple ua", "MyBot/1.0", "MyBot/1.0"},

		// Whitespace handling
		{"with whitespace", "  MyBot/1.0  ", "MyBot/1.0"},
		{"with tabs", "\tMyBot/1.0\t", "MyBot/1.0"},
		{"empty string", "", ""},
		{"only whitespace", "   ", ""},

		// Security: newline injection prevention
		{"with newline", "MyBot/1.0\nInjected-Header: value", "MyBot/1.0 Injected-Header: value"},
		{"with carriage return", "MyBot/1.0\rInjected-Header: value", "MyBot/1.0 Injected-Header: value"},
		{"with crlf", "MyBot/1.0\r\nInjected-Header: value", "MyBot/1.0  Injected-Header: value"},
		{"multiple newlines", "Line1\nLine2\nLine3", "Line1 Line2 Line3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := UserAgent(tt.input, true)
			if result != tt.expected {
				t.Errorf("UserAgent(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestValidateUserAgent_SecuritySanitization(t *testing.T) {
	// Focused security tests for HTTP header injection prevention
	maliciousInputs := []struct {
		name     string
		input    string
		hasSpace bool // Result should have spaces instead of newlines
	}{
		{"crlf injection", "MyBot/1.0\r\nX-Injected: malicious", true},
		{"lf injection", "MyBot/1.0\nX-Injected: malicious", true},
		{"cr injection", "MyBot/1.0\rX-Injected: malicious", true},
		{"multiple crlf", "A\r\nB\r\nC", true},
	}

	for _, tt := range maliciousInputs {
		t.Run(tt.name, func(t *testing.T) {
			result := UserAgent(tt.input, true)
			// Result should not contain \n or \r
			if strings.Contains(result, "\n") || strings.Contains(result, "\r") {
				t.Errorf("UserAgent(%q) still contains newline/CR characters: %q", tt.input, result)
			}
			// If input had newlines, result should have spaces
			if tt.hasSpace && !strings.Contains(result, " ") {
				t.Errorf("UserAgent(%q) should contain spaces after sanitization: %q", tt.input, result)
			}
		})
	}
}

func TestValidateURL_IDNHomograph(t *testing.T) {
	// IDN (Internationalized Domain Names) homograph attack tests
	tests := []struct {
		name        string
		url         string
		shouldAllow bool
	}{
		{
			name:        "normal domain",
			url:         "https://example.com",
			shouldAllow: true,
		},
		{
			name:        "punycode domain",
			url:         "https://xn--e1afmkfd.xn--p1ai", // пример.рф
			shouldAllow: true,
		},
		{
			name:        "mixed script (potential homograph)",
			url:         "https://раypal.com", // Note: contains Cyrillic 'а'
			shouldAllow: true,                 // URL doesn't block these
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := URL(tt.url)
			hasError := err != nil

			if tt.shouldAllow && hasError {
				t.Errorf("should allow URL %q, got error: %v", tt.url, err)
			}
			if !tt.shouldAllow && !hasError {
				t.Errorf("should reject URL %q", tt.url)
			}
		})
	}
}

func TestValidateURL_ExtremelyLong(t *testing.T) {
	tests := []struct {
		name      string
		urlLength int
		shouldErr bool
	}{
		{"normal length", 100, false},
		{"2000 chars", 2000, false},
		{"10000 chars", 10000, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create URL of specified length
			longPath := strings.Repeat("a", tt.urlLength-20)
			url := "https://example.com/" + longPath

			_, err := URL(url)

			if tt.shouldErr && err == nil {
				t.Errorf("expected error for %d char URL", tt.urlLength)
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("unexpected error for %d char URL: %v", tt.urlLength, err)
			}
		})
	}
}

func TestValidateUserAgent_ExtremeLength(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{"normal length", 100},
		{"very long", 1000},
		{"extremely long", 10000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			longUA := strings.Repeat("A", tt.length)
			result := UserAgent(longUA, true)

			// Should not panic, should return trimmed string
			if len(result) == 0 {
				t.Error("should return non-empty result")
			}
		})
	}
}

func TestValidateWaitFor_Injection(t *testing.T) {
	// Test that CSS selectors don't allow script injection
	tests := []struct {
		name     string
		selector string
		expected string
	}{
		{
			name:     "normal selector",
			selector: ".content",
			expected: ".content",
		},
		{
			name:     "selector with quotes",
			selector: "div[data-test='value']",
			expected: "div[data-test='value']",
		},
		// Note: WaitFor just trims, doesn't sanitize
		// These tests document current behavior
		{
			name:     "selector with angle brackets",
			selector: "<script>alert()</script>",
			expected: "<script>alert()</script>", // Not sanitized
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WaitFor(tt.selector, true)
			if result != tt.expected {
				t.Errorf("WaitFor() = %q, expected %q", result, tt.expected)
			}
		})
	}
}
