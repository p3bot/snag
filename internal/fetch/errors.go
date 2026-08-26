// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package fetch

import "errors"

var (
	ErrPageLoadTimeout  = errors.New("page load timeout exceeded")
	ErrAuthRequired     = errors.New("authentication required")
	ErrNavigationFailed = errors.New("page navigation failed")
)
