package analyzer

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ananchev/validate-tcx-deploy-script/internal/logger"
)

func checkFilePathsInScript(scriptFile string, lines map[int]string) {

	logger.Debug("checking file paths for '{s}'", "s", scriptFile)

	// sort by line number and check
	si := make([]int, 0, len(lines))
	for i := range lines {
		si = append(si, i)
	}
	sort.Ints(si)
	for _, i := range si {
		if fileExists(lines[i]) {
			logger.Info("'{s}' line '{ln}' is valid: file path '{fp}' exists", "s", scriptFile, "ln", i, "fp", lines[i])
		} else {
			logger.Error("'{s}' line '{ln}' is invalid: '{fp}' not found on file system", "s", scriptFile, "ln", i, "fp", lines[i])
		}
	}
}

func fileExists(path string) bool {
	path = strings.ReplaceAll(path, convertFrom, convertTo)
	fullPath := filepath.Join(sourceCodeRoot, path)
	logger.Debug("fullPath: '{f}'", "f", fullPath)
	_, err := os.Stat(fullPath)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}
