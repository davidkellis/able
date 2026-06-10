package driver

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverPackagesUsesLoaderPackageNamesAndSkipsTests(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "src", "core"), 0o755); err != nil {
		t.Fatalf("mkdir source: %v", err)
	}
	writeFile(t, filepath.Join(root, "package.yml"), "name: able\n")
	writeFile(t, filepath.Join(root, "src", "core", "value.able"), "package value\n")
	writeFile(t, filepath.Join(root, "src", "core", "value.test.able"), "package value_test\n")

	packages, err := DiscoverPackages([]SearchPath{{
		Path: filepath.Join(root, "src"),
		Kind: RootStdlib,
	}}, false)
	if err != nil {
		t.Fatalf("DiscoverPackages: %v", err)
	}
	if len(packages) != 1 {
		t.Fatalf("package count = %d, want 1 (%#v)", len(packages), packages)
	}
	if packages[0].Name != "able.core.value" {
		t.Fatalf("package name = %q, want able.core.value", packages[0].Name)
	}
	if len(packages[0].Files) != 1 || filepath.Base(packages[0].Files[0]) != "value.able" {
		t.Fatalf("unexpected package files %#v", packages[0].Files)
	}
}
