
## Features

- [ ] Base64 encription
- [ ] Keyvalue

## Commands

- [x] Version: Shows current version
- [x] Init: Inits a new project given a name and optional config type flag (default JSON)
- [x] Add: Adds a new empty component template
- [x] Remove: Removes a component template by namespace
- [ ] Generate: Generates a new manifest
- [ ] List: Lists all the manifests in order filtered by namespace
- [ ] Update: Self update to latest version
- [ ] Prune: Removes every generated manifest for every namespace (ask for confirmation)
- [ ] Apply: Performs a kubectl apply to the latest manifest (ask for confirmation)

## Enhancements

- [ ] Beautify UI
- [ ] YAML support
- [ ] Validate existing Maniplacer project befor initing another
- [ ] Regenerate templates dir if was deleted
- [x] Dockerfile for sandboxing
- [ ] After creating a component, ask if you want to override if already exists