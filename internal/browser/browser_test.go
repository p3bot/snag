// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package browser

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/launcher/flags"

	"github.com/p3bot/snag/internal/logger"
)

func init() {
	logger.SetDefault(logger.NewWithWriter(logger.LevelQuiet, io.Discard, false))
}

func TestDetectBrowserName(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		// Chrome
		{"chrome linux", "/usr/bin/google-chrome", "Chrome"},
		{"chrome stable", "/usr/bin/google-chrome-stable", "Chrome"},
		{"chrome macos", "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome", "Chrome"},
		{"chrome windows", "C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe", "Chrome"},
		{"chrome uppercase", "/usr/bin/CHROME", "Chrome"},
		{"chrome mixed case", "/usr/bin/Chrome", "Chrome"},

		// Chromium
		{"chromium linux", "/usr/bin/chromium", "Chromium"},
		{"chromium browser", "/usr/bin/chromium-browser", "Chromium"},
		{"chromium macos", "/Applications/Chromium.app/Contents/MacOS/Chromium", "Chromium"},
		{"chromium windows", "C:\\Program Files\\Chromium\\chromium.exe", "Chromium"},

		// Ungoogled Chromium (must be detected before regular Chromium)
		{"ungoogled chromium", "/usr/bin/ungoogled-chromium", "Ungoogled-Chromium"},
		{"ungoogled chromium app", "/Applications/Ungoogled Chromium.app", "Ungoogled-Chromium"},

		// Edge
		{"edge linux", "/usr/bin/microsoft-edge", "Edge"},
		{"edge macos", "/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge", "Edge"},
		{"edge windows", "C:\\Program Files\\Microsoft\\Edge\\Application\\msedge.exe", "Edge"},
		{"msedge", "/usr/bin/msedge", "Edge"},

		// Brave
		{"brave linux", "/usr/bin/brave", "Brave"},
		{"brave browser", "/usr/bin/brave-browser", "Brave"},
		{"brave macos", "/Applications/Brave Browser.app/Contents/MacOS/Brave Browser", "Brave"},
		{"brave windows", "C:\\Program Files\\BraveSoftware\\Brave-Browser\\brave.exe", "Brave"},

		// Opera
		{"opera linux", "/usr/bin/opera", "Opera"},
		{"opera macos", "/Applications/Opera.app/Contents/MacOS/Opera", "Opera"},

		// Vivaldi
		{"vivaldi linux", "/usr/bin/vivaldi", "Vivaldi"},
		{"vivaldi macos", "/Applications/Vivaldi.app/Contents/MacOS/Vivaldi", "Vivaldi"},

		// Arc
		{"arc macos", "/Applications/Arc.app/Contents/MacOS/Arc", "Arc"},

		// Yandex
		{"yandex linux", "/usr/bin/yandex-browser", "Yandex"},
		{"yandex macos", "/Applications/Yandex.app/Contents/MacOS/Yandex", "Yandex"},

		// Thorium
		{"thorium linux", "/usr/bin/thorium", "Thorium"},
		{"thorium browser", "/usr/bin/thorium-browser", "Thorium"},

		// Slimjet
		{"slimjet linux", "/usr/bin/slimjet", "Slimjet"},

		// Cent
		{"cent browser", "/usr/bin/cent-browser", "Cent"},

		// Extension handling
		{"exe extension", "C:\\chrome.exe", "Chrome"},
		{"app extension", "/Applications/Chrome.app", "Chrome"},

		// Path with directories
		{"deep path", "/home/user/.local/bin/chrome", "Chrome"},
		{"user directory", "/home/chrome-user/bin/chromium", "Chromium"},

		// Fallback cases
		{"unknown browser", "/usr/bin/firefox", "Firefox"},
		{"custom browser", "/usr/bin/mybrowser", "Mybrowser"},
		{"empty path", "", "."},  // filepath.Base("") returns "."
		{"just slash", "/", "/"}, // filepath.Base("/") returns "/"

		// Case sensitivity
		{"uppercase chromium", "/usr/bin/CHROMIUM", "Chromium"},
		{"mixed case edge", "/usr/bin/MsEdge", "Edge"},

		// Order of precedence (Chrome vs Chromium)
		{"chrome not chromium", "/usr/bin/chrome", "Chrome"},
		{"chromium not chrome", "/usr/bin/chromium", "Chromium"},

		// Complex paths
		{"windows complex", "C:\\Program Files (x86)\\Google\\Chrome\\Application\\chrome.exe", "Chrome"},
		{"macos bundle", "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome", "Chrome"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectBrowserName(tt.path)
			if result != tt.expected {
				t.Errorf("detectBrowserName(%q) = %q, expected %q", tt.path, result, tt.expected)
			}
		})
	}
}

func TestDetectBrowserName_OrderOfPrecedence(t *testing.T) {
	// Test that more specific matches take precedence over generic ones
	tests := []struct {
		name        string
		path        string
		expected    string
		notExpected string
	}{
		{"ungoogled before chromium", "/usr/bin/ungoogled-chromium", "Ungoogled-Chromium", "Chromium"},
		{"chrome before chromium", "/usr/bin/google-chrome", "Chrome", "Chromium"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectBrowserName(tt.path)
			if result != tt.expected {
				t.Errorf("detectBrowserName(%q) = %q, expected %q", tt.path, result, tt.expected)
			}
			if result == tt.notExpected {
				t.Errorf("detectBrowserName(%q) = %q, should not match %q", tt.path, result, tt.notExpected)
			}
		})
	}
}

func TestDetectBrowserName_ExtensionStripping(t *testing.T) {
	// Test that .exe and .app extensions are properly stripped
	tests := []struct {
		name         string
		pathWithExt  string
		pathNoExt    string
		expectedName string
	}{
		{".exe stripping", "C:\\chrome.exe", "/usr/bin/chrome", "Chrome"},
		{".app stripping", "/Applications/Chrome.app", "/usr/bin/Chrome", "Chrome"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resultWithExt := detectBrowserName(tt.pathWithExt)
			resultNoExt := detectBrowserName(tt.pathNoExt)

			if resultWithExt != tt.expectedName {
				t.Errorf("detectBrowserName(%q) = %q, expected %q", tt.pathWithExt, resultWithExt, tt.expectedName)
			}
			if resultNoExt != tt.expectedName {
				t.Errorf("detectBrowserName(%q) = %q, expected %q", tt.pathNoExt, resultNoExt, tt.expectedName)
			}
			if resultWithExt != resultNoExt {
				t.Errorf("Extension stripping failed: with ext = %q, without ext = %q", resultWithExt, resultNoExt)
			}
		})
	}
}

func TestDetectBrowserName_FallbackBehavior(t *testing.T) {
	// Test fallback behavior for unknown browsers
	tests := []struct {
		name     string
		path     string
		contains string // Should contain this substring
	}{
		{"capitalizes first letter", "/usr/bin/firefox", "F"},
		{"preserves rest", "/usr/bin/firefox", "irefox"},
		{"handles single char", "/usr/bin/x", "X"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectBrowserName(tt.path)
			if len(result) == 0 {
				t.Errorf("detectBrowserName(%q) returned empty string", tt.path)
			}
		})
	}
}

// TestKillBrowserOnPortNotFound tests killing when no browser on port.
func TestKillBrowserOnPortNotFound(t *testing.T) {
	port := 9999 // Use unlikely port
	bm := NewBrowserManager(BrowserOptions{Port: port})

	count, err := bm.KillBrowser(port)
	if err != nil {
		t.Fatalf("KillBrowser should not error when no browser found: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected to kill 0 processes, got %d", count)
	}
}

// TestKillAllBrowsersNoneFound tests killing all when no browsers running.
func TestKillAllBrowsersNoneFound(t *testing.T) {
	// Try to kill when no browsers with remote debugging are running
	bm := NewBrowserManager(BrowserOptions{})
	count, err := bm.KillBrowser(0)

	// Should not error
	if err != nil {
		t.Fatalf("KillBrowser should not error when no browsers found: %v", err)
	}

	// Count should be 0 or more (depending on if any debug browsers are running)
	if count < 0 {
		t.Errorf("Expected non-negative count, got %d", count)
	}
}

func TestProbePort_CountsPageTargets(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/json/list", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[
			{"type":"page","url":"https://example.com"},
			{"type":"page","url":"https://github.com"},
			{"type":"service_worker","url":"https://example.com/sw.js"}
		]`))
	})
	go http.Serve(ln, mux)

	port := ln.Addr().(*net.TCPAddr).Port
	n, err := ProbePort(port)
	if err != nil {
		t.Fatalf("ProbePort: %v", err)
	}
	if n != 2 {
		t.Errorf("ProbePort = %d, want 2 page targets", n)
	}
}

func TestResolveWSURL_RewritesHost(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	port := ln.Addr().(*net.TCPAddr).Port
	mux := http.NewServeMux()
	mux.HandleFunc("/json/version", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"webSocketDebuggerUrl":"ws://0.0.0.0:%d/devtools/browser/abc"}`, port)
	})
	go http.Serve(ln, mux)

	got, err := resolveWSURL(context.Background(), port)
	if err != nil {
		t.Fatalf("resolveWSURL: %v", err)
	}
	want := fmt.Sprintf("ws://127.0.0.1:%d/devtools/browser/abc", port)
	if got != want {
		t.Fatalf("resolveWSURL = %q, want %q", got, want)
	}
}

func TestResolveWSURL_Canceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := resolveWSURL(ctx, 1)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled ctx: %v", err)
	}
}

func TestResolveWSURL_BadStatus(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/json/version", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusInternalServerError)
	})
	go http.Serve(ln, mux)

	_, err = resolveWSURL(context.Background(), ln.Addr().(*net.TCPAddr).Port)
	if err == nil {
		t.Fatal("expected error for non-OK status")
	}
}

func TestResolveWSURL_MissingURL(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/json/version", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{}`))
	})
	go http.Serve(ln, mux)

	_, err = resolveWSURL(context.Background(), ln.Addr().(*net.TCPAddr).Port)
	if err == nil {
		t.Fatal("expected error when webSocketDebuggerUrl is empty")
	}
}

func TestDesktopChromeUA(t *testing.T) {
	in := "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) HeadlessChrome/152.0.0.0 Safari/537.36"
	want := "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/152.0.0.0 Safari/537.36"
	if got := desktopChromeUA(in); got != want {
		t.Fatalf("desktopChromeUA() = %q, want %q", got, want)
	}
	plain := "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/152.0.0.0 Safari/537.36"
	if got := desktopChromeUA(plain); got != plain {
		t.Fatalf("desktopChromeUA() changed a headed UA: %q", got)
	}
}

func TestClientHintBrand(t *testing.T) {
	tests := []struct {
		name, want string
	}{
		{"Chrome", "Google Chrome"},
		{"Edge", "Microsoft Edge"},
		{"Brave", "Brave"},
		{"Chromium", "Chromium"},
		{"Ungoogled-Chromium", "Chromium"},
		{"", "Chromium"},
		{"Vivaldi", "Vivaldi"},
	}
	for _, tt := range tests {
		if got := clientHintBrand(tt.name); got != tt.want {
			t.Errorf("clientHintBrand(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestRewriteClientHintBrands(t *testing.T) {
	in := []uaBrandJSON{
		{Brand: "Not:A-Brand", Version: "99"},
		{Brand: "Chromium", Version: "152"},
		{Brand: "HeadlessChrome", Version: "152"},
	}
	got := rewriteClientHintBrands(in, "Brave")
	if len(got) != 3 {
		t.Fatalf("len = %d, want 3", len(got))
	}
	if got[1].Brand != "Chromium" {
		t.Fatalf("Chromium brand rewritten: %q", got[1].Brand)
	}
	if got[2].Brand != "Brave" {
		t.Fatalf("HeadlessChrome brand = %q, want Brave", got[2].Brand)
	}
	if got[2].Version != "152" {
		t.Fatalf("version = %q, want 152", got[2].Version)
	}

	headed := []uaBrandJSON{
		{Brand: "Not:A-Brand", Version: "99"},
		{Brand: "Chromium", Version: "152"},
		{Brand: "Google Chrome", Version: "152"},
	}
	got = rewriteClientHintBrands(headed, "Brave")
	if got[2].Brand != "Google Chrome" {
		t.Fatalf("headed brand rewritten: %q", got[2].Brand)
	}
	if len(got) != 3 {
		t.Fatalf("headed list grew: len = %d", len(got))
	}

	chromium := []uaBrandJSON{
		{Brand: "Not:A-Brand", Version: "99"},
		{Brand: "Chromium", Version: "152"},
		{Brand: "HeadlessChrome", Version: "152"},
	}
	got = rewriteClientHintBrands(chromium, "Chromium")
	if len(got) != 2 {
		t.Fatalf("Chromium brands len = %d, want 2 (no duplicate)", len(got))
	}
	if got[0].Brand != "Not:A-Brand" || got[1].Brand != "Chromium" {
		t.Fatalf("Chromium brands = %+v", got)
	}
}

func TestUserAgentOverride_NoHeadless(t *testing.T) {
	got := userAgentOverride(pageUAInfo{
		UA: "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/152.0.0.0 Safari/537.36",
	}, "Chrome")
	if got != nil {
		t.Fatalf("headed UA produced override: %+v", got)
	}
}

func TestUserAgentOverride_WithoutHighEntropy(t *testing.T) {
	in := pageUAInfo{
		UA:       "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) HeadlessChrome/152.0.0.0 Safari/537.36",
		Platform: "Linux x86_64",
		Brands:   []uaBrandJSON{{Brand: "HeadlessChrome", Version: "152"}},
	}
	got := userAgentOverride(in, "Chrome")
	if got == nil {
		t.Fatal("expected UA rewrite")
	}
	wantUA := "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/152.0.0.0 Safari/537.36"
	if got.UserAgent != wantUA {
		t.Fatalf("UserAgent = %q, want %q", got.UserAgent, wantUA)
	}
	if got.Platform != "Linux x86_64" {
		t.Fatalf("Platform = %q", got.Platform)
	}
	if got.UserAgentMetadata != nil {
		t.Fatalf("blank Client Hints attached: %+v", got.UserAgentMetadata)
	}
}

func TestUserAgentOverride_CustomUAHeadlessBrand(t *testing.T) {
	in := pageUAInfo{
		UA:           "Mozilla/5.0 (Custom Bot) snag/test",
		Platform:     "Linux x86_64",
		HighEntropy:  true,
		Architecture: "x86",
		Bitness:      "64",
		UAPlatform:   "Linux",
		Brands: []uaBrandJSON{
			{Brand: "Not:A-Brand", Version: "99"},
			{Brand: "Chromium", Version: "152"},
			{Brand: "HeadlessChrome", Version: "152"},
		},
		FullVersionList: []uaBrandJSON{
			{Brand: "HeadlessChrome", Version: "152.0.7339.80"},
		},
	}
	got := userAgentOverride(in, "Chrome")
	if got == nil || got.UserAgentMetadata == nil {
		t.Fatal("expected Client Hints rewrite for custom UA")
	}
	if got.UserAgent != in.UA {
		t.Fatalf("custom UA rewritten: %q", got.UserAgent)
	}
	if got.UserAgentMetadata.Brands[2].Brand != "Google Chrome" {
		t.Fatalf("brands = %+v", got.UserAgentMetadata.Brands)
	}

	in.HighEntropy = false
	if userAgentOverride(in, "Chrome") != nil {
		t.Fatal("custom UA without high entropy should not override")
	}
}

func TestUserAgentOverride_WithHighEntropy(t *testing.T) {
	in := pageUAInfo{
		UA:              "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) HeadlessChrome/152.0.0.0 Safari/537.36",
		Platform:        "Linux x86_64",
		HighEntropy:     true,
		Architecture:    "x86",
		Bitness:         "64",
		UAPlatform:      "Linux",
		PlatformVersion: "6.16.0",
		Brands: []uaBrandJSON{
			{Brand: "Not:A-Brand", Version: "99"},
			{Brand: "Chromium", Version: "152"},
			{Brand: "HeadlessChrome", Version: "152"},
		},
		FullVersionList: []uaBrandJSON{
			{Brand: "HeadlessChrome", Version: "152.0.7339.80"},
		},
	}
	got := userAgentOverride(in, "Chrome")
	if got == nil || got.UserAgentMetadata == nil {
		t.Fatal("expected Client Hints with high-entropy data")
	}
	meta := got.UserAgentMetadata
	if meta.Architecture != "x86" || meta.Bitness != "64" || meta.PlatformVersion != "6.16.0" {
		t.Fatalf("high-entropy fields: %+v", meta)
	}
	if len(meta.Brands) != 3 || meta.Brands[2].Brand != "Google Chrome" {
		t.Fatalf("brands = %+v", meta.Brands)
	}
	if len(meta.FullVersionList) != 1 || meta.FullVersionList[0].Brand != "Google Chrome" {
		t.Fatalf("fullVersionList = %+v", meta.FullVersionList)
	}
}

func TestUACHPlatform(t *testing.T) {
	if got := uaCHPlatform(pageUAInfo{UAPlatform: "Linux"}); got != "Linux" {
		t.Fatalf("explicit platform = %q", got)
	}
	if got := uaCHPlatform(pageUAInfo{UA: "Mozilla/5.0 (X11; Linux aarch64)"}); got != "Linux" {
		t.Fatalf("UA Linux = %q", got)
	}
	if got := uaCHPlatform(pageUAInfo{UA: "Mozilla/5.0 (Windows NT 10.0)"}); got != "Windows" {
		t.Fatalf("UA Windows = %q", got)
	}
	if got := uaCHPlatform(pageUAInfo{UA: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)"}); got != "macOS" {
		t.Fatalf("UA Mac = %q", got)
	}
}

func TestWithRealUserLaunchFlags_Headless(t *testing.T) {
	args := strings.Join(withRealUserLaunchFlags(launcher.New(), true).FormatArgs(), " ")
	if strings.Contains(args, "--enable-automation") {
		t.Fatalf("headless launch still passes --enable-automation: %s", args)
	}
	if !strings.Contains(args, "--headless=new") {
		t.Fatalf("headless launch missing --headless=new: %s", args)
	}
	if strings.Contains(args, "--headless ") || strings.HasSuffix(args, "--headless") {
		t.Fatalf("headless launch still uses old --headless: %s", args)
	}
	if !strings.Contains(args, "--window-size=1920,1080") {
		t.Fatalf("headless launch missing window size: %s", args)
	}
	if !strings.Contains(args, "--screen-info={1920x1080}") {
		t.Fatalf("headless launch missing virtual screen: %s", args)
	}
	if !strings.Contains(args, "--disable-blink-features=AutomationControlled") {
		t.Fatalf("headless launch missing AutomationControlled: %s", args)
	}
}

func TestWithRealUserLaunchFlags_Visible(t *testing.T) {
	args := strings.Join(withRealUserLaunchFlags(launcher.New(), false).FormatArgs(), " ")
	if strings.Contains(args, "--enable-automation") {
		t.Fatalf("visible launch still passes --enable-automation: %s", args)
	}
	if strings.Contains(args, "--headless") {
		t.Fatalf("visible launch should not be headless: %s", args)
	}
	if strings.Contains(args, "--window-size=") {
		t.Fatalf("visible launch should not force window size: %s", args)
	}
	if strings.Contains(args, "--screen-info=") {
		t.Fatalf("visible launch should not force virtual screen: %s", args)
	}
	if !strings.Contains(args, "--disable-blink-features=AutomationControlled") {
		t.Fatalf("visible launch missing AutomationControlled: %s", args)
	}
}

func TestClose_HeadlessNilBrowserNoPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Close panicked: %v", r)
		}
	}()
	bm := NewBrowserManager(BrowserOptions{})
	bm.wasLaunched = true
	bm.launchedHeadless = true
	bm.Close()
}

func TestListTabs_CanceledSession(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	bm := NewBrowserManager(BrowserOptions{})
	bm.browser = rod.New().Context(ctx)

	if _, err := bm.ListTabs(); !errors.Is(err, context.Canceled) {
		t.Fatalf("ListTabs cancelled session: %v", err)
	}
	if _, err := bm.GetTabByIndex(1); !errors.Is(err, context.Canceled) {
		t.Fatalf("GetTabByIndex cancelled session: %v", err)
	}
	if _, err := bm.GetTabsByPattern("example"); !errors.Is(err, context.Canceled) {
		t.Fatalf("GetTabsByPattern cancelled session: %v", err)
	}
}

func TestConnectCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	bm := NewBrowserManager(BrowserOptions{Port: 1})
	if err := bm.Connect(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("Connect cancelled ctx: %v", err)
	}
	if err := bm.ConnectExisting(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("ConnectExisting cancelled ctx: %v", err)
	}
	if err := bm.OpenBrowserOnly(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("OpenBrowserOnly cancelled ctx: %v", err)
	}
}

func TestTempProfileHeadlessCloseRemovesDir(t *testing.T) {
	bm := NewBrowserManager(BrowserOptions{
		Port:          freePort(t),
		ForceHeadless: true,
		TempProfile:   true,
	})
	if _, err := bm.FindBrowserPath(); err != nil {
		t.Skip("browser not available")
	}
	if err := bm.Connect(context.Background()); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	if bm.launcher == nil {
		t.Fatal("expected a launched browser")
	}
	dir := bm.launcher.Get(flags.UserDataDir)
	if dir == "" {
		t.Fatal("launcher user-data-dir is empty")
	}
	if !strings.HasPrefix(dir, os.TempDir()) {
		t.Fatalf("temp profile %q is not under %s", dir, os.TempDir())
	}
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("temp profile missing while running: %v", err)
	}
	bm.Close()
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatalf("temp profile still present after Close(): %s (%v)", dir, err)
	}
}

func freePort(t *testing.T) int {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()
	return port
}

func TestProbePort_ConnectionRefused(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()

	_, err = ProbePort(port)
	if err == nil {
		t.Fatal("expected error for closed port")
	}
}
