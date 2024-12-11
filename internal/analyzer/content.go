package analyzer

import (
	"os"
	"path/filepath"

	"github.com/ananchev/validate-tcx-deploy-script/internal/logger"
	gitignore "github.com/sabhiram/go-gitignore"
)

// matchPattern checks if a path matches a given ignore pattern
func matchPattern(pattern, path string) bool {
	// Create a new GitIgnore object using the pattern
	ignore := gitignore.CompileIgnoreLines(pattern)

	// Check if the path should be ignored
	matched := ignore.MatchesPath(path)

	// Return the result
	return matched
}

// shouldIgnore checks if a given path should be ignored based on patterns
func shouldIgnore(path string, ignorePatterns []string) bool {
	for _, pattern := range ignorePatterns {
		if matchPattern(pattern, path) {
			logger.Debug("Excluding path '{path}' as it matches ignore pattern '{p}'", "path", path, "p", pattern)
			return true
		} else {
			logger.Debug("Path '{path}' does not match ignore pattern '{p}'", "path", path, "p", pattern)

		}
	}
	return false
}

// TraverseAndCollect traverses directories and collects paths according to ignore rules
func traverseAndCollect(root string, ignorePatterns []string) []string {
	var files []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			logger.Error("Error accessing path '{p}': {e}", "p", path, "e", err.Error())
			os.Exit(1)
		}

		relPath, err := filepath.Rel(root, path)
		if err != nil {
			logger.Error("Error calculating relative path for '{p}': {e}", "p", path, "e", err.Error())
			os.Exit(1)
		}

		// Skip ignored files and directories
		if shouldIgnore(relPath, ignorePatterns) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if !info.IsDir() {
			logger.Debug("Path '{relPath}' should be checked if existing in the script file.", "relPath", relPath)
			files = append(files, relPath)
		} else {
			logger.Debug("Excluding path '{relPath}' as it is a directory", "relPath", relPath)

		}
		return nil
	})

	if err != nil {
		logger.Error("Error walking the directory tree: {err}", "err", err)
		os.Exit(1) // Exit immediately with a non-zero status
	}
	return files
}

func compareFilesWithScripts(script string, validLines map[int]string, root string, ignorePatterns []string) {
	logger.Info("Comparison if all repositry files are referenced in the script started for '{script}'", "script", script)
	logger.Info("Repository root is '{r}'", "r", root)
	logger.Info("ignorePatterns are '{ignorePatterns}'", "ignorePatterns", ignorePatterns)

	filesFound := traverseAndCollect(root, ignorePatterns)
	logger.Info("'{files}' files found in the repository after skipping the ignore lines", "files", len(filesFound))
	for i := 0; i < len(filesFound); i++ {
		logger.Debug("\t'{f}'", "f", filesFound[i])
	}
	valueSet := make(map[string]struct{})
	logger.Info("'{valid}' valid lines found in script '{s}'", "valid", len(validLines), "s", script)
	for _, v := range validLines {
		logger.Debug("\t'{v}'", "v", v)
	}

	for _, value := range validLines {
		valueSet[value] = struct{}{}
	}
	logger.Debug("Searching for files in the repository that are not present as valid lines in the script '{s}'...", "s", script)
	// Iterate through the slice and check each item
	for _, item := range filesFound {
		// Check if the item exists in valueSet
		if _, ok := valueSet[item]; !ok {
			logger.Error("Filepath '{item}' does not exist in the script file '{script}'", "item", item, "script", script)
		} else {
			logger.Info("'{item}' is found in the script file '{script}'", "item", item, "script", script)
		}
	}

}
