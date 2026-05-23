package interpreter

import (
	"math"
	"strings"
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_LoweringEmitsI32StackOpsForFinalLiteralArithmetic(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Bin("-", ast.Bin("+", ast.Int(7), ast.Int(5)), ast.Int(3)),
	}, nil, nil)

	interp := NewBytecode()
	program, err := interp.lowerModuleToBytecode(module)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	for _, op := range []bytecodeOp{
		bytecodeOpConstI32,
		bytecodeOpBinaryI32Add,
		bytecodeOpBinaryI32Sub,
		bytecodeOpBoxI32,
	} {
		if !bytecodeProgramContainsOpcode(program, op) {
			t.Fatalf("expected lowering to emit opcode %d", op)
		}
	}
	if bytecodeProgramContainsOpcode(program, bytecodeOpBinaryIntAdd) ||
		bytecodeProgramContainsOpcode(program, bytecodeOpBinaryIntSub) {
		t.Fatalf("expected raw i32 stack lowering to avoid boxed binary opcodes")
	}
}

func TestBytecodeVM_I32StackLiteralArithmeticParity(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Bin("-", ast.Bin("+", ast.Int(7), ast.Int(5)), ast.Int(3)),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode i32 stack arithmetic mismatch: got=%#v want=%#v", got, want)
	}
	assertIntValue(t, got, runtime.IntegerI32, 9)
}

func TestBytecodeVM_I32StackLiteralArithmeticOverflowParity(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Bin("+", ast.Int(math.MaxInt32), ast.Int(1)),
	}, nil, nil)

	treeErr := evalModuleError(t, New(), module)
	if treeErr == nil || !strings.Contains(treeErr.Error(), "integer overflow") {
		t.Fatalf("expected tree integer overflow, got: %v", treeErr)
	}
	byteErr := runBytecodeModuleError(t, NewBytecode(), module)
	if byteErr == nil || !strings.Contains(byteErr.Error(), "integer overflow") {
		t.Fatalf("expected bytecode integer overflow, got: %v", byteErr)
	}
}

func TestBytecodeVM_LoweringEmitsI32SlotStackOpsForFinalParamArithmetic(t *testing.T) {
	def := ast.Fn(
		"inc",
		[]*ast.FunctionParameter{ast.Param("n", ast.Ty("i32"))},
		[]ast.Statement{
			ast.Bin("+", ast.ID("n"), ast.Int(1)),
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
	if program.frameLayout == nil || !program.frameLayout.hasTypedSlots {
		t.Fatalf("expected typed slot metadata")
	}
	if got := program.frameLayout.slotKinds[0]; got != bytecodeCellKindI32 {
		t.Fatalf("expected param slot kind i32, got %d", got)
	}
	for _, op := range []bytecodeOp{
		bytecodeOpLoadSlotI32,
		bytecodeOpConstI32,
		bytecodeOpBinaryI32Add,
		bytecodeOpBoxI32,
	} {
		if !bytecodeProgramContainsOpcode(program, op) {
			t.Fatalf("expected lowering to emit opcode %d", op)
		}
	}
	if bytecodeProgramContainsOpcode(program, bytecodeOpLoadSlot) ||
		bytecodeProgramContainsOpcode(program, bytecodeOpBinaryIntAdd) {
		t.Fatalf("expected final i32 param arithmetic to avoid boxed load/add opcodes")
	}
}

func TestBytecodeVM_I32SlotStackParamArithmeticParity(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Fn(
			"inc",
			[]*ast.FunctionParameter{ast.Param("n", ast.Ty("i32"))},
			[]ast.Statement{
				ast.Bin("+", ast.ID("n"), ast.Int(1)),
			},
			ast.Ty("i32"),
			nil,
			nil,
			false,
			false,
		),
		ast.Call("inc", ast.Int(41)),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode i32 slot arithmetic mismatch: got=%#v want=%#v", got, want)
	}
	assertIntValue(t, got, runtime.IntegerI32, 42)
}

func TestBytecodeVM_I32SlotStackParamArithmeticOverflowParity(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Fn(
			"inc",
			[]*ast.FunctionParameter{ast.Param("n", ast.Ty("i32"))},
			[]ast.Statement{
				ast.Bin("+", ast.ID("n"), ast.Int(1)),
			},
			ast.Ty("i32"),
			nil,
			nil,
			false,
			false,
		),
		ast.Call("inc", ast.Int(math.MaxInt32)),
	}, nil, nil)

	treeErr := evalModuleError(t, New(), module)
	if treeErr == nil || !strings.Contains(treeErr.Error(), "integer overflow") {
		t.Fatalf("expected tree integer overflow, got: %v", treeErr)
	}
	byteErr := runBytecodeModuleError(t, NewBytecode(), module)
	if byteErr == nil || !strings.Contains(byteErr.Error(), "integer overflow") {
		t.Fatalf("expected bytecode integer overflow, got: %v", byteErr)
	}
}

func TestBytecodeVM_LoweringEmitsI32StoreSlotForTypedLocalLiteralArithmetic(t *testing.T) {
	def := ast.Fn(
		"f",
		nil,
		[]ast.Statement{
			ast.Assign(ast.TypedP(ast.ID("x"), ast.Ty("i32")), ast.Bin("+", ast.Int(4), ast.Int(5))),
			ast.Bin("+", ast.ID("x"), ast.Int(1)),
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
	if program.frameLayout == nil || !program.frameLayout.hasTypedSlots {
		t.Fatalf("expected typed slot metadata")
	}
	storeSlot := -1
	for _, instr := range program.instructions {
		if instr.op == bytecodeOpStoreSlotI32 && instr.name == "x" {
			storeSlot = instr.target
			if !instr.discardResult {
				t.Fatalf("statement-position typed i32 declaration should discard its result")
			}
			break
		}
	}
	if storeSlot < 0 {
		t.Fatalf("expected typed local declaration to emit i32 slot store")
	}
	if got := program.frameLayout.slotKinds[storeSlot]; got != bytecodeCellKindI32 {
		t.Fatalf("expected local slot kind i32, got %d", got)
	}
	if !bytecodeProgramContainsOpcode(program, bytecodeOpLoadSlotI32) {
		t.Fatalf("expected final local arithmetic to emit i32 slot load")
	}
}

func TestBytecodeVM_LoweringKeepsFinalI32StoreSlotResult(t *testing.T) {
	def := ast.Fn(
		"f",
		nil,
		[]ast.Statement{
			ast.Assign(ast.TypedP(ast.ID("x"), ast.Ty("i32")), ast.Int(4)),
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
	for _, instr := range program.instructions {
		if instr.op == bytecodeOpStoreSlotI32 && instr.name == "x" {
			if instr.discardResult {
				t.Fatalf("final typed i32 declaration should keep assignment result")
			}
			return
		}
	}
	t.Fatalf("expected final typed i32 declaration to emit i32 slot store")
}

func TestBytecodeVM_StoreSlotI32DiscardResultStoresRawSlot(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.slots = make([]runtime.Value, 1)
	vm.pushI32(37)
	instr := &bytecodeInstruction{op: bytecodeOpStoreSlotI32, target: 0, discardResult: true}
	if err := vm.execStoreSlotI32(instr); err != nil {
		t.Fatalf("discarded i32 slot store failed: %v", err)
	}
	if vm.ip != 1 {
		t.Fatalf("ip after discarded i32 store = %d, want 1", vm.ip)
	}
	if len(vm.stack) != 0 {
		t.Fatalf("discarded i32 store stack = %#v, want empty", vm.stack)
	}
	raw, ok := vm.slots[0].(bytecodeRawI32SlotValue)
	if !ok || raw != bytecodeRawI32SlotValue(37) {
		t.Fatalf("discarded i32 store slot = %#v, want raw 37", vm.slots[0])
	}
	if err := vm.execLoadSlotOpcode(&bytecodeInstruction{op: bytecodeOpLoadSlot, target: 0}); err != nil {
		t.Fatalf("load discarded raw i32 slot: %v", err)
	}
	if _, ok := vm.stack[0].(bytecodeRawI32SlotValue); ok {
		t.Fatalf("visible load should materialize raw i32 slot, got %#v", vm.stack[0])
	}
	assertIntValue(t, vm.stack[0], runtime.IntegerI32, 37)
}

func TestBytecodeVM_LoweringEmitsI32CompoundAssignForTypedSlot(t *testing.T) {
	def := ast.Fn(
		"f",
		nil,
		[]ast.Statement{
			ast.Assign(ast.TypedP(ast.ID("x"), ast.Ty("i32")), ast.Int(1)),
			ast.AssignOp(ast.AssignmentAdd, ast.ID("x"), ast.Int(2)),
			ast.ID("x"),
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
	found := false
	for _, instr := range program.instructions {
		if instr.op == bytecodeOpCompoundAssignSlotI32 {
			found = true
			if instr.operator != string(ast.AssignmentAdd) {
				t.Fatalf("typed i32 compound operator = %q, want %q", instr.operator, ast.AssignmentAdd)
			}
			if !instr.discardResult {
				t.Fatalf("statement-position typed i32 compound assignment should discard its result")
			}
		}
		if instr.op == bytecodeOpCompoundAssignSlot {
			t.Fatalf("typed i32 compound assignment should avoid generic compound slot opcode")
		}
	}
	if !found {
		t.Fatalf("expected typed i32 compound assignment opcode")
	}
}

func TestBytecodeVM_I32CompoundAssignParity(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Fn(
			"f",
			nil,
			[]ast.Statement{
				ast.Assign(ast.TypedP(ast.ID("x"), ast.Ty("i32")), ast.Int(10)),
				ast.AssignOp(ast.AssignmentSub, ast.ID("x"), ast.Bin("+", ast.Int(3), ast.Int(4))),
				ast.ID("x"),
			},
			ast.Ty("i32"),
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
		t.Fatalf("bytecode i32 compound assignment mismatch: got=%#v want=%#v", got, want)
	}
	assertIntValue(t, got, runtime.IntegerI32, 3)
}

func TestBytecodeVM_I32CompoundAssignOverflowParity(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Fn(
			"f",
			nil,
			[]ast.Statement{
				ast.Assign(ast.TypedP(ast.ID("x"), ast.Ty("i32")), ast.Int(math.MaxInt32)),
				ast.AssignOp(ast.AssignmentAdd, ast.ID("x"), ast.Int(1)),
				ast.ID("x"),
			},
			ast.Ty("i32"),
			nil,
			nil,
			false,
			false,
		),
		ast.Call("f"),
	}, nil, nil)

	treeErr := evalModuleError(t, New(), module)
	if treeErr == nil || !strings.Contains(treeErr.Error(), "integer overflow") {
		t.Fatalf("expected tree integer overflow, got: %v", treeErr)
	}
	byteErr := runBytecodeModuleError(t, NewBytecode(), module)
	if byteErr == nil || !strings.Contains(byteErr.Error(), "integer overflow") {
		t.Fatalf("expected bytecode integer overflow, got: %v", byteErr)
	}
}

func TestBytecodeVM_I32CompoundAssignKeepsRHSFirstFallback(t *testing.T) {
	def := ast.Fn(
		"f",
		nil,
		[]ast.Statement{
			ast.Assign(ast.TypedP(ast.ID("x"), ast.Ty("i32")), ast.Int(1)),
			ast.AssignOp(
				ast.AssignmentAdd,
				ast.ID("x"),
				ast.AssignOp(ast.AssignmentAssign, ast.ID("x"), ast.Int(5)),
			),
			ast.ID("x"),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)
	module := ast.Mod([]ast.Statement{def, ast.Call("f")}, nil, nil)

	interp := NewBytecode()
	program, err := interp.lowerFunctionDefinitionBytecode(def)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	if bytecodeProgramContainsOpcode(program, bytecodeOpCompoundAssignSlotI32) {
		t.Fatalf("RHS with assignment side effects should stay on the generic compound path")
	}
	if !bytecodeProgramContainsOpcode(program, bytecodeOpCompoundAssignSlot) {
		t.Fatalf("expected generic compound assignment fallback")
	}
	want := mustEvalModule(t, New(), module)
	got := runBytecodeModuleWithInterpreter(t, interp, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode i32 compound RHS-first fallback mismatch: got=%#v want=%#v", got, want)
	}
	assertIntValue(t, got, runtime.IntegerI32, 10)
}

func TestBytecodeVM_CompoundAssignSlotI32DiscardResultStoresRawSlot(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.slots = []runtime.Value{bytecodeRawI32SlotValue(10)}
	vm.pushI32(4)
	instr := &bytecodeInstruction{op: bytecodeOpCompoundAssignSlotI32, target: 0, operator: string(ast.AssignmentAdd), discardResult: true}
	if err := vm.execCompoundAssignSlotI32(instr); err != nil {
		t.Fatalf("discarded i32 compound assignment failed: %v", err)
	}
	if vm.ip != 1 {
		t.Fatalf("ip after discarded i32 compound assignment = %d, want 1", vm.ip)
	}
	if len(vm.stack) != 0 {
		t.Fatalf("discarded i32 compound assignment stack = %#v, want empty", vm.stack)
	}
	raw, ok := vm.slots[0].(bytecodeRawI32SlotValue)
	if !ok || raw != bytecodeRawI32SlotValue(14) {
		t.Fatalf("discarded i32 compound assignment slot = %#v, want raw 14", vm.slots[0])
	}
}
