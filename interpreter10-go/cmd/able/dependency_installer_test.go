package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"able/interpreter10-go/pkg/driver"
)

func TestDependencyInstaller_PathDependency(t *testing.T) {
	root := t.TempDir()
	mainDir := filepath.Join(root, "app")
	depDir := filepath.Join(root, "dep")
	if err := os.MkdirAll(filepath.Join(mainDir, "src"), 0o755); err != nil {
		t.Fatalf("mkdir main: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(depDir, "src"), 0o755); err != nil {
		t.Fatalf("mkdir dep: %v", err)
	}

	writeFile(t, filepath.Join(mainDir, "package.yml"), `
name: app
version: 0.1.0
dependencies:
  dep:
    path: ../dep
`)

	writeFile(t, filepath.Join(depDir, "package.yml"), `
name: dep
version: 0.2.0
`)

	manifest, err := driver.LoadManifest(filepath.Join(mainDir, "package.yml"))
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}

	lock := driver.NewLockfile(manifest.Name, cliToolVersion)
	cacheDir := filepath.Join(root, ".able")
	installer := newDependencyInstaller(manifest, cacheDir)

	changed, logs, err := installer.Install(lock)
	if err != nil {
		t.Fatalf("Install returned error: %v", err)
	}
	if !changed {
		t.Fatalf("expected lockfile to change for new dependency")
	}
	if len(logs) == 0 {
		t.Fatalf("expected logging output for dependency resolution")
	}
	if len(lock.Packages) != 1 {
		t.Fatalf("lock packages = %#v", lock.Packages)
	}
	pkg := lock.Packages[0]
	if pkg.Name != "dep" || pkg.Version != "0.2.0" {
		t.Fatalf("lock entry unexpected: %#v", pkg)
	}
	if !strings.HasPrefix(pkg.Source, "path:") {
		t.Fatalf("expected path source, got %q", pkg.Source)
	}
	if len(pkg.Dependencies) != 0 {
		t.Fatalf("expected no transitive dependencies, got %#v", pkg.Dependencies)
	}
}

func TestDependencyInstaller_PathDependencyTransitive(t *testing.T) {
	root := t.TempDir()
	mainDir := filepath.Join(root, "app")
	depDir := filepath.Join(root, "dep")
	subDir := filepath.Join(root, "sub")

	for _, dir := range []string{
		filepath.Join(mainDir, "src"),
		filepath.Join(depDir, "src"),
		filepath.Join(subDir, "src"),
	} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}

	writeFile(t, filepath.Join(mainDir, "package.yml"), `
name: app
version: 0.1.0
dependencies:
  dep:
    path: ../dep
`)

	writeFile(t, filepath.Join(depDir, "package.yml"), `
name: dep
version: 1.0.0
dependencies:
  sub:
    path: ../sub
`)

	writeFile(t, filepath.Join(subDir, "package.yml"), `
name: sub
version: 2.0.0
`)

	manifest, err := driver.LoadManifest(filepath.Join(mainDir, "package.yml"))
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}

	lock := driver.NewLockfile(manifest.Name, cliToolVersion)
	cacheDir := filepath.Join(root, ".able")
	installer := newDependencyInstaller(manifest, cacheDir)

	changed, _, err := installer.Install(lock)
	if err != nil {
		t.Fatalf("Install returned error: %v", err)
	}
	if !changed {
		t.Fatalf("expected lockfile to record new dependencies")
	}
	if len(lock.Packages) != 2 {
		t.Fatalf("expected two packages in lock, got %#v", lock.Packages)
	}
	first := lock.Packages[0]
	second := lock.Packages[1]
	if first.Name != "dep" || second.Name != "sub" {
		t.Fatalf("unexpected package ordering: %#v", lock.Packages)
	}
	if len(first.Dependencies) != 1 || first.Dependencies[0].Name != "sub" {
		t.Fatalf("dep dependencies incorrect: %#v", first.Dependencies)
	}
	if len(second.Dependencies) != 0 {
		t.Fatalf("sub should have no dependencies, got %#v", second.Dependencies)
	}
}

func TestDependencyInstaller_RegistryDependency(t *testing.T) {
	root := t.TempDir()
	registry := filepath.Join(root, "registry")
	mathRoot := filepath.Join(registry, "default", "math", "1.0.0")
	helperRoot := filepath.Join(registry, "default", "helper", "0.5.0")

	for _, dir := range []string{
		filepath.Join(mathRoot, "src"),
		filepath.Join(helperRoot, "src"),
		filepath.Join(root, "app", "src"),
	} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}

	writeFile(t, filepath.Join(mathRoot, "package.yml"), `
name: math
version: 1.0.0
dependencies:
  helper: "0.5.0"
`)
	writeFile(t, filepath.Join(mathRoot, "src", "core.able"), `
package core

fn pow(x: i32) -> i32 {
  helper.helper_value() + x
}
`)
	writeFile(t, filepath.Join(helperRoot, "package.yml"), `
name: helper
version: 0.5.0
`)
	writeFile(t, filepath.Join(helperRoot, "src", "core.able"), `
package core

fn helper_value() -> i32 {
  10
}
`)

	appDir := filepath.Join(root, "app")
	writeFile(t, filepath.Join(appDir, "package.yml"), `
name: app
version: 0.1.0
dependencies:
  math: "1.0.0"
`)

	manifest, err := driver.LoadManifest(filepath.Join(appDir, "package.yml"))
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}

	cacheDir := filepath.Join(root, ".able")
	t.Setenv("ABLE_REGISTRY", registry)
	installer := newDependencyInstaller(manifest, cacheDir)
	lock := driver.NewLockfile(manifest.Name, cliToolVersion)

	changed, _, err := installer.Install(lock)
	if err != nil {
		t.Fatalf("Install error: %v", err)
	}
	if !changed {
		t.Fatalf("expected lockfile change for registry dependency")
	}
	if len(lock.Packages) != 2 {
		t.Fatalf("expected two packages in lockfile, got %#v", lock.Packages)
	}
	sort.Slice(lock.Packages, func(i, j int) bool {
		return lock.Packages[i].Name < lock.Packages[j].Name
	})
	mathPkg := lock.Packages[1]
	helperPkg := lock.Packages[0]
	if mathPkg.Name != "math" || helperPkg.Name != "helper" {
		t.Fatalf("unexpected lockfile entries: %#v", lock.Packages)
	}
	if len(mathPkg.Dependencies) != 1 || mathPkg.Dependencies[0].Name != "helper" {
		t.Fatalf("math dependencies incorrect: %#v", mathPkg.Dependencies)
	}
	if len(helperPkg.Dependencies) != 0 {
		t.Fatalf("helper should have no dependencies, got %#v", helperPkg.Dependencies)
	}
}

func TestDependencyInstaller_GitDependency(t *testing.T) {
	root := t.TempDir()
	repo := filepath.Join(root, "repo")
	if err := os.MkdirAll(filepath.Join(repo, "src"), 0o755); err != nil {
		t.Fatalf("mkdir repo: %v", err)
	}
	writeFile(t, filepath.Join(repo, "package.yml"), `
name: gitpkg
version: 0.2.0
`)
	writeFile(t, filepath.Join(repo, "src", "core.able"), `package gitpkg
fn value() -> string { "git" }
`)

	rev := initGitRepo(t, repo)

	mainDir := filepath.Join(root, "app")
	if err := os.MkdirAll(filepath.Join(mainDir, "src"), 0o755); err != nil {
		t.Fatalf("mkdir app: %v", err)
	}
	writeFile(t, filepath.Join(mainDir, "package.yml"), `
name: app
version: 0.1.0
dependencies:
  gitpkg:
    git: `+repo+`
    rev: `+rev+`
`)

	manifest, err := driver.LoadManifest(filepath.Join(mainDir, "package.yml"))
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}

	cacheDir := filepath.Join(root, "cache")
	installer := newDependencyInstaller(manifest, cacheDir)
	lock := driver.NewLockfile(manifest.Name, cliToolVersion)

	changed, _, err := installer.Install(lock)
	if err != nil {
		t.Fatalf("Install error: %v", err)
	}
	if !changed {
		t.Fatalf("expected lockfile change for git dependency")
	}
	if len(lock.Packages) != 1 {
		t.Fatalf("lock packages unexpected: %#v", lock.Packages)
	}
	pkg := lock.Packages[0]
	expectedSource := fmt.Sprintf("git+%s@%s", repo, rev)
	if pkg.Source != expectedSource {
		t.Fatalf("pkg.Source = %q, want %q", pkg.Source, expectedSource)
	}
	if pkg.Name != "gitpkg" {
		t.Fatalf("pkg.Name = %q, want gitpkg", pkg.Name)
	}
	cached := filepath.Join(cacheDir, "pkg", "src", pkg.Name, sanitizePathSegment(pkg.Version))
	if _, err := os.Stat(cached); err != nil {
		t.Fatalf("expected cached git package at %s: %v", cached, err)
	}
	if pkg.Version != rev {
		t.Fatalf("pkg.Version = %q, want %q", pkg.Version, rev)
	}
	if len(pkg.Dependencies) != 0 {
		t.Fatalf("expected no transitive dependencies for git package, got %#v", pkg.Dependencies)
	}
}

func TestDependencyInstaller_GitDependencyBranch(t *testing.T) {
	root := t.TempDir()
	repo := filepath.Join(root, "repo")
	if err := os.MkdirAll(filepath.Join(repo, "src"), 0o755); err != nil {
		t.Fatalf("mkdir repo: %v", err)
	}
	writeFile(t, filepath.Join(repo, "package.yml"), `
name: gitpkg
version: 0.3.0
`)
	writeFile(t, filepath.Join(repo, "src", "core.able"), `package gitpkg
fn value() -> string { "branch" }
`)

	rev := initGitRepo(t, repo)

	mainDir := filepath.Join(root, "app")
	if err := os.MkdirAll(filepath.Join(mainDir, "src"), 0o755); err != nil {
		t.Fatalf("mkdir app: %v", err)
	}
	writeFile(t, filepath.Join(mainDir, "package.yml"), `
name: app
version: 0.1.0
dependencies:
  gitpkg:
    git: `+repo+`
    branch: master
`)

	manifest, err := driver.LoadManifest(filepath.Join(mainDir, "package.yml"))
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}

	cacheDir := filepath.Join(root, "cache")
	installer := newDependencyInstaller(manifest, cacheDir)
	lock := driver.NewLockfile(manifest.Name, cliToolVersion)

	changed, _, err := installer.Install(lock)
	if err != nil {
		t.Fatalf("Install error: %v", err)
	}
	if !changed {
		t.Fatalf("expected lockfile change for git branch dependency")
	}
	if len(lock.Packages) != 1 {
		t.Fatalf("lock packages unexpected: %#v", lock.Packages)
	}
	pkg := lock.Packages[0]
	wantVersion := fmt.Sprintf("master@%s", rev)
	if pkg.Version != wantVersion {
		t.Fatalf("pkg.Version = %q, want %q", pkg.Version, wantVersion)
	}
	if pkg.Name != "gitpkg" {
		t.Fatalf("pkg.Name = %q, want gitpkg", pkg.Name)
	}
	expectedSource := fmt.Sprintf("git+%s@%s", repo, rev)
	if pkg.Source != expectedSource {
		t.Fatalf("pkg.Source = %q, want %q", pkg.Source, expectedSource)
	}
	cached := filepath.Join(cacheDir, "pkg", "src", pkg.Name, sanitizePathSegment(pkg.Version))
	if _, err := os.Stat(cached); err != nil {
		t.Fatalf("expected cached git package at %s: %v", cached, err)
	}
}
