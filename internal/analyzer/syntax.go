package analyzer

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/ananchev/validate-tcx-deploy-script/internal/logger"
)

// Package-level regex patterns (compiled once for performance)
var (
	parameterFlagPatterns  map[string]*regexp.Regexp // flagName -> regex for `-flagname`
	parameterValuePatterns map[string]*regexp.Regexp // flagName -> regex for `-flagname="value"`
	stylesheetUtilityRegex *regexp.Regexp
	stylesheetFlagsRegex   *regexp.Regexp
)

// Track executables per script for parity checking
var scriptExecutables map[string]map[string]bool // scriptFile -> set of unique executables

// Common shell commands to ignore when tracking executables
var shellCommands = map[string]bool{
	"echo": true, "cd": true, "mkdir": true, "rm": true, "cp": true, "mv": true,
	"chmod": true, "chown": true, "export": true, "set": true, "pwd": true,
	"ls": true, "dir": true, "del": true, "copy": true, "move": true,
	"rem": true, "@echo": true, "call": true, "if": true, "for": true, "goto": true,
	"pushd": true, "popd": true, "exit": true, "return": true,
}

// Store current script's target OS for validation
var currentScriptTargetOS string

// initializeRegexPatterns compiles all regex patterns once for efficiency
func initializeRegexPatterns(parameters []string) {
	parameterFlagPatterns = make(map[string]*regexp.Regexp)
	parameterValuePatterns = make(map[string]*regexp.Regexp)

	for _, flagName := range parameters {
		// Compile pattern for checking if flag exists: -flagname
		flagPattern := fmt.Sprintf(`-%s`, regexp.QuoteMeta(flagName))
		parameterFlagPatterns[flagName] = regexp.MustCompile(flagPattern)

		// Compile pattern for extracting value: -flagname="value"
		valuePattern := fmt.Sprintf(`-%s="([^"]+)"`, regexp.QuoteMeta(flagName))
		parameterValuePatterns[flagName] = regexp.MustCompile(valuePattern)
	}

	// Compile stylesheet-specific patterns
	stylesheetUtilityRegex = regexp.MustCompile(`install_xml_stylesheet_datasets`)
	stylesheetFlagsRegex = regexp.MustCompile(`-input="([^"]+)"|-filepath="([^"]+)"`)
}

func checkFileSyntax(filePath string, sourceCodeRoot string, targetOS string) {

	// Set current script's target OS for validation
	currentScriptTargetOS = targetOS

	fullPath := filepath.Join(sourceCodeRoot, filePath)

	file, err := os.Open(fullPath)
	if err != nil {
		logger.Error("Error opening '{f}'. {e}.", "f", filePath, "e", err.Error())
		return
	}
	defer file.Close()

	// Read lines from the file
	scanner := bufio.NewScanner(file)
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		parseLineAsCommand(filePath, line, lineNumber)
	}

	logger.Info("valid lines")
	logValidationResults("valid", filePath)
	logger.Info("stylesheet import")
	logValidationResults("stylesheet import", filePath)
	logger.Separate("lines with invalid syntax of referenced filepaths")
	hasInvalidLines := logValidationResults("invalid", filePath)
	if !hasInvalidLines {
		logger.Separate("none")
	}
	logger.Info("skipped lines")
	logValidationResults("skipped", filePath)
}

func logValidationResults(lineType string, filePath string) bool {
	var lines map[int]string
	switch lineType {
	case "valid":
		lines = analysisResult.File[filePath].Valid
	case "invalid":
		lines = analysisResult.File[filePath].Invalid
	case "skipped":
		lines = analysisResult.File[filePath].Skipped
	case "stylesheet import":
		lines = analysisResult.File[filePath].GetStyleSheetImportLines()
	}
	if len(lines) == 0 {
		logger.Info("No {lt} entries found", "lt", lineType)
		return false
	}

	// sort by line number and print
	si := make([]int, 0, len(lines))
	for i := range lines {
		si = append(si, i)
	}
	sort.Ints(si)
	for _, i := range si {

		if lineType == "invalid" {
			logger.Error("'{f}' line '{ln}' is invalid: '{val}'", "f", filePath, "ln", i, "val", lines[i])
		} else {
			logger.Info("'{f}' line '{ln}' is valid: '{val}'", "f", filePath, "ln", i, "val", lines[i])
		}
	}
	return true
}

func parseLineAsCommand(file string, line string, lineNumber int) {

	logger.Debug("parsing line '{ln} {l}'", "ln", lineNumber, "l", line)

	// Track executables for parity check
	trackExecutable(file, line)

	var skipLine bool = true

	for _, flagName := range pathParameters {

		logger.Debug("searching for flag '{f}'", "f", flagName)

		// Use pre-compiled regex (no compilation in loop!)
		re := parameterFlagPatterns[flagName]
		matches := re.FindStringSubmatch(line)
		logger.Debug("'{lm}' matches found for flag '{f}'", "lm", len(matches), "f", flagName)

		// Check if the line contains our flags of interest
		if len(matches) < 1 {
			// Continue and try with the next parameter
			logger.Debug("no '{p}' flag found ...", "p", flagName)
			continue
		}

		logger.Debug("checking if the '-{f}' flag definition is properly formatted", "f", flagName)

		// Use pre-compiled regex for value extraction
		re = parameterValuePatterns[flagName]
		matches = re.FindStringSubmatch(line)

		logger.Debug("'{lm}' matches found", "lm", len(matches))

		// Check if the flag found is properly formatted
		if len(matches) < 2 {
			logger.Debug("line '{l}': '-{s}' is present but not quoted properly", "l", lineNumber, "s", flagName)
			analysisResult.File[file].Invalid[lineNumber] = line
			skipLine = false // do not capture this line as skip line
			break
		} else if len(matches) == 2 {
			// Extract the file path
			logger.Debug("Formatting correct, extracting the file path in '-{f}'...", "f", flagName)
			filePath := matches[1]
			logger.Debug("filepath is: '{fp}'", "fp", filePath)

			// Validate path separators match target OS
			if err := validatePathSeparators(filePath, currentScriptTargetOS, lineNumber); err != nil {
				logger.Error("'{f}' {e}", "f", file, "e", err.Error())
				analysisResult.File[file].Invalid[lineNumber] = line + " [" + err.Error() + "]"
				skipLine = false
				break
			}

			analysisResult.File[file].Valid[lineNumber] = filePath

			skipLine = false // do not capture this line as skip line

			logger.Debug("is the line defining a call to 'install_xml_stylesheet_datasets' utility?")
			var (
				inputFile           string
				stylesheetsFilepath string
			)
			if isStylesheetImportLine(line, &inputFile, &stylesheetsFilepath) {
				analysisResult.File[file].StyleSheetImport[lineNumber] = StyleSheetImport{
					Line:         line,
					XMLsFilepath: stylesheetsFilepath,
					InputFile:    inputFile,
				}
			}
			break
		}
	}

	if skipLine {
		logger.Debug("line '{ln} {l}' does not contain any flag of interest", "ln", lineNumber, "l", line)
		analysisResult.File[file].Skipped[lineNumber] = line
	}

}

func isStylesheetImportLine(line string, input *string, filepath *string) bool {
	// Use pre-compiled regex
	if !stylesheetUtilityRegex.MatchString(line) {
		logger.Debug("'{l}' does not contain 'install_xml_stylesheet_datasets'", "l", line)
		return false
	}
	logger.Debug("'{l}' is refering to 'install_xml_stylesheet_datasets'", "l", line)
	logger.Debug("Extracting the values for input and filepath flags...")

	// Use pre-compiled regex
	matches := stylesheetFlagsRegex.FindAllStringSubmatch(line, -1)

	logger.Debug("Match is '{m}'", "m", matches)
	for _, match := range matches {
		if match[1] != "" {
			logger.Debug("Value of -input flag is is '{v}'", "v", match[1])
			*input = match[1]
		}
		if match[2] != "" {
			logger.Debug("Value of -filepath flag is is '{v}'", "v", match[2])
			*filepath = match[2]
		}
	}
	return true
}

// Method to get a map of line numbers to Line strings from StyleSheetImport struct
func (l Lines) GetStyleSheetImportLines() map[int]string {
	importLines := make(map[int]string)
	for lineNumber, styleSheetImport := range l.StyleSheetImport {
		importLines[lineNumber] = styleSheetImport.Line
	}
	return importLines
}

// validatePathSeparators checks that file paths use the correct separator for the target OS
func validatePathSeparators(filePath string, targetOS string, lineNumber int) error {
	hasBackslash := strings.Contains(filePath, `\`)
	hasForwardSlash := strings.Contains(filePath, `/`)

	if targetOS == "windows" {
		if hasForwardSlash {
			return fmt.Errorf("line %d: path '%s' contains forward slashes (/) but script targets Windows (use \\)",
				lineNumber, filePath)
		}
	} else if targetOS == "linux" {
		if hasBackslash {
			return fmt.Errorf("line %d: path '%s' contains backslashes (\\) but script targets Linux (use /)",
				lineNumber, filePath)
		}
	}

	return nil
}

// extractExecutableName extracts the executable name from a command line
func extractExecutableName(line string) string {
	line = strings.TrimSpace(line)

	// Skip empty lines and comments
	if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "REM") || strings.HasPrefix(line, "rem") {
		return ""
	}

	// Get first word (the command)
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return ""
	}

	cmd := parts[0]

	// Skip variable assignments (VAR=value)
	if strings.Contains(cmd, "=") {
		return ""
	}

	// Remove @ prefix from batch commands
	cmd = strings.TrimPrefix(cmd, "@")

	// Extract executable name from path (both / and \ separators)
	// Example: $TC_BIN/install_util -> install_util
	if idx := strings.LastIndexAny(cmd, "/\\"); idx != -1 {
		cmd = cmd[idx+1:]
	}

	// Remove variable references like $VAR or %VAR%
	if strings.HasPrefix(cmd, "$") || strings.HasPrefix(cmd, "%") {
		return ""
	}

	// Convert to lowercase for case-insensitive comparison
	cmd = strings.ToLower(cmd)

	// Skip common shell commands
	if shellCommands[cmd] {
		return ""
	}

	// Remove .exe extension for consistency
	cmd = strings.TrimSuffix(cmd, ".exe")
	cmd = strings.TrimSuffix(cmd, ".bat")
	cmd = strings.TrimSuffix(cmd, ".sh")

	return cmd
}

// trackExecutable records executable calls for parity checking
func trackExecutable(scriptFile string, line string) {
	executable := extractExecutableName(line)
	if executable == "" {
		return
	}

	if scriptExecutables == nil {
		scriptExecutables = make(map[string]map[string]bool)
	}
	if scriptExecutables[scriptFile] == nil {
		scriptExecutables[scriptFile] = make(map[string]bool)
	}

	scriptExecutables[scriptFile][executable] = true
}

// checkScriptParity verifies that Windows and Linux scripts call the same executables
func checkScriptParity(scripts []scriptDefinition) {
	if len(scripts) < 2 {
		return // Need at least 2 scripts to compare
	}

	logger.Heading(" ")
	logger.Separate("SCRIPT PARITY CHECK")
	logger.Separate("=====================================")
	logger.Separate("Checking that Windows and Linux scripts call the same executables...")

	// Group scripts by target OS
	windowsScripts := []string{}
	linuxScripts := []string{}

	for _, script := range scripts {
		if script.TargetOS == "windows" {
			windowsScripts = append(windowsScripts, script.Filename)
		} else if script.TargetOS == "linux" {
			linuxScripts = append(linuxScripts, script.Filename)
		}
	}

	// If we have both Windows and Linux scripts, compare them
	if len(windowsScripts) > 0 && len(linuxScripts) > 0 {
		// Collect all executables from Windows scripts
		windowsExecs := make(map[string]bool)
		for _, ws := range windowsScripts {
			for exec := range scriptExecutables[ws] {
				windowsExecs[exec] = true
			}
		}

		// Collect all executables from Linux scripts
		linuxExecs := make(map[string]bool)
		for _, ls := range linuxScripts {
			for exec := range scriptExecutables[ls] {
				linuxExecs[exec] = true
			}
		}

		// Find executables in Windows but not in Linux
		missingInLinux := []string{}
		for exec := range windowsExecs {
			if !linuxExecs[exec] {
				missingInLinux = append(missingInLinux, exec)
			}
		}

		// Find executables in Linux but not in Windows
		missingInWindows := []string{}
		for exec := range linuxExecs {
			if !windowsExecs[exec] {
				missingInWindows = append(missingInWindows, exec)
			}
		}

		// Report findings
		if len(missingInLinux) > 0 {
			sort.Strings(missingInLinux)
			logger.Error("Executables in Windows script(s) but missing in Linux script(s): {execs}",
				"execs", strings.Join(missingInLinux, ", "))
		}

		if len(missingInWindows) > 0 {
			sort.Strings(missingInWindows)
			logger.Error("Executables in Linux script(s) but missing in Windows script(s): {execs}",
				"execs", strings.Join(missingInWindows, ", "))
		}

		if len(missingInLinux) == 0 && len(missingInWindows) == 0 {
			logger.Separate("none")
		}
	}
}
