# Maniplacer

A Kubernetes manifest templating tool for creating repeatable manifests from Go templates and configuration files.

## Overview

Maniplacer helps you manage Kubernetes manifests efficiently by:
- **Scaffolding** new component templates with sensible defaults
- **Templating** manifests using Go's template engine with custom functions
- **Generating** validated YAML manifests from configuration files
- **Organizing** resources by namespace and repository structure

## Project Status

- **Latest tagged release**: `1.4.0`
- **Runtime and tooling**: Go `1.25`, Cobra CLI, Kubernetes `client-go`
- **Implemented commands**: `init`, `new`, `add`, `remove`, `generate`, `list`, `prune`, `docs`, `update`, `version`, `completion`, and `apply`
- **In progress**: `deployed` is registered but currently a placeholder
- **Planned**: Config-file naming convention for automatic namespace apply

## Features

- **Project Management**: Initialize projects and create repositories with `init` and `new`
- **Template Management**: Add and remove component templates
- **Smart Generation**: Generate manifests with intelligent config detection
- **Cluster Apply**: Apply the latest generated manifests to a Kubernetes cluster
- **Multi-format Support**: Works with JSON and YAML configuration files
- **Built-in Documentation**: Local documentation server with examples
- **Timestamped Outputs**: Each generation writes to a timestamped folder
- **Cleanup Tools**: Prune manifests and remove templates easily
- **Self-updating**: Update to latest version from GitHub releases
- **Input Validation**: Repository, namespace, component, and rendered YAML validation
- **Test Coverage**: Unit tests for configuration, rendering, templates, updates, and documentation
- **Shell Completion**: Auto-completion for bash, zsh, fish, powershell
- **Dry-Run Mode**: Validate generation without writing files

## Installation

### Installer

Download the installer so you can inspect it before running it:

```bash
curl -fsSLo /tmp/maniplacer-installer.sh \
  https://raw.githubusercontent.com/dantedelordran/maniplacer/main/installer.sh
bash /tmp/maniplacer-installer.sh
```

The installer supports Linux and macOS on AMD64/ARM64, plus Windows AMD64 from
Git Bash. It verifies the release asset's SHA-256 digest and installs to
`~/.local/bin` by default. Set `INSTALL_DIR` to choose another destination or
`MANIPLACER_NO_MODIFY_PATH=1` to leave shell startup files unchanged.

It requires `curl`, a SHA-256 tool (`sha256sum`, `shasum`, or `openssl`), and one
JSON parser (`jq`, `python3`, or Perl with `JSON::PP`).

### Via Source
```bash
git clone https://github.com/dantedelordran/maniplacer.git
cd maniplacer
make build
make install
```

### Build with Specific Version
```bash
VERSION=1.4.0 make build
# or
go build -ldflags "-X github.com/dantedelordran/maniplacer/internal/utils.Version=1.4.0" -o dist/maniplacer ./cmd
```

### Docker
```bash
docker build --build-arg VERSION=1.4.0 -t maniplacer:latest .
docker run --rm -v $(pwd):/workspace maniplacer:latest version
```

## Quick Start

This flow uses the generated `config.json`; it does not create a second config
file or trigger an interactive format choice.

### 1. Initialize a Project

```bash
maniplacer init my-k8s-project
cd my-k8s-project
```

When prompted to create a repository during `init`, answer `n`; the next step
creates it explicitly.

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
# - httproute     (HTTP routing rules)
# - secret        (secure storage for sensitive data)
# - configmap     (configuration key-value pairs)
# - hpa           (Horizontal Pod Autoscaler)
# - hcpolicy      (Health Check Policy)
```

### 4. Edit Configuration

`maniplacer new` creates `myapp/config.json` with the values used by the bundled
templates. Update the application-specific values before generating:

```json
{
  "name": "myapp",
  "namespace": "production",
  "image": "ghcr.io/example/myapp:1.2.3",
  "replicas": 3,
  "containerPort": 8080,
  "servicePort": 80,
  "configValue": "production"
}
```

Keep the other generated keys when adding `hpa`, `httproute`, or `hcpolicy`.

### 5. Generate Manifests

```bash
# Generate manifests from templates
maniplacer generate -n production -r myapp

# Use custom config file
maniplacer generate -c custom-config.json -n production -r myapp

# Validate without writing files
maniplacer generate --dry-run -n production -r myapp
```

### 6. Apply Latest Generated Manifests

```bash
# Apply latest generated manifests to the cluster configured by ~/.kube/config
maniplacer apply myapp -n production
```

### 7. List Generated Manifests

```bash
# List generated versions in a namespace
maniplacer list -n production -r myapp

# List manifests in default namespace
maniplacer list -r myapp
```

## Project Structure

```
my-k8s-project/
├── .maniplacer              # Project marker file
├── myapp/                   # Repository directory
│   ├── config.json          # Configuration values
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
kind: ConfigMap
metadata:
  name: {{ .name | Quote }}
  namespace: {{ .namespace | Quote }}
data:
  APP_ENV: {{ .configValue | Quote }}
```

### Built-in Functions
- **`Base64`** - Encode strings to Base64
- **`ToUpper`** - Convert to uppercase
- **`ToLower`** - Convert to lowercase
- **`Quote`** - Wrap in quotes

`Base64` is encoding, not encryption. Do not commit credentials in configuration
files or generated Secret manifests; use uncommitted input or an external secret
manager.

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
maniplacer init my-k8s-project

# Arguments:
# [project-name]    Optional project directory name
```

During initialization:
- Creates project root and marks it as a valid Maniplacer project
- Optionally creates a repository with templates/, manifests/, and config.json

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

# Validate without writing files
maniplacer generate --dry-run -r myrepo

# Allow templates to reference keys that are absent from the config
maniplacer generate --strict=false -r myrepo

# Available options:
# -n, --namespace   Template namespace (default: "default")
# -f, --format      Config format: json, yaml, yml (auto-detected if not specified)
# -r, --repo        Repository name (required)
# -c, --config      Custom path to config file (overrides default config file detection)
# --dry-run         Validate generation without writing files
# --strict          Fail on missing config keys and invalid YAML output (default: true)
```

By default generation is **strict**: a template that references a key missing from
the config file is an error, and rendered output must parse as valid YAML. This
prevents the literal string `<no value>` from being written into a manifest and
later applied to a cluster. Pass `--strict=false` for the previous lenient
behaviour.

### `maniplacer apply`
Apply the latest generated manifest version for a repository to the Kubernetes cluster configured by `~/.kube/config`.

```bash
# Apply latest generated manifests in the default namespace
maniplacer apply myrepo

# Apply latest generated manifests in a specific namespace
maniplacer apply myrepo -n production

# Available options:
# -n, --namespace   Namespace to apply resources (default: "default")
```

The command reads from `<repo>/manifests/<namespace>/<latest-timestamp>/`, prompts before creating a missing namespace, and uses server-side apply with field manager `maniplacer`.

### `maniplacer list`
Display generated manifest versions in a specific namespace and repository.

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

The update process fetches the latest release, compares semantic versions,
verifies the selected binary's GitHub-published SHA-256 digest, checks its
embedded version, and atomically replaces the executable. Self-update supports
Linux and macOS; Windows users should rerun the installer.

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

Access documentation at: `http://127.0.0.1:8000/docs/`

### `maniplacer version`
Display the current version of Maniplacer.

```bash
maniplacer version
```

### `maniplacer completion`
Generate shell completion scripts for bash, zsh, fish, or powershell.

```bash
# Bash
maniplacer completion bash > /etc/bash_completion.d/maniplacer

# Zsh
maniplacer completion zsh > "${fpath[1]}/_maniplacer"

# Fish
maniplacer completion fish > ~/.config/fish/completions/maniplacer.fish

# PowerShell
maniplacer completion powershell > $PROFILE
```

### `maniplacer deployed`
Registered as a future status command. The current implementation is a placeholder and does not list pods yet.

## Configuration Formats

### JSON Configuration
```json
{
  "name": "myapp",
  "namespace": "production",
  "replicas": 3,
  "image": "myapp:v1.2.3",
  "containerPort": 8080,
  "servicePort": 80,
  "configValue": "production"
}
```

### YAML Configuration
```yaml
name: myapp
namespace: production
replicas: 3
image: myapp:v1.2.3
containerPort: 8080
servicePort: 80
configValue: production
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
maniplacer init my-microservice
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

### Debug Mode
Enable debug logging for troubleshooting:

```bash
MANIPLACER_DEBUG=true maniplacer generate -r myapp
```

## Development

### Build from Source

```bash
# Clone repository
git clone https://github.com/dantedelordran/maniplacer.git
cd maniplacer

# Build with automatic version from git
make build

# Build with specific version
VERSION=1.4.0 make build

# Build for all platforms
make build-all

# Create release archive
make release
```

### Publishing a Release

Releases are built from version tags after `develop` has been merged into
`main`. The release workflow runs tests, builds and verifies every supported
binary, creates checksums and provenance attestations, and publishes the GitHub
release.

#### 1. Push `develop`

```bash
git switch develop
git push origin develop
```

This starts the CI workflow for `develop`.

#### 2. Open and merge a pull request

```bash
gh pr create \
  --base main \
  --head develop \
  --title "Release preparation for 1.5.0" \
  --body "Prepare Maniplacer 1.5.0 for release."
```

Wait for all CI checks to pass, review the pull request, and merge it without
deleting the long-lived `develop` branch:

```bash
gh pr merge --merge
```

You can also create and merge the pull request from the GitHub website.

#### 3. Update local `main`

```bash
git switch main
git pull --ff-only origin main
git log --oneline -5
```

Confirm that the merged changes appear in the log before creating the tag.

#### 4. Create and push the version tag

Use semantic versioning and follow the repository's existing tag style without
a `v` prefix:

```bash
git tag -a 1.5.0 -m "Maniplacer 1.5.0"
git push origin 1.5.0
```

Pushing the tag starts `.github/workflows/release.yml`. The tagged commit is
what gets built, so always create the tag from the updated `main` branch.

#### 5. Monitor the release

```bash
gh run list --workflow=release.yml
gh run watch
gh release view 1.5.0
```

Do not run `gh release create` manually; the release workflow creates the
release and uploads its assets.

### Run Tests

```bash
# Run all tests
make test

# Run with coverage
go test ./... -cover

# Run specific package tests
go test ./internal/utils/...
go test ./internal/templates/...
go test ./internal/cli/...
```

### Run Linters

```bash
# Run all linters
make lint

# Or directly
go vet ./...

# Optional, if installed
staticcheck ./...
```

### Project Structure

```
maniplacer/
├── cmd/                      # Application entry point
│   └── main.go
├── internal/                 # Internal packages
│   ├── cli/                  # CLI commands
│   ├── templates/            # Template definitions and functions
│   └── utils/                # Utility functions and validation
├── dist/                     # Build artifacts
├── Makefile                  # Build automation
├── go.mod                    # Go module definition
└── README.md                 # This file
```

## Best Practices

1. **Use descriptive repository names** - Name repos based on service or component (e.g., `frontend`, `api`, `database`)
2. **Organize by namespaces** - Separate templates for different environments (`development`, `staging`, `production`)
3. **Protect secrets** - Never commit plaintext credentials or generated Secret manifests
4. **Test templates** - Generate manifests in development before deploying to production
5. **Leverage template helpers** - Use built-in functions for common transformations
6. **Clean regularly** - Use `prune` to remove old manifests and `remove` to clean up unused templates
7. **Use dry-run first** - Validate templates with `--dry-run` before generating files
8. **Validate names** - Repository and namespace names follow Kubernetes naming conventions

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

**"Invalid repository name"**
- Repository names must follow Kubernetes naming conventions
- Use lowercase alphanumeric characters and hyphens only
- Must start and end with alphanumeric character
- Example: `my-app` (valid), `My_App` (invalid)

**"Invalid namespace"**
- Namespace names follow Kubernetes DNS-1123 label standards
- Cannot use reserved namespaces: `kube-system`, `kube-public`, `kube-node-lease`
- Example: `production` (valid), `KUBE-SYSTEM` (invalid)

### Enable Debug Logging

```bash
MANIPLACER_DEBUG=true maniplacer <command>
```

This shows detailed logs for troubleshooting.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

### Development Workflow

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests (`make test`)
5. Run linters (`make lint`)
6. Commit your changes (`git commit -m 'Add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

Built for Kubernetes developers who love clean, organized manifest management.
