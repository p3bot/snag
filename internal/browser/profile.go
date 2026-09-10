// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package browser

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

const (
	launchProfileApp    = "snag"
	launchProfileName   = "chrome"
	chromeSingletonLock = "SingletonLock"
	userDataDirPerm     = 0o700
)

// LaunchProfile describes the Chromium user-data-dir a snag launch would use.
type LaunchProfile struct {
	Path      string
	Ephemeral bool
	Exists    bool
}

// DefaultLaunchProfile is the persistent user-data-dir for unspecified launches.
// Linux: $XDG_STATE_HOME/snag/chrome when XDG_STATE_HOME is non-empty and
// absolute; otherwise $HOME/.local/state/snag/chrome.
// Darwin: $HOME/Library/Application Support/snag/chrome.
func DefaultLaunchProfile() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		dir, err := os.UserConfigDir()
		if err != nil {
			return "", fmt.Errorf("failed to resolve application support directory: %w", err)
		}
		return filepath.Join(dir, launchProfileApp, launchProfileName), nil
	case "linux":
		state, err := linuxStateHome()
		if err != nil {
			return "", err
		}
		return filepath.Join(state, launchProfileApp, launchProfileName), nil
	default:
		return "", fmt.Errorf("launch profile is not supported on %s", runtime.GOOS)
	}
}

// EnsureUserDataDir creates path and any missing parents at owner-only
// permission. These directories hold cookies and session state. It does not
// change the mode of a directory that already exists.
func EnsureUserDataDir(path string) error {
	if err := os.MkdirAll(path, userDataDirPerm); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	return nil
}

func linuxStateHome() (string, error) {
	xdg := os.Getenv("XDG_STATE_HOME")
	if xdg != "" && filepath.IsAbs(xdg) {
		return xdg, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to resolve home directory: %w", err)
	}
	return filepath.Join(home, ".local", "state"), nil
}

// ResolveLaunchProfile returns the directory a launch would pass as
// --user-data-dir. It does not create directories. Empty userDataDir means
// the XDG/darwin default. tempProfile is a unique temp dir at launch time.
func ResolveLaunchProfile(userDataDir string, tempProfile bool) (LaunchProfile, error) {
	if tempProfile {
		return LaunchProfile{Ephemeral: true}, nil
	}
	path := userDataDir
	if path == "" {
		var err error
		path, err = DefaultLaunchProfile()
		if err != nil {
			return LaunchProfile{}, err
		}
	}
	info, err := os.Stat(path)
	exists := err == nil && info.IsDir()
	return LaunchProfile{Path: path, Exists: exists}, nil
}

// profileLockPresent reports Chrome's in-use marker. SingletonLock is often a
// dangling symlink; Lstat so a missing target still counts.
func profileLockPresent(dir string) bool {
	if dir == "" {
		return false
	}
	_, err := os.Lstat(filepath.Join(dir, chromeSingletonLock))
	return err == nil
}
