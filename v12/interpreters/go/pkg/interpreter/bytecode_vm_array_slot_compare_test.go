package interpreter

import (
	"strings"
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_LoweringEmitsArrayReadSlotCompareSlotJump(t *testing.T) {
	def := ast.Fn(
		"partition_guard",
		[]*ast.FunctionParameter{
			ast.Param("arr", ast.Gen(ast.Ty("Array"), ast.Ty("i32"))),
			ast.Param("i", ast.Ty("i32")),
			ast.Param("pivot", ast.Ty("i32")),
		},
		[]ast.Statement{
			ast.IfExpr(
				ast.Bin(">=", ast.CallExpr(ast.Member(ast.ID("arr"), "read_slot"), ast.ID("i")), ast.ID("pivot")),
				ast.Block(ast.Ret(ast.ID("i"))),
			),
			ast.ID("i"),
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
	var sawFused bool
	for _, instr := range program.instructions {
		if instr.op != bytecodeOpJumpIfArrayReadSlotCompareSlotFalse {
			continue
		}
		sawFused = true
		if instr.operator != ">=" {
			t.Fatalf("array slot compare operator = %q, want >=", instr.operator)
		}
		if instr.argCount != 0 || instr.loopBreak != 1 || instr.loopContinue != 2 {
			t.Fatalf("array slot compare slots = receiver %d index %d right %d, want 0/1/2", instr.argCount, instr.loopBreak, instr.loopContinue)
		}
	}
	if !sawFused {
		t.Fatalf("expected lowering to emit array read_slot compare-slot conditional jump")
	}
	if bytecodeProgramContainsOpcode(program, bytecodeOpCallMemberArraySlot) {
		t.Fatalf("expected fused condition to skip standalone read_slot call opcode")
	}
}

func TestBytecodeVM_LoweringEmitsArrayReadSlotOpcodeForSlotExpression(t *testing.T) {
	def := ast.Fn(
		"load_value",
		[]*ast.FunctionParameter{
			ast.Param("arr", ast.Gen(ast.Ty("Array"), ast.Ty("i32"))),
			ast.Param("i", ast.Ty("i32")),
		},
		[]ast.Statement{
			ast.Assign(ast.ID("value"), ast.CallExpr(ast.Member(ast.ID("arr"), "read_slot"), ast.ID("i"))),
			ast.ID("value"),
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
	var sawRead bool
	for _, instr := range program.instructions {
		if instr.op != bytecodeOpArrayReadSlot {
			continue
		}
		sawRead = true
		if instr.argCount != 0 || instr.loopBreak != 1 {
			t.Fatalf("array read_slot slots = receiver %d index %d, want 0/1", instr.argCount, instr.loopBreak)
		}
	}
	if !sawRead {
		t.Fatalf("expected lowering to emit array read_slot opcode")
	}
	if bytecodeProgramContainsOpcode(program, bytecodeOpCallMemberArraySlot) {
		t.Fatalf("expected direct slot read to skip standalone read_slot call opcode")
	}
}

func TestBytecodeVM_JumpIfArrayReadSlotCompareSlotFalseFastPath(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	arr := interp.newArrayValue([]runtime.Value{
		runtime.NewSmallInt(5, runtime.IntegerI32),
	}, 0)
	program := &bytecodeProgram{}
	vm.storeCachedCanonicalArraySlotCall(program, 0, bytecodeInstruction{name: "read_slot", argCount: 1}, arr, bytecodeMemberMethodFastPathArrayReadSlot)
	instr := &bytecodeInstruction{
		op:           bytecodeOpJumpIfArrayReadSlotCompareSlotFalse,
		argCount:     0,
		loopBreak:    1,
		loopContinue: 2,
		target:       9,
		operator:     ">=",
	}
	vm.slots = []runtime.Value{
		arr,
		runtime.NewSmallInt(0, runtime.IntegerI32),
		runtime.NewSmallInt(4, runtime.IntegerI32),
	}

	if err := vm.execJumpIfArrayReadSlotCompareSlotFalse(instr, program); err != nil {
		t.Fatalf("array read_slot compare jump failed: %v", err)
	}
	if vm.ip != 1 {
		t.Fatalf("truthy compare should advance ip to 1, got %d", vm.ip)
	}

	vm.ip = 0
	vm.slots[2] = runtime.NewSmallInt(6, runtime.IntegerI32)
	if err := vm.execJumpIfArrayReadSlotCompareSlotFalse(instr, program); err != nil {
		t.Fatalf("array read_slot compare false jump failed: %v", err)
	}
	if vm.ip != 9 {
		t.Fatalf("false compare should jump to 9, got %d", vm.ip)
	}

	vm.ip = 0
	vm.slots[1] = runtime.NewSmallInt(-1, runtime.IntegerI32)
	err := vm.execJumpIfArrayReadSlotCompareSlotFalse(instr, program)
	if err == nil || !strings.Contains(err.Error(), "array index must be non-negative") {
		t.Fatalf("negative read_slot compare err = %v, want non-negative index error", err)
	}
}

func TestBytecodeVM_ArrayReadSlotOpcodeFastPath(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	arr := interp.newArrayValue([]runtime.Value{
		runtime.StringValue{Val: "zero"},
		runtime.StringValue{Val: "one"},
	}, 0)
	program := &bytecodeProgram{}
	vm.storeCachedCanonicalArraySlotCall(program, 0, bytecodeInstruction{name: "read_slot", argCount: 1}, arr, bytecodeMemberMethodFastPathArrayReadSlot)
	instr := &bytecodeInstruction{
		op:        bytecodeOpArrayReadSlot,
		argCount:  0,
		loopBreak: 1,
		name:      "read_slot",
	}
	vm.slots = []runtime.Value{
		arr,
		runtime.NewSmallInt(1, runtime.IntegerI32),
	}

	newProg, err := vm.execArrayReadSlot(instr, program)
	if err != nil {
		t.Fatalf("array read_slot opcode failed: %v", err)
	}
	if newProg != nil {
		t.Fatalf("array read_slot opcode should stay in current program")
	}
	if vm.ip != 1 {
		t.Fatalf("array read_slot opcode ip = %d, want 1", vm.ip)
	}
	if want := (runtime.StringValue{Val: "one"}); !valuesEqual(vm.stack[0], want) {
		t.Fatalf("array read_slot opcode result = %#v, want %#v", vm.stack[0], want)
	}

	vm = newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.storeCachedCanonicalArraySlotCall(program, 0, bytecodeInstruction{name: "read_slot", argCount: 1}, arr, bytecodeMemberMethodFastPathArrayReadSlot)
	vm.slots = []runtime.Value{
		arr,
		runtime.NewSmallInt(-1, runtime.IntegerI32),
	}
	_, err = vm.execArrayReadSlot(instr, program)
	if err == nil || !strings.Contains(err.Error(), "array index must be non-negative") {
		t.Fatalf("negative read_slot opcode err = %v, want non-negative index error", err)
	}
}
