package interpreter

import (
	"math"
	"math/big"
	"strings"
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
			expr: ast.Bin(".<<", ast.Int(1), ast.Int(-1)),
			msg:  "shift out of range",
		},
		{
			name: "LargeShift",
			expr: ast.Bin(".>>", ast.Int(1), ast.Int(32)),
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

func TestBitwiseRequiresIntegerOperands(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()
	cases := []ast.Expression{
		ast.Bin(".&", ast.Int(1), ast.Flt(1.0)),
		ast.Bin(".|", ast.Flt(1.0), ast.Int(1)),
		ast.Bin(".^", ast.Flt(1.0), ast.Flt(1.0)),
	}
	for idx, expr := range cases {
		_, err := interp.evaluateExpression(expr, env)
		if err == nil {
			t.Fatalf("case %d: expected error", idx)
		}
		if err.Error() != "Bitwise requires integer operands" {
			t.Fatalf("case %d: expected 'Bitwise requires integer operands', got %q", idx, err.Error())
		}
	}
}

func TestDivisionByZeroDiagnostics(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()
	cases := []ast.Expression{
		ast.Bin("/", ast.Int(4), ast.Int(0)),
		ast.Bin("//", ast.Int(4), ast.Int(0)),
		ast.Bin("%%", ast.Int(4), ast.Int(0)),
		ast.Bin("/%", ast.Int(4), ast.Int(0)),
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
		{"DivIntsPromoteToFloat", ast.Bin("/", ast.Int(5), ast.Int(2)), 2.5},
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

func TestDivModRequiresIntegerOperands(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()
	cases := []struct {
		name   string
		expr   ast.Expression
	}{
		{"QuotientWithFloat", ast.Bin("//", ast.Int(5), ast.Flt(2))},
		{"RemainderWithFloat", ast.Bin("%%", ast.Flt(5), ast.Int(2))},
		{"DivModStructWithFloat", ast.Bin("/%", ast.Flt(5), ast.Int(2))},
	}
	for _, tc := range cases {
		val, err := interp.evaluateExpression(tc.expr, env)
		if val != nil {
			t.Fatalf("%s: expected no value", tc.name)
		}
		if err == nil {
			t.Fatalf("%s: expected error", tc.name)
		}
		if !strings.Contains(err.Error(), "requires integer operands") {
			t.Fatalf("%s: expected integer operand diagnostic, got %q", tc.name, err.Error())
		}
	}
}

func TestDivModEuclidean(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()

	quot, err := interp.evaluateExpression(ast.Bin("//", ast.Int(-5), ast.Int(3)), env)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	qVal, ok := quot.(runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected integer quotient, got %#v", quot)
	}
	if qVal.Val.Cmp(big.NewInt(-2)) != 0 {
		t.Fatalf("expected quotient -2, got %v", qVal.Val)
	}

	rem, err := interp.evaluateExpression(ast.Bin("%%", ast.Int(-5), ast.Int(3)), env)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	rVal, ok := rem.(runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected integer remainder, got %#v", rem)
	}
	if rVal.Val.Cmp(big.NewInt(1)) != 0 {
		t.Fatalf("expected remainder 1, got %v", rVal.Val)
	}
}

func TestDivModStructResult(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()
	val, err := interp.evaluateExpression(ast.Bin("/%", ast.Int(7), ast.Int(3)), env)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	inst, ok := val.(*runtime.StructInstanceValue)
	if !ok {
		t.Fatalf("expected struct instance, got %#v", val)
	}
	if inst.Definition == nil || inst.Definition.Node == nil || inst.Definition.Node.ID == nil {
		t.Fatalf("expected DivMod struct definition")
	}
	if inst.Definition.Node.ID.Name != "DivMod" {
		t.Fatalf("expected DivMod struct, got %s", inst.Definition.Node.ID.Name)
	}
	if len(inst.TypeArguments) != 1 {
		t.Fatalf("expected one type argument, got %d", len(inst.TypeArguments))
	}
	if simple, ok := inst.TypeArguments[0].(*ast.SimpleTypeExpression); !ok || simple.Name.Name != "i32" {
		t.Fatalf("expected type argument i32, got %#v", inst.TypeArguments[0])
	}
	quotVal, ok := inst.Fields["quotient"].(runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected integer quotient field, got %#v", inst.Fields["quotient"])
	}
	if quotVal.Val.Cmp(big.NewInt(2)) != 0 {
		t.Fatalf("expected quotient 2, got %v", quotVal.Val)
	}
	remVal, ok := inst.Fields["remainder"].(runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected integer remainder field, got %#v", inst.Fields["remainder"])
	}
	if remVal.Val.Cmp(big.NewInt(1)) != 0 {
		t.Fatalf("expected remainder 1, got %v", remVal.Val)
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

func TestIntegerLiteralSuffixPreserved(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()
	i64 := ast.IntegerTypeI64
	literal := ast.IntBig(big.NewInt(9_007_199_254_740_993), &i64)
	val, err := interp.evaluateExpression(literal, env)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	intVal, ok := val.(runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected integer value, got %#v", val)
	}
	if intVal.TypeSuffix != runtime.IntegerI64 {
		t.Fatalf("expected type suffix i64, got %s", intVal.TypeSuffix)
	}
	expected := big.NewInt(0).SetInt64(0)
	expected.SetString("9007199254740993", 10)
	if intVal.Val.Cmp(expected) != 0 {
		t.Fatalf("expected %v, got %v", expected, intVal.Val)
	}
}

func TestIntegerArithmeticPromotion(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()
	i16 := ast.IntegerTypeI16
	u16 := ast.IntegerTypeU16
	expr := ast.Bin("+", ast.IntTyped(1, &i16), ast.IntTyped(2, &u16))
	val, err := interp.evaluateExpression(expr, env)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	intVal, ok := val.(runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected integer result, got %#v", val)
	}
	if intVal.TypeSuffix != runtime.IntegerI32 {
		t.Fatalf("expected promotion to i32, got %s", intVal.TypeSuffix)
	}
	if intVal.Val.Cmp(big.NewInt(3)) != 0 {
		t.Fatalf("expected value 3, got %v", intVal.Val)
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

func TestBitshiftRangeChecks(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()
	cases := []struct {
		name string
		expr ast.Expression
	}{
		{"ShiftLeftOutOfRange", ast.Bin(".<<", ast.Int(1), ast.Int(32))},
		{"ShiftRightOutOfRange", ast.Bin(".>>", ast.Int(1), ast.Int(33))},
	}
	for _, tc := range cases {
		if _, err := interp.evaluateExpression(tc.expr, env); err == nil {
			t.Fatalf("%s: expected shift out of range error", tc.name)
		} else if err.Error() != "shift out of range" {
			t.Fatalf("%s: unexpected error %v", tc.name, err)
		}
	}

	errModule := ast.Mod([]ast.Statement{
		ast.Assign(ast.ID("x"), ast.Int(1)),
		ast.AssignOp(ast.AssignmentShiftL, ast.ID("x"), ast.Int(32)),
	}, nil, nil)
	if _, _, err := interp.EvaluateModule(errModule); err == nil {
		t.Fatalf("expected compound shift to fail")
	} else if err.Error() != "shift out of range" {
		t.Fatalf("unexpected error from compound shift: %v", err)
	}

	okModule := ast.Mod([]ast.Statement{
		ast.Assign(ast.ID("x"), ast.Int(1)),
		ast.AssignOp(ast.AssignmentShiftL, ast.ID("x"), ast.Int(3)),
		ast.ID("x"),
	}, nil, nil)
	// Run on a fresh interpreter to avoid clashing with previous declarations.
	interp = New()
	result, _, err := interp.EvaluateModule(okModule)
	if err != nil {
		t.Fatalf("compound shift module failed: %v", err)
	}
	intVal, ok := result.(runtime.IntegerValue)
	if !ok || intVal.Val.Cmp(bigInt(8)) != 0 {
		t.Fatalf("expected 8 after shift, got %#v", result)
	}
}

func TestMixedNumericOperations(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()

	sum := ast.Bin("+", ast.Int(2), ast.Flt(1.5))
	val, err := interp.evaluateExpression(sum, env)
	if err != nil {
		t.Fatalf("mixed arithmetic failed: %v", err)
	}
	flt, ok := val.(runtime.FloatValue)
	if !ok {
		t.Fatalf("expected float result, got %#v", val)
	}
	if math.Abs(flt.Val-3.5) > 1e-9 {
		t.Fatalf("expected 3.5, got %v", flt.Val)
	}

	eq := ast.Bin("==", ast.Int(3), ast.Flt(3.0))
	eqVal, err := interp.evaluateExpression(eq, env)
	if err != nil {
		t.Fatalf("mixed equality failed: %v", err)
	}
	if b, ok := eqVal.(runtime.BoolValue); !ok || !b.Val {
		t.Fatalf("expected equality to hold, got %#v", eqVal)
	}

	lt := ast.Bin("<", ast.Int(2), ast.Flt(3.5))
	ltVal, err := interp.evaluateExpression(lt, env)
	if err != nil {
		t.Fatalf("mixed comparison failed: %v", err)
	}
	if b, ok := ltVal.(runtime.BoolValue); !ok || !b.Val {
		t.Fatalf("expected comparison to be true, got %#v", ltVal)
	}
}
