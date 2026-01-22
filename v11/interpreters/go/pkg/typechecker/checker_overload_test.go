package typechecker

import (
	"strings"
	"testing"

	"able/interpreter-go/pkg/ast"
)

func TestOverloadedFunctionReportsAmbiguousCall(t *testing.T) {
	checker := New()
	fn1 := ast.Fn(
		"collide",
		[]*ast.FunctionParameter{
			ast.Param("value", ast.UnionT(ast.Ty("String"), ast.Ty("i32"))),
		},
		[]ast.Statement{ast.Str("num")},
		ast.Ty("String"),
		nil,
		nil,
		false,
		false,
	)
	fn2 := ast.Fn(
		"collide",
		[]*ast.FunctionParameter{
			ast.Param("value", ast.UnionT(ast.Ty("String"), ast.Ty("bool"))),
		},
		[]ast.Statement{ast.Str("flag")},
		ast.Ty("String"),
		nil,
		nil,
		false,
		false,
	)
	call := ast.Call("collide", ast.Str("boom"))
	module := ast.NewModule([]ast.Statement{fn1, fn2, call}, nil, nil)

	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	foundAmbiguous := false
	for _, diag := range diags {
		if strings.Contains(diag.Message, "duplicate declaration") {
			t.Fatalf("unexpected duplicate declaration diagnostic: %v", diag.Message)
		}
		if strings.Contains(diag.Message, "ambiguous overload") {
			foundAmbiguous = true
		}
	}
	if !foundAmbiguous {
		t.Fatalf("expected ambiguous overload diagnostic, got %v", diags)
	}
}
