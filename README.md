# Maniplacer rebirth

A new reimaginated version of Maniplacer, a CLI tool for generating K8s yaml files

## How to install

```bash
curl -fsSL https://raw.github.com/dantedelordran/maniplacer/main/installer.sh | bash
```

## Want to try? Prepare a sandbox

```bash
docker build -t maniplacer:latest .
docker run -it --name sandbox maniplacer:latest
```
