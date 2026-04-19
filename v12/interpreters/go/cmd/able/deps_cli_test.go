package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"able/interpreter-go/pkg/driver"
)

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

import util.core::core

fn main() -> void {
  print(core.value())
}
`)

	cacheDir := filepath.Join(root, "cache")
	t.Setenv("ABLE_HOME", cacheDir)
	t.Setenv("ABLE_REGISTRY", registry)
	t.Setenv("ABLE_MODULE_PATHS", repoStdlibPath(t)+string(os.PathListSeparator)+repoKernelPath(t))

	enterWorkingDir(t, project)

	_ = runCLIExpectSuccess(t, "deps", "install")

	lockPath := filepath.Join(project, "package.lock")
	lock, err := driver.LoadLockfile(lockPath)
	if err != nil {
		t.Fatalf("LoadLockfile: %v", err)
	}
	if len(lock.Packages) != 3 {
		t.Fatalf("lock packages unexpected: %#v", lock.Packages)
	}
	pkg := requireLockedPackage(t, lock.Packages, "util")
	if pkg.Version != "1.0.0" {
		t.Fatalf("lock entry unexpected: %#v", pkg)
	}
	expectedSource := fmt.Sprintf("registry:default/%s/%s", pkg.Name, pkg.Version)
	if pkg.Source != expectedSource {
		t.Fatalf("pkg.Source = %q, want %q", pkg.Source, expectedSource)
	}
	requireLockedStdlibAndKernel(t, lock.Packages)

	stdout := runCLIExpectSuccess(t, "run")
	assertOutputContainsAll(t, stdout, "util dependency")

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

import util.core::core

fn main() -> void {
  print(core.value())
}
`)

	cacheDir := filepath.Join(root, "cache")
	t.Setenv("ABLE_HOME", cacheDir)
	t.Setenv("ABLE_REGISTRY", registry)
	t.Setenv("ABLE_MODULE_PATHS", repoStdlibPath(t)+string(os.PathListSeparator)+repoKernelPath(t))

	enterWorkingDir(t, project)

	_ = runCLIExpectSuccess(t, "deps", "install")

	writeFile(t, filepath.Join(project, "src", "main.able"), `
package main

import util.core::core

fn main() -> void {
  print(core.value() + 1)
}
`)

	_, stdout, stderr := runCLIExpectFailure(t, "run")
	if stdout != "" {
		t.Fatalf("expected no stdout when typecheck fails, got %q", stdout)
	}
	assertTextContainsAll(t, stderr, "typechecker:", "package export summary")
}
