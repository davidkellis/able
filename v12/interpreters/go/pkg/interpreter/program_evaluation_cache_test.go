package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/driver"
	"able/interpreter-go/pkg/runtime"
)

func testCachedProgramModule(pkg string, body []ast.Statement, imports []*ast.ImportStatement, file string) *driver.Module {
	mod := &driver.Module{
		Package: pkg,
		AST:     ast.Mod(body, imports, ast.Pkg([]interface{}{pkg}, false)),
		Files:   []string{file},
	}
	origins := make(map[ast.Node]string)
	ast.AnnotateOrigins(mod.AST, file, origins)
	mod.NodeOrigins = origins
	return mod
}

func TestCachedProgramNodeOriginsReusesMergedMap(t *testing.T) {
	depModule := testCachedProgramModule("dep", []ast.Statement{
		ast.Fn("provide", nil, []ast.Statement{ast.Ret(ast.Int(42))}, ast.Ty("i32"), nil, nil, false, false),
	}, nil, "dep/lib.able")
	mainModule := testCachedProgramModule("main", []ast.Statement{
		ast.Fn("main", nil, []ast.Statement{ast.Ret(ast.CallExpr(ast.Member(ast.ID("dep"), ast.ID("provide"))))}, ast.Ty("i32"), nil, nil, false, false),
	}, []*ast.ImportStatement{ast.Imp([]interface{}{"dep"}, false, nil, nil)}, "main/main.able")
	program := &driver.Program{
		Entry:   mainModule,
		Modules: []*driver.Module{depModule, mainModule},
	}
	t.Cleanup(func() { programEvaluationStateCache.Delete(program) })

	first := cachedProgramNodeOrigins(program)
	second := cachedProgramNodeOrigins(program)
	if first == nil || second == nil {
		t.Fatalf("expected merged node origins")
	}
	sentinel := ast.ID("sentinel")
	first[sentinel] = "sentinel.able"
	if got := second[sentinel]; got != "sentinel.able" {
		t.Fatalf("expected cached node origins map reuse, got %q", got)
	}
}

func TestCachedLoadedModuleBytecodeProgramSharedAcrossInterpreters(t *testing.T) {
	module := testCachedProgramModule("main", []ast.Statement{
		ast.Fn("main", nil, []ast.Statement{ast.Ret(ast.Int(42))}, ast.Ty("i32"), nil, nil, false, false),
	}, nil, "main/main.able")
	t.Cleanup(func() { loadedModuleBytecodeCache.Delete(module) })

	first, err := cachedLoadedModuleBytecodeProgram(NewBytecode(), module)
	if err != nil {
		t.Fatalf("cachedLoadedModuleBytecodeProgram first call: %v", err)
	}
	second, err := cachedLoadedModuleBytecodeProgram(NewBytecode(), module)
	if err != nil {
		t.Fatalf("cachedLoadedModuleBytecodeProgram second call: %v", err)
	}
	if first == nil || second == nil {
		t.Fatalf("expected cached lowered programs")
	}
	if first != second {
		t.Fatalf("expected loaded-module cache reuse")
	}

	fresh, err := NewBytecode().lowerModuleToBytecode(module.AST)
	if err != nil {
		t.Fatalf("lowerModuleToBytecode fresh call: %v", err)
	}
	if fresh == first {
		t.Fatalf("expected direct lowerModuleToBytecode to remain uncached")
	}
}

func TestEvaluateProgramBytecodeReusesLoadedModuleLowering(t *testing.T) {
	module := testCachedProgramModule("main", []ast.Statement{
		ast.Fn("main", nil, []ast.Statement{ast.Ret(ast.Int(42))}, ast.Ty("i32"), nil, nil, false, false),
		ast.CallExpr(ast.ID("main")),
	}, nil, "main/main.able")
	program := &driver.Program{
		Entry:   module,
		Modules: []*driver.Module{module},
	}
	t.Cleanup(func() {
		programEvaluationStateCache.Delete(program)
		loadedModuleBytecodeCache.Delete(module)
	})

	value, _, _, err := NewBytecode().EvaluateProgram(program, ProgramEvaluationOptions{SkipTypecheck: true})
	if err != nil {
		t.Fatalf("first EvaluateProgram: %v", err)
	}
	got, ok := value.(runtime.IntegerValue)
	gotInt, okInt := got.ToInt64()
	if !ok || !okInt || gotInt != 42 {
		t.Fatalf("first EvaluateProgram returned %#v, want 42", value)
	}

	cachedAny, ok := loadedModuleBytecodeCache.Load(module)
	if !ok {
		t.Fatalf("expected loaded module bytecode cache entry after first evaluation")
	}
	cachedProgram := cachedAny.(*cachedLoadedModuleBytecodeState).program
	if cachedProgram == nil {
		t.Fatalf("expected cached lowered module program")
	}

	value, _, _, err = NewBytecode().EvaluateProgram(program, ProgramEvaluationOptions{SkipTypecheck: true})
	if err != nil {
		t.Fatalf("second EvaluateProgram: %v", err)
	}
	got, ok = value.(runtime.IntegerValue)
	gotInt, okInt = got.ToInt64()
	if !ok || !okInt || gotInt != 42 {
		t.Fatalf("second EvaluateProgram returned %#v, want 42", value)
	}

	cachedAgain, ok := loadedModuleBytecodeCache.Load(module)
	if !ok {
		t.Fatalf("expected loaded module bytecode cache entry after second evaluation")
	}
	if cachedAgain.(*cachedLoadedModuleBytecodeState).program != cachedProgram {
		t.Fatalf("expected second evaluation to reuse cached lowered module program")
	}
}
