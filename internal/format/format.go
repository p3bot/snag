// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package format

import (
	"fmt"

	"github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/base"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/commonmark"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/strikethrough"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/table"
	"github.com/k3a/html2text"
	"github.com/p3bot/snag/internal/logger"
)

const (
	Markdown = "md"
	HTML     = "html"
	Text     = "text"
	PDF      = "pdf"
	PNG      = "png"
)

// BinaryPage is the page surface needed for PDF and PNG output.
type BinaryPage interface {
	PDF() ([]byte, error)
	ScreenshotPNG() ([]byte, error)
}

// Page is the page surface for text and binary formats.
type Page interface {
	HTML() (string, error)
	BinaryPage
}

// Extension is the filename suffix for a format name, including the dot.
func Extension(name string) string {
	switch name {
	case Markdown:
		return ".md"
	case HTML:
		return ".html"
	case Text:
		return ".txt"
	case PDF:
		return ".pdf"
	case PNG:
		return ".png"
	default:
		return ".md"
	}
}

// Render converts a loaded page to the named format. It does not write.
func Render(page Page, formatName string) ([]byte, error) {
	cc := NewContentConverter(formatName)
	if formatName == PDF || formatName == PNG {
		return cc.renderBinary(page)
	}
	html, err := page.HTML()
	if err != nil {
		return nil, fmt.Errorf("failed to extract HTML: %w", err)
	}
	return cc.Convert(html)
}

var markdownConverter = converter.NewConverter(
	converter.WithPlugins(
		base.NewBasePlugin(),
		commonmark.NewCommonmarkPlugin(),
		table.NewTablePlugin(),
		strikethrough.NewStrikethroughPlugin(),
	),
)

type ContentConverter struct {
	format string
}

func NewContentConverter(format string) *ContentConverter {
	return &ContentConverter{
		format: format,
	}
}

func (cc *ContentConverter) Convert(html string) ([]byte, error) {
	var content string
	var err error

	switch cc.format {
	case HTML:
		content = html
		logger.Verbose("Output format: HTML (passthrough)")

	case Markdown:
		logger.Verbose("Converting HTML to Markdown...")
		content, err = cc.convertToMarkdown(html)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrConversionFailed, err)
		}
		logger.Debug("Converted to %d bytes of Markdown", len(content))

	case Text:
		logger.Verbose("Extracting plain text...")
		content = cc.extractPlainText(html)
		logger.Debug("Extracted %d bytes of plain text", len(content))

	default:
		return nil, fmt.Errorf("unsupported format: %s", cc.format)
	}

	return []byte(content), nil
}

func (cc *ContentConverter) convertToMarkdown(html string) (string, error) {
	markdown, err := markdownConverter.ConvertString(html)
	if err != nil {
		return "", err
	}

	return markdown, nil
}

func (cc *ContentConverter) extractPlainText(htmlContent string) string {
	text := html2text.HTML2TextWithOptions(
		htmlContent,
		html2text.WithUnixLineBreaks(),
	)

	return text
}

func (cc *ContentConverter) renderBinary(page BinaryPage) ([]byte, error) {
	switch cc.format {
	case PDF:
		logger.Verbose("Generating PDF...")
		data, err := page.PDF()
		if err != nil {
			return nil, fmt.Errorf("failed to generate PDF: %w", err)
		}
		logger.Debug("Generated %d bytes of PDF", len(data))
		return data, nil

	case PNG:
		logger.Verbose("Capturing PNG screenshot...")
		data, err := page.ScreenshotPNG()
		if err != nil {
			return nil, fmt.Errorf("failed to capture PNG screenshot: %w", err)
		}
		logger.Debug("Captured %d bytes of PNG", len(data))
		return data, nil

	default:
		return nil, fmt.Errorf("unsupported binary format: %s", cc.format)
	}
}
