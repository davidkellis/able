package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_LoweringEmitsTypedFloatRegionForMultiOperationDeclaration(t *testing.T) {
	def := ast.Fn(
		"step",
		nil,
		[]ast.Statement{
			ast.Assign(ast.ID("a"), ast.Flt(2)),
			ast.Assign(ast.ID("b"), ast.Flt(3)),
			ast.Assign(ast.ID("c"), ast.Flt(4)),
			ast.Assign(ast.ID("d"), ast.Flt(5)),
			ast.Assign(ast.ID("result"), ast.Bin("+", ast.Bin("*", ast.ID("a"), ast.ID("b")), ast.Bin("*", ast.ID("c"), ast.ID("d")))),
			ast.ID("result"),
		},
		ast.Ty("f64"),
		nil,
		nil,
		false,
		false,
	)

	program, err := NewBytecode().lowerFunctionDefinitionBytecode(def)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	if !bytecodeProgramContainsOpcode(program, bytecodeOpStoreSlotFloatRegion) {
		t.Fatal("expected lowering to emit typed float region")
	}
	if len(program.floatRegions) != 1 {
		t.Fatalf("float region count = %d, want 1", len(program.floatRegions))
	}
	plan := program.floatRegions[0]
	if len(plan.steps) != 7 || plan.maxDepth != 3 {
		t.Fatalf("float region plan = %#v, want seven steps and depth three", plan)
	}
}

func TestBytecodeVM_TypedFloatRegionPreservesF64Parity(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Fn(
			"step",
			nil,
			[]ast.Statement{
				ast.Assign(ast.ID("a"), ast.Flt(2)),
				ast.Assign(ast.ID("b"), ast.Flt(3)),
				ast.Assign(ast.ID("c"), ast.Flt(4)),
				ast.Assign(ast.ID("d"), ast.Flt(5)),
				ast.Assign(ast.ID("result"), ast.Bin("-", ast.Bin("*", ast.ID("a"), ast.ID("b")), ast.Bin("/", ast.ID("d"), ast.ID("c")))),
				ast.ID("result"),
			},
			ast.Ty("f64"),
			nil,
			nil,
			false,
			false,
		),
		ast.Call("step"),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("typed float region mismatch: got=%#v want=%#v", got, want)
	}
	assertFloatValue(t, got, runtime.FloatF64, 4.75)
}

func TestBytecodeVM_TypedFloatRegionNormalizesF32AtEveryOperation(t *testing.T) {
	f32 := ast.FloatType(runtime.FloatF32)
	module := ast.Mod([]ast.Statement{
		ast.Fn(
			"step",
			nil,
			[]ast.Statement{
				ast.Assign(ast.ID("a"), ast.FltTyped(1.1, &f32)),
				ast.Assign(ast.ID("b"), ast.FltTyped(0.6, &f32)),
				ast.Assign(ast.ID("c"), ast.FltTyped(0.2, &f32)),
				ast.Assign(ast.ID("result"), ast.Bin("+", ast.Bin("*", ast.ID("a"), ast.ID("b")), ast.Bin("-", ast.ID("a"), ast.ID("c")))),
				ast.ID("result"),
			},
			ast.Ty("f32"),
			nil,
			nil,
			false,
			false,
		),
		ast.Call("step"),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("typed f32 region mismatch: got=%#v want=%#v", got, want)
	}
	a := normalizeFloat(runtime.FloatF32, 1.1)
	b := normalizeFloat(runtime.FloatF32, 0.6)
	c := normalizeFloat(runtime.FloatF32, 0.2)
	assertFloatValue(t, got, runtime.FloatF32, normalizeFloat(runtime.FloatF32, normalizeFloat(runtime.FloatF32, a*b)+normalizeFloat(runtime.FloatF32, a-c)))
}

func TestBytecodeVM_TypedFloatRegionFallsBackWhenRuntimeSlotBreaksProof(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	plan := bytecodeFloatRegionPlan{
		steps: []bytecodeFloatRegionStep{
			{kind: bytecodeFloatRegionLoadSlot, slot: 0},
			{kind: bytecodeFloatRegionConst, value: 2, floatKind: runtime.FloatF64},
			{kind: bytecodeFloatRegionAdd},
			{kind: bytecodeFloatRegionConst, value: 3, floatKind: runtime.FloatF64},
			{kind: bytecodeFloatRegionMul},
		},
		maxDepth: 2,
	}
	vm.currentProgram = &bytecodeProgram{floatRegions: []bytecodeFloatRegionPlan{plan}}
	vm.slots = []runtime.Value{runtime.NewSmallInt(1, runtime.IntegerI32), runtime.NilValue{}}
	instr := &bytecodeInstruction{op: bytecodeOpStoreSlotFloatRegion, target: 1, argCount: 0, discardResult: true}

	if err := vm.execStoreSlotFloatRegion(instr); err != nil {
		t.Fatalf("typed float region fallback failed: %v", err)
	}
	assertFloatValue(t, vm.slots[1], runtime.FloatF64, 9)
	if vm.stackDepth() != 0 {
		t.Fatalf("discarded fallback stack = %#v, want empty", vm.stack)
	}
}

func TestBytecodeVM_TypedFloatRegionSupportsLoopCarriedTargetAndMixedKinds(t *testing.T) {
	f32 := ast.FloatType(runtime.FloatF32)
	module := ast.Mod([]ast.Statement{
		ast.Fn(
			"step",
			nil,
			[]ast.Statement{
				ast.Assign(ast.ID("value"), ast.FltTyped(1.25, &f32)),
				ast.Assign(ast.ID("scale"), ast.Flt(2)),
				ast.AssignOp(ast.AssignmentAssign, ast.ID("value"), ast.Bin("+", ast.Bin("*", ast.ID("value"), ast.ID("scale")), ast.Flt(0.5))),
				ast.ID("value"),
			},
			ast.Ty("f64"),
			nil,
			nil,
			false,
			false,
		),
		ast.Call("step"),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("loop-carried mixed float region mismatch: got=%#v want=%#v", got, want)
	}
	assertFloatValue(t, got, runtime.FloatF64, 3)
}

func TestBytecodeVM_TypedFloatRegionLeavesSingleOperationsOnExistingPath(t *testing.T) {
	def := ast.Fn(
		"step",
		nil,
		[]ast.Statement{
			ast.Assign(ast.ID("a"), ast.Flt(2)),
			ast.Assign(ast.ID("b"), ast.Flt(3)),
			ast.Assign(ast.ID("result"), ast.Bin("*", ast.ID("a"), ast.ID("b"))),
			ast.ID("result"),
		},
		ast.Ty("f64"),
		nil,
		nil,
		false,
		false,
	)

	program, err := NewBytecode().lowerFunctionDefinitionBytecode(def)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	if bytecodeProgramContainsOpcode(program, bytecodeOpStoreSlotFloatRegion) {
		t.Fatal("single float operation should not use a region")
	}
	if !bytecodeProgramContainsOpcode(program, bytecodeOpStoreSlotFloatBinary) {
		t.Fatal("single float operation should retain the existing binary store")
	}
}

func TestBytecodeFloatSimpleTypeCheckPreservesExactNegatedFloatKind(t *testing.T) {
	f32 := ast.FloatType(runtime.FloatF32)
	ctx := &bytecodeLoweringContext{}
	if got := bytecodeExpressionSimpleTypeCheck(ctx, ast.NewUnaryExpression(ast.UnaryOperatorNegate, ast.FltTyped(1.25, &f32))); got != bytecodeSimpleTypeCheckF32 {
		t.Fatalf("negated f32 simple check = %v, want f32", got)
	}
	if got := bytecodeExpressionSimpleTypeCheck(ctx, ast.NewUnaryExpression(ast.UnaryOperatorNegate, ast.Flt(1.25))); got != bytecodeSimpleTypeCheckF64 {
		t.Fatalf("negated f64 simple check = %v, want f64", got)
	}
}
