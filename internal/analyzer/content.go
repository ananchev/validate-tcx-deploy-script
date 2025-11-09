package analyzer

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ananchev/validate-tcx-deploy-script/internal/logger"
	gitignore "github.com/sabhiram/go-gitignore"
)

// matchPattern checks if a path matches a given ignore pattern using gitignore-style matching.
// It returns true if the path should be ignored according to the pattern.
func matchPattern(pattern, path string) bool {
	// Create a new GitIgnore object using the pattern
	ignore := gitignore.CompileIgnoreLines(pattern)

	// Check if the path should be ignored
	matched := ignore.MatchesPath(path)

	// Return the result
	return matched
}

// shouldIgnore checks if a given path should be ignored based on a list of ignore patterns.
// It returns true if the path matches any of the provided patterns.
// Patterns follow gitignore-style syntax.
func shouldIgnore(path string, ignorePatterns []string) bool {
	for _, pattern := range ignorePatterns {
		if matchPattern(pattern, path) {
			logger.Debug("Excluding path '{path}' as it matches ignore pattern '{p}'", "path", path, "p", pattern)
			return true
		}
	}
	return false
}

// traverseAndCollect recursively walks a directory tree and collects file paths,
// respecting ignore patterns. It continues walking even if individual paths are
// inaccessible, collecting errors along the way.
//
// Parameters:
//   - root: the root directory to start traversing from
//   - ignorePatterns: list of gitignore-style patterns to exclude from results
//
// Returns:
//   - []string: slice of relative file paths found (excluding ignored and inaccessible paths)
//   - error: combined error if any paths were inaccessible during traversal (nil if no errors)
//
// The function logs errors immediately when encountered but continues traversing to collect
// as many valid paths as possible. Directories matching ignore patterns are skipped entirely.
func traverseAndCollect(root string, ignorePatterns []string) ([]string, error) {
	var files []string
	var errors []error

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		// Handle access errors first - before trying to use path/info
		if err != nil {
			logger.Error("Error accessing path '{p}': {e}", "p", path, "e", err.Error())
			errors = append(errors, fmt.Errorf("path %s: %w", path, err))

			// If it's a directory we can't access, skip it entirely
			// Note: info might be nil if the path doesn't exist at all
			if info != nil && info.IsDir() {
				return filepath.SkipDir
			}
			return nil // Skip this file, continue with siblings
		}

		// Now we know the path is accessible - calculate relative path
		relPath, err := filepath.Rel(root, path)
		if err != nil {
			logger.Error("Error calculating relative path for '{p}': {e}", "p", path, "e", err.Error())
			errors = append(errors, fmt.Errorf("relative path %s: %w", path, err))
			return nil // Skip this file, continue walking
		}

		// Check if path matches ignore patterns
		if shouldIgnore(relPath, ignorePatterns) {
			if info.IsDir() {
				logger.Debug("Skipping directory '{relPath}' (matches ignore pattern)", "relPath", relPath)
				return filepath.SkipDir // Don't descend into this directory
			}
			// File is ignored, skip it
			return nil
		}

		// Path is accessible and not ignored - process it
		if !info.IsDir() {
			logger.Debug("Path '{relPath}' should be checked if existing in the script file.", "relPath", relPath)
			files = append(files, relPath)
		} else {
			logger.Debug("Excluding path '{relPath}' as it is a directory", "relPath", relPath)
		}
		return nil
	})

	// Critical error from filepath.Walk itself
	if err != nil {
		errors = append(errors, fmt.Errorf("error walking directory tree: %w", err))
	}

	// Return partial results + error summary
	if len(errors) > 0 {
		return files, fmt.Errorf("encountered %d errors during traversal (see logs for details)", len(errors))
	}

	return files, nil
}

// compareFilesWithScripts compares files found in the repository with paths referenced
// in a deployment script. It verifies that all repository files are referenced in the script.
//
// Parameters:
//   - script: name of the script file being validated
//   - validLines: map of line numbers to file paths extracted from the script
//   - root: root directory of the repository to scan
//   - ignorePatterns: list of gitignore-style patterns to exclude from validation
//
// Returns:
//   - error: any errors encountered during directory traversal (nil if no errors)
//
// The function logs detailed information about files found, files referenced in the script,
// and any discrepancies. It continues validation even if some paths are inaccessible,
// logging errors but returning partial results.
func compareFilesWithScripts(script string, validLines map[int]string, root string, ignorePatterns []string) error {
	logger.Info("Comparison if all repositry files are referenced in the script started for '{script}'", "script", script)
	logger.Info("Repository root is '{r}'", "r", root)
	logger.Info("ignorePatterns are '{ignorePatterns}'", "ignorePatterns", ignorePatterns)

	filesFound, err := traverseAndCollect(root, ignorePatterns)

	// Log the results even if there were errors
	logger.Info("'{files}' files found in the repository after skipping the ignore lines", "files", len(filesFound))
	for i := 0; i < len(filesFound); i++ {
		logger.Debug("\t'{f}'", "f", filesFound[i])
	}

	// If there were errors during traversal, log summary but continue with comparison
	if err != nil {
		logger.Error("Errors occurred during directory traversal: {e}", "e", err.Error())
		// Don't return here - still do the comparison with partial results
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
	hasErrors := false
	for _, item := range filesFound {
		// Check if the item exists in valueSet
		if _, ok := valueSet[item]; !ok {
			logger.Error("Filepath '{item}' does not exist in the script file '{script}'", "item", item, "script", script)
			hasErrors = true
		} else {
			logger.Info("'{item}' is found in the script file '{script}'", "item", item, "script", script)
		}
	}

	if !hasErrors && len(filesFound) > 0 {
		logger.Info("All repository files are referenced in the script")
	} else if len(filesFound) == 0 {
		logger.Info("No files found in repository to check")
	}

	// Return the error at the end so caller knows issues occurred
	return err
}
