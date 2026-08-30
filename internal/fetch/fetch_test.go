// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package fetch

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/p3bot/snag/internal/browser"
)

func TestFetchCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if !errors.Is(fetchCanceled(ctx, errors.New("other")), context.Canceled) {
		t.Fatal("cancelled ctx should win")
	}
	if !errors.Is(fetchCanceled(context.Background(), context.Canceled), context.Canceled) {
		t.Fatal("Canceled err should win")
	}
	if fetchCanceled(context.Background(), context.DeadlineExceeded) != nil {
		t.Fatal("deadline is not process cancel")
	}
}

func TestFetchAlreadyCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	pf := NewPageFetcher(&browser.Page{}, 1)
	_, err := pf.Fetch(ctx, FetchOptions{URL: "https://example.com", Timeout: 1})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Fetch on cancelled ctx: %v", err)
	}
}

func TestWaitForSelectorAlreadyCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := WaitForSelector(ctx, &browser.Page{}, "body", time.Second)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("WaitForSelector on cancelled ctx: %v", err)
	}
}
