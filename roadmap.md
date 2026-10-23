# Maniplacer Roadmap

## Goal and implementation order

Complete a reliable `generate → apply → deployed` workflow. Implement the following
milestones as focused PRs, in order:

1. Atomic generation and unique manifest versions.
2. Shared namespace resolution and config-file naming conventions.
3. Shared Kubernetes clients and resource ownership labels.
4. The `deployed` command.
5. Apply improvements and cluster integration tests.

Unchecked items below are planned work. Command examples in those sections describe
the proposed interface, not functionality that is already available.

## Implemented baseline

### Features and tooling

- [x] Base64 encoding (not encryption).
- [x] JSON and YAML configuration support.
- [x] Strict template rendering: missing config keys and invalid YAML cause errors.
- [x] Generation dry-run without writing files.
- [x] Repository, namespace, component, and path validation.
- [x] Validate an existing Maniplacer project before initializing another.
- [x] Dockerfile for sandboxing.
- [x] Prompt before overwriting an existing component template.
- [x] Repository- and namespace-based project structure.
- [x] Expanded command descriptions and documentation.
- [x] Unit tests and cross-platform CI with race detection and static analysis.

### Commands

- [x] `version`: Show the current version.
- [x] `init`: Initialize a project with an optional name; repository scaffolding currently creates JSON config.
- [x] `new`: Create a repository inside a project.
- [x] `add`: Scaffold component templates for a repository and namespace.
- [x] `remove`: Remove component templates for a repository and namespace.
- [x] `generate`: Render manifests into a timestamped version directory.
- [x] `list`: List generated versions for a repository and namespace.
- [x] `update`: Self-update to the latest release.
- [x] `prune`: Remove generated versions for one repository and namespace, with confirmation.
- [x] `apply`: Apply the latest generated version using the Kubernetes API and server-side apply; prompt before creating a missing namespace.
- [x] `docs`: Serve local documentation.
- [x] `completion`: Generate shell completion scripts.
- [ ] `deployed`: List pods deployed by Maniplacer; currently a registered placeholder.

## 1. Atomic generation and unique manifest versions

**Problem:** `generate` creates its output directory before processing all templates.
A failed run can leave an incomplete version that `apply` subsequently selects as
the latest. Timestamps have second-level precision, so quick or concurrent runs can
also share an output directory.

**Starting points:** `internal/cli/generate.go` (`processTemplate` and generation
loop), `internal/cli/apply.go` (`getLatestManifest`),
`internal/cli/generate_render_test.go`.

- [ ] Render and validate the entire generation in a staging directory.
- [ ] Keep staging directories outside the directories searched by `apply`, on the same filesystem as the final destination.
- [ ] Publish the version with a directory rename only after every template succeeds.
- [ ] Clean up staging files on failure or cancellation.
- [ ] Give each generation a collision-resistant, chronologically sortable version identifier.
- [ ] Ignore staging/non-version entries when selecting the latest version, while retaining support for existing timestamped versions.
- [ ] Preserve the existing behavior of `generate --dry-run`: no output files or directories.

**Acceptance criteria**

- [ ] A failure in any template leaves the previous complete version as the latest available version.
- [ ] Two generations within the same second cannot overwrite or merge their outputs.
- [ ] Successful generation publishes every expected manifest together.
- [ ] Tests cover partial failure, version collisions, latest-version selection, and dry-run behavior.

## 2. Namespace resolution and config-file naming conventions

**Problem:** `--namespace` currently selects template/output directories, while
`.namespace` in the config controls the namespace in bundled manifests. These can
disagree. Generation also selects the template directory before loading config.

**Starting points:** `internal/cli/generate.go` (`FindConfigFile`, config loading,
and template-directory selection), `internal/cli/common.go` (`resolveTarget`),
`internal/cli/apply.go`, `internal/cli/config_test.go`.

### Proposed convention

```text
myapp/
├── config.yaml
├── config.staging.yaml
└── config.production.yaml
```

Support `config.<namespace>.json`, `config.<namespace>.yaml`, and
`config.<namespace>.yml`, alongside existing generic config filenames.

### Resolution rules

- [ ] Define config selection explicitly: `--config` selects a file; an explicit namespace can prefer a matching namespaced config before falling back to a generic config.
- [ ] Preserve format selection and existing generic JSON/YAML config support.
- [ ] Require an explicit choice when multiple environment configs are ambiguous; never silently choose an arbitrary environment.
- [ ] Resolve namespace in this order: explicitly supplied `--namespace`, filename suffix, config `namespace`, then `default`.
- [ ] Use `cmd.Flags().Changed("namespace")` to distinguish an explicit flag from its default value.
- [ ] Treat an explicit namespace flag as a deliberate override; otherwise, reject conflicting filename/config namespaces with a useful error.
- [ ] Validate the resolved namespace and set the template's `.namespace` to that value.
- [ ] Resolve config and namespace before choosing template and output directories.
- [ ] Extract shared resolution logic for `generate` and `apply`.
- [ ] Add `--config` / `-c` to `apply` so it can select the corresponding generated namespace without regenerating manifests.
- [ ] Reject rendered namespaced resources that conflict with the resolved target; keep cluster-scoped resources without a namespace.
- [ ] Check namespace consistency again before apply so an edited config cannot silently redirect existing manifests.

### Proposed workflow

```bash
maniplacer generate -r myapp -c config.production.yaml
maniplacer apply myapp -c config.production.yaml

# Deliberate namespace override
maniplacer generate -r myapp -c config.production.yaml -n staging
maniplacer apply myapp -c config.production.yaml -n staging
```

**Acceptance criteria**

- [ ] Template selection, output directory, rendered namespace, and apply target agree.
- [ ] Tests cover all supported extensions, generic configs, filename inference, explicit overrides, invalid namespaces, ambiguous configs, and conflicting values.
- [ ] Cluster-scoped resources remain valid and receive no injected namespace.
- [ ] README and embedded docs explain config selection, precedence, and migration behavior.

## 3. Shared Kubernetes clients and ownership labels

**Problem:** Kubernetes clients are package-level globals in `apply.go`, and applied
workloads have no consistent labels that `deployed` can use to identify their pods.

**Starting points:** `internal/cli/apply.go` (`initKubeClients`, `createResources`,
and `ensureNamespace`), `internal/templates/files/deployment.tmpl`.

- [ ] Extract client construction and shared Kubernetes operations into a small package such as `internal/kube`.
- [ ] Inject `kubernetes.Interface` and `dynamic.Interface` instead of relying on concrete global clients.
- [ ] Pass repository identity into the apply/labeling step.
- [ ] Add ownership labels before server-side apply, covering custom manifests as well as bundled templates:

  ```yaml
  app.kubernetes.io/managed-by: maniplacer
  maniplacer.io/repository: myapp
  ```

- [ ] Add labels to resource metadata and the pod-template paths below:

  | Resource | Pod label location |
  | --- | --- |
  | Deployment, StatefulSet, DaemonSet, Job | `spec.template.metadata.labels` |
  | CronJob | `spec.jobTemplate.spec.template.metadata.labels` |
  | Pod | `metadata.labels` |

- [ ] Preserve unrelated existing labels and leave workload selectors unchanged.
- [ ] Document that previously applied workloads need reapplying to receive ownership labels; existing pods become discoverable when their workload replaces them with labeled pods.

**Acceptance criteria**

- [ ] Pods created by supported workloads inherit the ownership labels.
- [ ] Tests cover all supported pod-template paths and preservation of existing labels/selectors.
- [ ] Apply and pod-listing code can use injected fake clients in unit tests.

## 4. Implement `deployed`

**Dependency:** Milestone 3 provides client reuse and ownership labels.

**Starting point:** `internal/cli/deployed.go`.

### Proposed interface

```bash
maniplacer deployed -r myapp -n production
maniplacer deployed --all-namespaces
maniplacer deployed -r myapp -o json
```

- [ ] Replace the placeholder with a `RunE` implementation that lists pods through the typed Kubernetes client.
- [ ] Always filter by `app.kubernetes.io/managed-by=maniplacer`.
- [ ] Add optional repository filtering with `--repo` / `-r`; omitting it lists all Maniplacer-managed repositories in scope.
- [ ] Support `--namespace` / `-n` (default: `default`) and `--all-namespaces` / `-A`; reject an explicit namespace combined with all-namespaces.
- [ ] Use server-side label selectors rather than fetching unrelated pods and filtering locally.
- [ ] Provide deterministic table output with namespace, name, ready containers, phase, restarts, and age.
- [ ] Support `--output` / `-o json` for scripts, keeping diagnostic messages separate from JSON output.
- [ ] Return a helpful empty-result message and actionable errors for RBAC denial or cluster connection failures.
- [ ] Use the command context for cancellation.

Example table:

```text
NAMESPACE   NAME                 READY   PHASE     RESTARTS   AGE
production  myapp-7c845f9-x2abc   1/1     Running   0          12m
```

**Acceptance criteria**

- [ ] Unrelated pods never appear in results.
- [ ] Fake-client tests cover repository/namespace filtering, all-namespaces, empty results, API errors, readiness/restarts, and JSON output.
- [ ] Command help, README, and embedded docs describe the completed command and label migration requirement.

## 5. Apply improvements and integration coverage

**Starting points:** `internal/cli/apply.go`, `internal/cli/generate.go`,
`internal/cli/root.go`, `.github/workflows/ci.yml`.

- [ ] Decode YAML streams document by document in generation validation and apply; process every document separated by `---` and skip empty documents.
- [ ] Reject invalid later documents during generation rather than validating only the first document.
- [ ] Honor standard kubeconfig loading, including `KUBECONFIG`, and add `--kubeconfig` and `--context` to cluster commands.
- [ ] Replace `context.TODO()` with the Cobra command context and support bounded request timeouts.
- [ ] Show the target context, namespace, and manifest version, then ask for apply confirmation as promised by the original roadmap.
- [ ] Add `--yes` for noninteractive confirmation handling, including missing-namespace creation, and document its behavior.
- [ ] Add `apply --dry-run=server` for Kubernetes API validation; ensure this mode creates or changes no resources, including namespaces.
- [ ] Document that generation validation checks rendering/YAML, whereas server dry-run validates against the cluster.
- [ ] Add unit tests for multi-document decoding, namespace handling, confirmation behavior, and error propagation.
- [ ] Add a kind-based integration test for `generate → apply → deployed`, including pod-label propagation and exclusion of unrelated pods.
- [ ] Verify real server-side apply, multi-document manifests, cluster-scoped resources, and dry-run behavior in integration tests; fake clients do not reproduce API-server behavior.
- [ ] Run cluster integration tests in an appropriate Linux CI job alongside existing unit-test jobs.

**Acceptance criteria**

- [ ] Every document in a multi-document manifest is validated and applied.
- [ ] Cluster commands honor selected kubeconfig/context and cancellation.
- [ ] Apply confirmation works interactively and can be explicitly bypassed for automation.
- [ ] Server dry-run leaves the cluster unchanged and reports missing namespaces clearly.
- [ ] The end-to-end workflow passes against a disposable kind cluster.

## 6. Follow-up features

After the core workflow is complete, prioritize:

- [ ] **Config overlays:** Merge shared defaults with environment-specific values; define precedence and map/list merge behavior.
- [ ] **Diff:** Preview differences between generated resources and live cluster state.
- [ ] **Version selection:** Add `apply --version <id>` to reapply a selected generated version, and expose identifiers clearly in `list`.
- [ ] **Rollout waiting:** Add `apply --wait --timeout <duration>` to report workload readiness and return an error on timeout or rollout failure.
- [ ] **Shared templates:** Reuse templates across namespaces/environments to reduce duplication under `templates/<namespace>`.
- [ ] **Scaffolding format:** Add the previously described config-format flag to `init` and `new`, retaining JSON as the default.

## Validation and next-session handoff

The existing `go test ./...` suite passed during the review that produced this
roadmap. That baseline does not cover the planned behavior above.

For implementation work, run the relevant focused tests, then the project checks:

```bash
go test ./...
go test -race ./...
go vet ./...
```

- Keep feature checkboxes unchecked until their acceptance criteria are met.
- Update command help, `README.md`, and `internal/cli/docs.html` alongside user-facing changes.
- Preserve existing generic configs and timestamped manifest history when introducing new conventions.
- Start the next session with **milestone 1**. Suggested prompt:

  > Read roadmap.md and implement milestone 1: atomic generation and unique manifest versions. Inspect the existing generation and latest-version selection code, add regression tests for failed and colliding generations, run the relevant checks, and update the completed roadmap checkboxes.
