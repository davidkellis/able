package stdlibpath

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveRepoOrInstalledSrcPrefersExplicitOverride(t *testing.T) {
	repoRoot := t.TempDir()
	overrideRoot := filepath.Join(t.TempDir(), "stdlib")
	overrideSrc := filepath.Join(overrideRoot, "src")
	siblingSrc := filepath.Join(repoRoot, "able-stdlib", "src")

	if err := os.MkdirAll(overrideSrc, 0o755); err != nil {
		t.Fatalf("mkdir override src: %v", err)
	}
	if err := os.MkdirAll(siblingSrc, 0o755); err != nil {
		t.Fatalf("mkdir sibling src: %v", err)
	}

	t.Setenv("ABLE_STDLIB_ROOT", overrideRoot)
	t.Setenv("ABLE_HOME", filepath.Join(t.TempDir(), ".able"))

	if got := ResolveRepoOrInstalledSrc(repoRoot); got != overrideSrc {
		t.Fatalf("ResolveRepoOrInstalledSrc() = %q, want %q", got, overrideSrc)
	}
}

func TestResolveRepoOrInstalledSrcPrefersCachedAbleHome(t *testing.T) {
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

	if got := ResolveRepoOrInstalledSrc(repoRoot); got != cacheSrc {
		t.Fatalf("ResolveRepoOrInstalledSrc() = %q, want %q", got, cacheSrc)
	}
}
