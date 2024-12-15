# Execution flow
```mermaid
---
config:
    noteAlign: left
---
sequenceDiagram
    autonumber
    participant User
    box Code Execution
    participant Main
    participant Logger
    participant Analyzer
    end

    User->>Main: Run application
    Note right of User: <executable> -c path/to/<config.yml>
    Note right of User: <config.yml> <br> - Deployment scripts filenames and target operating system <br> - Arguments for which to extract & check file paths <br> - Exclusions when checking repository content vs. scripts<br> - Local directory where TC configuriton files are stored
    
    Main->>Logger: Initialize logger
    Main->>Analyzer: Load Configuration
   

    

    loop Analyze the config-n deployment scripts for Linux & Windows
        Analyzer->>Analyzer: Check syntax of the script <br> arguments providing file paths
        Analyzer->>Analyzer: Are all file paths in the <br> script existing on file system?
        Analyzer->>Analyzer: Are there files on file system <br> not included in the script?
        Analyzer->>Analyzer: Check input file for stylesheet import <br> and if all XML files exist on disk
        critical Failed checks
            Analyzer->>Logger: Log error specifying failed check
        end
        opt If logging level = info
                Analyzer->>Logger: Log successful checks
        end

    end
    Logger->>User: Output analysis results
    Note left of Main: <Logfile path as set in config>.log
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