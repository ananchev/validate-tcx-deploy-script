package analyzer

import (
	"runtime"
	"strings"
	"testing"

	"github.com/ananchev/validate-tcx-deploy-script/internal/logger"
)

// Test setup - initialize logger for tests
func init() {
	// Initialize logger with minimal configuration for tests
	logger.InitLogger("", "error") // No file, error level only
}

// Tests for determinePathConversion()

func TestDeterminePathConversion_WindowsOnLinux(t *testing.T) {
	// What: Windows target on Linux runtime returns backslash to forward slash conversion
	// Note: This test only runs meaningfully on Linux, but should not error on Windows
	
	// Skip if not on Linux (test would pass but not be meaningful)
	if runtime.GOOS != "linux" {
		t.Skip("Skipping Windows-on-Linux test when not running on Linux")
	}

	from, to, err := determinePathConversion("windows", "test.bat")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if from != `\` {
		t.Errorf("Expected convertFrom to be '\\', got %q", from)
	}

	if to != `/` {
		t.Errorf("Expected convertTo to be '/', got %q", to)
	}
}

func TestDeterminePathConversion_LinuxOnWindows(t *testing.T) {
	// What: Linux target on Windows runtime returns forward slash to backslash conversion
	// Note: This test only runs meaningfully on Windows, but should not error on Linux
	
	// Skip if not on Windows (test would pass but not be meaningful)
	if runtime.GOOS != "windows" {
		t.Skip("Skipping Linux-on-Windows test when not running on Windows")
	}

	from, to, err := determinePathConversion("linux", "test.sh")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if from != `/` {
		t.Errorf("Expected convertFrom to be '/', got %q", from)
	}

	if to != `\` {
		t.Errorf("Expected convertTo to be '\\', got %q", to)
	}
}

func TestDeterminePathConversion_MatchingOS_Windows(t *testing.T) {
	// What: Windows target on Windows runtime returns empty (no conversion needed)
	
	// Skip if not on Windows
	if runtime.GOOS != "windows" {
		t.Skip("Skipping Windows-on-Windows test when not running on Windows")
	}

	from, to, err := determinePathConversion("windows", "test.bat")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if from != "" {
		t.Errorf("Expected empty convertFrom for matching OS, got %q", from)
	}

	if to != "" {
		t.Errorf("Expected empty convertTo for matching OS, got %q", to)
	}
}

func TestDeterminePathConversion_MatchingOS_Linux(t *testing.T) {
	// What: Linux target on Linux runtime returns empty (no conversion needed)
	
	// Skip if not on Linux
	if runtime.GOOS != "linux" {
		t.Skip("Skipping Linux-on-Linux test when not running on Linux")
	}

	from, to, err := determinePathConversion("linux", "test.sh")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if from != "" {
		t.Errorf("Expected empty convertFrom for matching OS, got %q", from)
	}

	if to != "" {
		t.Errorf("Expected empty convertTo for matching OS, got %q", to)
	}
}

func TestDeterminePathConversion_InvalidTargetOS(t *testing.T) {
	// What: Invalid target OS returns error
	
	from, to, err := determinePathConversion("macos", "test.sh")
	if err == nil {
		t.Fatal("Expected error for invalid target OS, got nil")
	}

	if from != "" {
		t.Errorf("Expected empty convertFrom on error, got %q", from)
	}

	if to != "" {
		t.Errorf("Expected empty convertTo on error, got %q", to)
	}

	// Verify error message contains useful information
	if !strings.Contains(err.Error(), "target_os") || !strings.Contains(err.Error(), "macos") {
		t.Errorf("Expected error message to mention target_os and invalid value, got: %v", err)
	}
}

// Tests for replaceInMap()

func TestReplaceInMap_ReplacesCharacters(t *testing.T) {
	// What: Replaces characters in map values
	inputMap := map[int]string{
		1: "path/to/file",
		2: "another/path/here",
		3: "no-replacement",
	}

	result := replaceInMap(inputMap, "/", `\`)

	expected := map[int]string{
		1: `path\to\file`,
		2: `another\path\here`,
		3: "no-replacement",
	}

	if len(result) != len(expected) {
		t.Fatalf("Expected %d entries, got %d", len(expected), len(result))
	}

	for key, expectedVal := range expected {
		if result[key] != expectedVal {
			t.Errorf("Key %d: expected %q, got %q", key, expectedVal, result[key])
		}
	}
}

func TestReplaceInMap_EmptyMap(t *testing.T) {
	// What: Empty map returns empty map
	inputMap := map[int]string{}

	result := replaceInMap(inputMap, "/", `\`)

	if len(result) != 0 {
		t.Errorf("Expected empty map, got %d entries", len(result))
	}
}

// Tests for replaceInIgnorePatterns()

func TestReplaceInIgnorePatterns_ReplacesBothSlices(t *testing.T) {
	// What: Replaces in both Global and StyleSheetsFolder slices
	patterns := ignorePatterns{
		Global:            []string{"dist/", "build/output", "*.log"},
		StyleSheetsFolder: []string{"temp/", "cache/files"},
	}

	result := replaceInIgnorePatterns(patterns, "/", `\`)

	expectedGlobal := []string{`dist\`, `build\output`, "*.log"}
	expectedStyleSheets := []string{`temp\`, `cache\files`}

	if len(result.Global) != len(expectedGlobal) {
		t.Fatalf("Expected %d global patterns, got %d", len(expectedGlobal), len(result.Global))
	}

	for i, expected := range expectedGlobal {
		if result.Global[i] != expected {
			t.Errorf("Global[%d]: expected %q, got %q", i, expected, result.Global[i])
		}
	}

	if len(result.StyleSheetsFolder) != len(expectedStyleSheets) {
		t.Fatalf("Expected %d stylesheet patterns, got %d", len(expectedStyleSheets), len(result.StyleSheetsFolder))
	}

	for i, expected := range expectedStyleSheets {
		if result.StyleSheetsFolder[i] != expected {
			t.Errorf("StyleSheetsFolder[%d]: expected %q, got %q", i, expected, result.StyleSheetsFolder[i])
		}
	}
}

func TestReplaceInIgnorePatterns_EmptyPatterns(t *testing.T) {
	// What: Empty patterns return empty patterns
	patterns := ignorePatterns{
		Global:            []string{},
		StyleSheetsFolder: []string{},
	}

	result := replaceInIgnorePatterns(patterns, "/", `\`)

	if len(result.Global) != 0 {
		t.Errorf("Expected empty Global slice, got %d entries", len(result.Global))
	}

	if len(result.StyleSheetsFolder) != 0 {
		t.Errorf("Expected empty StyleSheetsFolder slice, got %d entries", len(result.StyleSheetsFolder))
	}
}
