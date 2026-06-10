package interpreter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/driver"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_StructLiteralNamedFastInFunctionBody(t *testing.T) {
	pointDef := ast.StructDef(
		"Point",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("i32"), "x"),
			ast.FieldDef(ast.Ty("i32"), "y"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	makePoint := ast.Fn(
		"make_point",
		nil,
		[]ast.Statement{
			ast.Ret(ast.StructLit(
				[]*ast.StructFieldInitializer{
					ast.FieldInit(ast.Int(3), "x"),
					ast.FieldInit(ast.Int(5), "y"),
				},
				false,
				"Point",
				nil,
				nil,
			)),
		},
		ast.Ty("Point"),
		nil,
		nil,
		false,
		false,
	)

	interp := NewBytecode()
	env := interp.GlobalEnvironment()
	if _, err := interp.evaluateStructDefinition(pointDef, env); err != nil {
		t.Fatalf("evaluateStructDefinition failed: %v", err)
	}
	if _, err := interp.evaluateFunctionDefinition(makePoint, env); err != nil {
		t.Fatalf("evaluateFunctionDefinition failed: %v", err)
	}

	raw, ok := env.Lookup("make_point")
	if !ok {
		t.Fatalf("make_point not defined")
	}
	fn, ok := raw.(*runtime.FunctionValue)
	if !ok || fn == nil {
		t.Fatalf("make_point binding = %#v, want *runtime.FunctionValue", raw)
	}
	program, ok := fn.Bytecode.(*bytecodeProgram)
	if !ok || program == nil {
		t.Fatalf("make_point bytecode = %#v, want *bytecodeProgram", fn.Bytecode)
	}
	found := false
	foundPlan := false
	for ip, instr := range program.instructions {
		if instr.op == bytecodeOpStructLiteralNamedFast {
			found = true
			plan, ok := program.namedStructLiterals[ip]
			if !ok {
				t.Fatalf("named struct literal plan missing at ip=%d", ip)
			}
			if plan.definition == nil || plan.definition.Node != pointDef {
				t.Fatalf("named struct literal definition plan = %#v, want Point definition", plan.definition)
			}
			if len(plan.fieldOrder) != 2 || plan.fieldOrder[0] != 0 || plan.fieldOrder[1] != 1 {
				t.Fatalf("named struct literal field order = %#v, want [0 1]", plan.fieldOrder)
			}
			foundPlan = true
			break
		}
	}
	if !found {
		t.Fatalf("bytecode struct literal fast opcode not emitted in function body")
	}
	if !foundPlan {
		t.Fatalf("bytecode struct literal fast plan not emitted in function body")
	}

	module := ast.Mod([]ast.Statement{
		pointDef,
		makePoint,
		ast.Member(ast.CallExpr(ast.ID("make_point")), "x"),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode function-body struct literal mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_ParsedShorthandNamedStructLiteralUsesFastOpcode(t *testing.T) {
	module := mustParseModuleSource(t, `
struct FileLines {
  handle: i32,
  closed: bool
}

fn lines(handle: i32) -> FileLines {
  FileLines { handle, closed: false }
}

lines(7).handle
`)

	want := mustEvalModule(t, New(), module)
	interp := NewBytecode()
	got := runBytecodeModuleWithInterpreter(t, interp, module)
	if !valuesEqual(got, want) {
		t.Fatalf("parsed shorthand struct literal mismatch: got=%#v want=%#v", got, want)
	}

	raw, ok := interp.GlobalEnvironment().Lookup("lines")
	if !ok {
		t.Fatal("lines not defined")
	}
	fn, ok := raw.(*runtime.FunctionValue)
	if !ok || fn == nil {
		t.Fatalf("lines binding = %#v, want *runtime.FunctionValue", raw)
	}
	program, ok := fn.Bytecode.(*bytecodeProgram)
	if !ok || program == nil {
		t.Fatalf("lines bytecode = %#v, want *bytecodeProgram", fn.Bytecode)
	}
	if !bytecodeProgramContainsOpcode(program, bytecodeOpStructLiteralNamedFast) {
		t.Fatalf("parsed shorthand literal did not use the named struct literal opcode")
	}
}

func TestBytecodeVM_StdlibFsLinesUsesNamedStructLiteralOpcode(t *testing.T) {
	root := repositoryRoot()
	stdlibRoot := filepath.Join(root, "..", "able-stdlib", "src")
	if _, err := os.Stat(filepath.Join(stdlibRoot, "fs.able")); os.IsNotExist(err) {
		t.Skipf("canonical able-stdlib checkout not found at %s", stdlibRoot)
	} else if err != nil {
		t.Fatalf("stat canonical able-stdlib: %v", err)
	}
	t.Setenv("ABLE_STDLIB_ROOT", stdlibRoot)

	fixtureDir := filepath.Join(root, "v12", "fixtures", "exec", "06_12_28_stdlib_fs_lines")
	entryPath := filepath.Join(fixtureDir, "main.able")
	searchPaths, err := buildExecSearchPaths(entryPath, fixtureDir, fixtureManifest{})
	if err != nil {
		t.Fatalf("build fixture search paths: %v", err)
	}
	loader, err := driver.NewLoader(searchPaths)
	if err != nil {
		t.Fatalf("new loader: %v", err)
	}
	defer loader.Close()
	loaded, err := loader.Load(entryPath)
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}

	executor := NewSerialExecutor(nil)
	defer executor.Close()
	interp := newTestInterpreter(t, testExecBytecode, executor)
	mode := configureFixtureTypechecker(interp)
	_, _, _, err = interp.EvaluateProgram(loaded, ProgramEvaluationOptions{
		SkipTypecheck:    mode == typecheckModeOff,
		AllowDiagnostics: mode != typecheckModeOff,
	})
	if err != nil {
		t.Fatalf("evaluate fixture modules: %v", err)
	}
	fsPackage := ""
	for _, module := range loaded.Modules {
		if module == nil {
			continue
		}
		for _, file := range module.Files {
			if filepath.Clean(file) == filepath.Join(stdlibRoot, "fs.able") {
				if module.AST != nil && module.AST.Package != nil {
					fsPackage = strings.Join(identifiersToStrings(module.AST.Package.NamePath), ".")
				}
				break
			}
		}
		if fsPackage != "" {
			break
		}
	}
	if fsPackage == "" {
		t.Fatal("canonical fs module unavailable in loaded fixture")
	}
	fsEnv := interp.PackageEnvironment(fsPackage)
	if fsEnv == nil {
		known := make([]string, 0, len(interp.packageEnvs))
		for name := range interp.packageEnvs {
			known = append(known, name)
		}
		t.Fatalf("%s package environment unavailable; known packages: %v", fsPackage, known)
	}
	raw, ok := fsEnv.Lookup("lines")
	if !ok {
		t.Fatal("fs.lines not defined")
	}
	fn, ok := raw.(*runtime.FunctionValue)
	if !ok || fn == nil {
		t.Fatalf("fs.lines binding = %#v, want *runtime.FunctionValue", raw)
	}
	program, ok := fn.Bytecode.(*bytecodeProgram)
	if !ok || program == nil {
		t.Fatalf("fs.lines bytecode = %#v, want *bytecodeProgram", fn.Bytecode)
	}
	linesLiteralPlan, found := bytecodeNamedStructLiteralPlanForOpcode(program)
	if !found {
		t.Fatalf("fs.lines did not lower FileLines through named struct literal opcode")
	}
	if linesLiteralPlan.definition == nil || linesLiteralPlan.definition.Node == nil ||
		linesLiteralPlan.definition.Node.ID == nil || linesLiteralPlan.definition.Node.ID.Name != "FileLines" {
		t.Fatalf("fs.lines FileLines literal definition plan = %#v, want FileLines", linesLiteralPlan.definition)
	}
	path := filepath.Join(t.TempDir(), "lines.txt")
	if err := os.WriteFile(path, []byte("first\nsecond\n"), 0o600); err != nil {
		t.Fatalf("write line fixture: %v", err)
	}
	value, err := interp.CallFunction(fn, []runtime.Value{runtime.StringValue{Val: path}})
	if err != nil {
		t.Fatalf("call fs.lines: %v", err)
	}
	fileLines, ok := value.(*runtime.StructInstanceValue)
	if !ok {
		t.Fatalf("fs.lines returned %T, want *runtime.StructInstanceValue", value)
	}
	if fileLines.Definition == nil || fileLines.Definition.Node == nil || fileLines.Definition.Node.ID == nil || fileLines.Definition.Node.ID.Name != "FileLines" {
		t.Fatalf("fs.lines returned definition %#v, want FileLines", fileLines.Definition)
	}
	if _, ok := structNamedFieldValue(fileLines, "closed"); !ok {
		t.Fatalf("fs.lines result lacks closed field: %#v", fileLines)
	}
	closeMethod, err := interp.memberAccessOnValueWithOptions(fileLines, ast.ID("close"), fsEnv, true)
	if err != nil {
		t.Fatalf("resolve FileLines.close: %v", err)
	}
	if _, err := interp.CallFunction(closeMethod, nil); err != nil {
		t.Fatalf("FileLines.close: %v", err)
	}
	closedValue, ok := structNamedFieldValue(fileLines, "closed")
	if !ok || !valuesEqual(closedValue, runtime.BoolValue{Val: true}) {
		t.Fatalf("FileLines.close did not set closed=true: %#v", fileLines)
	}

	value, err = interp.CallFunction(fn, []runtime.Value{runtime.StringValue{Val: path}})
	if err != nil {
		t.Fatalf("call fresh fs.lines: %v", err)
	}
	fileLines, ok = value.(*runtime.StructInstanceValue)
	if !ok {
		t.Fatalf("fresh fs.lines returned %T, want *runtime.StructInstanceValue", value)
	}
	next, err := interp.memberAccessOnValueWithOptions(fileLines, ast.ID("next"), fsEnv, true)
	if err != nil {
		t.Fatalf("resolve FileLines.next: %v", err)
	}
	for call, want := range []string{"first", "second"} {
		got, err := interp.CallFunction(next, nil)
		if err != nil {
			t.Fatalf("FileLines.next call %d: %v", call, err)
		}
		if !valuesEqual(got, runtime.StringValue{Val: want}) {
			t.Fatalf("FileLines.next call %d = %#v, want %q", call, got, want)
		}
	}
	if _, err := interp.CallFunction(next, nil); err != nil {
		t.Fatalf("FileLines.next EOF: %v", err)
	}
	closedValue, ok = structNamedFieldValue(fileLines, "closed")
	if !ok || !valuesEqual(closedValue, runtime.BoolValue{Val: true}) {
		t.Fatalf("FileLines EOF did not set closed=true: %#v", fileLines)
	}
}

func bytecodeNamedStructLiteralPlanForOpcode(program *bytecodeProgram) (bytecodeNamedStructLiteralPlan, bool) {
	if program == nil {
		return bytecodeNamedStructLiteralPlan{}, false
	}
	for ip, instruction := range program.instructions {
		if instruction.op != bytecodeOpStructLiteralNamedFast {
			continue
		}
		plan, ok := program.namedStructLiterals[ip]
		return plan, ok
	}
	return bytecodeNamedStructLiteralPlan{}, false
}

func TestBytecodeVM_GenericStructLiteralNamedFastInFunctionBody(t *testing.T) {
	boxDef := ast.StructDef(
		"Box",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("T"), "value"),
		},
		ast.StructKindNamed,
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
	)
	makeBox := ast.Fn(
		"make_box",
		[]*ast.FunctionParameter{
			ast.Param("value", ast.Ty("i32")),
		},
		[]ast.Statement{
			ast.Ret(ast.StructLit(
				[]*ast.StructFieldInitializer{
					ast.FieldInit(ast.ID("value"), "value"),
				},
				false,
				"Box",
				nil,
				nil,
			)),
		},
		ast.Gen(ast.Ty("Box"), ast.Ty("i32")),
		nil,
		nil,
		false,
		false,
	)

	interp := NewBytecode()
	env := interp.GlobalEnvironment()
	if _, err := interp.evaluateStructDefinition(boxDef, env); err != nil {
		t.Fatalf("evaluateStructDefinition failed: %v", err)
	}
	if _, err := interp.evaluateFunctionDefinition(makeBox, env); err != nil {
		t.Fatalf("evaluateFunctionDefinition failed: %v", err)
	}

	raw, ok := env.Lookup("make_box")
	if !ok {
		t.Fatalf("make_box not defined")
	}
	fn, ok := raw.(*runtime.FunctionValue)
	if !ok || fn == nil {
		t.Fatalf("make_box binding = %#v, want *runtime.FunctionValue", raw)
	}
	program, ok := fn.Bytecode.(*bytecodeProgram)
	if !ok || program == nil {
		t.Fatalf("make_box bytecode = %#v, want *bytecodeProgram", fn.Bytecode)
	}
	found := false
	for ip, instr := range program.instructions {
		if instr.op != bytecodeOpStructLiteralNamedFast {
			continue
		}
		found = true
		plan, ok := program.namedStructLiterals[ip]
		if !ok {
			t.Fatalf("named struct literal plan missing at ip=%d", ip)
		}
		if plan.definition == nil || plan.definition.Node != boxDef {
			t.Fatalf("named struct literal definition plan = %#v, want Box definition", plan.definition)
		}
		if len(plan.fieldOrder) != 1 || plan.fieldOrder[0] != 0 {
			t.Fatalf("named struct literal field order = %#v, want [0]", plan.fieldOrder)
		}
		break
	}
	if !found {
		t.Fatalf("bytecode generic struct literal fast opcode not emitted")
	}

	module := ast.Mod([]ast.Statement{
		boxDef,
		makeBox,
		ast.Member(ast.Call("make_box", ast.Int(11)), "value"),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode generic function-body struct literal mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_ArrayStructLiteralNamedFastReturnsArrayValue(t *testing.T) {
	arrayDef := ast.StructDef(
		"Array",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("i64"), "storage_handle"),
			ast.FieldDef(ast.Ty("i32"), "length"),
			ast.FieldDef(ast.Ty("i32"), "capacity"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	structDef := &runtime.StructDefinitionValue{Node: arrayDef}
	lit := ast.StructLit(
		[]*ast.StructFieldInitializer{
			ast.FieldInit(ast.Int(0), "storage_handle"),
			ast.FieldInit(ast.Int(0), "length"),
			ast.FieldInit(ast.Int(0), "capacity"),
		},
		false,
		"Array",
		nil,
		nil,
	)
	handle := runtime.ArrayStoreNewReservedCapacity(8)
	instr := bytecodeInstruction{
		op:       bytecodeOpStructLiteralNamedFast,
		argCount: 3,
		node:     lit,
	}
	program := &bytecodeProgram{
		namedStructLiterals: map[int]bytecodeNamedStructLiteralPlan{
			0: {
				definition: structDef,
				fieldOrder: []int{0, 1, 2},
			},
		},
	}
	interp := NewBytecode()
	vm := interp.acquireBytecodeVM(interp.GlobalEnvironment())
	defer interp.releaseBytecodeVM(vm)
	vm.stack = append(vm.stack,
		runtime.NewSmallInt(handle, runtime.IntegerI64),
		runtime.NewSmallInt(0, runtime.IntegerI32),
		runtime.NewSmallInt(8, runtime.IntegerI32),
	)

	if err := vm.execStructLiteralNamedFast(&instr, program); err != nil {
		t.Fatalf("execStructLiteralNamedFast(Array): %v", err)
	}
	if len(vm.stack) != 1 {
		t.Fatalf("stack len = %d, want 1", len(vm.stack))
	}
	arr, ok := vm.stack[0].(*runtime.ArrayValue)
	if !ok {
		t.Fatalf("result = %T (%#v), want *runtime.ArrayValue", vm.stack[0], vm.stack[0])
	}
	if arr.Handle != handle {
		t.Fatalf("array handle = %d, want %d", arr.Handle, handle)
	}
	if arr.State == nil || arr.State.Capacity != 8 {
		t.Fatalf("array state = %#v, want capacity 8", arr.State)
	}
}

func TestBytecodeVM_StructLiteralNamedFastStabilizesRawReturnScratch(t *testing.T) {
	snapshotDef := ast.StructDef(
		"Snapshot",
		[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("u64"), "value")},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	definition := &runtime.StructDefinitionValue{Node: snapshotDef}
	literal := ast.StructLit(
		[]*ast.StructFieldInitializer{ast.FieldInit(ast.Int(0), "value")},
		false,
		"Snapshot",
		nil,
		nil,
	)
	program := &bytecodeProgram{namedStructLiterals: map[int]bytecodeNamedStructLiteralPlan{
		0: {definition: definition, fieldOrder: []int{0}},
	}}
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.stack = append(vm.stack, vm.rawIntegerReturnValue(runtime.IntegerU64, 42))

	instr := bytecodeInstruction{op: bytecodeOpStructLiteralNamedFast, argCount: 1, node: literal}
	if err := vm.execStructLiteralNamedFast(&instr, program); err != nil {
		t.Fatalf("execStructLiteralNamedFast: %v", err)
	}
	if len(vm.stack) != 1 {
		t.Fatalf("stack size = %d, want 1", len(vm.stack))
	}
	snapshot, ok := vm.stack[0].(*runtime.StructInstanceValue)
	if !ok || len(snapshot.Positional) != 1 {
		t.Fatalf("result = %#v, want one-field Snapshot", vm.stack[0])
	}

	vm.rawIntegerReturnScratch.Raw = 99
	assertIntValue(t, snapshot.Positional[0], runtime.IntegerU64, 42)
}

func TestBytecodeVM_LowerExpressionSingletonStructLiteralFastWithoutEnvDefinition(t *testing.T) {
	treeEmptyDef := ast.StructDef(
		"TreeEmpty",
		nil,
		ast.StructKindSingleton,
		nil,
		nil,
		false,
	)
	expr := ast.StructLit(nil, false, "TreeEmpty", nil, nil)

	interp := NewBytecode()
	env := interp.GlobalEnvironment()
	if _, err := interp.evaluateStructDefinition(treeEmptyDef, env); err != nil {
		t.Fatalf("evaluateStructDefinition failed: %v", err)
	}

	program, err := interp.lowerExpressionToBytecode(expr)
	if err != nil {
		t.Fatalf("lowerExpressionToBytecode failed: %v", err)
	}
	if !bytecodeProgramContainsOpcode(program, bytecodeOpStructLiteralNamedFast) {
		t.Fatalf("expected fast named struct literal opcode from env-free expression lowering")
	}

	found := false
	for ip, instr := range program.instructions {
		if instr.op != bytecodeOpStructLiteralNamedFast {
			continue
		}
		found = true
		plan, ok := program.namedStructLiterals[ip]
		if !ok {
			t.Fatalf("named struct literal plan missing at ip=%d", ip)
		}
		if plan.definition != nil {
			t.Fatalf("expected env-free expression lowering to defer definition lookup, got %#v", plan.definition)
		}
		if len(plan.fieldOrder) != 0 {
			t.Fatalf("expected env-free expression lowering to defer field-order plan, got %#v", plan.fieldOrder)
		}
	}
	if !found {
		t.Fatalf("fast named struct literal instruction missing")
	}

	vm := interp.acquireBytecodeVM(env)
	defer interp.releaseBytecodeVM(vm)
	got, err := vm.run(program)
	if err != nil {
		t.Fatalf("run lowered expression: %v", err)
	}
	defValue, ok := got.(*runtime.StructDefinitionValue)
	if !ok || defValue == nil {
		t.Fatalf("expected singleton struct definition result, got %T (%#v)", got, got)
	}
	if defValue.Node != treeEmptyDef {
		t.Fatalf("unexpected singleton definition on result: %#v", defValue)
	}
}
