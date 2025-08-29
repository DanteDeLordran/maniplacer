
## Features

- [x] Base64 encription

## Commands

- [x] Version: Shows current version
- [x] Init: Inits a new project given a name and optional config type flag (default JSON)
- [x] Add: Adds a new empty component template
- [x] Remove: Removes a component template by namespace
- [x] Generate: Generates a new manifest
- [x] List: Lists all the manifests in order filtered by namespace
- [x] Update: Self update to latest version
- [x] Prune: Removes every generated manifest for every namespace (ask for confirmation)\
- [x] New: Creates a new repo inside project
- [ ] Apply: Performs a kubectl apply to the latest manifest (ask for confirmation)

## Enhancements

- [x] YAML support
- [x] Validate existing Maniplacer project before initing another
- [x] Dockerfile for sandboxing
- [x] After creating a component, ask if you want to override if already exists
- [x] Change project structure
- [x] Remake long desc
- [ ] Use config file naming convenction for automatic namespace apply