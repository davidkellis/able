package interpreter

import (
	"os"
	"path/filepath"
	"testing"

	"able/interpreter-go/pkg/driver"
)

func TestFindStdlibRootsPrefersConfiguredRoot(t *testing.T) {
	configured := filepath.Join(t.TempDir(), "configured", "src")
	if err := os.MkdirAll(filepath.Join(configured, "able"), 0o755); err != nil {
		t.Fatalf("mkdir configured root: %v", err)
	}
	t.Setenv("ABLE_STDLIB_ROOT", configured)

	got := findStdlibRoots(t.TempDir())
	if len(got) != 1 || got[0] != configured {
		t.Fatalf("findStdlibRoots() = %v, want [%s]", got, configured)
	}
}

func TestFindStdlibRootsFindsSiblingWorkspace(t *testing.T) {
	root := t.TempDir()
	start := filepath.Join(root, "able", "v12", "fixtures", "exec")
	want := filepath.Join(root, "able-stdlib", "src")
	if err := os.MkdirAll(start, 0o755); err != nil {
		t.Fatalf("mkdir start: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(want, "able"), 0o755); err != nil {
		t.Fatalf("mkdir sibling stdlib: %v", err)
	}
	t.Setenv("ABLE_STDLIB_ROOT", "")

	got := findStdlibRoots(start)
	if len(got) != 1 || got[0] != want {
		t.Fatalf("findStdlibRoots() = %v, want [%s]", got, want)
	}
}

func TestFixtureStdlibRootSourceRecognizesSiblingCheckout(t *testing.T) {
	root := filepath.Join(t.TempDir(), "able-stdlib", "src")
	if err := os.MkdirAll(filepath.Join(root, "able"), 0o755); err != nil {
		t.Fatalf("mkdir sibling stdlib: %v", err)
	}
	t.Setenv("ABLE_STDLIB_ROOT", "")

	if got := fixtureStdlibRootSource(root); got != driver.StdlibSourceOverride {
		t.Fatalf("fixtureStdlibRootSource() = %v, want override", got)
	}
}

func TestBuildExecSearchPathsKeepsEntryStdlibRoot(t *testing.T) {
	root := t.TempDir()
	fixtureDir := filepath.Join(root, "able", "v12", "fixtures", "exec")
	entryPath := filepath.Join(fixtureDir, "main.able")
	stdlibRoot := filepath.Join(root, "able-stdlib")
	stdlibSrc := filepath.Join(stdlibRoot, "src")
	cacheHome := filepath.Join(root, "cache")
	cacheRoot := filepath.Join(cacheHome, "pkg", "src", "able", "0.1.0")
	cacheSrc := filepath.Join(cacheRoot, "src")

	for _, dir := range []string{fixtureDir, filepath.Join(stdlibSrc, "able"), filepath.Join(cacheSrc, "able")} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}
	for _, packageRoot := range []string{stdlibRoot, cacheRoot} {
		if err := os.WriteFile(filepath.Join(packageRoot, "package.yml"), []byte("name: able\nversion: 0.1.0\n"), 0o600); err != nil {
			t.Fatalf("write package.yml: %v", err)
		}
	}
	if err := os.WriteFile(entryPath, []byte("package example\n"), 0o600); err != nil {
		t.Fatalf("write entry: %v", err)
	}

	t.Setenv("ABLE_STDLIB_ROOT", "")
	t.Setenv("ABLE_HOME", cacheHome)
	paths, err := buildExecSearchPaths(entryPath, fixtureDir, fixtureManifest{})
	if err != nil {
		t.Fatalf("buildExecSearchPaths() error: %v", err)
	}

	var stdlibPaths []driver.SearchPath
	for _, path := range paths {
		canonical, err := driver.CanonicalizeStdlibCandidateRoot(path.Path)
		if err != nil || canonical != stdlibSrc {
			continue
		}
		stdlibPaths = append(stdlibPaths, path)
	}
	if len(stdlibPaths) != 1 {
		t.Fatalf("selected stdlib paths = %v, want one path for %s", stdlibPaths, stdlibSrc)
	}
	if got := stdlibPaths[0].StdlibSource; got != driver.StdlibSourceOverride {
		t.Fatalf("selected stdlib source = %v, want override", got)
	}
}
