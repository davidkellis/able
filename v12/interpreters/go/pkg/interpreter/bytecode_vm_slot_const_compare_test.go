package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_LoweringEmitsConditionalJumpForIntCompareSlotConstIf(t *testing.T) {
	def := ast.Fn(
		"loop_guard",
		[]*ast.FunctionParameter{ast.Param("n", ast.Ty("i32"))},
		[]ast.Statement{
			ast.IfExpr(
				ast.Bin(">=", ast.ID("n"), ast.Int(9)),
				ast.Block(ast.Bin("-", ast.ID("n"), ast.Int(1))),
			),
			ast.IfExpr(
				ast.Bin(">", ast.ID("n"), ast.Int(12)),
				ast.Block(ast.Bin("-", ast.ID("n"), ast.Int(2))),
			),
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
	compareJumps := 0
	for _, instr := range program.instructions {
		if instr.op == bytecodeOpJumpIfIntCompareSlotConstFalse {
			compareJumps++
			if instr.operator != ">=" && instr.operator != ">" {
				t.Fatalf("unexpected compare jump operator %q", instr.operator)
			}
			if !instr.hasIntRaw || instr.intImmediateRaw <= 0 {
				t.Fatalf("expected compare jump to carry raw immediate, got raw=%v value=%d", instr.hasIntRaw, instr.intImmediateRaw)
			}
		}
	}
	if compareJumps != 2 {
		t.Fatalf("expected two conditional compare slot-const jumps, got %d", compareJumps)
	}
	if bytecodeProgramContainsOpcode(program, bytecodeOpBinaryIntCompareSlotConst) {
		t.Fatalf("expected if-position compare to skip standalone bool-producing opcode")
	}
}

func TestBytecodeVM_JumpIfIntCompareSlotConstFalseFastPath(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.slots = []runtime.Value{runtime.NewSmallInt(10, runtime.IntegerI32)}
	instr := &bytecodeInstruction{
		op:              bytecodeOpJumpIfIntCompareSlotConstFalse,
		argCount:        0,
		target:          7,
		operator:        ">=",
		intImmediate:    runtime.NewSmallInt(9, runtime.IntegerI32),
		intImmediateRaw: 9,
		hasIntImmediate: true,
		hasIntRaw:       true,
	}

	if err := vm.execJumpIfIntCompareSlotConstFalse(instr, nil); err != nil {
		t.Fatalf("jump-if compare fast path failed: %v", err)
	}
	if vm.ip != 1 {
		t.Fatalf("truthy compare should advance ip to 1, got %d", vm.ip)
	}

	vm.ip = 0
	vm.slots[0] = runtime.NewSmallInt(8, runtime.IntegerI32)
	if err := vm.execJumpIfIntCompareSlotConstFalse(instr, nil); err != nil {
		t.Fatalf("jump-if compare false path failed: %v", err)
	}
	if vm.ip != 7 {
		t.Fatalf("false compare should jump to 7, got %d", vm.ip)
	}
}
