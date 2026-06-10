package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_EnvMatchLoweringAvoidsGenericMatchForSlotlessTypedBinding(t *testing.T) {
	interp := NewBytecode()
	expr := ast.Match(
		ast.Int(5),
		ast.Mc(ast.LitP(ast.Nil()), ast.Int(0)),
		ast.Mc(ast.TypedP(ast.ID("n"), ast.Ty("i32")), ast.Bin("+", ast.ID("n"), ast.Int(2))),
	)
	program, err := interp.lowerExpressionToBytecode(expr)
	if err != nil {
		t.Fatalf("lowerExpressionToBytecode failed: %v", err)
	}
	for _, instr := range program.instructions {
		if instr.op == bytecodeOpMatch {
			t.Fatalf("slotless env-match lowering emitted generic match opcode: %#v", program.instructions)
		}
	}

	got, err := newBytecodeVM(interp, interp.GlobalEnvironment()).run(program)
	if err != nil {
		t.Fatalf("bytecode run failed: %v", err)
	}
	intVal, ok := got.(runtime.IntegerValue)
	gotVal, gotOK := intVal.ToInt64()
	if !ok || !gotOK || intVal.TypeSuffix != runtime.IntegerI32 || gotVal != 7 {
		t.Fatalf("env-lowered match result = %#v, want i32 7", got)
	}
}

func TestBytecodeVM_EnvMatchLoweringKeepsTypedBindingScoped(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Match(
			ast.Int(5),
			ast.Mc(ast.TypedP(ast.ID("n"), ast.Ty("i32")), ast.ID("n")),
		),
	}, nil, nil)
	interp := NewBytecode()
	got, env, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("bytecode module failed: %v", err)
	}
	if !valuesEqual(got, runtime.NewSmallInt(5, runtime.IntegerI32)) {
		t.Fatalf("env-lowered match result = %#v, want i32 5", got)
	}
	if _, err := env.Get("n"); err == nil {
		t.Fatalf("env-lowered match binding leaked into outer scope")
	}
}

func TestBytecodeVM_EnvMatchLoweringKeepsDirectBodyDeclarationScoped(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Match(
			ast.Int(5),
			ast.Mc(ast.TypedP(ast.ID("n"), ast.Ty("i32")), ast.Assign(ast.ID("inner"), ast.Int(9))),
		),
	}, nil, nil)
	interp := NewBytecode()
	got, env, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("bytecode module failed: %v", err)
	}
	if !valuesEqual(got, runtime.NewSmallInt(9, runtime.IntegerI32)) {
		t.Fatalf("env-lowered match result = %#v, want i32 9", got)
	}
	if _, err := env.Get("n"); err == nil {
		t.Fatalf("env-lowered match pattern binding leaked into outer scope")
	}
	if _, err := env.Get("inner"); err == nil {
		t.Fatalf("env-lowered match body declaration leaked into outer scope")
	}
}

func TestBytecodeVM_EnvMatchLoweringKeepsGuardedMatchOnGenericOpcode(t *testing.T) {
	interp := NewBytecode()
	expr := ast.Match(
		ast.Int(5),
		ast.Mc(ast.TypedP(ast.ID("n"), ast.Ty("i32")), ast.ID("n"), ast.Bool(true)),
	)
	program, err := interp.lowerExpressionToBytecode(expr)
	if err != nil {
		t.Fatalf("lowerExpressionToBytecode failed: %v", err)
	}
	for _, instr := range program.instructions {
		if instr.op == bytecodeOpMatch {
			return
		}
	}
	t.Fatalf("guarded match should remain on generic match opcode, got %#v", program.instructions)
}
