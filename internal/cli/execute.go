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
