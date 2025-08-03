# Code Generator CLI

A plugin-based CLI that automatically generates code configurations by scanning files in the root directory.

## How it works

The CLI scans your root directory for configuration files and generates the corresponding code based on plugins:

- `terraform.yaml` → generates Terraform modules using `./pkg/plugins/terraform`

Each plugin defines the generation requirements and templates for its specific technology.

## Installation

```bash
go install
```

## Usage

```bash
# Generate all configurations based on files found in root
./codegen

# Generate specific configuration
./codegen terraform
```

## Plugin Structure

```
pkg/plugins/
├── terraform/     # Terraform generation logic
```

## Development

```bash
# Build the CLI
go build -o codegen ./cmd

# Run tasks
task [task-name]
```
