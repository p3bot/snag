// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package format

import (
	"fmt"
	"os"

	"github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/base"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/commonmark"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/strikethrough"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/table"
	"github.com/k3a/html2text"
	"github.com/p3bot/snag/internal/logger"
)

const (
	DefaultFileMode = 0644   // Owner RW, Group R, Other R
	BytesPerKB      = 1024.0 // Bytes in a kilobyte
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

func ProcessContent(page Page, formatName, outputFile string) error {
	converter := NewContentConverter(formatName)
	if formatName == PDF || formatName == PNG {
		return converter.ProcessPage(page, outputFile)
	}
	html, err := page.HTML()
	if err != nil {
		return fmt.Errorf("failed to extract HTML: %w", err)
	}
	return converter.Process(html, outputFile)
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

func (cc *ContentConverter) Process(html string, outputFile string) error {
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
			return fmt.Errorf("%w: %w", ErrConversionFailed, err)
		}
		logger.Debug("Converted to %d bytes of Markdown", len(content))

	case Text:
		logger.Verbose("Extracting plain text...")
		content = cc.extractPlainText(html)
		logger.Debug("Extracted %d bytes of plain text", len(content))

	default:
		return fmt.Errorf("unsupported format: %s", cc.format)
	}

	if outputFile != "" {
		return cc.writeToFile(content, outputFile)
	}

	return cc.writeToStdout(content)
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

func (cc *ContentConverter) writeToStdout(content string) error {
	logger.Verbose("Writing to stdout...")

	_, err := fmt.Print(content)
	if err != nil {
		return fmt.Errorf("failed to write to stdout: %w", err)
	}

	logger.Debug("Wrote %d bytes to stdout", len(content))

	return nil
}

func (cc *ContentConverter) writeToFile(content string, filename string) error {
	logger.Verbose("Writing to file: %s", filename)

	if _, err := os.Stat(filename); err == nil {
		logger.Verbose("Overwriting existing file: %s", filename)
	}

	err := os.WriteFile(filename, []byte(content), DefaultFileMode)
	if err != nil {
		return fmt.Errorf("failed to write to file %s: %w", filename, err)
	}

	sizeKB := float64(len(content)) / BytesPerKB
	logger.Success("Saved to %s (%.1f KB)", filename, sizeKB)

	return nil
}

func (cc *ContentConverter) ProcessPage(page BinaryPage, outputFile string) error {
	var data []byte
	var err error

	switch cc.format {
	case PDF:
		logger.Verbose("Generating PDF...")
		data, err = page.PDF()
		if err != nil {
			return fmt.Errorf("failed to generate PDF: %w", err)
		}
		logger.Debug("Generated %d bytes of PDF", len(data))

	case PNG:
		logger.Verbose("Capturing PNG screenshot...")
		data, err = page.ScreenshotPNG()
		if err != nil {
			return fmt.Errorf("failed to capture PNG screenshot: %w", err)
		}
		logger.Debug("Captured %d bytes of PNG", len(data))

	default:
		return fmt.Errorf("unsupported binary format: %s", cc.format)
	}

	if outputFile != "" {
		return cc.writeBinaryToFile(data, outputFile)
	}

	return cc.writeBinaryToStdout(data)
}

func (cc *ContentConverter) writeBinaryToStdout(data []byte) error {
	logger.Verbose("Writing binary data to stdout...")

	_, err := os.Stdout.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write to stdout: %w", err)
	}

	logger.Debug("Wrote %d bytes to stdout", len(data))

	return nil
}

func (cc *ContentConverter) writeBinaryToFile(data []byte, filename string) error {
	logger.Verbose("Writing binary data to file: %s", filename)

	if _, err := os.Stat(filename); err == nil {
		logger.Verbose("Overwriting existing file: %s", filename)
	}

	err := os.WriteFile(filename, data, DefaultFileMode)
	if err != nil {
		return fmt.Errorf("failed to write to file %s: %w", filename, err)
	}

	sizeKB := float64(len(data)) / BytesPerKB
	logger.Success("Saved to %s (%.1f KB)", filename, sizeKB)

	return nil
}
