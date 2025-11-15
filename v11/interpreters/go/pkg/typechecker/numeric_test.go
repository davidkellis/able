package typechecker

import (
	"able/interpreter10-go/pkg/ast"
	"strings"
	"testing"
)

func TestBinaryExpressionPromotesSignedIntegers(t *testing.T) {
	checker := New()
	i16 := ast.IntegerTypeI16
	i64 := ast.IntegerTypeI64
	expr := ast.Bin("+", ast.IntTyped(1, &i16), ast.IntTyped(2, &i64))
	assign := ast.Assign(ast.ID("sum"), expr)
	module := ast.NewModule([]ast.Statement{assign}, nil, nil)

	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	typ, ok := checker.infer[expr]
	if !ok {
		t.Fatalf("expected inferred type for expression")
	}
	if name := typeName(typ); name != "i64" {
		t.Fatalf("expected promoted type i64, got %s", name)
	}
}

func TestBinaryExpressionPromotesMixedSignedUnsigned(t *testing.T) {
	checker := New()
	i8 := ast.IntegerTypeI8
	u16 := ast.IntegerTypeU16
	expr := ast.Bin("+", ast.IntTyped(1, &i8), ast.IntTyped(2, &u16))
	assign := ast.Assign(ast.ID("value"), expr)
	module := ast.NewModule([]ast.Statement{assign}, nil, nil)

	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	typ, ok := checker.infer[expr]
	if !ok {
		t.Fatalf("expected inferred type for expression")
	}
	if name := typeName(typ); name != "i32" {
		t.Fatalf("expected promoted type i32, got %s", name)
	}
}

func TestBinaryExpressionFallsBackToUnsignedForU128(t *testing.T) {
	checker := New()
	u128 := ast.IntegerTypeU128
	i32 := ast.IntegerTypeI32
	expr := ast.Bin("+", ast.IntTyped(1, &u128), ast.IntTyped(2, &i32))
	assign := ast.Assign(ast.ID("sum"), expr)
	module := ast.NewModule([]ast.Statement{assign}, nil, nil)

	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	typ, ok := checker.infer[expr]
	if !ok {
		t.Fatalf("expected inferred type for expression")
	}
	if name := typeName(typ); name != "u128" {
		t.Fatalf("expected promoted type u128, got %s", name)
	}
}

func TestBinaryExpressionReportsOverflowWhenNoWidthAvailable(t *testing.T) {
	checker := New()
	i128 := ast.IntegerTypeI128
	u64 := ast.IntegerTypeU64
	expr := ast.Bin("+", ast.IntTyped(1, &i128), ast.IntTyped(2, &u64))
	assign := ast.Assign(ast.ID("value"), expr)
	module := ast.NewModule([]ast.Statement{assign}, nil, nil)

	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) == 0 {
		t.Fatalf("expected diagnostics for exceeding width")
	}
	if want := "exceeding available widths"; !strings.Contains(diags[0].Message, want) {
		t.Fatalf("expected diagnostic to mention %q, got %q", want, diags[0].Message)
	}
}

func TestBitwiseRequiresIntegerOperands(t *testing.T) {
	checker := New()
	expr := ast.Bin("&", ast.Int(1), ast.Str("bad"))
	assign := ast.Assign(ast.ID("value"), expr)
	module := ast.NewModule([]ast.Statement{assign}, nil, nil)

	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) == 0 {
		t.Fatalf("expected diagnostic for bitwise operands")
	}
	if want := "requires integer operands"; !strings.Contains(diags[0].Message, want) {
		t.Fatalf("expected diagnostic to mention %q, got %q", want, diags[0].Message)
	}
}

func TestComparisonRequiresNumericOperands(t *testing.T) {
	checker := New()
	assignString := ast.Assign(ast.ID("text"), ast.Str("hello"))
	compare := ast.Assign(ast.ID("flag"), ast.Bin("<", ast.ID("text"), ast.Int(1)))
	module := ast.NewModule([]ast.Statement{assignString, compare}, nil, nil)

	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) == 0 {
		t.Fatalf("expected comparison diagnostic")
	}
	if want := "requires numeric operands"; !strings.Contains(diags[0].Message, want) {
		t.Fatalf("expected diagnostic to mention %q, got %q", want, diags[0].Message)
	}
}
