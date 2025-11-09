package analyzer

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ananchev/validate-tcx-deploy-script/internal/logger"
)

// $TC_BIN/install_xml_stylesheet_datasets -u=$INSTALL_USER -p=$TC_USER_PASSWD -g=dba \
// -input="200-Stylesheets/import_stylesheet.txt" \
// -filepath="200-Stylesheets/" \
// -replace

type FilePathInfo struct {
	RelativePath string
	AbsolutePath string
}

// Below type is needed to define methods executed on a map [int]FilePathCollection
type FilePathMap map[int]FilePathInfo

// Paths method returns a map of index to path strings based on the type specified.
// Returns an error if pathType is not "relative" or "absolute".
func (fpm FilePathMap) Paths(pathType string) (map[int]string, error) {
	pathsMap := make(map[int]string)

	for key, file := range fpm {
		switch pathType {
		case "relative":
			pathsMap[key] = file.RelativePath
		case "absolute":
			pathsMap[key] = file.AbsolutePath
		default:
			return nil, fmt.Errorf("unrecognized path type %q: must be 'relative' or 'absolute'", pathType)
		}
	}
	return pathsMap, nil
}

// processStylesheetInputFile processes a single stylesheet import definition file.
// It reads the input file, extracts XML file references, validates they exist,
// and compares repository files with script references.
//
// Parameters:
//   - importDefinition: The stylesheet import definition containing input file and XML paths
//
// Returns:
//   - error: Any error encountered during processing, or nil on success
func processStylesheetInputFile(importDefinition StyleSheetImport) error {
	osLocalizedInputFileLocation := strings.ReplaceAll(importDefinition.InputFile, convertFrom, convertTo)
	osLocalizedXMLsFilePath := strings.ReplaceAll(importDefinition.XMLsFilepath, convertFrom, convertTo)

	inputFileFullPath := filepath.Join(sourceCodeRoot, osLocalizedInputFileLocation)

	// Open the input text file
	file, err := os.Open(inputFileFullPath)
	if err != nil {
		return fmt.Errorf("error opening %q: %w", inputFileFullPath, err)
	}
	defer file.Close() // Properly closes when function returns

	xmlFilesReferences := FilePathMap{}
	readLinesCount := 0
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		readLinesCount++

		// Split each line by the comma
		columns := strings.Split(line, ",")
		if len(columns) >= 2 {
			// Trim spaces, form the full path, and append to the slice with files to check if existing on the file system
			filePath := strings.ReplaceAll(importDefinition.XMLsFilepath, convertFrom, convertTo)
			fileName := strings.TrimSpace(columns[1])
			pathToStylesheetXML := filepath.Join(filePath, fileName)
			logger.Debug("stylesheet XML absolute path: '{p}'", "p", pathToStylesheetXML)
			logger.Debug("stylesheet XML relative path: '{p}'", "p", fileName)

			xmlFilesReferences[readLinesCount] = FilePathInfo{RelativePath: fileName, AbsolutePath: pathToStylesheetXML}
		} else {
			logger.Error("Line '{l}' is of invalid format", "l", line)
		}
	}

	logger.Info("Read '{n}' lines from '{f}'", "n", readLinesCount, "f", osLocalizedInputFileLocation)

	// Check for any errors encountered during scanning
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading %q: %w", inputFileFullPath, err)
	}

	// Get absolute paths for validation
	absolutePaths, err := xmlFilesReferences.Paths("absolute")
	if err != nil {
		return fmt.Errorf("error getting absolute paths: %w", err)
	}

	logger.Debug("Checking if all '{n}' stylesheet XMLs referenced in '{f}' exist...", "n", readLinesCount, "f", osLocalizedInputFileLocation)
	checkFilePathsInScript(importDefinition.InputFile, absolutePaths)

	// Get relative paths for comparison
	relativePaths, err := xmlFilesReferences.Paths("relative")
	if err != nil {
		return fmt.Errorf("error getting relative paths: %w", err)
	}

	logger.Debug("Comparison if all repositry files in '200-Stylesheets' are referenced in '{input}'", "input", osLocalizedInputFileLocation)
	xmlsLocation := filepath.Join(sourceCodeRoot, osLocalizedXMLsFilePath)

	if err := compareFilesWithScripts(osLocalizedInputFileLocation, relativePaths, xmlsLocation, ignores.StyleSheetsFolder); err != nil {
		return fmt.Errorf("stylesheet comparison errors: %w", err)
	}

	return nil
}

func checkStylesheetPaths(scriptFile string, styleSheetImport map[int]StyleSheetImport) {
	logger.Debug("checking stylesheet import paths for '{s}'", "s", scriptFile)
	logger.Debug("found stylesheet imports: '{s_imp_def}'", "s_imp_def", styleSheetImport)
	countImportDefs := len(styleSheetImport)
	index := 1

	for _, importDefinition := range styleSheetImport {
		osLocalizedInputFileLocation := strings.ReplaceAll(importDefinition.InputFile, convertFrom, convertTo)
		logger.Debug("input file '{i}' of '{aa}' is '{s}'", "i", index, "aa", countImportDefs, "s", osLocalizedInputFileLocation)
		index++

		// Process each stylesheet import file
		if err := processStylesheetInputFile(importDefinition); err != nil {
			logger.Error("Error processing stylesheet import file '{f}': {err}", "f", osLocalizedInputFileLocation, "err", err)
			// Continue processing other imports despite errors
			continue
		}
	}
}
