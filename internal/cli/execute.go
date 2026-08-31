// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cli

import "github.com/p3bot/snag/internal/logger"

// Execute runs the command tree. SIGINT/SIGTERM cancel the root context;
// main maps that to POSIX 128+signum via SignalExitCode.
func Execute() error {
	ctx, stop := signalContext()
	defer stop()
	err := rootCmd.ExecuteContext(ctx)
	if SignalExitCode() != 0 {
		logger.Verbose("cancelled")
	}
	return err
}

// FormatError formats the process-level Error: line using the parsed --color
// value. Invalid --color falls back to auto so the rejection itself is not
// painted with the bad mode.
func FormatError(err error) string {
	if err == nil {
		return ""
	}
	useColor, rerr := logger.ResolveColor(colorMode)
	if rerr != nil {
		useColor, _ = logger.ResolveColor(logger.ColorAuto)
	}
	return logger.FormatErrorLine(err, useColor)
}
