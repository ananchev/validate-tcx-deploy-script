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
	print("valid", filePath)
	logger.Info("stylesheet import")
	print("stylesheet import", filePath)
	logger.Separate("lines with invalid syntax of referenced filepaths")
	print("invalid", filePath)
	logger.Info("skipped lines")
	print("skipped", filePath)
}

func print(lineType string, filePath string) {
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
			logger.Error("'{f}' line '{ln}' is invalid: '{val}'", "f", filePath, "ln", i, "val", lines[i])
		} else {
			logger.Info("'{f}' line '{ln}' is valid: '{val}'", "f", filePath, "ln", i, "val", lines[i])
		}
	}
}

func parseLineAsCommand(file string, line string, lineNumber int) {

	logger.Debug("parsing line '{ln} {l}'", "ln", lineNumber, "l", line)

	var skipLine bool = true

	for _, flagName := range pathParameters {

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
			analysisResult.File[file].Invalid[lineNumber] = line
			skipLine = false // do not capture this line as skip line
			break
		} else if len(matches) == 2 {
			// Extract the file path
			logger.Debug("Formatting correct, extracting the file path in '-{f}'...", "f", flagName)
			filePath := matches[1]
			logger.Debug("filepath is: '{fp}'", "fp", filePath)
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
	regex := regexp.MustCompile(`install_xml_stylesheet_datasets`)
	if !regex.MatchString(line) {
		logger.Debug("'{l}' does not contain 'install_xml_stylesheet_datasets'", "l", line)
		return false
	}
	logger.Debug("'{l}' is refering to 'install_xml_stylesheet_datasets'", "l", line)
	logger.Debug("Extracting the values for input and filepath flags...")

	regex = regexp.MustCompile(`-input="([^"]+)"|-filepath="([^"]+)"`)
	matches := regex.FindAllStringSubmatch(line, -1)

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
