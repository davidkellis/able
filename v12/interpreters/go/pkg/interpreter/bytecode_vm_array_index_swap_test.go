package interpreter

import (
	"strings"
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_LoweringEmitsArrayIndexSwapSlotOpcode(t *testing.T) {
	def := ast.Fn(
		"swap",
		[]*ast.FunctionParameter{
			ast.Param("arr", ast.Gen(ast.Ty("Array"), ast.Ty("i32"))),
			ast.Param("a", ast.Ty("i32")),
			ast.Param("b", ast.Ty("i32")),
		},
		[]ast.Statement{
			ast.Assign(ast.ID("tmp"), ast.NewTypeCastExpression(ast.Index(ast.ID("arr"), ast.ID("a")), ast.Ty("i32"))),
			ast.AssignIndex(ast.ID("arr"), ast.ID("a"), ast.NewTypeCastExpression(ast.Index(ast.ID("arr"), ast.ID("b")), ast.Ty("i32"))),
			ast.AssignIndex(ast.ID("arr"), ast.ID("b"), ast.ID("tmp")),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)
	program, err := NewBytecode().lowerFunctionDefinitionBytecode(def)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	sawSwap := false
	for _, instr := range program.instructions {
		switch instr.op {
		case bytecodeOpArrayIndexSwapSlot:
			sawSwap = true
			if instr.argCount != 0 || instr.loopBreak != 1 || instr.loopContinue != 2 {
				t.Fatalf("swap slots = receiver %d first %d second %d, want 0/1/2", instr.argCount, instr.loopBreak, instr.loopContinue)
			}
			if instr.name != "i32" {
				t.Fatalf("swap cast name = %q, want i32", instr.name)
			}
		case bytecodeOpArrayIndexGetSlot, bytecodeOpArrayIndexSetSlot, bytecodeOpIndexGet, bytecodeOpIndexSet:
			t.Fatalf("swap pattern should avoid standalone index opcode %v", instr.op)
		}
	}
	if !sawSwap {
		t.Fatalf("expected lowering to emit array index swap slot opcode")
	}
}

func TestBytecodeVM_LoweringEmitsArrayIndexSwapSlotOpcodeForTypedTemp(t *testing.T) {
	def := ast.Fn(
		"swap",
		[]*ast.FunctionParameter{
			ast.Param("arr", ast.Gen(ast.Ty("Array"), ast.Ty("i32"))),
			ast.Param("a", ast.Ty("i32")),
			ast.Param("b", ast.Ty("i32")),
		},
		[]ast.Statement{
			ast.Assign(ast.TypedP(ast.ID("tmp"), ast.Ty("i32")), ast.NewTypeCastExpression(ast.Index(ast.ID("arr"), ast.ID("a")), ast.Ty("i32"))),
			ast.AssignIndex(ast.ID("arr"), ast.ID("a"), ast.NewTypeCastExpression(ast.Index(ast.ID("arr"), ast.ID("b")), ast.Ty("i32"))),
			ast.AssignIndex(ast.ID("arr"), ast.ID("b"), ast.ID("tmp")),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)
	program, err := NewBytecode().lowerFunctionDefinitionBytecode(def)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	if !bytecodeProgramContainsOpcode(program, bytecodeOpArrayIndexSwapSlot) {
		t.Fatalf("expected typed temp swap pattern to emit array index swap slot opcode")
	}
}

func TestBytecodeVM_LoweringEmitsArraySlotSwapSlotOpcode(t *testing.T) {
	def := ast.Fn(
		"swap",
		[]*ast.FunctionParameter{
			ast.Param("arr", ast.Gen(ast.Ty("Array"), ast.Ty("i32"))),
			ast.Param("a", ast.Ty("i32")),
			ast.Param("b", ast.Ty("i32")),
		},
		[]ast.Statement{
			ast.Assign(ast.ID("tmp"), ast.CallExpr(ast.Member(ast.ID("arr"), "read_slot"), ast.ID("a"))),
			ast.CallExpr(ast.Member(ast.ID("arr"), "write_slot"), ast.ID("a"), ast.CallExpr(ast.Member(ast.ID("arr"), "read_slot"), ast.ID("b"))),
			ast.CallExpr(ast.Member(ast.ID("arr"), "write_slot"), ast.ID("b"), ast.ID("tmp")),
		},
		ast.Ty("void"),
		nil,
		nil,
		false,
		false,
	)
	program, err := NewBytecode().lowerFunctionDefinitionBytecode(def)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	sawSwap := false
	for _, instr := range program.instructions {
		switch instr.op {
		case bytecodeOpArraySlotSwapSlot:
			sawSwap = true
			if instr.argCount != 0 || instr.loopBreak != 1 || instr.loopContinue != 2 {
				t.Fatalf("slot swap slots = receiver %d first %d second %d, want 0/1/2", instr.argCount, instr.loopBreak, instr.loopContinue)
			}
		case bytecodeOpCallMemberArraySlot:
			t.Fatalf("read_slot/write_slot swap pattern should avoid standalone array-slot call opcode")
		}
	}
	if !sawSwap {
		t.Fatalf("expected lowering to emit array slot swap opcode")
	}
}

func TestBytecodeVM_LoweringEmitsArraySlotSwapSlotOpcodeForTypedTemp(t *testing.T) {
	def := ast.Fn(
		"swap",
		[]*ast.FunctionParameter{
			ast.Param("arr", ast.Gen(ast.Ty("Array"), ast.Ty("i32"))),
			ast.Param("a", ast.Ty("i32")),
			ast.Param("b", ast.Ty("i32")),
		},
		[]ast.Statement{
			ast.Assign(ast.TypedP(ast.ID("tmp"), ast.Ty("i32")), ast.CallExpr(ast.Member(ast.ID("arr"), "read_slot"), ast.ID("a"))),
			ast.CallExpr(ast.Member(ast.ID("arr"), "write_slot"), ast.ID("a"), ast.CallExpr(ast.Member(ast.ID("arr"), "read_slot"), ast.ID("b"))),
			ast.CallExpr(ast.Member(ast.ID("arr"), "write_slot"), ast.ID("b"), ast.ID("tmp")),
		},
		ast.Ty("void"),
		nil,
		nil,
		false,
		false,
	)
	program, err := NewBytecode().lowerFunctionDefinitionBytecode(def)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	if !bytecodeProgramContainsOpcode(program, bytecodeOpArraySlotSwapSlot) {
		t.Fatalf("expected typed temp slot swap pattern to emit array slot swap opcode")
	}
}

func TestBytecodeVM_ArrayIndexSwapSlotFastPath(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	arr := interp.newArrayValue([]runtime.Value{
		runtime.NewSmallInt(1, runtime.IntegerI32),
		runtime.NewSmallInt(2, runtime.IntegerI32),
		runtime.NewSmallInt(3, runtime.IntegerI32),
	}, 3)
	if _, err := interp.ensureArrayState(arr, 0); err != nil {
		t.Fatalf("ensure array state: %v", err)
	}
	vm.slots = []runtime.Value{
		arr,
		runtime.NewSmallInt(0, runtime.IntegerI32),
		runtime.NewSmallInt(2, runtime.IntegerI32),
	}
	instr := &bytecodeInstruction{
		op:           bytecodeOpArrayIndexSwapSlot,
		argCount:     0,
		loopBreak:    1,
		loopContinue: 2,
		typeExpr:     ast.Ty("i32"),
		name:         "i32",
	}

	if err := vm.execArrayIndexSwapSlot(instr); err != nil {
		t.Fatalf("array index swap slot opcode failed: %v", err)
	}
	if vm.ip != 1 {
		t.Fatalf("array index swap slot opcode ip = %d, want 1", vm.ip)
	}
	if got := vm.stack[0].(runtime.IntegerValue).Int64Fast(); got != 1 {
		t.Fatalf("swap result = %d, want original first value 1", got)
	}
	state, err := interp.ensureArrayState(arr, 0)
	if err != nil {
		t.Fatalf("ensure array state after swap: %v", err)
	}
	if got := state.Values[0].(runtime.IntegerValue).Int64Fast(); got != 3 {
		t.Fatalf("slot 0 after swap = %d, want 3", got)
	}
	if got := state.Values[2].(runtime.IntegerValue).Int64Fast(); got != 1 {
		t.Fatalf("slot 2 after swap = %d, want 1", got)
	}
}

func TestBytecodeVM_ArrayIndexSwapSlotUsesI32RegisterIndexes(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	arr := interp.newArrayValue([]runtime.Value{
		runtime.NewSmallInt(1, runtime.IntegerI32),
		runtime.NewSmallInt(2, runtime.IntegerI32),
		runtime.NewSmallInt(3, runtime.IntegerI32),
	}, 3)
	if _, err := interp.ensureArrayState(arr, 0); err != nil {
		t.Fatalf("ensure array state: %v", err)
	}
	program := &bytecodeProgram{frameLayout: &bytecodeFrameLayout{
		slotCount:        3,
		slotKinds:        []bytecodeCellKind{bytecodeCellKindValue, bytecodeCellKindI32, bytecodeCellKindI32},
		hasTypedSlots:    true,
		i32RegisterFrame: true,
	}}
	vm.slots = []runtime.Value{arr, nil, nil}
	vm.activateI32RegisterFrame(program)
	if !vm.setI32RegisterRaw(1, 0) || !vm.setI32RegisterRaw(2, 2) {
		t.Fatalf("expected register frame to accept swap indexes")
	}

	if err := vm.execArrayIndexSwapSlot(&bytecodeInstruction{
		op:           bytecodeOpArrayIndexSwapSlot,
		argCount:     0,
		loopBreak:    1,
		loopContinue: 2,
		typeExpr:     ast.Ty("i32"),
		name:         "i32",
	}); err != nil {
		t.Fatalf("array index swap slot register-index opcode failed: %v", err)
	}
	if got := vm.stack[0].(runtime.IntegerValue).Int64Fast(); got != 1 {
		t.Fatalf("swap result = %d, want original first value 1", got)
	}
}

func TestBytecodeVM_ArrayIndexSwapSlotI32FastPathHandlesRawTrackedI32(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	arr := interp.newArrayValue([]runtime.Value{
		bytecodeRawI32SlotValue(11),
		runtime.NewSmallInt(22, runtime.IntegerI32),
	}, 2)
	if _, err := interp.ensureArrayState(arr, 0); err != nil {
		t.Fatalf("ensure array state: %v", err)
	}
	vm.slots = []runtime.Value{
		arr,
		runtime.NewSmallInt(0, runtime.IntegerI32),
		runtime.NewSmallInt(1, runtime.IntegerI32),
	}
	if err := vm.execArrayIndexSwapSlot(&bytecodeInstruction{
		op:           bytecodeOpArrayIndexSwapSlot,
		argCount:     0,
		loopBreak:    1,
		loopContinue: 2,
		typeExpr:     ast.Ty("i32"),
		name:         "i32",
	}); err != nil {
		t.Fatalf("array index swap slot opcode failed: %v", err)
	}
	if got, ok := vm.stack[0].(runtime.IntegerValue); !ok || got.TypeSuffix != runtime.IntegerI32 || got.Int64Fast() != 11 {
		t.Fatalf("swap result = %#v, want materialized i32 11", vm.stack[0])
	}
	state, err := interp.ensureArrayState(arr, 0)
	if err != nil {
		t.Fatalf("ensure array state after swap: %v", err)
	}
	if got, ok := state.Values[1].(runtime.IntegerValue); !ok || got.TypeSuffix != runtime.IntegerI32 || got.Int64Fast() != 11 {
		t.Fatalf("slot 1 after swap = %#v, want materialized i32 11", state.Values[1])
	}
}

func TestBytecodeVM_ArraySlotSwapSlotFastPathReturnsVoid(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	arr := interp.newArrayValue([]runtime.Value{
		runtime.NewSmallInt(1, runtime.IntegerI32),
		runtime.NewSmallInt(2, runtime.IntegerI32),
		runtime.NewSmallInt(3, runtime.IntegerI32),
	}, 3)
	if _, err := interp.ensureArrayState(arr, 0); err != nil {
		t.Fatalf("ensure array state: %v", err)
	}
	vm.slots = []runtime.Value{
		arr,
		runtime.NewSmallInt(0, runtime.IntegerI32),
		runtime.NewSmallInt(2, runtime.IntegerI32),
	}
	program := &bytecodeProgram{instructions: []bytecodeInstruction{{
		op:           bytecodeOpArraySlotSwapSlot,
		argCount:     0,
		loopBreak:    1,
		loopContinue: 2,
	}}}
	vm.storeCachedCanonicalArraySlotCallForArray(program, 0, arr, bytecodeMemberMethodFastPathArrayReadWriteSlot)

	if err := vm.execArraySlotSwapSlot(&program.instructions[0], program); err != nil {
		t.Fatalf("array slot swap opcode failed: %v", err)
	}
	if vm.ip != 1 {
		t.Fatalf("array slot swap opcode ip = %d, want 1", vm.ip)
	}
	if _, ok := vm.stack[0].(runtime.VoidValue); !ok {
		t.Fatalf("slot swap result = %#v, want void", vm.stack[0])
	}
	state, err := interp.ensureArrayState(arr, 0)
	if err != nil {
		t.Fatalf("ensure array state after swap: %v", err)
	}
	if got := state.Values[0].(runtime.IntegerValue).Int64Fast(); got != 3 {
		t.Fatalf("slot 0 after swap = %d, want 3", got)
	}
	if got := state.Values[2].(runtime.IntegerValue).Int64Fast(); got != 1 {
		t.Fatalf("slot 2 after swap = %d, want 1", got)
	}
}

func TestBytecodeVM_ArraySlotSwapSlotTrackedFastPathMaterializesNil(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	arr := interp.newArrayValue([]runtime.Value{
		nil,
		runtime.NewSmallInt(9, runtime.IntegerI32),
	}, 2)
	state, err := interp.ensureArrayState(arr, 0)
	if err != nil {
		t.Fatalf("ensure array state: %v", err)
	}
	state.Values[0] = nil
	handled, err := vm.resolveArraySlotSwapSlotFast(arr, runtime.NewSmallInt(0, runtime.IntegerI32), runtime.NewSmallInt(1, runtime.IntegerI32))
	if err != nil {
		t.Fatalf("array slot swap fast path failed: %v", err)
	}
	if !handled {
		t.Fatalf("expected tracked slot swap fast path to handle in-bounds small indexes")
	}
	if got := state.Values[0].(runtime.IntegerValue).Int64Fast(); got != 9 {
		t.Fatalf("slot 0 after swap = %d, want 9", got)
	}
	if _, ok := state.Values[1].(runtime.NilValue); !ok {
		t.Fatalf("slot 1 after swap = %#v, want materialized nil", state.Values[1])
	}
}

func TestBytecodeVM_ArraySlotSwapSlotUsesI32RegisterIndexes(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	arr := interp.newArrayValue([]runtime.Value{
		runtime.NewSmallInt(1, runtime.IntegerI32),
		runtime.NewSmallInt(2, runtime.IntegerI32),
		runtime.NewSmallInt(3, runtime.IntegerI32),
	}, 3)
	if _, err := interp.ensureArrayState(arr, 0); err != nil {
		t.Fatalf("ensure array state: %v", err)
	}
	program := &bytecodeProgram{
		instructions: []bytecodeInstruction{{
			op:           bytecodeOpArraySlotSwapSlot,
			argCount:     0,
			loopBreak:    1,
			loopContinue: 2,
		}},
		frameLayout: &bytecodeFrameLayout{
			slotCount:        3,
			slotKinds:        []bytecodeCellKind{bytecodeCellKindValue, bytecodeCellKindI32, bytecodeCellKindI32},
			hasTypedSlots:    true,
			i32RegisterFrame: true,
		},
	}
	vm.storeCachedCanonicalArraySlotCallForArray(program, 0, arr, bytecodeMemberMethodFastPathArrayReadWriteSlot)
	vm.slots = []runtime.Value{arr, nil, nil}
	vm.activateI32RegisterFrame(program)
	if !vm.setI32RegisterRaw(1, 0) || !vm.setI32RegisterRaw(2, 2) {
		t.Fatalf("expected register frame to accept swap indexes")
	}

	if err := vm.execArraySlotSwapSlot(&program.instructions[0], program); err != nil {
		t.Fatalf("array slot swap register-index opcode failed: %v", err)
	}
	if _, ok := vm.stack[0].(runtime.VoidValue); !ok {
		t.Fatalf("slot swap result = %#v, want void", vm.stack[0])
	}
	state, err := interp.ensureArrayState(arr, 0)
	if err != nil {
		t.Fatalf("ensure array state after swap: %v", err)
	}
	if got := state.Values[0].(runtime.IntegerValue).Int64Fast(); got != 3 {
		t.Fatalf("slot 0 after swap = %d, want 3", got)
	}
	if got := state.Values[2].(runtime.IntegerValue).Int64Fast(); got != 1 {
		t.Fatalf("slot 2 after swap = %d, want 1", got)
	}
}

func TestBytecodeVM_ArrayIndexSwapSlotSyncsSharedAliases(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	first := interp.newArrayValue([]runtime.Value{
		runtime.NewSmallInt(1, runtime.IntegerI32),
		runtime.NewSmallInt(2, runtime.IntegerI32),
	}, 2)
	if _, err := interp.ensureArrayState(first, 0); err != nil {
		t.Fatalf("ensure first array state: %v", err)
	}
	second, err := interp.arrayValueFromHandle(first.Handle, 0, 0)
	if err != nil {
		t.Fatalf("arrayValueFromHandle: %v", err)
	}
	if !first.TrackedAliases || !second.TrackedAliases {
		t.Fatalf("expected both aliases to be marked shared before swap")
	}
	vm.slots = []runtime.Value{
		first,
		runtime.NewSmallInt(0, runtime.IntegerI32),
		runtime.NewSmallInt(1, runtime.IntegerI32),
	}

	if err := vm.execArrayIndexSwapSlot(&bytecodeInstruction{
		op:           bytecodeOpArrayIndexSwapSlot,
		argCount:     0,
		loopBreak:    1,
		loopContinue: 2,
		typeExpr:     ast.Ty("i32"),
	}); err != nil {
		t.Fatalf("array index swap slot opcode failed: %v", err)
	}
	if got := second.Elements[0].(runtime.IntegerValue).Int64Fast(); got != 2 {
		t.Fatalf("shared alias slot 0 after swap = %d, want 2", got)
	}
	if got := second.Elements[1].(runtime.IntegerValue).Int64Fast(); got != 1 {
		t.Fatalf("shared alias slot 1 after swap = %d, want 1", got)
	}
}

func TestBytecodeVM_ArrayIndexSwapSlotPreservesCastError(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	arr := interp.newArrayValue([]runtime.Value{
		runtime.NewSmallInt(1, runtime.IntegerI32),
	}, 1)
	vm.slots = []runtime.Value{
		arr,
		runtime.NewSmallInt(-1, runtime.IntegerI32),
		runtime.NewSmallInt(0, runtime.IntegerI32),
	}
	err := vm.execArrayIndexSwapSlot(&bytecodeInstruction{
		op:           bytecodeOpArrayIndexSwapSlot,
		argCount:     0,
		loopBreak:    1,
		loopContinue: 2,
		typeExpr:     ast.Ty("i32"),
	})
	if err == nil || !strings.Contains(err.Error(), "cannot cast Error to i32") {
		t.Fatalf("negative index swap error = %v, want cast error", err)
	}
	state, stateErr := interp.ensureArrayState(arr, 0)
	if stateErr != nil {
		t.Fatalf("ensure array state after failed swap: %v", stateErr)
	}
	if got := state.Values[0].(runtime.IntegerValue).Int64Fast(); got != 1 {
		t.Fatalf("slot 0 after failed swap = %d, want unchanged 1", got)
	}
}
