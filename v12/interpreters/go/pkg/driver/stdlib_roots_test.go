package driver

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCanonicalizeStdlibCandidateRootPrefersSrcDir(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "src")
	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatalf("mkdir src: %v", err)
	}

	got, err := CanonicalizeStdlibCandidateRoot(root)
	if err != nil {
		t.Fatalf("CanonicalizeStdlibCandidateRoot returned error: %v", err)
	}
	if got != src {
		t.Fatalf("canonical root = %q, want %q", got, src)
	}
}

func TestCanonicalizeStdlibCandidateRootKeepsSourceDir(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "src")
	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatalf("mkdir src: %v", err)
	}

	got, err := CanonicalizeStdlibCandidateRoot(src)
	if err != nil {
		t.Fatalf("CanonicalizeStdlibCandidateRoot returned error: %v", err)
	}
	if got != src {
		t.Fatalf("canonical root = %q, want %q", got, src)
	}
}

func TestDetermineEntryRootMetadataClassifiesWorkspaceAbleRoot(t *testing.T) {
	root := t.TempDir()
	kind, source := determineEntryRootMetadata(root, "able", nil)
	if kind != RootStdlib {
		t.Fatalf("kind = %v, want RootStdlib", kind)
	}
	if source != StdlibSourceWorkspace {
		t.Fatalf("source = %v, want workspace", source)
	}
}

func TestDetermineEntryRootMetadataPrefersOverlappingSearchPathSource(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "src")
	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatalf("mkdir src: %v", err)
	}

	searchPaths := []SearchPath{{
		Path:         src,
		Kind:         RootStdlib,
		StdlibSource: StdlibSourceOverride,
	}}
	kind, source := determineEntryRootMetadata(src, "able", searchPaths)
	if kind != RootStdlib {
		t.Fatalf("kind = %v, want RootStdlib", kind)
	}
	if source != StdlibSourceOverride {
		t.Fatalf("source = %v, want override", source)
	}
}

func TestDetermineSearchRootMetadataPromotesAbleEnvRootToStdlib(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "src")
	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatalf("mkdir src: %v", err)
	}

	kind, source := determineSearchRootMetadata(SearchPath{
		Path:         src,
		Kind:         RootUser,
		StdlibSource: StdlibSourceEnv,
	}, src, "able")
	if kind != RootStdlib {
		t.Fatalf("kind = %v, want RootStdlib", kind)
	}
	if source != StdlibSourceEnv {
		t.Fatalf("source = %v, want env", source)
	}
}

func TestNewLoaderPreservesStdlibSourceClass(t *testing.T) {
	root := t.TempDir()
	loader, err := NewLoader([]SearchPath{{
		Path:         root,
		Kind:         RootStdlib,
		StdlibSource: StdlibSourceCache,
	}})
	if err != nil {
		t.Fatalf("NewLoader: %v", err)
	}
	defer loader.Close()

	if len(loader.searchPaths) != 1 {
		t.Fatalf("len(searchPaths) = %d, want 1", len(loader.searchPaths))
	}
	if loader.searchPaths[0].StdlibSource != StdlibSourceCache {
		t.Fatalf("stdlib source = %v, want cache", loader.searchPaths[0].StdlibSource)
	}
}

func TestResolveCanonicalStdlibSearchPathsManifestPrefersLockfile(t *testing.T) {
	root := t.TempDir()
	lockfileRoot := writeStdlibCandidateRoot(t, filepath.Join(root, "lockfile"))
	envRoot := writeStdlibCandidateRoot(t, filepath.Join(root, "env"))
	lockfileSrc := filepath.Join(lockfileRoot, "src")
	envSrc := filepath.Join(envRoot, "src")

	filtered, err := ResolveCanonicalStdlibSearchPaths([]SearchPath{
		{Path: lockfileRoot, Kind: RootStdlib, StdlibSource: StdlibSourceLockfile},
		{Path: envRoot, Kind: RootStdlib, StdlibSource: StdlibSourceEnv},
	}, true)
	if err == nil {
		t.Fatalf("expected collision error, got nil filtered=%v", filtered)
	}
	if !strings.Contains(err.Error(), "selected canonical stdlib root (lockfile)") ||
		!strings.Contains(err.Error(), lockfileSrc) ||
		!strings.Contains(err.Error(), "distinct visible stdlib root (env)") ||
		!strings.Contains(err.Error(), envSrc) {
		t.Fatalf("unexpected collision error: %v", err)
	}
}

func TestResolveCanonicalStdlibSearchPathsAdhocRejectsOverrideEnvCollision(t *testing.T) {
	root := t.TempDir()
	overrideRoot := writeStdlibCandidateRoot(t, filepath.Join(root, "override"))
	envRoot := writeStdlibCandidateRoot(t, filepath.Join(root, "env"))
	overrideSrc := filepath.Join(overrideRoot, "src")
	envSrc := filepath.Join(envRoot, "src")

	filtered, err := ResolveCanonicalStdlibSearchPaths([]SearchPath{
		{Path: overrideRoot, Kind: RootStdlib, StdlibSource: StdlibSourceOverride},
		{Path: envRoot, Kind: RootStdlib, StdlibSource: StdlibSourceEnv},
	}, false)
	if err == nil {
		t.Fatalf("expected collision error, got nil filtered=%v", filtered)
	}
	if !strings.Contains(err.Error(), "selected canonical stdlib root (override)") ||
		!strings.Contains(err.Error(), overrideSrc) ||
		!strings.Contains(err.Error(), "distinct visible stdlib root (env)") ||
		!strings.Contains(err.Error(), envSrc) {
		t.Fatalf("unexpected collision error: %v", err)
	}
}

func TestResolveCanonicalStdlibSearchPathsRejectsMultipleEnvRoots(t *testing.T) {
	root := t.TempDir()
	envRootA := writeStdlibCandidateRoot(t, filepath.Join(root, "env-a"))
	envRootB := writeStdlibCandidateRoot(t, filepath.Join(root, "env-b"))
	envSrcA := filepath.Join(envRootA, "src")
	envSrcB := filepath.Join(envRootB, "src")

	filtered, err := ResolveCanonicalStdlibSearchPaths([]SearchPath{
		{Path: envRootA, Kind: RootStdlib, StdlibSource: StdlibSourceEnv},
		{Path: envRootB, Kind: RootStdlib, StdlibSource: StdlibSourceEnv},
	}, false)
	if err == nil {
		t.Fatalf("expected collision error, got nil filtered=%v", filtered)
	}
	if !strings.Contains(err.Error(), "multiple env-provided `name: able` roots are visible") ||
		!strings.Contains(err.Error(), envSrcA) ||
		!strings.Contains(err.Error(), envSrcB) ||
		!strings.Contains(err.Error(), "ABLE_MODULE_PATHS / ABLE_PATH") {
		t.Fatalf("unexpected collision error: %v", err)
	}
}

func TestResolveCanonicalStdlibSearchPathsAdhocUsesSingleEnvRoot(t *testing.T) {
	root := t.TempDir()
	envRoot := writeStdlibCandidateRoot(t, filepath.Join(root, "env"))

	filtered, err := ResolveCanonicalStdlibSearchPaths([]SearchPath{
		{Path: envRoot, Kind: RootStdlib, StdlibSource: StdlibSourceEnv},
	}, false)
	if err != nil {
		t.Fatalf("ResolveCanonicalStdlibSearchPaths returned error: %v", err)
	}
	if len(filtered) != 1 {
		t.Fatalf("len(filtered) = %d, want 1", len(filtered))
	}
	if filtered[0].Path != filepath.Join(envRoot, "src") {
		t.Fatalf("filtered path = %q, want %q", filtered[0].Path, filepath.Join(envRoot, "src"))
	}
}

func TestResolveCanonicalStdlibSearchPathsAllowsDuplicateSameRoot(t *testing.T) {
	root := t.TempDir()
	stdlibRoot := writeStdlibCandidateRoot(t, filepath.Join(root, "stdlib"))

	filtered, err := ResolveCanonicalStdlibSearchPaths([]SearchPath{
		{Path: stdlibRoot, Kind: RootStdlib, StdlibSource: StdlibSourceOverride},
		{Path: filepath.Join(stdlibRoot, "src"), Kind: RootStdlib, StdlibSource: StdlibSourceEnv},
	}, false)
	if err != nil {
		t.Fatalf("ResolveCanonicalStdlibSearchPaths returned error: %v", err)
	}
	if len(filtered) != 1 {
		t.Fatalf("len(filtered) = %d, want 1", len(filtered))
	}
	if filtered[0].Path != filepath.Join(stdlibRoot, "src") {
		t.Fatalf("filtered path = %q, want %q", filtered[0].Path, filepath.Join(stdlibRoot, "src"))
	}
}

func writeStdlibCandidateRoot(t *testing.T, root string) string {
	t.Helper()
	src := filepath.Join(root, "src")
	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatalf("mkdir stdlib src: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "package.yml"), []byte("name: able\nversion: 0.0.1\n"), 0o600); err != nil {
		t.Fatalf("write package.yml: %v", err)
	}
	return root
}
