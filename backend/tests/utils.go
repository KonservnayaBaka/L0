package tests

import (
	"os"
	"path/filepath"
	"testing"
)

func EnsureRepoRoot(t *testing.T) func() {
	t.Helper()

	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("EnsureRepoRoot: getwd failed: %v", err)
	}

	cur := origWd
	for {
		if _, err := os.Stat(filepath.Join(cur, "go.mod")); err == nil {
			if err := os.Chdir(cur); err != nil {
				t.Fatalf("EnsureRepoRoot: chdir to %s failed: %v", cur, err)
			}
			return func() {
				_ = os.Chdir(origWd)
			}
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			t.Fatalf("EnsureRepoRoot: go.mod not found starting from %s", origWd)
		}
		cur = parent
	}
}
