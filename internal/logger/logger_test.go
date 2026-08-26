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

func TestShouldUseColor(t *testing.T) {
	// Note: This function checks environment variable and terminal status
	// We can only reliably test the NO_COLOR environment variable behavior

	// Save original NO_COLOR value
	originalNOCOLOR := os.Getenv("NO_COLOR")
	defer func() {
		if originalNOCOLOR != "" {
			os.Setenv("NO_COLOR", originalNOCOLOR)
		} else {
			os.Unsetenv("NO_COLOR")
		}
	}()

	// Test with NO_COLOR set
	t.Run("with NO_COLOR set", func(t *testing.T) {
		os.Setenv("NO_COLOR", "1")
		result := shouldUseColor()
		if result != false {
			t.Errorf("shouldUseColor() with NO_COLOR=1 should return false, got %v", result)
		}
	})

	// Test with NO_COLOR unset (result depends on terminal status)
	t.Run("with NO_COLOR unset", func(t *testing.T) {
		os.Unsetenv("NO_COLOR")
		result := shouldUseColor()
		// Result can be true or false depending on whether stderr is a terminal
		// Just verify it returns a boolean without panicking
		_ = result
	})

	// Test with empty NO_COLOR (should behave as unset)
	t.Run("with NO_COLOR empty", func(t *testing.T) {
		os.Setenv("NO_COLOR", "")
		result := shouldUseColor()
		// Empty NO_COLOR should behave as unset
		// Result depends on terminal status
		_ = result
	})
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

			// Color setting depends on shouldUseColor() which depends on environment
			// Just verify it's a boolean
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
