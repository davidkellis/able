package compiler

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"able/interpreter-go/pkg/driver"
)

func TestCompilerExperimentalExecutionContextNestedSpawnExecutes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping experimental execution-context build in short mode")
	}
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not available")
	}

	moduleRoot, workDir := compilerTestWorkDir(t, "ablec-execution-context")
	entryPath := filepath.Join(workDir, "main.able")
	source := `package compiler_execution_context

fn main() -> void {
  channel := __able_channel_new(1)
  outer := spawn {
    inner := spawn {
      __able_channel_send(channel, 42)
    }
    inner.value()
    __able_channel_receive(channel)
  }
  future_flush()
  print(outer.value())
}
`
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

	result, err := New(Options{
		PackageName:                  "main",
		RequireNoFallbacks:           requireNoFallbacksForFixtureGates(t),
		ExperimentalExecutionContext: true,
	}).Compile(program)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	if err := result.Write(workDir); err != nil {
		t.Fatalf("write output: %v", err)
	}
	harness := compilerHarnessSource(entryPath, nil, "goroutine")
	if err := os.WriteFile(filepath.Join(workDir, "main.go"), []byte(harness), 0o600); err != nil {
		t.Fatalf("write harness: %v", err)
	}

	binPath := filepath.Join(workDir, "compiled-execution-context")
	build := exec.Command("go", "build", "-race", "-o", binPath, ".")
	build.Dir = workDir
	build.Env = withEnv(os.Environ(), "GOCACHE", compilerExecGocache(moduleRoot))
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\n%s", err, string(output))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	run := exec.CommandContext(ctx, binPath)
	output, err := run.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		t.Fatalf("experimental execution-context binary timed out\n%s", string(output))
	}
	if err != nil {
		t.Fatalf("experimental execution-context binary failed: %v\n%s", err, string(output))
	}
	if got, want := splitLines(string(output)), []string{"42"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("stdout = %v, want %v", got, want)
	}
}
