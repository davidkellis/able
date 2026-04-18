package interpreter

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindTestStdlibRootPrefersCachedAbleHome(t *testing.T) {
	repoRoot := t.TempDir()
	cacheHome := filepath.Join(t.TempDir(), ".able")
	cacheSrc := filepath.Join(cacheHome, "pkg", "src", "able", "0.1.0", "src")
	siblingSrc := filepath.Join(repoRoot, "able-stdlib", "src")

	if err := os.MkdirAll(cacheSrc, 0o755); err != nil {
		t.Fatalf("mkdir cache src: %v", err)
	}
	if err := os.MkdirAll(siblingSrc, 0o755); err != nil {
		t.Fatalf("mkdir sibling src: %v", err)
	}

	t.Setenv("ABLE_STDLIB_ROOT", "")
	t.Setenv("ABLE_HOME", cacheHome)

	if got := findTestStdlibRoot(repoRoot); got != cacheSrc {
		t.Fatalf("findTestStdlibRoot() = %q, want %q", got, cacheSrc)
	}
}
