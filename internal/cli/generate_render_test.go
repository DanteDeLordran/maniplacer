package cli

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeTemplate writes a template file into a temp dir and returns its path.
func writeTemplate(t *testing.T, dir, name, content string) string {
	t.Helper()

	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("could not write template: %v", err)
	}

	return path
}

func TestProcessTemplateStrictRejectsMissingKey(t *testing.T) {
	dir := t.TempDir()
	outputDir := t.TempDir()

	templatePath := writeTemplate(t, dir, "deployment.yaml", "name: {{ .doesNotExist }}\n")
	config := map[string]any{"name": "myapp"}

	err := processTemplate(context.Background(), templatePath, outputDir, "deployment.yaml", config, false, true)
	if err == nil {
		t.Fatal("processTemplate() error = nil, want an error for the missing key")
	}

	// A failed render must not leave a manifest behind
	if _, statErr := os.Stat(filepath.Join(outputDir, "deployment.yaml")); !os.IsNotExist(statErr) {
		t.Error("processTemplate() wrote an output file despite failing")
	}
}

func TestProcessTemplateLenientAllowsMissingKey(t *testing.T) {
	dir := t.TempDir()
	outputDir := t.TempDir()

	templatePath := writeTemplate(t, dir, "deployment.yaml", "name: {{ .doesNotExist }}\n")

	err := processTemplate(context.Background(), templatePath, outputDir, "deployment.yaml", map[string]any{}, false, false)
	if err != nil {
		t.Fatalf("processTemplate() error = %v, want nil in lenient mode", err)
	}

	data, err := os.ReadFile(filepath.Join(outputDir, "deployment.yaml"))
	if err != nil {
		t.Fatalf("could not read generated manifest: %v", err)
	}
	if !strings.Contains(string(data), "<no value>") {
		t.Errorf("lenient render = %q, want it to contain %q", string(data), "<no value>")
	}
}

func TestProcessTemplateRendersConfig(t *testing.T) {
	dir := t.TempDir()
	outputDir := t.TempDir()

	templatePath := writeTemplate(t, dir, "service.yaml", "kind: Service\nmetadata:\n  name: {{ .name }}\n")
	config := map[string]any{"name": "myapp"}

	if err := processTemplate(context.Background(), templatePath, outputDir, "service.yaml", config, false, true); err != nil {
		t.Fatalf("processTemplate() error = %v", err)
	}

	data, err := os.ReadFile(filepath.Join(outputDir, "service.yaml"))
	if err != nil {
		t.Fatalf("could not read generated manifest: %v", err)
	}
	if !strings.Contains(string(data), "name: myapp") {
		t.Errorf("rendered manifest = %q, want it to contain %q", string(data), "name: myapp")
	}
}

func TestProcessTemplateStrictRejectsInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	outputDir := t.TempDir()

	templatePath := writeTemplate(t, dir, "broken.yaml", "kind: Service\n  name: {{ .name }}\n\tbad: indent\n")
	config := map[string]any{"name": "myapp"}

	err := processTemplate(context.Background(), templatePath, outputDir, "broken.yaml", config, false, true)
	if err == nil {
		t.Fatal("processTemplate() error = nil, want an error for output that is not valid YAML")
	}

	if _, statErr := os.Stat(filepath.Join(outputDir, "broken.yaml")); !os.IsNotExist(statErr) {
		t.Error("processTemplate() wrote an output file despite failing")
	}
}

func TestProcessTemplateDryRunWritesNothing(t *testing.T) {
	dir := t.TempDir()
	outputDir := t.TempDir()

	templatePath := writeTemplate(t, dir, "service.yaml", "kind: Service\nmetadata:\n  name: {{ .name }}\n")
	config := map[string]any{"name": "myapp"}

	if err := processTemplate(context.Background(), templatePath, outputDir, "service.yaml", config, true, true); err != nil {
		t.Fatalf("processTemplate() error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(outputDir, "service.yaml")); !os.IsNotExist(err) {
		t.Error("processTemplate() wrote a file in dry-run mode")
	}
}
