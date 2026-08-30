// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cli

import (
	"context"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
)

// caughtSignal: interrupt is intent (POSIX 128+signum), not a retryable failure.
var caughtSignal atomic.Int32

func signalContext() (context.Context, func()) {
	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		select {
		case sig, ok := <-ch:
			if !ok {
				return
			}
			if s, ok := sig.(syscall.Signal); ok {
				caughtSignal.Store(int32(s))
			}
			// Restore default SIGINT/SIGTERM so a second interrupt can kill
			// a stuck shutdown (Notify remains in force until Stop).
			signal.Stop(ch)
			cancel()
		case <-ctx.Done():
		}
	}()
	return ctx, func() {
		signal.Stop(ch)
		cancel()
	}
}

// SignalExitCode returns 128+signum when interrupted, else 0 (checked before error map).
func SignalExitCode() int {
	if s := caughtSignal.Load(); s != 0 {
		return 128 + int(s)
	}
	return 0
}
