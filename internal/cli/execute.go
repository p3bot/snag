// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cli

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// Execute runs the command tree. A signal closes the browser and exits
// immediately with POSIX 128+signum.
func Execute() error {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		fmt.Fprintf(os.Stderr, "\nReceived %v, cleaning up...\n", sig)

		browserMutex.Lock()
		if browserManager != nil {
			browserManager.Close()
		}
		browserMutex.Unlock()

		if sig == os.Interrupt {
			os.Exit(ExitCodeInterrupt)
		}
		os.Exit(ExitCodeSIGTERM)
	}()

	return rootCmd.Execute()
}
