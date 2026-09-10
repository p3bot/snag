// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package browser

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/p3bot/snag/internal/logger"
)

func (bm *BrowserManager) KillBrowser(port int) (int, error) {
	if port > 0 {
		return bm.killBrowserOnPort(port)
	}

	return bm.killAllBrowsers()
}

func (bm *BrowserManager) killBrowserOnPort(port int) (int, error) {
	logger.Verbose("Checking port %d...", port)

	cmd := exec.Command("lsof", "-ti", fmt.Sprintf(":%d", port))
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Debug("lsof failed: %v, output: %s", err, string(output))
		logger.Info("No browser running on port %d", port)
		return 0, nil
	}

	pidStr := strings.TrimSpace(string(output))
	if pidStr == "" {
		logger.Info("No browser running on port %d", port)
		return 0, nil
	}

	pidLines := strings.Split(pidStr, "\n")
	pid, err := strconv.Atoi(strings.TrimSpace(pidLines[0]))
	if err != nil {
		return 0, fmt.Errorf("failed to parse PID '%s': %w", pidLines[0], err)
	}

	logger.Verbose("Found browser process (PID %d) on port %d", pid, port)

	killCmd := exec.Command("kill", "-9", fmt.Sprintf("%d", pid))
	if err := killCmd.Run(); err != nil {
		return 0, fmt.Errorf("failed to kill browser process (PID %d): %w", pid, err)
	}

	logger.Success("Killed browser process (PID %d)", pid)
	return 1, nil
}

func (bm *BrowserManager) killAllBrowsers() (int, error) {
	logger.Verbose("Killing all browser processes with remote debugging...")

	// Detect browser path to get executable name
	path, err := bm.findBrowserPath()
	if err != nil {
		return 0, err
	}

	browserExe := filepath.Base(path)
	browserExe = strings.TrimSuffix(browserExe, ".app")

	logger.Debug("Searching for processes matching: %s with --remote-debugging-port", browserExe)

	cmd := exec.Command("ps", "aux")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("failed to list processes: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	var pids []string

	for _, line := range lines {
		if !strings.Contains(line, browserExe) {
			continue
		}
		if !strings.Contains(line, "--remote-debugging-port") {
			continue
		}
		if strings.Contains(line, "grep") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		pid := fields[1]
		pids = append(pids, pid)
		logger.Verbose("  Found PID %s: %s", pid, truncateCommandLine(line, 80))
	}

	if len(pids) == 0 {
		logger.Info("No browser processes found")
		return 0, nil
	}

	logger.Verbose("Killing %d process(es)...", len(pids))

	killedCount := 0
	for _, pid := range pids {
		killCmd := exec.Command("kill", "-9", pid)
		if err := killCmd.Run(); err != nil {
			logger.Warning("Failed to kill PID %s: %v", pid, err)
			continue
		}
		killedCount++
		logger.Debug("Killed PID %s", pid)
	}

	if killedCount > 0 {
		logger.Success("Killed %d process(es)", killedCount)
	}

	return killedCount, nil
}

func truncateCommandLine(line string, maxLen int) string {
	if len(line) <= maxLen {
		return line
	}
	return line[:maxLen] + "..."
}
