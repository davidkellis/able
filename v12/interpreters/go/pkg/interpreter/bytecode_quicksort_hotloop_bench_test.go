package interpreter

import (
	"path/filepath"
	"runtime"
	"testing"

	"able/interpreter-go/pkg/driver"
)

const quicksortHotloopBenchFixture = "v12/fixtures/exec/07_10_bytecode_quicksort_hotloop"

func suspendMemProfileSampling() func() {
	prev := runtime.MemProfileRate
	runtime.MemProfileRate = 0
	resumed := false
	return func() {
		if resumed {
			return
		}
		runtime.MemProfileRate = prev
		resumed = true
	}
}

func BenchmarkBytecodeQuicksortHotloopRuntime(b *testing.B) {
	resumeMemProfile := suspendMemProfileSampling()
	defer resumeMemProfile()

	root := repositoryRoot()
	if root == "" {
		b.Fatalf("repository root not found")
	}
	dir := filepath.Join(root, filepath.FromSlash(quicksortHotloopBenchFixture))
	manifest := readManifest(b, dir)
	entry := manifest.Entry
	if entry == "" {
		entry = "main.able"
	}
	entryPath := filepath.Join(dir, entry)

	searchPaths, err := buildExecSearchPaths(entryPath, dir, manifest)
	if err != nil {
		b.Fatalf("exec search paths: %v", err)
	}
	loader, err := driver.NewLoader(searchPaths)
	if err != nil {
		b.Fatalf("loader init: %v", err)
	}
	defer loader.Close()

	program, err := loader.Load(entryPath)
	if err != nil {
		b.Fatalf("load program: %v", err)
	}

	executor := NewSerialExecutor(nil)
	defer executor.Close()

	interp := NewBytecodeWithExecutor(executor)
	registerBenchPrint(interp)

	skipTypecheck := benchSkipTypecheck()
	_, entryEnv, _, err := interp.EvaluateProgram(program, ProgramEvaluationOptions{
		SkipTypecheck:    skipTypecheck,
		AllowDiagnostics: !skipTypecheck,
	})
	if err != nil {
		b.Fatalf("evaluate program: %v", err)
	}

	env := entryEnv
	if env == nil {
		env = interp.GlobalEnvironment()
	}
	mainValue, err := env.Get("main")
	if err != nil {
		b.Fatalf("lookup main: %v", err)
	}

	// Keep memprofiles focused on the steady-state runtime loop by running a
	// warmup call before sampling resumes.
	runtime.GC()
	if _, err := interp.CallFunction(mainValue, nil); err != nil {
		b.Fatalf("warmup call main: %v", err)
	}
	runtime.GC()
	resumeMemProfile()
	runtime.GC()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := interp.CallFunction(mainValue, nil); err != nil {
			b.Fatalf("call main: %v", err)
		}
	}
	executor.Flush()
}
