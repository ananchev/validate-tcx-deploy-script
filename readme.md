# Execution flow
```mermaid
sequenceDiagram
    autonumber
    participant User
    participant Main
    participant ConfigLoader
    participant Logger
    participant Analyzer

    Note over User: scripts-check -c path/to/config
    User->>Main: Run application
    Main->>ConfigLoader: Load Configuration
    ConfigLoader->>Logger: Initialize logger
    ConfigLoader->>Analyzer: Provide configurations

    loop Analyze each script
        Analyzer->>Analyzer: Check script syntax validity
        Analyzer->>Analyzer: Do all paths exist on file system?
        Analyzer->>Analyzer: Are all files from disk in the script?
        Analyzer->>Analyzer: Validate stylesheet imports

        alt Passed checks
            opt Log level permits info logging
                Analyzer->>Logger: Log info about successful checks
            end
        else Failed checks
            Analyzer->>Logger: Log error specifying failed check
        end
    end

    Logger->>User: Output analysis results
    
```

# Example configuration file
```yml
scripts:
  - filename:	DeploymentInstructions.bat
    target_os:	windows
  - filename:	DeploymentInstructions.sh
    target_os:	linux
path_parameters:
  - input
  - xml_file
  - name
  - path
  - file
source_code_root: "/path/to/configuration/repo"
ignore_patterns:
  global:
    - "001-Start_Automation"
    - "003-Infrastructure_Automation"
    - "040-SourceCode"
    - "058-Application"
    - "060-Binaries"
    - "070-BMIDE"
    - "090-BatchLoV"
    - "200-Stylesheets" # path for stylesheet XMLs is checked separately
    - "900-Import_Export_Tools"   
    - "950-APE"
    - "BNL-Import_Export_Tools"
    - "*.adoc"
    - "Policy and Best Practices.txt"
    - "readme.txt"
    - "README.md"
    - "DeploymentInstructions.bat"
    - "DeploymentInstructions.sh"
  stylesheets_folder:
    - "*.txt"
logfile: execution.log
```