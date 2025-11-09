package main

import (
	"flag"
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

	args := ProcessArgs()
	configurationParameters := getConfig(args.ConfigPath)

	err := logger.InitLogger(configurationParameters.Logfile, args.LogLevel)
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer logger.Close()

	analyzer.Run(configurationParameters)

}

func ProcessArgs() Args {
	var a Args

	f := flag.NewFlagSet("Default", 1)
	f.StringVar(&a.ConfigPath, "c", "config.yaml", "path to configuration file")
	f.StringVar(&a.LogLevel, "l", "error", "info, error, or debug logging")

	f.Parse(os.Args[1:])
	return a
}

func getConfig(filename string) analyzer.Parameters {
	var c analyzer.Parameters
	yamlFile, err := os.ReadFile(filename)
	if err != nil {
		panic("Error reading configuration file")
	}
	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		panic("Failed to unmarshal configuration yaml. Bad format?")
	}
	return c
}
