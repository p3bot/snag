// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/p3bot/snag/internal/browser"
	"github.com/p3bot/snag/internal/fetch"
	"github.com/p3bot/snag/internal/logger"
	"github.com/p3bot/snag/internal/output"
	"github.com/p3bot/snag/internal/validate"
)

func outputPageInfo(info *fetch.PageInfo, outputFile string) error {
	if info != nil {
		info.Slug = output.SlugifyTitle(info.Title, output.MaxSlugLength)
	}

	jsonData, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal page info to JSON: %w", err)
	}

	if outputFile == "" {
		fmt.Println(string(jsonData))
		return nil
	}

	if err := os.WriteFile(outputFile, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write info to file: %w", err)
	}

	logger.Success("Saved info to %s", outputFile)
	return nil
}

func handleInfoFromURL(cmd *cobra.Command, urlStr string) error {
	validatedURL, err := validate.URL(urlStr)
	if err != nil {
		return err
	}

	if err := validate.Timeout(timeout); err != nil {
		return err
	}

	if err := validate.Port(port); err != nil {
		return err
	}

	outputFile := strings.TrimSpace(flagOutput)
	if cmd.Flags().Changed("output") && outputFile != "" {
		if err := validate.OutputPath(outputFile); err != nil {
			return err
		}
	}

	opts, err := browserOptionsFromFlags(cmd, false, forceHead)
	if err != nil {
		return err
	}

	validatedWaitFor := validate.WaitFor(waitFor, cmd.Flags().Changed("wait-for"))

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

	page, err := bm.NewPage()
	if err != nil {
		return err
	}

	if closeTab || bm.LaunchedHeadless() {
		defer bm.ClosePage(page)
	}

	fetcher := fetch.NewPageFetcher(page, timeout)
	_, err = fetcher.Fetch(fetch.FetchOptions{
		URL:     validatedURL,
		Timeout: timeout,
		WaitFor: validatedWaitFor,
	})
	if err != nil {
		return err
	}

	pageInfo, err := fetch.ExtractPageInfo(page)
	if err != nil {
		return err
	}

	return outputPageInfo(pageInfo, outputFile)
}

func handleInfoFromTab(cmd *cobra.Command) error {
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

	outputFile := strings.TrimSpace(flagOutput)
	if cmd.Flags().Changed("output") && outputFile != "" {
		if err := validate.OutputPath(outputFile); err != nil {
			return err
		}
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

	var page *browser.Page

	if tabIndex, err := strconv.Atoi(tabValue); err == nil {
		logger.Verbose("Getting info from tab index: %d", tabIndex)
		page, err = bm.GetTabByIndex(tabIndex)
		if err != nil {
			if errors.Is(err, browser.ErrTabIndexInvalid) {
				logger.Error("Tab index out of range")
				logger.Info("Run 'snag --list-tabs' to see available tabs")
			}
			return err
		}
	} else {
		logger.Verbose("Getting info from tab matching pattern: %s", tabValue)
		matchedPages, err := bm.GetTabsByPattern(tabValue)
		if err != nil {
			if errors.Is(err, browser.ErrNoTabMatch) {
				logger.Error("No tab matches pattern '%s'", tabValue)
				logger.Info("Run 'snag --list-tabs' to see available tabs")
			}
			return err
		}

		if len(matchedPages) > 1 {
			logger.Error("Pattern '%s' matched %d tabs, --info requires exactly one", tabValue, len(matchedPages))
			logger.Info("Use a more specific pattern or tab index")
			return fmt.Errorf("pattern matched multiple tabs")
		}

		page = matchedPages[0]
	}

	if cmd.Flags().Changed("wait-for") {
		validatedWaitFor := validate.WaitFor(waitFor, true)
		if validatedWaitFor != "" {
			err := fetch.WaitForSelector(page, validatedWaitFor, time.Duration(timeout)*time.Second)
			if err != nil {
				return err
			}
		}
	}

	pageInfo, err := fetch.ExtractPageInfo(page)
	if err != nil {
		return err
	}

	return outputPageInfo(pageInfo, outputFile)
}
