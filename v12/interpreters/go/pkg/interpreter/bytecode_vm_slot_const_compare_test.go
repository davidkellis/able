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

func TestBytecodeVM_LoweringEmitsTypedIntegerSlotConstCompareJump(t *testing.T) {
	u8 := ast.IntegerTypeU8
	def := ast.Fn(
		"classify_byte",
		[]*ast.FunctionParameter{ast.Param("byte", ast.Ty("u8"))},
		[]ast.Statement{
			ast.IfExpr(
				ast.Bin("==", ast.ID("byte"), ast.IntTyped(45, &u8)),
				ast.Block(ast.ID("byte")),
			),
			ast.IfExpr(
				ast.Bin("<", ast.ID("byte"), ast.IntTyped(48, &u8)),
				ast.Block(ast.ID("byte")),
			),
			ast.IfExpr(
				ast.Bin(">=", ast.ID("byte"), ast.IntTyped(57, &u8)),
				ast.Block(ast.ID("byte")),
			),
			ast.ID("byte"),
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
		if instr.op != bytecodeOpJumpIfIntCompareSlotConstFalse {
			continue
		}
		compareJumps++
		if instr.operator != "==" && instr.operator != "<" && instr.operator != ">=" {
			t.Fatalf("unexpected typed compare jump operator %q", instr.operator)
		}
		if !instr.hasIntImmediate || instr.intImmediate.TypeSuffix != runtime.IntegerU8 {
			t.Fatalf("expected typed u8 immediate, got %#v", instr.intImmediate)
		}
		if !instr.hasIntRaw || instr.intImmediateRaw <= 0 {
			t.Fatalf("expected typed compare jump to carry raw immediate, got raw=%v value=%d", instr.hasIntRaw, instr.intImmediateRaw)
		}
	}
	if compareJumps != 3 {
		t.Fatalf("expected three typed conditional compare slot-const jumps, got %d", compareJumps)
	}
	if bytecodeProgramContainsOpcode(program, bytecodeOpBinaryIntCompareSlotConst) {
		t.Fatalf("expected if-position typed compare to skip standalone bool-producing opcode")
	}
}

func TestBytecodeVM_LoweringEmitsConditionalJumpForIntCompareSlotConstConjunctionIf(t *testing.T) {
	u8 := ast.IntegerTypeU8
	cond := ast.Bin(
		"&&",
		ast.Bin(">=", ast.ID("byte"), ast.IntTyped(48, &u8)),
		ast.Bin("<=", ast.ID("byte"), ast.IntTyped(57, &u8)),
	)
	ifExpr := ast.IfExpr(cond, ast.Block(ast.Int(1)))
	ifExpr.ElseBody = ast.Block(ast.Int(0))
	def := ast.Fn(
		"is_digit",
		[]*ast.FunctionParameter{ast.Param("byte", ast.Ty("u8"))},
		[]ast.Statement{ifExpr},
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
	compareJumps := 0
	lessEqualJumps := 0
	for _, instr := range program.instructions {
		switch instr.op {
		case bytecodeOpJumpIfIntCompareSlotConstFalse:
			compareJumps++
			if instr.operator != ">=" {
				t.Fatalf("unexpected lower-bound conjunction jump operator %q", instr.operator)
			}
			if !instr.hasIntImmediate || instr.intImmediate.TypeSuffix != runtime.IntegerU8 {
				t.Fatalf("expected typed u8 immediate, got %#v", instr.intImmediate)
			}
		case bytecodeOpJumpIfIntLessEqualSlotConstFalse:
			lessEqualJumps++
			if instr.operator != "<=" {
				t.Fatalf("unexpected upper-bound conjunction jump operator %q", instr.operator)
			}
			if !instr.hasIntImmediate || instr.intImmediate.TypeSuffix != runtime.IntegerU8 {
				t.Fatalf("expected typed u8 immediate, got %#v", instr.intImmediate)
			}
		}
	}
	if compareJumps != 1 || lessEqualJumps != 1 {
		t.Fatalf("expected one compare jump and one less-equal jump for conjunction, got compare=%d lessEqual=%d", compareJumps, lessEqualJumps)
	}
	if bytecodeProgramContainsOpcode(program, bytecodeOpDup) {
		t.Fatalf("expected conjunction in if-position to skip generic dup-based && lowering")
	}
	if bytecodeProgramContainsOpcode(program, bytecodeOpJumpIfFalse) {
		t.Fatalf("expected conjunction in if-position to skip generic jump-if-false lowering")
	}
	if bytecodeProgramContainsOpcode(program, bytecodeOpBinaryIntCompareSlotConst) {
		t.Fatalf("expected conjunction in if-position to skip standalone bool-producing compare opcodes")
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

func TestBytecodeVM_IntCompareSlotConstConjunctionIfParity(t *testing.T) {
	u8 := ast.IntegerTypeU8
	cond := ast.Bin(
		"&&",
		ast.Bin(">=", ast.ID("byte"), ast.IntTyped(48, &u8)),
		ast.Bin("<=", ast.ID("byte"), ast.IntTyped(57, &u8)),
	)
	ifExpr := ast.IfExpr(cond, ast.Block(ast.Int(1)))
	ifExpr.ElseBody = ast.Block(ast.Int(0))
	module := ast.Mod([]ast.Statement{
		ast.Fn(
			"is_digit",
			[]*ast.FunctionParameter{ast.Param("byte", ast.Ty("u8"))},
			[]ast.Statement{ifExpr},
			ast.Ty("i32"),
			nil,
			nil,
			false,
			false,
		),
		ast.Bin("+", ast.Call("is_digit", ast.IntTyped(52, &u8)), ast.Call("is_digit", ast.IntTyped(65, &u8))),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode digit-range conjunction parity mismatch: got=%#v want=%#v", got, want)
	}
}
