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
		case bytecodeOpArrayIndexGetSlot, bytecodeOpArrayIndexSetSlot, bytecodeOpIndexGet, bytecodeOpIndexSet:
			t.Fatalf("swap pattern should avoid standalone index opcode %v", instr.op)
		}
	}
	if !sawSwap {
		t.Fatalf("expected lowering to emit array index swap slot opcode")
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
