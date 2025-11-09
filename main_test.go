package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetConfig_Success(t *testing.T) {
	// Create a temporary valid YAML config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.yaml")

	validYAML := `scripts:
  - filename: test.sh
    target_os: linux
path_parameters:
  - input
  - file
source_code_root: '/test/path'
ignore_patterns:
  global:
    - '*.md'
  stylesheets_folder:
    - '*.txt'
logfile: test.log
`
	err := os.WriteFile(configPath, []byte(validYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Test getConfig
	config, err := getConfig(configPath)
	if err != nil {
		t.Fatalf("getConfig() failed: %v", err)
	}

	// Verify config was parsed correctly
	if len(config.Scripts) != 1 {
		t.Errorf("Expected 1 script, got %d", len(config.Scripts))
	}
	if config.Scripts[0].Filename != "test.sh" {
		t.Errorf("Expected filename 'test.sh', got '%s'", config.Scripts[0].Filename)
	}
	if config.Scripts[0].TargetOS != "linux" {
		t.Errorf("Expected target_os 'linux', got '%s'", config.Scripts[0].TargetOS)
	}
	if config.SourceCodeRoot != "/test/path" {
		t.Errorf("Expected source_code_root '/test/path', got '%s'", config.SourceCodeRoot)
	}
	if config.Logfile != "test.log" {
		t.Errorf("Expected logfile 'test.log', got '%s'", config.Logfile)
	}
}

func TestGetConfig_FileNotFound(t *testing.T) {
	_, err := getConfig("nonexistent_file.yaml")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}

	// Check error message contains the filename
	if err != nil && !contains(err.Error(), "nonexistent_file.yaml") {
		t.Errorf("Error message should contain filename, got: %v", err)
	}
	if err != nil && !contains(err.Error(), "not found") {
		t.Errorf("Error message should indicate file not found, got: %v", err)
	}
}

func TestGetConfig_InvalidYAML(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "invalid.yaml")

	invalidYAML := `scripts:
  - filename: test.sh
    target_os: linux
  invalid yaml structure here
  no proper indentation
`
	err := os.WriteFile(configPath, []byte(invalidYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err = getConfig(configPath)
	if err == nil {
		t.Error("Expected error for invalid YAML, got nil")
	}

	if err != nil && !contains(err.Error(), "invalid YAML format") {
		t.Errorf("Error should mention invalid YAML format, got: %v", err)
	}
}

func TestGetConfig_WithTabs(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "tabs.yaml")

	// YAML with tabs (intentionally invalid)
	yamlWithTabs := "scripts:\n\t- filename: test.sh\n\t  target_os: linux\n"

	err := os.WriteFile(configPath, []byte(yamlWithTabs), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err = getConfig(configPath)
	if err == nil {
		t.Error("Expected error for YAML with tabs, got nil")
	}

	// Check that error message specifically mentions tabs
	if err != nil && !contains(err.Error(), "tabs") {
		t.Errorf("Error should mention tabs, got: %v", err)
	}
	if err != nil && !contains(err.Error(), "spaces") {
		t.Errorf("Error should mention spaces, got: %v", err)
	}
}

func TestGetConfig_EmptyFile(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "empty.yaml")

	err := os.WriteFile(configPath, []byte(""), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Empty YAML should fail validation (missing required fields)
	_, err = getConfig(configPath)
	if err == nil {
		t.Error("Expected error for empty config, got nil")
	}

	if err != nil && !contains(err.Error(), "scripts") {
		t.Errorf("Error should mention missing scripts, got: %v", err)
	}
}

func TestGetConfig_UnreadableFile(t *testing.T) {
	if os.Getenv("SKIP_PERMISSION_TESTS") != "" {
		t.Skip("Skipping permission test")
	}

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "unreadable.yaml")

	err := os.WriteFile(configPath, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Make file unreadable (this might not work on all platforms)
	err = os.Chmod(configPath, 0000)
	if err != nil {
		t.Skip("Cannot change file permissions on this platform")
	}
	defer os.Chmod(configPath, 0644) // Restore for cleanup

	_, err = getConfig(configPath)
	if err == nil {
		t.Error("Expected error for unreadable file, got nil")
	}
}

func TestProcessArgs_Defaults(t *testing.T) {
	// Save original os.Args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Test with no arguments
	os.Args = []string{"cmd"}

	args := ProcessArgs()

	if args.ConfigPath != "config.yaml" {
		t.Errorf("Expected default config path 'config.yaml', got '%s'", args.ConfigPath)
	}
	if args.LogLevel != "error" {
		t.Errorf("Expected default log level 'error', got '%s'", args.LogLevel)
	}
}

func TestProcessArgs_CustomValues(t *testing.T) {
	// Save original os.Args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Test with custom arguments
	os.Args = []string{"cmd", "-c", "custom.yaml", "-l", "debug"}

	args := ProcessArgs()

	if args.ConfigPath != "custom.yaml" {
		t.Errorf("Expected config path 'custom.yaml', got '%s'", args.ConfigPath)
	}
	if args.LogLevel != "debug" {
		t.Errorf("Expected log level 'debug', got '%s'", args.LogLevel)
	}
}

func TestRun_ConfigNotFound(t *testing.T) {
	// Save original os.Args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Test with non-existent config
	os.Args = []string{"cmd", "-c", "definitely_does_not_exist.yaml"}

	err := run()
	if err == nil {
		t.Error("Expected error when config file doesn't exist, got nil")
	}

	if err != nil && !contains(err.Error(), "not found") {
		t.Errorf("Error should indicate file not found, got: %v", err)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && stringContains(s, substr)))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestAnalyzerParametersStructure verifies the analyzer.Parameters struct can be unmarshaled
func TestAnalyzerParametersStructure(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "complete_config.yaml")

	completeYAML := `scripts:
  - filename: deploy_win.bat
    target_os: windows
  - filename: deploy_linux.sh
    target_os: linux
path_parameters:
  - input
  - xml_file
  - name
  - path
  - file
source_code_root: '/project/config'
ignore_patterns:
  global:
    - '*.md'
    - '*.txt'
    - 'test_folder'
  stylesheets_folder:
    - '*.log'
    - '*.tmp'
logfile: validation.log
`
	err := os.WriteFile(configPath, []byte(completeYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	config, err := getConfig(configPath)
	if err != nil {
		t.Fatalf("getConfig() failed: %v", err)
	}

	// Verify all fields were parsed
	if len(config.Scripts) != 2 {
		t.Errorf("Expected 2 scripts, got %d", len(config.Scripts))
	}
	if len(config.PathParameters) != 5 {
		t.Errorf("Expected 5 path parameters, got %d", len(config.PathParameters))
	}
	if len(config.IgnorePatterns.Global) != 3 {
		t.Errorf("Expected 3 global ignore patterns, got %d", len(config.IgnorePatterns.Global))
	}
	if len(config.IgnorePatterns.StyleSheetsFolder) != 2 {
		t.Errorf("Expected 2 stylesheets ignore patterns, got %d", len(config.IgnorePatterns.StyleSheetsFolder))
	}
}

func TestGetConfig_MissingScripts(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "no_scripts.yaml")

	noScriptsYAML := `path_parameters:
  - input
source_code_root: '/test/path'
ignore_patterns:
  global:
    - '*.md'
  stylesheets_folder:
    - '*.txt'
logfile: test.log
`
	err := os.WriteFile(configPath, []byte(noScriptsYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err = getConfig(configPath)
	if err == nil {
		t.Error("Expected error for missing scripts, got nil")
	}

	if err != nil && !contains(err.Error(), "scripts") {
		t.Errorf("Error should mention scripts, got: %v", err)
	}
	if err != nil && !contains(err.Error(), "empty") {
		t.Errorf("Error should mention empty, got: %v", err)
	}
}

func TestGetConfig_MissingScriptFilename(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "no_filename.yaml")

	noFilenameYAML := `scripts:
  - target_os: linux
path_parameters:
  - input
source_code_root: '/test/path'
logfile: test.log
`
	err := os.WriteFile(configPath, []byte(noFilenameYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err = getConfig(configPath)
	if err == nil {
		t.Error("Expected error for missing filename, got nil")
	}

	if err != nil && !contains(err.Error(), "filename") {
		t.Errorf("Error should mention filename, got: %v", err)
	}
}

func TestGetConfig_MissingTargetOS(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "no_target_os.yaml")

	noTargetOSYAML := `scripts:
  - filename: test.sh
path_parameters:
  - input
source_code_root: '/test/path'
logfile: test.log
`
	err := os.WriteFile(configPath, []byte(noTargetOSYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err = getConfig(configPath)
	if err == nil {
		t.Error("Expected error for missing target_os, got nil")
	}

	if err != nil && !contains(err.Error(), "target_os") {
		t.Errorf("Error should mention target_os, got: %v", err)
	}
}

func TestGetConfig_InvalidTargetOS(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "invalid_target_os.yaml")

	invalidTargetOSYAML := `scripts:
  - filename: test.sh
    target_os: macos
path_parameters:
  - input
source_code_root: '/test/path'
logfile: test.log
`
	err := os.WriteFile(configPath, []byte(invalidTargetOSYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err = getConfig(configPath)
	if err == nil {
		t.Error("Expected error for invalid target_os, got nil")
	}

	if err != nil && !contains(err.Error(), "target_os") {
		t.Errorf("Error should mention target_os, got: %v", err)
	}
	if err != nil && !contains(err.Error(), "windows") && !contains(err.Error(), "linux") {
		t.Errorf("Error should mention valid values (windows/linux), got: %v", err)
	}
}

func TestGetConfig_MissingSourceCodeRoot(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "no_source_root.yaml")

	noSourceRootYAML := `scripts:
  - filename: test.sh
    target_os: linux
path_parameters:
  - input
logfile: test.log
`
	err := os.WriteFile(configPath, []byte(noSourceRootYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err = getConfig(configPath)
	if err == nil {
		t.Error("Expected error for missing source_code_root, got nil")
	}

	if err != nil && !contains(err.Error(), "source_code_root") {
		t.Errorf("Error should mention source_code_root, got: %v", err)
	}
}

func TestGetConfig_MissingPathParameters(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "no_path_params.yaml")

	noPathParamsYAML := `scripts:
  - filename: test.sh
    target_os: linux
source_code_root: '/test/path'
logfile: test.log
`
	err := os.WriteFile(configPath, []byte(noPathParamsYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err = getConfig(configPath)
	if err == nil {
		t.Error("Expected error for missing path_parameters, got nil")
	}

	if err != nil && !contains(err.Error(), "path_parameters") {
		t.Errorf("Error should mention path_parameters, got: %v", err)
	}
}
