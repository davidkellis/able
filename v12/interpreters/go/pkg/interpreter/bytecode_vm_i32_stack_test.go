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

func TestBytecodeVM_LoweringEmitsArrayReadSlotI32ForTypedLocal(t *testing.T) {
	def := ast.Fn(
		"load_value",
		[]*ast.FunctionParameter{
			ast.Param("arr", ast.Gen(ast.Ty("Array"), ast.Ty("i32"))),
			ast.Param("i", ast.Ty("i32")),
		},
		[]ast.Statement{
			ast.Assign(ast.TypedP(ast.ID("value"), ast.Ty("i32")), ast.CallExpr(ast.Member(ast.ID("arr"), "read_slot"), ast.ID("i"))),
			ast.ID("value"),
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
	if !bytecodeProgramContainsOpcode(program, bytecodeOpArrayReadSlotI32) {
		t.Fatalf("expected typed array read local to emit array read_slot i32 opcode")
	}
	if bytecodeProgramContainsOpcode(program, bytecodeOpArrayReadSlot) {
		t.Fatalf("typed array read local should avoid generic array read_slot opcode")
	}
	if bytecodeProgramContainsOpcode(program, bytecodeOpUnboxI32) {
		t.Fatalf("typed array read local should avoid unbox i32 opcode")
	}
	if !bytecodeProgramContainsOpcode(program, bytecodeOpStoreSlotI32) {
		t.Fatalf("expected typed array read local to emit i32 slot store")
	}
	for _, instr := range program.instructions {
		if instr.op == bytecodeOpStoreSlotNew && instr.name == "value" {
			t.Fatalf("typed array read local should avoid generic store slot new")
		}
	}
}

func TestBytecodeVM_ArrayReadSlotI32OpcodeFastPath(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	arr := interp.newArrayValue([]runtime.Value{
		runtime.NewSmallInt(5, runtime.IntegerI32),
		runtime.NewSmallInt(9, runtime.IntegerI32),
	}, 2)
	if _, err := interp.ensureArrayState(arr, 0); err != nil {
		t.Fatalf("ensure array state: %v", err)
	}
	program := &bytecodeProgram{frameLayout: &bytecodeFrameLayout{
		slotCount:        2,
		slotKinds:        []bytecodeCellKind{bytecodeCellKindValue, bytecodeCellKindI32},
		hasTypedSlots:    true,
		i32RegisterFrame: true,
	}}
	vm.slots = []runtime.Value{arr, nil}
	vm.activateI32RegisterFrame(program)
	if !vm.setI32RegisterRaw(1, 1) {
		t.Fatalf("expected register frame to accept read_slot i32 index")
	}
	vm.storeCachedCanonicalArraySlotCall(program, 0, bytecodeInstruction{name: "read_slot", argCount: 1}, arr, bytecodeMemberMethodFastPathArrayReadSlot)

	if err := vm.execArrayReadSlotI32(&bytecodeInstruction{
		op:        bytecodeOpArrayReadSlotI32,
		argCount:  0,
		loopBreak: 1,
		name:      "read_slot",
		node:      ast.CallExpr(ast.Member(ast.ID("arr"), "read_slot"), ast.ID("i")),
	}, program); err != nil {
		t.Fatalf("array read_slot i32 opcode failed: %v", err)
	}
	if vm.ip != 1 {
		t.Fatalf("array read_slot i32 opcode ip = %d, want 1", vm.ip)
	}
	raw, err := vm.popI32()
	if err != nil {
		t.Fatalf("pop i32 fast path failed: %v", err)
	}
	if raw != 9 {
		t.Fatalf("array read_slot i32 raw = %d, want 9", raw)
	}
}

func TestBytecodeVM_ArrayReadSlotI32OpcodePreservesTypedMismatch(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	arr := interp.newArrayValue([]runtime.Value{
		runtime.NewSmallInt(5, runtime.IntegerI32),
	}, 1)
	program := &bytecodeProgram{}
	vm.slots = []runtime.Value{
		arr,
		runtime.NewSmallInt(4, runtime.IntegerI32),
		nil,
	}
	vm.storeCachedCanonicalArraySlotCall(program, 0, bytecodeInstruction{name: "read_slot", argCount: 1}, arr, bytecodeMemberMethodFastPathArrayReadSlot)

	if err := vm.execArrayReadSlotI32(&bytecodeInstruction{
		op:        bytecodeOpArrayReadSlotI32,
		argCount:  0,
		loopBreak: 1,
		name:      "read_slot",
		node:      ast.CallExpr(ast.Member(ast.ID("arr"), "read_slot"), ast.ID("i")),
	}, program); err != nil {
		t.Fatalf("array read_slot i32 mismatch setup failed: %v", err)
	}
	if err := vm.execStoreSlotI32(&bytecodeInstruction{op: bytecodeOpStoreSlotI32, target: 2}); err != nil {
		t.Fatalf("array read_slot i32 mismatch store failed: %v", err)
	}
	if vm.slots[2] != nil {
		t.Fatalf("array read_slot i32 mismatch should not store slot value, got %#v", vm.slots[2])
	}
	if len(vm.stack) != 1 {
		t.Fatalf("array read_slot i32 mismatch stack len = %d, want 1", len(vm.stack))
	}
	errVal, ok := vm.stack[0].(runtime.ErrorValue)
	if !ok {
		t.Fatalf("array read_slot i32 mismatch result = %T, want runtime.ErrorValue", vm.stack[0])
	}
	if !strings.Contains(errVal.Message, "Typed pattern mismatch in assignment") {
		t.Fatalf("unexpected array read_slot i32 mismatch message: %q", errVal.Message)
	}
}

func TestBytecodeVM_TypedI32UnboxStoreCoercesRangeCompatibleInteger(t *testing.T) {
	u8 := ast.IntegerTypeU8
	module := ast.Mod([]ast.Statement{
		ast.Fn(
			"f",
			nil,
			[]ast.Statement{
				ast.Assign(ast.TypedP(ast.ID("x"), ast.Ty("i32")), ast.IntTyped(255, &u8)),
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

	interp := NewBytecode()
	fnProgram, err := interp.lowerFunctionDefinitionBytecode(module.Body[0].(*ast.FunctionDefinition))
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	if !bytecodeProgramContainsOpcode(fnProgram, bytecodeOpUnboxI32) {
		t.Fatalf("expected range-compatible integer typed store to emit unbox i32")
	}

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModuleWithInterpreter(t, interp, module)
	if !valuesEqual(got, want) {
		t.Fatalf("typed i32 unbox coercion mismatch: got=%#v want=%#v", got, want)
	}
	assertIntValue(t, got, runtime.IntegerI32, 255)
}

func TestBytecodeVM_UnboxI32MismatchPreservesAssignmentErrorValue(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.slots = make([]runtime.Value, 1)
	vm.stack = append(vm.stack, runtime.StringValue{Val: "oops"})

	if err := vm.execUnboxI32(&bytecodeInstruction{op: bytecodeOpUnboxI32, node: ast.Str("oops")}); err != nil {
		t.Fatalf("unbox i32 mismatch failed: %v", err)
	}
	if err := vm.execStoreSlotI32(&bytecodeInstruction{op: bytecodeOpStoreSlotI32, target: 0}); err != nil {
		t.Fatalf("store slot i32 fallback failed: %v", err)
	}
	if vm.slots[0] != nil {
		t.Fatalf("typed i32 mismatch should not store slot value, got %#v", vm.slots[0])
	}
	if len(vm.stack) != 1 {
		t.Fatalf("typed i32 mismatch stack len = %d, want 1", len(vm.stack))
	}
	errVal, ok := vm.stack[0].(runtime.ErrorValue)
	if !ok {
		t.Fatalf("typed i32 mismatch result = %T, want runtime.ErrorValue", vm.stack[0])
	}
	if !strings.Contains(errVal.Message, "Typed pattern mismatch in assignment") {
		t.Fatalf("unexpected typed i32 mismatch message: %q", errVal.Message)
	}
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
	if got, ok := vm.stack[0].(bytecodeRawI32SlotValue); !ok || got != 37 {
		t.Fatalf("visible load should preserve raw i32 slot carrier, got %#v", vm.stack[0])
	}
}

func TestBytecodeVM_I32RegisterFrameStoresDiscardedSlotOffValueFrame(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	program := &bytecodeProgram{frameLayout: &bytecodeFrameLayout{
		slotCount:        1,
		slotKinds:        []bytecodeCellKind{bytecodeCellKindI32},
		hasTypedSlots:    true,
		i32RegisterFrame: true,
	}}
	vm.slots = make([]runtime.Value, 1)
	vm.activateI32RegisterFrame(program)
	vm.pushI32(37)

	if err := vm.execStoreSlotI32(&bytecodeInstruction{op: bytecodeOpStoreSlotI32, target: 0, discardResult: true}); err != nil {
		t.Fatalf("discarded i32 register store failed: %v", err)
	}
	if vm.slots[0] != nil {
		t.Fatalf("discarded i32 register store should not write runtime slot, got %#v", vm.slots[0])
	}
	if raw, ok := vm.i32RegisterRaw(0); !ok || raw != 37 {
		t.Fatalf("i32 register slot = %d/%v, want 37/true", raw, ok)
	}
	if err := vm.execLoadSlotI32(&bytecodeInstruction{op: bytecodeOpLoadSlotI32, target: 0}); err != nil {
		t.Fatalf("load i32 register slot failed: %v", err)
	}
	raw, err := vm.popI32()
	if err != nil {
		t.Fatalf("pop i32 register load failed: %v", err)
	}
	if raw != 37 {
		t.Fatalf("raw i32 register load = %d, want 37", raw)
	}
	if err := vm.execLoadSlotOpcode(&bytecodeInstruction{op: bytecodeOpLoadSlot, target: 0}); err != nil {
		t.Fatalf("generic load i32 register slot failed: %v", err)
	}
	if got, ok := vm.stack[len(vm.stack)-1].(bytecodeRawI32SlotValue); !ok || got != 37 {
		t.Fatalf("generic load i32 register slot = %#v, want raw i32 37", vm.stack[len(vm.stack)-1])
	}
}

func TestBytecodeVM_TypedI32DiscardStoreSeedsRegisterFrame(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	program := &bytecodeProgram{frameLayout: &bytecodeFrameLayout{
		slotCount:        1,
		slotKinds:        []bytecodeCellKind{bytecodeCellKindI32},
		hasTypedSlots:    true,
		i32RegisterFrame: true,
	}}
	vm.slots = make([]runtime.Value, 1)
	vm.activateI32RegisterFrame(program)
	vm.stack = append(vm.stack, runtime.NewSmallInt(44, runtime.IntegerI32))

	instr := &bytecodeInstruction{
		op:            bytecodeOpStoreSlotNew,
		target:        0,
		storeTyped:    true,
		typeExpr:      ast.Ty("i32"),
		discardResult: true,
	}
	if err := vm.execStoreSlot(instr); err != nil {
		t.Fatalf("discarded typed i32 store failed: %v", err)
	}
	if len(vm.stack) != 0 {
		t.Fatalf("discarded typed i32 store stack = %#v, want empty", vm.stack)
	}
	if vm.slots[0] != nil {
		t.Fatalf("discarded typed i32 store should not retain boxed slot, got %#v", vm.slots[0])
	}
	if raw, ok := vm.i32RegisterRaw(0); !ok || raw != 44 {
		t.Fatalf("discarded typed i32 store raw register = %d/%v, want 44/true", raw, ok)
	}
}

func TestBytecodeVM_I32RegisterFrameLoopParity(t *testing.T) {
	sum := ast.Fn(
		"sum",
		[]*ast.FunctionParameter{ast.Param("n", ast.Ty("i32"))},
		[]ast.Statement{
			ast.Assign(ast.TypedP(ast.ID("total"), ast.Ty("i32")), ast.Int(0)),
			ast.Assign(ast.TypedP(ast.ID("i"), ast.Ty("i32")), ast.Int(0)),
			ast.Loop(
				ast.Iff(ast.Bin(">=", ast.ID("i"), ast.ID("n")), ast.Brk(nil, nil)),
				ast.AssignOp(ast.AssignmentAssign, ast.ID("total"), ast.Bin("+", ast.ID("total"), ast.ID("i"))),
				ast.AssignOp(ast.AssignmentAssign, ast.ID("i"), ast.Bin("+", ast.ID("i"), ast.Int(1))),
			),
			ast.ID("total"),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)
	module := ast.Mod([]ast.Statement{sum, ast.Call("sum", ast.Int(10))}, nil, nil)

	interp := NewBytecode()
	program, err := interp.lowerFunctionDefinitionBytecode(sum)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	if program.frameLayout == nil || !program.frameLayout.i32RegisterFrame {
		t.Fatalf("expected i32 register frame layout")
	}
	want := mustEvalModule(t, New(), module)
	got := runBytecodeModuleWithInterpreter(t, interp, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode i32 register loop mismatch: got=%#v want=%#v", got, want)
	}
	assertIntValue(t, got, runtime.IntegerI32, 45)
}

func TestBytecodeVM_I32RegisterFrameSurvivesInlineCallNameSlotArgs(t *testing.T) {
	id := ast.Fn(
		"id",
		[]*ast.FunctionParameter{ast.Param("x", ast.Ty("i32"))},
		[]ast.Statement{ast.ID("x")},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)
	caller := ast.Fn(
		"caller",
		[]*ast.FunctionParameter{ast.Param("n", ast.Ty("i32"))},
		[]ast.Statement{
			ast.Assign(ast.TypedP(ast.ID("x"), ast.Ty("i32")), ast.Bin("+", ast.ID("n"), ast.Int(1))),
			ast.Assign(ast.TypedP(ast.ID("y"), ast.Ty("i32")), ast.Call("id", ast.ID("x"))),
			ast.AssignOp(ast.AssignmentAssign, ast.ID("x"), ast.Bin("+", ast.ID("x"), ast.ID("y"))),
			ast.ID("x"),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)
	module := ast.Mod([]ast.Statement{id, caller, ast.Call("caller", ast.Int(10))}, nil, nil)

	interp := NewBytecode()
	program, err := interp.lowerFunctionDefinitionBytecode(caller)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	if program.frameLayout == nil || !program.frameLayout.i32RegisterFrame {
		t.Fatalf("expected direct-call caller to keep i32 register frame layout")
	}
	foundSlotArgCall := false
	foundDiscardedTypedStore := false
	for idx, instr := range program.instructions {
		if instr.op == bytecodeOpCallName && instr.name == "id" && instr.slotArgs {
			foundSlotArgCall = true
		}
		if (instr.op == bytecodeOpStoreSlotNew || instr.op == bytecodeOpStoreSlotI32) && instr.name == "y" {
			if instr.op == bytecodeOpStoreSlotNew && !instr.storeTyped {
				t.Fatalf("expected y store to stay typed")
			}
			if !instr.discardResult {
				t.Fatalf("expected statement-position typed i32 call store to discard result")
			}
			if idx+1 < len(program.instructions) && program.instructions[idx+1].op == bytecodeOpPop {
				t.Fatalf("discarded typed i32 call store should not need a following pop")
			}
			foundDiscardedTypedStore = true
		}
	}
	if !foundSlotArgCall {
		t.Fatalf("expected id(x) to lower to call-name slot args")
	}
	if !foundDiscardedTypedStore {
		t.Fatalf("expected y: i32 := id(x) to lower to a discarded typed store")
	}

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModuleWithInterpreter(t, interp, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode i32 register call-boundary mismatch: got=%#v want=%#v", got, want)
	}
	assertIntValue(t, got, runtime.IntegerI32, 22)
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
