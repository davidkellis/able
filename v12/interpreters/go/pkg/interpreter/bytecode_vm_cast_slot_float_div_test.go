package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_LoweringEmitsBinaryCastSlotFloatConstDivOpcode(t *testing.T) {
	i64 := ast.IntegerTypeI64
	def := ast.Fn(
		"f",
		nil,
		[]ast.Statement{
			ast.Assign(ast.ID("state"), ast.IntTyped(42, &i64)),
			ast.Bin("/", ast.NewTypeCastExpression(ast.ID("state"), ast.Ty("f64")), ast.Flt(2147483647.0)),
		},
		nil,
		nil,
		nil,
		false,
		false,
	)

	interp := NewBytecode()
	program, err := interp.lowerFunctionDefinitionBytecode(def)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	if !bytecodeProgramContainsOpcode(program, bytecodeOpBinaryCastSlotFloatConstDiv) {
		t.Fatalf("expected lowering to emit cast-slot-float-const div opcode")
	}
}

func TestBytecodeVM_LoweringSkipsMirroredBinaryCastSlotFloatConstDivOpcode(t *testing.T) {
	i64 := ast.IntegerTypeI64
	def := ast.Fn(
		"f",
		nil,
		[]ast.Statement{
			ast.Assign(ast.ID("state"), ast.IntTyped(42, &i64)),
			ast.Bin("/", ast.Flt(1.0), ast.NewTypeCastExpression(ast.ID("state"), ast.Ty("f64"))),
		},
		nil,
		nil,
		nil,
		false,
		false,
	)

	interp := NewBytecode()
	program, err := interp.lowerFunctionDefinitionBytecode(def)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	if bytecodeProgramContainsOpcode(program, bytecodeOpBinaryCastSlotFloatConstDiv) {
		t.Fatalf("expected mirrored const/cast division to stay on generic lowering")
	}
}

func TestBytecodeVM_BinaryCastSlotFloatConstDivParity(t *testing.T) {
	i64 := ast.IntegerTypeI64
	module := ast.Mod([]ast.Statement{
		ast.Assign(ast.ID("state"), ast.IntTyped(42, &i64)),
		ast.Bin("/", ast.NewTypeCastExpression(ast.ID("state"), ast.Ty("f64")), ast.Flt(2147483647.0)),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode cast-slot-float-const div mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_BinaryCastSlotFloatConstDivFastPathUsesI32RegisterLane(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.slots = []runtime.Value{nil}
	program := &bytecodeProgram{
		frameLayout: &bytecodeFrameLayout{
			slotCount:        1,
			slotKinds:        []bytecodeCellKind{bytecodeCellKindI32},
			i32RegisterFrame: true,
		},
	}
	vm.activateI32RegisterFrame(program)
	if !vm.setI32RegisterRaw(0, 42) {
		t.Fatalf("expected i32 register lane to accept raw slot value")
	}
	instr := &bytecodeInstruction{
		op:       bytecodeOpBinaryCastSlotFloatConstDiv,
		target:   0,
		value:    runtime.FloatValue{Val: 2, TypeSuffix: runtime.FloatF64},
		typeExpr: ast.Ty("f64"),
	}
	result, handled, err := vm.execBinaryCastSlotFloatConstDiv(instr)
	if err != nil {
		t.Fatalf("unexpected cast-slot-float-const div error: %v", err)
	}
	if !handled {
		t.Fatalf("expected cast-slot-float-const div opcode to handle register-backed slot")
	}
	assertFloatValue(t, result, runtime.FloatF64, 21)
	if _, ok := result.(bytecodeRawF64SlotValue); !ok {
		t.Fatalf("result = %#v, want raw f64 slot value", result)
	}
}

func TestBytecodeVM_LoweringEmitsStoreSlotCastSlotFloatConstDivOpcode(t *testing.T) {
	i64 := ast.IntegerTypeI64
	def := ast.Fn(
		"f",
		nil,
		[]ast.Statement{
			ast.Assign(ast.ID("state"), ast.IntTyped(42, &i64)),
			ast.Assign(ast.ID("x"), ast.Bin("/", ast.NewTypeCastExpression(ast.ID("state"), ast.Ty("f64")), ast.Flt(2147483647.0))),
			ast.ID("x"),
		},
		nil,
		nil,
		nil,
		false,
		false,
	)

	interp := NewBytecode()
	program, err := interp.lowerFunctionDefinitionBytecode(def)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	if !bytecodeProgramContainsOpcode(program, bytecodeOpStoreSlotCastSlotFloatConstDiv) {
		t.Fatalf("expected lowering to emit store cast-slot-float-const div opcode")
	}
}

func TestBytecodeVM_LoweringSkipsMirroredStoreSlotCastSlotFloatConstDivOpcode(t *testing.T) {
	i64 := ast.IntegerTypeI64
	def := ast.Fn(
		"f",
		nil,
		[]ast.Statement{
			ast.Assign(ast.ID("state"), ast.IntTyped(42, &i64)),
			ast.Assign(ast.ID("x"), ast.Bin("/", ast.Flt(1.0), ast.NewTypeCastExpression(ast.ID("state"), ast.Ty("f64")))),
			ast.ID("x"),
		},
		nil,
		nil,
		nil,
		false,
		false,
	)

	interp := NewBytecode()
	program, err := interp.lowerFunctionDefinitionBytecode(def)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	if bytecodeProgramContainsOpcode(program, bytecodeOpStoreSlotCastSlotFloatConstDiv) {
		t.Fatalf("expected mirrored const/cast store division to stay on generic lowering")
	}
}

func TestBytecodeVM_StoreSlotCastSlotFloatConstDivParity(t *testing.T) {
	i64 := ast.IntegerTypeI64
	module := ast.Mod([]ast.Statement{
		ast.Assign(ast.ID("state"), ast.IntTyped(42, &i64)),
		ast.Assign(ast.ID("x"), ast.Bin("/", ast.NewTypeCastExpression(ast.ID("state"), ast.Ty("f64")), ast.Flt(2147483647.0))),
		ast.ID("x"),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode store cast-slot-float-const div mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_FunctionParamMirroredFloatDivParity(t *testing.T) {
	module := mustParseModuleSource(t, `
fn scale(n: i32) -> f64 {
  1.0 / (n as f64) / (n as f64)
}

scale(6)
`)

	want := mustEvalModule(t, New(), module)
	byteInterp := NewBytecode()
	got := runBytecodeModuleWithInterpreter(t, byteInterp, module)

	if !valuesEqual(got, want) {
		t.Fatalf("bytecode mirrored float div mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_StoreSlotCastSlotFloatConstDivFastPathUsesI32RegisterLane(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.slots = []runtime.Value{nil, nil}
	program := &bytecodeProgram{
		frameLayout: &bytecodeFrameLayout{
			slotCount:        2,
			slotKinds:        []bytecodeCellKind{bytecodeCellKindI32, bytecodeCellKindValue},
			i32RegisterFrame: true,
		},
	}
	vm.activateI32RegisterFrame(program)
	if !vm.setI32RegisterRaw(0, 42) {
		t.Fatalf("expected i32 register lane to accept raw slot value")
	}
	instr := &bytecodeInstruction{
		op:            bytecodeOpStoreSlotCastSlotFloatConstDiv,
		target:        1,
		argCount:      0,
		value:         runtime.FloatValue{Val: 2, TypeSuffix: runtime.FloatF64},
		typeExpr:      ast.Ty("f64"),
		discardResult: true,
	}
	if err := vm.execStoreSlotCastSlotFloatConstDiv(instr); err != nil {
		t.Fatalf("unexpected store cast-slot-float-const div error: %v", err)
	}
	assertFloatValue(t, vm.slots[1], runtime.FloatF64, 21)
	if _, ok := vm.slots[1].(bytecodeRawF64SlotValue); !ok {
		t.Fatalf("stored value = %#v, want raw f64 slot value", vm.slots[1])
	}
	if len(vm.stack) != 0 {
		t.Fatalf("discarded store should leave stack empty, got %#v", vm.stack)
	}
}

func TestBytecodeVM_StoreSlotCastSlotFloatConstDivDiscardFastPathStoresRawFloatWithoutOwnedCell(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.slots = []runtime.Value{nil, nil}
	program := &bytecodeProgram{
		frameLayout: &bytecodeFrameLayout{
			slotCount:        2,
			slotKinds:        []bytecodeCellKind{bytecodeCellKindI32, bytecodeCellKindValue},
			i32RegisterFrame: true,
		},
	}
	vm.activateI32RegisterFrame(program)
	if !vm.setI32RegisterRaw(0, 42) {
		t.Fatalf("expected i32 register lane to accept raw slot value")
	}
	instr := &bytecodeInstruction{
		op:            bytecodeOpStoreSlotCastSlotFloatConstDiv,
		target:        1,
		argCount:      0,
		value:         runtime.FloatValue{Val: 2, TypeSuffix: runtime.FloatF64},
		typeExpr:      ast.Ty("f64"),
		discardResult: true,
	}
	if err := vm.execStoreSlotCastSlotFloatConstDiv(instr); err != nil {
		t.Fatalf("unexpected first store error: %v", err)
	}
	assertFloatValue(t, vm.slots[1], runtime.FloatF64, 21)
	if _, ok := vm.slots[1].(bytecodeRawF64SlotValue); !ok {
		t.Fatalf("first stored value = %#v, want raw f64 slot value", vm.slots[1])
	}
	if !vm.setI32RegisterRaw(0, 84) {
		t.Fatalf("expected i32 register lane to accept second raw slot value")
	}
	if err := vm.execStoreSlotCastSlotFloatConstDiv(instr); err != nil {
		t.Fatalf("unexpected second store error: %v", err)
	}
	assertFloatValue(t, vm.slots[1], runtime.FloatF64, 42)
	if len(vm.stack) != 0 {
		t.Fatalf("discarded store should leave stack empty, got %#v", vm.stack)
	}
	if vm.ownedFloatSlots != nil {
		t.Fatalf("expected raw discarded stores to avoid allocating float slot map")
	}
}

func TestBytecodeVM_StoreSlotCastSlotFloatConstDivDiscardFastPathUpdatesExistingOwnedSlotCellWithoutMap(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	existing := &runtime.FloatValue{Val: 7, TypeSuffix: runtime.FloatF64}
	vm.slots = []runtime.Value{nil, existing}
	program := &bytecodeProgram{
		frameLayout: &bytecodeFrameLayout{
			slotCount:        2,
			slotKinds:        []bytecodeCellKind{bytecodeCellKindI32, bytecodeCellKindValue},
			i32RegisterFrame: true,
		},
	}
	vm.activateI32RegisterFrame(program)
	if !vm.setI32RegisterRaw(0, 84) {
		t.Fatalf("expected i32 register lane to accept raw slot value")
	}
	instr := &bytecodeInstruction{
		op:            bytecodeOpStoreSlotCastSlotFloatConstDiv,
		target:        1,
		argCount:      0,
		value:         runtime.FloatValue{Val: 2, TypeSuffix: runtime.FloatF64},
		typeExpr:      ast.Ty("f64"),
		discardResult: true,
	}
	if err := vm.execStoreSlotCastSlotFloatConstDiv(instr); err != nil {
		t.Fatalf("unexpected store error: %v", err)
	}
	got, ok := vm.slots[1].(*runtime.FloatValue)
	if !ok || got == nil {
		t.Fatalf("stored cell = %#v, want owned float slot cell", vm.slots[1])
	}
	if got != existing {
		t.Fatalf("expected discarded store to update existing owned float slot cell in place")
	}
	if got.Val != 42 || got.TypeSuffix != runtime.FloatF64 {
		t.Fatalf("updated cell = %#v, want f64 42", got)
	}
	if vm.ownedFloatSlots != nil {
		t.Fatalf("expected direct in-place owned slot update to avoid allocating float slot map")
	}
}

func TestBytecodeVM_StoreSlotCastSlotFloatConstDivDiscardFastPathUsesRawI64SourceAndExistingOwnedTarget(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	existing := &runtime.FloatValue{Val: 7, TypeSuffix: runtime.FloatF64}
	vm.slots = []runtime.Value{&bytecodeRawI64SlotCell{Val: 84}, existing}
	instr := &bytecodeInstruction{
		op:            bytecodeOpStoreSlotCastSlotFloatConstDiv,
		target:        1,
		argCount:      0,
		value:         runtime.FloatValue{Val: 2, TypeSuffix: runtime.FloatF64},
		typeExpr:      ast.Ty("f64"),
		discardResult: true,
	}
	if err := vm.execStoreSlotCastSlotFloatConstDiv(instr); err != nil {
		t.Fatalf("unexpected raw-i64 source store error: %v", err)
	}
	got, ok := vm.slots[1].(*runtime.FloatValue)
	if !ok || got == nil {
		t.Fatalf("stored cell = %#v, want owned float slot cell", vm.slots[1])
	}
	if got != existing {
		t.Fatalf("expected discarded store to update existing owned float slot cell in place")
	}
	if got.Val != 42 || got.TypeSuffix != runtime.FloatF64 {
		t.Fatalf("updated cell = %#v, want f64 42", got)
	}
	if vm.ownedFloatSlots != nil {
		t.Fatalf("expected raw-i64 direct in-place update to avoid allocating float slot map")
	}
}
