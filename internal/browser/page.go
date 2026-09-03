// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package browser

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/p3bot/snag/internal/logger"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

// Page wraps a Rod page so CLI and domain packages do not import Rod.
type Page struct {
	rodPage *rod.Page
}

// PageMeta is the title and URL of a loaded page.
type PageMeta struct {
	Title string
	URL   string
}

func wrapPage(p *rod.Page) *Page {
	if p == nil {
		return nil
	}
	suppressJSDialogs(p)
	return &Page{rodPage: p}
}

// dialogPages tracks targets already prepared so wrapPage can run twice
// on the same tab.
var dialogPages sync.Map

// suppressJSDialogs stops alert/confirm/prompt from blocking fetch, and
// installs the fetch/XHR idle probe used by WaitHTTPIdle.
// Stubs dismiss (confirm false, prompt null). The CDP listener is a
// fallback: dismiss native dialogs, accept beforeunload so snag can leave.
// Attaching does not re-fire javascriptDialogOpening for a dialog that is
// already open, so dismiss once after the listener is running (ignore
// "no dialog"). Do not use a one-shot Rod HandleDialog wait: an empty
// event stops that waiter while a real dialog remains.
func suppressJSDialogs(p *rod.Page) {
	if p == nil {
		return
	}
	id := string(p.TargetID)
	if _, loaded := dialogPages.LoadOrStore(id, struct{}{}); loaded {
		return
	}

	_, err := p.EvalOnNewDocument(`window.alert = function() {};
window.confirm = function() { return false; };
window.prompt = function() { return null; };
(function() {
	var n = 0, quiet = Date.now();
	function bump(d) {
		n += d;
		if (n < 0) { n = 0; }
		if (n === 0) { quiet = Date.now(); }
	}
	window.__snagHTTPIdle = function(ms) {
		return n === 0 && (Date.now() - quiet) >= ms;
	};
	if (window.fetch) {
		var fetchFn = window.fetch;
		window.fetch = function() {
			bump(1);
			var p = fetchFn.apply(this, arguments);
			var done = function() { bump(-1); };
			if (p && p.then) { p.then(done, done); } else { done(); }
			return p;
		};
	}
	var send = XMLHttpRequest.prototype.send;
	XMLHttpRequest.prototype.send = function() {
		bump(1);
		this.addEventListener('loadend', function() { bump(-1); });
		return send.apply(this, arguments);
	};
})();`)
	if err != nil {
		logger.Debug("Failed to override JavaScript dialogs: %v", err)
	}

	wait := p.EachEvent(func(e *proto.PageJavascriptDialogOpening) {
		accept := e.Type == proto.PageDialogTypeBeforeunload
		logger.Debug("Handling JavaScript dialog (%s, accept=%t): %s", e.Type, accept, e.Message)
		if err := (proto.PageHandleJavaScriptDialog{Accept: accept}).Call(p); err != nil {
			logger.Debug("Failed to handle JavaScript dialog: %v", err)
		}
	})
	go wait()

	if err := (proto.PageHandleJavaScriptDialog{Accept: false}).Call(p); err != nil {
		logger.Debug("No existing JavaScript dialog to dismiss: %v", err)
	}
}

func (p *Page) HTML() (string, error) {
	if p == nil || p.rodPage == nil {
		return "", fmt.Errorf("page is nil")
	}
	return p.rodPage.HTML()
}

func (p *Page) Meta() (PageMeta, error) {
	if p == nil || p.rodPage == nil {
		return PageMeta{}, fmt.Errorf("page is nil")
	}
	info, err := p.rodPage.Info()
	if err != nil {
		return PageMeta{}, err
	}
	return PageMeta{Title: info.Title, URL: info.URL}, nil
}

func (p *Page) Close() error {
	if p == nil || p.rodPage == nil {
		return nil
	}
	return p.rodPage.Context(context.Background()).Close()
}

func (p *Page) rodCtx(ctx context.Context) (*rod.Page, error) {
	if p == nil || p.rodPage == nil {
		return nil, fmt.Errorf("page is nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return p.rodPage.Context(ctx), nil
}

func (p *Page) NavigateTimeout(url string, timeout time.Duration) error {
	if p == nil || p.rodPage == nil {
		return fmt.Errorf("page is nil")
	}
	return p.rodPage.Timeout(timeout).Navigate(url)
}

func (p *Page) Navigate(ctx context.Context, url string) error {
	page, err := p.rodCtx(ctx)
	if err != nil {
		return err
	}
	return page.Navigate(url)
}

// WaitLoad returns a wait that, after Navigate, blocks until the new
// document's window.onload (or ctx is done).
//
// Rod WaitNavigation(Load) matches the first lifecycle load. Enabling
// lifecycle events replays about:blank's load, so wait() can return while
// the target HTML is still a head-only shell. Poll until the main frame
// URL is no longer about:blank, then use Rod WaitLoad (readyState /
// window.onload) on that document.
func (p *Page) WaitLoad(ctx context.Context) (wait func() error, err error) {
	page, err := p.rodCtx(ctx)
	if err != nil {
		return nil, err
	}

	return func() error {
		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()
		for {
			info, infoErr := page.Info()
			if infoErr != nil {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				return infoErr
			}
			if info.URL != "" && !strings.HasPrefix(info.URL, "about:") {
				logger.Debug("Navigation committed to %s", info.URL)
				break
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-ticker.C:
			}
		}
		if err := page.WaitLoad(); err != nil {
			return err
		}
		return ctx.Err()
	}, nil
}

// WaitHTTPIdle returns a wait that, after load, blocks until fetch/XHR
// have been quiet for idle or ctx is done. idle is quiet time, not a
// deadline. In-flight counts come from the EvalOnNewDocument hook, which
// must be installed before Navigate (wrapPage). Websockets are not counted.
func (p *Page) WaitHTTPIdle(ctx context.Context, idle time.Duration) (wait func() error, err error) {
	page, err := p.rodCtx(ctx)
	if err != nil {
		return nil, err
	}
	if idle <= 0 {
		idle = HTTPIdle
	}
	ms := idle.Milliseconds()

	// fetch/XHR inflight is counted by the EvalOnNewDocument hook. Rod
	// WaitRequestIdle only runs its Add/Done callbacks once wait() is
	// called, so a fetch that started during parse is easy to miss.
	return func() error {
		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()
		for {
			res, evalErr := page.Eval(fmt.Sprintf(
				`() => typeof window.__snagHTTPIdle === 'function' && window.__snagHTTPIdle(%d)`,
				ms,
			))
			if evalErr != nil {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				logger.Debug("HTTP idle probe failed: %v", evalErr)
				return nil
			}
			if res.Value.Bool() {
				return nil
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-ticker.C:
			}
		}
	}, nil
}

func (p *Page) NavigationStatus() (int, error) {
	if p == nil || p.rodPage == nil {
		return 0, fmt.Errorf("page is nil")
	}
	// Hardcoded JavaScript only. Never evaluate user-provided scripts.
	statusCode, err := p.rodPage.Eval(`() => {
		return window.performance?.getEntriesByType?.('navigation')?.[0]?.responseStatus || 0;
	}`)
	if err != nil {
		return 0, err
	}
	return statusCode.Value.Int(), nil
}

func (p *Page) Has(selector string) (bool, error) {
	if p == nil || p.rodPage == nil {
		return false, fmt.Errorf("page is nil")
	}
	has, _, err := p.rodPage.Has(selector)
	return has, err
}

// Element is a DOM node on a Page.
type Element struct {
	rodElem *rod.Element
}

func (p *Page) Element(selector string, timeout time.Duration) (*Element, error) {
	if p == nil || p.rodPage == nil {
		return nil, fmt.Errorf("page is nil")
	}
	elem, err := p.rodPage.Timeout(timeout).Element(selector)
	if err != nil {
		return nil, err
	}
	return &Element{rodElem: elem}, nil
}

func (e *Element) WaitVisible() error {
	if e == nil || e.rodElem == nil {
		return fmt.Errorf("element is nil")
	}
	return e.rodElem.WaitVisible()
}

func (p *Page) PDF() ([]byte, error) {
	if p == nil || p.rodPage == nil {
		return nil, fmt.Errorf("page is nil")
	}
	stream, err := p.rodPage.PDF(&proto.PagePrintToPDF{
		PrintBackground: true,
	})
	if err != nil {
		return nil, fmt.Errorf("PDF generation failed: %w", err)
	}
	pdfData, err := io.ReadAll(stream)
	if err != nil {
		return nil, fmt.Errorf("failed to read PDF data: %w", err)
	}
	return pdfData, nil
}

func (p *Page) ScreenshotPNG() ([]byte, error) {
	if p == nil || p.rodPage == nil {
		return nil, fmt.Errorf("page is nil")
	}
	screenshotData, err := p.rodPage.Screenshot(true, &proto.PageCaptureScreenshot{
		Format: proto.PageCaptureScreenshotFormatPng,
	})
	if err != nil {
		return nil, fmt.Errorf("screenshot capture failed: %w", err)
	}
	return screenshotData, nil
}
