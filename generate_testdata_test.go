//go:build testdata

package pipass_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/thumbrise/pipass"
	"github.com/thumbrise/pipass/testdata"
)

// TestGenerateScaffolding acts as a CLI trigger under the 'testdata' build tag
func TestGenerateScaffolding(t *testing.T) {
	out, err := pipass.Compile(
		"generated",
		testdata.Stage{},
		testdata.Actor{},
		testdata.Trigger{},
		testdata.Config{},
	)
	if err != nil {
		t.Fatalf("pipeline test compile failure: %v", err)
	}

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("failed to resolve caller filepath")
	}

	projectRoot := filepath.Dir(currentFile)
	targetDir := filepath.Join(projectRoot, "testdata", "generated")

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		t.Fatalf("failed to create target directory: %v", err)
	}

	err = os.WriteFile(filepath.Join(targetDir, "generated.go"), out, 0600)
	if err != nil {
		t.Fatalf("failed to dump monolithic test scaffolding: %v", err)
	}
}
