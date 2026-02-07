package main

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func writeFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(strings.TrimSpace(contents)+"\n"), 0o644); err != nil {
		t.Fatalf("write file %s: %v", path, err)
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

func TestAblecBuildEmitsGoMod(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not available")
	}

	moduleRoot, err := filepath.Abs(filepath.Join(".", "..", ".."))
	if err != nil {
		t.Fatalf("module root: %v", err)
	}
	buildRoot := t.TempDir()

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

	code, _, stderr := captureCLI(t, []string{"-build", "-o", outDir, "-bin", binPath, entryPath})
	if code != 0 {
		t.Fatalf("ablec build returned exit code %d, stderr: %q", code, stderr)
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

	if !strings.Contains(content, "module able/compiled") {
		t.Fatalf("go.mod missing module header: %q", content)
	}
	if !strings.Contains(content, "require "+modulePath+" v0.0.0") {
		t.Fatalf("go.mod missing require for %s: %q", modulePath, content)
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
}
