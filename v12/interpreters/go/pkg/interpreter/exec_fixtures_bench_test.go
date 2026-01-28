package interpreter

import (
	"os"
	"path/filepath"
	"testing"

	"able/interpreter-go/pkg/driver"
	"able/interpreter-go/pkg/runtime"
)

const defaultBenchFixture = "v12/fixtures/exec/07_09_bytecode_iterator_yield"

func BenchmarkExecFixtureTreewalker(b *testing.B) {
	benchmarkExecFixture(b, testExecTreewalker)
}

func BenchmarkExecFixtureBytecode(b *testing.B) {
	benchmarkExecFixture(b, testExecBytecode)
}

func benchmarkExecFixture(b *testing.B, mode testExecMode) {
	b.Helper()
	dir := resolveBenchFixtureDir(b)
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

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		interp := newBenchmarkInterpreter(mode, executor)
		registerBenchPrint(interp)

		_, entryEnv, _, err := interp.EvaluateProgram(program, ProgramEvaluationOptions{
			SkipTypecheck: true,
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
		if _, err := interp.CallFunction(mainValue, nil); err != nil {
			b.Fatalf("call main: %v", err)
		}
		executor.Flush()
	}
}

func newBenchmarkInterpreter(mode testExecMode, executor Executor) *Interpreter {
	switch mode {
	case testExecBytecode:
		return NewBytecodeWithExecutor(executor)
	default:
		return NewWithExecutor(executor)
	}
}

func resolveBenchFixtureDir(b *testing.B) string {
	b.Helper()
	root := repositoryRoot()
	if root == "" {
		b.Fatalf("repository root not found")
	}
	if raw := os.Getenv("ABLE_BENCH_FIXTURE"); raw != "" {
		if filepath.IsAbs(raw) {
			return raw
		}
		return filepath.Join(root, filepath.FromSlash(raw))
	}
	return filepath.Join(root, filepath.FromSlash(defaultBenchFixture))
}

func registerBenchPrint(interp *Interpreter) {
	printFn := runtime.NativeFunctionValue{
		Name:  "print",
		Arity: 1,
		Impl: func(_ *runtime.NativeCallContext, _ []runtime.Value) (runtime.Value, error) {
			return runtime.NilValue{}, nil
		},
	}
	interp.GlobalEnvironment().Define("print", printFn)
}
