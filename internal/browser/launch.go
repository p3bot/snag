// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package browser

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/p3bot/snag/internal/logger"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

func (bm *BrowserManager) Connect(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if !bm.forceHeadless {
		logger.Verbose("Checking for existing browser instance on port %d...", bm.port)
		browser, err := bm.connectToExisting(ctx)
		if err == nil {
			if bm.openBrowser {
				logger.Verbose("Connected to existing browser (visible mode)")
			} else {
				logger.Verbose("Connected to existing browser instance")
			}
			bm.warnIgnoredLaunchFlags()
			if bm.userAgent != "" {
				logger.Warning("--user-agent ignored (browser already running with its own user agent)")
			}
			bm.browser = browser
			bm.wasLaunched = false
			return nil
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		logger.Verbose("No existing browser instance found")
	}

	headless := bm.forceHeadless || !bm.openBrowser

	if headless {
		logger.Verbose("Launching browser in headless mode...")
	} else {
		logger.Verbose("Launching browser in visible mode...")
	}

	browser, err := bm.launchBrowser(ctx, headless)
	if err != nil {
		return err
	}

	if headless {
		logger.Verbose("%s launched in headless mode", bm.browserName)
	} else {
		logger.Verbose("%s launched in visible mode", bm.browserName)
	}

	bm.browser = browser
	bm.wasLaunched = true
	bm.launchedHeadless = headless
	return nil
}

// ConnectExisting attaches to a browser already listening on the debug port.
func (bm *BrowserManager) ConnectExisting(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	b, err := bm.connectToExisting(ctx)
	if err != nil {
		return err
	}
	bm.browser = b
	return nil
}

func (bm *BrowserManager) connectToExisting(ctx context.Context) (*rod.Browser, error) {
	baseURL := fmt.Sprintf("http://127.0.0.1:%d", bm.port)
	logger.Debug("Attempting connection to: %s", baseURL)

	wsURL, err := resolveWSURL(ctx, bm.port)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrBrowserConnection, err)
	}
	logger.Debug("Resolved WebSocket URL: %s", wsURL)

	browser, err := connectRod(ctx, wsURL)
	if err != nil {
		logger.Debug("Connection failed: %v", err)
		return nil, fmt.Errorf("%w: %w", ErrBrowserConnection, err)
	}
	logger.Debug("Successfully connected to browser")

	return browser, nil
}

func (bm *BrowserManager) launchBrowser(ctx context.Context, headless bool) (*rod.Browser, error) {
	path, err := bm.findBrowserPath()
	if err != nil {
		return nil, err
	}

	l := withRealUserLaunchFlags(launcher.New().Context(ctx).
		Bin(path).
		Leakless(headless), headless)

	if bm.userAgent != "" {
		l = l.Set("user-agent", bm.userAgent)
		logger.Verbose("Using custom user agent: %s", bm.userAgent)
	}

	l, err = bm.applyLaunchProfile(l)
	if err != nil {
		return nil, err
	}

	l = l.Set("remote-debugging-port", fmt.Sprintf("%d", bm.port))

	controlURL, err := l.Launch()
	if err != nil {
		return nil, bm.wrapLaunchError(err)
	}
	logger.Debug("Browser launched with control URL: %s", controlURL)

	bm.launcher = l
	bm.wasLaunched = true
	bm.launchedHeadless = headless

	browser, err := connectRod(ctx, controlURL)
	if err != nil {
		logger.Debug("Failed to connect to launched browser: %v", err)
		// Close() does not kill a visible browser, so a failed start must
		// tear down here (interrupt and CDP failure both land on this path).
		bm.abandonLauncher(l)
		bm.launcher = nil
		bm.wasLaunched = false
		bm.launchedHeadless = false
		return nil, fmt.Errorf("%w: %w", ErrBrowserConnection, err)
	}
	logger.Debug("Successfully connected to launched browser")

	return browser, nil
}

func (bm *BrowserManager) OpenBrowserOnly(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	logger.Verbose("Checking for existing browser instance on port %d...", bm.port)
	if _, err := bm.connectToExisting(ctx); err == nil {
		logger.Success("Browser already running on port %d", bm.port)
		bm.warnIgnoredLaunchFlags()
		if bm.userAgent != "" {
			logger.Warning("--user-agent ignored (browser already running with its own user agent)")
		}
		logger.Info("You can connect to it using: snag <url>")
		return nil
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	path, err := bm.findBrowserPath()
	if err != nil {
		return err
	}

	l := withRealUserLaunchFlags(launcher.New().Context(ctx).
		Bin(path).
		Leakless(false), false).
		Set("remote-debugging-port", fmt.Sprintf("%d", bm.port))

	if bm.userAgent != "" {
		l = l.Set("user-agent", bm.userAgent)
		logger.Verbose("Using custom user agent: %s", bm.userAgent)
	}

	l, err = bm.applyLaunchProfile(l)
	if err != nil {
		return err
	}

	controlURL, err := l.Launch()
	if err != nil {
		return bm.wrapLaunchError(err)
	}

	abandon := func() {
		bm.abandonLauncher(l)
	}

	browser, err := connectRod(ctx, controlURL)
	if err != nil {
		abandon()
		return fmt.Errorf("%w: %w", ErrBrowserConnection, err)
	}

	_, err = browser.Page(proto.TargetCreateTarget{URL: "about:blank"})
	if err != nil {
		_ = browser.Context(context.Background()).Close()
		abandon()
		return fmt.Errorf("failed to create page: %w", err)
	}

	logger.Success("Browser opened on port %d", bm.port)
	logger.Info("Browser is running with remote debugging enabled")
	logger.Info("You can now connect to it using: snag <url>")

	return nil
}

func (bm *BrowserManager) warnIgnoredLaunchFlags() {
	if bm.userDataDir != "" {
		logger.Warning("--user-data-dir ignored (browser already running with its own profile)")
	}
	if bm.tempProfile {
		logger.Warning("--temp-profile ignored (browser already running with its own profile)")
	}
}

func (bm *BrowserManager) applyLaunchProfile(l *launcher.Launcher) (*launcher.Launcher, error) {
	if bm.tempProfile && bm.userDataDir != "" {
		return l, fmt.Errorf("user-data-dir and temp-profile are mutually exclusive")
	}
	if bm.tempProfile {
		logger.Verbose("Using ephemeral user data directory")
		return l, nil
	}

	dir := bm.userDataDir
	if dir == "" {
		var err error
		dir, err = DefaultLaunchProfile()
		if err != nil {
			return l, fmt.Errorf("failed to resolve launch profile: %w", err)
		}
		bm.userDataDir = dir
	}

	created := false
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		logger.Verbose("Creating user data directory: %s", dir)
		created = true
	}

	if err := EnsureUserDataDir(dir); err != nil {
		logger.Error("Failed to create user data directory: %s", dir)
		logger.ErrorWithSuggestion(
			"Cannot create user data directory",
			fmt.Sprintf("mkdir -p %s", dir),
		)
		return l, err
	}
	if created {
		logger.Verbose("User data directory created: %s", dir)
	}

	l = l.Set("user-data-dir", dir)
	logger.Verbose("Using user data directory: %s", dir)
	return l, nil
}

func extraInstanceSuggestion(port int) string {
	if port <= 0 || port == defaultRemoteDebugPort {
		port = extraInstanceDebugPort
	}
	return fmt.Sprintf("snag --temp-profile --force-headless --port %d <url>", port)
}

func (bm *BrowserManager) wrapLaunchError(err error) error {
	err = fmt.Errorf("failed to launch browser: %w", err)
	if bm.tempProfile || !profileLockPresent(bm.userDataDir) {
		return err
	}
	logger.ErrorWithSuggestion(
		"Launch profile is in use by another Chrome process",
		extraInstanceSuggestion(bm.port),
	)
	return fmt.Errorf("%w: %s: %w", ErrProfileInUse, bm.userDataDir, err)
}

func (bm *BrowserManager) abandonLauncher(l *launcher.Launcher) {
	if l == nil {
		return
	}
	l.Kill()
	if bm.tempProfile {
		l.Cleanup()
	}
}

func (bm *BrowserManager) NewPage() (*Page, error) {
	if bm.browser == nil {
		return nil, fmt.Errorf("browser not connected")
	}

	page, err := bm.browser.Page(proto.TargetCreateTarget{})
	if err != nil {
		return nil, fmt.Errorf("failed to create page: %w", err)
	}

	if err := applyRealUserPageIdentity(page, bm.browserName); err != nil {
		logger.Warning("Failed to apply real Chrome identity: %v", err)
	}
	if bm.launchedHeadless {
		if err := applyHeadlessWindowIdentity(page); err != nil {
			logger.Warning("Failed to apply headless window identity: %v", err)
		}
	}

	return wrapPage(page), nil
}

// connectRod attaches to a CDP endpoint without Rod's default laptop device
// (Mac Chrome 114 user agent and 1280x800 viewport).
func connectRod(ctx context.Context, controlURL string) (*rod.Browser, error) {
	browser := rod.New().
		NoDefaultDevice().
		Context(ctx).
		ControlURL(controlURL).
		Timeout(ConnectTimeout)
	if err := browser.Connect(); err != nil {
		return nil, err
	}
	return browser.CancelTimeout(), nil
}

// withRealUserLaunchFlags drops Chromium automation tells so a launched
// browser presents as a normal desktop Chrome: no --enable-automation,
// navigator.webdriver reports false, new headless mode, and a real virtual
// display (--screen-info) instead of CDP device-metrics emulation.
func withRealUserLaunchFlags(l *launcher.Launcher, headless bool) *launcher.Launcher {
	l = l.Delete("enable-automation").
		Set("disable-blink-features", "AutomationControlled")
	if headless {
		return l.HeadlessNew(true).
			Set("window-size", strconv.Itoa(headlessWindowWidth), strconv.Itoa(headlessWindowHeight)).
			Set("screen-info", fmt.Sprintf("{%dx%d}", headlessWindowWidth, headlessWindowHeight))
	}
	return l.Headless(false)
}

type uaBrandJSON struct {
	Brand   string `json:"brand"`
	Version string `json:"version"`
}

type pageUAInfo struct {
	UA              string        `json:"ua"`
	Platform        string        `json:"platform"`
	Brands          []uaBrandJSON `json:"brands"`
	Mobile          bool          `json:"mobile"`
	UAPlatform      string        `json:"uaPlatform"`
	Architecture    string        `json:"architecture"`
	Bitness         string        `json:"bitness"`
	Model           string        `json:"model"`
	PlatformVersion string        `json:"platformVersion"`
	FullVersionList []uaBrandJSON `json:"fullVersionList"`
	Wow64           bool          `json:"wow64"`
	HighEntropy     bool          `json:"highEntropy"`
}

func desktopChromeUA(ua string) string {
	return strings.ReplaceAll(ua, "HeadlessChrome", "Chrome")
}

func clientHintBrand(browserName string) string {
	switch browserName {
	case "Chrome":
		return "Google Chrome"
	case "Edge":
		return "Microsoft Edge"
	case "Ungoogled-Chromium", "":
		return "Chromium"
	default:
		return browserName
	}
}

func rewriteClientHintBrands(in []uaBrandJSON, headed string) []*proto.EmulationUserAgentBrandVersion {
	out := make([]*proto.EmulationUserAgentBrandVersion, 0, len(in))
	haveHeaded := false
	stripped := false
	headedVersion := ""
	for _, b := range in {
		if strings.Contains(b.Brand, "HeadlessChrome") {
			stripped = true
			headedVersion = b.Version
			continue
		}
		if b.Brand == headed {
			haveHeaded = true
		}
		out = append(out, &proto.EmulationUserAgentBrandVersion{
			Brand:   b.Brand,
			Version: b.Version,
		})
	}
	if stripped && !haveHeaded && headed != "" {
		out = append(out, &proto.EmulationUserAgentBrandVersion{
			Brand:   headed,
			Version: headedVersion,
		})
	}
	return out
}

func brandsHaveHeadlessChrome(in []uaBrandJSON) bool {
	for _, b := range in {
		if strings.Contains(b.Brand, "HeadlessChrome") {
			return true
		}
	}
	return false
}

func hasHeadlessChrome(info pageUAInfo) bool {
	return strings.Contains(info.UA, "HeadlessChrome") ||
		brandsHaveHeadlessChrome(info.Brands) ||
		brandsHaveHeadlessChrome(info.FullVersionList)
}

func uaCHPlatform(info pageUAInfo) string {
	if info.UAPlatform != "" {
		return info.UAPlatform
	}
	switch {
	case strings.Contains(info.UA, "Linux"):
		return "Linux"
	case strings.Contains(info.UA, "Windows"):
		return "Windows"
	case strings.Contains(info.UA, "Mac"):
		return "macOS"
	default:
		return ""
	}
}

// userAgentOverride rewrites HeadlessChrome in the user agent string and,
// when high-entropy collection succeeded, in Client Hint brands. A custom
// --user-agent string is left unchanged; brands are still rewritten so
// HeadlessChrome cannot leak through Sec-CH-UA. Blank architecture /
// platformVersion must not replace Chrome's real values.
func userAgentOverride(info pageUAInfo, browserName string) *proto.NetworkSetUserAgentOverride {
	headlessUA := strings.Contains(info.UA, "HeadlessChrome")
	if !hasHeadlessChrome(info) {
		return nil
	}
	if !info.HighEntropy && !headlessUA {
		// Custom UA with HeadlessChrome brands, but no hints to rewrite.
		return nil
	}
	req := &proto.NetworkSetUserAgentOverride{
		UserAgent: desktopChromeUA(info.UA),
		Platform:  info.Platform,
	}
	if !info.HighEntropy {
		return req
	}
	headed := clientHintBrand(browserName)
	req.UserAgentMetadata = &proto.EmulationUserAgentMetadata{
		Brands:          rewriteClientHintBrands(info.Brands, headed),
		FullVersionList: rewriteClientHintBrands(info.FullVersionList, headed),
		Platform:        uaCHPlatform(info),
		PlatformVersion: info.PlatformVersion,
		Architecture:    info.Architecture,
		Model:           info.Model,
		Mobile:          info.Mobile,
		Bitness:         info.Bitness,
		Wow64:           info.Wow64,
	}
	return req
}

// applyRealUserPageIdentity strips HeadlessChrome from a launched page's UA
// and Client Hints. Architecture, bitness, and platform version come from
// the browser's high-entropy hints; only the HeadlessChrome token is rewritten.
func applyRealUserPageIdentity(page *rod.Page, browserName string) error {
	res, err := page.Eval(`async () => {
		const ua = navigator.userAgent;
		const uad = navigator.userAgentData;
		const info = {
			ua: ua,
			platform: navigator.platform,
			brands: uad ? uad.brands : [],
			mobile: uad ? uad.mobile : false,
			uaPlatform: uad ? uad.platform : "",
			architecture: "",
			bitness: "",
			model: "",
			platformVersion: "",
			fullVersionList: [],
			wow64: false,
			highEntropy: false
		};
		const headlessBrand = info.brands.some(function (b) {
			return String(b.brand).includes("HeadlessChrome");
		});
		if ((!ua.includes("HeadlessChrome") && !headlessBrand) || !uad || !uad.getHighEntropyValues) {
			return JSON.stringify(info);
		}
		try {
			const high = await uad.getHighEntropyValues([
				"architecture",
				"bitness",
				"model",
				"platform",
				"platformVersion",
				"fullVersionList",
				"wow64"
			]);
			info.architecture = high.architecture || "";
			info.bitness = high.bitness || "";
			info.model = high.model || "";
			info.uaPlatform = high.platform || info.uaPlatform;
			info.platformVersion = high.platformVersion || "";
			info.fullVersionList = high.fullVersionList || [];
			info.wow64 = !!high.wow64;
			if (high.brands && high.brands.length) {
				info.brands = high.brands;
			}
			info.highEntropy = true;
		} catch (e) {}
		return JSON.stringify(info);
	}`)
	if err != nil {
		return err
	}

	var info pageUAInfo
	if err := json.Unmarshal([]byte(res.Value.Str()), &info); err != nil {
		return err
	}
	req := userAgentOverride(info, browserName)
	if req == nil {
		return nil
	}
	return page.SetUserAgent(req)
}

func applyHeadlessWindowIdentity(page *rod.Page) error {
	width, height := headlessWindowWidth, headlessWindowHeight
	// --window-size and --screen-info size the window and virtual display.
	// headless=new still reports outerWidth/outerHeight as 0 after
	// setWindowBounds. Do not replace the getter in page JS to hide that:
	// a non-native getter is itself a headless tell. Chromium may already
	// expose outerWidth as an own native property.
	return page.SetWindow(&proto.BrowserBounds{
		Width:       &width,
		Height:      &height,
		WindowState: proto.BrowserWindowStateNormal,
	})
}

// resolveWSURL reads webSocketDebuggerUrl from /json/version so connect can
// honour ctx. The advertised host is rewritten to 127.0.0.1 on the given port
// (Chrome may report [::1] or 0.0.0.0).
func resolveWSURL(ctx context.Context, port int) (string, error) {
	endpoint := fmt.Sprintf("http://127.0.0.1:%d/json/version", port)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}
	client := &http.Client{Timeout: ConnectTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GET %s: %s", endpoint, resp.Status)
	}

	var payload struct {
		WebSocketDebuggerURL string `json:"webSocketDebuggerUrl"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}
	if payload.WebSocketDebuggerURL == "" {
		return "", fmt.Errorf("no webSocketDebuggerUrl from %s", endpoint)
	}

	u, err := url.Parse(payload.WebSocketDebuggerURL)
	if err != nil {
		return "", err
	}
	u.Host = fmt.Sprintf("127.0.0.1:%d", port)
	return u.String(), nil
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
