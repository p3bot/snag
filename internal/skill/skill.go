// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package skill

import (
	_ "embed"
	"strings"
)

//go:embed skill.md
var embedded string

// Text returns the full agent skill contract as markdown, ready for stdout.
// A single trailing newline is guaranteed so writers do not double-terminate.
func Text() string {
	s := embedded
	if !strings.HasSuffix(s, "\n") {
		s += "\n"
	}
	return s
}
