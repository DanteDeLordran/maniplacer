# Maniplacer
A CLI tool for generating Kubernetes YAML manifests from templates with intelligent configuration handling.

## Features
- **Project scaffolding** - Initialize new projects with `init`
- **Template management** - Add, list, and remove templates (`add`, `list`, `remove`)
- **Smart manifest generation** - Generate manifests from templates with automatic config detection (`generate`)
- **Multiple config formats** - Support for JSON and YAML configuration files
- **Intelligent config detection** - Automatically detects and handles multiple config files
- **Easy updates** - Update the CLI easily via `update`
- **Namespace support** - Organize templates and manifests by namespaces
- **Rich template helpers** - Base64, ToUpper, ToLower, and more built-in functions
- **Timestamped outputs** - Each generation creates a unique timestamped folder

## How to Install

### Via Script
```bash
curl -fsSL https://raw.github.com/dantedelordran/maniplacer/main/installer.sh | bash
```

### Via Source
```bash
git clone https://github.com/dantedelordran/maniplacer.git
cd maniplacer
make build
```

## Quickstart

### Initialize a New Project
```bash
# Create a new project
maniplacer init my-project
cd my-project
```

This creates the following structure:
```bash
my-project/
├── templates/
│   └── default/
├── manifests/
├── config.json          # Default JSON config
├── config.yaml          # Alternative YAML config (optional)
└── .maniplacer          # Project marker file
```

### Prepare a Sandbox (Optional)
```bash
docker build -t maniplacer:latest .
docker run -it --name sandbox maniplacer:latest
```

## Configuration Files

Maniplacer supports multiple configuration formats and intelligent detection:

### Supported Formats
- **JSON**: `config.json`
- **YAML**: `config.yaml` or `config.yml`

### Smart Detection
- **Single config file**: Automatically detects and uses it
- **Multiple config files**: Prompts you to choose which one to use
- **Custom config files**: Use `-c` flag to specify any config file
- **Format specification**: Use `-f` flag to specify format explicitly

### Example Configuration Files

**config.json**
```json
{
  "appName": "my-app",
  "version": "1.0.0",
  "replicas": 3,
  "image": "nginx:latest",
  "ports": [80, 443]
}
```

**config.yaml**
```yaml
appName: my-app
version: "1.0.0"
replicas: 3
image: nginx:latest
ports:
  - 80
  - 443
```

## Usage Examples

### Basic Generation
```bash
# Auto-detect config file and generate manifests
maniplacer generate

# Use specific namespace
maniplacer generate -n production

# Specify repository
maniplacer generate -r my-repo
```

### Configuration Options
```bash
# Prefer YAML format (if multiple configs exist)
maniplacer generate -f yaml

# Use custom config file
maniplacer generate -c production.json

# Use custom config with explicit format
maniplacer generate -c my-config.txt -f json

# Combine all options
maniplacer generate -c prod.yaml -f yaml -n production -r backend
```

### Interactive Config Selection
When multiple config files exist, Maniplacer will prompt you:
```bash
Multiple configuration files found:
  1) config.json
  2) config.yaml

Please choose which config file to use (1-2): 1
Selected: config.json
Using JSON config file: /path/to/config.json
```

### Template Management
```bash
# Add a new template
maniplacer add deployment.yaml

# List all templates
maniplacer list

# Remove a template
maniplacer remove deployment.yaml

# Work with specific namespace
maniplacer add -n staging service.yaml
maniplacer list -n staging
```

## Command Reference

### `maniplacer init <project-name>`
Initialize a new Maniplacer project with the specified name.

### `maniplacer generate [flags]`
Generate Kubernetes manifests from templates and configuration.

**Flags:**
- `-f, --format string`: Config file format (json, yaml, yml). Auto-detects if not specified
- `-n, --namespace string`: Template namespace (default "default")
- `-r, --repo string`: Repository name
- `-c, --config string`: Custom path to config file

### `maniplacer add <template-file> [flags]`
Add a new template to the project.

**Flags:**
- `-n, --namespace string`: Target namespace (default "default")

### `maniplacer list [flags]`
List all templates in the project.

**Flags:**
- `-n, --namespace string`: Filter by namespace

### `maniplacer remove <template-file> [flags]`
Remove a template from the project.

**Flags:**
- `-n, --namespace string`: Target namespace (default "default")

### `maniplacer update [flags]`
Update Maniplacer to the latest version.

**Flags:**
- `-f, --force`: Force update even if already on latest version

## Template Helpers

Maniplacer provides built-in template functions for common operations:

```yaml
# String manipulation
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ .appName | ToLower }}
  labels:
    version: {{ .version | ToUpper }}

# Base64 encoding
data:
  config: {{ .configData | Base64 }}

# And more helpers available...
```

## Output Structure

Generated manifests are organized with timestamps for safe, repeatable generation:

```bash
manifests/
├── default/
│   ├── 2024-01-15_14-30-45/
│   │   ├── deployment.yaml
│   │   └── service.yaml
│   └── 2024-01-15_15-22-10/
│       ├── deployment.yaml
│       └── service.yaml
└── production/
    └── 2024-01-15_16-45-30/
        ├── deployment.yaml
        └── configmap.yaml
```

## Update to Latest Version

```bash
# Standard update
maniplacer update

# Force update
maniplacer update -f
```

## Best Practices

1. **Use descriptive config files** - Name your configs based on environment (e.g., `config-prod.yaml`)
2. **Organize by namespaces** - Separate templates for different environments or applications
3. **Version control everything** - Keep templates, configs, and generated manifests in git
4. **Test templates** - Generate manifests in development before deploying to production
5. **Leverage template helpers** - Use built-in functions for common transformations

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.