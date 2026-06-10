package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_ZeroFieldNamedStructLiteralReturnsDefinitionAndMatchesStructPattern(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.StructDef("Service", nil, ast.StructKindNamed, nil, nil, false),
		ast.Methods(ast.Ty("Service"), []*ast.FunctionDefinition{
			ast.Fn(
				"id",
				nil,
				[]ast.Statement{ast.Ret(ast.Int(7))},
				ast.Ty("i32"),
				nil,
				nil,
				true,
				false,
			),
		}, nil, nil),
		ast.Fn(
			"make_service",
			nil,
			[]ast.Statement{ast.Ret(ast.StructLit(nil, false, "Service", nil, nil))},
			ast.Ty("Service"),
			nil,
			nil,
			false,
			false,
		),
		ast.Assign(ast.ID("svc"), ast.Call("make_service")),
		ast.Match(
			ast.ID("svc"),
			ast.Mc(ast.StructP(nil, false, "Service"), ast.CallExpr(ast.Member(ast.ID("svc"), "id"))),
			ast.Mc(ast.Wc(), ast.Int(0)),
		),
	}, nil, nil)

	interp := NewBytecode()
	result, env, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	intResult, ok := result.(runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected integer result, got %T (%#v)", result, result)
	}
	if val, ok := intResult.ToInt64(); !ok || val != 7 {
		t.Fatalf("unexpected result: got=%#v want=7", result)
	}
	svc, err := env.Get("svc")
	if err != nil {
		t.Fatalf("missing svc binding: %v", err)
	}
	if _, ok := svc.(*runtime.StructDefinitionValue); !ok {
		t.Fatalf("svc binding = %T, want *runtime.StructDefinitionValue", svc)
	}

	program := mustBytecodeFunctionProgram(t, interp, "make_service")
	if !bytecodeProgramContainsOpcode(program, bytecodeOpStructLiteralNamedFast) {
		t.Fatalf("expected zero-field named literal to keep the named fast opcode")
	}
}
