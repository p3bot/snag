// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package fetch

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/p3bot/snag/internal/browser"
)

// PageInfo is JSON metadata about a web page.
type PageInfo struct {
	Title     string `json:"title"`
	URL       string `json:"url"`
	Domain    string `json:"domain"`
	Slug      string `json:"slug"`
	Timestamp string `json:"timestamp"`
}

func ExtractPageInfo(page *browser.Page) (*PageInfo, error) {
	if page == nil {
		return nil, fmt.Errorf("cannot extract info: page is nil")
	}

	meta, err := page.Meta()
	if err != nil {
		return nil, fmt.Errorf("failed to get page info: %w", err)
	}

	return &PageInfo{
		Title:     meta.Title,
		URL:       meta.URL,
		Domain:    extractDomain(meta.URL),
		Timestamp: time.Now().Format(time.RFC3339),
	}, nil
}

func extractDomain(urlStr string) string {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}

	// Hostname() strips the port and IPv6 brackets. LastIndex(":") would
	// cut inside an IPv6 literal (https://[::1]:8080 → "[::1").
	return strings.TrimPrefix(parsedURL.Hostname(), "www.")
}
