package typechecker

import (
	"strings"
	"testing"

	"able/interpreter-go/pkg/ast"
)

func TestUnionNormalizationWarnsOnRedundantMembers(t *testing.T) {
	checker := New()
	alias := ast.NewTypeAliasDefinition(
		ast.ID("MaybeInt"),
		ast.UnionT(ast.Nullable(ast.Ty("i32")), ast.Ty("i32")),
		nil,
		nil,
		false,
	)
	module := ast.NewModule([]ast.Statement{alias}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 1 {
		t.Fatalf("expected one diagnostic, got %v", diags)
	}
	if diags[0].Severity != SeverityWarning {
		t.Fatalf("expected warning severity, got %v", diags[0].Severity)
	}
	if !strings.Contains(diags[0].Message, "redundant union member i32") {
		t.Fatalf("expected redundant union warning, got %q", diags[0].Message)
	}
}

func TestNormalizeUnionTypesCollapsesDuplicates(t *testing.T) {
	normalized := normalizeUnionTypes([]Type{
		IntegerType{Suffix: "i32"},
		IntegerType{Suffix: "i32"},
	})
	if _, ok := normalized.(IntegerType); !ok {
		t.Fatalf("expected normalized union to collapse to IntegerType, got %T", normalized)
	}
}

func TestNormalizeUnionTypesCollapsesNil(t *testing.T) {
	normalized := normalizeUnionTypes([]Type{
		PrimitiveType{Kind: PrimitiveNil},
		PrimitiveType{Kind: PrimitiveString},
	})
	nullable, ok := normalized.(NullableType)
	if !ok {
		t.Fatalf("expected normalized union to be NullableType, got %T", normalized)
	}
	if name := typeName(nullable.Inner); name != "String" {
		t.Fatalf("expected nullable inner String, got %q", name)
	}
}
