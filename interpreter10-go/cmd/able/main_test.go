package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"able/interpreter10-go/pkg/driver"
	"able/interpreter10-go/pkg/interpreter"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
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

func containsPath(paths []string, target string) bool {
	for _, path := range paths {
		if path == target {
			return true
		}
	}
	return false
}

func TestRunEntryDirectFileNoManifest(t *testing.T) {
	dir := t.TempDir()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	defer func() {
		if chdirErr := os.Chdir(oldWD); chdirErr != nil {
			t.Fatalf("restore working directory: %v", chdirErr)
		}
	}()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}

	writeFile(t, filepath.Join(dir, "main.able"), `
fn main() {
  print("hello")
}
`)

	if code := runEntry([]string{"main.able"}); code != 0 {
		t.Fatalf("runEntry returned exit code %d, want 0", code)
	}
}

func TestRunEntryDirectFileWithManifest(t *testing.T) {
	dir := t.TempDir()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	defer func() {
		if chdirErr := os.Chdir(oldWD); chdirErr != nil {
			t.Fatalf("restore working directory: %v", chdirErr)
		}
	}()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}

	writeFile(t, filepath.Join(dir, "package.yml"), `
name: demo
targets:
  app: src/app.able
`)
	writeFile(t, filepath.Join(dir, "worker.able"), `
fn main() {
  print("worker")
}
`)

	if code := runEntry([]string{"worker.able"}); code != 0 {
		t.Fatalf("runEntry returned exit code %d, want 0", code)
	}
}

func TestRunShortcutAcceptsSourceFile(t *testing.T) {
	dir := t.TempDir()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	defer func() {
		if chdirErr := os.Chdir(oldWD); chdirErr != nil {
			t.Fatalf("restore working directory: %v", chdirErr)
		}
	}()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}

	writeFile(t, filepath.Join(dir, "solo.able"), `
fn main() {
  print("solo")
}
`)

	if code := run([]string{"solo.able"}); code != 0 {
		t.Fatalf("run returned exit code %d, want 0", code)
	}
}

func TestRunFileWithoutManifestStdlibAvailable(t *testing.T) {
	tempDir := t.TempDir()
	projectDir := filepath.Join(tempDir, "project")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("mkdir project: %v", err)
	}

	stdlibRoot := filepath.Join(tempDir, "stdlib")
	stdlibSrc := filepath.Join(stdlibRoot, "src")
	if err := os.MkdirAll(filepath.Join(stdlibSrc, "core"), 0o755); err != nil {
		t.Fatalf("mkdir stdlib core: %v", err)
	}

	writeFile(t, filepath.Join(stdlibRoot, "package.yml"), `
name: able
version: 0.0.1
`)
	writeFile(t, filepath.Join(stdlibSrc, "core", "thing.able"), `
package thing

fn stdlib_message() -> string {
  "std"
}
`)

	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	t.Setenv("ABLE_STD_LIB", stdlibSrc)

	defer func() {
		if chdirErr := os.Chdir(oldWD); chdirErr != nil {
			t.Fatalf("restore working directory: %v", chdirErr)
		}
	}()
	if err := os.Chdir(projectDir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}

	paths := collectSearchPaths()
	if !containsPath(paths, stdlibSrc) {
		t.Fatalf("expected search paths to include stdlib: %v", paths)
	}

	loader, err := driver.NewLoader([]string{stdlibSrc})
	if err != nil {
		t.Fatalf("NewLoader: %v", err)
	}
	defer loader.Close()
	if _, err := loader.Load(filepath.Join(stdlibSrc, "core", "thing.able")); err != nil {
		t.Fatalf("loader.Load stdlib stub: %v", err)
	}

	writeFile(t, filepath.Join(projectDir, "main.able"), `
package main

import able.core.thing as thing

fn main() {
  print(thing.stdlib_message())
}
`)

	code, stdout, stderr := captureCLI(t, []string{"main.able"})
	if code != 0 {
		t.Fatalf("run returned exit code %d, stderr: %q", code, stderr)
	}
	if !strings.Contains(stdout, "std") {
		t.Fatalf("expected stdout to contain std, got %q", stdout)
	}
}

func TestRunFileWithoutManifestMissingDependencyFails(t *testing.T) {
	tempDir := t.TempDir()
	projectDir := filepath.Join(tempDir, "project")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("mkdir project: %v", err)
	}

	stdlibRoot := filepath.Join(tempDir, "stdlib")
	stdlibSrc := filepath.Join(stdlibRoot, "src")
	if err := os.MkdirAll(filepath.Join(stdlibSrc, "core"), 0o755); err != nil {
		t.Fatalf("mkdir stdlib core: %v", err)
	}
	writeFile(t, filepath.Join(stdlibRoot, "package.yml"), `
name: able
version: 0.0.1
`)
	writeFile(t, filepath.Join(stdlibSrc, "core", "thing.able"), `
package thing

fn stdlib_message() -> string {
  "std"
}
`)

	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	t.Setenv("ABLE_STD_LIB", stdlibSrc)

	defer func() {
		if chdirErr := os.Chdir(oldWD); chdirErr != nil {
			t.Fatalf("restore working directory: %v", chdirErr)
		}
	}()
	if err := os.Chdir(projectDir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}

	writeFile(t, filepath.Join(projectDir, "main.able"), `
import helper.core as helper

fn main() {
  print(helper.value())
}
`)

	code, _, stderr := captureCLI(t, []string{"main.able"})
	if code == 0 {
		t.Fatalf("expected failure when dependency missing; stderr: %q", stderr)
	}
	if !strings.Contains(stderr, "package helper.core not found") && !strings.Contains(stderr, "loader: package helper.core not found") && !strings.Contains(stderr, "imports unknown package helper.core") {
		t.Fatalf("expected missing package error, got %q", stderr)
	}
}

func TestRunFileUsesEntryManifestLock(t *testing.T) {
	root := t.TempDir()
	manifestDir := filepath.Join(root, "foo")
	entryDir := filepath.Join(manifestDir, "bar")
	if err := os.MkdirAll(entryDir, 0o755); err != nil {
		t.Fatalf("mkdir entry dir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(manifestDir, "vendor", "helper"), 0o755); err != nil {
		t.Fatalf("mkdir helper vendor: %v", err)
	}

	writeFile(t, filepath.Join(manifestDir, "package.yml"), `
name: foo_app
targets:
  default: bar/baz.able
dependencies:
  helper:
    path: ./vendor/helper
`)
	writeFile(t, filepath.Join(entryDir, "baz.able"), `
fn main() {
  print("ran via manifest")
}
`)

	t.Setenv("ABLE_HOME", filepath.Join(root, "cache"))

	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	defer func() {
		if chdirErr := os.Chdir(oldWD); chdirErr != nil {
			t.Fatalf("restore working directory: %v", chdirErr)
		}
	}()
	if err := os.Chdir(root); err != nil {
		t.Fatalf("Chdir root: %v", err)
	}

	entryArg := filepath.Join("foo", "bar", "baz.able")
	if code, _, stderr := captureCLI(t, []string{entryArg}); code == 0 {
		t.Fatalf("expected failure without package.lock, stderr: %q", stderr)
	} else if !strings.Contains(stderr, "package.lock missing") {
		t.Fatalf("expected missing lockfile error, got %q", stderr)
	}

	lock := driver.NewLockfile("foo_app", cliToolVersion)
	lockPath := filepath.Join(manifestDir, "package.lock")
	if err := driver.WriteLockfile(lock, lockPath); err != nil {
		t.Fatalf("WriteLockfile: %v", err)
	}

	code, stdout, stderr := captureCLI(t, []string{entryArg})
	if code != 0 {
		t.Fatalf("expected success after lockfile write, exit %d (stderr: %q)", code, stderr)
	}
	if strings.Contains(stderr, "package.lock missing") {
		t.Fatalf("did not expect lockfile warning, got %q", stderr)
	}
	if !strings.Contains(stdout, "ran via manifest") {
		t.Fatalf("expected program output, got %q", stdout)
	}
}

func writeFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(strings.TrimSpace(contents)+"\n"), 0o644); err != nil {
		t.Fatalf("write file %s: %v", path, err)
	}
}

func initGitRepo(t *testing.T, dir string) string {
	t.Helper()
	repo, err := git.PlainInit(dir, false)
	if err != nil {
		t.Fatalf("PlainInit: %v", err)
	}
	worktree, err := repo.Worktree()
	if err != nil {
		t.Fatalf("Worktree: %v", err)
	}
	if err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if path == filepath.Join(dir, ".git") {
				return filepath.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		if strings.HasPrefix(rel, ".git/") {
			return nil
		}
		if _, err := worktree.Add(rel); err != nil {
			return err
		}
		return nil
	}); err != nil {
		t.Fatalf("stage files: %v", err)
	}
	hash, err := worktree.Commit("init", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Able CLI",
			Email: "able@example.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		t.Fatalf("Commit: %v", err)
	}
	return hash.String()
}

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

	for _, dir := range []string{filepath.Join(mainDir, "src"), filepath.Join(depDir, "src"), filepath.Join(subDir, "src")} {
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
	if err := os.MkdirAll(filepath.Join(mathRoot, "src"), 0o755); err != nil {
		t.Fatalf("mkdir math src: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(helperRoot, "src"), 0o755); err != nil {
		t.Fatalf("mkdir helper src: %v", err)
	}
	writeFile(t, filepath.Join(mathRoot, "src", "calc.able"), `package math
fn zero() -> i32 { 0 }
`)
	writeFile(t, filepath.Join(helperRoot, "src", "util.able"), `package helper
fn value() -> string { "helper" }
`)
	writeFile(t, filepath.Join(mathRoot, "package.yml"), `
name: math
version: 1.0.0
dependencies:
  helper: "0.5.0"
`)
	writeFile(t, filepath.Join(helperRoot, "package.yml"), `
name: helper
version: 0.5.0
`)

	t.Setenv("ABLE_REGISTRY", registry)

	mainDir := filepath.Join(root, "app")
	if err := os.MkdirAll(filepath.Join(mainDir, "src"), 0o755); err != nil {
		t.Fatalf("mkdir app: %v", err)
	}
	writeFile(t, filepath.Join(mainDir, "package.yml"), `
name: app
version: 0.1.0
dependencies:
  math: "1.0.0"
`)

	manifest, err := driver.LoadManifest(filepath.Join(mainDir, "package.yml"))
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}

	cacheDir := filepath.Join(root, "cache")
	installer := newDependencyInstaller(manifest, cacheDir)
	lock := driver.NewLockfile(manifest.Name, cliToolVersion)

	changed, logs, err := installer.Install(lock)
	if err != nil {
		t.Fatalf("Install error: %v", err)
	}
	if !changed {
		t.Fatalf("expected lockfile change for registry dependency")
	}
	if len(logs) == 0 {
		t.Fatalf("expected registry log entries")
	}
	if len(lock.Packages) != 2 {
		t.Fatalf("lock packages unexpected: %#v", lock.Packages)
	}
	helper := lock.Packages[0]
	math := lock.Packages[1]
	if helper.Name != "helper" || math.Name != "math" {
		t.Fatalf("unexpected package ordering: %#v", lock.Packages)
	}
	if math.Source == "" || !strings.HasPrefix(math.Source, "registry:") {
		t.Fatalf("expected registry source for math, got %q", math.Source)
	}
	helperCached := filepath.Join(cacheDir, "pkg", "src", helper.Name, sanitizePathSegment(helper.Version))
	if _, err := os.Stat(helperCached); err != nil {
		t.Fatalf("expected cached helper at %s: %v", helperCached, err)
	}
	mathCached := filepath.Join(cacheDir, "pkg", "src", math.Name, sanitizePathSegment(math.Version))
	if _, err := os.Stat(mathCached); err != nil {
		t.Fatalf("expected cached math at %s: %v", mathCached, err)
	}
	if len(math.Dependencies) != 1 || math.Dependencies[0].Name != "helper" {
		t.Fatalf("math dependencies incorrect: %#v", math.Dependencies)
	}
	if len(helper.Dependencies) != 0 {
		t.Fatalf("helper should have no dependencies, got %#v", helper.Dependencies)
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

func TestAbleDepsInstallAndRunWithCachedDependency(t *testing.T) {
	root := t.TempDir()
	registry := filepath.Join(root, "registry")
	pkgSrc := filepath.Join(registry, "default", "util", "1.0.0", "src")
	if err := os.MkdirAll(pkgSrc, 0o755); err != nil {
		t.Fatalf("mkdir registry src: %v", err)
	}
	writeFile(t, filepath.Join(pkgSrc, "package.yml"), `
name: util
version: 1.0.0
`)
	writeFile(t, filepath.Join(pkgSrc, "core.able"), `
package core

fn value() -> string {
  "util dependency"
}
`)

	project := filepath.Join(root, "app")
	if err := os.MkdirAll(filepath.Join(project, "src"), 0o755); err != nil {
		t.Fatalf("mkdir project src: %v", err)
	}
	writeFile(t, filepath.Join(project, "package.yml"), `
name: app
version: 0.1.0
targets:
  app: src/main.able
dependencies:
  util: "1.0.0"
`)
	writeFile(t, filepath.Join(project, "src", "main.able"), `
package main

import util.core as core

fn main() -> void {
  print(core.value())
}
`)

	cacheDir := filepath.Join(root, "cache")
	t.Setenv("ABLE_HOME", cacheDir)
	t.Setenv("ABLE_REGISTRY", registry)

	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	defer func() {
		_ = os.Chdir(originalWD)
	}()
	if err := os.Chdir(project); err != nil {
		t.Fatalf("Chdir project: %v", err)
	}

	if code, _, stderr := captureCLI(t, []string{"deps", "install"}); code != 0 {
		t.Fatalf("able deps install exited %d (stderr: %q)", code, stderr)
	}

	lockPath := filepath.Join(project, "package.lock")
	lock, err := driver.LoadLockfile(lockPath)
	if err != nil {
		t.Fatalf("LoadLockfile: %v", err)
	}
	if len(lock.Packages) != 1 {
		t.Fatalf("lock packages unexpected: %#v", lock.Packages)
	}
	pkg := lock.Packages[0]
	if pkg.Name != "util" || pkg.Version != "1.0.0" {
		t.Fatalf("lock entry unexpected: %#v", pkg)
	}
	expectedSource := fmt.Sprintf("registry:default/%s/%s", pkg.Name, pkg.Version)
	if pkg.Source != expectedSource {
		t.Fatalf("pkg.Source = %q, want %q", pkg.Source, expectedSource)
	}

	code, stdout, stderr := captureCLI(t, []string{"run"})
	if code != 0 {
		t.Fatalf("able run exited %d (stderr: %q)", code, stderr)
	}
	if !strings.Contains(stdout, "util dependency") {
		t.Fatalf("expected output to include dependency value, got %q", stdout)
	}
	if stderr != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}

	cached := filepath.Join(cacheDir, "pkg", "src", "util", "1.0.0")
	if _, err := os.Stat(cached); err != nil {
		t.Fatalf("expected cached dependency at %s: %v", cached, err)
	}
}

func TestAbleRunFailsOnTypecheckError(t *testing.T) {
	root := t.TempDir()
	registry := filepath.Join(root, "registry")
	pkgSrc := filepath.Join(registry, "default", "util", "1.0.0", "src")
	if err := os.MkdirAll(pkgSrc, 0o755); err != nil {
		t.Fatalf("mkdir registry src: %v", err)
	}
	writeFile(t, filepath.Join(pkgSrc, "package.yml"), `
name: util
version: 1.0.0
`)
	writeFile(t, filepath.Join(pkgSrc, "core.able"), `
package core

fn value() -> string {
  "util dependency"
}
`)

	project := filepath.Join(root, "app")
	if err := os.MkdirAll(filepath.Join(project, "src"), 0o755); err != nil {
		t.Fatalf("mkdir project src: %v", err)
	}
	writeFile(t, filepath.Join(project, "package.yml"), `
name: app
version: 0.1.0
targets:
  app: src/main.able
dependencies:
  util: "1.0.0"
`)
	writeFile(t, filepath.Join(project, "src", "main.able"), `
package main

import util.core as core

fn main() -> void {
  print(core.value())
}
`)

	cacheDir := filepath.Join(root, "cache")
	t.Setenv("ABLE_HOME", cacheDir)
	t.Setenv("ABLE_REGISTRY", registry)

	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	defer func() { _ = os.Chdir(originalWD) }()
	if err := os.Chdir(project); err != nil {
		t.Fatalf("Chdir project: %v", err)
	}

	if code, _, stderr := captureCLI(t, []string{"deps", "install"}); code != 0 {
		t.Fatalf("able deps install exited %d (stderr: %q)", code, stderr)
	}

	writeFile(t, filepath.Join(project, "src", "main.able"), `
package main

import util.core as core

fn main() -> void {
  print(core.value() + 1)
}
`)

	code, stdout, stderr := captureCLI(t, []string{"run"})
	if code == 0 {
		t.Fatalf("expected non-zero exit code for typecheck failure")
	}
	if stdout != "" {
		t.Fatalf("expected no stdout when typecheck fails, got %q", stdout)
	}
	if !strings.Contains(stderr, "typechecker:") {
		t.Fatalf("expected stderr to contain typechecker diagnostics, got %q", stderr)
	}
	if !strings.Contains(stderr, "package export summary") {
		t.Fatalf("expected stderr to include package export summary, got %q", stderr)
	}
}

func TestRunProgramTypecheckDetectsCrossPackageMismatch(t *testing.T) {
	root := t.TempDir()
	dep := t.TempDir()

	if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatalf("mkdir root src: %v", err)
	}
	writeFile(t, filepath.Join(root, "package.yml"), `
name: root_app
`)
	writeFile(t, filepath.Join(root, "src", "main.able"), `
import dep;

fn shout() -> string {
  dep.provide() + "!"
}

fn main() -> void {}
`)

	if err := os.MkdirAll(filepath.Join(dep, "src"), 0o755); err != nil {
		t.Fatalf("mkdir dep src: %v", err)
	}
	writeFile(t, filepath.Join(dep, "package.yml"), `
name: dep
`)
	writeFile(t, filepath.Join(dep, "src", "lib.able"), `
fn provide() -> i32 { 42 }
`)

	loader, err := driver.NewLoader([]string{filepath.Join(dep, "src")})
	if err != nil {
		t.Fatalf("NewLoader: %v", err)
	}
	defer loader.Close()

	entry := filepath.Join(root, "src", "main.able")
	program, err := loader.Load(entry)
	if err != nil {
		t.Fatalf("Load program: %v", err)
	}

	result, err := interpreter.TypecheckProgram(program)
	if err != nil {
		t.Fatalf("runProgramTypecheck returned error: %v", err)
	}
	if len(result.Diagnostics) == 0 {
		t.Fatalf("expected diagnostics from cross-package mismatch, got none")
	}
	if !strings.HasPrefix(result.Diagnostics[0].Package, "root_app") {
		t.Fatalf("expected diagnostic for root_app, got %s", result.Diagnostics[0].Package)
	}
	if want := "requires both operands"; !strings.Contains(result.Diagnostics[0].Diagnostic.Message, want) {
		t.Fatalf("expected diagnostic containing %q, got %q", want, result.Diagnostics[0].Diagnostic.Message)
	}
	if summary, ok := result.Packages["dep"]; !ok {
		t.Fatalf("expected summary for dep")
	} else if summary.Functions == nil {
		t.Fatalf("expected functions map in summary")
	}
}

func TestRunProgramTypecheckSucceedsWithImports(t *testing.T) {
	root := t.TempDir()
	dep := t.TempDir()

	if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatalf("mkdir root src: %v", err)
	}
	writeFile(t, filepath.Join(root, "package.yml"), `
name: root_app
`)
	writeFile(t, filepath.Join(root, "src", "main.able"), `
import dep;

fn shout() -> string {
  dep.provide() + "!"
}

fn main() -> void {}
`)

	if err := os.MkdirAll(filepath.Join(dep, "src"), 0o755); err != nil {
		t.Fatalf("mkdir dep src: %v", err)
	}
	writeFile(t, filepath.Join(dep, "package.yml"), `
name: dep
`)
	writeFile(t, filepath.Join(dep, "src", "lib.able"), `
fn provide() -> string { "ok" }
`)

	loader, err := driver.NewLoader([]string{filepath.Join(dep, "src")})
	if err != nil {
		t.Fatalf("NewLoader: %v", err)
	}
	defer loader.Close()

	entry := filepath.Join(root, "src", "main.able")
	program, err := loader.Load(entry)
	if err != nil {
		t.Fatalf("Load program: %v", err)
	}

	result, err := interpreter.TypecheckProgram(program)
	if err != nil {
		t.Fatalf("runProgramTypecheck returned error: %v", err)
	}
	if len(result.Diagnostics) != 0 {
		t.Fatalf("expected no diagnostics, got %v", result.Diagnostics)
	}
	if summary, ok := result.Packages["dep"]; !ok || summary.Functions == nil {
		t.Fatalf("expected exported function summary for dep")
	}
}

func TestAbleDepsInstallResolvesTransitivePathDependencies(t *testing.T) {
	root := t.TempDir()
	project := filepath.Join(root, "app")
	dep1 := filepath.Join(project, "deps", "dep1")
	dep2 := filepath.Join(project, "deps", "dep2")

	for _, dir := range []string{
		filepath.Join(project, "src"),
		filepath.Join(dep1, "src"),
		filepath.Join(dep2, "src"),
	} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}

	writeFile(t, filepath.Join(project, "package.yml"), `
name: app
version: 0.1.0
targets:
  app: src/main.able
dependencies:
  dep1:
    path: deps/dep1
`)
	writeFile(t, filepath.Join(project, "src", "main.able"), `
package main

import dep1.src.core as dep1

fn main() -> void {
  print(dep1.value())
}
`)

	writeFile(t, filepath.Join(dep1, "package.yml"), `
name: dep1
version: 1.0.0
dependencies:
  dep2:
    path: ../dep2
`)
	writeFile(t, filepath.Join(dep1, "src", "core.able"), `
package core

import dep2.src.core as dep2

fn value() -> string {
  dep2.value()
}
`)

	writeFile(t, filepath.Join(dep2, "package.yml"), `
name: dep2
version: 2.0.0
`)
	writeFile(t, filepath.Join(dep2, "src", "core.able"), `
package core

fn value() -> string {
  "transitive path value"
}
`)

	cacheDir := filepath.Join(root, "cache")
	t.Setenv("ABLE_HOME", cacheDir)

	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	defer func() { _ = os.Chdir(originalWD) }()
	if err := os.Chdir(project); err != nil {
		t.Fatalf("Chdir project: %v", err)
	}

	if code, _, stderr := captureCLI(t, []string{"deps", "install"}); code != 0 {
		t.Fatalf("able deps install exited %d (stderr: %q)", code, stderr)
	}

	lockPath := filepath.Join(project, "package.lock")
	lock, err := driver.LoadLockfile(lockPath)
	if err != nil {
		t.Fatalf("LoadLockfile: %v", err)
	}
	if len(lock.Packages) != 2 {
		t.Fatalf("expected 2 packages in lockfile, got %#v", lock.Packages)
	}
	pkgs := make(map[string]*driver.LockedPackage, len(lock.Packages))
	for _, pkg := range lock.Packages {
		if pkg == nil {
			continue
		}
		pkgs[pkg.Name] = pkg
	}
	dep1Lock, ok := pkgs["dep1"]
	if !ok {
		t.Fatalf("dep1 missing from lockfile: %#v", pkgs)
	}
	if len(dep1Lock.Dependencies) != 1 || dep1Lock.Dependencies[0].Name != "dep2" || dep1Lock.Dependencies[0].Version != "2.0.0" {
		t.Fatalf("dep1 dependency list unexpected: %#v", dep1Lock.Dependencies)
	}
	dep2Lock, ok := pkgs["dep2"]
	if !ok {
		t.Fatalf("dep2 missing from lockfile: %#v", pkgs)
	}
	if len(dep2Lock.Dependencies) != 0 {
		t.Fatalf("dep2 should have no dependencies, got %#v", dep2Lock.Dependencies)
	}

	code, stdout, stderr := captureCLI(t, []string{"run"})
	if code != 0 {
		t.Fatalf("able run exited %d (stderr: %q)", code, stderr)
	}
	if want := "transitive path value"; !strings.Contains(stdout, want) {
		t.Fatalf("expected stdout to contain %q, got %q", want, stdout)
	}
}

func TestAbleDepsInstallResolvesTransitiveRegistryDependencies(t *testing.T) {
	root := t.TempDir()
	project := filepath.Join(root, "app")
	registry := filepath.Join(root, "registry")

	for _, dir := range []string{
		filepath.Join(project, "src"),
		filepath.Join(registry, "default", "util", "1.0.0", "src"),
		filepath.Join(registry, "default", "helper", "1.0.0", "src"),
	} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}

	writeFile(t, filepath.Join(project, "package.yml"), `
name: app
version: 0.1.0
targets:
  app: src/main.able
dependencies:
  util: "1.0.0"
`)
	writeFile(t, filepath.Join(project, "src", "main.able"), `
package main

import util.core as util

fn main() -> void {
  print(util.value())
}
`)

	writeFile(t, filepath.Join(registry, "default", "helper", "1.0.0", "package.yml"), `
name: helper
version: 1.0.0
`)
	writeFile(t, filepath.Join(registry, "default", "helper", "1.0.0", "src", "core.able"), `
package core

fn value() -> string {
  "helper"
}
`)

	writeFile(t, filepath.Join(registry, "default", "util", "1.0.0", "package.yml"), `
name: util
version: 1.0.0
dependencies:
  helper: "1.0.0"
`)
	writeFile(t, filepath.Join(registry, "default", "util", "1.0.0", "src", "core.able"), `
package core

import helper.core as helper

fn value() -> string {
  helper.value() + " util"
}
`)

	cacheDir := filepath.Join(root, "cache")
	t.Setenv("ABLE_HOME", cacheDir)
	t.Setenv("ABLE_REGISTRY", registry)

	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	defer func() { _ = os.Chdir(originalWD) }()
	if err := os.Chdir(project); err != nil {
		t.Fatalf("Chdir project: %v", err)
	}

	if code, _, stderr := captureCLI(t, []string{"deps", "install"}); code != 0 {
		t.Fatalf("able deps install exited %d (stderr: %q)", code, stderr)
	}

	lockPath := filepath.Join(project, "package.lock")
	lock, err := driver.LoadLockfile(lockPath)
	if err != nil {
		t.Fatalf("LoadLockfile: %v", err)
	}
	if len(lock.Packages) != 2 {
		t.Fatalf("expected 2 packages in lockfile, got %#v", lock.Packages)
	}
	pkgs := make(map[string]*driver.LockedPackage, len(lock.Packages))
	for _, pkg := range lock.Packages {
		if pkg == nil {
			continue
		}
		pkgs[pkg.Name] = pkg
	}
	utilLock, ok := pkgs["util"]
	if !ok {
		t.Fatalf("util missing from lockfile: %#v", pkgs)
	}
	if len(utilLock.Dependencies) != 1 || utilLock.Dependencies[0].Name != "helper" || utilLock.Dependencies[0].Version != "1.0.0" {
		t.Fatalf("util dependency list unexpected: %#v", utilLock.Dependencies)
	}
	helperLock, ok := pkgs["helper"]
	if !ok {
		t.Fatalf("helper missing from lockfile: %#v", pkgs)
	}
	if len(helperLock.Dependencies) != 0 {
		t.Fatalf("helper should have no dependencies, got %#v", helperLock.Dependencies)
	}

	cacheManifest := filepath.Join(cacheDir, "pkg", "src", "util", "1.0.0", "package.yml")
	if _, err := os.Stat(cacheManifest); err != nil {
		t.Fatalf("expected cached manifest at %s: %v", cacheManifest, err)
	}

	code, stdout, stderr := captureCLI(t, []string{"run"})
	if code != 0 {
		t.Fatalf("able run exited %d (stderr: %q)", code, stderr)
	}
	if want := "helper util"; !strings.Contains(stdout, want) {
		t.Fatalf("expected stdout to contain %q, got %q", want, stdout)
	}
}

func TestAbleDepsInstallResolvesGitTransitiveDependencies(t *testing.T) {
	root := t.TempDir()
	project := filepath.Join(root, "app")
	repo := filepath.Join(root, "gitdep")

	for _, dir := range []string{
		filepath.Join(project, "src"),
		filepath.Join(repo, "src"),
		filepath.Join(repo, "vendor", "helper", "src"),
	} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}

	writeFile(t, filepath.Join(project, "package.yml"), `
name: app
version: 0.1.0
targets:
  app: src/main.able
dependencies:
  gitdep:
    git: `+repo+`
    rev: PLACEHOLDER
`)
	writeFile(t, filepath.Join(project, "src", "main.able"), `
package main

import gitdep.src.core as gitdep

fn main() -> void {
  print(gitdep.value())
}
`)

	writeFile(t, filepath.Join(repo, "package.yml"), `
name: gitdep
version: 1.0.0
dependencies:
  helper:
    path: vendor/helper
`)
	writeFile(t, filepath.Join(repo, "src", "core.able"), `
package core

import helper.src.core as helper

fn value() -> string {
  helper.value() + " from git"
}
`)
	writeFile(t, filepath.Join(repo, "vendor", "helper", "package.yml"), `
name: helper
version: 0.9.0
`)
	writeFile(t, filepath.Join(repo, "vendor", "helper", "src", "core.able"), `
package core

fn value() -> string {
  "helper"
}
`)

	rev := initGitRepo(t, repo)

	// Patch manifest with actual revision.
	writeFile(t, filepath.Join(project, "package.yml"), `
name: app
version: 0.1.0
targets:
  app: src/main.able
dependencies:
  gitdep:
    git: `+repo+`
    rev: `+rev+`
`)

	cacheDir := filepath.Join(root, "cache")
	t.Setenv("ABLE_HOME", cacheDir)

	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	defer func() { _ = os.Chdir(originalWD) }()
	if err := os.Chdir(project); err != nil {
		t.Fatalf("Chdir project: %v", err)
	}

	if code, _, stderr := captureCLI(t, []string{"deps", "install"}); code != 0 {
		t.Fatalf("able deps install exited %d (stderr: %q)", code, stderr)
	}

	lockPath := filepath.Join(project, "package.lock")
	lock, err := driver.LoadLockfile(lockPath)
	if err != nil {
		t.Fatalf("LoadLockfile: %v", err)
	}
	if len(lock.Packages) != 2 {
		t.Fatalf("expected 2 packages in lockfile, got %#v", lock.Packages)
	}
	pkgs := make(map[string]*driver.LockedPackage, len(lock.Packages))
	for _, pkg := range lock.Packages {
		if pkg == nil {
			continue
		}
		pkgs[pkg.Name] = pkg
	}
	gitLock, ok := pkgs["gitdep"]
	if !ok {
		t.Fatalf("gitdep missing from lockfile: %#v", pkgs)
	}
	if len(gitLock.Dependencies) != 1 || gitLock.Dependencies[0].Name != "helper" || gitLock.Dependencies[0].Version != "0.9.0" {
		t.Fatalf("gitdep dependency list unexpected: %#v", gitLock.Dependencies)
	}
	if expected := fmt.Sprintf("git+%s@%s", repo, rev); gitLock.Source != expected {
		t.Fatalf("gitdep source = %q, want %q", gitLock.Source, expected)
	}
	helperLock, ok := pkgs["helper"]
	if !ok {
		t.Fatalf("helper missing from lockfile: %#v", pkgs)
	}
	if len(helperLock.Dependencies) != 0 {
		t.Fatalf("helper should have no dependencies, got %#v", helperLock.Dependencies)
	}

	code, stdout, stderr := captureCLI(t, []string{"run"})
	if code != 0 {
		t.Fatalf("able run exited %d (stderr: %q)", code, stderr)
	}
	if want := "helper from git"; !strings.Contains(stdout, want) {
		t.Fatalf("expected stdout to contain %q, got %q", want, stdout)
	}
}

func TestAbleDepsUpdateAll(t *testing.T) {
	root := t.TempDir()
	registry := filepath.Join(root, "registry")
	project := filepath.Join(root, "app")

	if err := os.MkdirAll(filepath.Join(project, "src"), 0o755); err != nil {
		t.Fatalf("mkdir project src: %v", err)
	}
	for _, dir := range []string{
		filepath.Join(registry, "default", "util", "1.0.0", "src"),
		filepath.Join(registry, "default", "util", "1.1.0", "src"),
	} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir registry dir: %v", err)
		}
	}

	writeFile(t, filepath.Join(registry, "default", "util", "1.0.0", "package.yml"), `
name: util
version: 1.0.0
`)
	writeFile(t, filepath.Join(registry, "default", "util", "1.0.0", "src", "core.able"), `
package core

fn value() -> string {
  "util v1.0"
}
`)

	writeFile(t, filepath.Join(registry, "default", "util", "1.1.0", "package.yml"), `
name: util
version: 1.1.0
`)
	writeFile(t, filepath.Join(registry, "default", "util", "1.1.0", "src", "core.able"), `
package core

fn value() -> string {
  "util v1.1"
}
`)

	writeFile(t, filepath.Join(project, "package.yml"), `
name: app
version: 0.1.0
targets:
  app: src/main.able
dependencies:
  util: "1.0.0"
`)
	writeFile(t, filepath.Join(project, "src", "main.able"), `
package main

import util.core as util

fn main() -> void {
  print(util.value())
}
`)

	cacheDir := filepath.Join(root, "cache")
	t.Setenv("ABLE_HOME", cacheDir)
	t.Setenv("ABLE_REGISTRY", registry)

	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	defer func() { _ = os.Chdir(originalWD) }()
	if err := os.Chdir(project); err != nil {
		t.Fatalf("Chdir project: %v", err)
	}

	if code, _, stderr := captureCLI(t, []string{"deps", "install"}); code != 0 {
		t.Fatalf("able deps install exited %d (stderr: %q)", code, stderr)
	}

	writeFile(t, filepath.Join(project, "package.yml"), `
name: app
version: 0.1.0
targets:
  app: src/main.able
dependencies:
  util: "1.1.0"
`)

	code, _, stderr := captureCLI(t, []string{"deps", "update"})
	if code != 0 {
		t.Fatalf("able deps update exited %d (stderr: %q)", code, stderr)
	}

	lockPath := filepath.Join(project, "package.lock")
	lock, err := driver.LoadLockfile(lockPath)
	if err != nil {
		t.Fatalf("LoadLockfile: %v", err)
	}
	if len(lock.Packages) != 1 {
		t.Fatalf("lock packages unexpected: %#v", lock.Packages)
	}
	pkg := lock.Packages[0]
	if pkg.Name != "util" || pkg.Version != "1.1.0" {
		t.Fatalf("lock entry unexpected: %#v", pkg)
	}

	code, stdout, stderr := captureCLI(t, []string{"run"})
	if code != 0 {
		t.Fatalf("able run exited %d (stderr: %q)", code, stderr)
	}
	if want := "util v1.1"; !strings.Contains(stdout, want) {
		t.Fatalf("expected stdout to contain %q, got %q", want, stdout)
	}
}

func TestAbleDepsUpdateSpecificDependency(t *testing.T) {
	root := t.TempDir()
	registry := filepath.Join(root, "registry")
	project := filepath.Join(root, "app")

	for _, dir := range []string{
		filepath.Join(project, "src"),
		filepath.Join(registry, "default", "util", "1.0.0", "src"),
		filepath.Join(registry, "default", "util", "2.0.0", "src"),
		filepath.Join(registry, "default", "helper", "1.0.0", "src"),
	} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}

	writeFile(t, filepath.Join(registry, "default", "helper", "1.0.0", "package.yml"), `
name: helper
version: 1.0.0
`)
	writeFile(t, filepath.Join(registry, "default", "helper", "1.0.0", "src", "core.able"), `
package core

fn helper_value() -> string {
  "helper v1.0"
}
`)

	writeFile(t, filepath.Join(registry, "default", "util", "1.0.0", "package.yml"), `
name: util
version: 1.0.0
`)
	writeFile(t, filepath.Join(registry, "default", "util", "1.0.0", "src", "core.able"), `
package core

fn value() -> string {
  "util v1.0"
}
`)

	writeFile(t, filepath.Join(registry, "default", "util", "2.0.0", "package.yml"), `
name: util
version: 2.0.0
`)
	writeFile(t, filepath.Join(registry, "default", "util", "2.0.0", "src", "core.able"), `
package core

fn value() -> string {
  "util v2.0"
}
`)

	writeFile(t, filepath.Join(project, "package.yml"), `
name: app
version: 0.1.0
targets:
  app: src/main.able
dependencies:
  helper: "1.0.0"
  util: "1.0.0"
`)
	writeFile(t, filepath.Join(project, "src", "main.able"), `
package main

import helper.core as helper
import util.core as util

fn main() -> void {
  print(helper.helper_value() + " & " + util.value())
}
`)

	cacheDir := filepath.Join(root, "cache")
	t.Setenv("ABLE_HOME", cacheDir)
	t.Setenv("ABLE_REGISTRY", registry)

	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	defer func() { _ = os.Chdir(originalWD) }()
	if err := os.Chdir(project); err != nil {
		t.Fatalf("Chdir project: %v", err)
	}

	if code, _, stderr := captureCLI(t, []string{"deps", "install"}); code != 0 {
		t.Fatalf("able deps install exited %d (stderr: %q)", code, stderr)
	}

	writeFile(t, filepath.Join(project, "package.yml"), `
name: app
version: 0.1.0
targets:
  app: src/main.able
dependencies:
  helper: "1.0.0"
  util: "2.0.0"
`)

	code, _, stderr := captureCLI(t, []string{"deps", "update", "util"})
	if code != 0 {
		t.Fatalf("able deps update exited %d (stderr: %q)", code, stderr)
	}

	lockPath := filepath.Join(project, "package.lock")
	lock, err := driver.LoadLockfile(lockPath)
	if err != nil {
		t.Fatalf("LoadLockfile: %v", err)
	}
	if len(lock.Packages) != 2 {
		t.Fatalf("lock packages unexpected: %#v", lock.Packages)
	}
	versions := make(map[string]string, 2)
	for _, pkg := range lock.Packages {
		if pkg == nil {
			continue
		}
		versions[pkg.Name] = pkg.Version
	}
	if versions["util"] != "2.0.0" {
		t.Fatalf("expected util@2.0.0, got %v", versions["util"])
	}
	if versions["helper"] != "1.0.0" {
		t.Fatalf("expected helper@1.0.0, got %v", versions["helper"])
	}

	code, _, stderr = captureCLI(t, []string{"run"})
	if code != 0 {
		t.Fatalf("able run exited %d (stderr: %q)", code, stderr)
	}
	// Output may vary based on future formatting changes; ensure the run succeeded.
}

func TestAbleRunUsesCachedRegistryDependencyWhenOffline(t *testing.T) {
	root := t.TempDir()
	registry := filepath.Join(root, "registry")
	project := filepath.Join(root, "app")

	for _, dir := range []string{
		filepath.Join(project, "src"),
		filepath.Join(registry, "default", "util", "1.0.0", "src"),
	} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}

	writeFile(t, filepath.Join(registry, "default", "util", "1.0.0", "package.yml"), `
name: util
version: 1.0.0
`)
	writeFile(t, filepath.Join(registry, "default", "util", "1.0.0", "src", "core.able"), `
package core

fn value() -> string {
  "cached util"
}
`)

	writeFile(t, filepath.Join(project, "package.yml"), `
name: app
version: 0.1.0
targets:
  app: src/main.able
dependencies:
  util: "1.0.0"
`)
	writeFile(t, filepath.Join(project, "src", "main.able"), `
package main

import util.core as util

fn main() -> void {
  print(util.value())
}
`)

	cacheDir := filepath.Join(root, "cache")
	t.Setenv("ABLE_HOME", cacheDir)
	t.Setenv("ABLE_REGISTRY", registry)

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	defer func() { _ = os.Chdir(wd) }()
	if err := os.Chdir(project); err != nil {
		t.Fatalf("Chdir project: %v", err)
	}

	if code, _, stderr := captureCLI(t, []string{"deps", "install"}); code != 0 {
		t.Fatalf("able deps install exited %d (stderr: %q)", code, stderr)
	}

	lockPath := filepath.Join(project, "package.lock")
	lock, err := driver.LoadLockfile(lockPath)
	if err != nil {
		t.Fatalf("LoadLockfile: %v", err)
	}
	if len(lock.Packages) != 1 {
		t.Fatalf("expected single package in lockfile, got %#v", lock.Packages)
	}
	if err := os.RemoveAll(registry); err != nil {
		t.Fatalf("RemoveAll registry: %v", err)
	}

	code, stdout, stderr := captureCLI(t, []string{"run"})
	if code != 0 {
		t.Fatalf("able run exited %d (stderr: %q)", code, stderr)
	}
	if want := "cached util"; !strings.Contains(stdout, want) {
		t.Fatalf("expected stdout to contain %q, got %q", want, stdout)
	}
}

func captureCLI(t *testing.T, args []string) (int, string, string) {
	t.Helper()

	stdout := os.Stdout
	stderr := os.Stderr

	rOut, wOut, err := os.Pipe()
	if err != nil {
		t.Fatalf("stdout pipe: %v", err)
	}
	rErr, wErr, err := os.Pipe()
	if err != nil {
		t.Fatalf("stderr pipe: %v", err)
	}

	os.Stdout = wOut
	os.Stderr = wErr

	code := run(args)

	if err := wOut.Close(); err != nil {
		t.Fatalf("stdout close: %v", err)
	}
	if err := wErr.Close(); err != nil {
		t.Fatalf("stderr close: %v", err)
	}

	os.Stdout = stdout
	os.Stderr = stderr

	outBytes, err := io.ReadAll(rOut)
	if err != nil {
		t.Fatalf("stdout read: %v", err)
	}
	errBytes, err := io.ReadAll(rErr)
	if err != nil {
		t.Fatalf("stderr read: %v", err)
	}

	if err := rOut.Close(); err != nil {
		t.Fatalf("stdout pipe close: %v", err)
	}
	if err := rErr.Close(); err != nil {
		t.Fatalf("stderr pipe close: %v", err)
	}

	return code, string(outBytes), string(errBytes)
}
