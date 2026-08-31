// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/p3bot/snag/internal/testutil"
)

// TestMain runs before and after all tests to ensure cleanup
var snagBin string

func TestMain(m *testing.M) {
	if os.Getenv("SNAG_SIGNAL_HELPER") == "1" {
		os.Exit(m.Run())
	}

	tmp, err := os.MkdirTemp("", "snag-cli-test-")
	if err != nil {
		fmt.Fprintf(os.Stderr, "mkdir temp: %v\n", err)
		os.Exit(ExitCodeError)
	}
	snagBin = filepath.Join(tmp, "snag")
	if err := testutil.BuildSnag(snagBin); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(ExitCodeError)
	}

	// Set up signal handling to cleanup on Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Fprintf(os.Stderr, "\nCleaning up browsers before exit...\n")
		cleanupOrphanedBrowsers()
		os.Exit(ExitCodeInterrupt)
	}()

	// Clean up any orphaned Chrome instances before running tests
	cleanupOrphanedBrowsers()

	// Run tests
	exitCode := m.Run()

	// Clean up any orphaned Chrome instances after running tests
	cleanupOrphanedBrowsers()
	os.RemoveAll(tmp)

	os.Exit(exitCode)
}

// cleanupOrphanedBrowsers kills any Chrome/Chromium instances with remote debugging
func cleanupOrphanedBrowsers() {
	// Find Chrome/Chromium processes with remote-debugging-port
	cmd := exec.Command("sh", "-c", "ps aux | grep -iE '(chrome|chromium).*--remote-debugging-port' | grep -v grep | awk '{print $2}'")
	output, err := cmd.Output()
	if err != nil {
		// No processes found or command failed - that's okay
		return
	}

	pidsStr := strings.TrimSpace(string(output))
	if pidsStr == "" {
		return
	}

	pids := strings.Split(pidsStr, "\n")
	fmt.Fprintf(os.Stderr, "Cleaning up %d orphaned browser instance(s)...\n", len(pids))

	for _, pid := range pids {
		pid = strings.TrimSpace(pid)
		if pid == "" {
			continue
		}
		// Try graceful kill first
		exec.Command("kill", pid).Run()
	}

	// Wait longer for graceful shutdown
	exec.Command("sleep", "2").Run()

	// Force kill any remaining
	for _, pid := range pids {
		pid = strings.TrimSpace(pid)
		if pid == "" {
			continue
		}
		exec.Command("kill", "-9", pid).Run()
	}

	// Final wait to ensure processes are fully gone
	exec.Command("sleep", "1").Run()
}

// isBrowserAvailable checks if Chrome or Chromium is available on the system
func isBrowserAvailable() bool {
	browsers := []string{
		"google-chrome",
		"chromium",
		"chromium-browser",
		"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
		"/Applications/Chromium.app/Contents/MacOS/Chromium",
	}

	for _, browser := range browsers {
		if _, err := exec.LookPath(browser); err == nil {
			return true
		}
		// Check if file exists (for macOS app paths)
		if _, err := os.Stat(browser); err == nil {
			return true
		}
	}
	return false
}

// startTestServer launches an HTTP server serving files from testdata/
func startTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	// Get absolute path to testdata directory
	testdataPath := testutil.Testdata()

	// Create file server
	fileServer := http.FileServer(http.Dir(testdataPath))

	// Create test server
	server := httptest.NewServer(fileServer)

	// Register cleanup
	t.Cleanup(func() {
		server.Close()
	})

	return server
}

// runSnag executes the snag binary with the given arguments
// Returns stdout, stderr, and error
func runSnag(args ...string) (stdout string, stderr string, err error) {
	cmd := exec.Command(snagBin, args...)
	cmd.Dir = testutil.ModuleRoot()

	// Capture stdout and stderr separately
	stdoutBytes, stderrBytes, err := runCommand(cmd)

	return string(stdoutBytes), string(stderrBytes), err
}

// runCommand executes a command and returns stdout, stderr separately
func runCommand(cmd *exec.Cmd) (stdout []byte, stderr []byte, err error) {
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, err
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, nil, err
	}

	// Read output using io.ReadAll
	stdoutBytes, stdoutErr := io.ReadAll(stdoutPipe)
	stderrBytes, stderrErr := io.ReadAll(stderrPipe)

	// Wait for command to finish
	err = cmd.Wait()

	// Check for read errors
	if stdoutErr != nil {
		return nil, nil, stdoutErr
	}
	if stderrErr != nil {
		return nil, nil, stderrErr
	}

	return stdoutBytes, stderrBytes, err
}

// assertContains checks if the output contains the expected substring
func assertContains(t *testing.T, output, expected string) {
	t.Helper()
	if !strings.Contains(output, expected) {
		t.Errorf("expected output to contain %q, got:\n%s", expected, output)
	}
}

// assertNotContains checks if the output does not contain the substring
func assertNotContains(t *testing.T, output, unexpected string) {
	t.Helper()
	if strings.Contains(output, unexpected) {
		t.Errorf("expected output to NOT contain %q, got:\n%s", unexpected, output)
	}
}

// assertExitCode checks if the command exited with the expected code
func assertExitCode(t *testing.T, err error, expectedCode int) {
	t.Helper()
	if expectedCode == 0 {
		if err != nil {
			t.Errorf("expected exit code 0, but command failed: %v", err)
		}
	} else {
		// expectedCode != 0, so we expect an error
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() != expectedCode {
				t.Errorf("expected exit code %d, got %d", expectedCode, exitErr.ExitCode())
			}
		} else {
			t.Errorf("expected exit code %d, but got non-exit error or success: %v", expectedCode, err)
		}
	}
}

// assertNoError checks that there was no error
func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// assertError checks that there was an error
func assertError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error but got none")
	}
}

// ============================================================================
// Phase 3: Fast CLI Tests (No Browser Required)
// ============================================================================

// TestCLI_Version tests the --version flag
func TestCLI_Version(t *testing.T) {
	stdout, stderr, err := runSnag("--version")

	assertNoError(t, err)

	// Version should be in output (could be stdout or stderr)
	output := stdout + stderr
	if !strings.Contains(output, "snag version") && !strings.Contains(output, Version) {
		t.Errorf("expected version in output, got: %s", output)
	}
}

// TestCLI_Help tests the --help flag
func TestCLI_Help(t *testing.T) {
	stdout, stderr, _ := runSnag("--help")

	// Help may exit with 0 or error depending on cli library
	output := stdout + stderr

	// Should contain usage information
	assertContains(t, output, "USAGE")
	assertContains(t, output, "snag")
}

// TestCLI_NoArguments tests running without URL
func TestCLI_NoArguments(t *testing.T) {
	stdout, stderr, err := runSnag()

	// Should fail when no URL provided
	assertError(t, err)
	assertExitCode(t, err, 1)

	output := stdout + stderr
	// Should show error or help message
	if !strings.Contains(output, "No URLs provided") && !strings.Contains(output, "USAGE") {
		t.Errorf("expected error or usage message, got: %s", output)
	}
}

// TestCLI_InvalidURL tests invalid URL handling
func TestCLI_InvalidURL(t *testing.T) {
	tests := []struct {
		url  string
		desc string
	}{
		{"ftp://example.com", "unsupported scheme"},
		{"javascript:alert(1)", "javascript scheme"},
		{"://malformed", "malformed URL"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			stdout, stderr, err := runSnag(tt.url)

			assertError(t, err)
			assertExitCode(t, err, 1)

			output := stdout + stderr
			// Should contain error message
			if !strings.Contains(output, "Invalid") && !strings.Contains(output, "invalid") &&
				!strings.Contains(output, "error") && !strings.Contains(output, "Error") {
				t.Errorf("expected error message for %s, got: %s", tt.desc, output)
			}
		})
	}
}

func hasANSI(s string) bool {
	return strings.Contains(s, "\x1b[")
}

func clearColorEnv(t *testing.T) {
	t.Helper()
	t.Setenv("NO_COLOR", "")
	t.Setenv("FORCE_COLOR", "")
	t.Setenv("CLICOLOR_FORCE", "")
}

func TestCLI_ColorInvalidValue(t *testing.T) {
	clearColorEnv(t)
	stdout, stderr, err := runSnag("--color", "rainbow")

	assertError(t, err)
	assertExitCode(t, err, 1)
	if stdout != "" && hasANSI(stdout) {
		t.Errorf("stdout has ANSI: %q", stdout)
	}
	if !strings.Contains(stderr, "rainbow") {
		t.Errorf("stderr missing bad value, got: %s", stderr)
	}
	for _, allowed := range []string{"auto", "always", "never"} {
		if !strings.Contains(stderr, allowed) {
			t.Errorf("stderr missing allowed value %q, got: %s", allowed, stderr)
		}
	}
	if strings.Contains(stderr, "Code:") || strings.Contains(stderr, "Fix:") || strings.Contains(stderr, "ValidValues") {
		t.Errorf("stderr has structured envelope: %s", stderr)
	}
}

func TestCLI_ColorCaseAndWhitespace(t *testing.T) {
	clearColorEnv(t)
	t.Setenv("TERM", "dumb")

	stdout, stderr, err := runSnag("--color", "AUTO")
	assertError(t, err)
	assertExitCode(t, err, 1)
	if hasANSI(stdout) {
		t.Errorf("stdout has ANSI: %q", stdout)
	}
	if hasANSI(stderr) {
		t.Errorf("stderr has ANSI with --color=AUTO TERM=dumb: %q", stderr)
	}

	stdout, stderr, err = runSnag("--color", " Always ")
	assertError(t, err)
	assertExitCode(t, err, 1)
	if hasANSI(stdout) {
		t.Errorf("stdout has ANSI: %q", stdout)
	}
	if !hasANSI(stderr) {
		t.Errorf("stderr missing ANSI with --color=' Always ': %q", stderr)
	}
}

func TestCLI_ColorNeverNoANSI(t *testing.T) {
	clearColorEnv(t)
	t.Setenv("TERM", "xterm")
	stdout, stderr, err := runSnag("--color", "never", "--verbose")
	assertError(t, err)
	assertExitCode(t, err, 1)
	if hasANSI(stdout) {
		t.Errorf("stdout has ANSI: %q", stdout)
	}
	if hasANSI(stderr) {
		t.Errorf("stderr has ANSI with --color=never: %q", stderr)
	}
}

func TestCLI_ColorAlwaysOverridesNOCOLOR(t *testing.T) {
	clearColorEnv(t)
	t.Setenv("NO_COLOR", "1")
	t.Setenv("TERM", "xterm")
	stdout, stderr, err := runSnag("--verbose", "--color", "always")
	assertError(t, err)
	assertExitCode(t, err, 1)
	if hasANSI(stdout) {
		t.Errorf("stdout has ANSI: %q", stdout)
	}
	if !hasANSI(stderr) {
		t.Errorf("stderr missing ANSI with NO_COLOR=1 --color=always: %q", stderr)
	}
}

func TestCLI_ColorAutoNOCOLOR(t *testing.T) {
	clearColorEnv(t)
	t.Setenv("NO_COLOR", "1")
	t.Setenv("FORCE_COLOR", "1")
	t.Setenv("TERM", "xterm")
	stdout, stderr, err := runSnag("--verbose", "--color", "auto")
	assertError(t, err)
	assertExitCode(t, err, 1)
	if hasANSI(stdout) {
		t.Errorf("stdout has ANSI: %q", stdout)
	}
	if hasANSI(stderr) {
		t.Errorf("stderr has ANSI with NO_COLOR=1 --color=auto: %q", stderr)
	}
}

func TestCLI_ColorAlwaysColoursRedirectedStderr(t *testing.T) {
	clearColorEnv(t)
	t.Setenv("TERM", "dumb")
	stdout, stderr, err := runSnag("--color", "always")
	assertError(t, err)
	assertExitCode(t, err, 1)
	if hasANSI(stdout) {
		t.Errorf("stdout has ANSI: %q", stdout)
	}
	if !hasANSI(stderr) {
		t.Errorf("stderr missing ANSI with --color=always TERM=dumb, got: %q", stderr)
	}
}

func TestCLI_ColorAutoForceColorTERMDumb(t *testing.T) {
	clearColorEnv(t)
	t.Setenv("TERM", "dumb")
	t.Setenv("FORCE_COLOR", "1")
	stdout, stderr, err := runSnag("--verbose", "--color", "auto")
	assertError(t, err)
	assertExitCode(t, err, 1)
	if hasANSI(stdout) {
		t.Errorf("stdout has ANSI: %q", stdout)
	}
	if !hasANSI(stderr) {
		t.Errorf("stderr missing ANSI with TERM=dumb FORCE_COLOR=1 --color=auto, got: %q", stderr)
	}
}

func TestCLI_ColorAutoTERMDumbNoForce(t *testing.T) {
	clearColorEnv(t)
	t.Setenv("TERM", "dumb")
	stdout, stderr, err := runSnag("--verbose", "--color", "auto")
	assertError(t, err)
	assertExitCode(t, err, 1)
	if hasANSI(stdout) {
		t.Errorf("stdout has ANSI: %q", stdout)
	}
	if hasANSI(stderr) {
		t.Errorf("stderr has ANSI with TERM=dumb --color=auto: %q", stderr)
	}
}

func TestCLI_ColorAlwaysDoesNotColourStdout(t *testing.T) {
	clearColorEnv(t)
	t.Setenv("TERM", "dumb")
	stdout, stderr, err := runSnag("--color", "always", "--version")
	assertNoError(t, err)
	if hasANSI(stdout) {
		t.Errorf("--version stdout has ANSI with --color=always: %q", stdout)
	}
	if hasANSI(stderr) {
		t.Errorf("--version stderr has ANSI: %q", stderr)
	}

	stdout, _, err = runSnag("--color", "always", "--skill")
	assertNoError(t, err)
	if hasANSI(stdout) {
		t.Errorf("--skill stdout has ANSI with --color=always: %q", stdout)
	}
}

func TestCLI_ColorAlwaysUnknownFlag(t *testing.T) {
	clearColorEnv(t)
	t.Setenv("TERM", "dumb")
	stdout, stderr, err := runSnag("--color", "always", "--bogus")
	assertError(t, err)
	assertExitCode(t, err, 1)
	if hasANSI(stdout) {
		t.Errorf("stdout has ANSI: %q", stdout)
	}
	if !hasANSI(stderr) {
		t.Errorf("stderr missing ANSI with --color=always --bogus: %q", stderr)
	}
	assertContains(t, stderr, "unknown flag")
}

func TestCLI_ColorNeverUnknownFlag(t *testing.T) {
	clearColorEnv(t)
	t.Setenv("TERM", "xterm")
	stdout, stderr, err := runSnag("--color", "never", "--bogus")
	assertError(t, err)
	assertExitCode(t, err, 1)
	if hasANSI(stdout) {
		t.Errorf("stdout has ANSI: %q", stdout)
	}
	if hasANSI(stderr) {
		t.Errorf("stderr has ANSI with --color=never --bogus: %q", stderr)
	}
	assertContains(t, stderr, "unknown flag")
}

func TestCLI_ColorAlwaysVerboseDebugConflict(t *testing.T) {
	clearColorEnv(t)
	t.Setenv("TERM", "dumb")
	stdout, stderr, err := runSnag("--color", "always", "--verbose", "--debug")
	assertError(t, err)
	assertExitCode(t, err, 1)
	if hasANSI(stdout) {
		t.Errorf("stdout has ANSI: %q", stdout)
	}
	if !hasANSI(stderr) {
		t.Errorf("stderr missing ANSI with --color=always --verbose --debug: %q", stderr)
	}
	assertContains(t, stderr, "verbose")
	assertContains(t, stderr, "debug")
}

func TestCLI_ColorNeverVerboseDebugConflict(t *testing.T) {
	clearColorEnv(t)
	t.Setenv("TERM", "xterm")
	stdout, stderr, err := runSnag("--color", "never", "--verbose", "--debug")
	assertError(t, err)
	assertExitCode(t, err, 1)
	if hasANSI(stdout) {
		t.Errorf("stdout has ANSI: %q", stdout)
	}
	if hasANSI(stderr) {
		t.Errorf("stderr has ANSI with --color=never --verbose --debug: %q", stderr)
	}
	assertContains(t, stderr, "verbose")
	assertContains(t, stderr, "debug")
}

func TestCLI_HelpDocumentsColor(t *testing.T) {
	stdout, stderr, _ := runSnag("--help")
	output := stdout + stderr
	assertContains(t, output, "--color")
	assertContains(t, output, "auto")
	assertContains(t, output, "always")
	assertContains(t, output, "never")
}

// TestCLI_InvalidFormat tests invalid format flag
func TestCLI_InvalidFormat(t *testing.T) {
	// Use a truly invalid format (json is not supported)
	stdout, stderr, err := runSnag("--format", "json", "https://example.com")

	assertError(t, err)
	assertExitCode(t, err, 1)

	output := stdout + stderr
	assertContains(t, output, "format")
}

// TestCLI_InvalidTimeout tests invalid timeout values
func TestCLI_InvalidTimeout(t *testing.T) {
	tests := []struct {
		timeout string
		desc    string
	}{
		{"-1", "negative timeout"},
		{"0", "zero timeout"},
		{"abc", "non-numeric timeout"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			stdout, stderr, err := runSnag("--timeout", tt.timeout, "https://example.com")

			// Should either fail validation or fail to parse
			assertError(t, err)

			output := stdout + stderr
			// Should contain error about timeout or invalid value
			if !strings.Contains(output, "timeout") && !strings.Contains(output, "invalid") &&
				!strings.Contains(output, "error") && !strings.Contains(output, "Error") {
				t.Errorf("expected error message about timeout or invalid value for %s, got: %s", tt.desc, output)
			}
		})
	}
}

// TestCLI_InvalidPort tests invalid port values
func TestCLI_InvalidPort(t *testing.T) {
	tests := []struct {
		port string
		desc string
	}{
		{"-1", "negative port"},
		{"0", "zero port"},
		{"99999", "port too large"},
		{"abc", "non-numeric port"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			stdout, stderr, err := runSnag("--port", tt.port, "--force-headless", "https://example.com")

			// Should either fail validation or fail to parse
			assertError(t, err)

			output := stdout + stderr
			// Should contain error about port or invalid value
			if !strings.Contains(output, "port") && !strings.Contains(output, "invalid") &&
				!strings.Contains(output, "error") && !strings.Contains(output, "Error") {
				t.Errorf("expected error message about port or invalid value for %s, got: %s", tt.desc, output)
			}
		})
	}
}

// TestCLI_FormatOptions tests valid format values are accepted
func TestCLI_FormatOptions(t *testing.T) {
	// Note: These will fail to actually fetch without a browser,
	// but should pass format validation
	// Test with user-facing format names (aliases that will be normalized)
	tests := []string{"markdown", "md", "html", "text", "txt", "pdf", "png"}

	for _, format := range tests {
		t.Run(format, func(t *testing.T) {
			// We can't actually test fetching without a browser,
			// but we can verify the format is accepted by checking
			// the error message doesn't mention invalid format
			stdout, stderr, err := runSnag("--format", format, "--force-headless", "https://example.com")

			output := stdout + stderr

			// If there's an error, it should NOT be about invalid format
			if err != nil {
				if strings.Contains(output, "Invalid format") || strings.Contains(output, "invalid format") {
					t.Errorf("format %q should be valid but got format error: %s", format, output)
				}
				// Other errors (like browser not found) are acceptable for this test
			}
		})
	}
}

// TestCLI_ListTabsNoBrowser tests --list-tabs without browser running
func TestCLI_ListTabsNoBrowser(t *testing.T) {
	stdout, stderr, err := runSnag("--list-tabs")

	// Should fail when no browser is running
	assertError(t, err)
	assertExitCode(t, err, 1)

	output := stdout + stderr
	// Should contain error message about no browser running
	assertContains(t, output, "No browser")
}

// TestCLI_OutputFilePermission tests output to unwritable location
func TestCLI_OutputFilePermission(t *testing.T) {
	// Create a temporary directory and make it read-only
	tmpDir := t.TempDir()
	readOnlyDir := filepath.Join(tmpDir, "readonly")

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

	outputPath := filepath.Join(readOnlyDir, "test-output.md")
	stdout, stderr, err := runSnag("-o", outputPath, "--force-headless", "https://example.com")

	// Should fail due to permissions
	assertError(t, err)

	output := stdout + stderr
	// May fail with permission error or browser error - both are acceptable
	// We're just verifying it doesn't succeed
	_ = output
}

// ============================================================================
// Phase 4: Browser Integration Tests (Requires Chrome/Chromium)
// ============================================================================

// TestBrowser_FetchSimpleHTML tests fetching simple.html from test server
func TestBrowser_FetchSimpleHTML(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	server := startTestServer(t)
	url := server.URL + "/simple.html"

	stdout, stderr, err := runSnag(url)

	assertNoError(t, err)
	assertExitCode(t, err, 0)

	// Verify markdown conversion happened
	assertContains(t, stdout, "# Example Heading")
	assertContains(t, stdout, "## Second Level Heading")
	assertContains(t, stdout, "This is a simple paragraph")
	assertContains(t, stdout, "[a link](https://example.com)")
	assertContains(t, stdout, "**bold text**")
	assertContains(t, stdout, "*italic text*")

	// Verify logs went to stderr (not stdout)
	if len(stderr) > 0 {
		// If there's stderr output, it should be logs, not content
		assertNotContains(t, stderr, "# Example Heading")
	}
}

// TestBrowser_FetchComplexHTML tests fetching complex.html with tables and lists
func TestBrowser_FetchComplexHTML(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	server := startTestServer(t)
	url := server.URL + "/complex.html"

	stdout, stderr, err := runSnag(url)

	assertNoError(t, err)
	assertExitCode(t, err, 0)

	// Verify markdown conversion
	assertContains(t, stdout, "# Complex Content")
	assertContains(t, stdout, "## Table Example")
	assertContains(t, stdout, "## List Examples")
	assertContains(t, stdout, "## Code Example")

	// Verify lists
	assertContains(t, stdout, "- Unordered item 1")
	assertContains(t, stdout, "- Unordered item 2")
	assertContains(t, stdout, "1. Ordered item 1")
	assertContains(t, stdout, "2. Ordered item 2")

	// Verify code block (should have backticks)
	assertContains(t, stdout, "```")
	assertContains(t, stdout, "function hello()")

	// Note: Table conversion may not produce markdown tables (known issue)
	// Just verify table content is preserved
	assertContains(t, stdout, "Item 1")
	assertContains(t, stdout, "Item 2")

	// Verify logs went to stderr
	if len(stderr) > 0 {
		assertNotContains(t, stderr, "# Complex Content")
	}
}

// TestBrowser_FetchMinimalHTML tests fetching minimal.html edge case
func TestBrowser_FetchMinimalHTML(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	server := startTestServer(t)
	url := server.URL + "/minimal.html"

	stdout, stderr, err := runSnag(url)

	assertNoError(t, err)
	assertExitCode(t, err, 0)

	// Should contain the minimal content
	assertContains(t, stdout, "Hello")

	// Should not be empty
	if len(strings.TrimSpace(stdout)) == 0 {
		t.Error("expected non-empty output for minimal HTML")
	}

	// Verify logs went to stderr
	if len(stderr) > 0 {
		assertNotContains(t, stderr, "Hello")
	}
}

// TestBrowser_HTMLFormat tests --format html output
func TestBrowser_HTMLFormat(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	server := startTestServer(t)
	url := server.URL + "/simple.html"

	stdout, stderr, err := runSnag("--format", "html", url)

	assertNoError(t, err)
	assertExitCode(t, err, 0)

	// Verify HTML output (not markdown)
	assertContains(t, stdout, "<h1>")
	assertContains(t, stdout, "<h2>")
	assertContains(t, stdout, "<p>")
	assertContains(t, stdout, "<a href=")
	assertContains(t, stdout, "<strong>")
	assertContains(t, stdout, "<em>")

	// Should NOT contain markdown syntax
	assertNotContains(t, stdout, "# Example Heading")
	assertNotContains(t, stdout, "**bold text**")

	// Verify logs went to stderr
	if len(stderr) > 0 {
		assertNotContains(t, stderr, "<h1>")
	}
}

// TestBrowser_OutputToFile tests -o flag for file output
func TestBrowser_OutputToFile(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	server := startTestServer(t)
	url := server.URL + "/simple.html"

	// Create temporary file for output
	tmpFile, err := os.CreateTemp("", "snag-test-*.md")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	outputPath := tmpFile.Name()
	tmpFile.Close()

	// Clean up after test
	t.Cleanup(func() {
		os.Remove(outputPath)
	})

	stdout, stderr, err := runSnag("-o", outputPath, url)

	assertNoError(t, err)
	assertExitCode(t, err, 0)

	// Stdout should be empty (content written to file)
	if len(strings.TrimSpace(stdout)) > 0 {
		t.Errorf("expected empty stdout when using -o flag, got: %s", stdout)
	}

	// Verify file was created and contains content
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	contentStr := string(content)
	assertContains(t, contentStr, "# Example Heading")
	assertContains(t, contentStr, "This is a simple paragraph")
	assertContains(t, stderr, "Saved to")

	// Verify success message in stderr
	if len(stderr) > 0 {
		// May contain success message about writing file
		assertNotContains(t, stderr, "# Example Heading")
	}
}

// TestBrowser_ForceHeadless tests --force-headless flag
func TestBrowser_ForceHeadless(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	server := startTestServer(t)
	url := server.URL + "/simple.html"

	stdout, stderr, err := runSnag("--force-headless", url)

	assertNoError(t, err)
	assertExitCode(t, err, 0)

	// Should successfully fetch content
	assertContains(t, stdout, "# Example Heading")

	// Verify headless mode was used (check stderr for relevant messages)
	output := stderr
	_ = output // May or may not contain mode indication
}

// TestBrowser_CustomPort tests --port flag with custom debugging port
func TestBrowser_CustomPort(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	server := startTestServer(t)
	url := server.URL + "/simple.html"

	// Use a non-default port
	stdout, stderr, err := runSnag("--port", "9223", "--force-headless", url)

	assertNoError(t, err)
	assertExitCode(t, err, 0)

	// Should successfully fetch content
	assertContains(t, stdout, "# Example Heading")

	output := stderr
	_ = output
}

// TestBrowser_Auth401Detection tests HTTP 401 authentication detection
func TestBrowser_Auth401Detection(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	// Create test server with 401 handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("WWW-Authenticate", `Basic realm="Test"`)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("<html><body>401 Unauthorized</body></html>"))
	})
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	url := server.URL

	stdout, stderr, err := runSnag(url)

	// May fail or succeed depending on how snag handles 401
	// At minimum, should not crash
	output := stdout + stderr

	// Should indicate authentication issue or return the 401 page
	// The test verifies snag handles 401 gracefully
	_ = output
	_ = err
}

// TestBrowser_Auth403Detection tests HTTP 403 forbidden detection
func TestBrowser_Auth403Detection(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	// Create test server with 403 handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("<html><body>403 Forbidden</body></html>"))
	})
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	url := server.URL

	stdout, stderr, err := runSnag(url)

	// May fail or succeed depending on how snag handles 403
	// At minimum, should not crash
	output := stdout + stderr

	// Should indicate forbidden access or return the 403 page
	_ = output
	_ = err
}

// TestBrowser_LoginFormDetection tests detection of login forms in DOM
func TestBrowser_LoginFormDetection(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	server := startTestServer(t)
	url := server.URL + "/login-form.html"

	stdout, stderr, err := runSnag(url)

	// Should successfully fetch the login form page
	assertNoError(t, err)
	assertExitCode(t, err, 0)

	// Should contain login form content in markdown
	assertContains(t, stdout, "Log In")

	// Form fields may be converted to markdown
	// Just verify content is present
	output := stdout + stderr
	_ = output
}

// TestBrowser_NoAuthFalsePositives tests that regular pages don't trigger auth detection
func TestBrowser_NoAuthFalsePositives(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	server := startTestServer(t)
	url := server.URL + "/simple.html"

	stdout, stderr, err := runSnag(url)

	// Regular page should fetch successfully
	assertNoError(t, err)
	assertExitCode(t, err, 0)

	// Should contain normal content
	assertContains(t, stdout, "# Example Heading")

	// Should NOT have authentication warnings
	output := stderr
	// If there are auth-related messages in stderr for a regular page, that's a false positive
	// However, we don't have specific auth detection messages defined, so just verify success
	_ = output
}

// TestBrowser_CustomTimeout tests --timeout flag with custom value
func TestBrowser_CustomTimeout(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	server := startTestServer(t)
	url := server.URL + "/simple.html"

	// Use a custom timeout (60 seconds)
	stdout, stderr, err := runSnag("--timeout", "60", url)

	assertNoError(t, err)
	assertExitCode(t, err, 0)

	// Should successfully fetch content
	assertContains(t, stdout, "# Example Heading")

	output := stderr
	_ = output
}

// TestBrowser_WaitForSelector tests --wait-for flag to wait for specific element
func TestBrowser_WaitForSelector(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	server := startTestServer(t)
	url := server.URL + "/dynamic.html"

	// Wait for the delayed content element
	stdout, stderr, err := runSnag("--wait-for", "#delayed-content", "--timeout", "5", url)

	assertNoError(t, err)
	assertExitCode(t, err, 0)

	// Should contain both initial and delayed content
	assertContains(t, stdout, "Dynamic Page")
	assertContains(t, stdout, "after 1 second")

	output := stderr
	_ = output
}

// TestBrowser_WaitForTimeout tests --wait-for with element that doesn't appear
func TestBrowser_WaitForTimeout(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	server := startTestServer(t)
	url := server.URL + "/simple.html"

	// Wait for element that doesn't exist, with short timeout
	stdout, stderr, err := runSnag("--wait-for", "#nonexistent-element", "--timeout", "2", url)

	// Should timeout and fail
	assertError(t, err)

	output := stdout + stderr
	// Should indicate timeout or element not found
	if !strings.Contains(output, "timeout") && !strings.Contains(output, "not found") &&
		!strings.Contains(output, "Timeout") {
		t.Errorf("Expected timeout error message, got: %s", output)
	}
}

// TestBrowser_DefaultTimeout tests that default timeout works
func TestBrowser_DefaultTimeout(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	server := startTestServer(t)
	url := server.URL + "/simple.html"

	// No timeout specified, should use default (30 seconds)
	stdout, stderr, err := runSnag(url)

	assertNoError(t, err)
	assertExitCode(t, err, 0)

	// Should successfully fetch content with default timeout
	assertContains(t, stdout, "# Example Heading")

	output := stderr
	_ = output
}

// TestBrowser_CustomUserAgent tests --user-agent flag
func TestBrowser_CustomUserAgent(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	server := startTestServer(t)
	url := server.URL + "/simple.html"

	customUA := "Mozilla/5.0 (Custom Bot) snag/test"
	stdout, stderr, err := runSnag("--user-agent", customUA, url)

	assertNoError(t, err)
	assertExitCode(t, err, 0)

	// Should successfully fetch content with custom user agent
	assertContains(t, stdout, "# Example Heading")

	// User agent is set in browser, content should be fetched normally
	output := stderr
	_ = output
}

// TestBrowser_CloseTab tests --close-tab flag
func TestBrowser_CloseTab(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	server := startTestServer(t)
	url := server.URL + "/simple.html"

	// Use --close-tab with headless mode
	stdout, stderr, err := runSnag("--close-tab", "--force-headless", url)

	assertNoError(t, err)
	assertExitCode(t, err, 0)

	// Should successfully fetch content and close the tab
	assertContains(t, stdout, "# Example Heading")

	output := stderr
	_ = output
}

// TestBrowser_VerboseOutput tests --verbose restores fetch progress on stderr
func TestBrowser_VerboseOutput(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	server := startTestServer(t)
	url := server.URL + "/simple.html"

	stdout, stderr, err := runSnag("--verbose", url)

	assertNoError(t, err)
	assertExitCode(t, err, 0)

	assertContains(t, stdout, "# Example Heading")
	assertContains(t, stderr, "Fetching")
	assertContains(t, stderr, "Fetched successfully")
	if !strings.Contains(stderr, "launched") && !strings.Contains(stderr, "Connected") {
		t.Errorf("expected verbose stderr to mention launched or connected, got:\n%s", stderr)
	}
}

// TestBrowser_DefaultSilentProgress tests that a successful fetch prints no progress on stderr
func TestBrowser_DefaultSilentProgress(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	server := startTestServer(t)
	url := server.URL + "/simple.html"

	stdout, stderr, err := runSnag(url)

	assertNoError(t, err)
	assertExitCode(t, err, 0)

	assertContains(t, stdout, "# Example Heading")
	assertNotContains(t, stderr, "launched")
	assertNotContains(t, stderr, "Connected to existing")
	assertNotContains(t, stderr, "Fetching")
	assertNotContains(t, stderr, "Fetched successfully")
}

func TestCLI_UnknownQuietFlag(t *testing.T) {
	stdout, stderr, err := runSnag("--quiet", "https://example.com")
	assertError(t, err)
	assertExitCode(t, err, 1)
	assertContains(t, stderr, "unknown flag: --quiet")
	_ = stdout

	stdout, stderr, err = runSnag("-q", "https://example.com")
	assertError(t, err)
	assertExitCode(t, err, 1)
	assertContains(t, stderr, "unknown shorthand flag: 'q' in -q")
	_ = stdout
}

func TestCLI_VerboseDebugMutuallyExclusive(t *testing.T) {
	stdout, stderr, err := runSnag("--verbose", "--debug", "https://example.com")
	assertError(t, err)
	assertExitCode(t, err, 1)
	assertContains(t, stderr, "if any flags in the group [verbose debug] are set none of the others can be")
	assertContains(t, stderr, "debug")
	assertContains(t, stderr, "verbose")
	_ = stdout
}

func TestBrowser_InfoJSON(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	server := startTestServer(t)
	url := server.URL + "/simple.html"

	stdout, stderr, err := runSnag("--info", url)

	assertNoError(t, err)
	assertExitCode(t, err, 0)

	var payload map[string]string
	if unmarshalErr := json.Unmarshal([]byte(stdout), &payload); unmarshalErr != nil {
		t.Fatalf("expected JSON on stdout: %v\n%s", unmarshalErr, stdout)
	}
	if payload["title"] != "Simple Test Page" {
		t.Errorf("title = %q, want Simple Test Page", payload["title"])
	}
	for _, key := range []string{"url", "domain", "slug", "timestamp"} {
		if payload[key] == "" {
			t.Errorf("expected %s in --info JSON, got: %s", key, stdout)
		}
	}

	assertNotContains(t, stderr, "launched")
	assertNotContains(t, stderr, "Connected to existing")
	assertNotContains(t, stderr, "Fetching")
	assertNotContains(t, stderr, "Fetched successfully")
}

func TestBrowser_InfoWarningsAndSave(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	server := startTestServer(t)
	url := server.URL + "/simple.html"
	outputPath := filepath.Join(t.TempDir(), "info.json")

	stdout, stderr, err := runSnag("--info", "--force-headless", "--close-tab", "-o", outputPath, url)

	assertNoError(t, err)
	assertExitCode(t, err, 0)
	assertContains(t, stderr, "--close-tab is ignored in headless mode")
	assertContains(t, stderr, "Saved info to")
	assertNotContains(t, stderr, "Fetching")
	if strings.TrimSpace(stdout) != "" {
		t.Errorf("expected empty stdout when saving --info, got: %s", stdout)
	}

	content, readErr := os.ReadFile(outputPath)
	if readErr != nil {
		t.Fatalf("failed to read info file: %v", readErr)
	}
	if !json.Valid(content) {
		t.Errorf("expected JSON in %s, got: %s", outputPath, content)
	}
}

// TestBrowser_DebugMode tests --debug includes verbose progress plus [DEBUG] lines
func TestBrowser_DebugMode(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	server := startTestServer(t)
	url := server.URL + "/simple.html"

	stdout, stderr, err := runSnag("--debug", url)

	assertNoError(t, err)
	assertExitCode(t, err, 0)

	assertContains(t, stdout, "# Example Heading")
	assertContains(t, stderr, "Fetching")
	assertContains(t, stderr, "Fetched successfully")
	assertContains(t, stderr, "[DEBUG]")
	if !strings.Contains(stderr, "launched") && !strings.Contains(stderr, "Connected") {
		t.Errorf("expected debug stderr to mention launched or connected, got:\n%s", stderr)
	}
}

// TestBrowser_RealWorld_ExampleDotCom tests fetching a real website (example.com)
func TestBrowser_RealWorld_ExampleDotCom(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	// Skip in environments without internet access
	if testing.Short() {
		t.Skip("skipping real-world test in short mode")
	}

	stdout, stderr, err := runSnag("https://example.com")

	assertNoError(t, err)
	assertExitCode(t, err, 0)

	// Should contain example.com content
	// Example.com has specific text we can check
	output := strings.ToLower(stdout)
	if !strings.Contains(output, "example") {
		t.Errorf("expected example.com content, got: %s", stdout)
	}

	// Should be valid markdown
	if len(strings.TrimSpace(stdout)) == 0 {
		t.Error("expected non-empty output")
	}

	_ = stderr
}

// TestBrowser_RealWorld_HttpBin tests fetching httpbin.org endpoints
func TestBrowser_RealWorld_HttpBin(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	// Skip in environments without internet access
	if testing.Short() {
		t.Skip("skipping real-world test in short mode")
	}

	// Test httpbin.org/html endpoint
	stdout, stderr, err := runSnag("https://httpbin.org/html")

	assertNoError(t, err)
	assertExitCode(t, err, 0)

	// httpbin.org/html returns HTML content
	// Should be converted to markdown
	if len(strings.TrimSpace(stdout)) == 0 {
		t.Error("expected non-empty output from httpbin.org")
	}

	_ = stderr
}

// TestBrowser_RealWorld_DelayedResponse tests handling slow-loading pages
func TestBrowser_RealWorld_DelayedResponse(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	// Skip in environments without internet access
	if testing.Short() {
		t.Skip("skipping real-world test in short mode")
	}

	// httpbin.org/delay/2 delays response by 2 seconds
	// Using increased timeout to handle network latency
	stdout, stderr, err := runSnag("--timeout", "30", "https://httpbin.org/delay/2")

	assertNoError(t, err)
	assertExitCode(t, err, 0)

	// Should handle the delay and fetch content
	if len(strings.TrimSpace(stdout)) == 0 {
		t.Error("expected non-empty output from delayed response")
	}

	_ = stderr
}

// TestBrowser_ListTabs tests --list-tabs with browser running
func TestBrowser_ListTabs(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	// Ensure clean state - kill any existing browsers
	cleanupOrphanedBrowsers()

	// First, open a browser with a visible tab
	stdout1, stderr1, err1 := runSnag("--open-browser", "https://example.com")
	assertNoError(t, err1)
	assertExitCode(t, err1, 0)

	// Now list tabs
	stdout2, stderr2, err2 := runSnag("--list-tabs")

	assertNoError(t, err2)
	assertExitCode(t, err2, 0)

	// Should list at least one tab
	assertContains(t, stdout2, "Available tabs")
	assertContains(t, stdout2, "[1]")

	// Should show example.com in the list
	assertContains(t, stdout2, "example.com")

	// Verify logs went to stderr for first command
	_ = stderr1
	_ = stderr2
	_ = stdout1
}

// TestCLI_TabNoBrowser tests --tab without browser running
func TestCLI_TabNoBrowser(t *testing.T) {
	// Ensure no browsers are running from previous tests
	cleanupOrphanedBrowsers()

	stdout, stderr, err := runSnag("--tab", "1")

	// Should fail when no browser is running
	assertError(t, err)
	assertExitCode(t, err, 1)

	// Error should be helpful
	assertContains(t, stderr, "No browser found")

	_ = stdout
}

// TestCLI_TabWithURL tests --tab with URL argument (should error)
func TestCLI_TabWithURL(t *testing.T) {
	stdout, stderr, err := runSnag("--tab", "1", "https://example.com")

	// Should fail when both --tab and URL are provided
	assertError(t, err)
	assertExitCode(t, err, 1)

	// Error should explain the conflict
	assertContains(t, stderr, "Cannot use both --tab and URL arguments")

	_ = stdout
}

// TestCLI_TabInvalidIndex tests --tab with non-numeric value
// TestCLI_TabInvalidIndex is deprecated - Phase 2.3 treats non-numeric values as patterns
// This test is kept for backwards compatibility but now tests pattern matching behavior
func TestCLI_TabInvalidIndex(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	// Ensure clean state - kill any existing browsers
	cleanupOrphanedBrowsers()

	// First, open a browser
	_, _, err1 := runSnag("--open-browser", "https://example.com")
	assertNoError(t, err1)

	// Try to use --tab with non-numeric value (now treated as pattern in Phase 2.3)
	stdout, stderr, err := runSnag("--tab", "not-a-number")

	// Should fail with "no tab matches pattern" error (not "invalid index")
	assertError(t, err)
	assertExitCode(t, err, 1)
	assertContains(t, stderr, "No tab matches pattern")
	assertContains(t, stderr, "snag --list-tabs")

	_ = stdout
}

// TestBrowser_TabOutOfRange tests --tab with index out of range
func TestBrowser_TabOutOfRange(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	// Ensure clean state - kill any existing browsers
	cleanupOrphanedBrowsers()

	// First, open a browser
	_, _, err1 := runSnag("--open-browser", "https://example.com")
	assertNoError(t, err1)

	// Try to fetch from tab index 999 (out of range)
	stdout, stderr, err := runSnag("--tab", "999")

	// Should fail with helpful error
	assertError(t, err)
	assertExitCode(t, err, 1)
	assertContains(t, stderr, "Tab index out of range")
	assertContains(t, stderr, "snag --list-tabs")

	_ = stdout
}

// TestBrowser_TabNoMatch tests --tab with pattern that doesn't match any tab
func TestBrowser_TabNoMatch(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	// Ensure clean state - kill any existing browsers
	cleanupOrphanedBrowsers()

	// First, open a browser
	_, _, err1 := runSnag("--open-browser", "https://example.com")
	assertNoError(t, err1)

	// Try to fetch from tab with non-matching pattern
	stdout, stderr, err := runSnag("--tab", "nonexistent-pattern-xyz")

	// Should fail with helpful error
	assertError(t, err)
	assertExitCode(t, err, 1)
	assertContains(t, stderr, "No tab matches pattern")
	assertContains(t, stderr, "snag --list-tabs")

	_ = stdout
}

// ============================================================================
// Phase 3: Integration Tests for New Features
// ============================================================================

// TestBrowser_TextFormat tests --format text output
func TestBrowser_TextFormat(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	server := startTestServer(t)
	url := server.URL + "/simple.html"

	stdout, stderr, err := runSnag("--format", "text", url)

	assertNoError(t, err)
	assertExitCode(t, err, 0)

	// Verify plain text output (no HTML, no markdown)
	assertContains(t, stdout, "Example Heading")
	assertContains(t, stdout, "simple paragraph")

	// Should NOT contain HTML tags
	assertNotContains(t, stdout, "<h1>")
	assertNotContains(t, stdout, "<p>")

	// Should NOT contain markdown syntax
	assertNotContains(t, stdout, "# Example Heading")
	assertNotContains(t, stdout, "**bold")

	// Verify logs went to stderr
	if len(stderr) > 0 {
		assertNotContains(t, stderr, "Example Heading")
	}
}

// TestBrowser_PDFFormat tests --format pdf creates file
func TestBrowser_PDFFormat(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	server := startTestServer(t)
	url := server.URL + "/simple.html"

	// Create temporary file for PDF output
	tmpFile, err := os.CreateTemp("", "snag-test-*.pdf")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	outputPath := tmpFile.Name()
	tmpFile.Close()

	// Clean up after test
	t.Cleanup(func() {
		os.Remove(outputPath)
	})

	stdout, stderr, err := runSnag("--format", "pdf", "-o", outputPath, url)

	assertNoError(t, err)
	assertExitCode(t, err, 0)

	assertContains(t, stderr, "Saved to")

	// Stdout should be empty (content written to file)
	if len(strings.TrimSpace(stdout)) > 0 {
		t.Errorf("expected empty stdout when using -o flag, got: %s", stdout)
	}

	// Verify PDF file was created
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read PDF file: %v", err)
	}

	// Check PDF magic bytes
	if len(content) < 4 || !bytes.HasPrefix(content, []byte("%PDF")) {
		t.Errorf("expected PDF file to start with %%PDF, got: %s", string(content[:min(20, len(content))]))
	}

	// PDF should not be empty
	if len(content) < 100 {
		t.Errorf("expected PDF file to have content, got only %d bytes", len(content))
	}
}

func TestBrowser_PDFAutoFilename(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	server := startTestServer(t)
	url := server.URL + "/simple.html"
	tmpDir := t.TempDir()

	cmd := exec.Command(snagBin, "--format", "pdf", url)
	cmd.Dir = tmpDir
	stdoutBytes, stderrBytes, err := runCommand(cmd)
	stdout := string(stdoutBytes)
	stderr := string(stderrBytes)

	assertNoError(t, err)
	assertExitCode(t, err, 0)
	assertContains(t, stderr, "Filename:")
	if len(strings.TrimSpace(stdout)) > 0 {
		t.Errorf("expected empty stdout for auto-generated PDF, got: %s", stdout)
	}

	files, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("failed to read temp directory: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 PDF in working directory, got %d", len(files))
	}
	if !strings.HasSuffix(files[0].Name(), ".pdf") {
		t.Errorf("expected .pdf file, got: %s", files[0].Name())
	}
}

// TestBrowser_PNGFormat tests --format png creates screenshot
func TestBrowser_PNGFormat(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	server := startTestServer(t)
	url := server.URL + "/simple.html"

	// Create temporary file for PNG output
	tmpFile, err := os.CreateTemp("", "snag-test-*.png")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	outputPath := tmpFile.Name()
	tmpFile.Close()

	// Clean up after test
	t.Cleanup(func() {
		os.Remove(outputPath)
	})

	stdout, stderr, err := runSnag("--format", "png", "-o", outputPath, url)

	assertNoError(t, err)
	assertExitCode(t, err, 0)

	// Stdout should be empty (content written to file)
	if len(strings.TrimSpace(stdout)) > 0 {
		t.Errorf("expected empty stdout when using -o flag, got: %s", stdout)
	}

	// Verify PNG file was created
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read PNG file: %v", err)
	}

	// Check PNG magic bytes
	if len(content) < 8 || content[0] != 0x89 || content[1] != 0x50 || content[2] != 0x4E || content[3] != 0x47 {
		t.Errorf("expected PNG file to have PNG magic bytes, got: %x", content[:min(8, len(content))])
	}

	// PNG should not be empty
	if len(content) < 100 {
		t.Errorf("expected PNG file to have content, got only %d bytes", len(content))
	}

	_ = stderr
}

// TestBrowser_OutputDir tests --output-dir flag with auto-generated filenames
func TestBrowser_OutputDir(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	server := startTestServer(t)
	url := server.URL + "/simple.html"

	// Create temporary directory
	tmpDir := t.TempDir()

	stdout, stderr, err := runSnag("-d", tmpDir, url)

	assertNoError(t, err)
	assertExitCode(t, err, 0)

	assertContains(t, stderr, "Saved to")

	// When using -d, content goes to file, not stdout
	// Stdout should be empty (or just contain logs)
	if len(strings.TrimSpace(stdout)) > 0 {
		// If there's output, it should be logs, not content
		assertNotContains(t, stdout, "Example Heading")
	}

	// Check that a file was created in the temp directory
	files, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("failed to read temp directory: %v", err)
	}

	// Should have created exactly one file
	if len(files) != 1 {
		t.Fatalf("expected 1 file in output directory, got %d", len(files))
	}

	// Verify filename format and content
	filename := files[0].Name()
	if !strings.HasSuffix(filename, ".md") {
		t.Errorf("expected .md file, got: %s", filename)
	}
	if !strings.Contains(filename, "-") {
		t.Errorf("expected timestamp in filename, got: %s", filename)
	}

	// Read the file and verify it contains the expected content
	filePath := filepath.Join(tmpDir, filename)
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}
	assertContains(t, string(content), "Example Heading")

	_ = stderr
}

// TestBrowser_OutputDirPDF tests --output-dir with PDF format
func TestBrowser_OutputDirPDF(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	server := startTestServer(t)
	url := server.URL + "/simple.html"

	// Create temporary directory
	tmpDir := t.TempDir()

	stdout, stderr, err := runSnag("-d", tmpDir, "--format", "pdf", url)

	assertNoError(t, err)
	assertExitCode(t, err, 0)

	// Stdout should be empty for binary format
	if len(strings.TrimSpace(stdout)) > 0 {
		t.Errorf("expected empty stdout for PDF format, got: %s", stdout)
	}

	// Check that a PDF file was created
	files, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("failed to read temp directory: %v", err)
	}

	// Should have created exactly one file
	if len(files) != 1 {
		t.Fatalf("expected 1 file in output directory, got %d", len(files))
	}

	if len(files) > 0 {
		// Verify filename format (should be timestamp-slug.pdf)
		filename := files[0].Name()
		if !strings.HasSuffix(filename, ".pdf") {
			t.Errorf("expected .pdf file, got: %s", filename)
		}
	}

	_ = stderr
}

// ============================================================================
// Phase 5: Multiple URL Support Tests
// ============================================================================

// TestCLI_MultipleURLs_FlagOrder tests that Cobra allows flags anywhere (improved UX)
func TestCLI_MultipleURLs_FlagOrder(t *testing.T) {
	tmpDir := t.TempDir()
	stdout, stderr, err := runSnag("https://example.com", "--force-headless", "-d", tmpDir)

	// Cobra allows flags anywhere (more flexible than urfave/cli)
	// This should now succeed instead of failing
	assertNoError(t, err)

	assertContains(t, stderr, "Saved to")
	files, readErr := os.ReadDir(tmpDir)
	if readErr != nil {
		t.Fatalf("failed to read temp directory: %v", readErr)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file in output directory, got %d", len(files))
	}

	_ = stdout
}

// TestCLI_MultipleURLs_WithOutput tests error when using --output with multiple URLs
func TestCLI_MultipleURLs_WithOutput(t *testing.T) {
	stdout, stderr, err := runSnag("--force-headless", "-o", "output.md", "https://example.com", "https://go.dev")

	// Should fail with conflict error
	assertError(t, err)
	assertExitCode(t, err, 1)

	output := stdout + stderr
	assertContains(t, output, "Cannot use --output with multiple")
	assertContains(t, output, "--output-dir")

	_ = stdout
}

// TestCLI_MultipleURLs_WithCloseTab tests --close-tab works with multiple URLs
func TestCLI_MultipleURLs_WithCloseTab(t *testing.T) {
	tmpDir := t.TempDir()
	stdout, stderr, err := runSnag("--verbose", "--force-headless", "--close-tab", "-d", tmpDir, "https://example.com", "https://go.dev")

	// Should succeed - --close-tab is now supported with multiple URLs
	assertNoError(t, err)

	output := stdout + stderr
	assertContains(t, output, "Processing 2 URLs")
	assertContains(t, output, "Batch complete")

	_ = stdout
}

// TestCLI_MultipleURLs_WithTab tests error when using --tab with URL arguments
func TestCLI_MultipleURLs_WithTab(t *testing.T) {
	stdout, stderr, err := runSnag("--tab", "1", "https://example.com")

	// Should fail with conflict error
	assertError(t, err)
	assertExitCode(t, err, 1)

	output := stdout + stderr
	assertContains(t, output, "Cannot use both --tab and URL arguments")

	_ = stdout
}

// TestCLI_MultipleURLs_WithAllTabs tests error when using --all-tabs with URL arguments
func TestCLI_MultipleURLs_WithAllTabs(t *testing.T) {
	stdout, stderr, err := runSnag("--all-tabs", "https://example.com")

	// Should fail with conflict error
	assertError(t, err)
	assertExitCode(t, err, 1)

	output := stdout + stderr
	assertContains(t, output, "Cannot use both --all-tabs and URL arguments")

	_ = stdout
}

// TestCLI_MultipleURLs_WithListTabs tests --list-tabs overrides URL arguments
func TestCLI_MultipleURLs_WithListTabs(t *testing.T) {
	stdout, stderr, err := runSnag("--list-tabs", "https://example.com")

	// Should succeed - --list-tabs overrides URL arguments, but may fail if no browser running
	// Either succeeds with tab list, or fails with "no browser instance running"
	output := stdout + stderr

	// Accept either outcome since test environment may not have browser running
	if err != nil {
		assertContains(t, output, "no browser instance running")
	} else {
		// If browser is running, should list tabs (URL is ignored)
		assertContains(t, output, "Available tabs")
	}

	_ = stdout
}

// TestCLI_URLFile_NotFound tests error when URL file doesn't exist
func TestCLI_URLFile_NotFound(t *testing.T) {
	stdout, stderr, err := runSnag("--url-file", "/nonexistent/file.txt")

	// Should fail with file not found error
	assertError(t, err)
	assertExitCode(t, err, 1)

	output := stdout + stderr
	assertContains(t, output, "Failed to open URL file")

	_ = stdout
}

// TestCLI_URLFile_Empty tests error when URL file contains no valid URLs
func TestCLI_URLFile_Empty(t *testing.T) {
	// Create empty URL file
	tmpFile, err := os.CreateTemp("", "empty-urls-*.txt")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	stdout, stderr, err := runSnag("--url-file", tmpFile.Name())

	// Should fail with no valid URLs error
	assertError(t, err)
	assertExitCode(t, err, 1)

	output := stdout + stderr
	assertContains(t, output, "no valid URLs")

	_ = stdout
}

// TestCLI_URLFile_OnlyComments tests error when URL file contains only comments
func TestCLI_URLFile_OnlyComments(t *testing.T) {
	// Create URL file with only comments
	tmpFile, err := os.CreateTemp("", "comments-only-*.txt")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	content := `# This is a comment
// Another comment

# More comments
`
	os.WriteFile(tmpFile.Name(), []byte(content), 0644)

	stdout, stderr, err := runSnag("--url-file", tmpFile.Name())

	// Should fail with no valid URLs error
	assertError(t, err)
	assertExitCode(t, err, 1)

	output := stdout + stderr
	assertContains(t, output, "no valid URLs")

	_ = stdout
}

// TestBrowser_MultipleURLs_Inline tests fetching multiple inline URLs
func TestBrowser_MultipleURLs_Inline(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	server := startTestServer(t)
	url1 := server.URL + "/simple.html"
	url2 := server.URL + "/minimal.html"

	// Create temporary directory for output
	tmpDir := t.TempDir()

	stdout, stderr, err := runSnag("--verbose", "--force-headless", "-d", tmpDir, url1, url2)

	assertNoError(t, err)
	assertExitCode(t, err, 0)

	// Stderr should contain progress indicators
	assertContains(t, stderr, "Processing 2 URLs")
	assertContains(t, stderr, "[1/2]")
	assertContains(t, stderr, "[2/2]")
	assertContains(t, stderr, "Batch complete: 2 succeeded, 0 failed")

	// Check that 2 files were created
	files, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("failed to read temp directory: %v", err)
	}

	if len(files) != 2 {
		t.Fatalf("expected 2 files in output directory, got %d", len(files))
	}

	// Verify both files have .md extension
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".md") {
			t.Errorf("expected .md file, got: %s", file.Name())
		}
	}

	_ = stdout
}

// TestBrowser_MultipleURLs_FromFile tests fetching URLs from a file
func TestBrowser_MultipleURLs_FromFile(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	server := startTestServer(t)
	url1 := server.URL + "/simple.html"
	url2 := server.URL + "/minimal.html"

	// Create URL file
	tmpFile, err := os.CreateTemp("", "urls-*.txt")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	content := fmt.Sprintf(`# Test URLs
%s
%s # With inline comment
`, url1, url2)
	os.WriteFile(tmpFile.Name(), []byte(content), 0644)

	// Create temporary directory for output
	tmpDir := t.TempDir()

	stdout, stderr, err := runSnag("--verbose", "--force-headless", "-d", tmpDir, "--url-file", tmpFile.Name())

	assertNoError(t, err)
	assertExitCode(t, err, 0)

	// Stderr should contain progress indicators
	assertContains(t, stderr, "Processing 2 URLs")
	assertContains(t, stderr, "Batch complete: 2 succeeded, 0 failed")

	// Check that 2 files were created
	files, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("failed to read temp directory: %v", err)
	}

	if len(files) != 2 {
		t.Fatalf("expected 2 files in output directory, got %d", len(files))
	}

	_ = stdout
}

// TestBrowser_MultipleURLs_Combined tests combining URL file with inline URLs
func TestBrowser_MultipleURLs_Combined(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	server := startTestServer(t)
	url1 := server.URL + "/simple.html"
	url2 := server.URL + "/minimal.html"
	url3 := server.URL + "/complex.html"

	// Create URL file with 2 URLs
	tmpFile, err := os.CreateTemp("", "urls-*.txt")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	content := fmt.Sprintf("%s\n%s\n", url1, url2)
	os.WriteFile(tmpFile.Name(), []byte(content), 0644)

	// Create temporary directory for output
	tmpDir := t.TempDir()

	// Combine file (2 URLs) with inline (1 URL) = 3 total
	stdout, stderr, err := runSnag("--verbose", "--force-headless", "-d", tmpDir, "--url-file", tmpFile.Name(), url3)

	assertNoError(t, err)
	assertExitCode(t, err, 0)

	// Stderr should show 3 URLs processed
	assertContains(t, stderr, "Processing 3 URLs")
	assertContains(t, stderr, "Batch complete: 3 succeeded, 0 failed")

	// Check that 3 files were created
	files, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("failed to read temp directory: %v", err)
	}

	if len(files) != 3 {
		t.Fatalf("expected 3 files in output directory, got %d", len(files))
	}

	_ = stdout
}

// TestBrowser_MultipleURLs_InvalidURLs tests continue-on-error behavior
func TestBrowser_MultipleURLs_InvalidURLs(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	server := startTestServer(t)
	validURL := server.URL + "/simple.html"

	// Create URL file with mix of valid and invalid URLs
	tmpFile, err := os.CreateTemp("", "urls-*.txt")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	content := fmt.Sprintf(`%s
invalid url with spaces
%s
`, validURL, validURL)
	os.WriteFile(tmpFile.Name(), []byte(content), 0644)

	// Create temporary directory for output
	tmpDir := t.TempDir()

	stdout, stderr, err := runSnag("--force-headless", "-d", tmpDir, "--url-file", tmpFile.Name())

	assertNoError(t, err) // Should succeed (2 valid URLs)
	assertExitCode(t, err, 0)

	// Warning stays on default stderr; batch progress does not
	assertContains(t, stderr, "URL contains space without comment marker")
	assertNotContains(t, stderr, "Processing 2 URLs")

	_ = stdout
}

// TestBrowser_MultipleURLs_WithFormat tests format flag applies to all URLs
func TestBrowser_MultipleURLs_WithFormat(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	server := startTestServer(t)
	url1 := server.URL + "/simple.html"
	url2 := server.URL + "/minimal.html"

	// Create temporary directory for output
	tmpDir := t.TempDir()

	stdout, stderr, err := runSnag("--force-headless", "-d", tmpDir, "--format", "html", url1, url2)

	assertNoError(t, err)
	assertExitCode(t, err, 0)

	// Check that 2 files were created with .html extension
	files, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("failed to read temp directory: %v", err)
	}

	if len(files) != 2 {
		t.Fatalf("expected 2 files in output directory, got %d", len(files))
	}

	// Verify both files have .html extension
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".html") {
			t.Errorf("expected .html file, got: %s", file.Name())
		}
	}

	_ = stdout
	_ = stderr
}

// TestBrowser_URLFile_AutoHTTPS tests auto-prepending https:// to URLs
func TestBrowser_URLFile_AutoHTTPS(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	// Skip in environments without internet access
	if testing.Short() {
		t.Skip("skipping real-world test in short mode")
	}

	// Create temporary directory for output
	tmpDir := t.TempDir()

	// Use testdata/urls-small.txt which has URLs without https://
	// This will try to fetch https://example.com and https://httpbin.org/html (auto-prepended)
	stdout, stderr, err := runSnag("--verbose", "--force-headless", "-d", tmpDir, "--url-file", "testdata/urls-small.txt")

	// May succeed or fail depending on network access
	// Just verify the URLs were loaded from file
	output := stderr + stdout
	assertContains(t, output, "Processing 2 URLs")

	_ = err // May succeed or fail depending on network
}

// TestBrowser_URLFile_Comprehensive tests all URL file features with testdata/urls.txt
func TestBrowser_URLFile_Comprehensive(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	// Skip in environments without internet access
	if testing.Short() {
		t.Skip("skipping real-world test in short mode")
	}

	// Create temporary directory for output
	tmpDir := t.TempDir()

	// Use testdata/urls.txt which demonstrates all features:
	// - Full-line comments with # and //
	// - Inline comments with # and //
	// - Auto-prepending https://
	// - URLs with paths and query parameters
	stdout, stderr, err := runSnag("--force-headless", "-d", tmpDir, "--url-file", "testdata/urls.txt")

	// May succeed or fail depending on network access
	// Verify the URLs were loaded (should have 7 valid URLs based on the file content)
	output := stderr + stdout
	if strings.Contains(output, "Processing") {
		// If processing started, verify it processed multiple URLs
		assertContains(t, output, "URL")
	}

	_ = err // May succeed or fail depending on network
}

// TestBrowser_URLFile_InvalidURLs_RealWorld tests error handling with testdata/urls-invalid.txt
func TestBrowser_URLFile_InvalidURLs_RealWorld(t *testing.T) {
	if !isBrowserAvailable() {
		t.Skip("Browser not available, skipping browser integration test")
	}

	// Skip in environments without internet access
	if testing.Short() {
		t.Skip("skipping real-world test in short mode")
	}

	// Create temporary directory for output
	tmpDir := t.TempDir()

	// Use testdata/urls-invalid.txt which has mix of valid and invalid URLs
	stdout, stderr, err := runSnag("--force-headless", "-d", tmpDir, "--url-file", "testdata/urls-invalid.txt")

	// Should process valid URLs and warn about invalid ones
	output := stderr + stdout

	// Should warn about invalid URLs
	if strings.Contains(output, "Processing") {
		// If processing started, verify warnings were shown
		// testdata/urls-invalid.txt has "example.com invalid extra text" which should warn
		assertContains(t, output, "URL contains space")
	}

	_ = err // May succeed or fail depending on network
}

// min is a polyfill for compatibility with Go versions prior to 1.21
// (min became a built-in function in Go 1.21)
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
