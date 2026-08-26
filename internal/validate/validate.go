// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package validate

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/p3bot/snag/internal/format"
	"github.com/p3bot/snag/internal/logger"
	"github.com/p3bot/snag/internal/output"
)

func URL(urlStr string) (string, error) {
	if !strings.Contains(urlStr, "://") {
		urlStr = "https://" + urlStr
		logger.Verbose("No scheme provided, using: %s", urlStr)
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		logger.Error("Invalid URL: %s", urlStr)
		logger.Debug("URL parsing error: %v", err)
		logger.ErrorWithSuggestion(
			fmt.Sprintf("URL parsing failed: %v", err),
			"snag https://example.com",
		)
		return "", ErrInvalidURL
	}
	logger.Debug("Parsed URL - scheme: %s, host: %s, path: %s", parsedURL.Scheme, parsedURL.Host, parsedURL.Path)

	validSchemes := map[string]bool{
		"http":  true,
		"https": true,
		"file":  true,
	}

	if !validSchemes[parsedURL.Scheme] {
		logger.Error("Unsupported URL scheme: %s", parsedURL.Scheme)
		logger.ErrorWithSuggestion(
			"URL must use http://, https://, or file://",
			"snag https://example.com",
		)
		return "", ErrInvalidURL
	}

	if parsedURL.Scheme != "file" && parsedURL.Host == "" {
		logger.Error("Invalid URL: missing host")
		logger.ErrorWithSuggestion(
			"URL must include a hostname",
			"snag https://example.com",
		)
		return "", ErrInvalidURL
	}

	return urlStr, nil
}

func IsNonFetchableURL(urlStr string) bool {
	nonFetchablePrefixes := []string{
		"chrome://",
		"about:",
		"devtools://",
		"chrome-extension://",
		"edge://",
		"brave://",
	}

	urlLower := strings.ToLower(urlStr)
	for _, prefix := range nonFetchablePrefixes {
		if strings.HasPrefix(urlLower, prefix) {
			return true
		}
	}
	return false
}

func Timeout(timeout int) error {
	if timeout <= 0 {
		logger.Error("Invalid timeout: %d", timeout)
		logger.ErrorWithSuggestion(
			"Timeout must be a positive number of seconds",
			"snag <url> --timeout 30",
		)
		return fmt.Errorf("invalid timeout: %d", timeout)
	}
	return nil
}

func Port(port int) error {
	if port < 1024 || port > 65535 {
		logger.Error("Invalid port: %d", port)
		logger.ErrorWithSuggestion(
			"Port must be between 1024 and 65535",
			"snag <url> --port 9222",
		)
		return fmt.Errorf("invalid port: %d", port)
	}
	return nil
}

func OutputPath(path string) error {
	if path == "" {
		logger.Error("Output file path cannot be empty")
		logger.ErrorWithSuggestion(
			"Provide a valid file path",
			"snag <url> -o /path/to/output.md",
		)
		return fmt.Errorf("output file path cannot be empty")
	}

	if info, err := os.Stat(path); err == nil && info.IsDir() {
		logger.Error("Output path is a directory, not a file: %s", path)
		logger.ErrorWithSuggestion(
			"Specify a file path, not a directory",
			"snag <url> -o /path/to/file.md",
		)
		return fmt.Errorf("output path is a directory, not a file: %s", path)
	}

	if info, err := os.Stat(path); err == nil {
		if info.Mode().Perm()&0200 == 0 {
			logger.Error("Cannot write to read-only file: %s", path)
			logger.ErrorWithSuggestion(
				"Make file writable or choose different path",
				"chmod u+w "+path,
			)
			return fmt.Errorf("cannot write to read-only file: %s", path)
		}
	}

	dir := filepath.Dir(path)

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		logger.Error("Output directory does not exist: %s", dir)
		logger.ErrorWithSuggestion(
			fmt.Sprintf("Directory '%s' not found", dir),
			"snag <url> -o /path/to/existing/dir/output.md",
		)
		return fmt.Errorf("output directory does not exist: %s", dir)
	}

	f, err := os.CreateTemp(dir, ".snag-write-test-*")
	if err != nil {
		logger.Error("Cannot write to output directory: %s", dir)
		logger.ErrorWithSuggestion(
			"Permission denied or directory not writable",
			"snag <url> -o /path/to/writable/dir/output.md",
		)
		return fmt.Errorf("cannot write to output directory: %s", dir)
	}
	testFile := f.Name()
	f.Close()
	os.Remove(testFile)

	return nil
}

func NormalizeFormat(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ToLower(name)

	switch name {
	case "markdown":
		return format.Markdown
	case "txt":
		return format.Text
	default:
		return name
	}
}

func Format(name string) error {
	if name == "" {
		logger.Error("Format cannot be empty")
		logger.ErrorWithSuggestion(
			"Format must be specified",
			fmt.Sprintf("snag <url> --format %s", format.Markdown),
		)
		return fmt.Errorf("format cannot be empty")
	}

	validFormats := map[string]bool{
		format.Markdown: true,
		format.HTML:     true,
		format.Text:     true,
		format.PDF:      true,
		format.PNG:      true,
	}

	if !validFormats[name] {
		logger.Error("Invalid format '%s'. Supported: md, html, text, pdf, png", name)
		logger.ErrorWithSuggestion(
			"Choose a valid format",
			fmt.Sprintf("snag <url> --format %s", format.Markdown),
		)
		return fmt.Errorf("invalid format: %s", name)
	}

	return nil
}

func CheckExtensionMismatch(outputFile string, format string) bool {
	if outputFile == "" {
		return false
	}

	ext := strings.ToLower(filepath.Ext(outputFile))
	expectedExt := strings.ToLower(output.GetFileExtension(format))

	if ext != expectedExt {
		if ext == "" {
			logger.Warning("Writing %s format to file with no extension: %s", format, outputFile)
			return true
		}
		logger.Warning("Writing %s format to file with %s extension: %s", format, ext, outputFile)
		return true
	}

	return false
}

func Directory(dir string) error {
	info, err := os.Stat(dir)
	if os.IsNotExist(err) {
		logger.Error("Directory does not exist: %s", dir)
		logger.ErrorWithSuggestion(
			fmt.Sprintf("Directory '%s' not found", dir),
			"mkdir -p /path/to/dir && snag <url> -d /path/to/dir",
		)
		return fmt.Errorf("directory does not exist: %s", dir)
	}
	if err != nil {
		return fmt.Errorf("error accessing directory: %w", err)
	}

	if !info.IsDir() {
		logger.Error("Path is not a directory: %s", dir)
		return fmt.Errorf("not a directory: %s", dir)
	}

	testFile, err := os.CreateTemp(dir, ".snag-write-test-*")
	if err != nil {
		logger.Error("Directory not writable: %s", dir)
		logger.ErrorWithSuggestion(
			"Permission denied or directory not writable",
			"chmod u+w /path/to/dir",
		)
		return fmt.Errorf("directory not writable: %s", dir)
	}
	testFilePath := testFile.Name()
	testFile.Close()
	os.Remove(testFilePath)

	return nil
}

func WaitFor(selector string, flagSet bool) string {
	selector = strings.TrimSpace(selector)

	if selector == "" {
		if flagSet {
			logger.Warning("--wait-for is empty, ignoring")
		}
		return ""
	}

	return selector
}

func UserAgent(ua string, flagSet bool) string {
	ua = strings.TrimSpace(ua)

	if ua == "" {
		if flagSet {
			logger.Warning("--user-agent is empty, using default user agent")
		}
		return ""
	}

	ua = strings.ReplaceAll(ua, "\n", " ")
	ua = strings.ReplaceAll(ua, "\r", " ")

	return ua
}

func UserDataDir(path string) (string, error) {
	path = strings.TrimSpace(path)

	if path == "" {
		logger.Warning("--user-data-dir is empty, using default profile")
		return "", nil
	}

	if strings.HasPrefix(path, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		if path == "~" {
			path = homeDir
		} else if strings.HasPrefix(path, "~/") {
			path = filepath.Join(homeDir, path[2:])
		}
	}

	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		logger.Verbose("Creating user data directory: %s", path)
		if err := os.MkdirAll(path, 0755); err != nil {
			logger.Error("Failed to create user data directory: %s", path)
			logger.ErrorWithSuggestion(
				"Cannot create user data directory",
				fmt.Sprintf("mkdir -p %s", path),
			)
			return "", fmt.Errorf("failed to create directory: %w", err)
		}
		logger.Verbose("User data directory created: %s", path)
		return path, nil
	}
	if err != nil {
		logger.Error("Error accessing user data directory: %s", path)
		return "", fmt.Errorf("error accessing directory: %w", err)
	}

	if !info.IsDir() {
		logger.Error("Path is not a directory: %s", path)
		logger.ErrorWithSuggestion(
			"User data directory must be a directory, not a file",
			"snag --user-data-dir /path/to/directory <url>",
		)
		return "", fmt.Errorf("path is not a directory: %s", path)
	}

	testFile, err := os.CreateTemp(path, ".snag-permission-test-*")
	if err != nil {
		logger.Error("Permission denied accessing user data directory: %s", path)
		logger.ErrorWithSuggestion(
			"Cannot read/write to directory",
			fmt.Sprintf("chmod u+rw %s", path),
		)
		return "", fmt.Errorf("permission denied accessing user data directory: %s", path)
	}
	testFilePath := testFile.Name()
	testFile.Close()
	os.Remove(testFilePath)

	return path, nil
}
