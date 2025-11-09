package analyzer

import (
	"testing"
)

// TestInitializeRegexPatterns_Success tests successful regex compilation
// What it tests: Calling initializeRegexPatterns() -> All patterns compile without error
func TestInitializeRegexPatterns_Success(t *testing.T) {
	testParams := []string{"R", "i", "source", "target"}
	initializeRegexPatterns(testParams)

	// If we get here without panic, compilation succeeded
}

// TestInitializeRegexPatterns_PatternsNotNil tests that all patterns are initialized
// What it tests: After initializeRegexPatterns() -> All regex pattern pointers are non-nil
func TestInitializeRegexPatterns_PatternsNotNil(t *testing.T) {
	testParams := []string{"R", "i", "source", "target"}
	initializeRegexPatterns(testParams)

	// Check parameter-specific patterns are not nil
	if parameterFlagPatterns == nil {
		t.Error("parameterFlagPatterns should not be nil after initialization")
	}
	if parameterValuePatterns == nil {
		t.Error("parameterValuePatterns should not be nil after initialization")
	}
	if stylesheetUtilityRegex == nil {
		t.Error("stylesheetUtilityRegex should not be nil after initialization")
	}
	if stylesheetFlagsRegex == nil {
		t.Error("stylesheetFlagsRegex should not be nil after initialization")
	}

	// Check that specific parameter patterns were created
	for _, param := range testParams {
		if parameterFlagPatterns[param] == nil {
			t.Errorf("parameterFlagPatterns[%s] should not be nil", param)
		}
		if parameterValuePatterns[param] == nil {
			t.Errorf("parameterValuePatterns[%s] should not be nil", param)
		}
	}
}

// TestInitializeRegexPatterns_EmptyParams tests initialization with no parameters
// What it tests: Call with empty parameter list -> Creates only stylesheet patterns
func TestInitializeRegexPatterns_EmptyParams(t *testing.T) {
	testParams := []string{}
	initializeRegexPatterns(testParams)

	// Stylesheet patterns should still be created
	if stylesheetUtilityRegex == nil {
		t.Error("stylesheetUtilityRegex should not be nil even with empty params")
	}
	if stylesheetFlagsRegex == nil {
		t.Error("stylesheetFlagsRegex should not be nil even with empty params")
	}

	// Parameter maps should be empty but not nil
	if parameterFlagPatterns == nil {
		t.Error("parameterFlagPatterns should be initialized (empty map)")
	}
	if len(parameterFlagPatterns) != 0 {
		t.Error("parameterFlagPatterns should be empty")
	}
}

// TestParameterFlagPattern_Matching tests that flag patterns match correctly
// What it tests: Pattern for "-R" matches "-R" in line, but not "R" alone
func TestParameterFlagPattern_Matching(t *testing.T) {
	testParams := []string{"R", "i"}
	initializeRegexPatterns(testParams)

	// Test -R pattern
	rPattern := parameterFlagPatterns["R"]
	if rPattern == nil {
		t.Fatal("R pattern not initialized")
	}

	if !rPattern.MatchString("-R") {
		t.Error("Pattern should match '-R'")
	}
	if !rPattern.MatchString("-R=\"file.xml\"") {
		t.Error("Pattern should match '-R=\"file.xml\"'")
	}
	if rPattern.MatchString("R=\"file.xml\"") {
		t.Error("Pattern should NOT match 'R=\"file.xml\"' (missing dash)")
	}
}

// TestParameterValuePattern_Extraction tests that value patterns extract correctly
// What it tests: Pattern extracts "file.xml" from "-R=\"file.xml\""
func TestParameterValuePattern_Extraction(t *testing.T) {
	testParams := []string{"R"}
	initializeRegexPatterns(testParams)

	rValuePattern := parameterValuePatterns["R"]
	if rValuePattern == nil {
		t.Fatal("R value pattern not initialized")
	}

	input := "-R=\"config/file.xml\""
	matches := rValuePattern.FindStringSubmatch(input)

	if len(matches) < 2 {
		t.Fatalf("Expected at least 2 matches, got %d", len(matches))
	}

	expected := "config/file.xml"
	if matches[1] != expected {
		t.Errorf("Expected path %q, got %q", expected, matches[1])
	}
}

// TestStylesheetUtilityRegex_Matching tests stylesheet utility pattern
// What it tests: Pattern matches "install_xml_stylesheet_datasets" utility calls
func TestStylesheetUtilityRegex_Matching(t *testing.T) {
	testParams := []string{}
	initializeRegexPatterns(testParams)

	testCases := []struct {
		input    string
		expected bool
	}{
		{"install_xml_stylesheet_datasets -input=\"file.xml\"", true},
		{"./install_xml_stylesheet_datasets", true},
		{"echo install_xml_stylesheet_datasets", true}, // Contains the string
		{"some_other_utility", false},
		{"", false},
	}

	for _, tc := range testCases {
		matches := stylesheetUtilityRegex.MatchString(tc.input)
		if matches != tc.expected {
			t.Errorf("stylesheetUtilityRegex.MatchString(%q) = %v, expected %v",
				tc.input, matches, tc.expected)
		}
	}
}

// TestStylesheetFlagsRegex_Extraction tests stylesheet flags pattern
// What it tests: Pattern extracts values from -input="..." and -filepath="..."
func TestStylesheetFlagsRegex_Extraction(t *testing.T) {
	testParams := []string{}
	initializeRegexPatterns(testParams)

	input := "install_xml_stylesheet_datasets -input=\"data.xml\" -filepath=\"styles.xsl\""
	matches := stylesheetFlagsRegex.FindAllStringSubmatch(input, -1)

	if len(matches) != 2 {
		t.Fatalf("Expected 2 flag matches, got %d", len(matches))
	}

	// First match should be -input
	if matches[0][1] != "data.xml" {
		t.Errorf("Expected input value 'data.xml', got %q", matches[0][1])
	}

	// Second match should be -filepath
	if matches[1][2] != "styles.xsl" {
		t.Errorf("Expected filepath value 'styles.xsl', got %q", matches[1][2])
	}
}

// TestMultipleParameters_AllCompiled tests that all provided parameters are compiled
// What it tests: Pass 4 parameters -> All 4 have flag and value patterns
func TestMultipleParameters_AllCompiled(t *testing.T) {
	testParams := []string{"R", "i", "source", "target", "custom"}
	initializeRegexPatterns(testParams)

	for _, param := range testParams {
		if parameterFlagPatterns[param] == nil {
			t.Errorf("Flag pattern for %q not initialized", param)
		}
		if parameterValuePatterns[param] == nil {
			t.Errorf("Value pattern for %q not initialized", param)
		}
	}

	if len(parameterFlagPatterns) != len(testParams) {
		t.Errorf("Expected %d flag patterns, got %d", len(testParams), len(parameterFlagPatterns))
	}
	if len(parameterValuePatterns) != len(testParams) {
		t.Errorf("Expected %d value patterns, got %d", len(testParams), len(parameterValuePatterns))
	}
}
