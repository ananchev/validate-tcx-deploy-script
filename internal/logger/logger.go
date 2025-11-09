package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

var (
	InfoLogger      *log.Logger
	DebugLogger     *log.Logger
	ErrorLogger     *log.Logger
	SeparatorLogger *log.Logger
	HeadingLogger   *log.Logger
	logFile         *os.File // Store file handle for cleanup
)

// InitLogger initializes the logging system with the specified log file and level.
// logfile: path to the log file (empty string for stdout only)
// logLevel: "debug", "info", or "error" to control verbosity
func InitLogger(logfile string, logLevel string) error {
	var multi_writer io.Writer
	if logfile == "" {
		multi_writer = io.MultiWriter(os.Stdout)
	} else {
		file, err := os.OpenFile(logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0755)
		if err != nil {
			return fmt.Errorf("failed to open log file: %w", err)
		}
		logFile = file // Store for later cleanup
		multi_writer = io.MultiWriter(os.Stdout, file)
	}

	var debug_writer io.Writer
	var info_writer io.Writer

	if logLevel == "debug" {
		debug_writer = multi_writer
		info_writer = multi_writer
	} else if logLevel == "info" {
		debug_writer = io.Discard
		info_writer = multi_writer
	} else {
		debug_writer = io.Discard
		info_writer = io.Discard
	}

	InfoLogger = log.New(info_writer, "INFO: ", 0)
	ErrorLogger = log.New(multi_writer, "ERROR: ", 0)
	DebugLogger = log.New(debug_writer, "DEBUG: ", 0)
	SeparatorLogger = log.New(multi_writer, "", 0)
	HeadingLogger = log.New(multi_writer, "", log.Ldate|log.Ltime)

	return nil
}

// Close closes the log file if one was opened.
// Should be called with defer in main to ensure cleanup.
func Close() error {
	if logFile != nil {
		return logFile.Close()
	}
	return nil
}

// format_string replaces {key} placeholders with corresponding values.
// args should be provided as alternating key, value pairs.
func format_string(format string, args ...interface{}) string {
	if len(args) == 0 {
		return format
	}

	// Build replacement map for better performance
	replacements := make(map[string]string, len(args)/2)
	for i := 0; i < len(args)-1; i += 2 {
		key := fmt.Sprintf("{%v}", args[i])
		value := fmt.Sprint(args[i+1])
		replacements[key] = value
	}

	// Perform all replacements
	result := format
	for key, value := range replacements {
		result = strings.ReplaceAll(result, key, value)
	}
	return result
}

func write_to_log(loggerType int, format string, args ...interface{}) {
	log_msg := format_string(format, args...)
	switch loggerType {
	case 1:
		ErrorLogger.Println(log_msg)
	case 2:
		InfoLogger.Println(log_msg)
	case 3:
		// below adds caller info to the string to be logged
		// _, fn, line, _ := runtime.Caller(1)
		// format = filepath.Base(fn) + ":" + strconv.Itoa(line) + ": " + format
		DebugLogger.Println(log_msg)
	case 4:
		SeparatorLogger.Println(log_msg)
	case 5:
		HeadingLogger.Println(log_msg)
	}
}

// Error logs an error message. Always visible regardless of log level.
func Error(format string, args ...interface{}) {
	write_to_log(1, format, args...)
}

// Info logs an informational message. Visible when log level is "info" or "debug".
func Info(format string, args ...interface{}) {
	write_to_log(2, format, args...)
}

// Debug logs a debug message. Only visible when log level is "debug".
func Debug(format string, args ...interface{}) {
	write_to_log(3, format, args...)
}

// Separate logs a separator line without prefix. Always visible.
func Separate(format string, args ...interface{}) {
	write_to_log(4, format, args...)
}

// Heading logs a heading with timestamp. Always visible.
func Heading(format string, args ...interface{}) {
	write_to_log(5, format, args...)
}
