package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateRepoName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid simple", "myapp", false},
		{"valid with dash", "my-app", false},
		{"valid with numbers", "app123", false},
		{"valid mixed", "my-app-123", false},
		{"empty", "", true},
		{"starts with dash", "-myapp", true},
		{"ends with dash", "myapp-", true},
		{"uppercase", "MyApp", true},
		{"contains dot", "my.app", true},
		{"too long", "this-is-a-very-long-repository-name-that-exceeds-the-maximum-allowed-length-of-63-characters", true},
		{"contains underscore", "my_app", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRepoName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRepoName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateNamespace(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid simple", "default", false},
		{"valid with dash", "my-namespace", false},
		{"valid with numbers", "ns123", false},
		{"empty", "", true},
		{"reserved kube-system", "kube-system", true},
		{"reserved kube-public", "kube-public", true},
		{"reserved kube-node-lease", "kube-node-lease", true},
		{"uppercase", "Default", true},
		{"starts with number", "123ns", false},
		{"too long", "this-is-a-very-long-namespace-name-that-exceeds-the-maximum-allowed-length-of-63-characters", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNamespace(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateNamespace(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateProjectName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid simple", "myproject", false},
		{"valid with dash", "my-project", false},
		{"valid with dot", "my.project", false},
		{"valid subdomain", "my.project.name", false},
		{"empty", "", true},
		{"uppercase", "MyProject", true},
		{"starts with dash", "-myproject", true},
		{"too long", "this-is-a-very-long-project-name-that-exceeds-the-maximum-allowed-length-of-253-characters-and-should-fail-validation-because-it-is-way-too-long", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateProjectName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateProjectName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestSanitizeName(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{"already valid", "myapp", "myapp"},
		{"uppercase", "MyApp", "myapp"},
		{"with spaces", "my app", "my-app"},
		{"with underscore", "my_app", "my-app"},
		{"starts with dash", "-myapp", "myapp"},
		{"ends with dash", "myapp-", "myapp"},
		{"too long", "this-is-a-very-long-name-that-exceeds-the-maximum-allowed-length", "this-is-a-very-long-name-that-exceeds-the-maximum-allowed-lengt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeName(tt.input)
			if result != tt.expect {
				t.Errorf("SanitizeName(%q) = %q, want %q", tt.input, result, tt.expect)
			}
		})
	}
}

func TestIsPathTraversal(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect bool
	}{
		{"safe path", "myrepo", false},
		{"safe nested", "my/repo", false},
		{"dot dot", "..", true},
		{"dot slash", "./", true},
		{"parent dir", "../", true},
		{"hidden parent", "/..", true},
		{"windows style", "..\\", true},
		{"windows backslash", "\\..", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsPathTraversal(tt.input)
			if result != tt.expect {
				t.Errorf("IsPathTraversal(%q) = %v, want %v", tt.input, result, tt.expect)
			}
		})
	}
}

func TestValidateSafePath(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"safe path", "myrepo", false},
		{"safe nested", "my/repo", false},
		{"path traversal", "../etc", true},
		{"path traversal hidden", "..hidden", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSafePath(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSafePath(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestCreateManiplacerProject(t *testing.T) {
	tmpDir := t.TempDir()

	err := CreateManiplacerProject(tmpDir)
	if err != nil {
		t.Fatalf("CreateManiplacerProject() error = %v", err)
	}

	// Check if .maniplacer file exists
	markerPath := filepath.Join(tmpDir, ManiplacerMarker)
	if _, err := os.Stat(markerPath); os.IsNotExist(err) {
		t.Errorf("CreateManiplacerProject() did not create %s file", ManiplacerMarker)
	}

	// Verify file permissions
	info, err := os.Stat(markerPath)
	if err != nil {
		t.Fatalf("Could not stat marker file: %v", err)
	}

	if info.Mode().Perm() != FilePermission {
		t.Errorf("Marker file permissions = %v, want %v", info.Mode().Perm(), FilePermission)
	}
}

func TestIsValidProject(t *testing.T) {
	tmpDir := t.TempDir()

	// Test with no marker file
	if IsValidProject() {
		t.Error("IsValidProject() returned true for directory without marker file")
	}

	// Test with valid marker file
	err := CreateManiplacerProject(tmpDir)
	if err != nil {
		t.Fatalf("CreateManiplacerProject() error = %v", err)
	}

	// Change to temp directory
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	if !IsValidProject() {
		t.Error("IsValidProject() returned false for valid project")
	}
}
