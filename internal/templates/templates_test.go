package templates

import (
	"strings"
	"testing"
	"text/template"
)

func TestBase64Function(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{"simple", "hello", "aGVsbG8="},
		{"empty", "", ""},
		{"with spaces", "hello world", "aGVsbG8gd29ybGQ="},
		{"special chars", "test123!", "dGVzdDEyMyE="},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ManiplacerFuncs["Base64"].(func(string) string)(tt.input)
			if result != tt.expect {
				t.Errorf("Base64(%q) = %q, want %q", tt.input, result, tt.expect)
			}
		})
	}
}

func TestToUpperFunction(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{"simple", "hello", "HELLO"},
		{"empty", "", ""},
		{"mixed", "HeLLo WoRLd", "HELLO WORLD"},
		{"already upper", "HELLO", "HELLO"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ManiplacerFuncs["ToUpper"].(func(string) string)(tt.input)
			if result != tt.expect {
				t.Errorf("ToUpper(%q) = %q, want %q", tt.input, result, tt.expect)
			}
		})
	}
}

func TestToLowerFunction(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{"simple", "HELLO", "hello"},
		{"empty", "", ""},
		{"mixed", "HeLLo WoRLd", "hello world"},
		{"already lower", "hello", "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ManiplacerFuncs["ToLower"].(func(string) string)(tt.input)
			if result != tt.expect {
				t.Errorf("ToLower(%q) = %q, want %q", tt.input, result, tt.expect)
			}
		})
	}
}

func TestQuoteFunction(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{"simple", "hello", `"hello"`},
		{"empty", "", `""`},
		{"with spaces", "hello world", `"hello world"`},
		{"with quotes", `hello "world"`, `"hello \"world\""`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ManiplacerFuncs["Quote"].(func(string) string)(tt.input)
			if result != tt.expect {
				t.Errorf("Quote(%q) = %q, want %q", tt.input, result, tt.expect)
			}
		})
	}
}

func TestTemplateWithFunctions(t *testing.T) {
	tests := []struct {
		name     string
		template string
		data     map[string]any
		expect   string
	}{
		{
			name:     "base64 encoding",
			template: `{{ .value | Base64 }}`,
			data:     map[string]any{"value": "hello"},
			expect:   "aGVsbG8=",
		},
		{
			name:     "toUpper",
			template: `{{ .value | ToUpper }}`,
			data:     map[string]any{"value": "hello"},
			expect:   "HELLO",
		},
		{
			name:     "toLower",
			template: `{{ .value | ToLower }}`,
			data:     map[string]any{"value": "HELLO"},
			expect:   "hello",
		},
		{
			name:     "quote",
			template: `{{ .value | Quote }}`,
			data:     map[string]any{"value": "hello"},
			expect:   `"hello"`,
		},
		{
			name:     "chained functions",
			template: `{{ .value | ToUpper | Quote }}`,
			data:     map[string]any{"value": "hello"},
			expect:   `"HELLO"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := template.New("test").Funcs(ManiplacerFuncs).Parse(tt.template)
			if err != nil {
				t.Fatalf("Failed to parse template: %v", err)
			}

			var buf strings.Builder
			err = tmpl.Execute(&buf, tt.data)
			if err != nil {
				t.Fatalf("Failed to execute template: %v", err)
			}

			if buf.String() != tt.expect {
				t.Errorf("Template execution = %q, want %q", buf.String(), tt.expect)
			}
		})
	}
}

func TestAllowedComponents(t *testing.T) {
	expectedComponents := []string{
		"deployment",
		"service",
		"httproute",
		"secret",
		"configmap",
		"hpa",
		"hcpolicy",
	}

	if len(AllowedComponents) != len(expectedComponents) {
		t.Errorf("AllowedComponents has %d items, want %d", len(AllowedComponents), len(expectedComponents))
	}

	for i, comp := range expectedComponents {
		if i >= len(AllowedComponents) {
			t.Errorf("Missing component: %s", comp)
			continue
		}
		if AllowedComponents[i] != comp {
			t.Errorf("AllowedComponents[%d] = %q, want %q", i, AllowedComponents[i], comp)
		}
	}
}

func TestTemplateRegistry(t *testing.T) {
	// Check that all allowed components have templates
	for _, comp := range AllowedComponents {
		tmpl, exists := TemplateRegistry[comp]
		if !exists {
			t.Errorf("TemplateRegistry missing template for component: %s", comp)
			continue
		}
		if len(tmpl) == 0 {
			t.Errorf("Template for %s is empty", comp)
		}
	}
}
