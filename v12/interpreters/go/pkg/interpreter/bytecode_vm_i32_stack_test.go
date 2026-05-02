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
