package analyzer

import (
	"runtime"
	"strings"

	"github.com/ananchev/validate-tcx-deploy-script/internal/logger"
)

type Lines struct {
	Valid   map[int]string
	Invalid map[int]string
	Skipped map[int]string
	Missing []string
}

type Result struct {
	File map[string]Lines
}

var _pathParameters []string
var (
	convertFrom string
	convertTo   string
)

var analysisResult Result = Result{
	File: make(map[string]Lines),
}

func Run(params Parameters) {

	// this slice is used in syntax.go
	_pathParameters = params.PathParameters

	for _, script := range params.Scripts {

		// create a results set for each of our filepaths
		analysisResult.File[script.Filename] = Lines{
			Valid:   make(map[int]string),
			Invalid: make(map[int]string),
			Skipped: make(map[int]string),
			Missing: []string{},
		}

		logger.Heading(" ")
		logger.Separate("file '{filePath}'", "filePath", script.Filename)
		logger.Separate("=====================================")
		logger.Separate("SCRIPTS SYNTAX CHECK")
		checkFileSyntax(script.Filename, params.SourceCodeRoot)

		logger.Separate("FILE SYSTEM REFERENCES CHECK")
		logger.Separate("Only path definitions with valid syntax are checked.")
		logger.Separate("The erroring lines found in the script syntax check are ignored.")
		logger.Separate("-------")

		runtimeOS := runtime.GOOS
		logger.Debug("Check is executed on '{os}' filesystem", "os", runtimeOS)

		if script.TargetOS == "windows" && runtimeOS == "linux" {
			logger.Debug("Target OS for '{f}' is '{os}', but runtime OS is '{ros}'", "f", script.Filename, "os", script.TargetOS, "ros", runtimeOS)
			logger.Debug("Windows paths will be converted to Linux")
			convertFrom = `\`
			convertTo = `/`
		} else if script.TargetOS == "linux" && runtimeOS == "windows" {
			logger.Debug("Target OS for '{f}' is '{os}', but runtime OS is '{ros}'", "f", script.Filename, "os", script.TargetOS, "ros", runtimeOS)
			logger.Debug("Linux paths will be converted to Windows")
			convertFrom = `/`
			convertTo = `\`
		} else if script.TargetOS != "linux" && script.TargetOS != "windows" {
			logger.Error("Incorrect specification of script target_os in the configuration yaml. Must be 'linux' or 'windows'.")
			break
		} else {
			logger.Debug("Target OS for '{f}' is matching with the runtime OS", "f", script.Filename)
		}
		checkFilePaths(script.Filename, params.SourceCodeRoot, analysisResult.File[script.Filename].Valid)
		logger.Separate("-------")

		logger.Separate("DIRECTORY CONTENT CHECK")
		logger.Separate("File & directory patterns defined as 'ignore_patterns' in the configuration are ignored")
		logger.Separate("-------")

		if runtimeOS == "linux" {
			logger.Debug("We are running on '{ros}', replacing all '\\' in ignore_patterns with '/'", "ros", runtimeOS)
		} else if runtimeOS == "windows" {
			logger.Debug("We are running on '{ros}', replacing all '/' in ignore_patterns with '\\'", "ros", runtimeOS)
		}
		validLines := replaceInMap(analysisResult.File[script.Filename].Valid, convertFrom, convertTo)
		ignorePatterns := replaceInSlice(params.IgnorePatterns, convertFrom, convertTo)
		compareFilesWithScripts(script.Filename, validLines, params.SourceCodeRoot, ignorePatterns)
		logger.Separate("-------")
		logger.Separate(" ")
	}
}

func replaceInSlice(slice []string, oldChar, newChar string) []string {
	var result []string
	for _, item := range slice {
		updatedItem := strings.ReplaceAll(item, oldChar, newChar)
		result = append(result, updatedItem)
	}
	return result
}

func replaceInMap(inputMap map[int]string, oldChar, newChar string) map[int]string {
	updatedMap := make(map[int]string)
	for key, value := range inputMap {
		updatedValue := strings.ReplaceAll(value, oldChar, newChar)
		updatedMap[key] = updatedValue
	}
	return updatedMap
}
