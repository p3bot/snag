// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package fetch

import (
	"encoding/json"
	"testing"
)

// TestExtractDomain tests domain extraction from URLs
func TestExtractDomain(t *testing.T) {
	tests := []struct {
		url      string
		expected string
		desc     string
	}{
		// Standard domains
		{"https://example.com", "example.com", "simple domain"},
		{"https://example.com/path", "example.com", "with path"},
		{"https://example.com:8080", "example.com", "with port"},
		{"https://example.com:8080/path", "example.com", "with port and path"},

		// www prefix removal
		{"https://www.example.com", "example.com", "www prefix removed"},
		{"https://www.example.com/path", "example.com", "www with path"},

		// Subdomains
		{"https://api.example.com", "api.example.com", "subdomain kept"},
		{"https://docs.api.example.com", "docs.api.example.com", "nested subdomain"},

		// Edge cases
		{"", "", "empty URL"},
		{"invalid", "", "invalid URL"},
		{"file:///local/path", "", "file URL"},
		{"https://localhost:3000", "localhost", "localhost"},
		{"https://127.0.0.1:8080", "127.0.0.1", "IP address"},

		{"https://[::1]", "::1", "IPv6 loopback"},
		{"https://[::1]:8080", "::1", "IPv6 loopback with port"},
		{"https://[2001:db8::1]/path", "2001:db8::1", "IPv6 documentation address"},
		{"https://[2001:db8::1]:443/path", "2001:db8::1", "IPv6 with port and path"},
	}

	for _, tt := range tests {
		result := extractDomain(tt.url)
		if result != tt.expected {
			t.Errorf("extractDomain(%q) [%s] = %q, expected %q",
				tt.url, tt.desc, result, tt.expected)
		}
	}
}

// TestPageInfoJSON tests JSON marshalling of PageInfo
func TestPageInfoJSON(t *testing.T) {
	info := &PageInfo{
		Title:     "Test Page",
		URL:       "https://example.com",
		Domain:    "example.com",
		Slug:      "test-page",
		Timestamp: "2025-02-04T10:30:00+10:00",
	}

	jsonData, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal PageInfo: %v", err)
	}

	// Verify JSON structure
	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Check all fields are present
	expectedFields := []string{"title", "url", "domain", "slug", "timestamp"}
	for _, field := range expectedFields {
		if _, ok := parsed[field]; !ok {
			t.Errorf("Missing field %q in JSON output", field)
		}
	}

	// Check values
	if parsed["title"] != "Test Page" {
		t.Errorf("title = %v, expected 'Test Page'", parsed["title"])
	}
	if parsed["url"] != "https://example.com" {
		t.Errorf("url = %v, expected 'https://example.com'", parsed["url"])
	}
	if parsed["domain"] != "example.com" {
		t.Errorf("domain = %v, expected 'example.com'", parsed["domain"])
	}
	if parsed["slug"] != "test-page" {
		t.Errorf("slug = %v, expected 'test-page'", parsed["slug"])
	}
}

// TestPageInfoJSONRoundtrip tests JSON marshalling and unmarshalling
func TestPageInfoJSONRoundtrip(t *testing.T) {
	original := &PageInfo{
		Title:     "Example Domain",
		URL:       "https://example.com/",
		Domain:    "example.com",
		Slug:      "example-domain",
		Timestamp: "2025-02-04T14:30:22+10:00",
	}

	jsonData, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var parsed PageInfo
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if parsed.Title != original.Title {
		t.Errorf("Title mismatch: got %q, expected %q", parsed.Title, original.Title)
	}
	if parsed.URL != original.URL {
		t.Errorf("URL mismatch: got %q, expected %q", parsed.URL, original.URL)
	}
	if parsed.Domain != original.Domain {
		t.Errorf("Domain mismatch: got %q, expected %q", parsed.Domain, original.Domain)
	}
	if parsed.Slug != original.Slug {
		t.Errorf("Slug mismatch: got %q, expected %q", parsed.Slug, original.Slug)
	}
	if parsed.Timestamp != original.Timestamp {
		t.Errorf("Timestamp mismatch: got %q, expected %q", parsed.Timestamp, original.Timestamp)
	}
}

// TestPageInfoEmptyTitle tests slug generation when title is empty
func TestPageInfoEmptyTitle(t *testing.T) {
	info := &PageInfo{
		Title:     "",
		URL:       "https://example.com/page",
		Domain:    "example.com",
		Slug:      "", // Empty slug when title is empty
		Timestamp: "2025-02-04T10:30:00+10:00",
	}

	jsonData, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal PageInfo with empty title: %v", err)
	}

	// Should still produce valid JSON
	var parsed PageInfo
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal JSON with empty title: %v", err)
	}

	if parsed.Title != "" {
		t.Errorf("Expected empty title, got %q", parsed.Title)
	}
}

// TestExtractDomain_EdgeCases tests edge cases in domain extraction
func TestExtractDomain_EdgeCases(t *testing.T) {
	tests := []struct {
		url      string
		expected string
		desc     string
	}{
		// Multiple www
		{"https://www.www.example.com", "www.example.com", "double www"},

		// Uppercase (www. prefix not removed when uppercase)
		{"https://WWW.EXAMPLE.COM", "WWW.EXAMPLE.COM", "uppercase domain"},
		{"https://Example.Com", "Example.Com", "mixed case"},

		// Long domain
		{"https://subdomain.another.deep.example.com", "subdomain.another.deep.example.com", "deep subdomain"},

		// IDN domains
		{"https://www.例え.jp", "例え.jp", "IDN domain"},

		// Query strings and fragments (should be ignored)
		{"https://example.com?query=1", "example.com", "with query string"},
		{"https://example.com#section", "example.com", "with fragment"},
		{"https://example.com?query=1#section", "example.com", "with query and fragment"},
	}

	for _, tt := range tests {
		result := extractDomain(tt.url)
		if result != tt.expected {
			t.Errorf("extractDomain(%q) [%s] = %q, expected %q",
				tt.url, tt.desc, result, tt.expected)
		}
	}
}
