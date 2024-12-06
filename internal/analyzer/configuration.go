package analyzer

// Application configuration structure
type Parameters struct {
	SourceCodeFolder string   `yaml:"source_code_folder"`
	PathParameters   []string `yaml:"path_parameters"`
	Logfile          string   `yaml:"logfile"`
}
