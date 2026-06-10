package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_LoweringEmitsReturnConstIfForIntCompareSlotConstStatement(t *testing.T) {
	def := ast.Fn(
		"fib",
		[]*ast.FunctionParameter{ast.Param("n", ast.Ty("i32"))},
		[]ast.Statement{
			ast.IfExpr(
				ast.Bin("<", ast.ID("n"), ast.Int(3)),
				ast.Block(ast.Ret(ast.Int(1))),
			),
			ast.Bin(
				"+",
				ast.Call("fib", ast.Bin("-", ast.ID("n"), ast.Int(1))),
				ast.Call("fib", ast.Bin("-", ast.ID("n"), ast.Int(2))),
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
	found := false
	for _, instr := range program.instructions {
		if instr.op == bytecodeOpReturnConstIfIntLessEqualSlotConst && instr.operator == "<" && instr.hasIntRaw && instr.intImmediateRaw == 3 {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected fused return-const-if opcode to preserve < operator and raw immediate 3")
	}
	if bytecodeProgramContainsOpcode(program, bytecodeOpJumpIfIntCompareSlotConstFalse) {
		t.Fatalf("expected fused return-const-if compare shape to skip standalone compare jump opcode")
	}
}

func TestBytecodeVM_LoweringEmitsReturnConstIfForIntCompareConstSlotStatement(t *testing.T) {
	def := ast.Fn(
		"fib",
		[]*ast.FunctionParameter{ast.Param("n", ast.Ty("i32"))},
		[]ast.Statement{
			ast.IfExpr(
				ast.Bin(">", ast.Int(3), ast.ID("n")),
				ast.Block(ast.Ret(ast.Int(1))),
			),
			ast.Bin(
				"+",
				ast.Call("fib", ast.Bin("-", ast.ID("n"), ast.Int(1))),
				ast.Call("fib", ast.Bin("-", ast.ID("n"), ast.Int(2))),
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
	found := false
	for _, instr := range program.instructions {
		if instr.op == bytecodeOpReturnConstIfIntLessEqualSlotConst && instr.operator == "<" && instr.hasIntRaw && instr.intImmediateRaw == 3 {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected reversed compare to normalize into fused return-const-if slot-const opcode")
	}
	if bytecodeProgramContainsOpcode(program, bytecodeOpJumpIfIntCompareSlotConstFalse) {
		t.Fatalf("expected reversed compare return-if shape to skip standalone compare jump opcode")
	}
}

func TestBytecodeVM_ReturnConstIfIntLessEqualSlotConstCompareFastPath(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	returnVal := runtime.NewSmallInt(1, runtime.IntegerI32)
	vm.slots = []runtime.Value{runtime.NewSmallInt(2, runtime.IntegerI32)}
	instr := &bytecodeInstruction{
		op:              bytecodeOpReturnConstIfIntLessEqualSlotConst,
		argCount:        0,
		operator:        "<",
		value:           returnVal,
		intImmediate:    runtime.NewSmallInt(3, runtime.IntegerI32),
		intImmediateRaw: 3,
		hasIntImmediate: true,
		hasIntRaw:       true,
	}

	got, returned, err := vm.execReturnConstIfIntLessEqualSlotConst(instr, nil)
	if err != nil {
		t.Fatalf("unexpected compare return-const-if error: %v", err)
	}
	if !returned || !valuesEqual(got, returnVal) {
		t.Fatalf("expected compare return-const-if fast path to return %#v, got=%#v returned=%v", returnVal, got, returned)
	}

	vm.ip = 7
	vm.slots[0] = runtime.NewSmallInt(3, runtime.IntegerI32)
	got, returned, err = vm.execReturnConstIfIntLessEqualSlotConst(instr, nil)
	if err != nil {
		t.Fatalf("unexpected false compare return-const-if error: %v", err)
	}
	if returned || got != nil {
		t.Fatalf("expected false compare return-const-if fast path not to return, got=%#v returned=%v", got, returned)
	}
	if vm.ip != 8 {
		t.Fatalf("expected false compare return-const-if to advance ip, got %d", vm.ip)
	}
}

func TestBytecodeVM_ReturnIfIntLessEqualSlotConstSameSlotCompareFastPath(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	trueVal := runtime.NewSmallInt(1, runtime.IntegerI32)
	vm.slots = []runtime.Value{trueVal}
	instr := &bytecodeInstruction{
		op:              bytecodeOpReturnIfIntLessEqualSlotConst,
		target:          0,
		argCount:        0,
		operator:        "<",
		intImmediate:    runtime.NewSmallInt(2, runtime.IntegerI32),
		intImmediateRaw: 2,
		hasIntImmediate: true,
		hasIntRaw:       true,
	}

	got, returned, err := vm.execReturnIfIntLessEqualSlotConst(instr, nil)
	if err != nil {
		t.Fatalf("unexpected compare return-if error: %v", err)
	}
	if !returned || !valuesEqual(got, trueVal) {
		t.Fatalf("expected compare return-if same-slot fast path to return %#v, got=%#v returned=%v", trueVal, got, returned)
	}

	vm.ip = 7
	vm.slots[0] = runtime.NewSmallInt(2, runtime.IntegerI32)
	got, returned, err = vm.execReturnIfIntLessEqualSlotConst(instr, nil)
	if err != nil {
		t.Fatalf("unexpected false compare return-if error: %v", err)
	}
	if returned || got != nil {
		t.Fatalf("expected false compare return-if same-slot fast path not to return, got=%#v returned=%v", got, returned)
	}
	if vm.ip != 8 {
		t.Fatalf("expected false compare return-if to advance ip, got %d", vm.ip)
	}
}

func TestBytecodeVM_LoweringEmitsReturnIfForIntLessEqualConstSlotStatement(t *testing.T) {
	def := ast.Fn(
		"fib",
		[]*ast.FunctionParameter{ast.Param("n", ast.Ty("i32"))},
		[]ast.Statement{
			ast.IfExpr(
				ast.Bin(">=", ast.Int(1), ast.ID("n")),
				ast.Block(ast.Ret(ast.ID("n"))),
			),
			ast.Bin(
				"+",
				ast.Call("fib", ast.Bin("-", ast.ID("n"), ast.Int(1))),
				ast.Call("fib", ast.Bin("-", ast.ID("n"), ast.Int(2))),
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
	found := false
	for _, instr := range program.instructions {
		if instr.op == bytecodeOpReturnIfIntLessEqualSlotConst && instr.operator == "<=" && instr.hasIntRaw && instr.intImmediateRaw == 1 {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected reversed <= compare to normalize into fused return-if less-equal slot-const opcode")
	}
	if bytecodeProgramContainsOpcode(program, bytecodeOpJumpIfIntLessEqualSlotConstFalse) {
		t.Fatalf("expected reversed less-equal return-if shape to skip standalone less-equal jump opcode")
	}
}

func TestBytecodeVM_LoweringEmitsEqualityBasePrefixReturnConstIfStatements(t *testing.T) {
	def := ast.Fn(
		"fib",
		[]*ast.FunctionParameter{ast.Param("n", ast.Ty("i32"))},
		[]ast.Statement{
			ast.IfExpr(
				ast.Bin("==", ast.ID("n"), ast.Int(0)),
				ast.Block(ast.Ret(ast.Int(0))),
			),
			ast.IfExpr(
				ast.Bin("==", ast.ID("n"), ast.Int(1)),
				ast.Block(ast.Ret(ast.Int(1))),
			),
			ast.Bin(
				"+",
				ast.Call("fib", ast.Bin("-", ast.ID("n"), ast.Int(1))),
				ast.Call("fib", ast.Bin("-", ast.ID("n"), ast.Int(2))),
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
	if len(program.instructions) < 2 {
		t.Fatalf("expected equality-base prefix instructions, got %d instructions", len(program.instructions))
	}
	first := program.instructions[0]
	second := program.instructions[1]
	if first.op != bytecodeOpReturnConstIfIntLessEqualSlotConst || first.operator != "==" || !first.hasIntRaw || first.intImmediateRaw != 0 {
		t.Fatalf("expected first equality base guard to lower to fused return-const-if for n == 0, got %#v", first)
	}
	if second.op != bytecodeOpReturnConstIfIntLessEqualSlotConst || second.operator != "==" || !second.hasIntRaw || second.intImmediateRaw != 1 {
		t.Fatalf("expected second equality base guard to lower to fused return-const-if for n == 1, got %#v", second)
	}
	if bytecodeProgramContainsOpcode(program, bytecodeOpJumpIfIntCompareSlotConstFalse) {
		t.Fatalf("expected equality-base prefix to skip standalone compare jump opcodes")
	}
}

func TestBytecodeVM_LoweringEmitsEqualityPrefixPlusRangeReturnConstIfStatements(t *testing.T) {
	def := ast.Fn(
		"fib",
		[]*ast.FunctionParameter{ast.Param("n", ast.Ty("i32"))},
		[]ast.Statement{
			ast.IfExpr(
				ast.Bin("==", ast.ID("n"), ast.Int(0)),
				ast.Block(ast.Ret(ast.Int(0))),
			),
			ast.IfExpr(
				ast.Bin("<=", ast.ID("n"), ast.Int(2)),
				ast.Block(ast.Ret(ast.Int(1))),
			),
			ast.Bin(
				"+",
				ast.Call("fib", ast.Bin("-", ast.ID("n"), ast.Int(1))),
				ast.Call("fib", ast.Bin("-", ast.ID("n"), ast.Int(2))),
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
	if len(program.instructions) < 2 {
		t.Fatalf("expected mixed equality/range base prefix instructions, got %d instructions", len(program.instructions))
	}
	first := program.instructions[0]
	second := program.instructions[1]
	if first.op != bytecodeOpReturnConstIfIntLessEqualSlotConst || first.operator != "==" || !first.hasIntRaw || first.intImmediateRaw != 0 {
		t.Fatalf("expected first mixed base guard to lower to fused return-const-if for n == 0, got %#v", first)
	}
	if second.op != bytecodeOpReturnConstIfIntLessEqualSlotConst || second.operator != "<=" || !second.hasIntRaw || second.intImmediateRaw != 2 {
		t.Fatalf("expected second mixed base guard to lower to fused return-const-if for n <= 2, got %#v", second)
	}
	if bytecodeProgramContainsOpcode(program, bytecodeOpJumpIfIntLessEqualSlotConstFalse) {
		t.Fatalf("expected mixed equality/range prefix to skip standalone less-equal jump opcodes")
	}
}

func TestBytecodeVM_LoweringEmitsOutOfOrderEqualityPlusCurrentRangeReturnIfStatements(t *testing.T) {
	def := ast.Fn(
		"fib",
		[]*ast.FunctionParameter{ast.Param("n", ast.Ty("i32"))},
		[]ast.Statement{
			ast.IfExpr(
				ast.Bin("==", ast.ID("n"), ast.Int(2)),
				ast.Block(ast.Ret(ast.Int(1))),
			),
			ast.IfExpr(
				ast.Bin("<=", ast.ID("n"), ast.Int(1)),
				ast.Block(ast.Ret(ast.ID("n"))),
			),
			ast.Bin(
				"+",
				ast.Call("fib", ast.Bin("-", ast.ID("n"), ast.Int(1))),
				ast.Call("fib", ast.Bin("-", ast.ID("n"), ast.Int(2))),
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
	if len(program.instructions) < 2 {
		t.Fatalf("expected out-of-order equality/current-range base instructions, got %d instructions", len(program.instructions))
	}
	first := program.instructions[0]
	second := program.instructions[1]
	if first.op != bytecodeOpReturnConstIfIntLessEqualSlotConst || first.operator != "==" || !first.hasIntRaw || first.intImmediateRaw != 2 {
		t.Fatalf("expected first mixed base guard to lower to fused return-const-if for n == 2, got %#v", first)
	}
	if second.op != bytecodeOpReturnIfIntLessEqualSlotConst || second.operator != "<=" || !second.hasIntRaw || second.intImmediateRaw != 1 {
		t.Fatalf("expected second mixed base guard to lower to fused return-if for n <= 1, got %#v", second)
	}
	if bytecodeProgramContainsOpcode(program, bytecodeOpJumpIfIntCompareSlotConstFalse) ||
		bytecodeProgramContainsOpcode(program, bytecodeOpJumpIfIntLessEqualSlotConstFalse) {
		t.Fatalf("expected out-of-order equality/current-range prefix to skip standalone compare jump opcodes")
	}
}
