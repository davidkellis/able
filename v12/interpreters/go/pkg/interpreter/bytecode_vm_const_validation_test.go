package interpreter

import (
	"math/big"
	"strings"
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_IntegerLiteralValidationRemainsLazy(t *testing.T) {
	i8 := ast.IntegerTypeI8
	overflowLiteral := ast.NewIntegerLiteral(big.NewInt(200), &i8)
	fn := ast.Fn(
		"f",
		[]*ast.FunctionParameter{ast.Param("flag", ast.Ty("bool"))},
		[]ast.Statement{
			ast.IfExpr(ast.ID("flag"), ast.Block(ast.Ret(ast.Int(1)))),
			overflowLiteral,
		},
		nil,
		nil,
		nil,
		false,
		false,
	)
	module := ast.Mod([]ast.Statement{fn}, nil, nil)

	byteInterp := NewBytecode()
	runBytecodeModuleWithInterpreter(t, byteInterp, module)
	fnValue, err := byteInterp.GlobalEnvironment().Get("f")
	if err != nil {
		t.Fatalf("lookup f: %v", err)
	}

	result, err := byteInterp.CallFunction(fnValue, []runtime.Value{runtime.BoolValue{Val: true}})
	if err != nil {
		t.Fatalf("expected true path to avoid overflow literal evaluation: %v", err)
	}
	if !valuesEqual(result, runtime.NewSmallInt(1, runtime.IntegerI32)) {
		t.Fatalf("unexpected true-path result: got=%#v", result)
	}

	_, err = byteInterp.CallFunction(fnValue, []runtime.Value{runtime.BoolValue{Val: false}})
	if err == nil {
		t.Fatalf("expected overflow error when evaluating literal on false path")
	}
	if !strings.Contains(err.Error(), "integer overflow") {
		t.Fatalf("expected integer overflow error, got: %v", err)
	}
}

func TestBytecodeVM_IntegerLiteralLoweringUsesSmallIntWhenPossible(t *testing.T) {
	def := ast.Fn(
		"f",
		nil,
		[]ast.Statement{
			ast.Assign(ast.ID("x"), ast.Int(48)),
			ast.ID("x"),
		},
		nil,
		nil,
		nil,
		false,
		false,
	)
	program, err := NewBytecode().lowerFunctionDefinitionBytecode(def)
	if err != nil {
		t.Fatalf("lower small integer literal: %v", err)
	}
	var smallConst *bytecodeInstruction
	for idx := range program.instructions {
		if program.instructions[idx].op == bytecodeOpConst {
			smallConst = &program.instructions[idx]
			break
		}
	}
	if smallConst == nil {
		t.Fatalf("expected const instruction, got %#v", program.instructions)
	}
	intVal, ok := smallConst.value.(runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected integer const, got %#v", smallConst.value)
	}
	if !intVal.IsSmall() || intVal.Int64Fast() != 48 || intVal.TypeSuffix != runtime.IntegerI32 {
		t.Fatalf("unexpected small integer const: %#v", intVal)
	}

	large, ok := new(big.Int).SetString("9223372036854775808", 10)
	if !ok {
		t.Fatalf("failed to build large integer literal")
	}
	largeDef := ast.Fn(
		"g",
		nil,
		[]ast.Statement{ast.IntBig(large, nil)},
		nil,
		nil,
		nil,
		false,
		false,
	)
	largeProgram, err := NewBytecode().lowerFunctionDefinitionBytecode(largeDef)
	if err != nil {
		t.Fatalf("lower large integer literal: %v", err)
	}
	largeInt, ok := largeProgram.instructions[0].value.(runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected large integer const, got %#v", largeProgram.instructions[0].value)
	}
	if largeInt.IsSmall() || largeInt.Val == nil {
		t.Fatalf("large integer literal should stay big-backed, got %#v", largeInt)
	}
}

func TestBytecodeVM_ResetForRunPreservesConstCaches(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	program := &bytecodeProgram{
		instructions: []bytecodeInstruction{
			{op: bytecodeOpConst, value: runtime.NewSmallInt(1, runtime.IntegerI32)},
		},
	}

	validatedFirst := vm.validatedIntegerConstSlots(program)
	if len(validatedFirst) != len(program.instructions) {
		t.Fatalf("unexpected validation slice length: got=%d want=%d", len(validatedFirst), len(program.instructions))
	}
	tableFirst := vm.slotConstImmediateTable(program)
	if tableFirst == nil {
		t.Fatalf("expected immediate table cache")
	}

	vm.resetForRun(interp, interp.GlobalEnvironment())

	validatedSecond := vm.validatedIntegerConstSlots(program)
	if len(validatedSecond) != len(program.instructions) {
		t.Fatalf("unexpected validation slice length after reset: got=%d want=%d", len(validatedSecond), len(program.instructions))
	}
	if len(validatedFirst) > 0 && &validatedFirst[0] != &validatedSecond[0] {
		t.Fatalf("expected validation cache backing storage to be reused across pooled runs")
	}
	tableSecond := vm.slotConstImmediateTable(program)
	if tableSecond != tableFirst {
		t.Fatalf("expected immediate table cache to be reused across pooled runs")
	}
}
