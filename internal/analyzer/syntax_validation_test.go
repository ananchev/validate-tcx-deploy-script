package analyzer

import (
	"strings"
	"testing"
)

// TestValidatePathSeparators_WindowsValid tests Windows path with only backslashes
// What it tests: Windows script with "config\data\file.xml" -> Valid (no error)
func TestValidatePathSeparators_WindowsValid(t *testing.T) {
	err := validatePathSeparators("config\\data\\file.xml", "windows", 10)

	if err != nil {
		t.Errorf("Expected no error for Windows path with backslashes, got: %v", err)
	}
}

// TestValidatePathSeparators_WindowsInvalidForwardSlash tests Windows path with forward slashes
// What it tests: Windows script with "config/data/file.xml" -> Error (forward slash not allowed)
func TestValidatePathSeparators_WindowsInvalidForwardSlash(t *testing.T) {
	err := validatePathSeparators("config/data/file.xml", "windows", 10)

	if err == nil {
		t.Error("Expected error for Windows path with forward slashes, got nil")
	}
	if err != nil && !strings.Contains(err.Error(), "forward slashes") {
		t.Errorf("Error should mention forward slashes, got: %v", err)
	}
	if err != nil && !strings.Contains(err.Error(), "line 10") {
		t.Errorf("Error should mention line number, got: %v", err)
	}
}

// TestValidatePathSeparators_LinuxValid tests Linux path with only forward slashes
// What it tests: Linux script with "config/data/file.xml" -> Valid (no error)
func TestValidatePathSeparators_LinuxValid(t *testing.T) {
	err := validatePathSeparators("config/data/file.xml", "linux", 15)

	if err != nil {
		t.Errorf("Expected no error for Linux path with forward slashes, got: %v", err)
	}
}

// TestValidatePathSeparators_LinuxInvalidBackslash tests Linux path with backslashes
// What it tests: Linux script with "config\data\file.xml" -> Error (backslash not allowed)
func TestValidatePathSeparators_LinuxInvalidBackslash(t *testing.T) {
	err := validatePathSeparators("config\\data\\file.xml", "linux", 15)

	if err == nil {
		t.Error("Expected error for Linux path with backslashes, got nil")
	}
	if err != nil && !strings.Contains(err.Error(), "backslashes") {
		t.Errorf("Error should mention backslashes, got: %v", err)
	}
	if err != nil && !strings.Contains(err.Error(), "line 15") {
		t.Errorf("Error should mention line number, got: %v", err)
	}
}

// TestValidatePathSeparators_MixedSeparatorsWindows tests mixed separators on Windows
// What it tests: Windows script with "config/subfolder\file.xml" -> Error (mixed not allowed)
func TestValidatePathSeparators_MixedSeparatorsWindows(t *testing.T) {
	err := validatePathSeparators("config/subfolder\\file.xml", "windows", 20)

	if err == nil {
		t.Error("Expected error for mixed separators in Windows path, got nil")
	}
	if err != nil && !strings.Contains(err.Error(), "forward slashes") {
		t.Errorf("Error should mention forward slashes, got: %v", err)
	}
}

// TestValidatePathSeparators_MixedSeparatorsLinux tests mixed separators on Linux
// What it tests: Linux script with "config\subfolder/file.xml" -> Error (mixed not allowed)
func TestValidatePathSeparators_MixedSeparatorsLinux(t *testing.T) {
	err := validatePathSeparators("config\\subfolder/file.xml", "linux", 25)

	if err == nil {
		t.Error("Expected error for mixed separators in Linux path, got nil")
	}
	if err != nil && !strings.Contains(err.Error(), "backslashes") {
		t.Errorf("Error should mention backslashes, got: %v", err)
	}
}

// TestValidatePathSeparators_NoSeparatorsWindows tests filename only (no path separators) on Windows
// What it tests: Windows script with "file.xml" -> Valid (no separators to validate)
func TestValidatePathSeparators_NoSeparatorsWindows(t *testing.T) {
	err := validatePathSeparators("file.xml", "windows", 30)

	if err != nil {
		t.Errorf("Expected no error for filename without separators, got: %v", err)
	}
}

// TestValidatePathSeparators_NoSeparatorsLinux tests filename only (no path separators) on Linux
// What it tests: Linux script with "file.xml" -> Valid (no separators to validate)
func TestValidatePathSeparators_NoSeparatorsLinux(t *testing.T) {
	err := validatePathSeparators("file.xml", "linux", 35)

	if err != nil {
		t.Errorf("Expected no error for filename without separators, got: %v", err)
	}
}

// TestValidatePathSeparators_MultipleBackslashesWindows tests Windows path with multiple backslashes
// What it tests: Windows script with "config\\data\\subdir\\file.xml" -> Valid
func TestValidatePathSeparators_MultipleBackslashesWindows(t *testing.T) {
	err := validatePathSeparators("config\\data\\subdir\\file.xml", "windows", 40)

	if err != nil {
		t.Errorf("Expected no error for Windows path with multiple backslashes, got: %v", err)
	}
}

// TestValidatePathSeparators_MultipleForwardSlashesLinux tests Linux path with multiple forward slashes
// What it tests: Linux script with "config/data/subdir/file.xml" -> Valid
func TestValidatePathSeparators_MultipleForwardSlashesLinux(t *testing.T) {
	err := validatePathSeparators("config/data/subdir/file.xml", "linux", 45)

	if err != nil {
		t.Errorf("Expected no error for Linux path with multiple forward slashes, got: %v", err)
	}
}

// TestValidatePathSeparators_WindowsAbsolutePath tests Windows absolute path
// What it tests: Windows script with "C:\Program Files\TC\file.xml" -> Valid
func TestValidatePathSeparators_WindowsAbsolutePath(t *testing.T) {
	err := validatePathSeparators("C:\\Program Files\\TC\\file.xml", "windows", 50)

	if err != nil {
		t.Errorf("Expected no error for Windows absolute path, got: %v", err)
	}
}

// TestValidatePathSeparators_LinuxAbsolutePath tests Linux absolute path
// What it tests: Linux script with "/opt/teamcenter/config/file.xml" -> Valid
func TestValidatePathSeparators_LinuxAbsolutePath(t *testing.T) {
	err := validatePathSeparators("/opt/teamcenter/config/file.xml", "linux", 55)

	if err != nil {
		t.Errorf("Expected no error for Linux absolute path, got: %v", err)
	}
}

