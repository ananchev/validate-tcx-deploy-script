package analyzer

import (
	"os"
	"testing"
	
	"github.com/ananchev/validate-tcx-deploy-script/internal/logger"
)

// Helper to reset state before each test
func setupSyntaxTest() {
	scriptExecutables = make(map[string]map[string]bool)
	analysisResult.File = make(map[string]Lines)
	currentScriptTargetOS = "windows"
	
	// Initialize logger
	logger.InitLogger(os.DevNull, "error")
	
	// Initialize with common test parameters
	testParams := []string{"R", "i", "source", "target"}
	pathParameters = testParams  // Set package-level variable used by parseLineAsCommand
	initializeRegexPatterns(testParams)
}// Test initialization happens before each test
func initTestFile(filename string, targetOS string) {
	analysisResult.File[filename] = Lines{
		Valid:            make(map[int]string),
		StyleSheetImport: make(map[int]StyleSheetImport),
		Invalid:          make(map[int]string),
		Skipped:          make(map[int]string),
		Missing:          []string{},
	}
	currentScriptTargetOS = targetOS
}

// TestParseLineAsCommand_ValidRFlag tests parsing line with -R flag
// What it tests: Line "-R config\data.xml" -> Extracts "config\data.xml" as file path
func TestParseLineAsCommand_ValidRFlag(t *testing.T) {
	setupSyntaxTest()
	filename := "test_script.bat"
	initTestFile(filename, "windows")

	line := "-R=\"config\\data.xml\""
	lineNum := 10

	parseLineAsCommand(filename, line, lineNum)

	// Check that path was added to valid
	if _, exists := analysisResult.File[filename].Valid[lineNum]; !exists {
		t.Error("Expected line to be in valid paths")
	}

	if analysisResult.File[filename].Valid[lineNum] != "config\\data.xml" {
		t.Errorf("Expected path 'config\\data.xml', got '%s'", analysisResult.File[filename].Valid[lineNum])
	}
}

// TestParseLineAsCommand_ValidIFlag tests parsing line with -i flag
// What it tests: Line "-i data/import.xml" -> Extracts "data/import.xml" as file path
func TestParseLineAsCommand_ValidIFlag(t *testing.T) {
	setupSyntaxTest()
	filename := "test_script.sh"
	initTestFile(filename, "linux")

	line := "-i=\"data/import.xml\""
	lineNum := 15

	parseLineAsCommand(filename, line, lineNum)

	// Check that path was added to valid
	if _, exists := analysisResult.File[filename].Valid[lineNum]; !exists {
		t.Error("Expected line to be in valid paths")
	}

	if analysisResult.File[filename].Valid[lineNum] != "data/import.xml" {
		t.Errorf("Expected path 'data/import.xml', got '%s'", analysisResult.File[filename].Valid[lineNum])
	}
}

// TestParseLineAsCommand_WrongSeparatorWindows tests path separator validation on Windows
// What it tests: Windows script with "config/data.xml" -> Adds error for forward slash
func TestParseLineAsCommand_WrongSeparatorWindows(t *testing.T) {
	setupSyntaxTest()
	filename := "test_script.bat"
	initTestFile(filename, "windows")

	line := "-R=\"config/data.xml\""
	lineNum := 25

	parseLineAsCommand(filename, line, lineNum)

	// Check that error was added for wrong separator
	if _, exists := analysisResult.File[filename].Invalid[lineNum]; !exists {
		t.Error("Expected line to be in invalid paths due to wrong separator")
	}
}

// TestParseLineAsCommand_WrongSeparatorLinux tests path separator validation on Linux
// What it tests: Linux script with "config\data.xml" -> Adds error for backslash
func TestParseLineAsCommand_WrongSeparatorLinux(t *testing.T) {
	setupSyntaxTest()
	filename := "test_script.sh"
	initTestFile(filename, "linux")

	line := "-R=\"config\\data.xml\""
	lineNum := 30

	parseLineAsCommand(filename, line, lineNum)

	// Check that error was added for wrong separator
	if _, exists := analysisResult.File[filename].Invalid[lineNum]; !exists {
		t.Error("Expected line to be in invalid paths due to wrong separator")
	}
}

// TestParseLineAsCommand_NoFlags tests line without any recognized flags
// What it tests: Line "echo Hello World" -> Line added to skipped
func TestParseLineAsCommand_NoFlags(t *testing.T) {
	setupSyntaxTest()
	filename := "test_script.bat"
	initTestFile(filename, "windows")

	line := "echo Hello World"
	lineNum := 35

	parseLineAsCommand(filename, line, lineNum)

	// Check that line was skipped
	if _, exists := analysisResult.File[filename].Skipped[lineNum]; !exists {
		t.Error("Expected line to be in skipped")
	}
}

// TestParseLineAsCommand_TracksExecutable tests that TC utilities are tracked
// What it tests: Line "$TC_BIN/plmxml_import -R file.xml" -> Tracks "plmxml_import" in registry
func TestParseLineAsCommand_TracksExecutable(t *testing.T) {
	setupSyntaxTest()
	filename := "deploy_linux.sh"
	initTestFile(filename, "linux")

	scriptExecutables = make(map[string]map[string]bool)

	line := "$TC_BIN/plmxml_import -R=\"config/file.xml\""
	lineNum := 65

	parseLineAsCommand(filename, line, lineNum)

	// Verify executable was tracked
	executables, exists := scriptExecutables[filename]
	if !exists {
		t.Fatal("Expected script to be in registry")
	}

	if !executables["plmxml_import"] {
		t.Error("Expected plmxml_import to be tracked")
	}
}

// TestParseLineAsCommand_IgnoresShellCommands tests that shell commands are not tracked
// What it tests: Line "echo Starting deployment" -> Does not track "echo" in registry
func TestParseLineAsCommand_IgnoresShellCommands(t *testing.T) {
	setupSyntaxTest()
	filename := "deploy_linux.sh"
	initTestFile(filename, "linux")

	scriptExecutables = make(map[string]map[string]bool)

	line := "echo Starting deployment"
	lineNum := 70

	parseLineAsCommand(filename, line, lineNum)

	// Verify shell command was NOT tracked
	executables, exists := scriptExecutables[filename]
	if exists && len(executables) > 0 {
		if executables["echo"] {
			t.Error("Shell command 'echo' should not be tracked")
		}
	}
}

// TestParseLineAsCommand_VariableSubstitution tests lines with environment variables
// What it tests: Line "%TC_BIN%\plmxml_import -R file.xml" -> Extracts "file.xml", tracks "plmxml_import"
func TestParseLineAsCommand_VariableSubstitution(t *testing.T) {
	setupSyntaxTest()
	filename := "deploy_win.bat"
	initTestFile(filename, "windows")

	scriptExecutables = make(map[string]map[string]bool)

	line := "%TC_BIN%\\plmxml_import -R=\"config\\file.xml\""
	lineNum := 85

	parseLineAsCommand(filename, line, lineNum)

	// Should extract path
	if _, exists := analysisResult.File[filename].Valid[lineNum]; !exists {
		t.Error("Expected line to be in valid paths")
	}

	// Should track executable
	executables, exists := scriptExecutables[filename]
	if !exists {
		t.Fatal("Expected script to be in registry")
	}

	if !executables["plmxml_import"] {
		t.Error("Expected plmxml_import to be tracked")
	}
}

// TestParseLineAsCommand_MultipleFlags tests line with multiple flags
// What it tests: Line "-R file1.xml -i file2.xml" -> Extracts first valid flag's path
func TestParseLineAsCommand_MultipleFlags(t *testing.T) {
	setupSyntaxTest()
	filename := "test_script.sh"
	initTestFile(filename, "linux")

	line := "-R=\"config/file1.xml\" -i=\"data/file2.xml\""
	lineNum := 45

	parseLineAsCommand(filename, line, lineNum)

	// The function processes flags in order and breaks after first valid one
	if _, exists := analysisResult.File[filename].Valid[lineNum]; !exists {
		t.Error("Expected line to be in valid paths")
	}
}
