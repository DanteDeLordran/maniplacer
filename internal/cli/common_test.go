package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dantedelordran/maniplacer/internal/utils"
	"github.com/spf13/cobra"
)

// newTestCmd builds a command carrying the shared repo/namespace flags, with the
// given flag values already applied.
func newTestCmd(t *testing.T, flags map[string]string) *cobra.Command {
	t.Helper()

	cmd := &cobra.Command{Use: "test"}
	addRepoNamespaceFlags(cmd, "test namespace")

	for name, value := range flags {
		if err := cmd.Flags().Set(name, value); err != nil {
			t.Fatalf("could not set flag %q: %v", name, err)
		}
	}

	return cmd
}

// chdirToProject creates a valid Maniplacer project in a temp dir and makes it
// the working directory for the duration of the test.
func chdirToProject(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	if err := utils.CreateManiplacerProject(dir); err != nil {
		t.Fatalf("could not create test project: %v", err)
	}

	original, err := os.Getwd()
	if err != nil {
		t.Fatalf("could not get working directory: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(original); err != nil {
			t.Fatalf("could not restore working directory: %v", err)
		}
	})

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("could not change working directory: %v", err)
	}

	// macOS temp dirs are symlinked (/var -> /private/var); resolveTarget uses
	// os.Getwd, which reports the resolved path.
	resolved, err := os.Getwd()
	if err != nil {
		t.Fatalf("could not get working directory: %v", err)
	}

	return resolved
}

func TestResolveTarget(t *testing.T) {
	root := chdirToProject(t)

	cmd := newTestCmd(t, map[string]string{"repo": "myapp", "namespace": "staging"})

	target, err := resolveTarget(cmd)
	if err != nil {
		t.Fatalf("resolveTarget() error = %v", err)
	}

	if target.Repo != "myapp" {
		t.Errorf("Repo = %q, want %q", target.Repo, "myapp")
	}
	if target.Namespace != "staging" {
		t.Errorf("Namespace = %q, want %q", target.Namespace, "staging")
	}

	if want := filepath.Join(root, "myapp"); target.RepoPath() != want {
		t.Errorf("RepoPath() = %q, want %q", target.RepoPath(), want)
	}
	if want := filepath.Join(root, "myapp", "templates", "staging"); target.TemplatesDir() != want {
		t.Errorf("TemplatesDir() = %q, want %q", target.TemplatesDir(), want)
	}
	if want := filepath.Join(root, "myapp", "manifests", "staging"); target.ManifestsDir() != want {
		t.Errorf("ManifestsDir() = %q, want %q", target.ManifestsDir(), want)
	}
}

func TestResolveTargetDefaultsNamespace(t *testing.T) {
	chdirToProject(t)

	target, err := resolveTarget(newTestCmd(t, map[string]string{"repo": "myapp"}))
	if err != nil {
		t.Fatalf("resolveTarget() error = %v", err)
	}

	if target.Namespace != utils.DefaultNamespace {
		t.Errorf("Namespace = %q, want %q", target.Namespace, utils.DefaultNamespace)
	}
}

func TestResolveTargetRejectsInvalidInput(t *testing.T) {
	chdirToProject(t)

	tests := []struct {
		name  string
		flags map[string]string
	}{
		{"missing repo", map[string]string{}},
		{"invalid repo name", map[string]string{"repo": "My_App"}},
		{"path traversal in repo", map[string]string{"repo": "../evil"}},
		{"reserved namespace", map[string]string{"repo": "myapp", "namespace": "kube-system"}},
		{"invalid namespace", map[string]string{"repo": "myapp", "namespace": "Staging"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := resolveTarget(newTestCmd(t, tt.flags)); err == nil {
				t.Errorf("resolveTarget() error = nil, want an error")
			}
		})
	}
}

func TestResolveTargetRequiresProject(t *testing.T) {
	dir := t.TempDir()

	original, err := os.Getwd()
	if err != nil {
		t.Fatalf("could not get working directory: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(original); err != nil {
			t.Fatalf("could not restore working directory: %v", err)
		}
	})

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("could not change working directory: %v", err)
	}

	if _, err := resolveTarget(newTestCmd(t, map[string]string{"repo": "myapp"})); err == nil {
		t.Error("resolveTarget() error = nil, want an error outside a Maniplacer project")
	}
}
