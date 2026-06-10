package compiler

import (
	"strings"
	"testing"
)

// Repeated public await_lock contention exercises the generated Awaitable
// registration lifecycle. In particular, a wake must permit another register
// cycle when a competing task acquires the mutex before this task commits.
func TestCompilerPublicMutexAwaitLockRearmsAfterContention(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping compiled public mutex await-lock execution test in short mode")
	}
	t.Setenv("ABLE_EXECUTOR", "goroutine")

	source := strings.Join([]string{
		"package demo",
		"",
		"import able.kernel.{Mutex}",
		"",
		"fn worker(mutex: Mutex, rounds: i64) -> i64 {",
		"  subtotal := 0_i64",
		"  index := 0_i64",
		"  loop {",
		"    if index >= rounds { break }",
		"    committed := await [mutex.await_lock(fn() -> i64 {",
		"      do { index + 1_i64 } ensure { mutex.unlock() }",
		"    })]",
		"    subtotal = subtotal + committed",
		"    index = index + 1_i64",
		"  }",
		"  subtotal",
		"}",
		"",
		"fn main() -> void {",
		"  mutex := Mutex.new()",
		"  first := spawn { worker(mutex, 48_i64) }",
		"  second := spawn { worker(mutex, 48_i64) }",
		"  print((first.value()! as i64) + (second.value()! as i64))",
		"}",
		"",
	}, "\n")

	for _, tc := range []struct {
		name         string
		experimental bool
	}{
		{name: "default"},
		{name: "execution_context", experimental: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := compileAndRunExecSourceWithOptions(t, "ablec-public-mutex-await-lock", source, Options{
				PackageName:                  "main",
				EmitMain:                     true,
				ExperimentalExecutionContext: tc.experimental,
			})
			if got != "2352\n" {
				t.Fatalf("compiled public mutex await_lock output = %q, want %q", got, "2352\\n")
			}
		})
	}
}
