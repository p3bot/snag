// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package logger

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

// newTestLogger creates a logger for testing that writes to a buffer
func newTestLogger(level LogLevel, writer io.Writer) *Logger {
	return NewWithWriter(level, writer, false)
}

func TestLogger_Success(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(LevelNormal, &buf)

	logger.Success("Operation completed")

	output := buf.String()
	if !strings.Contains(output, "Operation completed") {
		t.Errorf("expected success message in output, got: %s", output)
	}
}

func TestLogger_Error(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(LevelNormal, &buf)

	logger.Error("Something went wrong")

	output := buf.String()
	if !strings.Contains(output, "Something went wrong") {
		t.Errorf("expected error message in output, got: %s", output)
	}
}

func TestLogger_Info(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(LevelNormal, &buf)

	logger.Info("Informational message")

	output := buf.String()
	if !strings.Contains(output, "Informational message") {
		t.Errorf("expected info message in output, got: %s", output)
	}
}

func TestLogger_Debug(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(LevelDebug, &buf)

	logger.Debug("Debug information")

	output := buf.String()
	if !strings.Contains(output, "Debug information") {
		t.Errorf("expected debug message in output, got: %s", output)
	}
}

func TestLogger_QuietMode(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(LevelQuiet, &buf)

	// These should be suppressed in quiet mode
	logger.Success("Success message")
	logger.Info("Info message")
	logger.Debug("Debug message")

	output := buf.String()
	if strings.Contains(output, "Success message") {
		t.Errorf("quiet mode should suppress success messages, got: %s", output)
	}
	if strings.Contains(output, "Info message") {
		t.Errorf("quiet mode should suppress info messages, got: %s", output)
	}

	// Errors should still appear in quiet mode
	logger.Error("Error message")
	output = buf.String()
	if !strings.Contains(output, "Error message") {
		t.Errorf("quiet mode should still show error messages, got: %s", output)
	}
}

func TestLogger_VerboseMode(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(LevelVerbose, &buf)

	logger.Success("Success message")
	logger.Info("Info message")

	output := buf.String()
	if !strings.Contains(output, "Success message") {
		t.Errorf("verbose mode should show success messages, got: %s", output)
	}
	if !strings.Contains(output, "Info message") {
		t.Errorf("verbose mode should show info messages, got: %s", output)
	}

	// Debug messages should NOT appear in verbose mode (only in debug mode)
	buf.Reset()
	logger.Debug("Debug message")
	output = buf.String()
	if strings.Contains(output, "Debug message") {
		t.Errorf("verbose mode should not show debug messages, got: %s", output)
	}
}

func TestLogger_StderrOnly(t *testing.T) {
	// This test verifies that logger writes to the provided writer (stderr in practice)
	// and not to stdout
	var stderr bytes.Buffer
	logger := newTestLogger(LevelNormal, &stderr)

	logger.Info("This should go to stderr")

	if stderr.Len() == 0 {
		t.Error("expected logger to write to stderr buffer")
	}

	output := stderr.String()
	if !strings.Contains(output, "This should go to stderr") {
		t.Errorf("expected message in stderr output, got: %s", output)
	}
}

func TestUseColor(t *testing.T) {
	tests := []struct {
		name      string
		mode      string
		noColor   string
		force     string
		clicolor  string
		term      string
		stderrTTY bool
		want      bool
	}{
		{name: "always overrides NO_COLOR and FORCE_COLOR", mode: ColorAlways, noColor: "1", force: "1", term: "xterm", stderrTTY: true, want: true},
		{name: "empty NO_COLOR is unset", mode: ColorAuto, noColor: "", force: "1", term: "xterm", stderrTTY: false, want: true},
		{name: "never", mode: ColorNever, term: "xterm", stderrTTY: true, want: false},
		{name: "never ignores FORCE_COLOR", mode: ColorNever, force: "1", term: "xterm", stderrTTY: true, want: false},
		{name: "never ignores NO_COLOR", mode: ColorNever, noColor: "1", term: "xterm", stderrTTY: true, want: false},
		{name: "always overrides non-TTY", mode: ColorAlways, term: "xterm", stderrTTY: false, want: true},
		{name: "always overrides TERM=dumb", mode: ColorAlways, term: "dumb", stderrTTY: false, want: true},
		{name: "auto FORCE_COLOR colours non-TTY TERM=dumb", mode: ColorAuto, force: "1", term: "dumb", stderrTTY: false, want: true},
		{name: "auto CLICOLOR_FORCE colours non-TTY TERM=dumb", mode: ColorAuto, clicolor: "1", term: "dumb", stderrTTY: false, want: true},
		{name: "auto TERM=dumb without force", mode: ColorAuto, term: "dumb", stderrTTY: true, want: false},
		{name: "auto TTY", mode: ColorAuto, term: "xterm", stderrTTY: true, want: true},
		{name: "auto non-TTY", mode: ColorAuto, term: "xterm", stderrTTY: false, want: false},
		{name: "FORCE_COLOR=0 is falsy", mode: ColorAuto, force: "0", term: "xterm", stderrTTY: false, want: false},
		{name: "FORCE_COLOR=false is falsy", mode: ColorAuto, force: "false", term: "xterm", stderrTTY: false, want: false},
		{name: "FORCE_COLOR=yes is truthy", mode: ColorAuto, force: "yes", term: "xterm", stderrTTY: false, want: true},
		{name: "NO_COLOR wins over FORCE_COLOR in auto", mode: ColorAuto, noColor: "1", force: "1", term: "xterm", stderrTTY: true, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("NO_COLOR", tt.noColor)
			t.Setenv("FORCE_COLOR", tt.force)
			t.Setenv("CLICOLOR_FORCE", tt.clicolor)
			t.Setenv("TERM", tt.term)
			got := UseColor(tt.mode, tt.stderrTTY)
			if got != tt.want {
				t.Errorf("UseColor(%q, %v) = %v, want %v", tt.mode, tt.stderrTTY, got, tt.want)
			}
		})
	}
}

func TestResolveColor(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	t.Setenv("FORCE_COLOR", "")
	t.Setenv("CLICOLOR_FORCE", "")
	t.Setenv("TERM", "dumb")

	got, err := ResolveColor(ColorAlways)
	if err != nil {
		t.Fatalf("ResolveColor(always): %v", err)
	}
	if !got {
		t.Fatal("ResolveColor(always) with TERM=dumb: want true")
	}

	got, err = ResolveColor("Always")
	if err != nil {
		t.Fatalf("ResolveColor(Always): %v", err)
	}
	if !got {
		t.Fatal("ResolveColor(Always) with TERM=dumb: want true")
	}

	got, err = ResolveColor(" always ")
	if err != nil {
		t.Fatalf("ResolveColor(%q): %v", " always ", err)
	}
	if !got {
		t.Fatal("ResolveColor(\" always \") with TERM=dumb: want true")
	}

	got, err = ResolveColor("NEVER")
	if err != nil {
		t.Fatalf("ResolveColor(NEVER): %v", err)
	}
	if got {
		t.Fatal("ResolveColor(NEVER): want false")
	}

	got, err = ResolveColor("")
	if err == nil {
		t.Fatal("ResolveColor empty: want error")
	}
	if got {
		t.Fatal("ResolveColor empty: want color false")
	}

	got, err = ResolveColor("  ")
	if err == nil {
		t.Fatal("ResolveColor whitespace: want error")
	}
	if got {
		t.Fatal("ResolveColor whitespace: want color false")
	}

	got, err = ResolveColor("rainbow")
	if err == nil {
		t.Fatal("ResolveColor(rainbow): want error")
	}
	if got {
		t.Fatal("ResolveColor(rainbow): want color false")
	}
	msg := err.Error()
	for _, needle := range []string{"rainbow", "auto", "always", "never"} {
		if !strings.Contains(msg, needle) {
			t.Errorf("ResolveColor error %q missing %q", msg, needle)
		}
	}
	if strings.Contains(msg, "Code:") || strings.Contains(msg, "Fix:") || strings.Contains(msg, "ValidValues") {
		t.Errorf("ResolveColor error has structured envelope: %q", msg)
	}
}

func TestFormatErrorLine(t *testing.T) {
	got := FormatErrorLine(io.EOF, false)
	if strings.Contains(got, "\x1b[") {
		t.Errorf("FormatErrorLine with color off has ANSI: %q", got)
	}
	if got != "Error: EOF" {
		t.Errorf("FormatErrorLine = %q, want %q", got, "Error: EOF")
	}

	got = FormatErrorLine(io.EOF, true)
	if !strings.Contains(got, "\x1b[") {
		t.Errorf("FormatErrorLine with color on missing ANSI: %q", got)
	}
	if !strings.Contains(got, "Error:") || !strings.Contains(got, "EOF") {
		t.Errorf("FormatErrorLine = %q, want Error: prefix and message", got)
	}
}

func TestNewLogger(t *testing.T) {
	// Test all log levels
	tests := []struct {
		name  string
		level LogLevel
	}{
		{"quiet level", LevelQuiet},
		{"normal level", LevelNormal},
		{"verbose level", LevelVerbose},
		{"debug level", LevelDebug},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := New(tt.level)

			if logger == nil {
				t.Fatal("NewLogger returned nil")
			}

			if logger.level != tt.level {
				t.Errorf("NewLogger(%v) level = %v, expected %v", tt.level, logger.level, tt.level)
			}

			if logger.writer == nil {
				t.Error("NewLogger created logger with nil writer")
			}

			// Verify writer is os.Stderr in production
			if logger.writer != os.Stderr {
				t.Errorf("NewLogger created logger with writer %v, expected os.Stderr", logger.writer)
			}

			_ = logger.color
		})
	}
}

func TestSetDefault_PackageFuncs(t *testing.T) {
	t.Cleanup(func() { SetDefault(New(LevelNormal)) })

	var buf bytes.Buffer
	SetDefault(NewWithWriter(LevelNormal, &buf, false))
	Info("from default")
	if !strings.Contains(buf.String(), "from default") {
		t.Errorf("Info after SetDefault = %q, want message", buf.String())
	}

	SetDefault(nil)
	Error("still works")
}

func TestSetDefault_Concurrent(t *testing.T) {
	t.Cleanup(func() { SetDefault(New(LevelNormal)) })

	var buf bytes.Buffer
	SetDefault(NewWithWriter(LevelNormal, &buf, false))

	done := make(chan struct{})
	go func() {
		for i := 0; i < 1000; i++ {
			Info("concurrent")
		}
		close(done)
	}()
	for i := 0; i < 1000; i++ {
		SetDefault(NewWithWriter(LevelQuiet, io.Discard, false))
		SetDefault(NewWithWriter(LevelNormal, &buf, false))
	}
	<-done
}
