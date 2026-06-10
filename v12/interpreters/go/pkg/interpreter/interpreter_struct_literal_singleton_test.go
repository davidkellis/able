package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestZeroFieldNamedStructLiteralReturnsDefinitionAndMatchesStructPattern(t *testing.T) {
	interp := New()
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
		ast.Assign(ast.ID("svc"), ast.StructLit(nil, false, "Service", nil, nil)),
		ast.Match(
			ast.ID("svc"),
			ast.Mc(ast.StructP(nil, false, "Service"), ast.CallExpr(ast.Member(ast.ID("svc"), "id"))),
			ast.Mc(ast.Wc(), ast.Int(0)),
		),
	}, nil, nil)

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
}

func TestZeroFieldNamedStructFunctionalUpdateAcceptsSingletonDefinition(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.StructDef("Done", nil, ast.StructKindNamed, nil, nil, false),
		ast.Assign(ast.ID("base"), ast.StructLit(nil, false, "Done", nil, nil)),
		ast.Match(
			ast.StructLit(nil, false, "Done", []ast.Expression{ast.ID("base")}, nil),
			ast.Mc(ast.StructP(nil, false, "Done"), ast.Int(1)),
			ast.Mc(ast.Wc(), ast.Int(0)),
		),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	intResult, ok := result.(runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected integer result, got %T (%#v)", result, result)
	}
	if val, ok := intResult.ToInt64(); !ok || val != 1 {
		t.Fatalf("unexpected result: got=%#v want=1", result)
	}
}
