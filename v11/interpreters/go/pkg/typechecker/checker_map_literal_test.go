package typechecker

import (
	"testing"

	"able/interpreter10-go/pkg/ast"
)

func TestMapLiteralInference(t *testing.T) {
	checker := New()
	assign := ast.Assign(
		ast.ID("headers"),
		ast.MapLit([]ast.MapLiteralElement{
			ast.MapEntry(ast.NewStringLiteral("content-type"), ast.NewStringLiteral("application/json")),
			ast.MapEntry(ast.NewStringLiteral("authorization"), ast.NewStringLiteral("token")),
		}),
	)
	module := ast.NewModule([]ast.Statement{assign}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
}

func TestMapLiteralSpreadTypeMismatch(t *testing.T) {
	checker := New()
	assign := ast.Assign(
		ast.ID("headers"),
		ast.MapLit([]ast.MapLiteralElement{
			ast.MapEntry(ast.NewStringLiteral("accept"), ast.NewStringLiteral("json")),
			ast.MapSpread(ast.NewStringLiteral("oops")),
		}),
	)
	module := ast.NewModule([]ast.Statement{assign}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 1 {
		t.Fatalf("expected diagnostic for invalid spread, got %v", diags)
	}
}

func TestMapLiteralInconsistentKeys(t *testing.T) {
	checker := New()
	assign := ast.Assign(
		ast.ID("codes"),
		ast.MapLit([]ast.MapLiteralElement{
			ast.MapEntry(ast.NewStringLiteral("ok"), ast.Int(200)),
			ast.MapEntry(ast.Bool(true), ast.Int(201)),
		}),
	)
	module := ast.NewModule([]ast.Statement{assign}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 1 {
		t.Fatalf("expected diagnostic for mixed key types, got %v", diags)
	}
}
