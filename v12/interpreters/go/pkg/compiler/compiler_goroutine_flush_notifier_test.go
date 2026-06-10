package compiler

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"able/interpreter-go/pkg/driver"
)

func TestCompilerGeneratedGoroutineFlushNotifiesAllWaiters(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping generated goroutine flush notifier test in short mode")
	}
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not available")
	}

	moduleRoot, workDir := compilerTestWorkDir(t, "ablec-goroutine-flush-notifier")
	entryPath := filepath.Join(workDir, "main.able")
	if err := os.WriteFile(entryPath, []byte("package compiler_goroutine_flush_notifier\n\nfn main() -> void {}\n"), 0o600); err != nil {
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

	result, err := New(Options{PackageName: "main"}).Compile(program)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	if err := result.Write(workDir); err != nil {
		t.Fatalf("write output: %v", err)
	}
	compiledSource, err := os.ReadFile(filepath.Join(workDir, "compiled.go"))
	if err != nil {
		t.Fatalf("read generated runtime: %v", err)
	}
	flushStart := strings.Index(string(compiledSource), "func (e *__able_goroutine_executor) Flush()")
	flushEnd := strings.Index(string(compiledSource), "func (e *__able_goroutine_executor) notifyFlushProgress()")
	if flushStart < 0 || flushEnd <= flushStart {
		t.Fatalf("generated goroutine flush/progress notifier functions are missing")
	}
	flushSource := string(compiledSource[flushStart:flushEnd])
	if !strings.Contains(flushSource, "e.progressCond.Wait()") {
		t.Fatalf("generated goroutine flush does not wait for progress notification")
	}
	if strings.Contains(flushSource, "time.Sleep(0)") {
		t.Fatalf("generated goroutine flush still busy-yields")
	}

	if err := os.WriteFile(filepath.Join(workDir, "goroutine_flush_notifier_test.go"), []byte(generatedGoroutineFlushNotifierTest), 0o600); err != nil {
		t.Fatalf("write generated notifier test: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	run := exec.CommandContext(ctx, "go", "test", "-race", "-run", "TestFlush", "-count=1", ".")
	run.Dir = workDir
	run.Env = withEnv(os.Environ(), "GOCACHE", compilerExecGocache(moduleRoot))
	output, err := run.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		t.Fatalf("generated goroutine flush notifier test timed out\n%s", string(output))
	}
	if err != nil {
		t.Fatalf("generated goroutine flush notifier test failed: %v\n%s", err, string(output))
	}
}

const generatedGoroutineFlushNotifierTest = `package main

import (
	"context"
	"testing"
	"time"

	"able/interpreter-go/pkg/runtime"
)

func TestFlushWakesConcurrentWaitersOnCompletion(t *testing.T) {
	exec := __able_new_goroutine_executor()
	started := make(chan struct{})
	release := make(chan struct{})
	handle := exec.RunFuture(nil, func(_ *runtime.Environment) (runtime.Value, error) {
		close(started)
		<-release
		return runtime.NilValue{}, nil
	})
	if handle == nil {
		t.Fatal("RunFuture returned nil handle")
	}
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("future did not start")
	}

	done := make(chan struct{}, 2)
	for range 2 {
		go func() {
			exec.Flush()
			done <- struct{}{}
		}()
	}
	select {
	case <-done:
		t.Fatal("Flush returned before the running future completed")
	case <-time.After(20 * time.Millisecond):
	}

	close(release)
	for range 2 {
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("Flush did not wake after future completion")
		}
	}
	if got := exec.PendingTasks(); got != 0 {
		t.Fatalf("pending tasks after flush = %d, want 0", got)
	}
}

func TestFlushReturnsWhenAllRemainingTasksAreBlocked(t *testing.T) {
	exec := __able_new_goroutine_executor()
	started := make(chan struct{})
	release := make(chan struct{})
	handle := exec.RunFuture(nil, func(_ *runtime.Environment) (runtime.Value, error) {
		close(started)
		<-release
		return runtime.NilValue{}, nil
	})
	if handle == nil {
		t.Fatal("RunFuture returned nil handle")
	}
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("future did not start")
	}
	exec.MarkBlocked(handle)
	done := make(chan struct{})
	go func() {
		exec.Flush()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Flush did not return for an all-blocked executor")
	}
	close(release)
	deadline := time.After(time.Second)
	for exec.PendingTasks() != 0 {
		select {
		case <-deadline:
			t.Fatal("blocked future did not complete after release")
		default:
			time.Sleep(time.Millisecond)
		}
	}
}

func TestFlushWaitsForCancellingBlockedTaskToUnwind(t *testing.T) {
	exec := __able_new_goroutine_executor()
	ready := make(chan struct{})
	started := make(chan struct{})
	var handle *runtime.FutureValue
	handle = exec.RunFuture(nil, func(_ *runtime.Environment) (runtime.Value, error) {
		close(started)
		<-ready
		<-handle.Context().Done()
		exec.MarkUnblocked(handle)
		return nil, context.Canceled
	})
	if handle == nil {
		t.Fatal("RunFuture returned nil handle")
	}
	close(ready)
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("future did not start")
	}
	exec.MarkBlocked(handle)
	handle.RequestCancel()
	exec.Flush()
	if got := handle.Status(); got != runtime.FutureCancelled {
		t.Fatalf("cancelled blocked future status = %v, want cancelled", got)
	}
}
`
