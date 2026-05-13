package interpreter

import (
	"math"
	"strings"
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_LoweringEmitsIntegerSlotConstHotOpcodes(t *testing.T) {
	def := ast.Fn(
		"f",
		[]*ast.FunctionParameter{ast.Param("n", ast.Ty("i32"))},
		[]ast.Statement{
			ast.Bin("+", ast.ID("n"), ast.Int(1)),
			ast.Bin("*", ast.ID("n"), ast.Int(10)),
			ast.Bin("<=", ast.ID("n"), ast.Int(2)),
			ast.Bin(">=", ast.ID("n"), ast.Int(3)),
			ast.Bin("-", ast.ID("n"), ast.Int(1)),
			ast.ID("n"),
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
	sawAddSlotConst := bytecodeProgramContainsOpcode(program, bytecodeOpBinaryIntAddSlotConst)
	sawSubSlotConst := bytecodeProgramContainsOpcode(program, bytecodeOpBinaryIntSubSlotConst)
	sawMulSlotConst := bytecodeProgramContainsOpcode(program, bytecodeOpBinaryIntMulSlotConst)
	sawLESlotConst := bytecodeProgramContainsOpcode(program, bytecodeOpBinaryIntLessEqualSlotConst)
	sawCompareSlotConst := bytecodeProgramContainsOpcode(program, bytecodeOpBinaryIntCompareSlotConst)
	if !sawAddSlotConst || !sawSubSlotConst || !sawMulSlotConst || !sawLESlotConst || !sawCompareSlotConst {
		t.Fatalf("expected lowering to emit slot-const opcodes: add=%v sub=%v mul=%v le=%v compare=%v", sawAddSlotConst, sawSubSlotConst, sawMulSlotConst, sawLESlotConst, sawCompareSlotConst)
	}
	for _, instr := range program.instructions {
		switch instr.op {
		case bytecodeOpBinaryIntAddSlotConst, bytecodeOpBinaryIntSubSlotConst, bytecodeOpBinaryIntMulSlotConst, bytecodeOpBinaryIntLessEqualSlotConst, bytecodeOpBinaryIntCompareSlotConst:
			if !instr.hasIntImmediate {
				t.Fatalf("expected slot-const opcode %v to carry typed integer-immediate metadata", instr.op)
			}
			if got, ok := instr.intImmediate.ToInt64(); !ok || got <= 0 {
				t.Fatalf("expected slot-const opcode %v to keep positive integer immediate, got=%v ok=%v", instr.op, got, ok)
			}
		}
	}
}

func TestBytecodeVM_LoweringFusesSlotConstSelfAssignment(t *testing.T) {
	def := ast.Fn(
		"f",
		nil,
		[]ast.Statement{
			ast.Assign(ast.ID("i"), ast.Int(1)),
			ast.AssignOp(ast.AssignmentAssign, ast.ID("i"), ast.Bin("+", ast.ID("i"), ast.Int(2))),
			ast.AssignOp(ast.AssignmentAssign, ast.ID("i"), ast.Bin("*", ast.ID("i"), ast.Int(3))),
			ast.AssignOp(ast.AssignmentAssign, ast.ID("i"), ast.Bin("-", ast.ID("i"), ast.Int(1))),
			ast.ID("i"),
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
	fused := 0
	for _, instr := range program.instructions {
		if instr.op != bytecodeOpStoreSlotBinaryIntSlotConst {
			continue
		}
		fused++
		if instr.name != "i" || (instr.operator != "+" && instr.operator != "-" && instr.operator != "*") {
			t.Fatalf("unexpected fused slot-const assignment: name=%q operator=%q", instr.name, instr.operator)
		}
		if !instr.hasIntImmediate {
			t.Fatalf("expected fused slot-const assignment to carry typed integer-immediate metadata")
		}
	}
	if fused != 3 {
		t.Fatalf("expected three fused slot-const self-assignments, got %d", fused)
	}
}

func TestBytecodeVM_LoweringDiscardsStatementSlotConstSelfAssignmentResult(t *testing.T) {
	def := ast.Fn(
		"f",
		nil,
		[]ast.Statement{
			ast.Assign(ast.ID("i"), ast.Int(1)),
			ast.AssignOp(ast.AssignmentAssign, ast.ID("i"), ast.Bin("+", ast.ID("i"), ast.Int(2))),
			ast.ID("i"),
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
	var sawDiscardStore bool
	for idx, instr := range program.instructions {
		if instr.op != bytecodeOpStoreSlotBinaryIntSlotConst {
			continue
		}
		sawDiscardStore = true
		if !instr.discardResult {
			t.Fatalf("expected statement-position fused self-assignment to discard result")
		}
		if idx+1 < len(program.instructions) && program.instructions[idx+1].op == bytecodeOpPop {
			t.Fatalf("expected statement-position fused self-assignment to skip following Pop")
		}
	}
	if !sawDiscardStore {
		t.Fatalf("expected fused slot-const self-assignment")
	}
}

func TestBytecodeVM_LoweringKeepsNestedSlotConstAssignmentResult(t *testing.T) {
	inner := ast.AssignOp(ast.AssignmentAssign, ast.ID("i"), ast.Bin("+", ast.ID("i"), ast.Int(2)))
	def := ast.Fn(
		"f",
		nil,
		[]ast.Statement{
			ast.Assign(ast.ID("i"), ast.Int(1)),
			ast.Assign(ast.ID("j"), inner),
			ast.ID("j"),
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
	var sawNestedStore bool
	for _, instr := range program.instructions {
		if instr.op != bytecodeOpStoreSlotBinaryIntSlotConst {
			continue
		}
		sawNestedStore = true
		if instr.discardResult {
			t.Fatalf("nested fused self-assignment result should remain available to outer expression")
		}
	}
	if !sawNestedStore {
		t.Fatalf("expected nested fused slot-const self-assignment")
	}
}

func TestBytecodeVM_LoweringKeepsTypedI32SelfAssignmentOnRawStore(t *testing.T) {
	def := ast.Fn(
		"f",
		nil,
		[]ast.Statement{
			ast.Assign(ast.TypedP(ast.ID("i"), ast.Ty("i32")), ast.Int(1)),
			ast.AssignOp(ast.AssignmentAssign, ast.ID("i"), ast.Bin("+", ast.ID("i"), ast.Int(2))),
			ast.ID("i"),
		},
		ast.Ty("i32"),
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
	if bytecodeProgramContainsOpcode(program, bytecodeOpStoreSlotBinaryIntSlotConst) {
		t.Fatalf("typed i32 self-assignment should stay on raw i32 store path")
	}
	if !bytecodeProgramContainsOpcode(program, bytecodeOpStoreSlotI32) {
		t.Fatalf("expected typed i32 self-assignment to emit StoreSlotI32")
	}
}

func TestBytecodeVM_SlotConstSelfAssignmentParity(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Fn(
			"f",
			nil,
			[]ast.Statement{
				ast.Assign(ast.ID("i"), ast.Int(3)),
				ast.AssignOp(ast.AssignmentAssign, ast.ID("i"), ast.Bin("+", ast.ID("i"), ast.Int(2))),
				ast.AssignOp(ast.AssignmentAssign, ast.ID("i"), ast.Bin("*", ast.ID("i"), ast.Int(3))),
				ast.AssignOp(ast.AssignmentAssign, ast.ID("i"), ast.Bin("-", ast.ID("i"), ast.Int(1))),
				ast.ID("i"),
			},
			nil,
			nil,
			nil,
			false,
			false,
		),
		ast.Call("f"),
	}, nil, nil)
	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode slot-const self-assignment mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_StoreSlotBinaryIntSlotConstFastPath(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.slots = []runtime.Value{runtime.NewSmallInt(4, runtime.IntegerI32)}
	instr := &bytecodeInstruction{
		op:              bytecodeOpStoreSlotBinaryIntSlotConst,
		target:          0,
		operator:        "+",
		intImmediate:    runtime.NewSmallInt(2, runtime.IntegerI32),
		hasIntImmediate: true,
	}
	if err := vm.execStoreSlotBinaryIntSlotConst(instr, nil); err != nil {
		t.Fatalf("unexpected store-slot fast-path error: %v", err)
	}
	got, ok := bytecodeDirectSmallI32Value(vm.slots[0])
	if !ok || got != 6 {
		t.Fatalf("stored slot = %#v, want small i32 6", vm.slots[0])
	}
	if len(vm.stack) != 1 {
		t.Fatalf("expected assignment result on stack, got len=%d", len(vm.stack))
	}
	stackGot, ok := bytecodeDirectSmallI32Value(vm.stack[0])
	if !ok || stackGot != 6 {
		t.Fatalf("stack result = %#v, want small i32 6", vm.stack[0])
	}
	if !vm.selfFastSlot0I32Valid || vm.selfFastSlot0I32Raw != 6 {
		t.Fatalf("expected slot0 raw lane to refresh to 6, valid=%v raw=%d", vm.selfFastSlot0I32Valid, vm.selfFastSlot0I32Raw)
	}
}

func TestBytecodeVM_StoreSlotBinaryIntSlotConstDiscardResultFastPath(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.slots = []runtime.Value{runtime.NewSmallInt(4, runtime.IntegerI32)}
	instr := &bytecodeInstruction{
		op:              bytecodeOpStoreSlotBinaryIntSlotConst,
		target:          0,
		operator:        "+",
		intImmediate:    runtime.NewSmallInt(2, runtime.IntegerI32),
		hasIntImmediate: true,
		discardResult:   true,
	}
	if err := vm.execStoreSlotBinaryIntSlotConst(instr, nil); err != nil {
		t.Fatalf("unexpected store-slot fast-path error: %v", err)
	}
	got, ok := bytecodeDirectSmallI32Value(vm.slots[0])
	if !ok || got != 6 {
		t.Fatalf("stored slot = %#v, want small i32 6", vm.slots[0])
	}
	if len(vm.stack) != 0 {
		t.Fatalf("discarded assignment result should not push stack value, got len=%d", len(vm.stack))
	}
}

func TestBytecodeVM_StoreSlotBinaryIntSlotConstMultiplyFastPath(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.slots = []runtime.Value{runtime.NewSmallInt(4, runtime.IntegerI32)}
	instr := &bytecodeInstruction{
		op:              bytecodeOpStoreSlotBinaryIntSlotConst,
		target:          0,
		operator:        "*",
		intImmediate:    runtime.NewSmallInt(3, runtime.IntegerI32),
		hasIntImmediate: true,
	}
	if err := vm.execStoreSlotBinaryIntSlotConst(instr, nil); err != nil {
		t.Fatalf("unexpected multiply store-slot fast-path error: %v", err)
	}
	got, ok := bytecodeDirectSmallI32Value(vm.slots[0])
	if !ok || got != 12 {
		t.Fatalf("stored slot = %#v, want small i32 12", vm.slots[0])
	}
	stackGot, ok := bytecodeDirectSmallI32Value(vm.stack[0])
	if !ok || stackGot != 12 {
		t.Fatalf("stack result = %#v, want small i32 12", vm.stack[0])
	}
	if !vm.selfFastSlot0I32Valid || vm.selfFastSlot0I32Raw != 12 {
		t.Fatalf("expected slot0 raw lane to refresh to 12, valid=%v raw=%d", vm.selfFastSlot0I32Valid, vm.selfFastSlot0I32Raw)
	}
}

func TestBytecodeVM_StoreSlotBinaryIntSlotConstSubtractFastPath(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.slots = []runtime.Value{runtime.NewSmallInt(4, runtime.IntegerI32)}
	instr := &bytecodeInstruction{
		op:              bytecodeOpStoreSlotBinaryIntSlotConst,
		target:          0,
		operator:        "-",
		intImmediate:    runtime.NewSmallInt(3, runtime.IntegerI32),
		hasIntImmediate: true,
	}
	if err := vm.execStoreSlotBinaryIntSlotConst(instr, nil); err != nil {
		t.Fatalf("unexpected subtract store-slot fast-path error: %v", err)
	}
	got, ok := bytecodeDirectSmallI32Value(vm.slots[0])
	if !ok || got != 1 {
		t.Fatalf("stored slot = %#v, want small i32 1", vm.slots[0])
	}
	stackGot, ok := bytecodeDirectSmallI32Value(vm.stack[0])
	if !ok || stackGot != 1 {
		t.Fatalf("stack result = %#v, want small i32 1", vm.stack[0])
	}
	if !vm.selfFastSlot0I32Valid || vm.selfFastSlot0I32Raw != 1 {
		t.Fatalf("expected slot0 raw lane to refresh to 1, valid=%v raw=%d", vm.selfFastSlot0I32Valid, vm.selfFastSlot0I32Raw)
	}
}

func TestBytecodeVM_StoreSlotBinaryIntSlotConstFastPathOverflow(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.slots = []runtime.Value{runtime.NewSmallInt(math.MaxInt32, runtime.IntegerI32)}
	instr := &bytecodeInstruction{
		op:              bytecodeOpStoreSlotBinaryIntSlotConst,
		target:          0,
		operator:        "+",
		intImmediate:    runtime.NewSmallInt(1, runtime.IntegerI32),
		hasIntImmediate: true,
	}
	err := vm.execStoreSlotBinaryIntSlotConst(instr, nil)
	if err == nil || !strings.Contains(err.Error(), "integer overflow") {
		t.Fatalf("expected integer overflow, got %v", err)
	}
	got, ok := bytecodeDirectSmallI32Value(vm.slots[0])
	if !ok || got != math.MaxInt32 {
		t.Fatalf("overflow should leave slot unchanged, got %#v", vm.slots[0])
	}
	if len(vm.stack) != 0 {
		t.Fatalf("overflow should not push assignment result, got len=%d", len(vm.stack))
	}
}

func TestBytecodeVM_StoreSlotBinaryIntSlotConstSubtractFastPathOverflow(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.slots = []runtime.Value{runtime.NewSmallInt(math.MinInt32, runtime.IntegerI32)}
	instr := &bytecodeInstruction{
		op:              bytecodeOpStoreSlotBinaryIntSlotConst,
		target:          0,
		operator:        "-",
		intImmediate:    runtime.NewSmallInt(1, runtime.IntegerI32),
		hasIntImmediate: true,
	}
	err := vm.execStoreSlotBinaryIntSlotConst(instr, nil)
	if err == nil || !strings.Contains(err.Error(), "integer overflow") {
		t.Fatalf("expected integer overflow, got %v", err)
	}
	got, ok := bytecodeDirectSmallI32Value(vm.slots[0])
	if !ok || got != math.MinInt32 {
		t.Fatalf("overflow should leave slot unchanged, got %#v", vm.slots[0])
	}
	if len(vm.stack) != 0 {
		t.Fatalf("overflow should not push assignment result, got len=%d", len(vm.stack))
	}
}

func TestBytecodeVM_StoreSlotBinaryIntSlotConstMultiplyFastPathOverflow(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.slots = []runtime.Value{runtime.NewSmallInt(math.MaxInt32, runtime.IntegerI32)}
	instr := &bytecodeInstruction{
		op:              bytecodeOpStoreSlotBinaryIntSlotConst,
		target:          0,
		operator:        "*",
		intImmediate:    runtime.NewSmallInt(2, runtime.IntegerI32),
		hasIntImmediate: true,
	}
	err := vm.execStoreSlotBinaryIntSlotConst(instr, nil)
	if err == nil || !strings.Contains(err.Error(), "integer overflow") {
		t.Fatalf("expected integer overflow, got %v", err)
	}
	got, ok := bytecodeDirectSmallI32Value(vm.slots[0])
	if !ok || got != math.MaxInt32 {
		t.Fatalf("overflow should leave slot unchanged, got %#v", vm.slots[0])
	}
	if len(vm.stack) != 0 {
		t.Fatalf("overflow should not push assignment result, got len=%d", len(vm.stack))
	}
}
