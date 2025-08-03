# Code Generator CLI

A extensible, plugin-based CLI that automatically generates infrastructure and deployment configurations by discovering and processing structured configuration files.

## Architecture Overview

### Plugin System
The CLI uses a plugin architecture where each plugin handles a specific technology or deployment target. Plugins implement a simple interface:

```go
type Plugin interface {
    Name() string                    // Plugin identifier
    ConfigFile() string              // Config file pattern to match
    Generate(ctx, configPath, outputDir) error  // Generation logic
}
```

### Directory Organization
Configuration files are organized in technology-specific directories for better project structure:

```
project-root/
├── technology-name/
│   └── config-files...          # Technology-specific configs
├── another-tech/
│   ├── service-a.yaml           # Multiple configs supported
│   └── service-b.yaml
└── generated/                   # Output directory
```

### Discovery Mechanism
The file scanner automatically:
1. Scans predefined technology directories 
2. Matches found configs to registered plugins
3. Executes generation for each matched plugin/config pair

### Plugin Implementation
Each plugin contains:
- **Configuration parsing** - YAML/JSON structure definitions
- **Template generation** - Technology-specific file generation
- **Output organization** - Structured file/directory creation
- **Integration logic** - Tool-specific automation (tasks, scripts)

## Usage Patterns

```bash
# Auto-discovery: Generate all found configurations
./codegen

# Targeted: Generate specific technology configurations  
./codegen [plugin-name]
```

## Extensibility

### Adding New Plugins
1. Create plugin directory: `pkg/plugins/[technology]/`
2. Implement the Plugin interface
3. Register in main.go
4. Define config file pattern and generation logic

### Plugin Types
- **Single-config plugins**: One config file per technology (e.g., infrastructure)
- **Multi-config plugins**: Multiple config files per technology (e.g., microservices)

## Development

```bash
# Build the CLI
go build -o codegen .

# Run tests
go test ./test/e2e/... -v

# Add new plugin
mkdir pkg/plugins/[technology]
```
