# Maniplacer

A powerful Kubernetes manifest templating tool that simplifies the creation and management of K8s resources through customizable templates and configuration-driven generation.

## Overview

Maniplacer helps you manage Kubernetes manifests efficiently by:
- **Scaffolding** new component templates with sensible defaults
- **Templating** manifests using Go's template engine with custom functions
- **Generating** production-ready manifests from configuration files
- **Organizing** resources by namespace and repository structure

## Features

- 🚀 **Project Management**: Initialize projects and create repositories with `init` and `new`
- 🔧 **Template Management**: Add, remove, and list component templates
- 📁 **Smart Generation**: Generate manifests with intelligent config detection
- 🔄 **Multi-format Support**: Works with JSON and YAML configuration files
- 📚 **Built-in Documentation**: Local documentation server with examples
- ⏰ **Timestamped Outputs**: Each generation creates a unique timestamped folder
- 🧹 **Cleanup Tools**: Prune manifests and remove templates easily
- 🔄 **Self-updating**: Update to latest version from GitHub releases

## Installation

### Via Curl (Recommended)
```bash
curl -fsSL https://raw.github.com/dantedelordran/maniplacer/main/installer.sh | bash
```

### Via Source
```bash
git clone https://github.com/dantedelordran/maniplacer.git
cd maniplacer
make build
```

## Quick Start

### 1. Initialize a Project

```bash
# Create a new Maniplacer project in current directory
maniplacer init

# Or create a new project with a specific name
maniplacer init --name my-k8s-project
```

### 2. Create a Repository

```bash
# Create a new repository within your project
maniplacer new myapp
```

### 3. Add Component Templates

```bash
# Add templates for common Kubernetes resources
maniplacer add deployment service configmap -n production -r myapp

# Available components:
# - deployment    (workloads with containers and replicas)
# - service       (network-accessible services)
# - httpRoute     (HTTP routing rules)
# - secret        (secure storage for sensitive data)
# - configmap     (configuration key-value pairs)
```

### 4. Create Configuration

Create a `config.yaml` (or `config.json`) in your repository directory:

```yaml
# config.yaml
name: myapp
namespace: production
replicas: 3
image: myapp:v1.2.3
port: 8080

secrets:
  db_password: supersecret123
  api_key: abc123xyz789
```

### 5. Generate Manifests

```bash
# Generate manifests from templates
maniplacer generate -n production -r myapp

# Use specific config format
maniplacer generate -f yaml -n production -r myapp

# Use custom config file
maniplacer generate -c custom-config.json -n production -r myapp
```

### 6. List Generated Manifests

```bash
# List all manifests in a namespace
maniplacer list -n production -r myapp

# List manifests in default namespace
maniplacer list -r myapp
```

## Project Structure

```
my-k8s-project/
├── .maniplacer              # Project marker file
├── myapp/                   # Repository directory
│   ├── config.yaml          # Configuration values
│   ├── templates/           # Template definitions
│   │   └── production/      # Namespace-specific templates
│   │       ├── deployment.yaml
│   │       ├── service.yaml
│   │       └── configmap.yaml
│   └── manifests/           # Generated outputs
│       └── production/      # Namespace-specific manifests
│           └── 2024-01-15_14-30-45/  # Timestamped generation
│               ├── deployment.yaml
│               ├── service.yaml
│               └── configmap.yaml
```

## Template Engine

Maniplacer uses Go's powerful template engine with custom functions:

### Basic Templating
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: {{ .name }}
  namespace: {{ .namespace }}
type: Opaque
data:
  {{- range $key, $value := .secrets }}
  {{ $key }}: {{ $value | Base64 | Quote }}
  {{- end }}
```

### Built-in Functions
- **`Base64`** - Encode strings to Base64
- **`ToUpper`** - Convert to uppercase
- **`ToLower`** - Convert to lowercase  
- **`Quote`** - Wrap in quotes

### Example Usage
```yaml
# Template
env:
- name: DATABASE_URL
  value: {{ .db_url | Quote }}
- name: APP_ENV
  value: {{ .environment | ToUpper | Quote }}
```

## Commands

### `maniplacer init`
Bootstrap a new Maniplacer project with the required folder structure.

```bash
# Initialize in current directory (with confirmation)
maniplacer init

# Create new project with specific name
maniplacer init --name my-k8s-project

# Available options:
# -n, --name        Project name (creates new directory if specified)
```

During initialization:
- Creates project root and marks it as a valid Maniplacer project
- Sets up required directories (templates/, manifests/)
- Optionally creates a default repository with config.json

### `maniplacer new`
Create a new repository within an existing Maniplacer project.

```bash
# Create a new repository
maniplacer new frontend
maniplacer new backend-api

# Each repository gets:
# - templates/ directory for resource templates
# - manifests/ directory for generated files
# - config.json configuration file
```

### `maniplacer add`
Scaffold new Kubernetes component templates.

```bash
# Add single component
maniplacer add deployment -n staging -r myrepo

# Add multiple components
maniplacer add deployment service secret -n production -r myapp

# Available options:
# -n, --namespace   Target namespace (default: "default")
# -r, --repo        Repository name (required)
```

### `maniplacer generate`
Generate manifests from templates and configuration.

```bash
# Basic generation
maniplacer generate -r myrepo

# Specify namespace and format
maniplacer generate -n production -f yaml -r myrepo

# Use custom config file
maniplacer generate -c /path/to/config.json -r myrepo

# Available options:
# -n, --namespace   Template namespace (default: "default")
# -f, --format      Config format: json, yaml, yml (auto-detected if not specified)
# -r, --repo        Repository name (required)
# -c, --config      Custom config file path
```

### `maniplacer list`
Display all generated manifests in a specific namespace and repository.

```bash
# List manifests in default namespace
maniplacer list -r myrepo

# List manifests in specific namespace
maniplacer list -n production -r backend-service

# Available options:
# -n, --namespace   Target namespace (default: "default")
# -r, --repo        Repository name (required)
```

### `maniplacer remove`
Remove component templates from the templates directory.

```bash
# Remove single component
maniplacer remove service -n production -r myrepo

# Remove multiple components
maniplacer remove deployment service configmap -n staging -r myapp

# Available options:
# -n, --namespace   Target namespace (default: "default")
# -r, --repo        Repository name (required)
```

### `maniplacer prune`
Delete all generated manifests in a specific namespace (with confirmation).

```bash
# Prune manifests in default namespace
maniplacer prune -r myrepo

# Prune manifests in specific namespace
maniplacer prune -n staging -r myapp

# Available options:
# -n, --namespace   Target namespace (default: "default")  
# -r, --repo        Repository name (required)
```

### `maniplacer update`
Update Maniplacer to the latest version from GitHub releases.

```bash
# Check for updates and update with confirmation
maniplacer update

# Force update without confirmation
maniplacer update --force

# Available options:
# -f, --force       Skip confirmation prompt
```

The update process:
1. Fetches the latest release from GitHub
2. Compares with current version
3. Downloads appropriate binary for your OS/architecture
4. Creates backup and replaces binary via update script

### `maniplacer docs`
Launch local documentation server with examples and function reference.

```bash
# Start docs server (default port 8000)
maniplacer docs

# Use custom port
maniplacer docs -p 9000

# Available options:
# -p, --port        Server port (default: "8000")
```

Access documentation at: `http://localhost:8000/docs`

## Configuration Formats

### JSON Configuration
```json
{
  "name": "myapp",
  "namespace": "production",
  "replicas": 3,
  "image": "myapp:v1.2.3",
  "secrets": {
    "db_password": "supersecret123",
    "api_key": "abc123xyz789"
  }
}
```

### YAML Configuration
```yaml
name: myapp
namespace: production
replicas: 3
image: myapp:v1.2.3
secrets:
  db_password: supersecret123
  api_key: abc123xyz789
```

## Smart Configuration Detection

Maniplacer intelligently handles configuration files:

### Multiple Config Files
When multiple config files exist, Maniplacer prompts you to choose:
```bash
Multiple configuration files found:
  1) config.json
  2) config.yaml

Please choose which config file to use (1-2): 1
Selected: config.json
```

### Config File Priority
- **Custom path** (`-c` flag): Highest priority
- **Format preference** (`-f` flag): Uses preferred format if available
- **Auto-detection**: Finds and uses available config files
- **Interactive selection**: Prompts when multiple files exist

## Advanced Usage

### Multiple Namespaces
Organize templates and manifests by environment:

```bash
# Development environment
maniplacer add deployment service -n development -r myapp
maniplacer generate -n development -r myapp

# Staging environment  
maniplacer add deployment service -n staging -r myapp
maniplacer generate -n staging -r myapp

# Production environment
maniplacer add deployment service -n production -r myapp
maniplacer generate -n production -r myapp
```

### Complex Templates
Create sophisticated templates with loops and conditionals:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ .name }}
  namespace: {{ .namespace }}
spec:
  replicas: {{ .replicas }}
  selector:
    matchLabels:
      app: {{ .name }}
  template:
    metadata:
      labels:
        app: {{ .name }}
    spec:
      containers:
      - name: {{ .name }}
        image: {{ .image }}
        ports:
        - containerPort: {{ .port }}
        env:
        {{- range $key, $value := .env }}
        - name: {{ $key | ToUpper }}
          value: {{ $value | Quote }}
        {{- end }}
        {{- if .secrets }}
        envFrom:
        - secretRef:
            name: {{ .name }}-secrets
        {{- end }}
```

### Workflow Examples

#### Complete Development Workflow
```bash
# 1. Initialize project
maniplacer init --name my-microservice
cd my-microservice

# 2. Create repositories for different services
maniplacer new frontend
maniplacer new backend
maniplacer new database

# 3. Add templates for each service
maniplacer add deployment service -n production -r frontend
maniplacer add deployment service configmap -n production -r backend
maniplacer add deployment service secret -n production -r database

# 4. Generate manifests
maniplacer generate -n production -r frontend
maniplacer generate -n production -r backend
maniplacer generate -n production -r database

# 5. List generated manifests
maniplacer list -n production -r frontend

# 6. Clean up when needed
maniplacer prune -n production -r frontend
```

## Best Practices

1. **Use descriptive repository names** - Name repos based on service or component (e.g., `frontend`, `api`, `database`)
2. **Organize by namespaces** - Separate templates for different environments (`development`, `staging`, `production`)
3. **Version control everything** - Keep templates, configs, and important generated manifests in git
4. **Test templates** - Generate manifests in development before deploying to production
5. **Leverage template helpers** - Use built-in functions for common transformations
6. **Clean regularly** - Use `prune` to remove old manifests and `remove` to clean up unused templates

## Troubleshooting

### Common Issues

**"Current directory is not a valid Maniplacer project"**
- Ensure you're in a directory with a `.maniplacer` file
- Run `maniplacer init` if you haven't initialized the project

**"No configuration file found"**
- Create a `config.json` or `config.yaml` file in your repository directory
- Use the `-c` flag to specify a custom config file path

**"Template directory not found"**
- Add templates using `maniplacer add` before generating
- Verify the namespace and repository names are correct

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Support

- 📖 **Documentation**: Run `maniplacer docs` for interactive examples
- 🐛 **Issues**: Report bugs on [GitHub Issues](https://github.com/dantedelordran/maniplacer/issues)
- 💬 **Discussions**: Join conversations in [GitHub Discussions](https://github.com/dantedelordran/maniplacer/discussions)

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

Made with ❤️ for Kubernetes developers who love clean, organized manifest management.