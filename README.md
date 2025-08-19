# Maniplacer

A CLI tool for generating K8s yaml files

## Features

- Scaffold new projects with init
- Add, list, and remove templates (add, list, remove)
- Generate manifests from templates and a JSON config (generate)
- Update the CLI easily via update
- Supports namespaces for templates and manifests
- Template helpers: Base64, ToUpper, ToLower, and more

## How to install

### Via script

```bash
curl -fsSL https://raw.github.com/dantedelordran/maniplacer/main/installer.sh | bash
```

### Via source

```bash
git clone https://github.com/dantedelordran/maniplacer.git
cd maniplacer
make build
```

## Quickstart

### Prepare a sandbox (or install directly in your machine)

```bash
docker build -t maniplacer:latest .
docker run -it --name sandbox maniplacer:latest
```

Creates the following structure

```bash
my-project/
├─ templates/
├─ manifests/
└─ config.json
└─ .maniplacer
```

## Update to latest version

```bash
maniplacer update
# or force update
maniplacer update -f
```