// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package format

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/p3bot/snag/internal/logger"
	"github.com/p3bot/snag/internal/testutil"
)

func init() {
	logger.SetDefault(logger.Discard())
}

func TestConvertToMarkdown_Headings(t *testing.T) {
	html := `<html><body>
		<h1>Heading 1</h1>
		<h2>Heading 2</h2>
		<h3>Heading 3</h3>
	</body></html>`

	converter := NewContentConverter(Markdown)
	md, err := converter.convertToMarkdown(html)
	if err != nil {
		t.Fatalf("convertToMarkdown failed: %v", err)
	}

	// Check for standard markdown heading format with space after #
	if !strings.Contains(md, "# Heading 1") {
		t.Errorf("expected h1 to convert to standard markdown heading '# Heading 1', got:\n%s", md)
	}
	if !strings.Contains(md, "## Heading 2") {
		t.Errorf("expected h2 to convert to standard markdown heading '## Heading 2', got:\n%s", md)
	}
	if !strings.Contains(md, "### Heading 3") {
		t.Errorf("expected h3 to convert to standard markdown heading '### Heading 3', got:\n%s", md)
	}
}

func TestConvertToMarkdown_Links(t *testing.T) {
	html := `<html><body>
		<a href="https://example.com">Example Link</a>
	</body></html>`

	converter := NewContentConverter(Markdown)
	md, err := converter.convertToMarkdown(html)
	if err != nil {
		t.Fatalf("convertToMarkdown failed: %v", err)
	}

	// Check for markdown link format [text](url) as a combined string
	if !strings.Contains(md, "[Example Link](https://example.com)") {
		t.Errorf("expected link to convert to markdown format [Example Link](https://example.com), got:\n%s", md)
	}
}

func TestConvertToMarkdown_Tables(t *testing.T) {
	html := `<html><body>
		<table>
			<thead>
				<tr><th>Name</th><th>Value</th></tr>
			</thead>
			<tbody>
				<tr><td>Item 1</td><td>100</td></tr>
			</tbody>
		</table>
	</body></html>`

	converter := NewContentConverter(Markdown)
	md, err := converter.convertToMarkdown(html)
	if err != nil {
		t.Fatalf("convertToMarkdown failed: %v", err)
	}

	// Verify proper markdown table syntax (not just content preservation)
	if !strings.Contains(md, "|") {
		t.Errorf("expected markdown table with pipe characters, got:\n%s", md)
	}

	// Verify header row
	if !strings.Contains(md, "Name") || !strings.Contains(md, "Value") {
		t.Errorf("expected table headers 'Name' and 'Value', got:\n%s", md)
	}

	// Verify data row
	if !strings.Contains(md, "Item 1") || !strings.Contains(md, "100") {
		t.Errorf("expected table data 'Item 1' and '100', got:\n%s", md)
	}

	// Verify separator row (some variation of dashes)
	if !strings.Contains(md, "---") {
		t.Errorf("expected table separator with dashes, got:\n%s", md)
	}
}

func TestConvertToMarkdown_Strikethrough(t *testing.T) {
	html := `<html><body>
		<p>This is <del>deleted</del> text.</p>
		<p>This is <s>strikethrough</s> text.</p>
		<p>This is <strike>old strikethrough</strike> text.</p>
	</body></html>`

	converter := NewContentConverter(Markdown)
	md, err := converter.convertToMarkdown(html)
	if err != nil {
		t.Fatalf("convertToMarkdown failed: %v", err)
	}

	// Verify strikethrough syntax
	if !strings.Contains(md, "~~") {
		t.Errorf("expected strikethrough with ~~ syntax, got:\n%s", md)
	}

	// Verify content is preserved for all three strikethrough tags
	if !strings.Contains(md, "deleted") {
		t.Errorf("expected 'deleted' content from <del> tag, got:\n%s", md)
	}
	if !strings.Contains(md, "strikethrough") {
		t.Errorf("expected 'strikethrough' content from <s> tag, got:\n%s", md)
	}
	if !strings.Contains(md, "old") {
		t.Errorf("expected 'old' content from <strike> tag, got:\n%s", md)
	}
}

func TestConvertToMarkdown_Lists(t *testing.T) {
	html := `<html><body>
		<ul>
			<li>Unordered 1</li>
			<li>Unordered 2</li>
		</ul>
		<ol>
			<li>Ordered 1</li>
			<li>Ordered 2</li>
		</ol>
	</body></html>`

	converter := NewContentConverter(Markdown)
	md, err := converter.convertToMarkdown(html)
	if err != nil {
		t.Fatalf("convertToMarkdown failed: %v", err)
	}

	// Check for unordered list syntax (*, -, or +) and content
	hasUnorderedMarker := strings.Contains(md, "* ") || strings.Contains(md, "- ") || strings.Contains(md, "+ ")
	if !hasUnorderedMarker {
		t.Errorf("expected unordered list markers (*, -, or +) in markdown, got:\n%s", md)
	}
	if !strings.Contains(md, "Unordered 1") || !strings.Contains(md, "Unordered 2") {
		t.Errorf("expected unordered list items in markdown, got:\n%s", md)
	}

	// Check for ordered list syntax (1., 2., etc.) and content
	hasOrderedMarker := strings.Contains(md, "1. ") || strings.Contains(md, "1) ")
	if !hasOrderedMarker {
		t.Errorf("expected ordered list markers (1., 2., etc.) in markdown, got:\n%s", md)
	}
	if !strings.Contains(md, "Ordered 1") || !strings.Contains(md, "Ordered 2") {
		t.Errorf("expected ordered list items in markdown, got:\n%s", md)
	}
}

func TestConvertToMarkdown_CodeBlocks(t *testing.T) {
	html := `<html><body>
		<pre><code>function hello() {
	console.log("Hello");
}</code></pre>
	</body></html>`

	converter := NewContentConverter(Markdown)
	md, err := converter.convertToMarkdown(html)
	if err != nil {
		t.Fatalf("convertToMarkdown failed: %v", err)
	}

	// Check for fenced code block markers
	if !strings.Contains(md, "```") {
		t.Errorf("expected code to be wrapped in fenced code block (```), got:\n%s", md)
	}

	// Check for code content
	if !strings.Contains(md, "hello") || !strings.Contains(md, "console.log") {
		t.Errorf("expected code block content in markdown, got:\n%s", md)
	}
}

func TestConvertToMarkdown_Minimal(t *testing.T) {
	html := `<html><body>Hello</body></html>`

	converter := NewContentConverter(Markdown)
	md, err := converter.convertToMarkdown(html)
	if err != nil {
		t.Fatalf("convertToMarkdown failed: %v", err)
	}

	// Check for basic content
	if !strings.Contains(md, "Hello") {
		t.Errorf("expected 'Hello' in markdown output, got:\n%s", md)
	}
}

func TestConvertToMarkdown_NestedLists(t *testing.T) {
	tests := []struct {
		name     string
		filename string
	}{
		{
			name:     "nested unordered lists",
			filename: testutil.Testdata("nested-lists-ul.html"),
		},
		{
			name:     "nested ordered lists",
			filename: testutil.Testdata("nested-lists-ol.html"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			htmlBytes, err := os.ReadFile(tt.filename)
			if err != nil {
				t.Fatalf("failed to read test file: %v", err)
			}

			converter := NewContentConverter(Markdown)
			md, err := converter.convertToMarkdown(string(htmlBytes))
			if err != nil {
				t.Fatalf("convertToMarkdown failed: %v", err)
			}

			// Verify nested structure is preserved
			if !strings.Contains(md, "Item 1") {
				t.Errorf("expected 'Item 1' in output, got:\n%s", md)
			}
			if !strings.Contains(md, "Nested 1.1") {
				t.Errorf("expected nested items in output, got:\n%s", md)
			}
			if !strings.Contains(md, "Item 2") {
				t.Errorf("expected 'Item 2' in output, got:\n%s", md)
			}
		})
	}
}

func TestConvertToMarkdown_ComplexTables(t *testing.T) {
	tests := []struct {
		name     string
		filename string
	}{
		{
			name:     "table with colspan and footer",
			filename: testutil.Testdata("table-complex.html"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			htmlBytes, err := os.ReadFile(tt.filename)
			if err != nil {
				t.Fatalf("failed to read test file: %v", err)
			}

			converter := NewContentConverter(Markdown)
			md, err := converter.convertToMarkdown(string(htmlBytes))
			if err != nil {
				t.Fatalf("convertToMarkdown failed: %v", err)
			}

			// Verify table content is preserved
			if !strings.Contains(md, "Header") {
				t.Errorf("expected 'Header' in output, got:\n%s", md)
			}
			if !strings.Contains(md, "Cell 1") || !strings.Contains(md, "Cell 2") {
				t.Errorf("expected cell content in output, got:\n%s", md)
			}
		})
	}
}

func TestConvertToMarkdown_Images(t *testing.T) {
	htmlBytes, err := os.ReadFile(testutil.Testdata("images.html"))
	if err != nil {
		t.Fatalf("failed to read test file: %v", err)
	}

	converter := NewContentConverter(Markdown)
	md, err := converter.convertToMarkdown(string(htmlBytes))
	if err != nil {
		t.Fatalf("convertToMarkdown failed: %v", err)
	}

	// Check for markdown image syntax: ![alt](url)
	if !strings.Contains(md, "![Example Image]") {
		t.Errorf("expected markdown image syntax in output, got:\n%s", md)
	}
	if !strings.Contains(md, "example.com/image.png") {
		t.Errorf("expected image URL in output, got:\n%s", md)
	}
}

func TestConvertToMarkdown_Blockquotes(t *testing.T) {
	tests := []struct {
		name     string
		filename string
	}{
		{
			name:     "simple blockquote",
			filename: testutil.Testdata("blockquotes.html"),
		},
		{
			name:     "nested blockquotes",
			filename: testutil.Testdata("blockquotes-nested.html"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			htmlBytes, err := os.ReadFile(tt.filename)
			if err != nil {
				t.Fatalf("failed to read test file: %v", err)
			}

			converter := NewContentConverter(Markdown)
			md, err := converter.convertToMarkdown(string(htmlBytes))
			if err != nil {
				t.Fatalf("convertToMarkdown failed: %v", err)
			}

			// Check for blockquote syntax (> )
			if !strings.Contains(md, ">") {
				t.Errorf("expected blockquote marker (>) in output, got:\n%s", md)
			}
			if !strings.Contains(md, "quote") {
				t.Errorf("expected quote content in output, got:\n%s", md)
			}
		})
	}
}

func TestConvertToMarkdown_DefinitionLists(t *testing.T) {
	htmlBytes, err := os.ReadFile(testutil.Testdata("definition-lists.html"))
	if err != nil {
		t.Fatalf("failed to read test file: %v", err)
	}

	converter := NewContentConverter(Markdown)
	md, err := converter.convertToMarkdown(string(htmlBytes))
	if err != nil {
		t.Fatalf("convertToMarkdown failed: %v", err)
	}

	// Verify terms and definitions are preserved
	if !strings.Contains(md, "Term 1") {
		t.Errorf("expected 'Term 1' in output, got:\n%s", md)
	}
	if !strings.Contains(md, "Definition 1") {
		t.Errorf("expected 'Definition 1' in output, got:\n%s", md)
	}
}

func TestConvertToMarkdown_MalformedHTML(t *testing.T) {
	tests := []struct {
		name     string
		filename string
	}{
		{
			name:     "unclosed tags",
			filename: testutil.Testdata("malformed-unclosed.html"),
		},
		{
			name:     "mismatched tags",
			filename: testutil.Testdata("malformed-mismatched.html"),
		},
		{
			name:     "wrong order",
			filename: testutil.Testdata("malformed-wrong-order.html"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			htmlBytes, err := os.ReadFile(tt.filename)
			if err != nil {
				t.Fatalf("failed to read test file: %v", err)
			}

			converter := NewContentConverter(Markdown)
			md, err := converter.convertToMarkdown(string(htmlBytes))

			// Should not panic or error, just handle gracefully
			if err != nil {
				t.Logf("Warning: malformed HTML caused error: %v", err)
			}

			// Should still produce some output
			if len(strings.TrimSpace(md)) == 0 && err == nil {
				t.Error("expected some output even for malformed HTML")
			}
		})
	}
}

func TestConvertToMarkdown_SpecialCharacters(t *testing.T) {
	htmlBytes, err := os.ReadFile(testutil.Testdata("special-characters.html"))
	if err != nil {
		t.Fatalf("failed to read test file: %v", err)
	}

	converter := NewContentConverter(Markdown)
	md, err := converter.convertToMarkdown(string(htmlBytes))
	if err != nil {
		t.Fatalf("convertToMarkdown failed: %v", err)
	}

	// Verify special characters are handled
	if !strings.Contains(md, "Copyright") {
		t.Errorf("expected 'Copyright' in output, got:\n%s", md)
	}
	// Note: HTML entities may be converted to actual symbols
	// Just verify content is preserved in some form
}

// Phase 3: Text extraction tests

func TestExtractPlainText_Headings(t *testing.T) {
	html := `<html><body>
		<h1>Main Title</h1>
		<h2>Subtitle</h2>
		<p>Paragraph text</p>
	</body></html>`

	converter := NewContentConverter(Text)
	text := converter.extractPlainText(html)

	// Check for text content (no HTML tags, no markdown syntax)
	if !strings.Contains(text, "Main Title") {
		t.Errorf("expected 'Main Title' in text output, got:\n%s", text)
	}
	if !strings.Contains(text, "Subtitle") {
		t.Errorf("expected 'Subtitle' in text output, got:\n%s", text)
	}
	if !strings.Contains(text, "Paragraph text") {
		t.Errorf("expected 'Paragraph text' in text output, got:\n%s", text)
	}

	// Should NOT contain HTML tags
	if strings.Contains(text, "<h1>") || strings.Contains(text, "</h1>") {
		t.Errorf("expected no HTML tags in text output, got:\n%s", text)
	}

	// Should NOT contain markdown syntax
	if strings.Contains(text, "# Main Title") {
		t.Errorf("expected no markdown syntax in text output, got:\n%s", text)
	}
}

func TestExtractPlainText_Links(t *testing.T) {
	html := `<html><body>
		<p>Visit <a href="https://example.com">our website</a> for more info.</p>
	</body></html>`

	converter := NewContentConverter(Text)
	text := converter.extractPlainText(html)

	// Should contain the URL (plain text extraction shows URLs)
	if !strings.Contains(text, "example.com") {
		t.Errorf("expected URL 'example.com' in output, got:\n%s", text)
	}

	// Should contain surrounding text
	if !strings.Contains(text, "Visit") || !strings.Contains(text, "more info") {
		t.Errorf("expected surrounding text in output, got:\n%s", text)
	}

	// Should NOT contain HTML tags
	if strings.Contains(text, "<a href") {
		t.Errorf("expected no HTML tags in text output, got:\n%s", text)
	}

	// Should NOT contain markdown link syntax
	if strings.Contains(text, "[our website](") {
		t.Errorf("expected no markdown syntax in text output, got:\n%s", text)
	}
}

func TestExtractPlainText_Formatting(t *testing.T) {
	html := `<html><body>
		<p>This is <strong>bold</strong> and <em>italic</em> text.</p>
		<p>This has <del>strikethrough</del> text.</p>
	</body></html>`

	converter := NewContentConverter(Text)
	text := converter.extractPlainText(html)

	// Should contain the text content
	if !strings.Contains(text, "bold") {
		t.Errorf("expected 'bold' in text output, got:\n%s", text)
	}
	if !strings.Contains(text, "italic") {
		t.Errorf("expected 'italic' in text output, got:\n%s", text)
	}
	if !strings.Contains(text, "strikethrough") {
		t.Errorf("expected 'strikethrough' in text output, got:\n%s", text)
	}

	// Should NOT contain HTML tags
	if strings.Contains(text, "<strong>") || strings.Contains(text, "<em>") {
		t.Errorf("expected no HTML tags in text output, got:\n%s", text)
	}

	// Should NOT contain markdown syntax
	if strings.Contains(text, "**bold**") || strings.Contains(text, "*italic*") {
		t.Errorf("expected no markdown syntax in text output, got:\n%s", text)
	}
}

func TestExtractPlainText_Scripts(t *testing.T) {
	html := `<html><head>
		<script>console.log("test");</script>
	</head><body>
		<p>Visible content</p>
		<script>alert("popup");</script>
	</body></html>`

	converter := NewContentConverter(Text)
	text := converter.extractPlainText(html)

	// Should contain visible content
	if !strings.Contains(text, "Visible content") {
		t.Errorf("expected 'Visible content' in text output, got:\n%s", text)
	}

	// Should NOT contain script content
	if strings.Contains(text, "console.log") || strings.Contains(text, "alert") {
		t.Errorf("expected no script content in text output, got:\n%s", text)
	}
}

func TestExtractPlainText_Lists(t *testing.T) {
	html := `<html><body>
		<ul>
			<li>Item 1</li>
			<li>Item 2</li>
		</ul>
		<ol>
			<li>First</li>
			<li>Second</li>
		</ol>
	</body></html>`

	converter := NewContentConverter(Text)
	text := converter.extractPlainText(html)

	// Should contain list items
	if !strings.Contains(text, "Item 1") {
		t.Errorf("expected 'Item 1' in text output, got:\n%s", text)
	}
	if !strings.Contains(text, "Item 2") {
		t.Errorf("expected 'Item 2' in text output, got:\n%s", text)
	}
	if !strings.Contains(text, "First") {
		t.Errorf("expected 'First' in text output, got:\n%s", text)
	}
	if !strings.Contains(text, "Second") {
		t.Errorf("expected 'Second' in text output, got:\n%s", text)
	}

	// Should NOT contain HTML tags
	if strings.Contains(text, "<li>") || strings.Contains(text, "<ul>") {
		t.Errorf("expected no HTML tags in text output, got:\n%s", text)
	}
}

func TestExtractPlainText_Minimal(t *testing.T) {
	html := `<html><body>Hello World</body></html>`

	converter := NewContentConverter(Text)
	text := converter.extractPlainText(html)

	// Should contain the text
	if !strings.Contains(text, "Hello World") {
		t.Errorf("expected 'Hello World' in text output, got:\n%s", text)
	}

	// Should be plain text (no tags)
	if strings.Contains(text, "<") || strings.Contains(text, ">") {
		t.Errorf("expected no HTML in text output, got:\n%s", text)
	}
}

func TestExtractPlainText_Empty(t *testing.T) {
	html := `<html><body></body></html>`

	converter := NewContentConverter(Text)
	text := converter.extractPlainText(html)

	// Should be empty or whitespace only
	trimmed := strings.TrimSpace(text)
	if trimmed != "" {
		t.Errorf("expected empty text output, got: %q", text)
	}
}

func TestExtractPlainText_HiddenElements(t *testing.T) {
	htmlBytes, err := os.ReadFile(testutil.Testdata("hidden-elements.html"))
	if err != nil {
		t.Fatalf("failed to read test file: %v", err)
	}

	converter := NewContentConverter(Text)
	text := converter.extractPlainText(string(htmlBytes))

	// Should contain visible content
	if !strings.Contains(text, "Visible content") {
		t.Errorf("expected 'Visible content', got:\n%s", text)
	}

	// Should NOT contain style, script, or noscript content
	if strings.Contains(text, "color: red") {
		t.Error("should not contain CSS")
	}
	if strings.Contains(text, "console.log") {
		t.Error("should not contain JavaScript")
	}
}

type fakePage struct {
	html    string
	htmlErr error
	pdf     []byte
	pdfErr  error
	png     []byte
	pngErr  error
}

func (f fakePage) HTML() (string, error)          { return f.html, f.htmlErr }
func (f fakePage) PDF() ([]byte, error)           { return f.pdf, f.pdfErr }
func (f fakePage) ScreenshotPNG() ([]byte, error) { return f.png, f.pngErr }

func TestProcessContent_MarkdownToFile(t *testing.T) {
	out := filepath.Join(t.TempDir(), "out.md")
	err := ProcessContent(fakePage{html: "<h1>Hello</h1>"}, Markdown, out)
	if err != nil {
		t.Fatalf("ProcessContent: %v", err)
	}
	got, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if !strings.Contains(string(got), "Hello") {
		t.Errorf("expected heading text in markdown, got: %s", got)
	}
}

func TestProcessContent_PDFUsesPagePDF(t *testing.T) {
	out := filepath.Join(t.TempDir(), "out.pdf")
	payload := []byte("%PDF-fake")
	err := ProcessContent(fakePage{html: "<h1>ignored</h1>", pdf: payload}, PDF, out)
	if err != nil {
		t.Fatalf("ProcessContent: %v", err)
	}
	got, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if string(got) != string(payload) {
		t.Errorf("PDF output = %q, want %q", got, payload)
	}
}
