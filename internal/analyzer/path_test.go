package analyzer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ananchev/validate-tcx-deploy-script/internal/logger"
)

// Test setup - initialize logger for tests
func init() {
	// Initialize logger with minimal configuration for tests
	logger.InitLogger("", "error") // No file, error level only
}

// Tests for fileExists() and checkFilePathsInScript()

func TestFileExists_True(t *testing.T) {
	// What: File exists returns true
	// Setup: Create a temp file
	tmpDir, err := os.MkdirTemp("", "path-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "testfile.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Set sourceCodeRoot to temp dir for test
	originalRoot := sourceCodeRoot
	sourceCodeRoot = tmpDir
	defer func() { sourceCodeRoot = originalRoot }()

	// Test
	result := fileExists("testfile.txt")
	if !result {
		t.Errorf("Expected fileExists to return true for existing file")
	}
}

func TestFileExists_False(t *testing.T) {
	// What: File missing returns false
	tmpDir, err := os.MkdirTemp("", "path-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Set sourceCodeRoot to temp dir for test
	originalRoot := sourceCodeRoot
	sourceCodeRoot = tmpDir
	defer func() { sourceCodeRoot = originalRoot }()

	// Test with non-existent file
	result := fileExists("nonexistent.txt")
	if result {
		t.Errorf("Expected fileExists to return false for missing file")
	}
}
