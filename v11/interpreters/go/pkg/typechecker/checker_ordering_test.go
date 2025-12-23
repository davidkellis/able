package typechecker

import (
	"able/interpreter-go/pkg/ast"
	"testing"
)

func TestStringCmpReturnsOrdering(t *testing.T) {
	checker := New()
	call := ast.CallExpr(ast.Member(ast.Str("a"), "cmp"), ast.Str("b"))
	fn := ast.Fn(
		"cmp_wrapper",
		nil,
		[]ast.Statement{
			ast.Block(call),
		},
		ast.Ty("Ordering"),
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
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	if typ, ok := checker.infer[call]; !ok {
		t.Fatalf("expected inference entry for cmp call")
	} else if typeName(typ) != "Ordering" {
		t.Fatalf("expected cmp call to return Ordering, got %q", typeName(typ))
	}
}

func TestOrdConstraintAllowsCmp(t *testing.T) {
	checker := New()
	generic := ast.GenericParam("T", ast.InterfaceConstr(ast.Ty("Ord")))
	compareFn := ast.Fn(
		"compare",
		[]*ast.FunctionParameter{
			ast.Param("left", ast.Ty("T")),
			ast.Param("right", ast.Ty("T")),
		},
		[]ast.Statement{
			ast.Block(ast.CallExpr(ast.Member(ast.ID("left"), "cmp"), ast.ID("right"))),
		},
		ast.Ty("Ordering"),
		[]*ast.GenericParameter{generic},
		nil,
		false,
		false,
	)
	module := ast.NewModule([]ast.Statement{compareFn}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
}
