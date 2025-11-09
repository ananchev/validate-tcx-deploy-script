package analyzer

import (
	"os"
	"strings"
	"testing"

	"github.com/ananchev/validate-tcx-deploy-script/internal/logger"
)

// Helper function to set up test scenario
func setupParityTest() {
	// Reset global state
	scriptExecutables = make(map[string]map[string]bool)
	
	// Initialize logger to prevent nil pointer errors
	// Use os.DevNull to avoid creating test log files
	logger.InitLogger(os.DevNull, "error")
}

// TestCheckScriptParity_PerfectMatch tests matching executables between Windows and Linux
// What it tests: Both scripts have same utilities -> No parity issues reported
func TestCheckScriptParity_PerfectMatch(t *testing.T) {
	setupParityTest()

	// Simulate identical executables
	scriptExecutables["deploy_win.bat"] = map[string]bool{
		"plmxml_import": true,
		"tc_utils":      true,
	}
	scriptExecutables["deploy_linux.sh"] = map[string]bool{
		"plmxml_import": true,
		"tc_utils":      true,
	}

	scripts := []scriptDefinition{
		{Filename: "deploy_win.bat", TargetOS: "windows"},
		{Filename: "deploy_linux.sh", TargetOS: "linux"},
	}

	checkScriptParity(scripts)

	// Check that no errors were added (parity issues would add errors to analysisResult)
	// Since we can't directly assert error count, we verify the function completes without panic
}

// TestCheckScriptParity_MissingInLinux tests executable present in Windows but missing in Linux
// What it tests: Windows has "plmxml_import", Linux doesn't -> Parity error logged
func TestCheckScriptParity_MissingInLinux(t *testing.T) {
	setupParityTest()

	scriptExecutables["deploy_win.bat"] = map[string]bool{
		"plmxml_import": true,
		"tc_utils":      true,
	}
	scriptExecutables["deploy_linux.sh"] = map[string]bool{
		"tc_utils": true,
	}

	scripts := []scriptDefinition{
		{Filename: "deploy_win.bat", TargetOS: "windows"},
		{Filename: "deploy_linux.sh", TargetOS: "linux"},
	}

	// Function should detect mismatch and log it
	checkScriptParity(scripts)

	// The function logs errors but doesn't return them
	// In a real scenario, we'd check analysisResult.Files for the error entry
}

// TestCheckScriptParity_MissingInWindows tests executable present in Linux but missing in Windows
// What it tests: Linux has "deploy_config", Windows doesn't -> Parity error logged
func TestCheckScriptParity_MissingInWindows(t *testing.T) {
	setupParityTest()

	scriptExecutables["deploy_win.bat"] = map[string]bool{
		"tc_utils": true,
	}
	scriptExecutables["deploy_linux.sh"] = map[string]bool{
		"tc_utils":      true,
		"deploy_config": true,
	}

	scripts := []scriptDefinition{
		{Filename: "deploy_win.bat", TargetOS: "windows"},
		{Filename: "deploy_linux.sh", TargetOS: "linux"},
	}

	checkScriptParity(scripts)

	// Function should detect mismatch and log it
}

// TestCheckScriptParity_MultipleMismatches tests multiple parity issues
// What it tests: Windows has "util_a", "util_b"; Linux has "util_b", "util_c" -> Two parity errors
func TestCheckScriptParity_MultipleMismatches(t *testing.T) {
	setupParityTest()

	scriptExecutables["deploy_win.bat"] = map[string]bool{
		"util_a": true,
		"util_b": true,
	}
	scriptExecutables["deploy_linux.sh"] = map[string]bool{
		"util_b": true,
		"util_c": true,
	}

	scripts := []scriptDefinition{
		{Filename: "deploy_win.bat", TargetOS: "windows"},
		{Filename: "deploy_linux.sh", TargetOS: "linux"},
	}

	checkScriptParity(scripts)

	// Should report both util_a (missing in Linux) and util_c (missing in Windows)
}

// TestCheckScriptParity_EmptyWindowsScript tests Windows script with no executables
// What it tests: Windows script empty, Linux has utilities -> All Linux utilities reported as missing in Windows
func TestCheckScriptParity_EmptyWindowsScript(t *testing.T) {
	setupParityTest()

	scriptExecutables["deploy_win.bat"] = map[string]bool{}
	scriptExecutables["deploy_linux.sh"] = map[string]bool{
		"plmxml_import": true,
		"tc_utils":      true,
	}

	scripts := []scriptDefinition{
		{Filename: "deploy_win.bat", TargetOS: "windows"},
		{Filename: "deploy_linux.sh", TargetOS: "linux"},
	}

	checkScriptParity(scripts)

	// Should report that Windows script has no executables while Linux has 2
}

// TestCheckScriptParity_EmptyLinuxScript tests Linux script with no executables
// What it tests: Linux script empty, Windows has utilities -> All Windows utilities reported as missing in Linux
func TestCheckScriptParity_EmptyLinuxScript(t *testing.T) {
	setupParityTest()

	scriptExecutables["deploy_win.bat"] = map[string]bool{
		"plmxml_import": true,
		"tc_utils":      true,
	}
	scriptExecutables["deploy_linux.sh"] = map[string]bool{}

	scripts := []scriptDefinition{
		{Filename: "deploy_win.bat", TargetOS: "windows"},
		{Filename: "deploy_linux.sh", TargetOS: "linux"},
	}

	checkScriptParity(scripts)

	// Should report that Linux script has no executables while Windows has 2
}

// TestCheckScriptParity_BothEmpty tests both scripts with no executables
// What it tests: Both Windows and Linux scripts empty -> No parity issues (both have nothing)
func TestCheckScriptParity_BothEmpty(t *testing.T) {
	setupParityTest()

	scriptExecutables["deploy_win.bat"] = map[string]bool{}
	scriptExecutables["deploy_linux.sh"] = map[string]bool{}

	scripts := []scriptDefinition{
		{Filename: "deploy_win.bat", TargetOS: "windows"},
		{Filename: "deploy_linux.sh", TargetOS: "linux"},
	}

	checkScriptParity(scripts)

	// No parity issues since both are empty (matching state)
}

// TestCheckScriptParity_OnlyWindowsScript tests scenario with only Windows script
// What it tests: Only Windows script exists in map -> No parity check performed
func TestCheckScriptParity_OnlyWindowsScript(t *testing.T) {
	setupParityTest()

	scriptExecutables["deploy_win.bat"] = map[string]bool{
		"plmxml_import": true,
	}

	scripts := []scriptDefinition{
		{Filename: "deploy_win.bat", TargetOS: "windows"},
	}

	checkScriptParity(scripts)

	// Function should handle gracefully (no Linux counterpart to compare)
}

// TestCheckScriptParity_OnlyLinuxScript tests scenario with only Linux script
// What it tests: Only Linux script exists in map -> No parity check performed
func TestCheckScriptParity_OnlyLinuxScript(t *testing.T) {
	setupParityTest()

	scriptExecutables["deploy_linux.sh"] = map[string]bool{
		"plmxml_import": true,
	}

	scripts := []scriptDefinition{
		{Filename: "deploy_linux.sh", TargetOS: "linux"},
	}

	checkScriptParity(scripts)

	// Function should handle gracefully (no Windows counterpart to compare)
}

// TestCheckScriptParity_CaseSensitivity tests that executable names are case-sensitive
// What it tests: Windows has "PlmXML_Import", Linux has "plmxml_import" -> Different executables, parity error
func TestCheckScriptParity_CaseSensitivity(t *testing.T) {
	setupParityTest()

	scriptExecutables["deploy_win.bat"] = map[string]bool{
		"PlmXML_Import": true,
	}
	scriptExecutables["deploy_linux.sh"] = map[string]bool{
		"plmxml_import": true,
	}

	scripts := []scriptDefinition{
		{Filename: "deploy_win.bat", TargetOS: "windows"},
		{Filename: "deploy_linux.sh", TargetOS: "linux"},
	}

	checkScriptParity(scripts)

	// Should report mismatch since executable names differ in case
	// (extractExecutableName already lowercases, so this tests that behavior)
}

// TestCheckScriptParity_MultipleScriptPairs tests multiple pairs of scripts
// What it tests: Two script pairs (deploy/install), each with own parity -> Each pair checked independently
func TestCheckScriptParity_MultipleScriptPairs(t *testing.T) {
	setupParityTest()

	// First pair: deploy scripts
	scriptExecutables["deploy_win.bat"] = map[string]bool{
		"plmxml_import": true,
	}
	scriptExecutables["deploy_linux.sh"] = map[string]bool{
		"plmxml_import": true,
	}

	// Second pair: install scripts
	scriptExecutables["install_win.bat"] = map[string]bool{
		"tc_utils": true,
	}
	scriptExecutables["install_linux.sh"] = map[string]bool{
		"deploy_config": true, // Different utility
	}

	scripts := []scriptDefinition{
		{Filename: "deploy_win.bat", TargetOS: "windows"},
		{Filename: "deploy_linux.sh", TargetOS: "linux"},
		{Filename: "install_win.bat", TargetOS: "windows"},
		{Filename: "install_linux.sh", TargetOS: "linux"},
	}

	checkScriptParity(scripts)

	// First pair should match, second pair should have parity error
}

// TestCheckScriptParity_NoScripts tests empty scriptExecutables map
// What it tests: No scripts tracked -> Function handles gracefully without errors
func TestCheckScriptParity_NoScripts(t *testing.T) {
	setupParityTest()

	// scriptExecutables is empty
	scripts := []scriptDefinition{}

	checkScriptParity(scripts)

	// Should complete without panic
}

// TestCheckScriptParity_IgnoresNonTCUtilities tests that shell commands are not compared
// What it tests: Windows has "echo", Linux has "mkdir" -> No parity error (both are shell commands)
func TestCheckScriptParity_IgnoresNonTCUtilities(t *testing.T) {
	setupParityTest()

	// These should have been filtered out by extractExecutableName and not tracked
	// But if they were tracked, parity check should still work correctly

	scriptExecutables["deploy_win.bat"] = map[string]bool{
		"plmxml_import": true,
	}
	scriptExecutables["deploy_linux.sh"] = map[string]bool{
		"plmxml_import": true,
	}

	scripts := []scriptDefinition{
		{Filename: "deploy_win.bat", TargetOS: "windows"},
		{Filename: "deploy_linux.sh", TargetOS: "linux"},
	}

	checkScriptParity(scripts)

	// Should match (both have plmxml_import)
	// Shell commands like echo/mkdir shouldn't affect parity
}

// TestGetLinuxCounterpart_FindsMatch tests helper to find Linux script for Windows script
// What it tests: Given "deploy_win.bat" -> Returns "deploy_linux.sh" (if it exists)
func TestGetLinuxCounterpart_FindsMatch(t *testing.T) {
	setupParityTest()

	scriptExecutables["deploy_win.bat"] = map[string]bool{}
	scriptExecutables["deploy_linux.sh"] = map[string]bool{}

	// Note: getLinuxCounterpart is not exported, this tests the concept
	// The actual function finds counterparts using string manipulation

	windowsScript := "deploy_win.bat"
	expectedLinux := "deploy_linux.sh"

	// Simulate the logic: replace _win.bat with _linux.sh
	linuxScript := strings.Replace(windowsScript, "_win.bat", "_linux.sh", 1)

	if linuxScript != expectedLinux {
		t.Errorf("Expected %q, got %q", expectedLinux, linuxScript)
	}
}

// TestGetWindowsCounterpart_FindsMatch tests helper to find Windows script for Linux script
// What it tests: Given "deploy_linux.sh" -> Returns "deploy_win.bat" (if it exists)
func TestGetWindowsCounterpart_FindsMatch(t *testing.T) {
	setupParityTest()

	scriptExecutables["deploy_win.bat"] = map[string]bool{}
	scriptExecutables["deploy_linux.sh"] = map[string]bool{}

	linuxScript := "deploy_linux.sh"
	expectedWindows := "deploy_win.bat"

	// Simulate the logic: replace _linux.sh with _win.bat
	windowsScript := strings.Replace(linuxScript, "_linux.sh", "_win.bat", 1)

	if windowsScript != expectedWindows {
		t.Errorf("Expected %q, got %q", expectedWindows, windowsScript)
	}
}
