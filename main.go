package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"

	"github.com/ananchev/validate-tcx-deploy-script/internal/analyzer"
	"github.com/ananchev/validate-tcx-deploy-script/internal/logger"
	"gopkg.in/yaml.v3"
)

// Args command-line parameters
type Args struct {
	ConfigPath string
	LogLevel   string
}

func main() {
	err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	args := ProcessArgs()

	configurationParameters, err := getConfig(args.ConfigPath)
	if err != nil {
		return err
	}

	err = logger.InitLogger(configurationParameters.Logfile, args.LogLevel)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer logger.Close()

	analyzer.Run(configurationParameters)
	return nil
}

func ProcessArgs() Args {
	var a Args

	f := flag.NewFlagSet("Default", 1)
	f.StringVar(&a.ConfigPath, "c", "config.yaml", "path to configuration file")
	f.StringVar(&a.LogLevel, "l", "error", "info, error, or debug logging")

	f.Parse(os.Args[1:])
	return a
}

func getConfig(filename string) (analyzer.Parameters, error) {
	var c analyzer.Parameters

	// Check if file exists first for better error message
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return c, fmt.Errorf("configuration file '%s' not found", filename)
	}

	yamlFile, err := os.ReadFile(filename)
	if err != nil {
		return c, fmt.Errorf("error reading configuration file '%s': %w", filename, err)
	}

	// Check for tabs in YAML (common mistake that causes parsing errors)
	if bytes.Contains(yamlFile, []byte("\t")) {
		return c, fmt.Errorf("invalid YAML in '%s': file contains tabs. YAML requires spaces for indentation, not tabs", filename)
	}

	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		return c, fmt.Errorf("invalid YAML format in '%s': %w", filename, err)
	}

	// Validate the configuration
	err = validateConfig(&c)
	if err != nil {
		return c, fmt.Errorf("configuration validation failed in '%s': %w", filename, err)
	}

	return c, nil
}

func validateConfig(c *analyzer.Parameters) error {
	// Validate scripts list
	if len(c.Scripts) == 0 {
		return fmt.Errorf("'scripts' list cannot be empty")
	}

	// Validate each script
	for i, script := range c.Scripts {
		if script.Filename == "" {
			return fmt.Errorf("script at index %d is missing 'filename'", i)
		}
		if script.TargetOS == "" {
			return fmt.Errorf("script '%s' is missing 'target_os'", script.Filename)
		}
		if script.TargetOS != "windows" && script.TargetOS != "linux" {
			return fmt.Errorf("script '%s' has invalid 'target_os': '%s' (must be 'windows' or 'linux')",
				script.Filename, script.TargetOS)
		}
	}

	// Validate source_code_root
	if c.SourceCodeRoot == "" {
		return fmt.Errorf("'source_code_root' is required")
	}

	// Validate path_parameters
	if len(c.PathParameters) == 0 {
		return fmt.Errorf("'path_parameters' list cannot be empty")
	}

	return nil
}
