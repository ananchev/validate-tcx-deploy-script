package analyzer

import (
	"testing"
)

// TestExtractExecutableName_SimpleCommand tests extraction from simple command without path
// What it tests: "install_data" -> "install_data"
func TestExtractExecutableName_SimpleCommand(t *testing.T) {
	result := extractExecutableName("install_data -input=file.xml")
	expected := "install_data"

	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

// TestExtractExecutableName_WithLinuxPath tests extraction from command with Linux path
// What it tests: "$TC_BIN/install_data -input=file" -> "install_data"
func TestExtractExecutableName_WithLinuxPath(t *testing.T) {
	result := extractExecutableName("$TC_BIN/install_data -input=file")
	expected := "install_data"

	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

// TestExtractExecutableName_WithWindowsPath tests extraction from Windows path with .exe
// What it tests: "%TC_BIN%\install_data.exe -input=file" -> "install_data"
func TestExtractExecutableName_WithWindowsPath(t *testing.T) {
	result := extractExecutableName("%TC_BIN%\\install_data.exe -input=file")
	expected := "install_data"

	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

// TestExtractExecutableName_ShellCommandEcho tests that echo is ignored
// What it tests: "echo Starting deployment" -> "" (empty, ignored)
func TestExtractExecutableName_ShellCommandEcho(t *testing.T) {
	result := extractExecutableName("echo Starting deployment")
	expected := ""

	if result != expected {
		t.Errorf("Expected '%s' (shell command should be ignored), got '%s'", expected, result)
	}
}

// TestExtractExecutableName_ShellCommandCd tests that cd is ignored
// What it tests: "cd /opt/teamcenter" -> "" (empty, ignored)
func TestExtractExecutableName_ShellCommandCd(t *testing.T) {
	result := extractExecutableName("cd /opt/teamcenter")
	expected := ""

	if result != expected {
		t.Errorf("Expected '%s' (shell command should be ignored), got '%s'", expected, result)
	}
}

// TestExtractExecutableName_ShellCommandMkdir tests that mkdir is ignored
// What it tests: "mkdir -p output" -> "" (empty, ignored)
func TestExtractExecutableName_ShellCommandMkdir(t *testing.T) {
	result := extractExecutableName("mkdir -p output")
	expected := ""

	if result != expected {
		t.Errorf("Expected '%s' (shell command should be ignored), got '%s'", expected, result)
	}
}

// TestExtractExecutableName_Comment tests that comments are ignored
// What it tests: "# This is a comment" -> "" (empty, ignored)
func TestExtractExecutableName_Comment(t *testing.T) {
	result := extractExecutableName("# This is a comment")
	expected := ""

	if result != expected {
		t.Errorf("Expected '%s' (comment should be ignored), got '%s'", expected, result)
	}
}

// TestExtractExecutableName_REMComment tests that REM comments are ignored
// What it tests: "REM This is a batch comment" -> "" (empty, ignored)
func TestExtractExecutableName_REMComment(t *testing.T) {
	result := extractExecutableName("REM This is a batch comment")
	expected := ""

	if result != expected {
		t.Errorf("Expected '%s' (REM comment should be ignored), got '%s'", expected, result)
	}
}

// TestExtractExecutableName_VariableAssignment tests that variable assignments are ignored
// What it tests: "TC_ROOT=/opt/tc" -> "" (empty, ignored)
func TestExtractExecutableName_VariableAssignment(t *testing.T) {
	result := extractExecutableName("TC_ROOT=/opt/tc")
	expected := ""

	if result != expected {
		t.Errorf("Expected '%s' (variable assignment should be ignored), got '%s'", expected, result)
	}
}

// TestExtractExecutableName_WithArguments tests extraction when command has multiple arguments
// What it tests: "install_data -input=file -path=data -mode=silent" -> "install_data"
func TestExtractExecutableName_WithArguments(t *testing.T) {
	result := extractExecutableName("install_data -input=file -path=data -mode=silent")
	expected := "install_data"

	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

// TestExtractExecutableName_EmptyLine tests that empty lines return empty string
// What it tests: "" -> "" (empty)
func TestExtractExecutableName_EmptyLine(t *testing.T) {
	result := extractExecutableName("")
	expected := ""

	if result != expected {
		t.Errorf("Expected '%s' (empty line), got '%s'", expected, result)
	}
}

// TestExtractExecutableName_WhitespaceOnly tests that whitespace-only lines return empty
// What it tests: "    " -> "" (empty)
func TestExtractExecutableName_WhitespaceOnly(t *testing.T) {
	result := extractExecutableName("    ")
	expected := ""

	if result != expected {
		t.Errorf("Expected '%s' (whitespace only), got '%s'", expected, result)
	}
}

// TestExtractExecutableName_CaseInsensitive tests that extraction is case-insensitive
// What it tests: "INSTALL_DATA" and "Install_Data" both -> "install_data"
func TestExtractExecutableName_CaseInsensitive(t *testing.T) {
	result1 := extractExecutableName("INSTALL_DATA -input=file")
	result2 := extractExecutableName("Install_Data -input=file")
	expected := "install_data"

	if result1 != expected {
		t.Errorf("Expected '%s' (lowercase), got '%s'", expected, result1)
	}
	if result2 != expected {
		t.Errorf("Expected '%s' (lowercase), got '%s'", expected, result2)
	}
}

// TestExtractExecutableName_BatchCommandWithAt tests @ prefix removal
// What it tests: "@echo off" -> "" (empty, shell command)
func TestExtractExecutableName_BatchCommandWithAt(t *testing.T) {
	result := extractExecutableName("@echo off")
	expected := ""

	if result != expected {
		t.Errorf("Expected '%s' (@echo should be ignored), got '%s'", expected, result)
	}
}

// TestExtractExecutableName_DotSlashPath tests ./ prefix handling
// What it tests: "./install_data" -> "install_data"
func TestExtractExecutableName_DotSlashPath(t *testing.T) {
	result := extractExecutableName("./install_data -input=file")
	expected := "install_data"

	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

// TestTrackExecutable_AddNew tests that the first executable is tracked correctly
// What it tests: Empty registry -> Track "install_data" -> Should be in registry
func TestTrackExecutable_AddNew(t *testing.T) {
	// SETUP: Clear package state
	scriptExecutables = nil

	// EXECUTE: Track executable
	trackExecutable("script.sh", "$TC_BIN/install_data -input=file")

	// ASSERT: Verify it's in the map
	if scriptExecutables == nil {
		t.Fatal("scriptExecutables map not initialized")
	}
	if scriptExecutables["script.sh"] == nil {
		t.Fatal("script.sh not in map")
	}
	if !scriptExecutables["script.sh"]["install_data"] {
		t.Error("Expected install_data to be tracked")
	}
}

// TestTrackExecutable_SameExecutableMultipleTimes tests that duplicate calls are handled as a set
// What it tests: Call install_data 3 times in same script -> Should only have 1 entry (no duplicates)
func TestTrackExecutable_SameExecutableMultipleTimes(t *testing.T) {
	// SETUP: Clear state
	scriptExecutables = nil

	// EXECUTE: Track same executable 3 times with different arguments
	trackExecutable("script.sh", "$TC_BIN/install_data -input=file1.xml")
	trackExecutable("script.sh", "$TC_BIN/install_data -input=file2.xml")
	trackExecutable("script.sh", "./install_data -input=file3.xml")

	// ASSERT: Should still be tracked once (map[string]bool means unique set)
	if len(scriptExecutables["script.sh"]) != 1 {
		t.Errorf("Expected 1 unique executable, got %d", len(scriptExecutables["script.sh"]))
	}
	if !scriptExecutables["script.sh"]["install_data"] {
		t.Error("Expected install_data to be tracked")
	}
}

// TestTrackExecutable_DifferentExecutablesSameScript tests tracking multiple different executables
// What it tests: Track install_data, configure_plmxml, import_dataset -> All 3 should be in registry
func TestTrackExecutable_DifferentExecutablesSameScript(t *testing.T) {
	// SETUP
	scriptExecutables = nil

	// EXECUTE: Track 3 different executables in same script
	trackExecutable("script.sh", "install_data -input=file")
	trackExecutable("script.sh", "configure_plmxml -path=config")
	trackExecutable("script.sh", "import_dataset -file=data")

	// ASSERT: All 3 should be tracked
	if len(scriptExecutables["script.sh"]) != 3 {
		t.Errorf("Expected 3 executables, got %d", len(scriptExecutables["script.sh"]))
	}
	if !scriptExecutables["script.sh"]["install_data"] {
		t.Error("Expected install_data")
	}
	if !scriptExecutables["script.sh"]["configure_plmxml"] {
		t.Error("Expected configure_plmxml")
	}
	if !scriptExecutables["script.sh"]["import_dataset"] {
		t.Error("Expected import_dataset")
	}
}

// TestTrackExecutable_MultipleScriptsSeparate tests that different scripts are tracked separately
// What it tests: win.bat tracks install_data, linux.sh tracks configure_plmxml -> Each has its own list
func TestTrackExecutable_MultipleScriptsSeparate(t *testing.T) {
	// SETUP
	scriptExecutables = nil

	// EXECUTE: Track executables in 2 different scripts
	trackExecutable("win.bat", "install_data.exe -input=file")
	trackExecutable("linux.sh", "configure_plmxml -path=config")

	// ASSERT: Each script has its own set
	if len(scriptExecutables) != 2 {
		t.Errorf("Expected 2 scripts, got %d", len(scriptExecutables))
	}
	if !scriptExecutables["win.bat"]["install_data"] {
		t.Error("Expected install_data in win.bat")
	}
	if !scriptExecutables["linux.sh"]["configure_plmxml"] {
		t.Error("Expected configure_plmxml in linux.sh")
	}
	// Verify they don't cross-contaminate
	if scriptExecutables["win.bat"]["configure_plmxml"] {
		t.Error("configure_plmxml should not be in win.bat")
	}
	if scriptExecutables["linux.sh"]["install_data"] {
		t.Error("install_data should not be in linux.sh")
	}
}

// TestTrackExecutable_IgnoresShellCommands tests that shell commands are not tracked
// What it tests: echo, cd, mkdir -> Registry should be empty (all ignored)
func TestTrackExecutable_IgnoresShellCommands(t *testing.T) {
	// SETUP
	scriptExecutables = nil

	// EXECUTE: Try to track shell commands (should be ignored)
	trackExecutable("script.sh", "echo Starting deployment")
	trackExecutable("script.sh", "cd /opt/tc")
	trackExecutable("script.sh", "mkdir -p output")
	trackExecutable("script.sh", "export TC_ROOT=/opt/tc")

	// ASSERT: Nothing tracked (all shell commands)
	if scriptExecutables["script.sh"] != nil && len(scriptExecutables["script.sh"]) != 0 {
		t.Errorf("Expected 0 executables (all shell commands), got %d", len(scriptExecutables["script.sh"]))
	}
}

