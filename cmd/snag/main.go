// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Command snag fetches web page content using a Chromium-based browser.
// Entry point only: run the command tree and exit.
package main

import (
	"fmt"
	"os"

	"github.com/p3bot/snag/internal/cli"
)

func main() {
	err := cli.Execute()
	if code := cli.SignalExitCode(); code != 0 {
		os.Exit(code)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, cli.FormatError(err))
		os.Exit(cli.ExitCodeError)
	}
}
