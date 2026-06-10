package interpreter

import (
	"fmt"
	"os"
	"path/filepath"
	goRuntime "runtime"
	"strings"
	"testing"

	"able/interpreter-go/pkg/driver"
)

const linkedListIteratorFilterMapBenchFixture = "v12/fixtures/bench/linked_list_iterator_filter_map_i64_small"
const linkedListIteratorCollectBenchFixture = "v12/fixtures/bench/linked_list_iterator_collect_i64_small"
const heapI32BenchFixture = "v12/fixtures/bench/heap_i32_small"
const persistentSortedSetBenchFixture = "v12/fixtures/bench/persistent_sorted_set_i32_small"

func TestBytecodeLinkedListIteratorFilterMapBenchWarmup(t *testing.T) {
	runBytecodeBenchWarmup(t, linkedListIteratorFilterMapBenchFixture)
}

func TestBytecodeLinkedListIteratorCollectBenchWarmup(t *testing.T) {
	runBytecodeBenchWarmup(t, linkedListIteratorCollectBenchFixture)
}

func TestBytecodeHeapI32BenchWarmup(t *testing.T) {
	runBytecodeBenchWarmup(t, heapI32BenchFixture)
}

func TestBytecodePersistentSortedSetBenchWarmup(t *testing.T) {
	runBytecodeBenchWarmup(t, persistentSortedSetBenchFixture)
}

func TestBytecodePersistentSortedSetSlotLoweringCoverage(t *testing.T) {
	root := repositoryRoot()
	if root == "" {
		t.Fatalf("repository root not found")
	}
	stdlibPath := filepath.Join(root, "..", "able-stdlib", "src", "collections", "persistent_sorted_set.able")
	if _, err := os.Stat(stdlibPath); err != nil {
		if os.IsNotExist(err) {
			t.Skipf("canonical able-stdlib checkout not found at %s", stdlibPath)
		}
		t.Fatalf("stat persistent sorted set stdlib: %v", err)
	}

	row, err := runBytecodeOpcodeAuditTarget(bytecodeOpcodeAuditTarget{
		Name: "persistent_sorted_set",
		Path: stdlibPath,
	})
	if err != nil {
		t.Fatalf("audit persistent sorted set: %v", err)
	}
	for _, name := range []string{"insert_tree", "remove_tree"} {
		fn, ok := bytecodeOpcodeAuditFunctionByName(row, name)
		if !ok {
			t.Fatalf("audit missing function %s", name)
		}
		if !fn.SlotFrame {
			t.Fatalf("%s should use slot frame", name)
		}
		if got := fn.TrackedOpcodeHits["Match"]; got != 0 {
			t.Fatalf("%s generic Match opcode count = %d, want 0", name, got)
		}
		if got := fn.TrackedOpcodeHits["JumpIfNotTypedPattern"]; got == 0 {
			t.Fatalf("%s should emit typed-pattern slot jumps", name)
		}
		if got := fn.TrackedOpcodeHits["LoadSlotStructField"]; got == 0 {
			t.Fatalf("%s should emit slot struct-field loads", name)
		}
	}
}

func bytecodeOpcodeAuditFunctionByName(row bytecodeOpcodeAuditBenchmarkRow, name string) (bytecodeOpcodeAuditFunctionRow, bool) {
	for _, fn := range row.Functions {
		if fn.Name == name {
			return fn, true
		}
	}
	return bytecodeOpcodeAuditFunctionRow{}, false
}

func runBytecodeBenchWarmup(t *testing.T, fixturePath string) {
	t.Helper()

	root := repositoryRoot()
	if root == "" {
		t.Fatalf("repository root not found")
	}
	dir := filepath.Join(root, filepath.FromSlash(fixturePath))
	entryPath := filepath.Join(dir, "main.able")

	restoreWD, err := chdirBenchRuntime(dir)
	if err != nil {
		t.Fatalf("chdir run-from: %v", err)
	}
	defer restoreWD()

	searchPaths, err := buildExecSearchPaths(entryPath, dir, fixtureManifest{})
	if err != nil {
		t.Fatalf("bench search paths: %v", err)
	}
	loader, err := driver.NewLoader(searchPaths)
	if err != nil {
		t.Fatalf("loader init: %v", err)
	}
	defer loader.Close()

	program, err := loader.Load(entryPath)
	if err != nil {
		t.Fatalf("load program: %v", err)
	}

	executor, err := NewExecutorFromEnvironment()
	if err != nil {
		t.Fatalf("executor from environment: %v", err)
	}
	if closer, ok := executor.(interface{ Close() }); ok {
		defer closer.Close()
	}

	interp := NewBytecodeWithExecutor(executor)
	registerBenchPrint(interp)

	_, entryEnv, _, err := interp.EvaluateProgram(program, ProgramEvaluationOptions{
		SkipTypecheck:    true,
		AllowDiagnostics: false,
	})
	if err != nil {
		t.Fatalf("evaluate program: %v", err)
	}

	env := entryEnv
	if env == nil {
		env = interp.GlobalEnvironment()
	}
	mainValue, err := env.Get("main")
	if err != nil {
		t.Fatalf("lookup main: %v", err)
	}

	goRuntime.GC()
	if _, err := interp.CallFunction(mainValue, nil); err != nil {
		var searchPathSummary []string
		for _, sp := range searchPaths {
			searchPathSummary = append(searchPathSummary, fmt.Sprintf("%s[%d,%s]", sp.Path, sp.Kind, sp.StdlibSource))
		}
		var moduleSummary []string
		for _, mod := range program.Modules {
			if mod == nil {
				continue
			}
			moduleSummary = append(moduleSummary, fmt.Sprintf("%s=>%s", mod.Package, strings.Join(mod.Files, ",")))
		}
		t.Fatalf(
			"warmup call main: %v; search_paths=%v; modules=%v; trace=%#v",
			err,
			searchPathSummary,
			moduleSummary,
			interp.BytecodeTrace(20),
		)
	}
	executor.Flush()
}
