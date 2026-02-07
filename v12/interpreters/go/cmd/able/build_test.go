package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestBuildTargetFromManifest(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not available")
	}

	projectDir := t.TempDir()
	writeFile(t, filepath.Join(projectDir, "package.yml"), `
name: demo
targets:
  app: main.able
`)
	writeFile(t, filepath.Join(projectDir, "main.able"), `
extern go fn __able_os_exit(code: i32) -> void {}

fn main() -> void {
  __able_os_exit(0)
}
`)

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
		t.Fatalf("Chdir: %v", err)
	}

	code, _, stderr := captureCLI(t, []string{"build", "app"})
	if code != 0 {
		t.Fatalf("build returned exit code %d, stderr: %q", code, stderr)
	}

	outDir := filepath.Join(projectDir, "target", "compiled", "app")
	binPath := filepath.Join(outDir, "app")
	if runtime.GOOS == "windows" {
		binPath += ".exe"
	}
	if _, err := os.Stat(binPath); err != nil {
		t.Fatalf("expected binary at %s: %v", binPath, err)
	}
}

func TestBuildOutputOutsideModuleRoot(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not available")
	}

	moduleRoot, err := filepath.Abs(filepath.Join(".", "..", ".."))
	if err != nil {
		t.Fatalf("module root: %v", err)
	}
	buildRoot := t.TempDir()
	if isWithinDir(buildRoot, moduleRoot) {
		t.Skip("temp dir is within module root; unable to validate external output")
	}

	outDir := filepath.Join(buildRoot, "out")
	binPath := filepath.Join(buildRoot, "compiled-bin")
	if runtime.GOOS == "windows" {
		binPath += ".exe"
	}

	projectDir := t.TempDir()
	entryPath := filepath.Join(projectDir, "main.able")
	writeFile(t, entryPath, `
extern go fn __able_os_exit(code: i32) -> void {}

fn main() -> void {
  __able_os_exit(0)
}
`)

	code, _, stderr := captureCLI(t, []string{"build", "--out", outDir, "--bin", binPath, entryPath})
	if code != 0 {
		t.Fatalf("build returned exit code %d, stderr: %q", code, stderr)
	}
	if _, err := os.Stat(binPath); err != nil {
		t.Fatalf("expected binary at %s: %v", binPath, err)
	}

	goModPath := filepath.Join(outDir, "go.mod")
	data, err := os.ReadFile(goModPath)
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	content := string(data)
	modulePath, err := readGoModulePath(moduleRoot)
	if err != nil {
		t.Fatalf("read module path: %v", err)
	}
	if !strings.Contains(content, "replace "+modulePath+" => ./v12/interpreters/go") {
		t.Fatalf("go.mod missing replace for %s: %q", modulePath, content)
	}

	moduleCopy := filepath.Join(outDir, "v12", "interpreters", "go", "go.mod")
	if _, err := os.Stat(moduleCopy); err != nil {
		t.Fatalf("expected interpreter module copy at %s: %v", moduleCopy, err)
	}
	parserCopy := filepath.Join(outDir, "v12", "parser", "tree-sitter-able", "src", "parser.c")
	if _, err := os.Stat(parserCopy); err != nil {
		t.Fatalf("expected parser sources at %s: %v", parserCopy, err)
	}

	relocated := filepath.Join(buildRoot, "relocated")
	if err := os.Rename(outDir, relocated); err != nil {
		t.Fatalf("relocate output: %v", err)
	}
	build := exec.Command("go", "build", ".")
	build.Dir = relocated
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build failed in relocated output: %v\n%s", err, string(output))
	}
}
