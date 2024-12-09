package analyzer

type scriptDefinition struct {
	Filename string `yaml:"filename"`
	TargetOS string `yaml:"target_os"`
}

// Application configuration structure
type Parameters struct {
	Scripts        []scriptDefinition `yaml:"scripts"`
	PathParameters []string           `yaml:"path_parameters"`
	SourceCodeRoot string             `yaml:"source_code_root"`
	Logfile        string             `yaml:"logfile"`
}
