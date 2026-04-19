package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"able/interpreter-go/pkg/driver"
)

func TestRunEntryDirectFileNoManifest(t *testing.T) {
	dir := t.TempDir()
	enterWorkingDir(t, dir)

	writeFile(t, filepath.Join(dir, "main.able"), `
fn main() {
  print("hello")
}
`)

	if code := runEntry([]string{"main.able"}, interpreterTreewalker); code != 0 {
		t.Fatalf("runEntry returned exit code %d, want 0", code)
	}
}

func TestRunEntryDirectFileWithManifest(t *testing.T) {
	dir := t.TempDir()
	enterWorkingDir(t, dir)

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

	if code := runEntry([]string{"worker.able"}, interpreterTreewalker); code != 0 {
		t.Fatalf("runEntry returned exit code %d, want 0", code)
	}
}

func TestRunShortcutAcceptsSourceFile(t *testing.T) {
	dir := t.TempDir()
	enterWorkingDir(t, dir)

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
	writeFile(t, filepath.Join(stdlibSrc, "package.yml"), `
name: able
version: 0.0.1
`)
	writeFile(t, filepath.Join(stdlibSrc, "core", "thing.able"), `
package thing

fn stdlib_message() -> string {
  "std"
}
`)

	t.Setenv("ABLE_MODULE_PATHS", stdlibSrc)
	enterWorkingDir(t, projectDir)

	paths := collectSearchPaths(tempDir, searchPathOptions{})
	if !containsSearchPath(paths, stdlibSrc) && !containsSearchPath(paths, stdlibRoot) {
		t.Fatalf("expected search paths to include stdlib: %v", paths)
	}

	loader, err := driver.NewLoader([]driver.SearchPath{{Path: stdlibSrc, Kind: driver.RootStdlib}})
	if err != nil {
		t.Fatalf("NewLoader: %v", err)
	}
	defer loader.Close()
	if _, err := loader.Load(filepath.Join(stdlibSrc, "core", "thing.able")); err != nil {
		t.Fatalf("loader.Load stdlib stub: %v", err)
	}

	writeFile(t, filepath.Join(projectDir, "main.able"), `
package main

import able.core.thing::thing

fn main() {
  print(thing.stdlib_message())
}
`)

	stdout := runCLIExpectSuccess(t, "main.able")
	assertOutputContainsAll(t, stdout, "std")
}

func TestRunIgnoresTestModulesUnlessWithTests(t *testing.T) {
	dir := t.TempDir()
	enterWorkingDir(t, dir)

	writeFile(t, filepath.Join(dir, "package.yml"), `
name: demo
targets:
  app: main.able
`)
	writeFile(t, filepath.Join(dir, "main.able"), `
fn main() {
  print("main")
}
`)
	writeFile(t, filepath.Join(dir, "side.test.able"), `
print("test")
`)

	stdout := runCLIExpectSuccess(t, "run")
	if strings.Contains(stdout, "test") {
		t.Fatalf("expected test modules to be skipped, got stdout %q", stdout)
	}
	assertOutputContainsAll(t, stdout, "main")

	stdout = runCLIExpectSuccess(t, "run", "--with-tests")
	assertOutputContainsAll(t, stdout, "test", "main")
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

	paths := collectSearchPaths(tempDir, searchPathOptions{})
	if !containsSearchPath(paths, extraOne) || !containsSearchPath(paths, extraTwo) {
		t.Fatalf("expected search paths to include %s and %s, got %v", extraOne, extraTwo, paths)
	}
}

func TestFindStdlibRootsPreferFlattenedLayout(t *testing.T) {
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

	roots := findStdlibRoots(nested)
	if len(roots) == 0 || roots[0] != stdlibDir {
		t.Fatalf("expected findStdlibRoots to return %q first, got %v", stdlibDir, roots)
	}
}

func TestFindStdlibRootsDetectsAbleStdlibLayout(t *testing.T) {
	tempDir := t.TempDir()
	repoRoot := filepath.Join(tempDir, "repo")
	workDir := filepath.Join(repoRoot, "workspace")
	stdlibDir := filepath.Join(repoRoot, "able-stdlib", "src")

	if err := os.MkdirAll(workDir, 0o755); err != nil {
		t.Fatalf("mkdir workspace: %v", err)
	}
	if err := os.MkdirAll(stdlibDir, 0o755); err != nil {
		t.Fatalf("mkdir stdlib: %v", err)
	}

	roots := findStdlibRoots(workDir)
	found := false
	for _, root := range roots {
		if filepath.Clean(root) == filepath.Clean(stdlibDir) {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected findStdlibRoots to include %q, got %v", stdlibDir, roots)
	}
}

func TestFindKernelRootsDetectsV12Layout(t *testing.T) {
	tempDir := t.TempDir()
	repoRoot := filepath.Join(tempDir, "repo")
	workDir := filepath.Join(repoRoot, "workspace")
	kernelDir := filepath.Join(repoRoot, "v12", "kernel", "src")

	if err := os.MkdirAll(workDir, 0o755); err != nil {
		t.Fatalf("mkdir workspace: %v", err)
	}
	if err := os.MkdirAll(kernelDir, 0o755); err != nil {
		t.Fatalf("mkdir kernel: %v", err)
	}

	roots := findKernelRoots(workDir)
	found := false
	for _, root := range roots {
		if filepath.Clean(root) == filepath.Clean(kernelDir) {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected findKernelRoots to include %q, got %v", kernelDir, roots)
	}
}

func TestRunFileAutoDetectsCachedStdlib(t *testing.T) {
	root := t.TempDir()
	cacheDir := filepath.Join(root, "cache")
	stdlibSrc := filepath.Join(cacheDir, "pkg", "src", "able", defaultStdlibVersion, "src")
	stdlibRoot := filepath.Join(cacheDir, "pkg", "src", "able", defaultStdlibVersion)
	appRoot := filepath.Join(root, "app")

	if err := os.MkdirAll(stdlibSrc, 0o755); err != nil {
		t.Fatalf("mkdir stdlib: %v", err)
	}
	if err := os.MkdirAll(appRoot, 0o755); err != nil {
		t.Fatalf("mkdir app: %v", err)
	}

	writeFile(t, filepath.Join(stdlibRoot, "package.yml"), "name: able\nversion: "+defaultStdlibVersion+"\n")
	writeFile(t, filepath.Join(stdlibSrc, "custom.able"), `
package custom

fn greeting() -> string { "hello from cached stdlib" }
`)

	writeFile(t, filepath.Join(appRoot, "main.able"), `
package main

import able.custom.{greeting}

fn main() {
  print(greeting())
}
`)

	t.Setenv("ABLE_HOME", cacheDir)

	enterWorkingDir(t, appRoot)
	stdout := runCLIExpectSuccess(t, "main.able")
	assertOutputContainsAll(t, stdout, "hello from cached stdlib")
}

func TestRunUsesManifestLockForStdlibAndKernel(t *testing.T) {
	root := t.TempDir()
	depsRoot := filepath.Join(root, "deps")
	stdlibSrc := filepath.Join(depsRoot, "stdlib", "src")
	kernelSrc := filepath.Join(depsRoot, "kernel", "src")
	appRoot := filepath.Join(root, "app")

	for _, dir := range []string{
		stdlibSrc,
		kernelSrc,
		filepath.Join(appRoot, "src"),
	} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}

	writeFile(t, filepath.Join(depsRoot, "stdlib", "package.yml"), "name: able\nversion: 9.9.9\n")
	writeFile(t, filepath.Join(stdlibSrc, "locktest.able"), `
package locktest

fn greeting() -> string { "hello from locked stdlib" }
`)

	writeFile(t, filepath.Join(depsRoot, "kernel", "package.yml"), "name: kernel\nversion: 1.0.0\n")
	writeFile(t, filepath.Join(kernelSrc, "boot.able"), `
package boot

fn kernel_ready() -> bool { true }
`)

	writeFile(t, filepath.Join(appRoot, "package.yml"), `
name: sample
version: 0.0.1
dependencies:
  able: "9.9.9"
targets:
  app: src/main.able
`)
	writeFile(t, filepath.Join(appRoot, "package.lock"), `
root: sample
packages:
  - name: able
    version: 9.9.9
    source: path:../deps/stdlib/src
  - name: kernel
    version: 1.0.0
    source: path:../deps/kernel/src
`)
	writeFile(t, filepath.Join(appRoot, "src", "main.able"), `
package main

import able.locktest.{greeting}
import able.kernel.boot.{kernel_ready}

fn main() {
  if kernel_ready() {
    print(greeting())
  }
}
`)

	enterWorkingDir(t, appRoot)
	stdout := runCLIExpectSuccess(t, "run")
	assertOutputContainsAll(t, stdout, "hello from locked stdlib")
}

func TestRunRejectsManifestLockStdlibCollisionWithEnvRoot(t *testing.T) {
	root := t.TempDir()
	depsRoot := filepath.Join(root, "deps")
	stdlibSrc := filepath.Join(depsRoot, "stdlib", "src")
	conflictRoot := filepath.Join(root, "conflict-stdlib")
	conflictSrc := filepath.Join(conflictRoot, "src")
	appRoot := filepath.Join(root, "app")

	for _, dir := range []string{
		stdlibSrc,
		conflictSrc,
		filepath.Join(appRoot, "src"),
	} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}

	writeFile(t, filepath.Join(depsRoot, "stdlib", "package.yml"), "name: able\nversion: 9.9.9\n")
	writeFile(t, filepath.Join(stdlibSrc, "locktest.able"), `
package locktest

fn greeting() -> string { "hello from locked stdlib" }
`)

	writeFile(t, filepath.Join(conflictRoot, "package.yml"), "name: able\nversion: 8.8.8\n")
	writeFile(t, filepath.Join(conflictSrc, "other.able"), `
package other

fn greeting() -> string { "hello from env stdlib" }
`)

	writeFile(t, filepath.Join(appRoot, "package.yml"), `
name: sample
version: 0.0.1
dependencies:
  able: "9.9.9"
targets:
  app: src/main.able
`)
	writeFile(t, filepath.Join(appRoot, "package.lock"), `
root: sample
packages:
  - name: able
    version: 9.9.9
    source: path:../deps/stdlib/src
`)
	writeFile(t, filepath.Join(appRoot, "src", "main.able"), `
package main

import able.locktest.{greeting}

fn main() {
  print(greeting())
}
`)

	enterWorkingDir(t, appRoot)
	t.Setenv("ABLE_MODULE_PATHS", conflictSrc)

	_, _, stderr := runCLIExpectFailure(t, "run")
	assertTextContainsAll(t, stderr,
		"stdlib collision",
		"selected canonical stdlib root (lockfile)",
		stdlibSrc,
		"distinct visible stdlib root (env)",
		conflictSrc,
	)
}

func TestRunRejectsAdhocStdlibCollisionBetweenOverrideAndEnvRoot(t *testing.T) {
	root := t.TempDir()
	appRoot := filepath.Join(root, "app")
	homeDir := filepath.Join(root, "home")
	overrideRoot := filepath.Join(root, "override-stdlib")
	overrideSrc := filepath.Join(overrideRoot, "src")
	envRoot := filepath.Join(root, "env-stdlib")
	envSrc := filepath.Join(envRoot, "src")

	for _, dir := range []string{
		appRoot,
		overrideSrc,
		envSrc,
	} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}

	writeFile(t, filepath.Join(overrideRoot, "package.yml"), "name: able\nversion: 1.0.0\n")
	writeFile(t, filepath.Join(overrideSrc, "override.able"), `
package override

fn greeting() -> string { "override" }
`)

	writeFile(t, filepath.Join(envRoot, "package.yml"), "name: able\nversion: 2.0.0\n")
	writeFile(t, filepath.Join(envSrc, "env.able"), `
package env

fn greeting() -> string { "env" }
`)

	t.Setenv("ABLE_HOME", homeDir)
	t.Setenv("ABLE_MODULE_PATHS", envSrc)

	if err := saveGlobalOverrides(map[string]string{
		normalizeGitURL(defaultStdlibGitURL): overrideRoot,
	}); err != nil {
		t.Fatalf("save stdlib override: %v", err)
	}

	writeFile(t, filepath.Join(appRoot, "main.able"), `
package main

fn main() {
  print("hello")
}
`)

	enterWorkingDir(t, appRoot)

	_, _, stderr := runCLIExpectFailure(t, "main.able")
	assertTextContainsAll(t, stderr,
		"stdlib collision",
		"selected canonical stdlib root (override)",
		overrideSrc,
		"distinct visible stdlib root (env)",
		envSrc,
	)
}

func TestRunRejectsAdhocStdlibCollisionBetweenEnvAndCache(t *testing.T) {
	root := t.TempDir()
	appRoot := filepath.Join(root, "app")
	cacheDir := filepath.Join(root, "cache")
	cacheRoot := filepath.Join(cacheDir, "pkg", "src", "able", defaultStdlibVersion)
	cacheSrc := filepath.Join(cacheRoot, "src")
	envRoot := filepath.Join(root, "env-stdlib")
	envSrc := filepath.Join(envRoot, "src")

	for _, dir := range []string{
		appRoot,
		cacheSrc,
		envSrc,
	} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}

	writeFile(t, filepath.Join(cacheRoot, "package.yml"), "name: able\nversion: "+defaultStdlibVersion+"\n")
	writeFile(t, filepath.Join(cacheSrc, "cached.able"), `
package cached

fn greeting() -> string { "cached" }
`)

	writeFile(t, filepath.Join(envRoot, "package.yml"), "name: able\nversion: 2.0.0\n")
	writeFile(t, filepath.Join(envSrc, "env.able"), `
package env

fn greeting() -> string { "env" }
`)

	writeFile(t, filepath.Join(appRoot, "main.able"), `
package main

fn main() {
  print("hello")
}
`)

	t.Setenv("ABLE_HOME", cacheDir)
	t.Setenv("ABLE_MODULE_PATHS", envSrc)

	enterWorkingDir(t, appRoot)

	_, _, stderr := runCLIExpectFailure(t, "main.able")
	assertTextContainsAll(t, stderr,
		"stdlib collision",
		"selected canonical stdlib root (env)",
		envSrc,
		"distinct visible stdlib root (cache)",
		cacheSrc,
	)
}

func TestRunDynimportRejectsAdhocStdlibCollisionBetweenEnvAndCache(t *testing.T) {
	root := t.TempDir()
	appRoot := filepath.Join(root, "app")
	cacheDir := filepath.Join(root, "cache")
	cacheRoot := filepath.Join(cacheDir, "pkg", "src", "able", defaultStdlibVersion)
	cacheSrc := filepath.Join(cacheRoot, "src")
	envRoot := filepath.Join(root, "env-stdlib")
	envSrc := filepath.Join(envRoot, "src")

	for _, dir := range []string{
		appRoot,
		cacheSrc,
		envSrc,
	} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}

	writeFile(t, filepath.Join(cacheRoot, "package.yml"), "name: able\nversion: "+defaultStdlibVersion+"\n")
	writeFile(t, filepath.Join(cacheSrc, "custom.able"), `
package custom

fn greeting() -> string { "cached" }
`)

	writeFile(t, filepath.Join(envRoot, "package.yml"), "name: able\nversion: 2.0.0\n")
	writeFile(t, filepath.Join(envSrc, "custom.able"), `
package custom

fn greeting() -> string { "env" }
`)

	writeFile(t, filepath.Join(appRoot, "main.able"), `
package main

dynimport able.custom.{greeting}

fn main() {
  print(greeting())
}
`)

	t.Setenv("ABLE_HOME", cacheDir)
	t.Setenv("ABLE_MODULE_PATHS", envSrc)

	enterWorkingDir(t, appRoot)

	_, _, stderr := runCLIExpectFailure(t, "main.able")
	assertTextContainsAll(t, stderr,
		"stdlib collision",
		"selected canonical stdlib root (env)",
		envSrc,
		"distinct visible stdlib root (cache)",
		cacheSrc,
	)
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
	writeFile(t, filepath.Join(stdlibSrc, "package.yml"), `
name: able
version: 0.0.1
`)
	writeFile(t, filepath.Join(stdlibSrc, "core", "thing.able"), `
package thing

fn stdlib_message() -> string {
  "std"
}
`)

	t.Setenv("ABLE_MODULE_PATHS", stdlibSrc)
	enterWorkingDir(t, projectDir)

	writeFile(t, filepath.Join(projectDir, "main.able"), `
import helper.core::helper

fn main() {
  print(helper.value())
}
`)

	_, _, stderr := runCLIExpectFailure(t, "main.able")
	assertTextContainsAny(t, stderr,
		"package helper.core not found",
		"loader: package helper.core not found",
		"imports unknown package helper.core",
	)
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

	enterWorkingDir(t, root)

	entryArg := filepath.Join("foo", "bar", "baz.able")
	_, _, stderr := runCLIExpectFailure(t, entryArg)
	assertTextContainsAll(t, stderr, "package.lock missing")

	lock := driver.NewLockfile("foo_app", cliToolVersion)
	lockPath := filepath.Join(manifestDir, "package.lock")
	if err := driver.WriteLockfile(lock, lockPath); err != nil {
		t.Fatalf("WriteLockfile: %v", err)
	}

	stdout := runCLIExpectSuccess(t, entryArg)
	assertOutputContainsAll(t, stdout, "ran via manifest")
}

func TestCheckCommandSucceeds(t *testing.T) {
	dir := t.TempDir()
	enterWorkingDir(t, dir)

	writeFile(t, filepath.Join(dir, "main.able"), `
fn main() {
}
`)

	stdout := runCLIExpectSuccess(t, "check", "main.able")
	assertOutputContainsAll(t, stdout, "typecheck: ok")
}

func TestCheckCommandReportsDiagnostics(t *testing.T) {
	dir := t.TempDir()
	enterWorkingDir(t, dir)

	writeFile(t, filepath.Join(dir, "broken.able"), `
fn main() {
  value := 1 + "oops"
  value
}
`)

	_, stdout, stderr := runCLIExpectFailure(t, "check", "broken.able")
	if stdout != "" {
		t.Fatalf("expected no stdout on failure, got %q", stdout)
	}
	assertTextContainsAll(t, stderr, "requires numeric operands")
}
