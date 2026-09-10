// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package output

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/p3bot/snag/internal/format"
	"github.com/p3bot/snag/internal/logger"
)

var (
	slugNonAlphanumeric = regexp.MustCompile(`[^a-z0-9]+`)
	slugMultipleHyphens = regexp.MustCompile(`-+`)
)

const (
	MaxSlugLength   = 80
	DefaultFileMode = 0644
	bytesPerKB      = 1024.0
)

func SlugifyTitle(title string, maxLen int) string {
	slug := strings.ToLower(title)

	slug = slugNonAlphanumeric.ReplaceAllString(slug, "-")

	slug = slugMultipleHyphens.ReplaceAllString(slug, "-")

	slug = strings.Trim(slug, "-")

	if len(slug) > maxLen {
		slug = slug[:maxLen]
		slug = strings.TrimRight(slug, "-")
	}

	return slug
}

func GenerateURLSlug(urlStr string) string {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "page"
	}

	hostname := parsedURL.Hostname()
	if hostname == "" {
		return "page"
	}

	name := hostname
	if port := parsedURL.Port(); port != "" {
		name = hostname + "-" + port
	}

	slug := SlugifyTitle(name, MaxSlugLength)
	if slug == "" {
		return "page"
	}
	return slug
}

func GenerateFilename(title string, name string, timestamp time.Time, urlStr string) string {
	timePrefix := timestamp.Format("2006-01-02-150405")

	titleSlug := SlugifyTitle(title, MaxSlugLength)
	logger.Debug("Title '%s' slugified to '%s'", title, titleSlug)

	if titleSlug == "" {
		titleSlug = GenerateURLSlug(urlStr)
		logger.Debug("Empty title slug, using URL slug: %s", titleSlug)
	}

	ext := format.Extension(name)

	filename := fmt.Sprintf("%s-%s%s", timePrefix, titleSlug, ext)
	logger.Debug("Generated filename: %s", filename)

	return filename
}

func ResolveConflict(dir, filename string) (string, error) {
	fullPath := filepath.Join(dir, filename)
	logger.Debug("Checking for conflicts: %s", fullPath)

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		logger.Debug("No conflict, using original filename")
		return filename, nil
	} else if err != nil {
		return "", fmt.Errorf("failed to check file existence: %w", err)
	}

	ext := filepath.Ext(filename)
	nameWithoutExt := strings.TrimSuffix(filename, ext)

	counter := 1
	for {
		newFilename := fmt.Sprintf("%s-%d%s", nameWithoutExt, counter, ext)
		newFullPath := filepath.Join(dir, newFilename)

		_, err := os.Stat(newFullPath)
		if os.IsNotExist(err) {
			return newFilename, nil
		} else if err != nil {
			return "", fmt.Errorf("failed to check file existence: %w", err)
		}

		if counter > 10000 {
			return "", fmt.Errorf("too many conflicts for filename: %s", filename)
		}

		counter++
	}
}

// Write sends data to path, or to stdout when path is empty.
func Write(data []byte, path string) error {
	if path == "" {
		return writeStdout(data)
	}
	return writeFile(data, path)
}

func writeStdout(data []byte) error {
	logger.Verbose("Writing to stdout...")
	if _, err := os.Stdout.Write(data); err != nil {
		return fmt.Errorf("failed to write to stdout: %w", err)
	}
	logger.Debug("Wrote %d bytes to stdout", len(data))
	return nil
}

func writeFile(data []byte, filename string) error {
	logger.Verbose("Writing to file: %s", filename)
	if _, err := os.Stat(filename); err == nil {
		logger.Verbose("Overwriting existing file: %s", filename)
	}
	if err := os.WriteFile(filename, data, DefaultFileMode); err != nil {
		return fmt.Errorf("failed to write to file %s: %w", filename, err)
	}
	sizeKB := float64(len(data)) / bytesPerKB
	logger.Success("Saved to %s (%.1f KB)", filename, sizeKB)
	return nil
}
