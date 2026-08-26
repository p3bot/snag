// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package browser

import "errors"

var (
	ErrBrowserNotFound   = errors.New("no Chromium-based browser found")
	ErrBrowserConnection = errors.New("failed to connect to browser")
	ErrNoBrowserRunning  = errors.New("no browser instance running with remote debugging")
	ErrTabIndexInvalid   = errors.New("tab index out of range")
	ErrNoTabMatch        = errors.New("no tab matches pattern")
)
