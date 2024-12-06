package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
)

// parseLineAsCommand treats each line as if it were a command with flags and parses it.
func parseLineAsCommand(line string, lineNumber int) (map[string]string, error) {
	// Create a new FlagSet to parse the line
	fs := flag.NewFlagSet("lineParser", flag.ContinueOnError)

	// Define possible flags
	xmlFilePath := fs.String("xml_file", "", "Path to the XML file")
	filePath := fs.String("file", "", "General file path")
	filePathAlt := fs.String("file_path", "", "Alternative file path")

	// This disables standard error output on parse failure to avoid cluttering output with usage info
	fs.SetOutput(nil)

	// Split the line into arguments
	args := strings.Fields(line)

	// Parse the arguments
	err := fs.Parse(args)
	if err != nil {
		return nil, fmt.Errorf("Line %d: could not parse line as command - %v", lineNumber, err)
	}

	// Collect results
	results := make(map[string]string)
	if *xmlFilePath != "" {
		results["xml_file"] = *xmlFilePath
	}
	if *filePath != "" {
		results["file"] = *filePath
	}
	if *filePathAlt != "" {
		results["file_path"] = *filePathAlt
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("Line %d: none of the expected parameters were found or properly enclosed", lineNumber)
	}

	return results, nil
}

func main() {
	// Define the command-line flag for the input file
	inputFile := flag.String("inputFile", "", "Path to the input file containing lines to parse")
	flag.Parse()

	if *inputFile == "" {
		fmt.Println("Error: You must specify an input file using the -inputFile flag.")
		return
	}

	file, err := os.Open(*inputFile)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer file.Close()

	// Read lines from the file
	scanner := bufio.NewScanner(file)
	lineNumber := 0
	validEntries := make(map[int]map[string]string)
	var errorLog []string

	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()

		results, err := parseLineAsCommand(line, lineNumber)
		if err != nil {
			errorLog = append(errorLog, err.Error())
		} else {
			validEntries[lineNumber] = results
		}
	}

	fmt.Println("Valid entries:")
	for lineNumber, params := range validEntries {
		fmt.Printf("Line %d:\n", lineNumber)
		for key, value := range params {
			fmt.Printf("  %s: %s\n", key, value)
		}
	}

	fmt.Println("\nError Log:")
	for _, error := range errorLog {
		fmt.Println(error)
	}
}
