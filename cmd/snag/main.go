// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Command snag fetches web page content using a Chromium-based browser.
// Entry point only: run the command tree and exit.
package main

import (
	"os"

	"github.com/p3bot/snag/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(cli.ExitCodeError)
	}
}
