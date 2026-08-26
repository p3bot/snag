// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package validate

import "errors"

var (
	ErrInvalidURL  = errors.New("invalid URL")
	ErrNoValidURLs = errors.New("no valid URLs provided")
)
