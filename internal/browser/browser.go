// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package browser

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/p3bot/snag/internal/logger"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/devices"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

const (
	ConnectTimeout   = 10 * time.Second
	StabilizeTimeout = 3 * time.Second
)

type BrowserManager struct {
	browser          *rod.Browser
	launcher         *launcher.Launcher
	port             int
	wasLaunched      bool
	launchedHeadless bool
	userAgent        string
	userDataDir      string
	forceHeadless    bool
	openBrowser      bool
	browserName      string
}

type BrowserOptions struct {
	Port          int
	ForceHeadless bool
	OpenBrowser   bool
	UserAgent     string
	UserDataDir   string
}

type TabInfo struct {
	Index int
	URL   string
	Title string
	ID    string
}

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

func NewBrowserManager(opts BrowserOptions) *BrowserManager {
	return &BrowserManager{
		port:          opts.Port,
		userAgent:     opts.UserAgent,
		userDataDir:   opts.UserDataDir,
		forceHeadless: opts.ForceHeadless,
		openBrowser:   opts.OpenBrowser,
	}
}

func (bm *BrowserManager) Connect() error {
	if !bm.forceHeadless {
		logger.Verbose("Checking for existing browser instance on port %d...", bm.port)
		if browser, err := bm.connectToExisting(); err == nil {
			if bm.openBrowser {
				logger.Success("Connected to existing browser (visible mode)")
			} else {
				logger.Success("Connected to existing browser instance")
			}
			if bm.userDataDir != "" {
				logger.Warning("--user-data-dir ignored (browser already running with its own profile)")
			}
			if bm.userAgent != "" {
				logger.Warning("--user-agent ignored (browser already running with its own user agent)")
			}
			bm.browser = browser
			bm.wasLaunched = false
			return nil
		}
		logger.Verbose("No existing browser instance found")
	}

	headless := bm.forceHeadless || !bm.openBrowser

	if headless {
		logger.Verbose("Launching browser in headless mode...")
	} else {
		logger.Info("Launching browser in visible mode...")
	}

	browser, err := bm.launchBrowser(headless)
	if err != nil {
		return err
	}

	if headless {
		logger.Success("%s launched in headless mode", bm.browserName)
	} else {
		logger.Success("%s launched in visible mode", bm.browserName)
	}

	bm.browser = browser
	bm.wasLaunched = true
	bm.launchedHeadless = headless
	return nil
}

// ConnectExisting attaches to a browser already listening on the debug port.
func (bm *BrowserManager) ConnectExisting() error {
	b, err := bm.connectToExisting()
	if err != nil {
		return err
	}
	bm.browser = b
	return nil
}

func (bm *BrowserManager) connectToExisting() (*rod.Browser, error) {
	baseURL := fmt.Sprintf("http://127.0.0.1:%d", bm.port)
	logger.Debug("Attempting connection to: %s", baseURL)

	wsURL, err := launcher.ResolveURL(baseURL)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrBrowserConnection, err)
	}
	logger.Debug("Resolved WebSocket URL: %s", wsURL)

	browser := rod.New().ControlURL(wsURL).Timeout(ConnectTimeout)

	if err := browser.Connect(); err != nil {
		logger.Debug("Connection failed: %v", err)
		return nil, fmt.Errorf("%w: %w", ErrBrowserConnection, err)
	}
	logger.Debug("Successfully connected to browser")

	return browser.CancelTimeout(), nil
}

func (bm *BrowserManager) launchBrowser(headless bool) (*rod.Browser, error) {
	path, err := bm.findBrowserPath()
	if err != nil {
		return nil, err
	}

	l := launcher.New().
		Bin(path).
		Headless(headless).
		Leakless(headless).
		Set("disable-blink-features", "AutomationControlled")

	if bm.userAgent != "" {
		l = l.Set("user-agent", bm.userAgent)
		logger.Verbose("Using custom user agent: %s", bm.userAgent)
	}

	if bm.userDataDir != "" {
		l = l.Set("user-data-dir", bm.userDataDir)
		logger.Verbose("Using custom user data directory: %s", bm.userDataDir)
	}

	l = l.Set("remote-debugging-port", fmt.Sprintf("%d", bm.port))

	controlURL, err := l.Launch()
	if err != nil {
		return nil, fmt.Errorf("failed to launch browser: %w", err)
	}
	logger.Debug("Browser launched with control URL: %s", controlURL)

	bm.launcher = l

	browser := rod.New().ControlURL(controlURL).Timeout(ConnectTimeout)

	if err := browser.Connect(); err != nil {
		logger.Debug("Failed to connect to launched browser: %v", err)
		return nil, fmt.Errorf("%w: %w", ErrBrowserConnection, err)
	}
	logger.Debug("Successfully connected to launched browser")

	return browser.CancelTimeout(), nil
}

func (bm *BrowserManager) OpenBrowserOnly() error {
	logger.Verbose("Checking for existing browser instance on port %d...", bm.port)
	if _, err := bm.connectToExisting(); err == nil {
		logger.Success("Browser already running on port %d", bm.port)
		if bm.userDataDir != "" {
			logger.Warning("--user-data-dir ignored (browser already running with its own profile)")
		}
		if bm.userAgent != "" {
			logger.Warning("--user-agent ignored (browser already running with its own user agent)")
		}
		logger.Info("You can connect to it using: snag <url>")
		return nil
	}

	path, err := bm.findBrowserPath()
	if err != nil {
		return err
	}

	l := launcher.New().
		Bin(path).
		Leakless(false).
		Headless(false).
		Set("disable-blink-features", "AutomationControlled").
		Set("remote-debugging-port", fmt.Sprintf("%d", bm.port))

	if bm.userAgent != "" {
		l = l.Set("user-agent", bm.userAgent)
		logger.Verbose("Using custom user agent: %s", bm.userAgent)
	}

	if bm.userDataDir != "" {
		l = l.Set("user-data-dir", bm.userDataDir)
		logger.Verbose("Using custom user data directory: %s", bm.userDataDir)
	}

	controlURL, err := l.Launch()
	if err != nil {
		return fmt.Errorf("failed to launch browser: %w", err)
	}

	browser := rod.New().ControlURL(controlURL).Timeout(ConnectTimeout)
	if err := browser.Connect(); err != nil {
		return fmt.Errorf("%w: %w", ErrBrowserConnection, err)
	}

	_, err = browser.Page(proto.TargetCreateTarget{URL: "about:blank"})
	if err != nil {
		browser.Close()
		return fmt.Errorf("failed to create page: %w", err)
	}

	logger.Success("Browser opened on port %d", bm.port)
	logger.Info("Browser is running with remote debugging enabled")
	logger.Info("You can now connect to it using: snag <url>")

	return nil
}

func (bm *BrowserManager) NewPage() (*Page, error) {
	if bm.browser == nil {
		return nil, fmt.Errorf("browser not connected")
	}

	page, err := bm.browser.Page(proto.TargetCreateTarget{})
	if err != nil {
		return nil, fmt.Errorf("failed to create page: %w", err)
	}

	if bm.launchedHeadless {
		// Set a sensible default viewport for headless mode (1920x1080 Full HD)
		err := page.SetViewport(&proto.EmulationSetDeviceMetricsOverride{
			Width:             1920,
			Height:            1080,
			DeviceScaleFactor: 1,
			Mobile:            false,
		})
		if err != nil {
			logger.Debug("Failed to set headless viewport: %v", err)
		}
	} else {
		// Clear default viewport emulation so page fills the browser window
		if err := page.Emulate(devices.Clear); err != nil {
			logger.Debug("Failed to clear viewport emulation: %v", err)
		}
	}

	return wrapPage(page), nil
}

func (bm *BrowserManager) Close() {
	if bm.browser == nil {
		return
	}

	if bm.wasLaunched && bm.launchedHeadless {
		logger.Verbose("Closing headless browser...")
		if err := bm.browser.Close(); err != nil {
			logger.Warning("Failed to close browser: %v", err)
		}

		if bm.launcher != nil {
			bm.launcher.Kill()
			if bm.userDataDir == "" {
				bm.launcher.Cleanup()
			}
		}
	} else if bm.wasLaunched && !bm.launchedHeadless {
		logger.Verbose("Leaving visible browser running")
	} else {
		logger.Verbose("Leaving existing browser instance running")
	}
}

func (bm *BrowserManager) ClosePage(page *Page) {
	if page == nil {
		return
	}

	logger.Verbose("Closing page...")
	if err := page.Close(); err != nil {
		logger.Warning("Failed to close page: %v", err)
	}
}

func (bm *BrowserManager) WasLaunched() bool {
	return bm.wasLaunched
}

func (bm *BrowserManager) LaunchedHeadless() bool {
	return bm.launchedHeadless
}

func (bm *BrowserManager) Name() string {
	return bm.browserName
}

func (bm *BrowserManager) FindBrowserPath() (string, error) {
	return bm.findBrowserPath()
}

// ProbePort reports how many page tabs are open on a debugging port.
// It uses Chrome's HTTP /json/list so it does not attach a DevTools session
// (Rod Close() would send Browser.close and quit the user's Chrome).
func ProbePort(port int) (tabCount int, err error) {
	client := &http.Client{Timeout: ConnectTimeout}
	url := fmt.Sprintf("http://127.0.0.1:%d/json/list", port)
	resp, err := client.Get(url)
	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrBrowserConnection, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("%w: GET %s: %s", ErrBrowserConnection, url, resp.Status)
	}

	var targets []struct {
		Type string `json:"type"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&targets); err != nil {
		return 0, fmt.Errorf("%w: decode tab list: %w", ErrBrowserConnection, err)
	}

	n := 0
	for _, t := range targets {
		if t.Type == "page" {
			n++
		}
	}
	return n, nil
}

type pageWithInfo struct {
	page  *rod.Page
	url   string
	title string
	id    string
}

func (bm *BrowserManager) getSortedPagesWithInfo() ([]pageWithInfo, error) {
	if bm.browser == nil {
		return nil, ErrNoBrowserRunning
	}

	pages, err := bm.browser.Pages()
	if err != nil {
		return nil, fmt.Errorf("failed to get pages: %w", err)
	}

	pagesWithInfo := make([]pageWithInfo, 0, len(pages))
	for i, page := range pages {
		info, err := page.Info()
		if err != nil {
			logger.Warning("Failed to get info for tab at position %d (will be excluded from list): %v", i+1, err)
			logger.Debug("Tab page object: %+v", page)
			continue
		}
		pagesWithInfo = append(pagesWithInfo, pageWithInfo{
			page:  page,
			url:   info.URL,
			title: info.Title,
			id:    string(page.TargetID),
		})
	}

	if len(pagesWithInfo) < len(pages) {
		excluded := len(pages) - len(pagesWithInfo)
		logger.Warning("Excluded %d tab(s) due to inaccessible page info", excluded)
	}

	sort.Slice(pagesWithInfo, func(i, j int) bool {
		if pagesWithInfo[i].url != pagesWithInfo[j].url {
			return pagesWithInfo[i].url < pagesWithInfo[j].url
		}
		if pagesWithInfo[i].title != pagesWithInfo[j].title {
			return pagesWithInfo[i].title < pagesWithInfo[j].title
		}
		return pagesWithInfo[i].id < pagesWithInfo[j].id
	})

	return pagesWithInfo, nil
}

func (bm *BrowserManager) ListTabs() ([]TabInfo, error) {
	pagesWithInfo, err := bm.getSortedPagesWithInfo()
	if err != nil {
		return nil, err
	}

	tabs := make([]TabInfo, len(pagesWithInfo))
	for i, pwi := range pagesWithInfo {
		tabs[i] = TabInfo{
			Index: i + 1,
			URL:   pwi.url,
			Title: pwi.title,
			ID:    pwi.id,
		}
	}

	return tabs, nil
}

func (bm *BrowserManager) GetTabByIndex(index int) (*Page, error) {
	pagesWithInfo, err := bm.getSortedPagesWithInfo()
	if err != nil {
		return nil, err
	}

	if index < 1 || index > len(pagesWithInfo) {
		return nil, fmt.Errorf("%w: tab index %d (valid range: 1-%d)", ErrTabIndexInvalid, index, len(pagesWithInfo))
	}

	arrayIndex := index - 1

	logger.Verbose("Selected tab [%d] from sorted order: %s", index, pagesWithInfo[arrayIndex].url)

	return wrapPage(pagesWithInfo[arrayIndex].page), nil
}

func (bm *BrowserManager) GetTabsByPattern(pattern string) ([]*Page, error) {
	pagesWithInfo, err := bm.getSortedPagesWithInfo()
	if err != nil {
		return nil, err
	}

	if len(pagesWithInfo) == 0 {
		return nil, fmt.Errorf("%w: '%s' (no tabs open)", ErrNoTabMatch, pattern)
	}

	logger.Debug("Matching pattern '%s' against %d tabs", pattern, len(pagesWithInfo))
	patternLower := strings.ToLower(pattern)

	var exactMatches []*Page
	for i, pwi := range pagesWithInfo {
		if strings.EqualFold(pwi.url, pattern) {
			logger.Verbose("Matched tab [%d] via exact URL: %s", i+1, pwi.url)
			exactMatches = append(exactMatches, wrapPage(pwi.page))
		}
	}
	if len(exactMatches) > 0 {
		return exactMatches, nil
	}

	var substringMatches []*Page
	for i, pwi := range pagesWithInfo {
		if strings.Contains(strings.ToLower(pwi.url), patternLower) {
			logger.Verbose("Matched tab [%d] via substring: %s", i+1, pwi.url)
			substringMatches = append(substringMatches, wrapPage(pwi.page))
		}
	}
	if len(substringMatches) > 0 {
		return substringMatches, nil
	}

	re, err := regexp.Compile("(?i)" + pattern)
	if err != nil {
		logger.Debug("Pattern is not valid regex: %v", err)
		return nil, fmt.Errorf("invalid regex pattern '%s': %w", pattern, err)
	}

	var regexMatches []*Page
	for i, pwi := range pagesWithInfo {
		if re.MatchString(pwi.url) {
			logger.Verbose("Matched tab [%d] via regex: %s", i+1, pwi.url)
			regexMatches = append(regexMatches, wrapPage(pwi.page))
		}
	}
	if len(regexMatches) > 0 {
		logger.Debug("Found %d regex matches for pattern '%s'", len(regexMatches), pattern)
		return regexMatches, nil
	}

	return nil, fmt.Errorf("%w: '%s'", ErrNoTabMatch, pattern)
}

func (bm *BrowserManager) GetTabsByRange(start, end int) ([]*Page, error) {
	pagesWithInfo, err := bm.getSortedPagesWithInfo()
	if err != nil {
		return nil, err
	}

	if start < 1 {
		return nil, fmt.Errorf("tab range must start from 1 (got %d)", start)
	}
	if start > end {
		return nil, fmt.Errorf("invalid range: start must be <= end (got %d-%d)", start, end)
	}

	if start > len(pagesWithInfo) {
		return nil, fmt.Errorf("tab index %d out of range in range %d-%d (only %d tabs open)", start, start, end, len(pagesWithInfo))
	}
	if end > len(pagesWithInfo) {
		return nil, fmt.Errorf("tab index %d out of range in range %d-%d (only %d tabs open)", end, start, end, len(pagesWithInfo))
	}

	rangeWithInfo := pagesWithInfo[start-1 : end]

	rangeTabs := make([]*Page, len(rangeWithInfo))
	for i, pwi := range rangeWithInfo {
		rangeTabs[i] = wrapPage(pwi.page)
	}

	logger.Verbose("Selected %d tabs from sorted range [%d-%d]", len(rangeTabs), start, end)
	return rangeTabs, nil
}

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

func truncateCommandLine(line string, maxLen int) string {
	if len(line) <= maxLen {
		return line
	}
	return line[:maxLen] + "..."
}
