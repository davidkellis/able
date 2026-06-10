package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_StaticMemberPackageIdentityUsesPrecomputedKeyWithoutAllocation(t *testing.T) {
	name := "able.math"
	namePath := []string{"able", "math"}
	identityKey := bytecodePackageIdentityKey(name, namePath)
	want, ok := bytecodeStaticMemberPackageIdentity(name, namePath, "", bytecodeStaticMemberReceiverPackage)
	if !ok {
		t.Fatal("fallback package identity was not available")
	}
	got, ok := bytecodeStaticMemberPackageIdentity(name, namePath, identityKey, bytecodeStaticMemberReceiverPackage)
	if !ok || got != want {
		t.Fatalf("precomputed package identity = %#v, %v; want %#v, true", got, ok, want)
	}

	allocs := testing.AllocsPerRun(1000, func() {
		got, ok = bytecodeStaticMemberReceiverIdentityForValue(runtime.PackageValue{
			Name:        name,
			NamePath:    namePath,
			IdentityKey: identityKey,
		})
	})
	if !ok || got != want {
		t.Fatalf("runtime package identity = %#v, %v; want %#v, true", got, ok, want)
	}
	if allocs != 0 {
		t.Fatalf("precomputed package identity allocations = %.2f, want 0", allocs)
	}
}

func TestBytecodeVM_StaticMemberCallInlinesSingleOverloadSlotlessFunction(t *testing.T) {
	t.Setenv("ABLE_BYTECODE_STATS", "1")

	interp := NewBytecode()
	env := interp.GlobalEnvironment()
	vm := newBytecodeVM(interp, env)
	fnDef := ast.Fn(
		"value",
		[]*ast.FunctionParameter{ast.Param("n", ast.Ty("i32"))},
		[]ast.Statement{ast.ID("n")},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)
	fnProgram := finalizeBytecodeProgramMetadata(&bytecodeProgram{
		instructions: []bytecodeInstruction{
			{op: bytecodeOpLoadName, name: "n", nameSimple: true},
			{op: bytecodeOpReturn},
		},
	})
	fn := &runtime.FunctionValue{Declaration: fnDef, Closure: env}
	setFunctionBytecodeProgram(fn, fnProgram)
	overload := &runtime.FunctionOverloadValue{Overloads: []*runtime.FunctionValue{fn}}
	receiver := &runtime.StructDefinitionValue{Node: ast.StructDef("Box", nil, ast.StructKindNamed, nil, nil, false)}
	callNode := ast.NewFunctionCall(ast.Member(ast.ID("Box"), "value"), []ast.Expression{ast.Int(9)}, nil, false)
	vm.stack = []runtime.Value{receiver, runtime.NewSmallInt(9, runtime.IntegerI32)}

	newProg, err := vm.execStaticMemberCallable(
		overload,
		bytecodeInstruction{name: "value", argCount: 1},
		0,
		1,
		callNode,
		callNode,
		&bytecodeProgram{},
		true,
	)
	if err != nil {
		t.Fatalf("single-overload static call failed: %v", err)
	}
	if newProg != fnProgram {
		t.Fatalf("single-overload static call returned program %#v, want %#v", newProg, fnProgram)
	}
	if len(vm.stack) != 0 {
		t.Fatalf("stack len after inline setup = %d, want 0", len(vm.stack))
	}
	if got, err := vm.env.Get("n"); err != nil || !valuesEqual(got, runtime.NewSmallInt(9, runtime.IntegerI32)) {
		t.Fatalf("inline call env binding n = %#v, %v; want 9", got, err)
	}

	stats := interp.BytecodeStats()
	if stats.CallMemberStaticInlineHits != 1 {
		t.Fatalf("CallMemberStaticInlineHits = %d, want 1; stats=%#v", stats.CallMemberStaticInlineHits, stats)
	}
	if stats.CallMemberStaticGenericHits != 0 {
		t.Fatalf("CallMemberStaticGenericHits = %d, want 0; stats=%#v", stats.CallMemberStaticGenericHits, stats)
	}
}

func TestBytecodeVM_StaticMemberCallLoweringAvoidsReceiverLoadName(t *testing.T) {
	t.Setenv("ABLE_BYTECODE_STATS", "1")

	boxDef := ast.StructDef("Box", nil, ast.StructKindNamed, nil, nil, false)
	valueFn := ast.Fn(
		"value",
		nil,
		[]ast.Statement{ast.Int(7)},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)
	module := ast.Mod([]ast.Statement{
		boxDef,
		ast.Methods(ast.Ty("Box"), []*ast.FunctionDefinition{valueFn}, nil, nil),
		ast.Assign(ast.ID("i"), ast.Int(0)),
		ast.Assign(ast.ID("sum"), ast.Int(0)),
		ast.Loop(
			ast.Iff(ast.Bin(">=", ast.ID("i"), ast.Int(4)), ast.Brk(nil, nil)),
			ast.AssignOp(
				ast.AssignmentAssign,
				ast.ID("sum"),
				ast.Bin("+", ast.ID("sum"), ast.CallExpr(ast.Member(ast.ID("Box"), "value"))),
			),
			ast.AssignOp(
				ast.AssignmentAssign,
				ast.ID("i"),
				ast.Bin("+", ast.ID("i"), ast.Int(1)),
			),
		),
		ast.ID("sum"),
	}, nil, nil)

	interp := NewBytecode()
	program, err := interp.lowerModuleToBytecode(module)
	if err != nil {
		t.Fatalf("lowerModuleToBytecode: %v", err)
	}
	if !bytecodeProgramContainsOpcode(program, bytecodeOpLoadStaticReceiver) {
		t.Fatalf("expected static receiver load opcode")
	}
	if !bytecodeProgramContainsOpcode(program, bytecodeOpCallStaticMember) {
		t.Fatalf("expected static member call opcode")
	}
	for _, instr := range program.instructions {
		if instr.op == bytecodeOpLoadName && instr.name == "Box" {
			t.Fatalf("static receiver should not lower as LoadName Box")
		}
	}

	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	got, err := vm.run(program)
	if err != nil {
		t.Fatalf("bytecode execution failed: %v", err)
	}
	if intVal, ok := got.(runtime.IntegerValue); !ok || intVal.BigInt().Int64() != 28 {
		t.Fatalf("static member loop result = %#v, want 28", got)
	}
	stats := interp.BytecodeStats()
	if got := stats.LoadNameLookupsByName["Box"]; got != 0 {
		t.Fatalf("static receiver should avoid Box LoadName lookups, got %d", got)
	}
	if stats.CallMemberStaticCacheHits == 0 {
		t.Fatalf("expected static member cache hits, got stats=%#v", stats)
	}
}

func TestBytecodeVM_StaticMemberCandidateFallsBackForOrdinaryEnvReceiver(t *testing.T) {
	boxDef := ast.StructDef(
		"Box",
		[]*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("i32"), "n")},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	valueFn := ast.Fn(
		"value",
		[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Self"))},
		[]ast.Statement{ast.Member(ast.ID("self"), "n")},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)
	module := ast.Mod([]ast.Statement{
		boxDef,
		ast.Methods(ast.Ty("Box"), []*ast.FunctionDefinition{valueFn}, nil, nil),
		ast.Assign(ast.ID("box"), ast.StructLit([]*ast.StructFieldInitializer{
			ast.FieldInit(ast.Int(41), "n"),
		}, false, "Box", nil, nil)),
		ast.CallExpr(ast.Member(ast.ID("box"), "value")),
	}, nil, nil)

	interp := NewBytecode()
	program, err := interp.lowerModuleToBytecode(module)
	if err != nil {
		t.Fatalf("lowerModuleToBytecode: %v", err)
	}
	if !bytecodeProgramContainsOpcode(program, bytecodeOpCallStaticMember) {
		t.Fatalf("expected env receiver member call to use static-candidate opcode")
	}
	got := runBytecodeModuleWithInterpreter(t, interp, module)
	if intVal, ok := got.(runtime.IntegerValue); !ok || intVal.BigInt().Int64() != 41 {
		t.Fatalf("ordinary receiver fallback result = %#v, want 41", got)
	}
}

func TestBytecodeVM_StaticReceiverLoadsBeforeArguments(t *testing.T) {
	interp := NewBytecode()
	module := ast.Mod([]ast.Statement{
		ast.Assign(ast.ID("called"), ast.Int(0)),
		ast.CallExpr(
			ast.Member(ast.ID("Missing"), "value"),
			ast.Assign(ast.ID("called"), ast.Int(1)),
		),
	}, nil, nil)

	err := runBytecodeModuleError(t, interp, module)
	if err == nil {
		t.Fatalf("expected missing receiver error")
	}
	called, lookupErr := interp.GlobalEnvironment().Get("called")
	if lookupErr != nil {
		t.Fatalf("called lookup failed: %v", lookupErr)
	}
	if intVal, ok := called.(runtime.IntegerValue); !ok || intVal.BigInt().Int64() != 0 {
		t.Fatalf("argument side effect ran before missing receiver error: called=%#v", called)
	}
}
