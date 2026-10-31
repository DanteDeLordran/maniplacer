package cli

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"testing"
)

// fakeCluster serves the part of the Kubernetes API that apply talks to and
// records each server-side apply it receives. Tests reach it only through
// KUBECONFIG, so every test here also proves that variable is honored.
type fakeCluster struct {
	mu      sync.Mutex
	applied []string
}

func (c *fakeCluster) appliedPaths() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return slices.Clone(c.applied)
}

func (c *fakeCluster) serverSideApply(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Content-Type"); got != "application/apply-patch+yaml" {
			t.Errorf("Content-Type = %q, want a server-side apply", got)
		}
		if got := r.URL.Query().Get("fieldManager"); got != "maniplacer" {
			t.Errorf("fieldManager = %q, want %q", got, "maniplacer")
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("could not read apply body: %v", err)
		}

		c.mu.Lock()
		c.applied = append(c.applied, r.URL.Path)
		c.mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}
}

// startFakeCluster points KUBECONFIG at a fake API server, and HOME at an empty
// directory so a real ~/.kube/config can never be used.
func startFakeCluster(t *testing.T) *fakeCluster {
	t.Helper()

	cluster := &fakeCluster{}

	serveJSON := func(body string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, body)
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api", serveJSON(`{"kind": "APIVersions", "versions": ["v1"]}`))
	mux.HandleFunc("GET /apis", serveJSON(`{"kind": "APIGroupList", "apiVersion": "v1", "groups": []}`))
	mux.HandleFunc("GET /api/v1", serveJSON(`{
		"kind": "APIResourceList",
		"groupVersion": "v1",
		"resources": [
			{"name": "configmaps", "namespaced": true, "kind": "ConfigMap", "verbs": ["patch"]},
			{"name": "namespaces", "namespaced": false, "kind": "Namespace", "verbs": ["patch"]}
		]
	}`))
	mux.HandleFunc("GET /api/v1/namespaces/{name}", func(w http.ResponseWriter, r *http.Request) {
		serveJSON(fmt.Sprintf(`{"kind": "Namespace", "apiVersion": "v1", "metadata": {"name": %q}}`, r.PathValue("name")))(w, r)
	})
	mux.HandleFunc("PATCH /api/v1/namespaces/{name}", cluster.serverSideApply(t))
	mux.HandleFunc("PATCH /api/v1/namespaces/{namespace}/configmaps/{name}", cluster.serverSideApply(t))

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	kubeconfig := filepath.Join(t.TempDir(), "kubeconfig")
	content := fmt.Sprintf(`apiVersion: v1
kind: Config
clusters: [{name: test, cluster: {server: %q}}]
users: [{name: test, user: {}}]
contexts: [{name: test, context: {cluster: test, user: test}}]
current-context: test
`, server.URL)
	if err := os.WriteFile(kubeconfig, []byte(content), 0600); err != nil {
		t.Fatalf("could not write kubeconfig: %v", err)
	}

	t.Setenv("KUBECONFIG", kubeconfig)
	t.Setenv("HOME", t.TempDir())

	return cluster
}

// writeVersion publishes a generated version containing one manifest.
func writeVersion(t *testing.T, repo target, version, manifest string) {
	t.Helper()

	dir := filepath.Join(repo.ManifestsDir(), version)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("could not create version directory: %v", err)
	}
	writeTemplate(t, dir, "manifest.yaml", manifest)
}

func newApplyRepo(t *testing.T, manifest string) {
	t.Helper()

	repo := target{Root: chdirToProject(t), Repo: "myapp", Namespace: "production"}
	writeVersion(t, repo, "2026-01-01_00-00-00", manifest)
}

func TestApplyAppliesEveryDocument(t *testing.T) {
	cluster := startFakeCluster(t)
	newApplyRepo(t, `apiVersion: v1
kind: ConfigMap
metadata:
  name: first
  namespace: production
---
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: second
---
apiVersion: v1
kind: Namespace
metadata:
  name: cluster-scoped
`)

	if err := runCLI(t, "apply", "myapp", "-n", "production"); err != nil {
		t.Fatalf("apply error = %v", err)
	}

	want := []string{
		"/api/v1/namespaces/production/configmaps/first",
		"/api/v1/namespaces/production/configmaps/second",
		"/api/v1/namespaces/cluster-scoped",
	}
	if got := cluster.appliedPaths(); !slices.Equal(got, want) {
		t.Errorf("applied = %v, want %v", got, want)
	}
}

func TestApplyRejectsNamespaceMismatch(t *testing.T) {
	cluster := startFakeCluster(t)
	newApplyRepo(t, `apiVersion: v1
kind: ConfigMap
metadata:
  name: wrong
  namespace: other
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: right
  namespace: production
`)

	if err := runCLI(t, "apply", "myapp", "-n", "production"); err == nil {
		t.Fatal("apply error = nil, want a namespace mismatch")
	}

	want := []string{"/api/v1/namespaces/production/configmaps/right"}
	if got := cluster.appliedPaths(); !slices.Equal(got, want) {
		t.Errorf("applied = %v, want %v", got, want)
	}
}

func TestApplySkipsFileWithInvalidDocument(t *testing.T) {
	cluster := startFakeCluster(t)
	newApplyRepo(t, "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: first\n---\n: : bad: [unclosed\n")

	if err := runCLI(t, "apply", "myapp", "-n", "production"); err == nil {
		t.Fatal("apply error = nil, want a parse error")
	}

	if got := cluster.appliedPaths(); len(got) != 0 {
		t.Errorf("applied = %v, want nothing from a file that does not parse", got)
	}
}

func TestApplyReportsManifestWithoutKind(t *testing.T) {
	cluster := startFakeCluster(t)
	newApplyRepo(t, "metadata:\n  name: no-kind\n")

	if err := runCLI(t, "apply", "myapp", "-n", "production"); err == nil {
		t.Fatal("apply error = nil, want an error instead of a silent skip")
	}

	if got := cluster.appliedPaths(); len(got) != 0 {
		t.Errorf("applied = %v, want nothing", got)
	}
}

func TestGetLatestManifest(t *testing.T) {
	tests := []struct {
		name    string
		entries []string
		want    string
		wantErr bool
	}{
		{
			name:    "fractional version after second precision version",
			entries: []string{"2026-01-01_00-00-00", "2026-01-01_00-00-00.000000001"},
			want:    "2026-01-01_00-00-00.000000001",
		},
		{
			name:    "second precision version after fractional version",
			entries: []string{"2026-01-01_00-00-00.999999999", "2026-01-01_00-00-01"},
			want:    "2026-01-01_00-00-01",
		},
		{
			name:    "ignores directories that are not versions",
			entries: []string{"2026-01-01_00-00-00", "notes"},
			want:    "2026-01-01_00-00-00",
		},
		{
			name:    "no versions",
			entries: []string{"notes"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			for _, entry := range tt.entries {
				if err := os.Mkdir(filepath.Join(dir, entry), 0755); err != nil {
					t.Fatalf("could not create entry: %v", err)
				}
			}

			got, err := getLatestManifest(dir)
			if (err != nil) != tt.wantErr {
				t.Fatalf("getLatestManifest() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && got != filepath.Join(dir, tt.want) {
				t.Errorf("getLatestManifest() = %q, want %q", got, filepath.Join(dir, tt.want))
			}
		})
	}
}
