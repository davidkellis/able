package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"able/interpreter10-go/pkg/driver"
)

func TestFindManifest(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "package.yml"), []byte("name: test\n"), 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	child := filepath.Join(root, "src", "app")
	if err := os.MkdirAll(child, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	found, err := findManifest(child)
	if err != nil {
		t.Fatalf("findManifest returned error: %v", err)
	}
	want := filepath.Join(root, "package.yml")
	if found != want {
		t.Fatalf("findManifest = %q, want %q", found, want)
	}
}

func TestResolveAbleHomeEnv(t *testing.T) {
	tmp := t.TempDir()
	target := filepath.Join(tmp, "cache")
	t.Setenv("ABLE_HOME", target)

	got, err := resolveAbleHome()
	if err != nil {
		t.Fatalf("resolveAbleHome error: %v", err)
	}
	if got != target {
		t.Fatalf("resolveAbleHome = %q, want %q", got, target)
	}
}

func TestResolveAbleHomeDefault(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("ABLE_HOME", "")
	t.Setenv("HOME", tmp)

	got, err := resolveAbleHome()
	if err != nil {
		t.Fatalf("resolveAbleHome error: %v", err)
	}
	if want := filepath.Join(tmp, ".able"); got != want {
		t.Fatalf("resolveAbleHome = %q, want %q", got, want)
	}
}

func TestLoadLockfileForManifest_NoDepsMissingLock(t *testing.T) {
	root := t.TempDir()
	manifest := &driver.Manifest{
		Path: filepath.Join(root, "package.yml"),
	}
	lock, err := loadLockfileForManifest(manifest)
	if err != nil {
		t.Fatalf("loadLockfileForManifest returned error: %v", err)
	}
	if lock != nil {
		t.Fatalf("expected nil lock when no dependencies, got %#v", lock)
	}
}

func TestLoadLockfileForManifest_WithDepsMissingLock(t *testing.T) {
	root := t.TempDir()
	manifestDir := filepath.Join(root, "project")
	if err := os.MkdirAll(manifestDir, 0o755); err != nil {
		t.Fatalf("mkdir project: %v", err)
	}
	manifest := &driver.Manifest{
		Path: manifestDir + "/package.yml",
		Dependencies: map[string]*driver.DependencySpec{
			"stdlib": {Version: "~> 0.1"},
		},
	}
	_, err := loadLockfileForManifest(manifest)
	if err == nil {
		t.Fatalf("expected error when lockfile missing with dependencies")
	}
	if !strings.Contains(err.Error(), "package.lock missing") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBuildExecutionSearchPaths_DefaultCacheLayout(t *testing.T) {
	cache := t.TempDir()
	t.Setenv("ABLE_HOME", cache)

	root := t.TempDir()
	manifest := &driver.Manifest{
		Path: filepath.Join(root, "package.yml"),
	}
	lock := &driver.Lockfile{
		Packages: []*driver.LockedPackage{
			{Name: "able_stdlib", Version: "0.1.0"},
		},
	}

	paths, err := buildExecutionSearchPaths(manifest, lock)
	if err != nil {
		t.Fatalf("buildExecutionSearchPaths returned error: %v", err)
	}

	want := filepath.Join(cache, "pkg", "src", "able_stdlib", "0.1.0")
	if !containsPath(paths, want) {
		t.Fatalf("expected cache path %q in %v", want, paths)
	}
	if !containsPath(paths, filepath.Dir(manifest.Path)) {
		t.Fatalf("expected manifest root in search paths: %v", paths)
	}
}
