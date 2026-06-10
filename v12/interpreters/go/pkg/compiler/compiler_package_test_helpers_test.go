package compiler

import (
	"os"
	"path/filepath"
	"testing"

	"able/interpreter-go/pkg/driver"
)

func compileNoFallbackPackage(t *testing.T, pkgName string, files map[string]string) *Result {
	return compileNoFallbackPackageWithOptions(t, pkgName, files, Options{
		PackageName:        "main",
		RequireNoFallbacks: true,
		EmitMain:           true,
	})
}

func compileNoFallbackPackageWithOptions(t *testing.T, pkgName string, files map[string]string, opts Options) *Result {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "package.yml"), []byte("name: "+pkgName+"\n"), 0o600); err != nil {
		t.Fatalf("write package.yml: %v", err)
	}
	entryPath := filepath.Join(root, "main.able")
	for rel, content := range files {
		path := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
		}
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatalf("write %s: %v", rel, err)
		}
	}

	loader, err := driver.NewLoader(nil)
	if err != nil {
		t.Fatalf("loader init: %v", err)
	}
	t.Cleanup(func() { loader.Close() })

	program, err := loader.Load(entryPath)
	if err != nil {
		t.Fatalf("load program: %v", err)
	}

	opts.EntryPath = entryPath
	result, err := New(opts).Compile(program)
	if err != nil {
		t.Fatalf("compile with no fallbacks: %v", err)
	}
	if len(result.Fallbacks) != 0 {
		t.Fatalf("expected no fallbacks, got %v", result.Fallbacks)
	}
	return result
}
