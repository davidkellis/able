package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_LoweringEmitsBoolSlotJumpForIfCondition(t *testing.T) {
	def := ast.Fn(
		"pick",
		[]*ast.FunctionParameter{ast.Param("flag", ast.Ty("bool"))},
		[]ast.Statement{
			ast.NewIfExpression(
				ast.ID("flag"),
				ast.Block(ast.Int(1)),
				nil,
				ast.Block(ast.Int(2)),
			),
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
	if program.frameLayout == nil || program.frameLayout.slotKinds[0] != bytecodeCellKindBool {
		t.Fatalf("expected bool param slot metadata, layout=%#v", program.frameLayout)
	}
	if !bytecodeProgramContainsOpcode(program, bytecodeOpJumpIfBoolSlotFalse) {
		t.Fatalf("expected lowering to emit bool-slot jump")
	}
	if bytecodeProgramContainsOpcode(program, bytecodeOpLoadSlot) {
		t.Fatalf("expected bool-slot jump to avoid loading the condition slot")
	}
}

func TestBytecodeVM_LoweringEmitsBoolSlotJumpForWhileCondition(t *testing.T) {
	def := ast.Fn(
		"once",
		[]*ast.FunctionParameter{ast.Param("flag", ast.Ty("bool"))},
		[]ast.Statement{
			ast.Assign(ast.TypedP(ast.ID("n"), ast.Ty("i32")), ast.Int(0)),
			ast.While(
				ast.ID("flag"),
				ast.Block(
					ast.AssignOp(ast.AssignmentAssign, ast.ID("flag"), ast.Bool(false)),
					ast.AssignOp(ast.AssignmentAssign, ast.ID("n"), ast.Bin("+", ast.ID("n"), ast.Int(1))),
				),
			),
			ast.ID("n"),
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
	if !bytecodeProgramContainsOpcode(program, bytecodeOpJumpIfBoolSlotFalse) {
		t.Fatalf("expected while lowering to emit bool-slot jump")
	}
}

func TestBytecodeVM_BoolSlotJumpParity(t *testing.T) {
	pick := ast.Fn(
		"pick",
		[]*ast.FunctionParameter{ast.Param("flag", ast.Ty("bool"))},
		[]ast.Statement{
			ast.NewIfExpression(
				ast.ID("flag"),
				ast.Block(ast.Int(1)),
				nil,
				ast.Block(ast.Int(2)),
			),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)
	once := ast.Fn(
		"once",
		[]*ast.FunctionParameter{ast.Param("flag", ast.Ty("bool"))},
		[]ast.Statement{
			ast.Assign(ast.TypedP(ast.ID("n"), ast.Ty("i32")), ast.Int(0)),
			ast.While(
				ast.ID("flag"),
				ast.Block(
					ast.AssignOp(ast.AssignmentAssign, ast.ID("flag"), ast.Bool(false)),
					ast.AssignOp(ast.AssignmentAssign, ast.ID("n"), ast.Bin("+", ast.ID("n"), ast.Int(1))),
				),
			),
			ast.ID("n"),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)
	module := ast.Mod([]ast.Statement{
		pick,
		once,
		ast.Bin("+",
			ast.Bin("+", ast.Call("pick", ast.Bool(true)), ast.Call("pick", ast.Bool(false))),
			ast.Call("once", ast.Bool(true)),
		),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bool-slot jump parity mismatch: got=%#v want=%#v", got, want)
	}
	assertIntValue(t, got, runtime.IntegerI32, 4)
}

func TestBytecodeVM_BoolSlotJumpFastPath(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.slots = []runtime.Value{runtime.BoolValue{Val: false}}
	instr := &bytecodeInstruction{op: bytecodeOpJumpIfBoolSlotFalse, argCount: 0, target: 9}

	if err := vm.execJumpIfBoolSlotFalse(instr); err != nil {
		t.Fatalf("false bool-slot jump failed: %v", err)
	}
	if vm.ip != 9 {
		t.Fatalf("expected false branch to jump to 9, got %d", vm.ip)
	}

	vm.ip = 3
	vm.slots[0] = runtime.BoolValue{Val: true}
	if err := vm.execJumpIfBoolSlotFalse(instr); err != nil {
		t.Fatalf("true bool-slot jump failed: %v", err)
	}
	if vm.ip != 4 {
		t.Fatalf("expected true branch to advance to 4, got %d", vm.ip)
	}
}
