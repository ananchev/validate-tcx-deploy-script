package analyzer

import (
	"bufio"
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

// Paths method returns a map of index to path strings based on the type specified
func (fpm FilePathMap) Paths(pathType string) map[int]string {
	pathsMap := make(map[int]string)

	for key, file := range fpm {
		switch pathType {
		case "relative":
			pathsMap[key] = file.RelativePath
		case "absolute":
			pathsMap[key] = file.AbsolutePath
		default:
			panic("Unrecognized path type. FilePathMap.Paths(pathType string) accepts 'relative' or 'absolute' ")
		}
	}
	return pathsMap
}

func checkStylesheetPaths(scriptFile string, styleSheetImport map[int]StyleSheetImport) {
	logger.Debug("checking stylesheet import paths for '{s}'", "s", scriptFile)
	logger.Debug("found stylesheet imports: '{s_imp_def}'", "s_imp_def", styleSheetImport)
	countImportDefs := len(styleSheetImport)
	index := 1
	for _, importDefinition := range styleSheetImport {
		osLocalizedInputFileLocation := strings.ReplaceAll(importDefinition.InputFile, convertFrom, convertTo)
		osLocalizedXMLsFilePath := strings.ReplaceAll(importDefinition.XMLsFilepath, convertFrom, convertTo)

		logger.Debug("input file '{i}' of '{aa}' is '{s}'", "i", index, "aa", countImportDefs, "s", osLocalizedInputFileLocation)
		index++

		inputFileFullPath := filepath.Join(sourceCodeRoot, osLocalizedInputFileLocation)
		// Open the input text file
		file, err := os.Open(inputFileFullPath)
		if err != nil {
			logger.Error("Error opening '{f}': {err}", "f", inputFileFullPath, "err", err)
			continue
		}
		defer file.Close()

		//xmlFilesReferences := make(map[int]FilePathInfo)
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
			logger.Error("Error reading '{f}': {err}", "f", inputFileFullPath, "err", err)
			continue
		}

		logger.Debug("Checking if all '{n}' stylesheet XMLs referenced in '{f}' exist...", "n", readLinesCount, "f", osLocalizedInputFileLocation)
		checkFilePathsInScript(importDefinition.InputFile, xmlFilesReferences.Paths("absolute"))

		logger.Debug("Comparison if all repositry files in '200-Stylesheets' are referenced in '{input}'", "input", osLocalizedInputFileLocation)
		xmlsLocation := filepath.Join(sourceCodeRoot, osLocalizedXMLsFilePath)
		if err := compareFilesWithScripts(osLocalizedInputFileLocation, xmlFilesReferences.Paths("relative"), xmlsLocation, ignores.StyleSheetsFolder); err != nil {
			logger.Error("Errors occurred during stylesheet comparison for '{input}': {e}", "input", osLocalizedInputFileLocation, "e", err.Error())
			// Continue processing despite errors
		}

	}

}
