package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"able/interpreter10-go/pkg/driver"
)

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

func TestCollectSearchPathsIncludesAbleModulePaths(t *testing.T) {
	tempDir := t.TempDir()
	extraOne := filepath.Join(tempDir, "depA")
	extraTwo := filepath.Join(tempDir, "depB")
	for _, dir := range []string{extraOne, extraTwo} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}

	joined := strings.Join([]string{extraOne, extraTwo}, string(os.PathListSeparator))
	t.Setenv("ABLE_MODULE_PATHS", joined)

	paths := collectSearchPaths()
	if !containsPath(paths, extraOne) || !containsPath(paths, extraTwo) {
		t.Fatalf("expected search paths to include %s and %s, got %v", extraOne, extraTwo, paths)
	}
}

func TestFindStdlibRootPrefersFlattenedLayout(t *testing.T) {
	tempDir := t.TempDir()
	repoDir := filepath.Join(tempDir, "repo")
	nested := filepath.Join(repoDir, "nested", "project")
	stdlibDir := filepath.Join(repoDir, "stdlib", "src")

	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("mkdir nested: %v", err)
	}
	if err := os.MkdirAll(stdlibDir, 0o755); err != nil {
		t.Fatalf("mkdir stdlib: %v", err)
	}

	found := findStdlibRoot(nested)
	if found != stdlibDir {
		t.Fatalf("expected findStdlibRoot to return %q, got %q", stdlibDir, found)
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
	if !strings.Contains(stderr, "package helper.core not found") &&
		!strings.Contains(stderr, "loader: package helper.core not found") &&
		!strings.Contains(stderr, "imports unknown package helper.core") {
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

func TestCheckCommandSucceeds(t *testing.T) {
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
}
`)

	code, stdout, stderr := captureCLI(t, []string{"check", "main.able"})
	if code != 0 {
		t.Fatalf("expected able check success, exit %d (stderr: %q)", code, stderr)
	}
	if !strings.Contains(stdout, "typecheck: ok") {
		t.Fatalf("expected typecheck success message, got stdout=%q", stdout)
	}
	if stderr != "" {
		t.Fatalf("expected no stderr output, got %q", stderr)
	}
}

func TestCheckCommandReportsDiagnostics(t *testing.T) {
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

	writeFile(t, filepath.Join(dir, "broken.able"), `
fn main() {
  value := 1 + "oops"
  value
}
`)

	code, stdout, stderr := captureCLI(t, []string{"check", "broken.able"})
	if code == 0 {
		t.Fatalf("expected able check failure for diagnostics, stdout=%q stderr=%q", stdout, stderr)
	}
	if stdout != "" {
		t.Fatalf("expected no stdout on failure, got %q", stdout)
	}
	if !strings.Contains(stderr, "requires numeric operands") {
		t.Fatalf("expected diagnostic in stderr, got %q", stderr)
	}
}
