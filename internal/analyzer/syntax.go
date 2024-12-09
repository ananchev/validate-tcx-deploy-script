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

func checkFileSyntax(filePath string, sourceCodeRoot string) {

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
	logger.Separate("-------")
	print("valid", filePath)
	logger.Separate("lines with invalid syntax of referenced filepaths")
	logger.Separate("-------")
	print("invalid", filePath)
	logger.Info("skipped lines")
	logger.Separate("-------")
	print("skipped", filePath)
}

func print(lineType string, filePath string) {
	var lines map[int]string
	switch lineType {
	case "valid":
		lines = AnalyzisResult.File[filePath].Valid
	case "invalid":
		lines = AnalyzisResult.File[filePath].Invalid
	case "skipped":
		lines = AnalyzisResult.File[filePath].Skipped
	}
	if len(lines) == 0 {
		logger.Info("No {lt} entries found", "lt", lineType)
		return
	}

	// sort by line number and print
	si := make([]int, 0, len(lines))
	for i := range lines {
		si = append(si, i)
	}
	sort.Ints(si)
	for _, i := range si {

		if lineType == "invalid" {
			logger.Error("\t{ln}:\t'{val}'", "ln", i, "val", lines[i])
		} else {
			logger.Info("\t{ln}:\t'{val}'", "ln", i, "val", lines[i])
		}
	}
}

func parseLineAsCommand(file string, line string, lineNumber int) {

	logger.Debug("parsing line '{ln} {l}'", "ln", lineNumber, "l", line)

	var skipLine bool = true

	for _, flagName := range _pathParameters {

		logger.Debug("searching for flag '{f}'", "f", flagName)

		pattern := fmt.Sprintf(`-%s`, regexp.QuoteMeta(flagName))
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(line)
		logger.Debug("'{lm}' matches found for flag '{f}'", "lm", len(matches), "f", flagName)

		// Check if the line contains our flags of interest
		if len(matches) < 1 {
			// Continue and try with the next parameter
			logger.Debug("no '{p}' flag found ...", "p", flagName)
			continue
		}

		logger.Debug("checking if the '-{f}' flag definition is properly formatted", "f", flagName)

		pattern = fmt.Sprintf(`-%s="([^"]+)"`, regexp.QuoteMeta(flagName))
		re = regexp.MustCompile(pattern)
		matches = re.FindStringSubmatch(line)

		logger.Debug("'{lm}' matches found", "lm", len(matches))

		// Check if the flag found is properly formatted
		if len(matches) < 2 {
			logger.Debug("line '{l}': '-{s}' is present but not quoted properly", "l", lineNumber, "s", flagName)
			AnalyzisResult.File[file].Invalid[lineNumber] = line
			skipLine = false
			break
		} else if len(matches) == 2 {
			// Extract the file path
			logger.Debug("Formatting correct, extracting the file path in '-{f}'...", "f", flagName)
			filePath := matches[1]
			logger.Debug("filepath is: '{fp}'", "fp", filePath)
			AnalyzisResult.File[file].Valid[lineNumber] = filePath
			skipLine = false
			break
		}
	}

	if skipLine {
		logger.Debug("line '{ln} {l}' does not contain any flag of interest", "ln", lineNumber, "l", line)
		AnalyzisResult.File[file].Skipped[lineNumber] = line
	}

}
