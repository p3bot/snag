// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"strings"
	"testing"
)

// TestFormatSection tests the section header formatting.
func TestFormatSection(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		expected string
	}{
		{
			name:     "Normal section title",
			title:    "Version Information",
			expected: "\nVersion Information\n───────────────────\n",
		},
		{
			name:     "Short title",
			title:    "Test",
			expected: "\nTest\n────\n",
		},
		{
			name:     "Empty title",
			title:    "",
			expected: "\n\n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dr := &DoctorReport{}
			result := dr.formatSection(tt.title)
			if result != tt.expected {
				t.Errorf("formatSection(%q) =\n%q\nexpected:\n%q", tt.title, result, tt.expected)
			}
		})
	}
}

// TestFormatItem tests the item formatting.
func TestFormatItem(t *testing.T) {
	tests := []struct {
		name     string
		label    string
		value    string
		expected string
	}{
		{
			name:     "Normal item",
			label:    "snag version",
			value:    "0.0.5",
			expected: "  snag version:        0.0.5\n",
		},
		{
			name:     "Long value",
			label:    "Path",
			value:    "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			expected: "  Path:                /Applications/Google Chrome.app/Contents/MacOS/Google Chrome\n",
		},
		{
			name:     "Empty value",
			label:    "Test",
			value:    "",
			expected: "  Test:                \n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dr := &DoctorReport{}
			result := dr.formatItem(tt.label, tt.value)
			if result != tt.expected {
				t.Errorf("formatItem(%q, %q) =\n%q\nexpected:\n%q", tt.label, tt.value, result, tt.expected)
			}
		})
	}
}

// TestFormatCheck tests the checkmark formatting.
func TestFormatCheck(t *testing.T) {
	tests := []struct {
		name     string
		label    string
		value    string
		ok       bool
		expected string
	}{
		{
			name:     "Success checkmark",
			label:    "Port 9222",
			value:    "Running (7 tabs open)",
			ok:       true,
			expected: "  Port 9222:           ✓ Running (7 tabs open)\n",
		},
		{
			name:     "Failure X mark",
			label:    "Port 9222",
			value:    "Not running",
			ok:       false,
			expected: "  Port 9222:           ✗ Not running\n",
		},
		{
			name:     "Browser found",
			label:    "Chrome",
			value:    "/Users/test/Library/Application Support/Google/Chrome",
			ok:       true,
			expected: "  Chrome:              ✓ /Users/test/Library/Application Support/Google/Chrome\n",
		},
		{
			name:     "Browser not found",
			label:    "Detected",
			value:    "No Chromium-based browser found",
			ok:       false,
			expected: "  Detected:            ✗ No Chromium-based browser found\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dr := &DoctorReport{}
			result := dr.formatCheck(tt.label, tt.value, tt.ok)
			if result != tt.expected {
				t.Errorf("formatCheck(%q, %q, %v) =\n%q\nexpected:\n%q", tt.label, tt.value, tt.ok, result, tt.expected)
			}
		})
	}
}

// TestFormatPortStatus tests port status formatting.
func TestFormatPortStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   *PortStatus
		expected string
	}{
		{
			name: "Running with tabs",
			status: &PortStatus{
				Port:     9222,
				Running:  true,
				TabCount: 7,
			},
			expected: "  Port 9222:           ✓ Running (7 tabs open)\n",
		},
		{
			name: "Not running",
			status: &PortStatus{
				Port:    9222,
				Running: false,
			},
			expected: "  Port 9222:           ✗ Not running\n",
		},
		{
			name: "Custom port running",
			status: &PortStatus{
				Port:     9223,
				Running:  true,
				TabCount: 3,
			},
			expected: "  Port 9223:           ✓ Running (3 tabs open)\n",
		},
		{
			name: "Zero tabs",
			status: &PortStatus{
				Port:     9222,
				Running:  true,
				TabCount: 0,
			},
			expected: "  Port 9222:           ✓ Running (0 tabs open)\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dr := &DoctorReport{}
			result := dr.formatPortStatus(tt.status)
			if result != tt.expected {
				t.Errorf("formatPortStatus() =\n%q\nexpected:\n%q", result, tt.expected)
			}
		})
	}
}

// TestDoctorReportString tests the full String() output.
func TestDoctorReportString(t *testing.T) {
	report := &DoctorReport{
		SnagVersion:    "0.0.5",
		LatestVersion:  "0.0.6",
		GoVersion:      "go1.25.3",
		OS:             "darwin",
		Arch:           "arm64",
		WorkingDir:     "/Users/test/projects/snag",
		BrowserName:    "Chrome",
		BrowserPath:    "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
		BrowserVersion: "Google Chrome 141.0.7390.123",
		ProfilePath:    "/Users/test/Library/Application Support/Google/Chrome",
		ProfileExists:  true,
		DefaultPortStatus: &PortStatus{
			Port:     9222,
			Running:  true,
			TabCount: 7,
		},
		EnvVars: map[string]string{
			"CHROME_PATH":   "",
			"CHROMIUM_PATH": "",
		},
	}

	output := report.String()

	// Test that all major sections are present
	expectedSections := []string{
		"snag Doctor Report",
		"==================",
		"Repository:    https://github.com/p3bot/snag",
		"Report Issue:  https://github.com/p3bot/snag/issues/new",
		"Version Information",
		"Working Directory",
		"Browser Detection",
		"Profile Location",
		"Connection Status",
		"Environment Variables",
	}

	for _, section := range expectedSections {
		if !strings.Contains(output, section) {
			t.Errorf("String() output missing section: %q", section)
		}
	}

	// Test that version information is present
	expectedContent := []string{
		"snag version:        0.0.5",
		"Latest version:      0.0.6 (update available)",
		"Go version:          go1.25.3",
		"OS/Arch:             darwin/arm64",
		"/Users/test/projects/snag",
		"Detected:            Chrome",
		"Version:             Google Chrome 141.0.7390.123",
		"✓ /Users/test/Library/Application Support/Google/Chrome",
		"✓ Running (7 tabs open)",
		"CHROME_PATH:         (not set)",
		"CHROMIUM_PATH:       (not set)",
	}

	for _, content := range expectedContent {
		if !strings.Contains(output, content) {
			t.Errorf("String() output missing content: %q", content)
		}
	}
}

// TestDoctorReportString_NoUpdateAvailable tests when versions match.
func TestDoctorReportString_NoUpdateAvailable(t *testing.T) {
	report := &DoctorReport{
		SnagVersion:   "0.0.5",
		LatestVersion: "0.0.5",
		GoVersion:     "go1.25.3",
		OS:            "linux",
		Arch:          "amd64",
		WorkingDir:    "/home/user/snag",
		EnvVars:       map[string]string{},
	}

	output := report.String()

	// Should show version but NOT "update available"
	if !strings.Contains(output, "Latest version:      0.0.5") {
		t.Error("String() should show latest version")
	}
	if strings.Contains(output, "(update available)") {
		t.Error("String() should not show update available when versions match")
	}
}

// TestDoctorReportString_NoLatestVersion tests when GitHub check fails.
func TestDoctorReportString_NoLatestVersion(t *testing.T) {
	report := &DoctorReport{
		SnagVersion:   "0.0.5",
		LatestVersion: "", // Empty when GitHub check fails
		GoVersion:     "go1.25.3",
		OS:            "darwin",
		Arch:          "arm64",
		WorkingDir:    "/Users/test/snag",
		EnvVars:       map[string]string{},
	}

	output := report.String()

	// Should NOT show "Latest version" line at all
	if strings.Contains(output, "Latest version:") {
		t.Error("String() should not show Latest version when check failed")
	}
	// But should still show snag version
	if !strings.Contains(output, "snag version:        0.0.5") {
		t.Error("String() should show snag version")
	}
}

// TestDoctorReportString_NoBrowser tests when no browser is detected.
func TestDoctorReportString_NoBrowser(t *testing.T) {
	report := &DoctorReport{
		SnagVersion:  "0.0.5",
		GoVersion:    "go1.25.3",
		OS:           "linux",
		Arch:         "amd64",
		WorkingDir:   "/home/user/snag",
		BrowserError: ErrBrowserNotFound,
		DefaultPortStatus: &PortStatus{
			Port:    9222,
			Running: false,
		},
		EnvVars: map[string]string{
			"CHROME_PATH":   "",
			"CHROMIUM_PATH": "",
		},
	}

	output := report.String()

	expectedContent := []string{
		"Browser Detection",
		"✗ No Chromium-based browser found",
		"Path:                (none)",
		"Version:             (none)",
	}

	for _, content := range expectedContent {
		if !strings.Contains(output, content) {
			t.Errorf("String() output missing content for no browser: %q", content)
		}
	}

	// Should NOT show Profile Location section when no browser
	if strings.Contains(output, "Profile Location") {
		t.Error("String() should not show Profile Location when no browser detected")
	}
}

// TestDoctorReportString_CustomPort tests with custom port.
func TestDoctorReportString_CustomPort(t *testing.T) {
	report := &DoctorReport{
		SnagVersion: "0.0.5",
		GoVersion:   "go1.25.3",
		OS:          "darwin",
		Arch:        "arm64",
		WorkingDir:  "/Users/test/snag",
		DefaultPortStatus: &PortStatus{
			Port:    9222,
			Running: false,
		},
		CustomPortStatus: &PortStatus{
			Port:     9223,
			Running:  true,
			TabCount: 3,
		},
		EnvVars: map[string]string{},
	}

	output := report.String()

	expectedContent := []string{
		"Connection Status",
		"Port 9222:           ✗ Not running",
		"Port 9223:           ✓ Running (3 tabs open)",
	}

	for _, content := range expectedContent {
		if !strings.Contains(output, content) {
			t.Errorf("String() output missing custom port content: %q", content)
		}
	}
}

// TestDoctorReportString_BrowserNoVersion tests browser detected but version unknown.
func TestDoctorReportString_BrowserNoVersion(t *testing.T) {
	report := &DoctorReport{
		SnagVersion:    "0.0.5",
		GoVersion:      "go1.25.3",
		OS:             "linux",
		Arch:           "amd64",
		WorkingDir:     "/home/user/snag",
		BrowserName:    "Chromium",
		BrowserPath:    "/usr/bin/chromium",
		BrowserVersion: "", // Version check failed
		EnvVars:        map[string]string{},
	}

	output := report.String()

	if !strings.Contains(output, "Detected:            Chromium") {
		t.Error("String() should show detected browser")
	}
	if !strings.Contains(output, "Version:             (unknown)") {
		t.Error("String() should show (unknown) when version check fails")
	}
}

// TestDoctorReportString_EnvVarsSet tests with environment variables set.
func TestDoctorReportString_EnvVarsSet(t *testing.T) {
	report := &DoctorReport{
		SnagVersion: "0.0.5",
		GoVersion:   "go1.25.3",
		OS:          "linux",
		Arch:        "amd64",
		WorkingDir:  "/home/user/snag",
		EnvVars: map[string]string{
			"CHROME_PATH":   "/opt/google/chrome/chrome",
			"CHROMIUM_PATH": "",
		},
	}

	output := report.String()

	if !strings.Contains(output, "CHROME_PATH:         /opt/google/chrome/chrome") {
		t.Error("String() should show set environment variable")
	}
	if !strings.Contains(output, "CHROMIUM_PATH:       (not set)") {
		t.Error("String() should show (not set) for empty env var")
	}
}

// TestCollectDoctorInfo tests the data collection function.
func TestCollectDoctorInfo(t *testing.T) {
	report, err := CollectDoctorInfo(9222)

	if err != nil {
		t.Fatalf("CollectDoctorInfo() returned error: %v", err)
	}

	// Verify basic fields are populated
	if report.SnagVersion == "" {
		t.Error("SnagVersion should not be empty")
	}
	if report.GoVersion == "" {
		t.Error("GoVersion should not be empty")
	}
	if report.OS == "" {
		t.Error("OS should not be empty")
	}
	if report.Arch == "" {
		t.Error("Arch should not be empty")
	}
	if report.WorkingDir == "" {
		t.Error("WorkingDir should not be empty")
	}
	if report.EnvVars == nil {
		t.Error("EnvVars should not be nil")
	}

	// Verify environment variables are checked
	if _, exists := report.EnvVars["CHROME_PATH"]; !exists {
		t.Error("EnvVars should contain CHROME_PATH")
	}
	if _, exists := report.EnvVars["CHROMIUM_PATH"]; !exists {
		t.Error("EnvVars should contain CHROMIUM_PATH")
	}

	// Verify port status is checked
	if report.DefaultPortStatus == nil {
		t.Error("DefaultPortStatus should not be nil")
	}
	if report.DefaultPortStatus.Port != 9222 {
		t.Errorf("DefaultPortStatus.Port = %d, expected 9222", report.DefaultPortStatus.Port)
	}
}

// TestCollectDoctorInfo_CustomPort tests collection with custom port.
func TestCollectDoctorInfo_CustomPort(t *testing.T) {
	report, err := CollectDoctorInfo(9223)

	if err != nil {
		t.Fatalf("CollectDoctorInfo() returned error: %v", err)
	}

	// Should check both default and custom port
	if report.DefaultPortStatus == nil {
		t.Error("DefaultPortStatus should not be nil")
	}
	if report.CustomPortStatus == nil {
		t.Error("CustomPortStatus should not be nil when custom port specified")
	}
	if report.CustomPortStatus.Port != 9223 {
		t.Errorf("CustomPortStatus.Port = %d, expected 9223", report.CustomPortStatus.Port)
	}
}

// TestCollectDoctorInfo_DefaultPort tests that custom status is nil for default port.
func TestCollectDoctorInfo_DefaultPort(t *testing.T) {
	report, err := CollectDoctorInfo(9222)

	if err != nil {
		t.Fatalf("CollectDoctorInfo() returned error: %v", err)
	}

	// Should only check default port, not create custom port status
	if report.DefaultPortStatus == nil {
		t.Error("DefaultPortStatus should not be nil")
	}
	if report.CustomPortStatus != nil {
		t.Error("CustomPortStatus should be nil when port is default 9222")
	}
}

// TestDoctorReportPrint tests that Print() calls String().
func TestDoctorReportPrint(t *testing.T) {
	report := &DoctorReport{
		SnagVersion: "0.0.5",
		GoVersion:   "go1.25.3",
		OS:          "darwin",
		Arch:        "arm64",
		WorkingDir:  "/Users/test/snag",
		EnvVars:     map[string]string{},
	}

	// Get the String() output
	expected := report.String()

	// Print() should produce identical output to String()
	// This is a simple test - in real testing you'd capture stdout
	// but since Print() just calls String(), we trust it works
	if expected == "" {
		t.Error("String() should not return empty output")
	}

	// Verify Print() doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Print() panicked: %v", r)
		}
	}()

	// Call Print() to verify it doesn't panic
	// Output goes to stdout but we're just testing for panics
	report.Print()
}

// TestCheckLatestVersion tests the GitHub version check (may fail if offline).
func TestCheckLatestVersion(t *testing.T) {
	// This test may fail if offline or GitHub is down
	// It should not panic or hang
	version := checkLatestVersion()

	// Version might be empty if network fails, that's OK
	// Just verify it doesn't panic and completes in reasonable time
	t.Logf("Latest version from GitHub: %q (empty if network error)", version)
}

// TestCheckPortConnection tests port connection checking.
func TestCheckPortConnection(t *testing.T) {
	// Test connection to port that's likely not running
	status := checkPortConnection(19999) // Unlikely port

	if status == nil {
		t.Fatal("checkPortConnection() should never return nil")
	}
	if status.Port != 19999 {
		t.Errorf("Port = %d, expected 19999", status.Port)
	}

	// Port is likely not running
	// This is just testing the function doesn't crash
	t.Logf("Port %d status: Running=%v, TabCount=%d", status.Port, status.Running, status.TabCount)
}
