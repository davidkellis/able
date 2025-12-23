package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"able/interpreter-go/pkg/driver"
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
	if len(lock.Packages) != 3 {
		t.Fatalf("lock packages = %#v", lock.Packages)
	}
	depPkg := findLockedPackage(lock.Packages, "dep")
	if depPkg == nil {
		t.Fatalf("missing dep entry: %#v", lock.Packages)
	}
	if depPkg.Version != "0.2.0" {
		t.Fatalf("dep version unexpected: %#v", depPkg)
	}
	if !strings.HasPrefix(depPkg.Source, "path:") {
		t.Fatalf("expected path source, got %q", depPkg.Source)
	}
	if len(depPkg.Dependencies) != 0 {
		t.Fatalf("expected no transitive dependencies, got %#v", depPkg.Dependencies)
	}
	if stdlib := findLockedPackage(lock.Packages, "able"); stdlib == nil {
		t.Fatalf("missing stdlib entry: %#v", lock.Packages)
	}
	if kernel := findLockedPackage(lock.Packages, "kernel"); kernel == nil {
		t.Fatalf("missing kernel entry: %#v", lock.Packages)
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
	if len(lock.Packages) != 4 {
		t.Fatalf("expected four packages in lock, got %#v", lock.Packages)
	}
	if dep := findLockedPackage(lock.Packages, "dep"); dep == nil {
		t.Fatalf("expected dep package in lock")
	} else if len(dep.Dependencies) != 1 || dep.Dependencies[0].Name != "sub" {
		t.Fatalf("dep dependencies incorrect: %#v", dep.Dependencies)
	}
	if sub := findLockedPackage(lock.Packages, "sub"); sub == nil {
		t.Fatalf("expected sub package in lock")
	} else if len(sub.Dependencies) != 0 {
		t.Fatalf("sub should have no dependencies, got %#v", sub.Dependencies)
	}
	if stdlib := findLockedPackage(lock.Packages, "able"); stdlib == nil {
		t.Fatalf("expected stdlib package in lock")
	}
	if kernel := findLockedPackage(lock.Packages, "kernel"); kernel == nil {
		t.Fatalf("expected kernel package in lock")
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
	if len(lock.Packages) != 4 {
		t.Fatalf("expected four packages in lockfile, got %#v", lock.Packages)
	}
	mathPkg := findLockedPackage(lock.Packages, "math")
	helperPkg := findLockedPackage(lock.Packages, "helper")
	stdlibPkg := findLockedPackage(lock.Packages, "able")
	kernelPkg := findLockedPackage(lock.Packages, "kernel")
	if mathPkg == nil || helperPkg == nil || stdlibPkg == nil || kernelPkg == nil {
		t.Fatalf("missing expected lockfile entries: %#v", lock.Packages)
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
	if len(lock.Packages) != 3 {
		t.Fatalf("lock packages unexpected: %#v", lock.Packages)
	}
	pkg := findLockedPackage(lock.Packages, "gitpkg")
	if pkg == nil {
		t.Fatalf("missing gitpkg entry: %#v", lock.Packages)
	}
	expectedSource := fmt.Sprintf("git+%s@%s", repo, rev)
	if pkg.Source != expectedSource {
		t.Fatalf("pkg.Source = %q, want %q", pkg.Source, expectedSource)
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
	if stdlib := findLockedPackage(lock.Packages, "able"); stdlib == nil {
		t.Fatalf("missing stdlib entry: %#v", lock.Packages)
	}
	if kernel := findLockedPackage(lock.Packages, "kernel"); kernel == nil {
		t.Fatalf("missing kernel entry: %#v", lock.Packages)
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
	if len(lock.Packages) != 3 {
		t.Fatalf("lock packages unexpected: %#v", lock.Packages)
	}
	pkg := findLockedPackage(lock.Packages, "gitpkg")
	if pkg == nil {
		t.Fatalf("missing gitpkg entry: %#v", lock.Packages)
	}
	wantVersion := fmt.Sprintf("master@%s", rev)
	if pkg.Version != wantVersion {
		t.Fatalf("pkg.Version = %q, want %q", pkg.Version, wantVersion)
	}
	expectedSource := fmt.Sprintf("git+%s@%s", repo, rev)
	if pkg.Source != expectedSource {
		t.Fatalf("pkg.Source = %q, want %q", pkg.Source, expectedSource)
	}
	cached := filepath.Join(cacheDir, "pkg", "src", pkg.Name, sanitizePathSegment(pkg.Version))
	if _, err := os.Stat(cached); err != nil {
		t.Fatalf("expected cached git package at %s: %v", cached, err)
	}
	if stdlib := findLockedPackage(lock.Packages, "able"); stdlib == nil {
		t.Fatalf("missing stdlib entry: %#v", lock.Packages)
	}
	if kernel := findLockedPackage(lock.Packages, "kernel"); kernel == nil {
		t.Fatalf("missing kernel entry: %#v", lock.Packages)
	}
}

func TestDependencyInstaller_PinsBundledStdlib(t *testing.T) {
	root := t.TempDir()

	stdlibRoot := filepath.Join(root, "stdlib")
	stdlibSrc := filepath.Join(stdlibRoot, "src")
	if err := os.MkdirAll(stdlibSrc, 0o755); err != nil {
		t.Fatalf("mkdir stdlib: %v", err)
	}
	writeFile(t, filepath.Join(stdlibRoot, "package.yml"), `
name: able
version: 1.2.3
`)

	kernelRoot := filepath.Join(root, "kernel")
	kernelSrc := filepath.Join(kernelRoot, "src")
	if err := os.MkdirAll(kernelSrc, 0o755); err != nil {
		t.Fatalf("mkdir kernel: %v", err)
	}
	writeFile(t, filepath.Join(kernelRoot, "package.yml"), "name: kernel\n")

	appRoot := filepath.Join(root, "app")
	if err := os.MkdirAll(filepath.Join(appRoot, "src"), 0o755); err != nil {
		t.Fatalf("mkdir app: %v", err)
	}
	manifestPath := filepath.Join(appRoot, "package.yml")
	writeFile(t, manifestPath, `
name: sample
version: 0.0.1
`)

	manifest, err := driver.LoadManifest(manifestPath)
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}
	lock := driver.NewLockfile(manifest.Name, cliToolVersion)
	installer := newDependencyInstaller(manifest, filepath.Join(root, ".able"))

	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	defer func() {
		_ = os.Chdir(oldWD)
	}()
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir root: %v", err)
	}

	changed, logs, err := installer.Install(lock)
	if err != nil {
		t.Fatalf("Install returned error: %v (logs: %v)", err, logs)
	}
	if !changed {
		t.Fatalf("expected lockfile to include stdlib entry")
	}
	if len(lock.Packages) != 2 {
		t.Fatalf("expected stdlib and kernel entries, got %#v", lock.Packages)
	}
	stdlib := findLockedPackage(lock.Packages, "able")
	if stdlib == nil || stdlib.Version != "1.2.3" {
		t.Fatalf("unexpected stdlib lock entry: %#v", stdlib)
	}
	if stdlib.Source != fmt.Sprintf("path:%s", stdlibSrc) {
		t.Fatalf("expected stdlib source %s, got %s", stdlibSrc, stdlib.Source)
	}
	kernel := findLockedPackage(lock.Packages, "kernel")
	if kernel == nil || kernel.Source == "" {
		t.Fatalf("expected kernel entry, got %#v", kernel)
	}
	if kernel.Source != fmt.Sprintf("path:%s", kernelSrc) {
		t.Fatalf("expected kernel source %s, got %s", kernelSrc, kernel.Source)
	}
}
