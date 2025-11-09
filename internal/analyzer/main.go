package analyzer

import (
	"runtime"

	"github.com/ananchev/validate-tcx-deploy-script/internal/logger"
)

type Lines struct {
	Valid            map[int]string
	StyleSheetImport map[int]StyleSheetImport
	Invalid          map[int]string
	Skipped          map[int]string
	Missing          []string
}

type Result struct {
	File map[string]Lines
}

type StyleSheetImport struct {
	Line         string
	InputFile    string
	XMLsFilepath string
}

var pathParameters []string
var sourceCodeRoot string
var ignores ignorePatterns
var (
	convertFrom string
	convertTo   string
)

var analysisResult Result = Result{
	File: make(map[string]Lines),
}

// processScript processes a single deployment script, performing syntax checks,
// file system validation, and directory content checks.
//
// Parameters:
//   - script: The script configuration to process
//   - params: The global parameters containing path settings and ignore patterns
//
// Returns:
//   - error: Any error encountered during processing
func processScript(script scriptDefinition, params Parameters) error {
	// create a results set for each of our filepaths
	analysisResult.File[script.Filename] = Lines{
		Valid:            make(map[int]string),
		StyleSheetImport: make(map[int]StyleSheetImport),
		Invalid:          make(map[int]string),
		Skipped:          make(map[int]string),
		Missing:          []string{},
	}

	logger.Heading(" ")
	logger.Separate("file '{filePath}'", "filePath", script.Filename)
	logger.Separate("=====================================")
	logger.Separate("SCRIPT SYNTAX CHECK")
	checkFileSyntax(script.Filename, params.SourceCodeRoot, script.TargetOS)

	logger.Separate("FILE SYSTEM REFERENCES CHECK")
	logger.Separate("Only path definitions with valid syntax are checked.")
	logger.Separate("The erroring lines found in the script syntax check are ignored.")

	runtimeOS := runtime.GOOS
	logger.Debug("Check is executed on '{os}' filesystem", "os", runtimeOS)

	// Determine path conversion requirements
	var err error
	convertFrom, convertTo, err = determinePathConversion(script.TargetOS, script.Filename)
	if err != nil {
		logger.Error(err.Error())
		return err
	}

	if convertFrom != "" {
		logger.Debug("Target OS for '{f}' is '{os}', but runtime OS is '{ros}'", "f", script.Filename, "os", script.TargetOS, "ros", runtimeOS)
		logger.Debug("Path separators will be converted from '{from}' to '{to}'", "from", convertFrom, "to", convertTo)
	} else {
		logger.Debug("Target OS for '{f}' is matching with the runtime OS", "f", script.Filename)
	}

	// process ignore patterns defined in the configuration to reflect the OS and script
	ignores = replaceInIgnorePatterns(params.IgnorePatterns, convertFrom, convertTo)

	checkFilePathsInScript(script.Filename, analysisResult.File[script.Filename].Valid)
	checkStylesheetPaths(script.Filename, analysisResult.File[script.Filename].StyleSheetImport)

	logger.Separate("DIRECTORY CONTENT CHECK")
	logger.Separate("File & directory patterns defined as 'ignore_patterns' in the configuration are ignored")

	if runtimeOS == "linux" {
		logger.Debug("We are running on '{ros}', replacing all '\\' in ignore_patterns with '/'", "ros", runtimeOS)
	} else if runtimeOS == "windows" {
		logger.Debug("We are running on '{ros}', replacing all '/' in ignore_patterns with '\\'", "ros", runtimeOS)
	}
	validLines := replaceInMap(analysisResult.File[script.Filename].Valid, convertFrom, convertTo)

	if err := compareFilesWithScripts(script.Filename, validLines, params.SourceCodeRoot, ignores.Global); err != nil {
		logger.Error("Errors occurred during file comparison for '{script}': {e}", "script", script.Filename, "e", err.Error())
		// Continue processing despite errors
	}
	logger.Separate(" ")

	return nil
}

func Run(params Parameters) {

	// initialize the package level variables
	pathParameters = params.PathParameters
	sourceCodeRoot = params.SourceCodeRoot

	// Initialize regex patterns once for performance
	initializeRegexPatterns(params.PathParameters)

	for _, script := range params.Scripts {
		if err := processScript(script, params); err != nil {
			// Error already logged in processScript, continue with other scripts
			continue
		}
	}

	// Check script parity (same executables in Windows and Linux scripts)
	checkScriptParity(params.Scripts)
}
