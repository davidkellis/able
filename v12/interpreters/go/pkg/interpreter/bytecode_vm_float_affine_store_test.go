package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_LoweringEmitsStoreSlotFloatAffine(t *testing.T) {
	i32 := ast.IntegerTypeI32
	def := ast.Fn(
		"pixel_ci",
		nil,
		[]ast.Statement{
			ast.Assign(ast.ID("y"), ast.IntTyped(17, &i32)),
			ast.Assign(ast.ID("ci"), ast.Bin("-", ast.Bin("/", ast.Bin("*", ast.Flt(2.0), ast.NewTypeCastExpression(ast.ID("y"), ast.Ty("f64"))), ast.NewTypeCastExpression(ast.ID("SIZE"), ast.Ty("f64"))), ast.Flt(1.0))),
			ast.ID("ci"),
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
	if !bytecodeProgramContainsOpcode(program, bytecodeOpStoreSlotFloatAffine) {
		t.Fatalf("expected lowering to emit fused float affine store")
	}
	if len(program.floatAffineStores) != 1 {
		t.Fatalf("expected one float affine plan, got %#v", program.floatAffineStores)
	}
	for _, plan := range program.floatAffineStores {
		if plan.sourceSlot < 0 || plan.divisorName != "SIZE" || plan.targetKind != runtime.FloatF64 {
			t.Fatalf("unexpected float affine plan: %#v", plan)
		}
		if plan.scaleVal != 2.0 || plan.offsetVal != 1.0 {
			t.Fatalf("unexpected float affine constants: %#v", plan)
		}
	}
}

func TestBytecodeVM_StoreSlotFloatAffineParity(t *testing.T) {
	i32 := ast.IntegerTypeI32
	module := ast.Mod([]ast.Statement{
		ast.Assign(ast.ID("SIZE"), ast.IntTyped(800, &i32)),
		ast.Fn(
			"pixel_ci",
			nil,
			[]ast.Statement{
				ast.Assign(ast.ID("y"), ast.IntTyped(17, &i32)),
				ast.Assign(ast.ID("ci"), ast.Bin("-", ast.Bin("/", ast.Bin("*", ast.Flt(2.0), ast.NewTypeCastExpression(ast.ID("y"), ast.Ty("f64"))), ast.NewTypeCastExpression(ast.ID("SIZE"), ast.Ty("f64"))), ast.Flt(1.0))),
				ast.ID("ci"),
			},
			ast.Ty("f64"),
			nil,
			nil,
			false,
			false,
		),
		ast.Call("pixel_ci"),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode float affine store mismatch: got=%#v want=%#v", got, want)
	}
	assertFloatValue(t, got, runtime.FloatF64, -0.9575)
}

func TestBytecodeVM_StoreSlotFloatAffineFastPathUsesI32RegisterLaneAndGlobalDivisor(t *testing.T) {
	interp := NewBytecode()
	env := interp.GlobalEnvironment()
	env.Define("SIZE", runtime.NewSmallInt(800, runtime.IntegerI32))
	vm := newBytecodeVM(interp, env)
	vm.slots = []runtime.Value{nil, nil}
	program := &bytecodeProgram{
		frameLayout: &bytecodeFrameLayout{
			slotCount:        2,
			slotKinds:        []bytecodeCellKind{bytecodeCellKindI32, bytecodeCellKindValue},
			i32RegisterFrame: true,
		},
		floatAffineStores: map[int]bytecodeStoreSlotFloatAffinePlan{
			0: {
				sourceSlot:  0,
				divisorSlot: -1,
				divisorName: "SIZE",
				targetKind:  runtime.FloatF64,
				scaleVal:    2.0,
				scaleKind:   runtime.FloatF64,
				offsetVal:   1.5,
				offsetKind:  runtime.FloatF64,
			},
		},
	}
	vm.currentProgram = program
	vm.activateI32RegisterFrame(program)
	if !vm.setI32RegisterRaw(0, 17) {
		t.Fatalf("expected i32 register lane to accept raw slot value")
	}
	instr := &bytecodeInstruction{
		op:            bytecodeOpStoreSlotFloatAffine,
		target:        1,
		discardResult: true,
	}
	if err := vm.execStoreSlotFloatAffine(instr); err != nil {
		t.Fatalf("unexpected float affine store error: %v", err)
	}
	assertFloatValue(t, vm.slots[1], runtime.FloatF64, -1.4575)
	if _, ok := vm.slots[1].(bytecodeRawF64SlotValue); !ok {
		t.Fatalf("stored value = %#v, want raw f64 slot value", vm.slots[1])
	}
	if len(vm.stack) != 0 {
		t.Fatalf("discarded float affine store should leave stack empty, got %#v", vm.stack)
	}
}
