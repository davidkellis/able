package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_LoweringEmitsSlotSlotCompareJump(t *testing.T) {
	def := ast.Fn(
		"guard",
		[]*ast.FunctionParameter{
			ast.Param("lo", ast.Ty("i32")),
			ast.Param("hi", ast.Ty("i32")),
			ast.Param("i", ast.Ty("i32")),
			ast.Param("j", ast.Ty("i32")),
		},
		[]ast.Statement{
			ast.IfExpr(ast.Bin(">=", ast.ID("lo"), ast.ID("hi")), ast.Block(ast.Ret(ast.ID("lo")))),
			ast.IfExpr(ast.Bin(">", ast.ID("i"), ast.ID("j")), ast.Block(ast.Ret(ast.ID("i")))),
			ast.IfExpr(ast.Bin("<=", ast.ID("i"), ast.ID("j")), ast.Block(ast.Ret(ast.ID("j")))),
			ast.ID("lo"),
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
	var compareJumps int
	for _, instr := range program.instructions {
		if instr.op != bytecodeOpJumpIfIntCompareSlotFalse {
			continue
		}
		compareJumps++
		switch instr.operator {
		case ">=", ">", "<=":
		default:
			t.Fatalf("unexpected slot-slot compare operator %q", instr.operator)
		}
		if instr.argCount < 0 || instr.loopBreak < 0 {
			t.Fatalf("slot-slot compare did not carry slot operands: %#v", instr)
		}
	}
	if compareJumps != 3 {
		t.Fatalf("slot-slot compare jump count = %d, want 3", compareJumps)
	}
}

func TestBytecodeVM_JumpIfIntCompareSlotFalseFastPath(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.slots = []runtime.Value{
		runtime.NewSmallInt(10, runtime.IntegerI32),
		runtime.NewSmallInt(9, runtime.IntegerI32),
	}
	instr := &bytecodeInstruction{
		op:        bytecodeOpJumpIfIntCompareSlotFalse,
		argCount:  0,
		loopBreak: 1,
		target:    7,
		operator:  ">=",
	}

	if err := vm.execJumpIfIntCompareSlotFalse(instr); err != nil {
		t.Fatalf("slot-slot compare jump failed: %v", err)
	}
	if vm.ip != 1 {
		t.Fatalf("truthy slot-slot compare should advance ip to 1, got %d", vm.ip)
	}

	vm.ip = 0
	vm.slots[0] = runtime.NewSmallInt(8, runtime.IntegerI32)
	if err := vm.execJumpIfIntCompareSlotFalse(instr); err != nil {
		t.Fatalf("false slot-slot compare jump failed: %v", err)
	}
	if vm.ip != 7 {
		t.Fatalf("false slot-slot compare should jump to 7, got %d", vm.ip)
	}
}
