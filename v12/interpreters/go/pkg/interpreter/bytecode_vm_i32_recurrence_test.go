package interpreter

import (
	"math"
	"strings"
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVMDetectsI32RecurrenceKernel(t *testing.T) {
	program := lowerI32RecurrenceTestProgram(t)
	if program.i32RecurrenceKernel == nil {
		t.Fatalf("expected i32 recurrence kernel")
	}
}

func TestBytecodeVMI32RecurrenceKernelParity(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		i32RecurrenceTestFunction("fib"),
		ast.Call("fib", ast.Int(10)),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode i32 recurrence mismatch: got=%#v want=%#v", got, want)
	}
	assertIntValue(t, got, runtime.IntegerI32, 55)
}

func TestBytecodeVMI32RecurrenceKernelOverflowParity(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Fn(
			"boom",
			[]*ast.FunctionParameter{ast.Param("n", ast.Ty("i32"))},
			[]ast.Statement{
				ast.IfExpr(
					ast.Bin("<=", ast.ID("n"), ast.Int(0)),
					ast.Block(ast.Ret(ast.Int(math.MaxInt32))),
				),
				ast.Bin(
					"+",
					ast.Call("boom", ast.Bin("-", ast.ID("n"), ast.Int(1))),
					ast.Call("boom", ast.Bin("-", ast.ID("n"), ast.Int(1))),
				),
			},
			ast.Ty("i32"),
			nil,
			nil,
			false,
			false,
		),
		ast.Call("boom", ast.Int(1)),
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

func lowerI32RecurrenceTestProgram(t *testing.T) *bytecodeProgram {
	t.Helper()
	interp := NewBytecode()
	program, err := interp.lowerFunctionDefinitionBytecode(i32RecurrenceTestFunction("fib"))
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	return program
}

func i32RecurrenceTestFunction(name string) *ast.FunctionDefinition {
	return ast.Fn(
		name,
		[]*ast.FunctionParameter{ast.Param("n", ast.Ty("i32"))},
		[]ast.Statement{
			ast.IfExpr(
				ast.Bin("<=", ast.ID("n"), ast.Int(2)),
				ast.Block(ast.Ret(ast.Int(1))),
			),
			ast.Bin(
				"+",
				ast.Call(name, ast.Bin("-", ast.ID("n"), ast.Int(1))),
				ast.Call(name, ast.Bin("-", ast.ID("n"), ast.Int(2))),
			),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)
}
