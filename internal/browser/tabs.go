// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package browser

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/p3bot/snag/internal/logger"

	"github.com/go-rod/rod"
)

type TabInfo struct {
	Index int
	URL   string
	Title string
	ID    string
}

type pageWithInfo struct {
	page  *rod.Page
	url   string
	title string
	id    string
}

func browserCtxErr(b *rod.Browser) error {
	if b == nil {
		return nil
	}
	ctx := b.GetContext()
	if ctx == nil {
		return nil
	}
	return ctx.Err()
}

func (bm *BrowserManager) getSortedPagesWithInfo() ([]pageWithInfo, error) {
	if bm.browser == nil {
		return nil, ErrNoBrowserRunning
	}
	if err := browserCtxErr(bm.browser); err != nil {
		return nil, err
	}

	pages, err := bm.browser.Pages()
	if err != nil {
		if e := browserCtxErr(bm.browser); e != nil {
			return nil, e
		}
		return nil, fmt.Errorf("failed to get pages: %w", err)
	}

	pagesWithInfo := make([]pageWithInfo, 0, len(pages))
	for i, page := range pages {
		info, err := page.Info()
		if err != nil {
			if e := browserCtxErr(bm.browser); e != nil {
				return nil, e
			}
			logger.Warning("Failed to get info for tab at position %d (will be excluded from list): %v", i+1, err)
			logger.Debug("Tab page object: %+v", page)
			continue
		}
		pagesWithInfo = append(pagesWithInfo, pageWithInfo{
			page:  page,
			url:   info.URL,
			title: info.Title,
			id:    string(page.TargetID),
		})
	}

	if len(pagesWithInfo) < len(pages) {
		excluded := len(pages) - len(pagesWithInfo)
		logger.Warning("Excluded %d tab(s) due to inaccessible page info", excluded)
	}

	sort.Slice(pagesWithInfo, func(i, j int) bool {
		if pagesWithInfo[i].url != pagesWithInfo[j].url {
			return pagesWithInfo[i].url < pagesWithInfo[j].url
		}
		if pagesWithInfo[i].title != pagesWithInfo[j].title {
			return pagesWithInfo[i].title < pagesWithInfo[j].title
		}
		return pagesWithInfo[i].id < pagesWithInfo[j].id
	})

	return pagesWithInfo, nil
}

func pagesFromInfo(pagesWithInfo []pageWithInfo) []*Page {
	pages := make([]*Page, len(pagesWithInfo))
	for i, pwi := range pagesWithInfo {
		pages[i] = wrapPage(pwi.page)
	}
	return pages
}

// GetPages returns every tab in the same sorted order as ListTabs, from one listing.
func (bm *BrowserManager) GetPages() ([]*Page, error) {
	pagesWithInfo, err := bm.getSortedPagesWithInfo()
	if err != nil {
		return nil, err
	}
	return pagesFromInfo(pagesWithInfo), nil
}

func (bm *BrowserManager) ListTabs() ([]TabInfo, error) {
	pagesWithInfo, err := bm.getSortedPagesWithInfo()
	if err != nil {
		return nil, err
	}

	tabs := make([]TabInfo, len(pagesWithInfo))
	for i, pwi := range pagesWithInfo {
		tabs[i] = TabInfo{
			Index: i + 1,
			URL:   pwi.url,
			Title: pwi.title,
			ID:    pwi.id,
		}
	}

	return tabs, nil
}

func (bm *BrowserManager) GetTabByIndex(index int) (*Page, error) {
	pagesWithInfo, err := bm.getSortedPagesWithInfo()
	if err != nil {
		return nil, err
	}

	if index < 1 || index > len(pagesWithInfo) {
		return nil, fmt.Errorf("%w: tab index %d (valid range: 1-%d)", ErrTabIndexInvalid, index, len(pagesWithInfo))
	}

	arrayIndex := index - 1

	logger.Verbose("Selected tab [%d] from sorted order: %s", index, pagesWithInfo[arrayIndex].url)

	return wrapPage(pagesWithInfo[arrayIndex].page), nil
}

func (bm *BrowserManager) GetTabsByPattern(pattern string) ([]*Page, error) {
	pagesWithInfo, err := bm.getSortedPagesWithInfo()
	if err != nil {
		return nil, err
	}

	if len(pagesWithInfo) == 0 {
		return nil, fmt.Errorf("%w: '%s' (no tabs open)", ErrNoTabMatch, pattern)
	}

	logger.Debug("Matching pattern '%s' against %d tabs", pattern, len(pagesWithInfo))
	patternLower := strings.ToLower(pattern)

	var exactMatches []*Page
	for i, pwi := range pagesWithInfo {
		if strings.EqualFold(pwi.url, pattern) {
			logger.Verbose("Matched tab [%d] via exact URL: %s", i+1, pwi.url)
			exactMatches = append(exactMatches, wrapPage(pwi.page))
		}
	}
	if len(exactMatches) > 0 {
		return exactMatches, nil
	}

	var substringMatches []*Page
	for i, pwi := range pagesWithInfo {
		if strings.Contains(strings.ToLower(pwi.url), patternLower) {
			logger.Verbose("Matched tab [%d] via substring: %s", i+1, pwi.url)
			substringMatches = append(substringMatches, wrapPage(pwi.page))
		}
	}
	if len(substringMatches) > 0 {
		return substringMatches, nil
	}

	re, err := regexp.Compile("(?i)" + pattern)
	if err != nil {
		logger.Debug("Pattern is not valid regex: %v", err)
		return nil, fmt.Errorf("invalid regex pattern '%s': %w", pattern, err)
	}

	var regexMatches []*Page
	for i, pwi := range pagesWithInfo {
		if re.MatchString(pwi.url) {
			logger.Verbose("Matched tab [%d] via regex: %s", i+1, pwi.url)
			regexMatches = append(regexMatches, wrapPage(pwi.page))
		}
	}
	if len(regexMatches) > 0 {
		logger.Debug("Found %d regex matches for pattern '%s'", len(regexMatches), pattern)
		return regexMatches, nil
	}

	return nil, fmt.Errorf("%w: '%s'", ErrNoTabMatch, pattern)
}

func (bm *BrowserManager) GetTabsByRange(start, end int) ([]*Page, error) {
	pagesWithInfo, err := bm.getSortedPagesWithInfo()
	if err != nil {
		return nil, err
	}

	if start < 1 {
		return nil, fmt.Errorf("tab range must start from 1 (got %d)", start)
	}
	if start > end {
		return nil, fmt.Errorf("invalid range: start must be <= end (got %d-%d)", start, end)
	}

	if start > len(pagesWithInfo) {
		return nil, fmt.Errorf("tab index %d out of range in range %d-%d (only %d tabs open)", start, start, end, len(pagesWithInfo))
	}
	if end > len(pagesWithInfo) {
		return nil, fmt.Errorf("tab index %d out of range in range %d-%d (only %d tabs open)", end, start, end, len(pagesWithInfo))
	}

	rangeTabs := pagesFromInfo(pagesWithInfo[start-1 : end])

	logger.Verbose("Selected %d tabs from sorted range [%d-%d]", len(rangeTabs), start, end)
	return rangeTabs, nil
}
