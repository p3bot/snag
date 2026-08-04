// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
)

type DoctorReport struct {
	SnagVersion   string
	LatestVersion string
	GoVersion     string
	OS            string
	Arch          string
	WorkingDir    string

	BrowserName    string
	BrowserPath    string
	BrowserVersion string
	BrowserError   error

	ProfilePath   string
	ProfileExists bool

	DefaultPortStatus *PortStatus
	CustomPortStatus  *PortStatus // nil if --port not specified

	EnvVars map[string]string
}

// PortStatus contains information about a browser debugging port.
type PortStatus struct {
	Port     int
	Running  bool
	TabCount int
	Error    error
}

func CollectDoctorInfo(customPort int) (*DoctorReport, error) {
	report := &DoctorReport{
		SnagVersion: version,
		GoVersion:   runtime.Version(),
		OS:          runtime.GOOS,
		Arch:        runtime.GOARCH,
		EnvVars:     make(map[string]string),
	}

	// Get working directory
	if wd, err := os.Getwd(); err == nil {
		report.WorkingDir = wd
	} else {
		report.WorkingDir = "(unknown)"
	}

	// Collect environment variables
	report.EnvVars["CHROME_PATH"] = os.Getenv("CHROME_PATH")
	report.EnvVars["CHROMIUM_PATH"] = os.Getenv("CHROMIUM_PATH")

	bm := NewBrowserManager(BrowserOptions{Port: customPort})

	path, err := bm.findBrowserPath()
	if err != nil {
		report.BrowserError = err
	} else {
		report.BrowserPath = path
		report.BrowserName = bm.browserName

		version, err := bm.GetBrowserVersion()
		if err == nil {
			report.BrowserVersion = version
		}

		profilePath, exists := bm.GetProfilePath()
		report.ProfilePath = profilePath
		report.ProfileExists = exists
	}

	report.DefaultPortStatus = checkPortConnection(9222)

	if customPort != 9222 {
		report.CustomPortStatus = checkPortConnection(customPort)
	}

	report.LatestVersion = checkLatestVersion()

	return report, nil
}

func checkPortConnection(port int) *PortStatus {
	status := &PortStatus{
		Port:    port,
		Running: false,
	}

	bm := NewBrowserManager(BrowserOptions{Port: port})

	done := make(chan bool, 1)
	var tabCount int
	var connErr error

	go func() {
		browser, err := bm.connectToExisting()
		if err != nil {
			connErr = err
			done <- false
			return
		}

		pages, err := browser.Pages()
		if err == nil {
			tabCount = len(pages)
		}

		done <- true
	}()

	select {
	case success := <-done:
		if success {
			status.Running = true
			status.TabCount = tabCount
		} else {
			status.Error = connErr
		}
	case <-time.After(3 * time.Second):
		status.Error = fmt.Errorf("connection timeout")
	}

	return status
}

func checkLatestVersion() string {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get("https://api.github.com/repos/p3bot/snag/releases/latest")
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	var release struct {
		TagName string `json:"tag_name"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return ""
	}

	version := strings.TrimPrefix(release.TagName, "v")
	return version
}

func (dr *DoctorReport) String() string {
	var buf strings.Builder

	buf.WriteString("snag Doctor Report\n")
	buf.WriteString("==================\n")
	buf.WriteString("Repository:    https://github.com/p3bot/snag\n")
	buf.WriteString("Report Issue:  https://github.com/p3bot/snag/issues/new\n")

	buf.WriteString(dr.formatSection("Version Information"))
	buf.WriteString(dr.formatItem("snag version", dr.SnagVersion))
	if dr.LatestVersion != "" {
		if dr.LatestVersion != dr.SnagVersion {
			buf.WriteString(dr.formatItem("Latest version", fmt.Sprintf("%s (update available)", dr.LatestVersion)))
		} else {
			buf.WriteString(dr.formatItem("Latest version", dr.LatestVersion))
		}
	}
	buf.WriteString(dr.formatItem("Go version", dr.GoVersion))
	buf.WriteString(dr.formatItem("OS/Arch", fmt.Sprintf("%s/%s", dr.OS, dr.Arch)))

	buf.WriteString(dr.formatSection("Working Directory"))
	buf.WriteString(fmt.Sprintf("  %s\n", dr.WorkingDir))

	buf.WriteString(dr.formatSection("Browser Detection"))
	if dr.BrowserError != nil {
		buf.WriteString(dr.formatCheck("Detected", "No Chromium-based browser found", false))
		buf.WriteString(dr.formatItem("Path", "(none)"))
		buf.WriteString(dr.formatItem("Version", "(none)"))
	} else {
		buf.WriteString(dr.formatItem("Detected", dr.BrowserName))
		buf.WriteString(dr.formatItem("Path", dr.BrowserPath))
		if dr.BrowserVersion != "" {
			buf.WriteString(dr.formatItem("Version", dr.BrowserVersion))
		} else {
			buf.WriteString(dr.formatItem("Version", "(unknown)"))
		}
	}

	if dr.ProfilePath != "" {
		buf.WriteString(dr.formatSection("Profile Location"))
		buf.WriteString(dr.formatCheck(dr.BrowserName, dr.ProfilePath, dr.ProfileExists))
	}

	buf.WriteString(dr.formatSection("Connection Status"))
	if dr.DefaultPortStatus != nil {
		buf.WriteString(dr.formatPortStatus(dr.DefaultPortStatus))
	}
	if dr.CustomPortStatus != nil {
		buf.WriteString(dr.formatPortStatus(dr.CustomPortStatus))
	}

	buf.WriteString(dr.formatSection("Environment Variables"))
	for k, v := range dr.EnvVars {
		if v == "" {
			v = "(not set)"
		}
		buf.WriteString(dr.formatItem(k, v))
	}

	return buf.String()
}

func (dr *DoctorReport) Print() {
	fmt.Print(dr.String())
}

func (dr *DoctorReport) formatPortStatus(status *PortStatus) string {
	label := fmt.Sprintf("Port %d", status.Port)
	if status.Running {
		value := fmt.Sprintf("Running (%d tabs open)", status.TabCount)
		return dr.formatCheck(label, value, true)
	}
	return dr.formatCheck(label, "Not running", false)
}

func (dr *DoctorReport) formatSection(title string) string {
	return fmt.Sprintf("\n%s\n%s\n", title, strings.Repeat("─", len(title)))
}

func (dr *DoctorReport) formatItem(label, value string) string {
	return fmt.Sprintf("  %-20s %s\n", label+":", value)
}

func (dr *DoctorReport) formatCheck(label, value string, ok bool) string {
	mark := "✗"
	if ok {
		mark = "✓"
	}
	return fmt.Sprintf("  %-20s %s %s\n", label+":", mark, value)
}
