package cli

import "testing"

func TestReleaseBinaryName(t *testing.T) {
	tests := []struct {
		goos, goarch string
		want         string
		wantErr      bool
	}{
		{goos: "linux", goarch: "amd64", want: "maniplacer-linux-amd64"},
		{goos: "darwin", goarch: "arm64", want: "maniplacer-darwin-arm64"},
		{goos: "windows", goarch: "amd64", want: "maniplacer-windows-amd64.exe"},
		{goos: "linux", goarch: "386", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.goos+"/"+tt.goarch, func(t *testing.T) {
			got, err := releaseBinaryName(tt.goos, tt.goarch)
			if (err != nil) != tt.wantErr {
				t.Fatalf("releaseBinaryName() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("releaseBinaryName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIsNewerVersion(t *testing.T) {
	tests := []struct {
		current, latest string
		want            bool
		wantErr         bool
	}{
		{current: "1.3.15", latest: "1.4.0", want: true},
		{current: "v1.4.0", latest: "1.4.0"},
		{current: "2.0.0", latest: "1.4.0"},
		{current: "dev", latest: "1.4.0", want: true},
		{current: "1.4.0-2-gabcdef", latest: "1.4.0"},
		{current: "1.4.0", latest: "not-semver", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.current+"_to_"+tt.latest, func(t *testing.T) {
			got, err := isNewerVersion(tt.current, tt.latest)
			if (err != nil) != tt.wantErr {
				t.Fatalf("isNewerVersion() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("isNewerVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}
