package interpreter

import (
	"math"
	"testing"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
)

func TestBitshiftRangeDiagnostics(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()
	cases := []struct {
		name string
		expr ast.Expression
		msg  string
	}{
		{
			name: "NegativeShift",
			expr: ast.Bin("<<", ast.Int(1), ast.Int(-1)),
			msg:  "shift out of range",
		},
		{
			name: "LargeShift",
			expr: ast.Bin(">>", ast.Int(1), ast.Int(32)),
			msg:  "shift out of range",
		},
	}
	for _, tc := range cases {
		_, err := interp.evaluateExpression(tc.expr, env)
		if err == nil {
			t.Fatalf("%s: expected error", tc.name)
		}
		if err.Error() != tc.msg {
			t.Fatalf("%s: expected %q, got %q", tc.name, tc.msg, err.Error())
		}
	}
}

func TestBitwiseRequiresI32Operands(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()
	cases := []ast.Expression{
		ast.Bin("&", ast.Int(1), ast.Flt(1.0)),
		ast.Bin("|", ast.Flt(1.0), ast.Int(1)),
		ast.Bin("^", ast.Flt(1.0), ast.Flt(1.0)),
	}
	for idx, expr := range cases {
		_, err := interp.evaluateExpression(expr, env)
		if err == nil {
			t.Fatalf("case %d: expected error", idx)
		}
		if err.Error() != "Bitwise requires i32 operands" {
			t.Fatalf("case %d: expected 'Bitwise requires i32 operands', got %q", idx, err.Error())
		}
	}
}

func TestDivisionByZeroDiagnostics(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()
	cases := []ast.Expression{
		ast.Bin("/", ast.Int(4), ast.Int(0)),
		ast.Bin("%", ast.Int(4), ast.Int(0)),
	}
	for idx, expr := range cases {
		_, err := interp.evaluateExpression(expr, env)
		if err == nil {
			t.Fatalf("case %d: expected error", idx)
		}
		if err.Error() != "division by zero" {
			t.Fatalf("case %d: expected 'division by zero', got %q", idx, err.Error())
		}
	}
}

func TestArithmeticMixedNumericTypes(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()
	cases := []struct {
		name   string
		expr   ast.Expression
		expect float64
	}{
		{"AddMixed", ast.Bin("+", ast.Int(1), ast.Flt(1.5)), 2.5},
		{"SubMixed", ast.Bin("-", ast.Flt(2.5), ast.Int(1)), 1.5},
		{"MulMixed", ast.Bin("*", ast.Int(2), ast.Flt(3.5)), 7},
		{"DivMixed", ast.Bin("/", ast.Flt(7), ast.Int(2)), 3.5},
	}
	for _, tc := range cases {
		val, err := interp.evaluateExpression(tc.expr, env)
		if err != nil {
			t.Fatalf("%s: unexpected error %v", tc.name, err)
		}
		fv, ok := val.(runtime.FloatValue)
		if !ok {
			t.Fatalf("%s: expected float result, got %#v", tc.name, val)
		}
		if math.Abs(fv.Val-tc.expect) > 1e-9 {
			t.Fatalf("%s: expected %v, got %v", tc.name, tc.expect, fv.Val)
		}
	}
}

func TestModuloMixedNumericTypes(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()
	cases := []struct {
		name   string
		expr   ast.Expression
		expect float64
	}{
		{"ModuloIntFloat", ast.Bin("%", ast.Int(5), ast.Flt(2)), 1},
		{"ModuloFloatInt", ast.Bin("%", ast.Flt(5), ast.Int(2)), 1},
	}
	for _, tc := range cases {
		val, err := interp.evaluateExpression(tc.expr, env)
		if err != nil {
			t.Fatalf("%s: unexpected error %v", tc.name, err)
		}
		fv, ok := val.(runtime.FloatValue)
		if ok {
			if math.Abs(fv.Val-tc.expect) > 1e-9 {
				t.Fatalf("%s: expected %v, got %v", tc.name, tc.expect, fv.Val)
			}
			continue
		}
		iv, ok := val.(runtime.IntegerValue)
		if !ok {
			t.Fatalf("%s: unexpected result %#v", tc.name, val)
		}
		if iv.Val.Cmp(bigInt(int64(tc.expect))) != 0 {
			t.Fatalf("%s: expected %v, got %v", tc.name, tc.expect, iv.Val)
		}
	}
}

func TestMixedNumericComparisons(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()
	cases := []struct {
		name   string
		expr   ast.Expression
		expect bool
	}{
		{"LessMixed", ast.Bin("<", ast.Int(2), ast.Flt(3.5)), true},
		{"GreaterMixed", ast.Bin(">", ast.Flt(3.5), ast.Int(2)), true},
		{"EqualMixed", ast.Bin("==", ast.Int(3), ast.Flt(3)), true},
		{"NotEqualMixed", ast.Bin("!=", ast.Int(3), ast.Flt(4)), true},
		{"LessMixedFalse", ast.Bin("<", ast.Flt(5), ast.Int(4)), false},
	}
	for _, tc := range cases {
		val, err := interp.evaluateExpression(tc.expr, env)
		if err != nil {
			t.Fatalf("%s: unexpected error %v", tc.name, err)
		}
		bv, ok := val.(runtime.BoolValue)
		if !ok {
			t.Fatalf("%s: expected bool, got %#v", tc.name, val)
		}
		if bv.Val != tc.expect {
			t.Fatalf("%s: expected %v, got %v", tc.name, tc.expect, bv.Val)
		}
	}
}

func TestUnsupportedBinaryOperatorError(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()
	cases := []struct {
		name string
		expr ast.Expression
		msg  string
	}{
		{"StringTimes", ast.Bin("*", ast.Str("hi"), ast.Str("there")), "Arithmetic requires numeric operands"},
		{"BoolPlus", ast.Bin("+", ast.Bool(true), ast.Bool(false)), "Arithmetic requires numeric operands"},
	}
	for _, tc := range cases {
		_, err := interp.evaluateExpression(tc.expr, env)
		if err == nil {
			t.Fatalf("%s: expected error", tc.name)
		}
		if err.Error() != tc.msg {
			t.Fatalf("%s: expected %q, got %q", tc.name, tc.msg, err.Error())
		}
	}
}

func TestStringConcatRequiresStrings(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()
	_, err := interp.evaluateExpression(ast.Bin("+", ast.Str("hi"), ast.Int(2)), env)
	if err == nil {
		t.Fatalf("expected concat error")
	}
	if err.Error() != "Arithmetic requires numeric operands" {
		t.Fatalf("expected string concat error, got %q", err.Error())
	}
}

func TestComparisonTypeErrors(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()
	cases := []ast.Expression{
		ast.Bin("<", ast.Str("a"), ast.Int(1)),
		ast.Bin("<", ast.Bool(true), ast.Int(1)),
	}
	for idx, expr := range cases {
		_, err := interp.evaluateExpression(expr, env)
		if err == nil {
			t.Fatalf("case %d: expected error", idx)
		}
		if err.Error() != "Arithmetic requires numeric operands" {
			t.Fatalf("case %d: unexpected error %v", idx, err)
		}
	}
}

func TestLogicalOperandErrors(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()
	cases := []struct {
		name string
		expr ast.Expression
		msg  string
	}{
		{"LogicalAndLeft", ast.Bin("&&", ast.Int(1), ast.Bool(true)), "Logical operands must be bool"},
		{"LogicalAndRight", ast.Bin("&&", ast.Bool(true), ast.Int(1)), "Logical operands must be bool"},
		{"LogicalOrLeft", ast.Bin("||", ast.Int(0), ast.Bool(false)), "Logical operands must be bool"},
		{"LogicalOrRight", ast.Bin("||", ast.Bool(false), ast.Int(1)), "Logical operands must be bool"},
	}
	for _, tc := range cases {
		_, err := interp.evaluateExpression(tc.expr, env)
		if err == nil {
			t.Fatalf("%s: expected error", tc.name)
		}
		if err.Error() != tc.msg {
			t.Fatalf("%s: expected %q, got %q", tc.name, tc.msg, err.Error())
		}
	}
}

func TestRangeBoundariesMustBeNumeric(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()
	_, err := interp.evaluateExpression(ast.Range(ast.Bool(true), ast.Int(5), true), env)
	if err == nil {
		t.Fatalf("expected range error")
	}
	if err.Error() != "Range boundaries must be numeric" {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = interp.evaluateExpression(ast.Range(ast.Int(1), ast.Bool(false), false), env)
	if err == nil {
		t.Fatalf("expected range error for end")
	}
	if err.Error() != "Range boundaries must be numeric" {
		t.Fatalf("unexpected error for end: %v", err)
	}
}
