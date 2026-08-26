// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cli

import "errors"

var (
	ErrTabURLConflict     = errors.New("cannot use both --tab and URL arguments")
	ErrOutputFlagConflict = errors.New("--output cannot be used with multiple content sources, use --output-dir instead")
)
