package typechecker

import (
	"math/big"
	"testing"

	"able/interpreter-go/pkg/ast"
)

func TestExplicitGenericTypeArgumentAdoptsUnsuffixedIntegerLiteral(t *testing.T) {
	checker := New()
	identity := genericIdentityDefinition()
	large, ok := new(big.Int).SetString("3000000000", 10)
	if !ok {
		t.Fatal("parse test literal")
	}
	call := ast.Call("identity", ast.IntBig(large, nil))
	call.TypeArguments = []ast.TypeExpression{ast.Ty("i64")}
	module := ast.NewModule([]ast.Statement{identity, ast.Assign(ast.ID("value"), call)}, nil, nil)

	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected explicit i64 context to adopt the literal, got %v", diags)
	}
	if got := typeName(checker.infer[call]); got != "i64" {
		t.Fatalf("call inferred as %s, want i64", got)
	}
}

func TestExpectedGenericReturnAdoptsUnsuffixedIntegerLiteral(t *testing.T) {
	checker := New()
	identity := genericIdentityDefinition()
	large, ok := new(big.Int).SetString("3000000000", 10)
	if !ok {
		t.Fatal("parse test literal")
	}
	call := ast.Call("identity", ast.IntBig(large, nil))
	assign := ast.Assign(ast.TypedP(ast.ID("value"), ast.Ty("i64")), call)
	module := ast.NewModule([]ast.Statement{identity, assign}, nil, nil)

	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected result context to adopt the literal, got %v", diags)
	}
	if got := typeName(checker.infer[call]); got != "i64" {
		t.Fatalf("call inferred as %s, want i64", got)
	}
}

func genericIdentityDefinition() *ast.FunctionDefinition {
	return ast.Fn(
		"identity",
		[]*ast.FunctionParameter{ast.Param("value", ast.Ty("T"))},
		[]ast.Statement{ast.ID("value")},
		ast.Ty("T"),
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
		false,
	)
}
