package cli

import (
	"bytes"
	"testing"

	"github.com/dantedelordran/maniplacer/internal/utils"
)

func TestRootVersionIsMachineReadable(t *testing.T) {
	var output bytes.Buffer
	rootCmd.SetOut(&output)
	rootCmd.SetArgs([]string{"--version"})
	t.Cleanup(func() {
		rootCmd.SetOut(nil)
		rootCmd.SetArgs(nil)
	})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("rootCmd.Execute() error = %v", err)
	}
	if got, want := output.String(), utils.Version+"\n"; got != want {
		t.Errorf("--version output = %q, want %q", got, want)
	}
}
