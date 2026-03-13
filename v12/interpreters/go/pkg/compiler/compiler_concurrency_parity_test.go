package compiler

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	goruntime "runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"able/interpreter-go/pkg/driver"
	"able/interpreter-go/pkg/interpreter"
)

const compilerConcurrencyParityFixtureEnv = "ABLE_COMPILER_CONCURRENCY_PARITY_FIXTURES"

func TestCompilerConcurrencyParityFixtures(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping compiler concurrency parity fixtures in short mode")
	}
	root := filepath.Join(repositoryRoot(), "v12", "fixtures", "exec")
	if _, err := os.Stat(root); os.IsNotExist(err) {
		root = filepath.Join("..", "..", "fixtures", "exec")
	}
	fixtures := resolveCompilerConcurrencyParityFixtures()
	if len(fixtures) == 0 {
		t.Skip("no concurrency fixtures configured")
	}
	for _, rel := range fixtures {
		rel := rel
		t.Run(rel, func(t *testing.T) {
			dir := filepath.Join(root, filepath.FromSlash(rel))
			manifest, err := interpreter.LoadFixtureManifest(dir)
			if err != nil {
				t.Fatalf("read manifest: %v", err)
			}
			if shouldSkipTarget(manifest.SkipTargets, "go") {
				return
			}
			if manifest.Expect.TypecheckDiagnostics != nil && len(manifest.Expect.TypecheckDiagnostics) > 0 {
				return
			}

			tree := runTreewalkerFixtureOutcome(t, dir, manifest)
			compiled := runCompiledFixtureOutcome(t, dir, manifest)

			if tree.Exit != compiled.Exit {
				t.Fatalf("exit mismatch: treewalker=%d compiled=%d", tree.Exit, compiled.Exit)
			}
			if !reflect.DeepEqual(tree.Stdout, compiled.Stdout) {
				t.Fatalf("stdout mismatch: treewalker=%v compiled=%v", tree.Stdout, compiled.Stdout)
			}
			if !reflect.DeepEqual(tree.Stderr, compiled.Stderr) {
				t.Fatalf("stderr mismatch: treewalker=%v compiled=%v", tree.Stderr, compiled.Stderr)
			}
		})
	}
}

func resolveCompilerConcurrencyParityFixtures() []string {
	if raw := strings.TrimSpace(os.Getenv(compilerConcurrencyParityFixtureEnv)); raw != "" {
		parts := strings.FieldsFunc(raw, func(r rune) bool {
			return r == ',' || r == ';' || r == '\n' || r == '\t' || r == ' '
		})
		fixtures := make([]string, 0, len(parts))
		seen := map[string]struct{}{}
		for _, part := range parts {
			name := strings.TrimSpace(part)
			if name == "" {
				continue
			}
			if _, ok := seen[name]; ok {
				continue
			}
			seen[name] = struct{}{}
			fixtures = append(fixtures, name)
		}
		return fixtures
	}
	return []string{
		"06_01_compiler_spawn_await",
		"06_01_compiler_await_future",
		"12_01_bytecode_spawn_basic",
		"12_01_bytecode_await_default",
		"12_02_async_spawn_combo",
		"12_02_future_fairness_cancellation",
		"12_03_spawn_future_status_error",
		"12_04_future_handle_value_view",
		"12_05_concurrency_channel_ping_pong",
		"12_05_mutex_lock_unlock",
		"12_06_await_fairness_cancellation",
		"12_07_channel_mutex_error_types",
		"12_08_blocking_io_concurrency",
		"15_04_background_work_flush",
	}
}

func TestCompilerFutureFlushReturnsWithBlockedGoroutineTasks(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping compiler goroutine flush parity test in short mode")
	}
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not available")
	}

	moduleRoot, workDir := compilerTestWorkDir(t, "ablec-goroutine-flush")

	source := `package compiler_future_flush_goroutine

fn main() -> void {
  handle := __able_channel_new(0)
  spawn {
    __able_channel_receive(handle)
  }
  future_flush()
  print(` + "`pending ${future_pending_tasks()}`" + `)
}
`
	entryPath := filepath.Join(workDir, "main.able")
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

	comp := New(Options{
		PackageName:        "main",
		RequireNoFallbacks: requireNoFallbacksForFixtureGates(t),
	})
	result, err := comp.Compile(program)
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

	binPath := filepath.Join(workDir, "compiled-goroutine-flush")
	if goruntime.GOOS == "windows" {
		binPath += ".exe"
	}
	build := exec.Command("go", "build", "-o", binPath, ".")
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
		t.Fatalf("compiled run timed out; future_flush likely blocked with blocked tasks\n%s", string(output))
	}
	if err != nil {
		t.Fatalf("compiled run failed: %v\n%s", err, string(output))
	}

	lines := splitLines(string(output))
	if len(lines) != 1 {
		t.Fatalf("expected one stdout line, got %v", lines)
	}
	line := lines[0]
	if !strings.HasPrefix(line, "pending ") {
		t.Fatalf("expected pending line, got %q", line)
	}
	pendingText := strings.TrimSpace(strings.TrimPrefix(line, "pending "))
	pending, err := strconv.Atoi(pendingText)
	if err != nil {
		t.Fatalf("parse pending count %q: %v", pendingText, err)
	}
	if pending < 1 {
		t.Fatalf("expected blocked task to remain pending after flush, got %d", pending)
	}

	tree := runTreewalkerFixtureOutcome(t, workDir, interpreter.FixtureManifest{
		Entry:    "main.able",
		Executor: "goroutine",
	})
	if tree.Exit != 0 {
		t.Fatalf("treewalker run failed: stderr=%v exit=%d", tree.Stderr, tree.Exit)
	}
	if len(tree.Stdout) != 1 {
		t.Fatalf("expected one treewalker stdout line, got %v", tree.Stdout)
	}
	treeLine := tree.Stdout[0]
	if !strings.HasPrefix(treeLine, "pending ") {
		t.Fatalf("expected treewalker pending line, got %q", treeLine)
	}
	treePendingText := strings.TrimSpace(strings.TrimPrefix(treeLine, "pending "))
	treePending, err := strconv.Atoi(treePendingText)
	if err != nil {
		t.Fatalf("parse treewalker pending count %q: %v", treePendingText, err)
	}
	if treePending < 1 {
		t.Fatalf("expected blocked task to remain pending after treewalker flush, got %d", treePending)
	}
}
