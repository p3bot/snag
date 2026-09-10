// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package browser

import (
	"context"
	"errors"
	"time"

	"github.com/p3bot/snag/internal/logger"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

const (
	ConnectTimeout   = 10 * time.Second
	StabilizeTimeout = 3 * time.Second

	defaultRemoteDebugPort = 9222
	extraInstanceDebugPort = 9223

	// headlessWindow is the virtual display and window size for launched
	// headless sessions. Chromium's headless default is 800x600; --screen-info
	// sizes the display (Chrome 142+) and --window-size matches it.
	headlessWindowWidth  = 1920
	headlessWindowHeight = 1080
)

type BrowserManager struct {
	browser          *rod.Browser
	launcher         *launcher.Launcher
	port             int
	wasLaunched      bool
	launchedHeadless bool
	userAgent        string
	userDataDir      string
	tempProfile      bool
	forceHeadless    bool
	openBrowser      bool
	browserName      string
}

type BrowserOptions struct {
	Port          int
	ForceHeadless bool
	OpenBrowser   bool
	UserAgent     string
	UserDataDir   string
	TempProfile   bool
}

func NewBrowserManager(opts BrowserOptions) *BrowserManager {
	return &BrowserManager{
		port:          opts.Port,
		userAgent:     opts.UserAgent,
		userDataDir:   opts.UserDataDir,
		tempProfile:   opts.TempProfile,
		forceHeadless: opts.ForceHeadless,
		openBrowser:   opts.OpenBrowser,
	}
}

func (bm *BrowserManager) Close() {
	if bm.wasLaunched && bm.launchedHeadless {
		logger.Verbose("Closing headless browser...")
		if bm.browser != nil {
			// Cleanup must not use the cancelled process context.
			if err := bm.browser.Context(context.Background()).Close(); err != nil && !errors.Is(err, context.Canceled) {
				logger.Warning("Failed to close browser: %v", err)
			}
		}
		if bm.launcher != nil {
			bm.abandonLauncher(bm.launcher)
		}
		return
	}
	if bm.wasLaunched {
		logger.Verbose("Leaving visible browser running")
		return
	}
	if bm.browser != nil {
		logger.Verbose("Leaving existing browser instance running")
	}
}

func (bm *BrowserManager) ClosePage(page *Page) {
	if page == nil {
		return
	}

	logger.Verbose("Closing page...")
	if err := page.Close(); err != nil && !errors.Is(err, context.Canceled) {
		logger.Warning("Failed to close page: %v", err)
	}
}

func (bm *BrowserManager) WasLaunched() bool {
	return bm.wasLaunched
}

func (bm *BrowserManager) LaunchedHeadless() bool {
	return bm.launchedHeadless
}

func (bm *BrowserManager) Name() string {
	return bm.browserName
}
