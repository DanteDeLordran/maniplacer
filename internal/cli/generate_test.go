package cli

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

const configMapTemplate = "kind: ConfigMap\nmetadata:\n  name: {{ .name }}\n  namespace: {{ .namespace }}\n"

// runCLI executes the real command tree with flags back at their defaults, since
// cobra keeps flag values and their Changed state between executions.
func runCLI(t *testing.T, args ...string) error {
	t.Helper()

	for cmd, names := range map[*cobra.Command][]string{
		generateCmd: {"repo", "namespace", "format", "config", "dry-run", "strict"},
		applyCmd:    {"namespace"},
	} {
		for _, name := range names {
			flag := cmd.Flags().Lookup(name)
			if err := flag.Value.Set(flag.DefValue); err != nil {
				t.Fatalf("could not reset flag %q: %v", name, err)
			}
			flag.Changed = false
		}
	}

	rootCmd.SetArgs(args)
	t.Cleanup(func() { rootCmd.SetArgs(nil) })

	return rootCmd.ExecuteContext(context.Background())
}

// newTestRepo creates the repo "myapp" in a temp project, with the given JSON
// config and a templates directory for namespace.
func newTestRepo(t *testing.T, namespace, config string) target {
	t.Helper()

	repo := target{Root: chdirToProject(t), Repo: "myapp", Namespace: namespace}

	if err := os.MkdirAll(repo.TemplatesDir(), 0755); err != nil {
		t.Fatalf("could not create templates directory: %v", err)
	}
	writeTemplate(t, repo.RepoPath(), "config.json", config)

	return repo
}

// versions returns the names of the published versions for repo.
func versions(t *testing.T, repo target) []string {
	t.Helper()

	entries, err := os.ReadDir(repo.ManifestsDir())
	if err != nil {
		t.Fatalf("could not read manifests directory: %v", err)
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	return names
}

func TestGenerateFailureKeepsPreviousVersionLatest(t *testing.T) {
	repo := newTestRepo(t, "default", `{"name": "myapp", "namespace": "default"}`)
	writeTemplate(t, repo.TemplatesDir(), "a-good.yaml", configMapTemplate)

	if err := runCLI(t, "generate", "-r", "myapp"); err != nil {
		t.Fatalf("generate error = %v", err)
	}
	published := versions(t, repo)

	// a-good.yaml still renders, so this used to publish a partial version
	writeTemplate(t, repo.TemplatesDir(), "z-broken.yaml", "name: {{ .missing }}\n")
	if err := runCLI(t, "generate", "-r", "myapp"); err == nil {
		t.Fatal("generate error = nil, want an error for the broken template")
	}

	if got := versions(t, repo); !slices.Equal(got, published) {
		t.Errorf("versions = %v, want only %v", got, published)
	}

	latest, err := getLatestManifest(repo.ManifestsDir())
	if err != nil {
		t.Fatalf("getLatestManifest() error = %v", err)
	}
	if want := filepath.Join(repo.ManifestsDir(), published[0]); latest != want {
		t.Errorf("latest = %q, want %q", latest, want)
	}

	// The only thing beside the namespace folder would be leftover staging
	entries, err := os.ReadDir(filepath.Dir(repo.ManifestsDir()))
	if err != nil {
		t.Fatalf("could not read manifests root: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("manifests root holds %d entries, want only the namespace folder", len(entries))
	}
}

func TestGenerateRapidRunsPublishSeparateVersions(t *testing.T) {
	repo := newTestRepo(t, "default", `{"name": "myapp", "namespace": "default"}`)
	writeTemplate(t, repo.TemplatesDir(), "configmap.yaml", configMapTemplate)

	const runs = 3
	for i := range runs {
		if err := runCLI(t, "generate", "-r", "myapp"); err != nil {
			t.Fatalf("generate error = %v", err)
		}
		// Well within one second, but past the coarser clock tick on Windows
		if i < runs-1 {
			time.Sleep(20 * time.Millisecond)
		}
	}

	published := versions(t, repo)
	if len(published) != runs {
		t.Fatalf("versions = %v, want %d of them", published, runs)
	}

	for _, version := range published {
		if _, err := os.Stat(filepath.Join(repo.ManifestsDir(), version, "configmap.yaml")); err != nil {
			t.Errorf("version %s is incomplete: %v", version, err)
		}
	}
}

func TestGenerateDryRunPublishesNothing(t *testing.T) {
	repo := newTestRepo(t, "default", `{"name": "myapp", "namespace": "default"}`)
	writeTemplate(t, repo.TemplatesDir(), "configmap.yaml", configMapTemplate)

	if err := runCLI(t, "generate", "-r", "myapp", "--dry-run"); err != nil {
		t.Fatalf("generate error = %v", err)
	}

	if _, err := os.Stat(filepath.Dir(repo.ManifestsDir())); !os.IsNotExist(err) {
		t.Errorf("dry-run created the manifests root, stat error = %v", err)
	}
}

func TestGenerateNamespaceResolution(t *testing.T) {
	tests := []struct {
		name       string
		namespace  string
		config     string
		args       []string
		wantRender string
		wantErr    bool
	}{
		{
			name:       "explicit flag overrides config",
			namespace:  "production",
			config:     `{"name": "myapp", "namespace": "default"}`,
			args:       []string{"-n", "production"},
			wantRender: "namespace: production",
		},
		{
			name:       "explicit default overrides config",
			namespace:  "default",
			config:     `{"name": "myapp", "namespace": "production"}`,
			args:       []string{"-n", "default"},
			wantRender: "namespace: default",
		},
		{
			name:       "config agrees with default target",
			namespace:  "default",
			config:     `{"name": "myapp", "namespace": "default"}`,
			wantRender: "namespace: default",
		},
		{
			name:       "config without a namespace uses the target",
			namespace:  "default",
			config:     `{"name": "myapp"}`,
			wantRender: "namespace: default",
		},
		{
			name:      "config disagrees with default target",
			namespace: "default",
			config:    `{"name": "myapp", "namespace": "production"}`,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newTestRepo(t, tt.namespace, tt.config)
			writeTemplate(t, repo.TemplatesDir(), "configmap.yaml", configMapTemplate)

			err := runCLI(t, append([]string{"generate", "-r", "myapp"}, tt.args...)...)
			if tt.wantErr {
				if err == nil {
					t.Fatal("generate error = nil, want a namespace conflict")
				}
				if _, statErr := os.Stat(filepath.Dir(repo.ManifestsDir())); !os.IsNotExist(statErr) {
					t.Errorf("generate wrote output despite failing, stat error = %v", statErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("generate error = %v", err)
			}

			latest, err := getLatestManifest(repo.ManifestsDir())
			if err != nil {
				t.Fatalf("getLatestManifest() error = %v", err)
			}
			data, err := os.ReadFile(filepath.Join(latest, "configmap.yaml"))
			if err != nil {
				t.Fatalf("could not read generated manifest: %v", err)
			}
			if !strings.Contains(string(data), tt.wantRender) {
				t.Errorf("rendered manifest = %q, want it to contain %q", data, tt.wantRender)
			}
		})
	}
}

func TestGenerateRejectsHardcodedNamespaceConflict(t *testing.T) {
	repo := newTestRepo(t, "production", `{"name": "myapp"}`)
	writeTemplate(t, repo.TemplatesDir(), "configmap.yaml", "kind: ConfigMap\nmetadata:\n  namespace: other\n")

	if err := runCLI(t, "generate", "-r", "myapp", "-n", "production"); err == nil {
		t.Fatal("generate error = nil, want a namespace conflict")
	}

	if got := versions(t, repo); len(got) != 0 {
		t.Errorf("versions = %v, want none", got)
	}
}
