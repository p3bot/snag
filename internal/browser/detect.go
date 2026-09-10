// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package browser

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/p3bot/snag/internal/logger"

	"github.com/go-rod/rod/lib/launcher"
)

func (bm *BrowserManager) findBrowserPath() (string, error) {
	path, exists := launcher.LookPath()
	if !exists {
		return "", ErrBrowserNotFound
	}

	bm.browserName = detectBrowserName(path)

	logger.Debug("Found browser at: %s", path)

	return path, nil
}

type browserDetectionRule struct {
	pattern          string
	name             string
	exclude          string
	profilePathMac   string
	profilePathLinux string
}

var browserDetectionRules = []browserDetectionRule{
	{"ungoogled", "Ungoogled-Chromium", "", "Chromium", "chromium"},
	{"chrome", "Chrome", "chromium", "Google/Chrome", "google-chrome"},
	{"chromium", "Chromium", "", "Chromium", "chromium"},
	{"msedge", "Edge", "", "Microsoft Edge", "microsoft-edge"},
	{"edge", "Edge", "", "Microsoft Edge", "microsoft-edge"},
	{"brave", "Brave", "", "BraveSoftware/Brave-Browser", "BraveSoftware/Brave-Browser"},
	{"opera", "Opera", "", "com.operasoftware.Opera", "opera"},
	{"vivaldi", "Vivaldi", "", "Vivaldi", "vivaldi"},
	{"arc", "Arc", "", "Arc", ""},
	{"yandex", "Yandex", "", "Yandex/YandexBrowser", "yandex-browser"},
	{"thorium", "Thorium", "", "Thorium", "thorium"},
	{"slimjet", "Slimjet", "", "Slimjet", "slimjet"},
	{"cent", "Cent", "", "CentBrowser", "cent-browser"},
}

func detectBrowserName(path string) string {
	base := filepath.Base(path)
	baseName := strings.TrimSuffix(base, ".exe")
	baseName = strings.TrimSuffix(baseName, ".app")
	lowerName := strings.ToLower(baseName)

	for _, rule := range browserDetectionRules {
		if strings.Contains(lowerName, rule.pattern) {
			if rule.exclude != "" && strings.Contains(lowerName, rule.exclude) {
				continue
			}
			return rule.name
		}
	}

	if len(baseName) > 0 {
		return strings.ToUpper(baseName[:1]) + baseName[1:]
	}

	return "Browser"
}

func (bm *BrowserManager) FindBrowserPath() (string, error) {
	return bm.findBrowserPath()
}

func (bm *BrowserManager) GetBrowserVersion() (string, error) {
	path, err := bm.findBrowserPath()
	if err != nil {
		return "", err
	}

	cmd := exec.Command(path, "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get browser version: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

func (bm *BrowserManager) GetProfilePath() (string, bool) {
	path, err := bm.findBrowserPath()
	if err != nil {
		return "", false
	}

	baseName := strings.ToLower(filepath.Base(path))
	baseName = strings.TrimSuffix(baseName, ".exe")
	baseName = strings.TrimSuffix(baseName, ".app")

	var matchedRule *browserDetectionRule
	for i := range browserDetectionRules {
		if strings.Contains(baseName, browserDetectionRules[i].pattern) {
			if browserDetectionRules[i].exclude != "" && strings.Contains(baseName, browserDetectionRules[i].exclude) {
				continue
			}
			matchedRule = &browserDetectionRules[i]
			break
		}
	}

	if matchedRule == nil {
		return "", false
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", false
	}

	var profilePath string
	if runtime.GOOS == "darwin" {
		if matchedRule.profilePathMac == "" {
			return "", false
		}
		profilePath = filepath.Join(home, "Library", "Application Support", matchedRule.profilePathMac)
	} else {
		if matchedRule.profilePathLinux == "" {
			return "", false
		}
		profilePath = filepath.Join(home, ".config", matchedRule.profilePathLinux)
	}

	_, err = os.Stat(profilePath)
	exists := err == nil

	return profilePath, exists
}
