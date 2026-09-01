// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package skill_test

import (
	"strings"
	"testing"

	"github.com/p3bot/snag/internal/skill"
)

func TestTextFrontmatterAndH1(t *testing.T) {
	text := skill.Text()
	if !strings.HasPrefix(text, "---\nname: snag\n") {
		t.Fatal("skill must open with Agent Skills frontmatter (name: snag)")
	}
	if !strings.Contains(text, "description:") {
		t.Fatal("skill frontmatter must include description")
	}
	if !strings.Contains(text, "\n# Fetch web pages with snag\n") {
		t.Fatal("skill must have H1 after frontmatter")
	}
	if !strings.HasSuffix(text, "\n") {
		t.Fatal("Text must end with a newline")
	}
}

func TestTextRequiredGuidance(t *testing.T) {
	text := skill.Text()
	needles := []string{
		"snag <url>",
		"--url-file",
		"--list-tabs",
		"--open-browser",
		"--kill-browser",
		"--doctor",
		"--verbose",
		"--debug",
		"--wait-for",
		"--info",
		"--all-tabs",
		"PDF/PNG never go to stdout",
		"## Authenticated Page",
		"log in in that window",
		`snag -t "pattern"`,
	}
	for _, n := range needles {
		if !strings.Contains(text, n) {
			t.Errorf("skill missing required guidance %q", n)
		}
	}
}
