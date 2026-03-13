package compiler

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"able/interpreter-go/pkg/driver"
)

func TestCompilerZeroFieldStructIdentifierValue(t *testing.T) {
	source := `extern go fn __able_os_exit(code: i32) -> void {}

struct Marker {}

fn marker() -> Marker { Marker }

fn main() {
  _value := marker()
  __able_os_exit(0)
}
`
	compileAndRunSource(t, "ablec-singleton-", source)
}

func TestCompilerSingletonStaticOverloadDispatch(t *testing.T) {
	source := `extern go fn __able_os_exit(code: i32) -> void {}

struct Factory {}

methods Factory {
  fn make() -> i32 {
    2
  }

  fn make(_first: i32, _second: i32) -> i32 {
    Factory.make()
  }
}

fn main() {
  result := Factory.make(3, 7)
  if result == 2 {
    __able_os_exit(0)
  }
  __able_os_exit(1)
}
`
	compileAndRunSource(t, "ablec-singleton-overload-", source)
}

func compileAndRunSource(t *testing.T, tempPrefix string, source string) {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping singleton compiler integration test in short mode")
	}
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not available")
	}

	moduleRoot, workDir := compilerTestWorkDir(t, tempPrefix)

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
	if output, err := run.CombinedOutput(); err != nil {
		t.Fatalf("run failed: %v\n%s", err, string(output))
	}
}
