// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package fetch

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/p3bot/snag/internal/browser"
	"github.com/p3bot/snag/internal/logger"
)

type PageFetcher struct {
	page    *browser.Page
	timeout time.Duration
}

type FetchOptions struct {
	URL     string
	Timeout int
	WaitFor string
}

func NewPageFetcher(page *browser.Page, timeout int) *PageFetcher {
	if page == nil {
		logger.Warning("NewPageFetcher called with nil page")
	}
	return &PageFetcher{
		page:    page,
		timeout: time.Duration(timeout) * time.Second,
	}
}

func pageLoadTimeout(opts FetchOptions) error {
	logger.Error("Page load timeout exceeded (%ds)", opts.Timeout)
	logger.ErrorWithSuggestion(
		"The page took too long to load",
		fmt.Sprintf("snag %s --timeout 60", opts.URL),
	)
	return ErrPageLoadTimeout
}

func fetchCanceled(ctx context.Context, err error) error {
	if ctx != nil && ctx.Err() != nil {
		return ctx.Err()
	}
	if errors.Is(err, context.Canceled) {
		return err
	}
	return nil
}

func (pf *PageFetcher) Fetch(ctx context.Context, opts FetchOptions) (string, error) {
	if pf.page == nil {
		return "", fmt.Errorf("cannot fetch: page is nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}

	logger.Verbose("Fetching %s...", opts.URL)

	waitCtx, cancel := context.WithTimeout(ctx, pf.timeout)
	defer cancel()

	waitIdle, err := pf.page.WaitHTTPIdle(waitCtx, browser.HTTPIdle)
	if err != nil {
		if e := fetchCanceled(ctx, err); e != nil {
			return "", e
		}
		return "", fmt.Errorf("%w: %w", ErrNavigationFailed, err)
	}

	waitLoad, err := pf.page.WaitLoad(waitCtx)
	if err != nil {
		if e := fetchCanceled(ctx, err); e != nil {
			return "", e
		}
		return "", fmt.Errorf("%w: %w", ErrNavigationFailed, err)
	}

	logger.Verbose("Navigating to %s (timeout: %ds)...", opts.URL, opts.Timeout)

	err = pf.page.Navigate(waitCtx, opts.URL)
	if err != nil {
		if e := fetchCanceled(ctx, err); e != nil {
			return "", e
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return "", pageLoadTimeout(opts)
		}
		return "", fmt.Errorf("%w: %w", ErrNavigationFailed, err)
	}

	logger.Verbose("Waiting for page load...")
	if err := waitLoad(); err != nil {
		if e := fetchCanceled(ctx, err); e != nil {
			return "", e
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return "", pageLoadTimeout(opts)
		}
		return "", fmt.Errorf("%w: %w", ErrNavigationFailed, err)
	}

	logger.Verbose("Waiting for HTTP to go idle...")
	if err := waitIdle(); err != nil {
		if e := fetchCanceled(ctx, err); e != nil {
			return "", e
		}
		logger.Verbose("HTTP did not go idle: %v", err)
	}

	if opts.WaitFor != "" {
		err := WaitForSelector(ctx, pf.page, opts.WaitFor, pf.timeout)
		if err != nil {
			if e := fetchCanceled(ctx, err); e != nil {
				return "", e
			}
			if errors.Is(err, context.DeadlineExceeded) {
				logger.ErrorWithSuggestion(
					fmt.Sprintf("Selector not found within %ds", opts.Timeout),
					fmt.Sprintf("snag --wait-for '%s' --timeout 60 %s", opts.WaitFor, opts.URL),
				)
			}
			return "", err
		}
	}

	if authErr := pf.detectAuth(); authErr != nil {
		return "", authErr
	}

	logger.Verbose("Extracting HTML content...")
	html, err := pf.page.HTML()
	if err != nil {
		if e := fetchCanceled(ctx, err); e != nil {
			return "", e
		}
		return "", fmt.Errorf("failed to extract HTML: %w", err)
	}

	logger.Debug("Extracted %d bytes of HTML", len(html))
	logger.Verbose("Fetched successfully")

	return html, nil
}

func (pf *PageFetcher) detectAuth() error {
	if pf.page == nil {
		return fmt.Errorf("cannot detect auth: page is nil")
	}

	status, err := pf.page.NavigationStatus()
	if err != nil {
		logger.Debug("Failed to get HTTP status via JavaScript: %v", err)
	} else if status > 0 {
		logger.Debug("HTTP status code: %d", status)

		if status == 401 || status == 403 {
			logger.Error("Authentication required (HTTP %d)", status)
			logger.ErrorWithSuggestion(
				"This page requires authentication",
				"snag --open-browser "+pf.getURL(),
			)
			return ErrAuthRequired
		}
	}

	hasLogin, err := pf.page.Has("input[type='password']")
	if err == nil && hasLogin {
		hasUsername, _ := pf.page.Has("input[type='text'], input[type='email'], input[name*='user'], input[name*='login']")
		hasSubmit, _ := pf.page.Has("button[type='submit'], input[type='submit']")

		if hasUsername && hasSubmit {
			logger.Debug("Detected login form on page")

			meta, err := pf.page.Meta()
			if err == nil && (strings.Contains(strings.ToLower(meta.Title), "login") ||
				strings.Contains(strings.ToLower(meta.Title), "sign in") ||
				strings.Contains(strings.ToLower(meta.URL), "/login") ||
				strings.Contains(strings.ToLower(meta.URL), "/signin") ||
				strings.Contains(strings.ToLower(meta.URL), "/auth")) {

				logger.Warning("This appears to be a login page")
				logger.ErrorWithSuggestion(
					"Authentication may be required",
					"snag --open-browser "+pf.getURL(),
				)
			}
		}
	}

	return nil
}

func (pf *PageFetcher) getURL() string {
	if pf.page == nil {
		logger.Warning("getURL called with nil page")
		return ""
	}
	meta, err := pf.page.Meta()
	if err != nil {
		return ""
	}
	return meta.URL
}

func WaitForSelector(ctx context.Context, page *browser.Page, selector string, timeout time.Duration) error {
	if page == nil {
		return fmt.Errorf("cannot wait for selector: page is nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	logger.Verbose("Waiting for selector: %s", selector)

	elem, err := page.Element(selector, timeout)
	if err != nil {
		if e := fetchCanceled(ctx, err); e != nil {
			return e
		}
		if errors.Is(err, context.DeadlineExceeded) {
			logger.Error("Timeout waiting for selector: %s", selector)
			return fmt.Errorf("timeout waiting for selector %s: %w", selector, err)
		}
		return fmt.Errorf("failed to find selector %s: %w", selector, err)
	}

	err = elem.WaitVisible()
	if err != nil {
		if e := fetchCanceled(ctx, err); e != nil {
			return e
		}
		if errors.Is(err, context.DeadlineExceeded) {
			logger.Error("Timeout waiting for selector to be visible: %s", selector)
			return fmt.Errorf("timeout waiting for selector %s to be visible: %w", selector, err)
		}
		return fmt.Errorf("selector %s not visible: %w", selector, err)
	}

	logger.Verbose("Selector found: %s", selector)
	return nil
}
