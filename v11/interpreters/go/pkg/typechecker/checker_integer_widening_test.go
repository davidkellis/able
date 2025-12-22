package typechecker

import (
	"testing"

	"able/interpreter-go/pkg/ast"
)

func TestFunctionCallAllowsIntegerWidening(t *testing.T) {
	checker := New()
	fn := ast.Fn(
		"takes_i64",
		[]*ast.FunctionParameter{
			ast.Param("value", ast.Ty("i64")),
		},
		[]ast.Statement{
			ast.Block(ast.Nil()),
		},
		ast.Ty("void"),
		nil,
		nil,
		false,
		false,
	)
	call := ast.Call("takes_i64", ast.Int(1))
	module := ast.NewModule([]ast.Statement{fn, call}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics for widening call, got %v", diags)
	}
}

func TestFunctionReturnAllowsIntegerWidening(t *testing.T) {
	checker := New()
	fn := ast.Fn(
		"make_i64",
		nil,
		[]ast.Statement{
			ast.Assign(ast.ID("value"), ast.Int(1)),
			ast.Ret(ast.ID("value")),
		},
		ast.Ty("i64"),
		nil,
		nil,
		false,
		false,
	)
	module := ast.NewModule([]ast.Statement{fn}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics for widening return, got %v", diags)
	}
}
