package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_SlotLoweringEmitsTypedPrimitiveMatchOpcode(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Fn("f", nil, []ast.Statement{
			ast.Assign(ast.ID("x"), ast.NewTypeCastExpression(ast.Int(7), ast.Ty("u8"))),
			ast.Match(
				ast.ID("x"),
				ast.Mc(
					ast.TypedP(ast.ID("byte"), ast.Ty("u8")),
					ast.Bin("+", ast.NewTypeCastExpression(ast.ID("byte"), ast.Ty("i32")), ast.Int(1)),
				),
				ast.Mc(ast.LitP(ast.Nil()), ast.Int(0)),
			),
		}, ast.Ty("i32"), nil, nil, false, false),
		ast.Call("f"),
	}, nil, nil)

	interp := NewBytecode()
	got := runBytecodeModuleWithInterpreter(t, interp, module)
	want := mustEvalModule(t, New(), module)
	if !valuesEqual(got, want) {
		t.Fatalf("slot-lowered primitive match mismatch: got=%#v want=%#v", got, want)
	}

	program := mustBytecodeFunctionProgram(t, interp, "f")
	if program.frameLayout == nil {
		t.Fatalf("expected function f to use slot frame layout")
	}
	if !bytecodeProgramContainsOpcode(program, bytecodeOpJumpIfNotTypedPattern) {
		t.Fatalf("expected slot-lowered typed-pattern match opcode")
	}
	if !bytecodeProgramContainsOpcode(program, bytecodeOpJumpIfNotNil) {
		t.Fatalf("expected slot-lowered nil-pattern match opcode")
	}
	if bytecodeProgramContainsOpcode(program, bytecodeOpMatch) {
		t.Fatalf("slot-lowered primitive match should not emit generic Match opcode")
	}
}

func TestBytecodeVM_SlotLoweringEmitsTypedNominalMatchOpcode(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.StructDef("Node", nil, ast.StructKindNamed, nil, nil, false),
		ast.Fn("f", []*ast.FunctionParameter{
			ast.Param("value", ast.Nullable(ast.Ty("Node"))),
		}, []ast.Statement{
			ast.Match(
				ast.ID("value"),
				ast.Mc(ast.LitP(ast.Nil()), ast.Int(0)),
				ast.Mc(ast.TypedP(ast.ID("node"), ast.Ty("Node")), ast.Int(1)),
			),
		}, ast.Ty("i32"), nil, nil, false, false),
		ast.Call("f", ast.StructLit(nil, false, "Node", nil, nil)),
	}, nil, nil)

	interp := NewBytecode()
	got := runBytecodeModuleWithInterpreter(t, interp, module)
	if intResult, ok := got.(runtime.IntegerValue); !ok {
		t.Fatalf("expected integer result, got %T (%#v)", got, got)
	} else if val, ok := intResult.ToInt64(); !ok || val != 1 {
		t.Fatalf("unexpected nominal match result: got=%#v want=1", got)
	}

	program := mustBytecodeFunctionProgram(t, interp, "f")
	if program.frameLayout == nil {
		t.Fatalf("expected function f to use slot frame layout")
	}
	if !bytecodeProgramContainsOpcode(program, bytecodeOpJumpIfNotTypedPattern) {
		t.Fatalf("expected slot-lowered nominal typed-pattern match opcode")
	}
	if bytecodeProgramContainsOpcode(program, bytecodeOpMatch) {
		t.Fatalf("slot-lowered nominal match should not emit generic Match opcode")
	}
}

func mustBytecodeFunctionProgram(t *testing.T, interp *Interpreter, name string) *bytecodeProgram {
	t.Helper()
	raw, err := interp.GlobalEnvironment().Get(name)
	if err != nil {
		t.Fatalf("lookup function %s: %v", name, err)
	}
	fn, ok := raw.(*runtime.FunctionValue)
	if !ok || fn == nil {
		t.Fatalf("expected function %s, got %T", name, raw)
	}
	program, ok := fn.Bytecode.(*bytecodeProgram)
	if !ok || program == nil {
		t.Fatalf("expected bytecode program for %s", name)
	}
	return program
}
