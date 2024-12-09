package analyzer

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ananchev/validate-tcx-deploy-script/internal/logger"
)

var _sourceCodeRoot string

func checkFilePaths(scriptFile string, sourceCodeRoot string, lines map[int]string) {

	logger.Debug("checking file paths for '{s}'", "s", scriptFile)
	_sourceCodeRoot = sourceCodeRoot

	// sort by line number and check
	si := make([]int, 0, len(lines))
	for i := range lines {
		si = append(si, i)
	}
	sort.Ints(si)
	for _, i := range si {
		if fileExists(lines[i]) {
			logger.Info("Line {ln} is valid: file paths '{fp}' exists", "ln", i, "fp", lines[i])
		} else {
			logger.Error("Line {ln} is invalid: '{fp}' not found on file system", "ln", i, "fp", lines[i])
		}
	}
}

func fileExists(path string) bool {
	p := covertFilePaths(path)

	fullPath := filepath.Join(_sourceCodeRoot, p)
	_, err := os.Stat(fullPath)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

func covertFilePaths(path string) (convertedPath string) {
	if len(convertFrom) > 0 {
		logger.Debug("Converting the '{cf}' in '{s}' to '{ct}'...", "cf", convertFrom, "s", path, "ct", convertTo)
		res := strings.Replace(path, convertFrom, convertTo, -1)
		logger.Debug("Resulting path is '{res}'", "res", res)
		return res
	}
	return path
}
