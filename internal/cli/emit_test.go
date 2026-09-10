// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cli

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/p3bot/snag/internal/browser"
	"github.com/p3bot/snag/internal/format"
	"github.com/p3bot/snag/internal/logger"
)

func TestResolveOutputPath_ExplicitFile(t *testing.T) {
	logger.SetDefault(logger.Discard())
	cfg := &Config{OutputFile: "out.md", Format: format.Markdown}
	got, err := resolveOutputPath(nil, cfg, "https://example.com", time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if got != "out.md" {
		t.Fatalf("got %q, want out.md", got)
	}
}

func TestResolveOutputPath_Stdout(t *testing.T) {
	logger.SetDefault(logger.Discard())
	cfg := &Config{Format: format.Markdown}
	got, err := resolveOutputPath(nil, cfg, "https://example.com", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Fatalf("got %q, want empty stdout path", got)
	}
}

func TestResolveOutputPath_Dir(t *testing.T) {
	logger.SetDefault(logger.Discard())
	dir := t.TempDir()
	cfg := &Config{OutputDir: dir, Format: format.HTML}
	ts := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	got, err := resolveOutputPath(nil, cfg, "https://example.com", ts)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Dir(got) != dir {
		t.Fatalf("dir %q, want %q", filepath.Dir(got), dir)
	}
	base := filepath.Base(got)
	if !strings.HasPrefix(base, "2026-01-02-030405-") || !strings.HasSuffix(base, ".html") {
		t.Fatalf("filename %q", base)
	}
}

func TestResolveOutputPath_BinaryAutoName(t *testing.T) {
	logger.SetDefault(logger.Discard())
	cfg := &Config{Format: format.PDF}
	ts := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	got, err := resolveOutputPath(nil, cfg, "https://example.com", ts)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(got, ".pdf") {
		t.Fatalf("got %q, want .pdf", got)
	}
}

func TestFinishBatch(t *testing.T) {
	if err := finishBatch(2, 0); err != nil {
		t.Fatalf("all success: %v", err)
	}
	if err := finishBatch(1, 2); err == nil {
		t.Fatal("expected failure error")
	}
}

func TestProcessBatch_CanceledContext(t *testing.T) {
	logger.SetDefault(logger.Discard())
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	bm := browser.NewBrowserManager(browser.BrowserOptions{})
	cfg := &Config{Format: format.Markdown, Timeout: 1}

	err := processBatchTabs(ctx, bm, []*browser.Page{{}}, cfg)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("processBatchTabs: %v", err)
	}
	err = processBatchURLs(ctx, bm, []string{"https://example.com"}, cfg)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("processBatchURLs: %v", err)
	}
}

func TestProcessBatchURLs_LastItemCanceledDuringPrepare(t *testing.T) {
	logger.SetDefault(logger.Discard())
	origNew, origPrep, origEmit := newBatchPage, runPreparePage, runEmitPage
	t.Cleanup(func() {
		newBatchPage, runPreparePage, runEmitPage = origNew, origPrep, origEmit
	})

	newBatchPage = func(*browser.BrowserManager) (*browser.Page, error) {
		return &browser.Page{}, nil
	}
	n := 0
	runPreparePage = func(context.Context, *browser.Page, *Config, string) error {
		n++
		if n < 2 {
			return nil
		}
		return context.Canceled
	}
	runEmitPage = func(context.Context, *browser.Page, *Config, string, time.Time) error {
		return nil
	}

	err := processBatchURLs(context.Background(), browser.NewBrowserManager(browser.BrowserOptions{}),
		[]string{"https://a.example", "https://b.example"},
		&Config{Format: format.Markdown, Timeout: 1})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v, want context.Canceled", err)
	}
	if strings.Contains(err.Error(), "batch processing") {
		t.Fatalf("last-item cancel must not become finishBatch: %v", err)
	}
}

func TestProcessBatchURLs_LastItemCanceledDuringEmit(t *testing.T) {
	logger.SetDefault(logger.Discard())
	origNew, origPrep, origEmit := newBatchPage, runPreparePage, runEmitPage
	t.Cleanup(func() {
		newBatchPage, runPreparePage, runEmitPage = origNew, origPrep, origEmit
	})

	newBatchPage = func(*browser.BrowserManager) (*browser.Page, error) {
		return &browser.Page{}, nil
	}
	runPreparePage = func(context.Context, *browser.Page, *Config, string) error { return nil }
	n := 0
	runEmitPage = func(context.Context, *browser.Page, *Config, string, time.Time) error {
		n++
		if n < 2 {
			return nil
		}
		return context.Canceled
	}

	err := processBatchURLs(context.Background(), browser.NewBrowserManager(browser.BrowserOptions{}),
		[]string{"https://a.example", "https://b.example"},
		&Config{Format: format.Markdown, Timeout: 1})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v, want context.Canceled", err)
	}
	if strings.Contains(err.Error(), "batch processing") {
		t.Fatalf("last-item cancel must not become finishBatch: %v", err)
	}
}

func TestProcessBatchURLs_PrepareErrorIsFailure(t *testing.T) {
	logger.SetDefault(logger.Discard())
	origNew, origPrep, origEmit := newBatchPage, runPreparePage, runEmitPage
	t.Cleanup(func() {
		newBatchPage, runPreparePage, runEmitPage = origNew, origPrep, origEmit
	})

	newBatchPage = func(*browser.BrowserManager) (*browser.Page, error) {
		return &browser.Page{}, nil
	}
	runPreparePage = func(context.Context, *browser.Page, *Config, string) error {
		return errors.New("load failed")
	}
	runEmitPage = func(context.Context, *browser.Page, *Config, string, time.Time) error {
		return nil
	}

	err := processBatchURLs(context.Background(), browser.NewBrowserManager(browser.BrowserOptions{}),
		[]string{"https://a.example"},
		&Config{Format: format.Markdown, Timeout: 1})
	if err == nil || !strings.Contains(err.Error(), "batch processing completed with 1 failures") {
		t.Fatalf("got %v, want counted failure", err)
	}
}

func TestConfigFetchOptions(t *testing.T) {
	c := &Config{Timeout: 15, WaitFor: ".ready"}
	opts := c.fetchOptions("https://example.com")
	if opts.URL != "https://example.com" || opts.Timeout != 15 || opts.WaitFor != ".ready" {
		t.Fatalf("%+v", opts)
	}
}

func TestNewEmitConfig_EmptyOutputFlag(t *testing.T) {
	logger.SetDefault(logger.Discard())
	resetCLIFlags()
	t.Cleanup(resetCLIFlags)

	if err := rootCmd.Flags().Set("output", "  "); err != nil {
		t.Fatal(err)
	}
	_, err := newEmitConfig(rootCmd, false)
	if err == nil {
		t.Fatal("expected error for empty --output")
	}
}

func TestNewEmitConfig_OmittedOutput(t *testing.T) {
	logger.SetDefault(logger.Discard())
	resetCLIFlags()
	t.Cleanup(resetCLIFlags)

	cfg, err := newEmitConfig(rootCmd, false)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.OutputFile != "" {
		t.Fatalf("OutputFile = %q, want empty (stdout)", cfg.OutputFile)
	}
}
