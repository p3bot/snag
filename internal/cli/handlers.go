// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/p3bot/snag/internal/browser"
	"github.com/p3bot/snag/internal/doctor"
	"github.com/p3bot/snag/internal/logger"
	"github.com/p3bot/snag/internal/validate"
)

func snag(ctx context.Context, config *Config) error {
	bm := browser.NewBrowserManager(config.BrowserOptions())

	defer func() {
		if config.CloseTab {
			logger.Verbose("Cleanup: closing tab and browser if needed")
		}
		bm.Close()
	}()

	err := bm.Connect(ctx)
	if err != nil {
		if e := abortErr(ctx, err); e != nil {
			return e
		}
		if errors.Is(err, browser.ErrBrowserNotFound) {
			logger.Error("No Chromium-based browser found")
			logger.ErrorWithSuggestion(
				"Install Chrome, Chromium, Edge, or Brave to use snag",
				"brew install --cask google-chrome",
			)
		}
		return err
	}

	page, err := bm.NewPage()
	if err != nil {
		if e := abortErr(ctx, err); e != nil {
			return e
		}
		return err
	}

	if config.CloseTab {
		defer bm.ClosePage(page)
	}

	if err := preparePage(ctx, page, config, config.URL); err != nil {
		return err
	}
	return emitPage(ctx, page, config, config.URL, time.Now())
}

func connectToExistingBrowser(ctx context.Context, port int) (*browser.BrowserManager, error) {
	bm := browser.NewBrowserManager(browser.BrowserOptions{
		Port: port,
	})

	if err := bm.ConnectExisting(ctx); err != nil {
		if e := abortErr(ctx, err); e != nil {
			return nil, e
		}
		logger.Error("No browser found. Try running 'snag --open-browser' first")
		return nil, browser.ErrNoBrowserRunning
	}

	return bm, nil
}

func stripURLParams(url string) string {
	if idx := strings.Index(url, "?"); idx != -1 {
		url = url[:idx]
	}

	if idx := strings.Index(url, "#"); idx != -1 {
		url = url[:idx]
	}

	return url
}

func formatTabLine(index int, title, url string, maxLength int, verbose bool) string {
	if verbose {
		if title == "" {
			return fmt.Sprintf("  [%d] %s", index, url)
		}
		return fmt.Sprintf("  [%d] %s - %s", index, url, title)
	}

	cleanURL := stripURLParams(url)

	prefix := fmt.Sprintf("  [%d] ", index)
	prefixLen := len(prefix)

	const maxURLLen = MaxDisplayURLLength
	displayURL := cleanURL
	if len(displayURL) > maxURLLen {
		displayURL = cleanURL[:maxURLLen-3] + "..."
	}

	titleBudget := maxLength - prefixLen - len(displayURL)
	if title != "" {
		titleBudget -= 3
	}

	if title == "" {
		return fmt.Sprintf("%s%s", prefix, displayURL)
	}

	if len(title) > titleBudget && titleBudget > 3 {
		title = title[:titleBudget-3] + "..."
	}

	return fmt.Sprintf("%s%s (%s)", prefix, displayURL, title)
}

func displayTabList(tabs []browser.TabInfo, w io.Writer, verbose bool) {
	if len(tabs) == 0 {
		fmt.Fprintf(w, "No tabs open in browser\n")
		return
	}

	fmt.Fprintf(w, "Available tabs in browser (%d tabs, sorted by URL):\n", len(tabs))
	for _, tab := range tabs {
		line := formatTabLine(tab.Index, tab.Title, tab.URL, MaxTabLineLength, verbose)
		fmt.Fprintf(w, "%s\n", line)
	}
}

func warnProfileFlagsIfConnecting(cmd *cobra.Command) {
	if cmd.Flags().Changed("user-data-dir") {
		logger.Warning("--user-data-dir ignored when connecting to existing browser")
	}
	if cmd.Flags().Changed("temp-profile") {
		logger.Warning("--temp-profile ignored when connecting to existing browser")
	}
}

func handleListTabs(cmd *cobra.Command) error {
	warnProfileFlagsIfConnecting(cmd)
	ctx := cmd.Context()
	bm, err := connectToExistingBrowser(ctx, port)
	if err != nil {
		return err
	}

	tabs, err := bm.ListTabs()
	if err != nil {
		if e := abortErr(ctx, err); e != nil {
			return e
		}
		return err
	}

	displayTabList(tabs, os.Stdout, verbose)

	return nil
}

func handleAllTabs(cmd *cobra.Command) error {
	if cmd.Flags().Changed("user-agent") {
		logger.Warning("--user-agent is ignored with --all-tabs (cannot change existing tabs' user agents)")
	}
	warnProfileFlagsIfConnecting(cmd)
	if cmd.Flags().Changed("timeout") && waitFor == "" {
		logger.Warning("--timeout is ignored without --wait-for when using --all-tabs")
	}

	cfg, err := newEmitConfig(cmd, true)
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	bm, err := connectToExistingBrowser(ctx, cfg.Port)
	if err != nil {
		return err
	}

	pages, err := bm.GetPages()
	if err != nil {
		if e := abortErr(ctx, err); e != nil {
			return e
		}
		return err
	}

	if len(pages) == 0 {
		logger.Info("No tabs open in browser")
		return nil
	}

	logger.Verbose("Processing %d tabs...", len(pages))
	return processBatchTabs(ctx, bm, pages, cfg)
}

func handleTabFetch(cmd *cobra.Command) error {
	tabValue := strings.TrimSpace(tab)
	if tabValue == "" {
		logger.Error("Tab pattern cannot be empty")
		return fmt.Errorf("tab pattern cannot be empty")
	}

	if cmd.Flags().Changed("user-agent") {
		logger.Warning("--user-agent is ignored with --tab (cannot change existing tab's user agent)")
	}
	warnProfileFlagsIfConnecting(cmd)
	if cmd.Flags().Changed("timeout") && !cmd.Flags().Changed("wait-for") {
		logger.Warning("--timeout is ignored without --wait-for when using --tab")
	}

	cfg, err := newEmitConfig(cmd, false)
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	bm, err := connectToExistingBrowser(ctx, cfg.Port)
	if err != nil {
		return err
	}

	// Check for tab range pattern (e.g., "1-5")
	if strings.Contains(tabValue, "-") {
		parts := strings.SplitN(tabValue, "-", 2)
		if len(parts) == 2 {
			start, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
			end, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))

			if err1 == nil && err2 == nil && start > 0 && end > 0 {
				if cmd.Flags().Changed("output") {
					logger.Error("Cannot use --output with multiple tabs. Use --output-dir instead")
					return ErrOutputFlagConflict
				}

				if err := cfg.requireAutoDir(); err != nil {
					return err
				}
				return handleTabRange(ctx, bm, start, end, cfg)
			}
		}
	}

	var page *browser.Page
	var multipleMatches bool
	var matchedPages []*browser.Page

	// Try parsing as tab index
	if tabIndex, err := strconv.Atoi(tabValue); err == nil {
		logger.Verbose("Fetching from tab index: %d", tabIndex)
		page, err = bm.GetTabByIndex(tabIndex)
		if err != nil {
			if e := abortErr(ctx, err); e != nil {
				return e
			}
			if errors.Is(err, browser.ErrTabIndexInvalid) {
				logger.Error("Tab index out of range")
				logger.Info("Run 'snag --list-tabs' to see available tabs")
			}
			return err
		}
		logger.Verbose("Connected to tab [%d] from sorted order (by URL)", tabIndex)
	} else {
		// Pattern matching
		logger.Verbose("Fetching from tab matching pattern: %s", tabValue)
		matchedPages, err = bm.GetTabsByPattern(tabValue)
		if err != nil {
			if e := abortErr(ctx, err); e != nil {
				return e
			}
			if errors.Is(err, browser.ErrNoTabMatch) {
				logger.Error("No tab matches pattern '%s'", tabValue)
				logger.Info("Run 'snag --list-tabs' to see available tabs")
			}
			return err
		}

		if len(matchedPages) == 1 {
			page = matchedPages[0]
			logger.Verbose("Connected to tab matching pattern: %s", tabValue)
		} else {
			multipleMatches = true
			if cmd.Flags().Changed("output") {
				logger.Error("Cannot use --output with multiple tabs. Use --output-dir instead")
				logger.Info("Pattern '%s' matched %d tabs", tabValue, len(matchedPages))
				return ErrOutputFlagConflict
			}
			logger.Verbose("Pattern '%s' matched %d tabs", tabValue, len(matchedPages))
		}
	}

	if multipleMatches {
		if err := cfg.requireAutoDir(); err != nil {
			return err
		}
		return handleTabPatternBatch(ctx, bm, matchedPages, tabValue, cfg)
	}

	info, err := page.Meta()
	if err != nil {
		if e := abortErr(ctx, err); e != nil {
			return e
		}
		return fmt.Errorf("failed to get page info: %w", err)
	}

	logger.Verbose("Fetching content from: %s", info.URL)

	if err := preparePage(ctx, page, cfg, ""); err != nil {
		return err
	}
	if err := emitPage(ctx, page, cfg, info.URL, time.Now()); err != nil {
		if cfg.CloseTab {
			closeExistingTab(bm, page)
		}
		return err
	}
	if cfg.CloseTab {
		closeExistingTab(bm, page)
	}
	return nil
}

// closeExistingTab closes a tab in an attached browser. Close failures are
// warnings: content was already fetched. Chrome exits if this was the last tab.
func closeExistingTab(bm *browser.BrowserManager, page *browser.Page) {
	if page == nil {
		return
	}

	if bm != nil {
		tabs, err := bm.ListTabs()
		if err == nil && len(tabs) == 1 {
			logger.Info("Closing last tab, browser will close")
		}
	}

	if err := page.Close(); err != nil {
		logger.Verbose("Failed to close tab: %v", err)
	}
}

func handleTabRange(ctx context.Context, bm *browser.BrowserManager, start, end int, cfg *Config) error {
	pages, err := bm.GetTabsByRange(start, end)
	if err != nil {
		if e := abortErr(ctx, err); e != nil {
			return e
		}
		logger.Error("Failed to get tab range: %v", err)
		logger.Info("Run 'snag --list-tabs' to see available tabs")
		return err
	}

	logger.Verbose("Processing %d tabs from range [%d-%d]...", len(pages), start, end)
	return processBatchTabs(ctx, bm, pages, cfg)
}

func handleTabPatternBatch(ctx context.Context, bm *browser.BrowserManager, pages []*browser.Page, pattern string, cfg *Config) error {
	logger.Verbose("Processing %d tabs matching pattern '%s'...", len(pages), pattern)
	return processBatchTabs(ctx, bm, pages, cfg)
}

func handleOpenURLsInBrowser(cmd *cobra.Command, urls []string) error {
	// Warn about ignored flags
	if cmd.Flags().Changed("output") {
		logger.Warning("--output ignored with --open-browser (no content fetching)")
	}
	if cmd.Flags().Changed("output-dir") {
		logger.Warning("--output-dir ignored with --open-browser (no content fetching)")
	}
	if cmd.Flags().Changed("format") {
		logger.Warning("--format ignored with --open-browser (no content fetching)")
	}
	if cmd.Flags().Changed("timeout") {
		logger.Warning("--timeout ignored with --open-browser (no content fetching)")
	}
	if cmd.Flags().Changed("wait-for") {
		logger.Warning("--wait-for ignored with --open-browser (no content fetching)")
	}
	if closeTab {
		logger.Warning("--close-tab ignored with --open-browser (no content fetching)")
	}

	// Validate all URLs before expensive browser connection
	var validatedURLs []string
	for _, urlStr := range urls {
		validatedURL, err := validate.URL(urlStr)
		if err != nil {
			logger.Warning("Skipping invalid URL '%s': %v", urlStr, err)
			continue
		}
		validatedURLs = append(validatedURLs, validatedURL)
	}

	if len(validatedURLs) == 0 {
		logger.Error("No valid URLs to open")
		return fmt.Errorf("no valid URLs provided")
	}

	logger.Verbose("Opening %d valid URL%s in browser...", len(validatedURLs), plural(len(validatedURLs)))

	opts, err := browserOptionsFromFlags(cmd, true, false)
	if err != nil {
		return err
	}

	bm := browser.NewBrowserManager(opts)

	ctx := cmd.Context()
	err = bm.Connect(ctx)
	if err != nil {
		if e := abortErr(ctx, err); e != nil {
			return e
		}
		return err
	}

	for i, validatedURL := range validatedURLs {
		if e := abortErr(ctx, nil); e != nil {
			return e
		}
		current := i + 1
		logger.Verbose("[%d/%d] Opening: %s", current, len(validatedURLs), validatedURL)

		page, err := bm.NewPage()
		if err != nil {
			if e := abortErr(ctx, err); e != nil {
				return e
			}
			logger.Error("[%d/%d] Failed to create page: %v", current, len(validatedURLs), err)
			continue
		}

		err = page.NavigateTimeout(validatedURL, time.Duration(timeout)*time.Second)
		if err != nil {
			if e := abortErr(ctx, err); e != nil {
				return e
			}
			logger.Error("[%d/%d] Failed to navigate: %v", current, len(validatedURLs), err)
			continue
		}

		logger.Verbose("[%d/%d] Opened: %s", current, len(validatedURLs), validatedURL)
	}

	logger.Success("Browser will remain open with %d tabs", len(validatedURLs))

	// Don't close browser - leave it running for user
	return nil
}

func handleMultipleURLs(cmd *cobra.Command, urls []string) error {
	cfg, err := newEmitConfig(cmd, true)
	if err != nil {
		return err
	}

	opts, err := browserOptionsFromFlags(cmd, false, forceHead)
	if err != nil {
		return err
	}
	cfg.UserAgent = opts.UserAgent
	cfg.UserDataDir = opts.UserDataDir
	cfg.TempProfile = opts.TempProfile

	var validatedURLs []string
	for _, urlStr := range urls {
		validatedURL, err := validate.URL(urlStr)
		if err != nil {
			logger.Warning("Skipping invalid URL '%s': %v", urlStr, err)
			continue
		}
		validatedURLs = append(validatedURLs, validatedURL)
	}

	if len(validatedURLs) == 0 {
		logger.Error("No valid URLs to process")
		return validate.ErrNoValidURLs
	}

	logger.Verbose("Processing %d URL%s...", len(validatedURLs), plural(len(validatedURLs)))

	bm := browser.NewBrowserManager(opts)
	defer bm.Close()

	ctx := cmd.Context()
	err = bm.Connect(ctx)
	if err != nil {
		if e := abortErr(ctx, err); e != nil {
			return e
		}
		return err
	}

	if cfg.CloseTab && cfg.ForceHeadless {
		logger.Warning("--close-tab is ignored in headless mode (tabs close automatically)")
	}

	return processBatchURLs(ctx, bm, validatedURLs, cfg)
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

func loadURLsFromReader(ctx context.Context, reader io.Reader, source string) ([]string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	type result struct {
		urls []string
		err  error
	}
	ch := make(chan result, 1)
	go func() {
		urls, err := scanURLLines(reader, source)
		ch <- result{urls, err}
	}()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case r := <-ch:
		return r.urls, r.err
	}
}

func scanURLLines(reader io.Reader, source string) ([]string, error) {
	var urls []string
	scanner := bufio.NewScanner(reader)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}

		hasComment := false
		for _, marker := range []string{" #", " //"} {
			if idx := strings.Index(line, marker); idx != -1 {
				line = strings.TrimSpace(line[:idx])
				hasComment = true
				break
			}
		}

		if !hasComment && strings.Contains(line, " ") {
			logger.Warning("Line %d: URL contains space without comment marker - skipping: %s", lineNum, line)
			continue
		}

		if !strings.HasPrefix(line, "http://") && !strings.HasPrefix(line, "https://") && !strings.HasPrefix(line, "file://") {
			line = "https://" + line
		}

		if _, err := validate.URL(line); err != nil {
			logger.Warning("Line %d: Invalid URL - skipping: %s", lineNum, scanner.Text())
			continue
		}

		urls = append(urls, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading from %s: %w", source, err)
	}

	if len(urls) == 0 {
		return nil, validate.ErrNoValidURLs
	}

	logger.Verbose("Loaded %d URLs from %s", len(urls), source)
	return urls, nil
}

func loadURLsFromFile(ctx context.Context, filename string) ([]string, error) {
	if filename == "-" {
		return loadURLsFromReader(ctx, os.Stdin, "stdin")
	}

	file, err := os.Open(filename)
	if err != nil {
		logger.Error("Failed to open URL file: %s", filename)
		return nil, fmt.Errorf("failed to open URL file: %w", err)
	}
	defer file.Close()

	return loadURLsFromReader(ctx, file, filename)
}

func handleKillBrowser(cmd *cobra.Command) error {
	portChanged := cmd.Flags().Changed("port")

	bm := browser.NewBrowserManager(browser.BrowserOptions{
		Port: port,
	})

	var targetPort int
	if portChanged {
		targetPort = port
	} else {
		targetPort = 0
	}

	_, err := bm.KillBrowser(targetPort)
	return err
}

func handleDoctor(cmd *cobra.Command) error {
	userDataDirSet := cmd.Flags().Changed("user-data-dir")
	profileDir := ""
	if userDataDirSet {
		expanded, err := validate.ExpandHome(userDataDir)
		if err != nil {
			profileDir = strings.TrimSpace(userDataDir)
		} else {
			profileDir = expanded
		}
	}

	report, err := doctor.CollectDoctorInfo(Version, port, doctor.ProfileFlags{
		UserDataDir: profileDir,
		TempProfile: tempProfile,
	})
	if err != nil {
		logger.Verbose("Warning: Some diagnostic information could not be collected: %v", err)
	}

	report.Print()
	return nil
}
