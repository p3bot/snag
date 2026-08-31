// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cli

import (
	"io"
	"strings"
	"testing"

	"github.com/p3bot/snag/internal/logger"
)

func TestFormatError(t *testing.T) {
	t.Cleanup(func() { colorMode = logger.ColorAuto })
	t.Setenv("NO_COLOR", "")
	t.Setenv("FORCE_COLOR", "")
	t.Setenv("CLICOLOR_FORCE", "")
	t.Setenv("TERM", "dumb")

	if got := FormatError(nil); got != "" {
		t.Errorf("FormatError(nil) = %q, want empty", got)
	}

	colorMode = logger.ColorAlways
	got := FormatError(io.EOF)
	if !strings.Contains(got, "\x1b[") {
		t.Errorf("FormatError always missing ANSI: %q", got)
	}
	if !strings.Contains(got, "Error:") || !strings.Contains(got, "EOF") {
		t.Errorf("FormatError always = %q, want Error: prefix and message", got)
	}

	colorMode = logger.ColorNever
	got = FormatError(io.EOF)
	if strings.Contains(got, "\x1b[") {
		t.Errorf("FormatError never has ANSI: %q", got)
	}
	if got != "Error: EOF" {
		t.Errorf("FormatError never = %q, want %q", got, "Error: EOF")
	}

	colorMode = "rainbow"
	got = FormatError(io.EOF)
	if strings.Contains(got, "\x1b[") {
		t.Errorf("FormatError invalid mode with TERM=dumb has ANSI: %q", got)
	}
	if got != "Error: EOF" {
		t.Errorf("FormatError invalid mode = %q, want %q", got, "Error: EOF")
	}
}
