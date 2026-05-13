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

func TestBytecodeVM_LoweringEmitsArrayIndexSlotCompareSlotJump(t *testing.T) {
	def := ast.Fn(
		"partition_guard",
		[]*ast.FunctionParameter{
			ast.Param("arr", ast.Gen(ast.Ty("Array"), ast.Ty("i32"))),
			ast.Param("i", ast.Ty("i32")),
			ast.Param("pivot", ast.Ty("i32")),
		},
		[]ast.Statement{
			ast.IfExpr(
				ast.Bin(">=", ast.NewTypeCastExpression(ast.Index(ast.ID("arr"), ast.ID("i")), ast.Ty("i32")), ast.ID("pivot")),
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
		if instr.op != bytecodeOpJumpIfArrayIndexSlotCompareSlotFalse {
			continue
		}
		sawFused = true
		if instr.operator != ">=" {
			t.Fatalf("array index compare operator = %q, want >=", instr.operator)
		}
		if instr.argCount != 0 || instr.loopBreak != 1 || instr.loopContinue != 2 {
			t.Fatalf("array index compare slots = receiver %d index %d right %d, want 0/1/2", instr.argCount, instr.loopBreak, instr.loopContinue)
		}
		if got := typeExpressionToString(instr.typeExpr); got != "i32" {
			t.Fatalf("array index compare cast type = %q, want i32", got)
		}
		if instr.name != "i32" {
			t.Fatalf("array index compare cached cast name = %q, want i32", instr.name)
		}
	}
	if !sawFused {
		t.Fatalf("expected lowering to emit array index compare-slot conditional jump")
	}
	if bytecodeProgramContainsOpcode(program, bytecodeOpArrayIndexGetSlot) {
		t.Fatalf("expected fused condition to skip standalone array index get opcode")
	}
	if bytecodeProgramContainsOpcode(program, bytecodeOpCast) {
		t.Fatalf("expected fused condition to absorb standalone cast opcode")
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

func TestBytecodeVM_JumpIfArrayIndexSlotCompareSlotFalseFastPath(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	arr := interp.newArrayValue([]runtime.Value{
		runtime.NewSmallInt(5, runtime.IntegerI32),
	}, 0)
	instr := &bytecodeInstruction{
		op:           bytecodeOpJumpIfArrayIndexSlotCompareSlotFalse,
		argCount:     0,
		loopBreak:    1,
		loopContinue: 2,
		target:       9,
		operator:     ">=",
		name:         "i32",
		typeExpr:     ast.Ty("i32"),
	}
	vm.slots = []runtime.Value{
		arr,
		runtime.NewSmallInt(0, runtime.IntegerI32),
		runtime.NewSmallInt(4, runtime.IntegerI32),
	}

	if err := vm.execJumpIfArrayIndexSlotCompareSlotFalse(instr); err != nil {
		t.Fatalf("array index compare jump failed: %v", err)
	}
	if vm.ip != 1 {
		t.Fatalf("truthy compare should advance ip to 1, got %d", vm.ip)
	}

	vm.ip = 0
	vm.slots[2] = runtime.NewSmallInt(6, runtime.IntegerI32)
	if err := vm.execJumpIfArrayIndexSlotCompareSlotFalse(instr); err != nil {
		t.Fatalf("array index compare false jump failed: %v", err)
	}
	if vm.ip != 9 {
		t.Fatalf("false compare should jump to 9, got %d", vm.ip)
	}

	vm.ip = 0
	vm.slots[1] = runtime.NewSmallInt(-1, runtime.IntegerI32)
	err := vm.execJumpIfArrayIndexSlotCompareSlotFalse(instr)
	if err == nil {
		t.Fatalf("negative index compare should preserve cast/index error")
	}
}

func TestBytecodeVM_ArrayIndexSlotCompareI32RawCastFastPath(t *testing.T) {
	cases := []struct {
		name  string
		value runtime.Value
		want  int64
	}{
		{name: "i32", value: runtime.NewSmallInt(5, runtime.IntegerI32), want: 5},
		{name: "u8", value: runtime.NewSmallInt(255, runtime.IntegerU8), want: 255},
		{name: "u32_wraps", value: runtime.NewSmallInt(2147483648, runtime.IntegerU32), want: -2147483648},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := bytecodeArrayIndexCastSmallI32Raw(tc.value)
			if !ok {
				t.Fatalf("expected raw i32 cast fast path to handle %T", tc.value)
			}
			if got != tc.want {
				t.Fatalf("raw i32 cast = %d, want %d", got, tc.want)
			}
		})
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

func TestBytecodeVM_ArrayReadSlotOpcodeTraceStillRecordsWhenEnabled(t *testing.T) {
	interp := NewBytecode()
	interp.bytecodeTraceEnabled = true
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	arr := interp.newArrayValue([]runtime.Value{
		runtime.StringValue{Val: "zero"},
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
		runtime.NewSmallInt(0, runtime.IntegerI32),
	}

	if _, err := vm.execArrayReadSlot(instr, program); err != nil {
		t.Fatalf("array read_slot opcode failed: %v", err)
	}
	snapshot := interp.BytecodeTrace(0)
	if !snapshot.Enabled || snapshot.TotalHits != 1 {
		t.Fatalf("array read_slot trace snapshot = %#v, want one enabled hit", snapshot)
	}
	entry := snapshot.Entries[0]
	if entry.Op != "call_member" || entry.Name != "read_slot" || entry.Lookup != "resolved_method" || entry.Dispatch != "array_read_slot_tracked_fast" {
		t.Fatalf("array read_slot trace entry = %#v", entry)
	}
}
