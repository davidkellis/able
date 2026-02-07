package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func TestBuildTargetFromManifest(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not available")
	}

	moduleRoot, err := filepath.Abs(filepath.Join(".", "..", ".."))
	if err != nil {
		t.Fatalf("module root: %v", err)
	}
	tmpRoot := filepath.Join(moduleRoot, "tmp")
	if err := os.MkdirAll(tmpRoot, 0o755); err != nil {
		t.Fatalf("mkdir tmp: %v", err)
	}
	buildRoot, err := os.MkdirTemp(tmpRoot, "able-build-")
	if err != nil {
		t.Fatalf("build root: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(buildRoot) })

	outDir := filepath.Join(buildRoot, "out")
	binPath := filepath.Join(buildRoot, "compiled-bin")
	if runtime.GOOS == "windows" {
		binPath += ".exe"
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

	code, _, stderr := captureCLI(t, []string{"build", "--out", outDir, "--bin", binPath, "app"})
	if code != 0 {
		t.Fatalf("build returned exit code %d, stderr: %q", code, stderr)
	}
	if _, err := os.Stat(binPath); err != nil {
		t.Fatalf("expected binary at %s: %v", binPath, err)
	}
}
