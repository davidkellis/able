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

func TestCompilerGeneratedSerialFlushTracksWorkerHandoff(t *testing.T) {
	result := compileNoFallbackSource(t, "package demo\n\nfn main() -> void {}\n")
	source := string(result.Files["compiled.go"])
	required := []string{
		"workerInFlight int",
		"e.workerInFlight++",
		"len(e.queue) > 0 || e.workerInFlight > 0 ||",
		"e.runSerialTask(task, true)",
		"e.swapCurrent(task.handle, workerTask)",
		"if workerTask && e.workerInFlight > 0",
		"e.workerInFlight--",
	}
	for _, fragment := range required {
		if !strings.Contains(source, fragment) {
			t.Fatalf("generated serial executor is missing handoff accounting %q", fragment)
		}
	}
}

func TestCompilerGeneratedSerialFlushWaitsAcrossWorkerHandoff(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping generated serial flush handoff test in short mode")
	}
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not available")
	}

	moduleRoot, workDir := compilerTestWorkDir(t, "ablec-serial-flush-handoff")
	entryPath := filepath.Join(workDir, "main.able")
	if err := os.WriteFile(entryPath, []byte("package compiler_serial_flush_handoff\n\nfn main() -> void {}\n"), 0o600); err != nil {
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

	result, err := New(Options{PackageName: "main", RequireNoFallbacks: true}).Compile(program)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	if err := result.Write(workDir); err != nil {
		t.Fatalf("write output: %v", err)
	}
	if err := os.WriteFile(filepath.Join(workDir, "serial_flush_handoff_test.go"), []byte(generatedSerialFlushHandoffTest), 0o600); err != nil {
		t.Fatalf("write generated handoff test: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	run := exec.CommandContext(ctx, "go", "test", "-race", "-run", "TestSerialFlush", "-count=20", ".")
	run.Dir = workDir
	run.Env = withEnv(os.Environ(), "GOCACHE", compilerExecGocache(moduleRoot))
	output, err := run.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		t.Fatalf("generated serial flush handoff test timed out\n%s", string(output))
	}
	if err != nil {
		t.Fatalf("generated serial flush handoff test failed: %v\n%s", err, string(output))
	}
}

const generatedSerialFlushHandoffTest = `package main

import (
	"sync"
	"testing"
	"time"

	"able/interpreter-go/pkg/runtime"
)

func newSerialFlushHandoffExecutor() *__able_serial_executor {
	exec := &__able_serial_executor{
		blocked: make(map[*runtime.FutureValue]__able_serial_task),
	}
	exec.cond = sync.NewCond(&exec.mu)
	return exec
}

func dequeueSerialFlushHandoffTask(t *testing.T, exec *__able_serial_executor) (*runtime.FutureValue, __able_serial_task) {
	t.Helper()
	handle := runtime.NewFuture()
	exec.mu.Lock()
	exec.queue = append(exec.queue, __able_serial_task{handle: handle})
	exec.mu.Unlock()
	task, ok := exec.nextTask()
	if !ok {
		t.Fatal("nextTask unexpectedly reported a closed executor")
	}
	return handle, task
}

func assertSerialFlushStillWaiting(t *testing.T, done <-chan struct{}, phase string) {
	t.Helper()
	select {
	case <-done:
		t.Fatalf("Flush returned during %s", phase)
	case <-time.After(10 * time.Millisecond):
	}
}

func TestSerialFlushWaitsAcrossDequeuedToActiveHandoff(t *testing.T) {
	exec := newSerialFlushHandoffExecutor()
	handle, task := dequeueSerialFlushHandoffTask(t, exec)
	if task.handle != handle {
		t.Fatal("nextTask returned the wrong future")
	}

	done := make(chan struct{})
	go func() {
		exec.Flush()
		close(done)
	}()
	assertSerialFlushStillWaiting(t, done, "worker handoff")

	prevCurrent, prevActive, prevPaused := exec.swapCurrent(handle, true)
	assertSerialFlushStillWaiting(t, done, "active execution")
	exec.restoreCurrent(prevCurrent, prevActive, prevPaused)

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Flush did not return after active execution ended")
	}
}

func TestSerialFlushReturnsWhenHandedOffTaskPauses(t *testing.T) {
	exec := newSerialFlushHandoffExecutor()
	handle, _ := dequeueSerialFlushHandoffTask(t, exec)
	prevCurrent, prevActive, prevPaused := exec.swapCurrent(handle, true)

	exec.mu.Lock()
	exec.paused = true
	exec.cond.Broadcast()
	exec.mu.Unlock()

	done := make(chan struct{})
	go func() {
		exec.Flush()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Flush did not return when the handed-off task paused")
	}
	exec.restoreCurrent(prevCurrent, prevActive, prevPaused)
}
`
