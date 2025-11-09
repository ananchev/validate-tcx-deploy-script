package analyzer

import (
	"testing"

	"github.com/ananchev/validate-tcx-deploy-script/internal/logger"
)

// Test setup - initialize logger for tests
func init() {
	// Initialize logger with minimal configuration for tests
	logger.InitLogger("", "error") // No file, error level only
}

// Tests for FilePathMap.Paths() method

func TestPaths_Relative(t *testing.T) {
	// What: Returns map of relative paths correctly
	fpm := FilePathMap{
		1: FilePathInfo{RelativePath: "file1.xml", AbsolutePath: "/full/path/file1.xml"},
		2: FilePathInfo{RelativePath: "file2.xml", AbsolutePath: "/full/path/file2.xml"},
	}

	result, err := fpm.Paths("relative")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 paths, got %d", len(result))
	}

	if result[1] != "file1.xml" {
		t.Errorf("Expected 'file1.xml', got '%s'", result[1])
	}

	if result[2] != "file2.xml" {
		t.Errorf("Expected 'file2.xml', got '%s'", result[2])
	}
}

func TestPaths_Absolute(t *testing.T) {
	// What: Returns map of absolute paths correctly
	fpm := FilePathMap{
		1: FilePathInfo{RelativePath: "file1.xml", AbsolutePath: "/full/path/file1.xml"},
		2: FilePathInfo{RelativePath: "file2.xml", AbsolutePath: "/full/path/file2.xml"},
	}

	result, err := fpm.Paths("absolute")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 paths, got %d", len(result))
	}

	if result[1] != "/full/path/file1.xml" {
		t.Errorf("Expected '/full/path/file1.xml', got '%s'", result[1])
	}

	if result[2] != "/full/path/file2.xml" {
		t.Errorf("Expected '/full/path/file2.xml', got '%s'", result[2])
	}
}

func TestPaths_InvalidType(t *testing.T) {
	// What: Returns error for invalid path type (not panic!)
	fpm := FilePathMap{
		1: FilePathInfo{RelativePath: "file1.xml", AbsolutePath: "/full/path/file1.xml"},
	}

	result, err := fpm.Paths("invalid")
	if err == nil {
		t.Fatal("Expected error for invalid path type, got nil")
	}

	if result != nil {
		t.Errorf("Expected nil result on error, got %v", result)
	}

	// Verify error message is meaningful
	if err.Error() == "" || len(err.Error()) < 10 {
		t.Errorf("Expected meaningful error message, got: %v", err)
	}
}
