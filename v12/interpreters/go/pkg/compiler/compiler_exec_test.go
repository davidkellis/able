package compiler

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"able/interpreter-go/pkg/driver"
)

func TestCompilerExecHarness(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping compiler exec harness in short mode")
	}
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not available")
	}
	moduleRoot, workDir := compilerTestWorkDir(t, "ablec")

	source := `extern go fn __able_os_exit(code: i32) -> void {}

fn add(x: i32, y: i32) -> i32 {
  x + y
}

fn main() {
  __able_os_exit(add(1, 2))
}
`
	entryPath := filepath.Join(workDir, "app.able")
	if err := os.WriteFile(entryPath, []byte(source), 0o600); err != nil {
		t.Fatalf("write source: %v", err)
	}

	loader, err := driver.NewLoader(nil)
	if err != nil {
		t.Fatalf("loader init: %v", err)
	}
	t.Cleanup(func() { loader.Close() })

	program, err := loader.Load(entryPath)
	if err != nil {
		t.Fatalf("load program: %v", err)
	}

	outputDir := filepath.Join(workDir, "out")
	comp := New(Options{
		PackageName: "main",
		EmitMain:    true,
		EntryPath:   entryPath,
	})
	result, err := comp.Compile(program)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	if err := result.Write(outputDir); err != nil {
		t.Fatalf("write output: %v", err)
	}

	binPath := filepath.Join(workDir, "compiled-bin")
	build := exec.Command("go", "build", "-o", binPath, ".")
	build.Dir = outputDir
	build.Env = withEnv(os.Environ(), "GOCACHE", compilerExecGocache(moduleRoot))
	if runtime.GOOS == "windows" {
		binPath += ".exe"
	}
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\n%s", err, string(output))
	}

	run := exec.Command(binPath)
	output, err := run.CombinedOutput()
	if err == nil {
		t.Fatalf("expected exit code 3, got 0")
	}
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected exit error, got %v\n%s", err, string(output))
	}
	if exitErr.ExitCode() != 3 {
		t.Fatalf("expected exit code 3, got %d\n%s", exitErr.ExitCode(), string(output))
	}
}
