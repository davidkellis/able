package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"able/interpreter-go/pkg/driver"
)

func TestSetupInstallsStdlibAndKernelAndRunSupportsBothExecModes(t *testing.T) {
	root := t.TempDir()
	homeDir := filepath.Join(root, "home")
	projectDir := filepath.Join(root, "project")
	stdlibRoot := filepath.Join(root, "stdlib")
	stdlibSrc := filepath.Join(stdlibRoot, "src")

	if err := os.MkdirAll(filepath.Join(stdlibSrc, "core"), 0o755); err != nil {
		t.Fatalf("mkdir stdlib core: %v", err)
	}
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("mkdir project: %v", err)
	}

	writeFile(t, filepath.Join(stdlibRoot, "package.yml"), `
name: able
version: `+defaultStdlibVersion+`
`)
	writeFile(t, filepath.Join(stdlibSrc, "package.yml"), `
name: able
version: `+defaultStdlibVersion+`
`)
	writeFile(t, filepath.Join(stdlibSrc, "core", "thing.able"), `
package thing

fn stdlib_message() -> string {
  "stdlib-smoke"
}
`)
	writeFile(t, filepath.Join(projectDir, "main.able"), `
package main

import able.core.thing::thing

fn main() {
  print(thing.stdlib_message())
}
`)

	t.Setenv("ABLE_HOME", homeDir)
	t.Setenv("ABLE_PATH", "")
	t.Setenv("ABLE_MODULE_PATHS", "")

	if err := saveGlobalOverrides(map[string]string{
		normalizeGitURL(defaultStdlibGitURL): stdlibRoot,
	}); err != nil {
		t.Fatalf("save stdlib override: %v", err)
	}

	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	defer func() {
		if chdirErr := os.Chdir(oldWD); chdirErr != nil {
			t.Fatalf("restore working directory: %v", chdirErr)
		}
	}()
	if err := os.Chdir(projectDir); err != nil {
		t.Fatalf("Chdir project: %v", err)
	}

	code, stdout, stderr := captureCLI(t, []string{"setup"})
	if code != 0 {
		t.Fatalf("able setup returned exit code %d, stderr: %q", code, stderr)
	}
	if !strings.Contains(stdout, "setup complete") {
		t.Fatalf("expected setup completion message, got stdout %q", stdout)
	}

	lockPath := filepath.Join(homeDir, "setup.lock")
	lock, err := driver.LoadLockfile(lockPath)
	if err != nil {
		t.Fatalf("load setup lockfile %s: %v", lockPath, err)
	}
	stdlibPkg := findLockedPackage(lock.Packages, "able")
	if stdlibPkg == nil {
		t.Fatalf("setup lock missing stdlib package: %#v", lock.Packages)
	}
	kernelPkg := findLockedPackage(lock.Packages, "kernel")
	if kernelPkg == nil {
		t.Fatalf("setup lock missing kernel package: %#v", lock.Packages)
	}

	stdlibPath := strings.TrimPrefix(stdlibPkg.Source, "path:")
	if stdlibPath == stdlibPkg.Source || stdlibPath == "" {
		t.Fatalf("stdlib source is not a path source: %q", stdlibPkg.Source)
	}
	if info, err := os.Stat(resolvePackageSrcPath(stdlibPath)); err != nil || !info.IsDir() {
		t.Fatalf("stdlib source path invalid: %s (%v)", resolvePackageSrcPath(stdlibPath), err)
	}

	kernelPath := strings.TrimPrefix(kernelPkg.Source, "path:")
	if kernelPath == kernelPkg.Source || kernelPath == "" {
		t.Fatalf("kernel source is not a path source: %q", kernelPkg.Source)
	}
	if _, err := os.Stat(filepath.Join(kernelPath, "kernel.able")); err != nil {
		t.Fatalf("kernel source missing kernel.able at %s: %v", kernelPath, err)
	}

	code, stdout, stderr = captureCLI(t, []string{"run", "main.able"})
	if code != 0 {
		t.Fatalf("treewalker run returned exit code %d, stderr: %q", code, stderr)
	}
	if !strings.Contains(stdout, "stdlib-smoke") {
		t.Fatalf("treewalker run missing stdlib output; stdout=%q stderr=%q", stdout, stderr)
	}

	code, stdout, stderr = captureCLI(t, []string{"--exec-mode", "bytecode", "run", "main.able"})
	if code != 0 {
		t.Fatalf("bytecode run returned exit code %d, stderr: %q", code, stderr)
	}
	if !strings.Contains(stdout, "stdlib-smoke") {
		t.Fatalf("bytecode run missing stdlib output; stdout=%q stderr=%q", stdout, stderr)
	}
}
