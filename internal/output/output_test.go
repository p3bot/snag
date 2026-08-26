// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package output

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/p3bot/snag/internal/format"
	"github.com/p3bot/snag/internal/logger"
)

func init() {
	logger.SetDefault(logger.Discard())
}

// TestSlugifyTitle tests URL-safe slug generation from page titles
func TestSlugifyTitle(t *testing.T) {
	tests := []struct {
		title    string
		maxLen   int
		expected string
		desc     string
	}{
		// Basic slugification
		{"Example Domain", 80, "example-domain", "simple title"},
		{"GitHub - Project Page", 80, "github-project-page", "title with dash"},
		{"Docs   -   The Go Language", 80, "docs-the-go-language", "multiple spaces"},

		// Special characters
		{"!!!Test???", 80, "test", "only special chars"},
		{"Hello@World#2024", 80, "hello-world-2024", "mixed special chars"},
		{"UTF-8 Encoding: 你好", 80, "utf-8-encoding", "unicode characters"},

		// Edge cases
		{"", 80, "", "empty title"},
		{"   ", 80, "", "only whitespace"},
		{"---", 80, "", "only hyphens"},
		{"Test", 80, "test", "single word"},

		// Truncation
		{"This is a very long title that exceeds the maximum length allowed for slugs", 20, "this-is-a-very-long", "truncation at maxLen"},
		{"abcdefghijklmnopqrstuvwxyz", 10, "abcdefghij", "truncation exact"},
		{"hello-world-how-are-you", 15, "hello-world-how", "truncation with hyphen"},

		// Case conversion
		{"UPPERCASE TITLE", 80, "uppercase-title", "all uppercase"},
		{"MixedCase Title", 80, "mixedcase-title", "mixed case"},

		// Multiple consecutive hyphens
		{"Test  -  Multiple  -  Hyphens", 80, "test-multiple-hyphens", "collapse hyphens"},
		{"a----b", 80, "a-b", "many consecutive hyphens"},
	}

	for _, tt := range tests {
		result := SlugifyTitle(tt.title, tt.maxLen)
		if result != tt.expected {
			t.Errorf("SlugifyTitle(%q, %d) [%s] = %q, expected %q",
				tt.title, tt.maxLen, tt.desc, result, tt.expected)
		}
	}
}

// TestGenerateURLSlug tests fallback slug generation from URLs
func TestGenerateURLSlug(t *testing.T) {
	tests := []struct {
		url      string
		expected string
		desc     string
	}{
		// Valid URLs
		{"https://example.com", "example-com", "simple domain"},
		{"https://www.github.com/user/repo", "www-github-com", "subdomain"},
		{"https://api.example.com:8080", "api-example-com-8080", "with port"},
		{"http://localhost:3000", "localhost-3000", "localhost"},
		{"https://[::1]", "1", "IPv6 loopback"},
		{"https://[::1]:8080", "1-8080", "IPv6 loopback with port"},
		{"https://[2001:db8::1]/path", "2001-db8-1", "IPv6 documentation address"},
		{"https://[2001:db8::1]:443/path", "2001-db8-1-443", "IPv6 with port"},

		// Edge cases
		{"file:///path/to/file.html", "page", "file URL"},
		{"invalid-url", "page", "invalid URL"},
		{"", "page", "empty URL"},

		// Complex hostnames
		{"https://my-service.cloud-provider.example.com", "my-service-cloud-provider-example-com", "multi-level subdomain"},
	}

	for _, tt := range tests {
		result := GenerateURLSlug(tt.url)
		if result != tt.expected {
			t.Errorf("GenerateURLSlug(%q) [%s] = %q, expected %q",
				tt.url, tt.desc, result, tt.expected)
		}
	}
}

// TestGetFileExtension tests format to file extension mapping
func TestGetFileExtension(t *testing.T) {
	tests := []struct {
		format   string
		expected string
	}{
		{format.Markdown, ".md"},
		{format.HTML, ".html"},
		{format.Text, ".txt"},
		{format.PDF, ".pdf"},
		{format.PNG, ".png"},
		{"unknown", ".md"}, // Default fallback
		{"", ".md"},        // Empty fallback
	}

	for _, tt := range tests {
		result := GetFileExtension(tt.format)
		if result != tt.expected {
			t.Errorf("GetFileExtension(%q) = %q, expected %q",
				tt.format, result, tt.expected)
		}
	}
}

// TestGenerateFilename tests complete filename generation
func TestGenerateFilename(t *testing.T) {
	// Use fixed timestamp for predictable output
	timestamp := time.Date(2025, 10, 21, 14, 30, 45, 0, time.UTC)

	tests := []struct {
		title    string
		format   string
		url      string
		expected string
		desc     string
	}{
		// Normal cases
		{
			"Example Domain",
			format.Markdown,
			"https://example.com",
			"2025-10-21-143045-example-domain.md",
			"basic markdown",
		},
		{
			"GitHub Repository",
			format.HTML,
			"https://github.com",
			"2025-10-21-143045-github-repository.html",
			"html format",
		},
		{
			"Documentation Page",
			format.Text,
			"https://docs.example.com",
			"2025-10-21-143045-documentation-page.txt",
			"text format",
		},
		{
			"PDF Report",
			format.PDF,
			"https://example.com",
			"2025-10-21-143045-pdf-report.pdf",
			"pdf format",
		},
		{
			"Screenshot",
			format.PNG,
			"https://example.com",
			"2025-10-21-143045-screenshot.png",
			"png format",
		},

		// Empty title fallback to URL
		{
			"",
			format.Markdown,
			"https://example.com",
			"2025-10-21-143045-example-com.md",
			"empty title uses URL",
		},
		{
			"   ",
			format.Markdown,
			"https://www.github.com",
			"2025-10-21-143045-www-github-com.md",
			"whitespace title uses URL",
		},

		// Special characters in title
		{
			"Hello@World#2024!!!",
			format.Markdown,
			"https://example.com",
			"2025-10-21-143045-hello-world-2024.md",
			"special chars removed",
		},
	}

	for _, tt := range tests {
		result := GenerateFilename(tt.title, tt.format, timestamp, tt.url)
		if result != tt.expected {
			t.Errorf("GenerateFilename(%q, %q, ..., %q) [%s] = %q, expected %q",
				tt.title, tt.format, tt.url, tt.desc, result, tt.expected)
		}
	}
}

// TestResolveConflict tests filename conflict resolution
func TestResolveConflict(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	// Test 1: No conflict - file doesn't exist
	filename, err := ResolveConflict(tmpDir, "test.md")
	if err != nil {
		t.Fatalf("ResolveConflict failed: %v", err)
	}
	if filename != "test.md" {
		t.Errorf("expected 'test.md', got %q", filename)
	}

	// Test 2: Create file, should get conflict
	testFile := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	filename, err = ResolveConflict(tmpDir, "test.md")
	if err != nil {
		t.Fatalf("ResolveConflict failed: %v", err)
	}
	if filename != "test-1.md" {
		t.Errorf("expected 'test-1.md', got %q", filename)
	}

	// Test 3: Create test-1.md, should get test-2.md
	testFile1 := filepath.Join(tmpDir, "test-1.md")
	if err := os.WriteFile(testFile1, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test-1 file: %v", err)
	}

	filename, err = ResolveConflict(tmpDir, "test.md")
	if err != nil {
		t.Fatalf("ResolveConflict failed: %v", err)
	}
	if filename != "test-2.md" {
		t.Errorf("expected 'test-2.md', got %q", filename)
	}

	// Test 4: Different extension
	filename, err = ResolveConflict(tmpDir, "test.pdf")
	if err != nil {
		t.Fatalf("ResolveConflict failed: %v", err)
	}
	if filename != "test.pdf" {
		t.Errorf("expected 'test.pdf' (different ext), got %q", filename)
	}

	// Test 5: File with multiple dots in name
	multiDotFile := filepath.Join(tmpDir, "test.backup.md")
	if err := os.WriteFile(multiDotFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create multi-dot file: %v", err)
	}

	filename, err = ResolveConflict(tmpDir, "test.backup.md")
	if err != nil {
		t.Fatalf("ResolveConflict failed: %v", err)
	}
	if filename != "test.backup-1.md" {
		t.Errorf("expected 'test.backup-1.md', got %q", filename)
	}
}

// TestResolveConflict_NonexistentDirectory tests error handling
func TestResolveConflict_NonexistentDirectory(t *testing.T) {
	// This should work because we're just checking if files exist
	// The directory check happens elsewhere (validateDirectory)
	nonexistentDir := "/nonexistent/test/directory"

	// Should return original filename since directory doesn't exist
	// (no files can exist there)
	filename, err := ResolveConflict(nonexistentDir, "test.md")
	if err != nil {
		t.Fatalf("ResolveConflict with nonexistent dir failed: %v", err)
	}
	if filename != "test.md" {
		t.Errorf("expected 'test.md', got %q", filename)
	}
}

// TestSlugifyTitle_Truncation tests edge cases in truncation
func TestSlugifyTitle_Truncation(t *testing.T) {
	// Test that truncation doesn't leave trailing hyphens
	tests := []struct {
		title    string
		maxLen   int
		expected string
	}{
		{"hello-world-test", 11, "hello-world"}, // Should truncate and remove trailing hyphen
		{"test-", 10, "test"},                   // Trailing hyphen removed
		{"a-b-c-d-e-f-g", 5, "a-b-c"},           // Multiple truncations
	}

	for _, tt := range tests {
		result := SlugifyTitle(tt.title, tt.maxLen)
		if result != tt.expected {
			t.Errorf("SlugifyTitle(%q, %d) = %q, expected %q",
				tt.title, tt.maxLen, result, tt.expected)
		}

		// Verify no trailing hyphens
		if strings.HasSuffix(result, "-") {
			t.Errorf("SlugifyTitle(%q, %d) has trailing hyphen: %q",
				tt.title, tt.maxLen, result)
		}

		// Verify no leading hyphens
		if strings.HasPrefix(result, "-") {
			t.Errorf("SlugifyTitle(%q, %d) has leading hyphen: %q",
				tt.title, tt.maxLen, result)
		}
	}
}

func TestSlugifyTitle_UnicodeNormalization(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		expected string
	}{
		{
			name:     "emoji",
			title:    "Hello 🚀 World",
			expected: "hello-world",
		},
		{
			name:     "chinese characters",
			title:    "中文标题 English",
			expected: "english",
		},
		{
			name:     "arabic text",
			title:    "عربي Arabic Text",
			expected: "arabic-text",
		},
		{
			name:     "mixed unicode",
			title:    "Café ☕ 2025",
			expected: "caf-2025",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SlugifyTitle(tt.title, 80)
			if result != tt.expected {
				t.Errorf("SlugifyTitle() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestGenerateFilename_InvalidChars(t *testing.T) {
	// Test filesystem-restricted characters
	tests := []struct {
		name  string
		title string
	}{
		{"less than", "File<Name"},
		{"greater than", "File>Name"},
		{"colon", "File:Name"},
		{"quote", "File\"Name"},
		{"pipe", "File|Name"},
		{"question mark", "File?Name"},
		{"asterisk", "File*Name"},
	}

	timestamp := time.Date(2025, 10, 21, 14, 30, 45, 0, time.UTC)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateFilename(tt.title, format.Markdown, timestamp, "https://example.com")

			// Result should not contain filesystem-restricted chars
			restricted := []string{"<", ">", ":", "\"", "|", "?", "*"}
			for _, char := range restricted {
				if strings.Contains(result, char) {
					t.Errorf("filename %q should not contain restricted char %q", result, char)
				}
			}
		})
	}
}

func TestResolveConflict_HighCount(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test.md and test-1.md through test-50.md
	baseFile := filepath.Join(tmpDir, "test.md")
	os.WriteFile(baseFile, []byte("test"), 0644)

	for i := 1; i <= 50; i++ {
		conflictFile := filepath.Join(tmpDir, fmt.Sprintf("test-%d.md", i))
		os.WriteFile(conflictFile, []byte("test"), 0644)
	}

	// Should return test-51.md
	filename, err := ResolveConflict(tmpDir, "test.md")
	if err != nil {
		t.Fatalf("ResolveConflict failed: %v", err)
	}

	expected := "test-51.md"
	if filename != expected {
		t.Errorf("expected %q, got %q", expected, filename)
	}
}
