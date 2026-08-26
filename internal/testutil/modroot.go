// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package testutil locates the module root and shared testdata for tests
// that no longer run with CWD at the repository root.
package testutil

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

var (
	rootOnce sync.Once
	rootDir  string
	rootErr  error
)

// ModuleRoot walks parents from CWD until it finds go.mod.
func ModuleRoot() string {
	rootOnce.Do(func() {
		dir, err := os.Getwd()
		if err != nil {
			rootErr = fmt.Errorf("getwd: %w", err)
			return
		}
		for {
			if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
				rootDir = dir
				return
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				rootErr = fmt.Errorf("go.mod not found from %s", dir)
				return
			}
			dir = parent
		}
	})
	if rootErr != nil {
		panic(rootErr)
	}
	return rootDir
}

// Testdata joins names under the module-root testdata directory.
func Testdata(elem ...string) string {
	parts := append([]string{ModuleRoot(), "testdata"}, elem...)
	return filepath.Join(parts...)
}

// BuildSnag builds ./cmd/snag into dest.
func BuildSnag(dest string) error {
	cmd := exec.Command("go", "build", "-o", dest, "./cmd/snag")
	cmd.Dir = ModuleRoot()
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("go build ./cmd/snag: %w\n%s", err, out)
	}
	return nil
}
