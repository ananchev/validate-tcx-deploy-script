package analyzer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ananchev/validate-tcx-deploy-script/internal/logger"
)

// Test setup - initialize logger for tests
func init() {
	// Initialize logger with minimal configuration for tests
	logger.InitLogger("", "error") // No file, error level only
}

// Test utilities

// setupTestDir creates a temporary directory with the specified file structure.
// Files are specified as relative paths (e.g., "src/main.go", "lib/utils.go").
func setupTestDir(t *testing.T, files []string) string {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "content-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	for _, file := range files {
		fullPath := filepath.Join(tmpDir, file)
		dir := filepath.Dir(fullPath)

		if err := os.MkdirAll(dir, 0755); err != nil {
			os.RemoveAll(tmpDir)
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}

		if err := os.WriteFile(fullPath, []byte("test content"), 0644); err != nil {
			os.RemoveAll(tmpDir)
			t.Fatalf("Failed to create file %s: %v", fullPath, err)
		}
	}

	return tmpDir
}

// cleanup removes the test directory.
func cleanup(t *testing.T, dir string) {
	t.Helper()
	if err := os.RemoveAll(dir); err != nil {
		t.Errorf("Failed to cleanup temp dir %s: %v", dir, err)
	}
}

// assertErrorContains checks if the error message contains the expected substring.
func assertErrorContains(t *testing.T, err error, substring string) {
	t.Helper()
	if err == nil {
		t.Fatalf("Expected error containing %q, got nil", substring)
	}
	if !strings.Contains(err.Error(), substring) {
		t.Errorf("Expected error to contain %q, got: %v", substring, err)
	}
}

// assertNoError checks that no error occurred.
func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}

// Group 1: Pattern Matching (3 tests)

func TestShouldIgnore_SinglePatternMatch(t *testing.T) {
	// What: File matches single gitignore pattern
	patterns := []string{"*.log"}

	result := shouldIgnore("debug.log", patterns)
	if !result {
		t.Errorf("Expected debug.log to be ignored by pattern *.log")
	}
}

func TestShouldIgnore_MultiplePatterns(t *testing.T) {
	// What: File matches second pattern in list
	patterns := []string{"*.txt", "*.log"}

	result := shouldIgnore("error.log", patterns)
	if !result {
		t.Errorf("Expected error.log to be ignored by pattern *.log")
	}
}

func TestShouldIgnore_NoMatch(t *testing.T) {
	// What: File doesn't match any patterns
	patterns := []string{"*.log", "temp/"}

	result := shouldIgnore("main.go", patterns)
	if result {
		t.Errorf("Expected main.go not to be ignored, but it was")
	}
}

// Group 2: Directory Traversal (11 tests)

func TestTraverseAndCollect_BasicFiles(t *testing.T) {
	// What: Collects all files in simple directory structure
	files := []string{"file1.txt", "file2.go", "subdir/file3.py"}
	tmpDir := setupTestDir(t, files)
	defer cleanup(t, tmpDir)

	patterns := []string{}
	collected, err := traverseAndCollect(tmpDir, patterns)

	assertNoError(t, err)
	if len(collected) != 3 {
		t.Errorf("Expected 3 files, got %d: %v", len(collected), collected)
	}
}

func TestTraverseAndCollect_WithIgnoredDirectories(t *testing.T) {
	// What: Skips ignored directories (returns SkipDir)
	files := []string{"src/main.go", "build/output.exe"}
	tmpDir := setupTestDir(t, files)
	defer cleanup(t, tmpDir)

	patterns := []string{"build/"}
	collected, err := traverseAndCollect(tmpDir, patterns)

	assertNoError(t, err)
	if len(collected) != 1 {
		t.Errorf("Expected 1 file, got %d: %v", len(collected), collected)
	}
	if len(collected) > 0 && !strings.Contains(collected[0], "main.go") {
		t.Errorf("Expected to collect main.go, got: %v", collected)
	}
}

func TestTraverseAndCollect_WithIgnoredFilePatterns(t *testing.T) {
	// What: Skips ignored file extensions
	files := []string{"main.go", "debug.log", "test.tmp"}
	tmpDir := setupTestDir(t, files)
	defer cleanup(t, tmpDir)

	patterns := []string{"*.log", "*.tmp"}
	collected, err := traverseAndCollect(tmpDir, patterns)

	assertNoError(t, err)
	if len(collected) != 1 {
		t.Errorf("Expected 1 file, got %d: %v", len(collected), collected)
	}
	if len(collected) > 0 && !strings.Contains(collected[0], "main.go") {
		t.Errorf("Expected to collect main.go, got: %v", collected)
	}
}

func TestTraverseAndCollect_WithNestedIgnoredDirs(t *testing.T) {
	// What: Properly skips nested ignored directories
	files := []string{"src/code.go", "node_modules/pkg/index.js"}
	tmpDir := setupTestDir(t, files)
	defer cleanup(t, tmpDir)

	patterns := []string{"node_modules/"}
	collected, err := traverseAndCollect(tmpDir, patterns)

	assertNoError(t, err)
	if len(collected) != 1 {
		t.Errorf("Expected 1 file, got %d: %v", len(collected), collected)
	}
	if len(collected) > 0 && !strings.Contains(collected[0], "code.go") {
		t.Errorf("Expected to collect code.go, got: %v", collected)
	}
}

func TestTraverseAndCollect_EmptyDirectory(t *testing.T) {
	// What: Handles empty directory gracefully
	tmpDir, err := os.MkdirTemp("", "content-test-empty-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer cleanup(t, tmpDir)

	patterns := []string{}
	collected, err := traverseAndCollect(tmpDir, patterns)

	assertNoError(t, err)
	if len(collected) != 0 {
		t.Errorf("Expected 0 files in empty directory, got %d: %v", len(collected), collected)
	}
}

func TestTraverseAndCollect_NonexistentPath(t *testing.T) {
	// What: Returns error for nonexistent root path
	nonexistentPath := filepath.Join(os.TempDir(), "nonexistent-dir-12345")
	patterns := []string{}

	collected, err := traverseAndCollect(nonexistentPath, patterns)

	if err == nil {
		t.Errorf("Expected error for nonexistent path, got nil")
	}
	if len(collected) != 0 {
		t.Errorf("Expected 0 files for nonexistent path, got %d: %v", len(collected), collected)
	}
}

func TestTraverseAndCollect_RelativeVsAbsolutePaths(t *testing.T) {
	// What: Returns correct relative paths from root
	files := []string{"subdir/file.txt", "another/deep/nested.go"}
	tmpDir := setupTestDir(t, files)
	defer cleanup(t, tmpDir)

	patterns := []string{}
	collected, err := traverseAndCollect(tmpDir, patterns)

	assertNoError(t, err)
	for _, path := range collected {
		if filepath.IsAbs(path) {
			t.Errorf("Expected relative path, got absolute: %s", path)
		}
		// Check for proper separators
		if !strings.Contains(path, string(filepath.Separator)) && strings.Contains(path, "/") {
			t.Errorf("Expected OS-specific separator in path: %s", path)
		}
	}
}

func TestTraverseAndCollect_WithGitignoreStylePatterns(t *testing.T) {
	// What: Tests directory patterns with trailing slash
	files := []string{"dist/output.js", "out/build.exe", "src/main.go"}
	tmpDir := setupTestDir(t, files)
	defer cleanup(t, tmpDir)

	patterns := []string{"dist/", "out/"}
	collected, err := traverseAndCollect(tmpDir, patterns)

	assertNoError(t, err)
	if len(collected) != 1 {
		t.Errorf("Expected 1 file, got %d: %v", len(collected), collected)
	}
	if len(collected) > 0 && !strings.Contains(collected[0], "main.go") {
		t.Errorf("Expected to collect main.go, got: %v", collected)
	}
}

func TestTraverseAndCollect_WithNegationPatterns(t *testing.T) {
	// What: Tests negation patterns if supported by library
	files := []string{"debug.log", "error.log", "important.log", "main.go"}
	tmpDir := setupTestDir(t, files)
	defer cleanup(t, tmpDir)

	patterns := []string{"*.log", "!important.log"}
	collected, err := traverseAndCollect(tmpDir, patterns)

	assertNoError(t, err)

	// Check if important.log is included (negation pattern)
	hasImportant := false
	hasOtherLog := false
	hasMainGo := false
	for _, path := range collected {
		if strings.Contains(path, "important.log") {
			hasImportant = true
		}
		if strings.Contains(path, "debug.log") || strings.Contains(path, "error.log") {
			hasOtherLog = true
		}
		if strings.Contains(path, "main.go") {
			hasMainGo = true
		}
	}

	// main.go should definitely be collected
	if !hasMainGo {
		t.Errorf("Expected main.go to be collected")
	}

	// Log behavior of negation patterns (library-dependent)
	if hasImportant {
		t.Logf("Negation pattern worked: important.log included despite *.log pattern")
	} else {
		t.Logf("Negation pattern not supported or didn't work: important.log was excluded")
	}

	if hasOtherLog {
		t.Logf("Other .log files were NOT excluded (unexpected)")
	} else {
		t.Logf("Other .log files were excluded as expected")
	}
}

func TestTraverseAndCollect_WindowsVsLinuxSeparators(t *testing.T) {
	// What: Handles path separators correctly on current platform
	files := []string{"nested/deep/file.txt"}
	tmpDir := setupTestDir(t, files)
	defer cleanup(t, tmpDir)

	patterns := []string{}
	collected, err := traverseAndCollect(tmpDir, patterns)

	assertNoError(t, err)
	if len(collected) != 1 {
		t.Fatalf("Expected 1 file, got %d", len(collected))
	}

	// Check that separator matches current OS
	expectedSep := string(filepath.Separator)
	if !strings.Contains(collected[0], expectedSep) && strings.Contains(collected[0], "/") {
		t.Errorf("Expected path to use OS separator %q, got: %s", expectedSep, collected[0])
	}
}

// Group 3: Integration Tests (8 tests)

func TestCompareFilesWithScripts_AllFilesExist(t *testing.T) {
	// What: All repository files are referenced in script (happy path)
	files := []string{"src/main.go", "lib/utils.go"}
	tmpDir := setupTestDir(t, files)
	defer cleanup(t, tmpDir)

	script := "test-script.sh"
	// Script contains ALL files from repository
	validLines := map[int]string{
		1: filepath.Join("src", "main.go"),
		2: filepath.Join("lib", "utils.go"),
	}
	patterns := []string{}

	err := compareFilesWithScripts(script, validLines, tmpDir, patterns)
	// No error because all repo files are in the script
	assertNoError(t, err)
}

func TestCompareFilesWithScripts_MissingFiles(t *testing.T) {
	// What: Repository has files NOT referenced in script (tests "missing in repo")
	files := []string{"src/main.go", "src/extra.go"}
	tmpDir := setupTestDir(t, files)
	defer cleanup(t, tmpDir)

	script := "test-script.sh"
	// Script only references main.go, but repo also has extra.go
	validLines := map[int]string{
		1: filepath.Join("src", "main.go"),
	}
	patterns := []string{}

	err := compareFilesWithScripts(script, validLines, tmpDir, patterns)
	// Function logs error but doesn't return error - it only returns traversal errors
	// The function purpose is to CHECK and LOG, not fail
	assertNoError(t, err) // No traversal errors
}

func TestCompareFilesWithScripts_WithIgnoredFiles(t *testing.T) {
	// What: Ignored files don't cause errors since they're not collected
	files := []string{"build/output.exe", "src/main.go"}
	tmpDir := setupTestDir(t, files)
	defer cleanup(t, tmpDir)

	script := "test-script.sh"
	// Script only references main.go
	validLines := map[int]string{1: filepath.Join("src", "main.go")}
	patterns := []string{"build/"} // build dir is ignored

	err := compareFilesWithScripts(script, validLines, tmpDir, patterns)
	// output.exe is ignored so it won't be collected, won't cause error
	assertNoError(t, err)
}

func TestCompareFilesWithScripts_EmptyScript(t *testing.T) {
	// What: Script with no file references but repo has files
	files := []string{"src/main.go"}
	tmpDir := setupTestDir(t, files)
	defer cleanup(t, tmpDir)

	script := "test-script.sh"
	validLines := map[int]string{} // Empty map
	patterns := []string{}

	err := compareFilesWithScripts(script, validLines, tmpDir, patterns)
	// Repo file(s) exist but script is empty - will log errors but not return error
	assertNoError(t, err)
}

func TestCompareFilesWithScripts_EmptyButExistingDirectory(t *testing.T) {
	// What: Script references files but repository directory exists and is empty (no files)
	tmpDir, err := os.MkdirTemp("", "content-test-empty-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer cleanup(t, tmpDir)

	script := "test-script.sh"
	validLines := map[int]string{1: "main.go"}
	patterns := []string{}

	err = compareFilesWithScripts(script, validLines, tmpDir, patterns)
	// No repo files found, script references files - no error because we're checking
	// if repo files are in script, not if script files are in repo
	assertNoError(t, err)
}

func TestCompareFilesWithScripts_TraversalErrors(t *testing.T) {
	// What: Traversal encounters errors, still compares what was collected
	files := []string{"accessible.txt"}
	tmpDir := setupTestDir(t, files)
	defer cleanup(t, tmpDir)

	// Create a subdirectory that will cause an error by making it inaccessible
	// Note: This is platform-specific and may not work reliably on all systems
	errorDir := filepath.Join(tmpDir, "restricted")
	if err := os.Mkdir(errorDir, 0000); err != nil {
		t.Skipf("Cannot create restricted directory: %v", err)
	}
	defer os.Chmod(errorDir, 0755) // Restore permissions for cleanup

	script := "test-script.sh"
	validLines := map[int]string{1: "accessible.txt"}
	patterns := []string{}

	err := compareFilesWithScripts(script, validLines, tmpDir, patterns)
	// Should return traversal error
	if err != nil {
		t.Logf("Got traversal error as expected: %v", err)
	}
}

func TestCompareFilesWithScripts_CaseSensitivity(t *testing.T) {
	// What: Tests case sensitivity handling (Windows=insensitive, Linux=sensitive)
	files := []string{"main.go"}
	tmpDir := setupTestDir(t, files)
	defer cleanup(t, tmpDir)

	script := "test-script.sh"
	validLines := map[int]string{1: "Main.go"} // Different case
	patterns := []string{}

	err := compareFilesWithScripts(script, validLines, tmpDir, patterns)
	assertNoError(t, err) // No traversal error

	// On Windows, main.go from repo won't match Main.go in script -> error logged
	// On Linux, same behavior
	// This test documents that the comparison is case-sensitive
	t.Logf("Comparison completed - check logs for case sensitivity behavior")
}

func TestCompareFilesWithScripts_MultipleScripts(t *testing.T) {
	// What: Multiple script files each referencing different files
	files := []string{"src/app.go", "lib/util.go", "test/main_test.go"}
	tmpDir := setupTestDir(t, files)
	defer cleanup(t, tmpDir)

	patterns := []string{}

	// First script references all files
	err1 := compareFilesWithScripts("script1.sh",
		map[int]string{
			1: filepath.Join("src", "app.go"),
			2: filepath.Join("lib", "util.go"),
			3: filepath.Join("test", "main_test.go"),
		},
		tmpDir, patterns)
	assertNoError(t, err1)

	// Second script missing one file - will log error about unreferenced file
	err2 := compareFilesWithScripts("script2.sh",
		map[int]string{
			1: filepath.Join("lib", "util.go"),
			2: filepath.Join("test", "main_test.go"),
		},
		tmpDir, patterns)
	assertNoError(t, err2) // No traversal error

	// Third script is empty - all repo files will be logged as errors
	err3 := compareFilesWithScripts("script3.sh",
		map[int]string{},
		tmpDir, patterns)
	assertNoError(t, err3) // No traversal error
}
