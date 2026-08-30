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
	"time"

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
	return &Page{rodPage: p}
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

func (p *Page) NavigateTimeout(url string, timeout time.Duration) error {
	if p == nil || p.rodPage == nil {
		return fmt.Errorf("page is nil")
	}
	return p.rodPage.Timeout(timeout).Navigate(url)
}

func (p *Page) WaitStable(d time.Duration) error {
	if p == nil || p.rodPage == nil {
		return fmt.Errorf("page is nil")
	}
	return p.rodPage.WaitStable(d)
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
