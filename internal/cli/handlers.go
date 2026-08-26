// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/p3bot/snag/internal/browser"
	"github.com/p3bot/snag/internal/doctor"
	"github.com/p3bot/snag/internal/fetch"
	"github.com/p3bot/snag/internal/format"
	"github.com/p3bot/snag/internal/logger"
	"github.com/p3bot/snag/internal/output"
	"github.com/p3bot/snag/internal/validate"
)

func snag(config *Config) error {
	bm := browser.NewBrowserManager(config.BrowserOptions())

	browserMutex.Lock()
	browserManager = bm
	browserMutex.Unlock()

	defer func() {
		if config.CloseTab {
			logger.Verbose("Cleanup: closing tab and browser if needed")
		}
		bm.Close()
		browserMutex.Lock()
		browserManager = nil
		browserMutex.Unlock()
	}()

	err := bm.Connect()
	if err != nil {
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
		return err
	}

	if config.CloseTab {
		defer bm.ClosePage(page)
	}

	fetcher := fetch.NewPageFetcher(page, config.Timeout)

	_, err = fetcher.Fetch(fetch.FetchOptions{
		URL:     config.URL,
		Timeout: config.Timeout,
		WaitFor: config.WaitFor,
	})
	if err != nil {
		return err
	}

	if config.OutputDir != "" {
		info, err := page.Meta()
		if err != nil {
			return fmt.Errorf("failed to get page info: %w", err)
		}

		config.OutputFile, err = generateOutputFilename(
			info.Title, config.URL, config.Format,
			time.Now(), config.OutputDir,
		)
		if err != nil {
			return err
		}
	}

	// For binary formats without -o or -d: auto-generate filename in current directory
	// Binary formats (PDF, PNG) should NEVER output to stdout (corrupts terminal)
	if config.OutputFile == "" && (config.Format == format.PDF || config.Format == format.PNG) {
		info, err := page.Meta()
		if err != nil {
			return fmt.Errorf("failed to get page info: %w", err)
		}

		config.OutputFile, err = generateOutputFilename(
			info.Title, config.URL, config.Format,
			time.Now(), ".",
		)
		if err != nil {
			return err
		}
		logger.Info("Filename: %s", config.OutputFile)
	}

	return format.ProcessContent(page, config.Format, config.OutputFile)
}

func generateOutputFilename(title, url, format string,
	timestamp time.Time, outputDir string) (string, error) {
	filename := output.GenerateFilename(title, format, timestamp, url)

	finalFilename, err := output.ResolveConflict(outputDir, filename)
	if err != nil {
		return "", fmt.Errorf("failed to resolve filename conflict: %w", err)
	}

	return filepath.Join(outputDir, finalFilename), nil
}

func connectToExistingBrowser(port int) (*browser.BrowserManager, error) {
	bm := browser.NewBrowserManager(browser.BrowserOptions{
		Port: port,
	})

	browserMutex.Lock()
	browserManager = bm
	browserMutex.Unlock()

	if err := bm.ConnectExisting(); err != nil {
		browserMutex.Lock()
		browserManager = nil
		browserMutex.Unlock()
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

func handleListTabs(cmd *cobra.Command) error {
	bm, err := connectToExistingBrowser(port)
	if err != nil {
		return err
	}
	defer func() {
		browserMutex.Lock()
		browserManager = nil
		browserMutex.Unlock()
	}()

	tabs, err := bm.ListTabs()
	if err != nil {
		return err
	}

	displayTabList(tabs, os.Stdout, verbose)

	return nil
}

func handleAllTabs(cmd *cobra.Command) error {
	outputFormat := validate.NormalizeFormat(flagFormat)
	outDir := strings.TrimSpace(outputDir)
	if outDir == "" {
		outDir = "."
	}

	if cmd.Flags().Changed("user-agent") {
		logger.Warning("--user-agent is ignored with --all-tabs (cannot change existing tabs' user agents)")
	}
	if cmd.Flags().Changed("user-data-dir") {
		logger.Warning("--user-data-dir ignored when connecting to existing browser")
	}
	if cmd.Flags().Changed("timeout") && waitFor == "" {
		logger.Warning("--timeout is ignored without --wait-for when using --all-tabs")
	}

	if err := validate.Format(outputFormat); err != nil {
		return err
	}

	if err := validate.Timeout(timeout); err != nil {
		return err
	}

	if err := validate.Directory(outDir); err != nil {
		return err
	}

	bm, err := connectToExistingBrowser(port)
	if err != nil {
		return err
	}
	defer func() {
		browserMutex.Lock()
		browserManager = nil
		browserMutex.Unlock()
	}()

	tabs, err := bm.ListTabs()
	if err != nil {
		return err
	}

	if len(tabs) == 0 {
		logger.Info("No tabs open in browser")
		return nil
	}

	timestamp := time.Now()

	logger.Info("Processing %d tabs...", len(tabs))

	successCount := 0
	failureCount := 0

	for _, tab := range tabs {
		if validate.IsNonFetchableURL(tab.URL) {
			logger.Warning("[%d/%d] Skipping tab: %s (not fetchable)", tab.Index, len(tabs), tab.URL)
			continue
		}

		logger.Info("[%d/%d] Processing: %s", tab.Index, len(tabs), tab.URL)

		page, err := bm.GetTabByIndex(tab.Index)
		if err != nil {
			logger.Error("[%d/%d] Failed to get tab: %v", tab.Index, len(tabs), err)
			failureCount++
			continue
		}

		if waitFor != "" {
			err := fetch.WaitForSelector(page, waitFor, time.Duration(timeout)*time.Second)
			if err != nil {
				logger.Error("[%d/%d] Wait failed: %v", tab.Index, len(tabs), err)
				failureCount++
				continue
			}
		}

		outputPath, err := generateOutputFilename(
			tab.Title, tab.URL, outputFormat,
			timestamp, outDir,
		)
		if err != nil {
			logger.Error("[%d/%d] Failed to generate filename: %v", tab.Index, len(tabs), err)
			failureCount++
			continue
		}

		if err := format.ProcessContent(page, outputFormat, outputPath); err != nil {
			logger.Error("[%d/%d] Failed to process content: %v", tab.Index, len(tabs), err)
			failureCount++
			if closeTab {
				if err := page.Close(); err != nil {
					logger.Verbose("[%d/%d] Failed to close tab: %v", tab.Index, len(tabs), err)
				}
			}
			continue
		}

		successCount++

		if closeTab {
			if tab.Index == len(tabs) {
				logger.Info("Closing last tab, browser will close")
			}
			if err := page.Close(); err != nil {
				logger.Verbose("[%d/%d] Failed to close tab: %v", tab.Index, len(tabs), err)
			}
		}
	}

	logger.Success("Batch complete: %d succeeded, %d failed", successCount, failureCount)

	if failureCount > 0 {
		return fmt.Errorf("batch processing completed with %d failures", failureCount)
	}

	return nil
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
	if cmd.Flags().Changed("user-data-dir") {
		logger.Warning("--user-data-dir ignored when connecting to existing browser")
	}
	if cmd.Flags().Changed("timeout") && !cmd.Flags().Changed("wait-for") {
		logger.Warning("--timeout is ignored without --wait-for when using --tab")
	}

	// Validate early before expensive browser connection
	outputFormat := validate.NormalizeFormat(flagFormat)
	validatedWaitFor := validate.WaitFor(waitFor, cmd.Flags().Changed("wait-for"))
	outputFile := strings.TrimSpace(flagOutput)

	if err := validate.Format(outputFormat); err != nil {
		return err
	}

	if err := validate.Timeout(timeout); err != nil {
		return err
	}

	if outputFile != "" {
		if err := validate.OutputPath(outputFile); err != nil {
			return err
		}
		validate.CheckExtensionMismatch(outputFile, outputFormat)
	}

	bm, err := connectToExistingBrowser(port)
	if err != nil {
		return err
	}
	defer func() {
		browserMutex.Lock()
		browserManager = nil
		browserMutex.Unlock()
	}()

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

				return handleTabRange(cmd, bm, start, end)
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
			if errors.Is(err, browser.ErrTabIndexInvalid) {
				logger.Error("Tab index out of range")
				logger.Info("Run 'snag --list-tabs' to see available tabs")
			}
			return err
		}
		logger.Success("Connected to tab [%d] from sorted order (by URL)", tabIndex)
	} else {
		// Pattern matching
		logger.Verbose("Fetching from tab matching pattern: %s", tabValue)
		matchedPages, err = bm.GetTabsByPattern(tabValue)
		if err != nil {
			if errors.Is(err, browser.ErrNoTabMatch) {
				logger.Error("No tab matches pattern '%s'", tabValue)
				logger.Info("Run 'snag --list-tabs' to see available tabs")
			}
			return err
		}

		if len(matchedPages) == 1 {
			page = matchedPages[0]
			logger.Success("Connected to tab matching pattern: %s", tabValue)
		} else {
			multipleMatches = true
			if cmd.Flags().Changed("output") {
				logger.Error("Cannot use --output with multiple tabs. Use --output-dir instead")
				logger.Info("Pattern '%s' matched %d tabs", tabValue, len(matchedPages))
				return ErrOutputFlagConflict
			}
			logger.Info("Pattern '%s' matched %d tabs", tabValue, len(matchedPages))
		}
	}

	if multipleMatches {
		return handleTabPatternBatch(cmd, matchedPages, tabValue)
	}

	// Single tab fetch (validation already done earlier)
	info, err := page.Meta()
	if err != nil {
		return fmt.Errorf("failed to get page info: %w", err)
	}

	logger.Info("Fetching content from: %s", info.URL)

	if validatedWaitFor != "" {
		err := fetch.WaitForSelector(page, validatedWaitFor, time.Duration(timeout)*time.Second)
		if err != nil {
			return err
		}
	}

	// For binary formats without -o or -d: auto-generate filename
	if outputFile == "" && (outputFormat == format.PDF || outputFormat == format.PNG) {
		outputFile, err = generateOutputFilename(
			info.Title, info.URL, outputFormat,
			time.Now(), ".",
		)
		if err != nil {
			return err
		}
		logger.Info("Filename: %s", outputFile)
	}

	err = format.ProcessContent(page, outputFormat, outputFile)
	if closeTab {
		closeExistingTab(page)
	}
	return err
}

// closeExistingTab closes a tab in an attached browser. Close failures are
// warnings: content was already fetched. Chrome exits if this was the last tab.
func closeExistingTab(page *browser.Page) {
	if page == nil {
		return
	}

	browserMutex.Lock()
	bm := browserManager
	browserMutex.Unlock()
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

func processBatchTabs(pages []*browser.Page, config *Config) error {
	timestamp := time.Now()

	successCount := 0
	failureCount := 0

	for i, page := range pages {
		current := i + 1
		total := len(pages)

		info, err := page.Meta()
		if err != nil {
			logger.Error("[%d/%d] Failed to get tab info: %v", current, total, err)
			failureCount++
			continue
		}

		logger.Info("[%d/%d] Processing: %s", current, total, info.URL)

		if config.WaitFor != "" {
			err := fetch.WaitForSelector(page, config.WaitFor, time.Duration(config.Timeout)*time.Second)
			if err != nil {
				logger.Error("[%d/%d] Wait failed: %v", current, total, err)
				failureCount++
				continue
			}
		}

		outputPath, err := generateOutputFilename(
			info.Title, info.URL, config.Format,
			timestamp, config.OutputDir,
		)
		if err != nil {
			logger.Error("[%d/%d] Failed to generate filename: %v", current, total, err)
			failureCount++
			continue
		}

		if err := format.ProcessContent(page, config.Format, outputPath); err != nil {
			logger.Error("[%d/%d] Failed to process content: %v", current, total, err)
			failureCount++
			if config.CloseTab {
				closeExistingTab(page)
			}
			continue
		}

		successCount++

		if config.CloseTab {
			closeExistingTab(page)
		}
	}

	logger.Success("Batch complete: %d succeeded, %d failed", successCount, failureCount)

	if failureCount > 0 {
		return fmt.Errorf("batch processing completed with %d failures", failureCount)
	}

	return nil
}

func handleTabRange(cmd *cobra.Command, bm *browser.BrowserManager, start, end int) error {
	outputFormat := validate.NormalizeFormat(flagFormat)
	validatedWaitFor := validate.WaitFor(waitFor, cmd.Flags().Changed("wait-for"))
	outDir := strings.TrimSpace(outputDir)
	if outDir == "" {
		outDir = "."
	}

	if err := validate.Format(outputFormat); err != nil {
		return err
	}

	if err := validate.Timeout(timeout); err != nil {
		return err
	}

	if err := validate.Directory(outDir); err != nil {
		return err
	}

	pages, err := bm.GetTabsByRange(start, end)
	if err != nil {
		logger.Error("Failed to get tab range: %v", err)
		logger.Info("Run 'snag --list-tabs' to see available tabs")
		return err
	}

	logger.Info("Processing %d tabs from range [%d-%d]...", len(pages), start, end)

	config := &Config{
		Format:    outputFormat,
		WaitFor:   validatedWaitFor,
		Timeout:   timeout,
		OutputDir: outDir,
		CloseTab:  closeTab,
	}

	return processBatchTabs(pages, config)
}

func handleTabPatternBatch(cmd *cobra.Command, pages []*browser.Page, pattern string) error {
	outputFormat := validate.NormalizeFormat(flagFormat)
	validatedWaitFor := validate.WaitFor(waitFor, cmd.Flags().Changed("wait-for"))
	outDir := strings.TrimSpace(outputDir)
	if outDir == "" {
		outDir = "."
	}

	if err := validate.Format(outputFormat); err != nil {
		return err
	}

	if err := validate.Timeout(timeout); err != nil {
		return err
	}

	if err := validate.Directory(outDir); err != nil {
		return err
	}

	logger.Info("Processing %d tabs matching pattern '%s'...", len(pages), pattern)

	config := &Config{
		Format:    outputFormat,
		WaitFor:   validatedWaitFor,
		Timeout:   timeout,
		OutputDir: outDir,
		CloseTab:  closeTab,
	}

	return processBatchTabs(pages, config)
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

	logger.Info("Opening %d valid URL%s in browser...", len(validatedURLs), plural(len(validatedURLs)))

	opts, err := browserOptionsFromFlags(cmd, true, false)
	if err != nil {
		return err
	}

	bm := browser.NewBrowserManager(opts)

	browserMutex.Lock()
	browserManager = bm
	browserMutex.Unlock()
	defer func() {
		browserMutex.Lock()
		browserManager = nil
		browserMutex.Unlock()
	}()

	err = bm.Connect()
	if err != nil {
		return err
	}

	for i, validatedURL := range validatedURLs {
		current := i + 1
		logger.Info("[%d/%d] Opening: %s", current, len(validatedURLs), validatedURL)

		page, err := bm.NewPage()
		if err != nil {
			logger.Error("[%d/%d] Failed to create page: %v", current, len(validatedURLs), err)
			continue
		}

		err = page.NavigateTimeout(validatedURL, time.Duration(timeout)*time.Second)
		if err != nil {
			logger.Error("[%d/%d] Failed to navigate: %v", current, len(validatedURLs), err)
			continue
		}

		logger.Success("[%d/%d] Opened: %s", current, len(validatedURLs), validatedURL)
	}

	logger.Success("Browser will remain open with %d tabs", len(validatedURLs))

	// Don't close browser - leave it running for user
	return nil
}

func handleMultipleURLs(cmd *cobra.Command, urls []string) error {
	outputFile := strings.TrimSpace(flagOutput)
	outDir := strings.TrimSpace(outputDir)

	outputFormat := validate.NormalizeFormat(flagFormat)
	if err := validate.Format(outputFormat); err != nil {
		return err
	}

	if err := validate.Timeout(timeout); err != nil {
		return err
	}

	if err := validate.Port(port); err != nil {
		return err
	}

	if outputFile != "" {
		if err := validate.OutputPath(outputFile); err != nil {
			return err
		}
	}

	if cmd.Flags().Changed("output-dir") && outDir == "" {
		outDir = "."
	}

	if outDir != "" {
		if err := validate.Directory(outDir); err != nil {
			return err
		}
	}

	opts, err := browserOptionsFromFlags(cmd, false, forceHead)
	if err != nil {
		return err
	}

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

	logger.Info("Processing %d URL%s...", len(validatedURLs), plural(len(validatedURLs)))

	bm := browser.NewBrowserManager(opts)
	browserMutex.Lock()
	browserManager = bm
	browserMutex.Unlock()
	defer func() {
		bm.Close()
		browserMutex.Lock()
		browserManager = nil
		browserMutex.Unlock()
	}()

	err = bm.Connect()
	if err != nil {
		return err
	}

	if closeTab && forceHead {
		logger.Warning("--close-tab is ignored in headless mode (tabs close automatically)")
	}

	validatedWaitFor := validate.WaitFor(waitFor, cmd.Flags().Changed("wait-for"))

	timestamp := time.Now()

	successCount := 0
	failureCount := 0

	for i, validatedURL := range validatedURLs {
		current := i + 1
		total := len(validatedURLs)

		logger.Info("[%d/%d] Fetching: %s", current, total, validatedURL)

		page, err := bm.NewPage()
		if err != nil {
			logger.Error("[%d/%d] Failed to create page: %v", current, total, err)
			failureCount++
			continue
		}

		fetcher := fetch.NewPageFetcher(page, timeout)
		_, err = fetcher.Fetch(fetch.FetchOptions{
			URL:     validatedURL,
			Timeout: timeout,
			WaitFor: validatedWaitFor,
		})
		if err != nil {
			logger.Error("[%d/%d] Failed to fetch: %v", current, total, err)
			bm.ClosePage(page)
			failureCount++
			continue
		}

		info, err := page.Meta()
		if err != nil {
			logger.Error("[%d/%d] Failed to get page info: %v", current, total, err)
			bm.ClosePage(page)
			failureCount++
			continue
		}

		outputPath, err := generateOutputFilename(
			info.Title, validatedURL, outputFormat,
			timestamp, outDir,
		)
		if err != nil {
			logger.Error("[%d/%d] Failed to generate filename: %v", current, total, err)
			bm.ClosePage(page)
			failureCount++
			continue
		}

		if err := format.ProcessContent(page, outputFormat, outputPath); err != nil {
			logger.Error("[%d/%d] Failed to save content: %v", current, total, err)
			bm.ClosePage(page)
			failureCount++
			continue
		}

		if bm.LaunchedHeadless() || closeTab {
			bm.ClosePage(page)
		}

		successCount++
	}

	logger.Success("Batch complete: %d succeeded, %d failed", successCount, failureCount)

	if failureCount > 0 {
		return fmt.Errorf("batch processing completed with %d failures", failureCount)
	}

	return nil
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

func loadURLsFromReader(reader io.Reader, source string) ([]string, error) {
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

func loadURLsFromFile(filename string) ([]string, error) {
	if filename == "-" {
		return loadURLsFromReader(os.Stdin, "stdin")
	}

	file, err := os.Open(filename)
	if err != nil {
		logger.Error("Failed to open URL file: %s", filename)
		return nil, fmt.Errorf("failed to open URL file: %w", err)
	}
	defer file.Close()

	return loadURLsFromReader(file, filename)
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
	report, err := doctor.CollectDoctorInfo(Version, port)
	if err != nil {
		logger.Verbose("Warning: Some diagnostic information could not be collected: %v", err)
	}

	report.Print()
	return nil
}
