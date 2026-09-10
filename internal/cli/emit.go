// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cli

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/p3bot/snag/internal/browser"
	"github.com/p3bot/snag/internal/fetch"
	"github.com/p3bot/snag/internal/format"
	"github.com/p3bot/snag/internal/logger"
	"github.com/p3bot/snag/internal/output"
	"github.com/p3bot/snag/internal/validate"
)

func (c *Config) fetchOptions(url string) fetch.FetchOptions {
	return fetch.FetchOptions{
		URL:     url,
		Timeout: c.Timeout,
		WaitFor: c.WaitFor,
	}
}

func newEmitConfig(cmd *cobra.Command, autoDir bool) (*Config, error) {
	outputFormat := validate.NormalizeFormat(flagFormat)
	if err := validate.Format(outputFormat); err != nil {
		return nil, err
	}
	if err := validate.Timeout(timeout); err != nil {
		return nil, err
	}
	if err := validate.Port(port); err != nil {
		return nil, err
	}

	outputFile := strings.TrimSpace(flagOutput)
	outDir := strings.TrimSpace(outputDir)
	if cmd.Flags().Changed("output-dir") && outDir == "" {
		outDir = "."
	}
	if autoDir && outDir == "" {
		outDir = "."
	}

	if cmd.Flags().Changed("output") || outputFile != "" {
		if err := validate.OutputPath(outputFile); err != nil {
			return nil, err
		}
		validate.CheckExtensionMismatch(outputFile, outputFormat)
	}
	if outDir != "" {
		if err := validate.Directory(outDir); err != nil {
			return nil, err
		}
	}

	return &Config{
		OutputFile:    outputFile,
		OutputDir:     outDir,
		Format:        outputFormat,
		Timeout:       timeout,
		WaitFor:       validate.WaitFor(waitFor, cmd.Flags().Changed("wait-for")),
		Port:          port,
		CloseTab:      closeTab,
		ForceHeadless: forceHead,
	}, nil
}

func (c *Config) requireAutoDir() error {
	if c.OutputDir != "" {
		return nil
	}
	c.OutputDir = "."
	return validate.Directory(c.OutputDir)
}

func generateOutputFilename(title, url, formatName string,
	timestamp time.Time, outputDir string) (string, error) {
	filename := output.GenerateFilename(title, formatName, timestamp, url)

	finalFilename, err := output.ResolveConflict(outputDir, filename)
	if err != nil {
		return "", fmt.Errorf("failed to resolve filename conflict: %w", err)
	}

	return filepath.Join(outputDir, finalFilename), nil
}

func resolveOutputPath(page *browser.Page, cfg *Config, fallbackURL string, ts time.Time) (string, error) {
	if cfg.OutputFile != "" {
		return cfg.OutputFile, nil
	}

	binary := cfg.Format == format.PDF || cfg.Format == format.PNG
	if cfg.OutputDir == "" && !binary {
		return "", nil
	}

	dir := cfg.OutputDir
	announce := false
	if dir == "" {
		dir = "."
		announce = true
	}

	title := ""
	urlStr := fallbackURL
	if page != nil {
		info, err := page.Meta()
		if err != nil {
			return "", fmt.Errorf("failed to get page info: %w", err)
		}
		title = info.Title
		if info.URL != "" {
			urlStr = info.URL
		}
	}

	path, err := generateOutputFilename(title, urlStr, cfg.Format, ts, dir)
	if err != nil {
		return "", err
	}
	if announce {
		logger.Info("Filename: %s", path)
	}
	return path, nil
}

func emitPage(ctx context.Context, page *browser.Page, cfg *Config, fallbackURL string, ts time.Time) error {
	path, err := resolveOutputPath(page, cfg, fallbackURL, ts)
	if err != nil {
		return err
	}
	data, err := format.Render(page, cfg.Format)
	if err != nil {
		if e := abortErr(ctx, err); e != nil {
			return e
		}
		return err
	}
	if err := output.Write(data, path); err != nil {
		if e := abortErr(ctx, err); e != nil {
			return e
		}
		return err
	}
	return nil
}

func preparePage(ctx context.Context, page *browser.Page, cfg *Config, url string) error {
	fetcher := fetch.NewPageFetcher(page, cfg.Timeout)
	var err error
	if url != "" {
		err = fetcher.Fetch(ctx, cfg.fetchOptions(url))
	} else {
		err = fetcher.Ready(ctx, cfg.fetchOptions(""))
	}
	if err != nil {
		if e := abortErr(ctx, err); e != nil {
			return e
		}
		return err
	}
	return nil
}

var (
	newBatchPage   = func(bm *browser.BrowserManager) (*browser.Page, error) { return bm.NewPage() }
	runPreparePage = preparePage
	runEmitPage    = emitPage
)

func finishBatch(success, fail int) error {
	logger.Verbose("Batch complete: %d succeeded, %d failed", success, fail)
	if fail > 0 {
		return fmt.Errorf("batch processing completed with %d failures", fail)
	}
	return nil
}

func processBatchURLs(ctx context.Context, bm *browser.BrowserManager, urls []string, cfg *Config) error {
	ts := time.Now()
	success, fail := 0, 0
	total := len(urls)

	for i, u := range urls {
		if e := abortErr(ctx, nil); e != nil {
			return e
		}
		current := i + 1
		logger.Verbose("[%d/%d] Fetching: %s", current, total, u)

		page, err := newBatchPage(bm)
		if err != nil {
			if e := abortErr(ctx, err); e != nil {
				return e
			}
			logger.Error("[%d/%d] Failed to create page: %v", current, total, err)
			fail++
			continue
		}

		if err := runPreparePage(ctx, page, cfg, u); err != nil {
			if e := abortErr(ctx, err); e != nil {
				return e
			}
			logger.Error("[%d/%d] Failed to fetch: %v", current, total, err)
			bm.ClosePage(page)
			fail++
			continue
		}

		if err := runEmitPage(ctx, page, cfg, u, ts); err != nil {
			if e := abortErr(ctx, err); e != nil {
				return e
			}
			logger.Error("[%d/%d] Failed to save content: %v", current, total, err)
			bm.ClosePage(page)
			fail++
			continue
		}

		if bm.LaunchedHeadless() || cfg.CloseTab {
			bm.ClosePage(page)
		}
		success++
	}

	return finishBatch(success, fail)
}

func processBatchTabs(ctx context.Context, bm *browser.BrowserManager, pages []*browser.Page, cfg *Config) error {
	ts := time.Now()
	success, fail := 0, 0
	total := len(pages)

	for i, page := range pages {
		if e := abortErr(ctx, nil); e != nil {
			return e
		}
		current := i + 1

		info, err := page.Meta()
		if err != nil {
			if e := abortErr(ctx, err); e != nil {
				return e
			}
			logger.Error("[%d/%d] Failed to get tab info: %v", current, total, err)
			fail++
			continue
		}

		if validate.IsNonFetchableURL(info.URL) {
			logger.Warning("[%d/%d] Skipping tab: %s (not fetchable)", current, total, info.URL)
			continue
		}

		logger.Verbose("[%d/%d] Processing: %s", current, total, info.URL)

		if err := runPreparePage(ctx, page, cfg, ""); err != nil {
			if e := abortErr(ctx, err); e != nil {
				return e
			}
			logger.Error("[%d/%d] Failed to prepare tab: %v", current, total, err)
			fail++
			continue
		}

		if err := runEmitPage(ctx, page, cfg, info.URL, ts); err != nil {
			if e := abortErr(ctx, err); e != nil {
				return e
			}
			logger.Error("[%d/%d] Failed to process content: %v", current, total, err)
			if cfg.CloseTab {
				closeExistingTab(bm, page)
			}
			fail++
			continue
		}

		success++
		if cfg.CloseTab {
			closeExistingTab(bm, page)
		}
	}

	return finishBatch(success, fail)
}
