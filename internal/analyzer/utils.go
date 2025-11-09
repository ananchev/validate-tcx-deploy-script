package analyzer

import (
	"fmt"
	"runtime"
	"strings"
)

// determinePathConversion determines the path separator conversion needed
// based on the target OS and runtime OS.
//
// Parameters:
//   - targetOS: The target operating system ("windows" or "linux")
//   - scriptFilename: The script filename for error messages
//
// Returns:
//   - convertFrom: The character to convert from
//   - convertTo: The character to convert to
//   - error: Any validation error
func determinePathConversion(targetOS, scriptFilename string) (convertFrom, convertTo string, err error) {
	runtimeOS := runtime.GOOS

	// Validate target OS
	if targetOS != "linux" && targetOS != "windows" {
		return "", "", fmt.Errorf("incorrect specification of script target_os for %q: must be 'linux' or 'windows', got %q", scriptFilename, targetOS)
	}

	// Determine conversion
	if targetOS == "windows" && runtimeOS == "linux" {
		return `\`, `/`, nil
	} else if targetOS == "linux" && runtimeOS == "windows" {
		return `/`, `\`, nil
	}

	// No conversion needed (matching OS)
	return "", "", nil
}

// replaceInMap replaces all occurrences of oldChar with newChar in map values.
// Returns a new map with updated values.
func replaceInMap(inputMap map[int]string, oldChar, newChar string) map[int]string {
	updatedMap := make(map[int]string)
	for key, value := range inputMap {
		updatedValue := strings.ReplaceAll(value, oldChar, newChar)
		updatedMap[key] = updatedValue
	}
	return updatedMap
}

// replaceInIgnorePatterns replaces all occurrences of oldChar with newChar
// in both Global and StyleSheetsFolder pattern slices.
// Returns a new ignorePatterns struct with updated values.
func replaceInIgnorePatterns(patterns ignorePatterns, oldChar, newChar string) ignorePatterns {
	// Helper function to replace characters in a slice of strings
	replaceInSlice := func(slice []string, oldChar, newChar string) []string {
		var result []string
		for _, item := range slice {
			updatedItem := strings.ReplaceAll(item, oldChar, newChar)
			result = append(result, updatedItem)
		}
		return result
	}

	// Perform replacements on Global and StyleSheetsFolder slices
	patterns.Global = replaceInSlice(patterns.Global, oldChar, newChar)
	patterns.StyleSheetsFolder = replaceInSlice(patterns.StyleSheetsFolder, oldChar, newChar)

	return patterns
}
