package typechecker

import (
	"math/big"
	"strings"
	"testing"

	"able/interpreter10-go/pkg/ast"
)

func TestStructLiteralReportsLiteralOverflow(t *testing.T) {
	checker := New()
	structDef := ast.StructDef(
		"Packet",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("u8"), "value"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	literal := ast.StructLit(
		[]*ast.StructFieldInitializer{
			ast.FieldInit(ast.Int(512), "value"),
		},
		false,
		"Packet",
		nil,
		nil,
	)
	module := ast.NewModule([]ast.Statement{structDef, literal}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 1 {
		t.Fatalf("expected literal overflow diagnostic, got %v", diags)
	}
	if !strings.Contains(diags[0].Message, "literal 512 does not fit in u8") {
		t.Fatalf("expected literal overflow message, got %q", diags[0].Message)
	}
}

func TestStructLiteralReportsFieldTypeMismatch(t *testing.T) {
	checker := New()
	structDef := ast.StructDef(
		"Carrier",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("bool"), "flag"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	literal := ast.StructLit(
		[]*ast.StructFieldInitializer{
			ast.FieldInit(ast.Int(1), "flag"),
		},
		false,
		"Carrier",
		nil,
		nil,
	)
	module := ast.NewModule([]ast.Statement{structDef, literal}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 1 {
		t.Fatalf("expected type mismatch diagnostic, got %v", diags)
	}
	if !strings.Contains(diags[0].Message, "struct field 'flag' expects") {
		t.Fatalf("expected struct field mismatch message, got %q", diags[0].Message)
	}
}
func TestLiteralMismatchHelper(t *testing.T) {
	value := IntegerType{Suffix: "i32", Literal: big.NewInt(512)}
	expected := IntegerType{Suffix: "u8"}
	if msg, ok := literalMismatchMessage(value, expected); !ok || !strings.Contains(msg, "literal 512") {
		t.Fatalf("expected literal mismatch message, got ok=%v msg=%q", ok, msg)
	}
}
