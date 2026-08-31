// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package logger

import (
	"fmt"
	"os"
	"strings"
)

// Colour modes for --color.
const (
	ColorAuto   = "auto"
	ColorAlways = "always"
	ColorNever  = "never"
)

func envTruthy(name string) bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(name)))
	return v != "" && v != "0" && v != "false" && v != "no"
}

func stderrIsTTY() bool {
	fi, err := os.Stderr.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

// UseColor applies the stderr colour contract. mode must already be
// auto, always, or never (callers trim and lowercase first).
func UseColor(mode string, stderrTTY bool) bool {
	switch mode {
	case ColorNever:
		return false
	case ColorAlways:
		return true
	case ColorAuto:
		if os.Getenv("NO_COLOR") != "" {
			return false
		}
		if envTruthy("FORCE_COLOR") || envTruthy("CLICOLOR_FORCE") {
			return true
		}
		if os.Getenv("TERM") == "dumb" {
			return false
		}
		return stderrTTY
	default:
		return false
	}
}

// ResolveColor validates mode and reports whether stderr diagnostics should
// use ANSI. The caller installs that bool on a Logger. Mode is trimmed and
// matched case-insensitively.
func ResolveColor(mode string) (bool, error) {
	normalized := strings.ToLower(strings.TrimSpace(mode))
	switch normalized {
	case ColorAuto, ColorAlways, ColorNever:
		return UseColor(normalized, stderrIsTTY()), nil
	default:
		return false, fmt.Errorf("invalid --color value %q (allowed: auto, always, never)", mode)
	}
}

// FormatErrorLine formats the process-level Error: line for stderr.
func FormatErrorLine(err error, color bool) string {
	prefix := "Error:"
	if color {
		prefix = colorRed + "Error:" + colorReset
	}
	return prefix + " " + err.Error()
}
