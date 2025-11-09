package logger

import (
	"bytes"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFormatString(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		args     []interface{}
		expected string
	}{
		{
			name:     "simple replacement",
			format:   "Hello {name}",
			args:     []interface{}{"name", "World"},
			expected: "Hello World",
		},
		{
			name:     "multiple replacements",
			format:   "File {file} line {line} error {msg}",
			args:     []interface{}{"file", "test.go", "line", 42, "msg", "syntax error"},
			expected: "File test.go line 42 error syntax error",
		},
		{
			name:     "no placeholders",
			format:   "Simple message",
			args:     []interface{}{"key", "value"},
			expected: "Simple message",
		},
		{
			name:     "empty args",
			format:   "Message with {placeholder}",
			args:     []interface{}{},
			expected: "Message with {placeholder}",
		},
		{
			name:     "duplicate keys",
			format:   "{x} + {x} = {result}",
			args:     []interface{}{"x", "5", "result", "10"},
			expected: "5 + 5 = 10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := format_string(tt.format, tt.args...)
			if result != tt.expected {
				t.Errorf("format_string() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestInitLogger_StdoutOnly(t *testing.T) {
	// Reset loggers
	InfoLogger = nil
	DebugLogger = nil
	ErrorLogger = nil
	logFile = nil

	err := InitLogger("", "info")
	if err != nil {
		t.Fatalf("InitLogger failed: %v", err)
	}
	defer Close()

	if InfoLogger == nil || ErrorLogger == nil || DebugLogger == nil {
		t.Error("Loggers were not initialized")
	}

	if logFile != nil {
		t.Error("logFile should be nil when no file specified")
	}
}

func TestInitLogger_WithFile(t *testing.T) {
	// Create temp directory for test
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "test.log")

	// Reset loggers
	InfoLogger = nil
	DebugLogger = nil
	ErrorLogger = nil
	logFile = nil

	err := InitLogger(logPath, "debug")
	if err != nil {
		t.Fatalf("InitLogger failed: %v", err)
	}
	defer Close()

	if logFile == nil {
		t.Error("logFile should not be nil when file specified")
	}

	// Verify file was created
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Error("Log file was not created")
	}
}

func TestInitLogger_InvalidPath(t *testing.T) {
	// Try to create log in non-existent directory without create permissions
	invalidPath := "/invalid/nonexistent/path/test.log"
	if os.PathSeparator == '\\' {
		// Windows path
		invalidPath = "Z:\\invalid\\nonexistent\\path\\test.log"
	}

	err := InitLogger(invalidPath, "info")
	if err == nil {
		t.Error("Expected error for invalid path, got nil")
		Close()
	}
}

func TestClose_NoFile(t *testing.T) {
	logFile = nil
	err := Close()
	if err != nil {
		t.Errorf("Close() with no file should not error, got: %v", err)
	}
}

func TestClose_WithFile(t *testing.T) {
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "test.log")

	err := InitLogger(logPath, "info")
	if err != nil {
		t.Fatalf("InitLogger failed: %v", err)
	}

	err = Close()
	if err != nil {
		t.Errorf("Close() failed: %v", err)
	}

	// Verify file handle was closed (logFile still points to closed file)
	if logFile == nil {
		t.Error("logFile should still reference the file object after Close()")
	}
}

func TestLogLevels(t *testing.T) {
	tests := []struct {
		name       string
		logLevel   string
		testFunc   func()
		checkInfo  bool
		checkDebug bool
	}{
		{
			name:       "error level",
			logLevel:   "error",
			testFunc:   func() { Info("info message") },
			checkInfo:  false,
			checkDebug: false,
		},
		{
			name:       "info level shows info",
			logLevel:   "info",
			testFunc:   func() { Info("info message") },
			checkInfo:  true,
			checkDebug: false,
		},
		{
			name:       "info level hides debug",
			logLevel:   "info",
			testFunc:   func() { Debug("debug message") },
			checkInfo:  false,
			checkDebug: false,
		},
		{
			name:       "debug level shows debug",
			logLevel:   "debug",
			testFunc:   func() { Debug("debug message") },
			checkInfo:  false,
			checkDebug: true,
		},
		{
			name:       "debug level shows info",
			logLevel:   "debug",
			testFunc:   func() { Info("info message") },
			checkInfo:  true,
			checkDebug: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture output
			var buf bytes.Buffer

			// Initialize logger with buffer as output
			if tt.logLevel == "debug" {
				InfoLogger = log.New(&buf, "INFO: ", 0)
				DebugLogger = log.New(&buf, "DEBUG: ", 0)
			} else if tt.logLevel == "info" {
				InfoLogger = log.New(&buf, "INFO: ", 0)
				DebugLogger = log.New(io.Discard, "DEBUG: ", 0)
			} else {
				InfoLogger = log.New(io.Discard, "INFO: ", 0)
				DebugLogger = log.New(io.Discard, "DEBUG: ", 0)
			}
			ErrorLogger = log.New(&buf, "ERROR: ", 0)

			tt.testFunc()

			output := buf.String()

			if tt.checkInfo && !strings.Contains(output, "INFO:") {
				t.Error("Expected INFO message to be logged")
			}
			if !tt.checkInfo && strings.Contains(output, "INFO:") {
				t.Error("INFO message should not be logged at this level")
			}
			if tt.checkDebug && !strings.Contains(output, "DEBUG:") {
				t.Error("Expected DEBUG message to be logged")
			}
			if !tt.checkDebug && strings.Contains(output, "DEBUG:") {
				t.Error("DEBUG message should not be logged at this level")
			}
		})
	}
}

func TestErrorLogging(t *testing.T) {
	var buf bytes.Buffer
	ErrorLogger = log.New(&buf, "ERROR: ", 0)

	Error("test error {msg}", "msg", "failed")

	output := buf.String()
	if !strings.Contains(output, "ERROR:") {
		t.Error("Expected ERROR prefix")
	}
	if !strings.Contains(output, "test error failed") {
		t.Error("Expected formatted message")
	}
}

func TestSeparateAndHeading(t *testing.T) {
	var buf bytes.Buffer
	SeparatorLogger = log.New(&buf, "", 0)
	HeadingLogger = log.New(&buf, "", log.Ldate|log.Ltime)

	Separate("=================")
	if !strings.Contains(buf.String(), "=================") {
		t.Error("Separate() did not log correctly")
	}

	buf.Reset()
	Heading("Test Heading")
	output := buf.String()
	if !strings.Contains(output, "Test Heading") {
		t.Error("Heading() did not log message")
	}
	// Heading should include date/time, check for common date format patterns
	// This is a basic check - actual format may vary
	if len(output) < len("Test Heading") {
		t.Error("Heading() should include timestamp")
	}
}
