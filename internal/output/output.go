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

const MaxSlugLength = 80

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

func GetFileExtension(name string) string {
	switch name {
	case format.Markdown:
		return ".md"
	case format.HTML:
		return ".html"
	case format.Text:
		return ".txt"
	case format.PDF:
		return ".pdf"
	case format.PNG:
		return ".png"
	default:
		return ".md"
	}
}

func GenerateFilename(title string, name string, timestamp time.Time, urlStr string) string {
	timePrefix := timestamp.Format("2006-01-02-150405")

	titleSlug := SlugifyTitle(title, MaxSlugLength)
	logger.Debug("Title '%s' slugified to '%s'", title, titleSlug)

	if titleSlug == "" {
		titleSlug = GenerateURLSlug(urlStr)
		logger.Debug("Empty title slug, using URL slug: %s", titleSlug)
	}

	ext := GetFileExtension(name)

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
