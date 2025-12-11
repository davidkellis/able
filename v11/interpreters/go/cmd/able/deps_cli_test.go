package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"able/interpreter10-go/pkg/driver"
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
	if len(lock.Packages) != 3 {
		t.Fatalf("lock packages unexpected: %#v", lock.Packages)
	}
	pkg := findLockedPackage(lock.Packages, "util")
	if pkg == nil || pkg.Version != "1.0.0" {
		t.Fatalf("lock entry unexpected: %#v", pkg)
	}
	expectedSource := fmt.Sprintf("registry:default/%s/%s", pkg.Name, pkg.Version)
	if pkg.Source != expectedSource {
		t.Fatalf("pkg.Source = %q, want %q", pkg.Source, expectedSource)
	}
	if stdlib := findLockedPackage(lock.Packages, "able"); stdlib == nil {
		t.Fatalf("expected stdlib entry in lock: %#v", lock.Packages)
	}
	if kernel := findLockedPackage(lock.Packages, "kernel"); kernel == nil {
		t.Fatalf("expected kernel entry in lock: %#v", lock.Packages)
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

import util.core::core

fn main() -> void {
  print(core.value())
}
`)

	cacheDir := filepath.Join(root, "cache")
	t.Setenv("ABLE_HOME", cacheDir)
	t.Setenv("ABLE_REGISTRY", registry)
	t.Setenv("ABLE_MODULE_PATHS", repoStdlibPath(t)+string(os.PathListSeparator)+repoKernelPath(t))

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

import util.core::core

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
