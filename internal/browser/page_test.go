// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package browser

import "testing"

func TestWrapPage_Nil(t *testing.T) {
	if wrapPage(nil) != nil {
		t.Fatal("wrapPage(nil) should return nil")
	}
}

func TestPageNilGuards(t *testing.T) {
	var p *Page

	if _, err := p.HTML(); err == nil {
		t.Error("HTML: expected error for nil page")
	}
	if _, err := p.Meta(); err == nil {
		t.Error("Meta: expected error for nil page")
	}
	if err := p.NavigateTimeout("https://example.com", 0); err == nil {
		t.Error("NavigateTimeout: expected error for nil page")
	}
	if err := p.WaitStable(0); err == nil {
		t.Error("WaitStable: expected error for nil page")
	}
	if _, err := p.NavigationStatus(); err == nil {
		t.Error("NavigationStatus: expected error for nil page")
	}
	if _, err := p.Has("body"); err == nil {
		t.Error("Has: expected error for nil page")
	}
	if _, err := p.Element("body", 0); err == nil {
		t.Error("Element: expected error for nil page")
	}
	if _, err := p.PDF(); err == nil {
		t.Error("PDF: expected error for nil page")
	}
	if _, err := p.ScreenshotPNG(); err == nil {
		t.Error("ScreenshotPNG: expected error for nil page")
	}
	if err := p.Close(); err != nil {
		t.Errorf("Close: nil page should be a no-op, got %v", err)
	}

	empty := &Page{}
	if _, err := empty.HTML(); err == nil {
		t.Error("HTML: expected error for empty page")
	}

	var e *Element
	if err := e.WaitVisible(); err == nil {
		t.Error("WaitVisible: expected error for nil element")
	}
}
