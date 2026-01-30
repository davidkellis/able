package interpreter

import (
	"math"
	"math/big"
	"strings"
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestCastNumericSemantics(t *testing.T) {
	u16 := ast.IntegerTypeU16
	i16 := ast.IntegerTypeI16
	i32 := ast.IntegerTypeI32

	wrapExpr := ast.NewTypeCastExpression(ast.IntTyped(256, &u16), ast.Ty("u8"))
	treeWrap, byteWrap := evalCastBoth(t, wrapExpr)
	assertIntValue(t, treeWrap, runtime.IntegerU8, 0)
	assertIntValue(t, byteWrap, runtime.IntegerU8, 0)

	negWrapExpr := ast.NewTypeCastExpression(ast.IntTyped(-1, &i16), ast.Ty("u8"))
	treeNegWrap, byteNegWrap := evalCastBoth(t, negWrapExpr)
	assertIntValue(t, treeNegWrap, runtime.IntegerU8, 255)
	assertIntValue(t, byteNegWrap, runtime.IntegerU8, 255)

	intToFloatExpr := ast.NewTypeCastExpression(ast.IntTyped(42, &i32), ast.Ty("f64"))
	treeIntFloat, byteIntFloat := evalCastBoth(t, intToFloatExpr)
	assertFloatValue(t, treeIntFloat, runtime.FloatF64, 42.0)
	assertFloatValue(t, byteIntFloat, runtime.FloatF64, 42.0)

	floatToFloatExpr := ast.NewTypeCastExpression(ast.Flt(1.5), ast.Ty("f32"))
	treeFloatFloat, byteFloatFloat := evalCastBoth(t, floatToFloatExpr)
	assertFloatValue(t, treeFloatFloat, runtime.FloatF32, 1.5)
	assertFloatValue(t, byteFloatFloat, runtime.FloatF32, 1.5)

	floatToIntExpr := ast.NewTypeCastExpression(ast.Flt(3.9), ast.Ty("i32"))
	treeFloatInt, byteFloatInt := evalCastBoth(t, floatToIntExpr)
	assertIntValue(t, treeFloatInt, runtime.IntegerI32, 3)
	assertIntValue(t, byteFloatInt, runtime.IntegerI32, 3)
}

func TestCastErrors(t *testing.T) {
	invalid := ast.Mod([]ast.Statement{
		ast.NewTypeCastExpression(ast.Str("nope"), ast.Ty("i32")),
	}, nil, nil)
	assertCastErrorContains(t, invalid, "cannot cast String to i32")

	outOfRange := ast.Mod([]ast.Statement{
		ast.NewTypeCastExpression(ast.Flt(1e40), ast.Ty("i32")),
	}, nil, nil)
	assertCastErrorContains(t, outOfRange, "integer overflow")
}

func evalCastBoth(t *testing.T, expr ast.Expression) (runtime.Value, runtime.Value) {
	module := ast.Mod([]ast.Statement{expr}, nil, nil)
	tree := mustEvalModule(t, New(), module)
	byte := runBytecodeModule(t, module)
	return tree, byte
}

func assertCastErrorContains(t *testing.T, module *ast.Module, substr string) {
	t.Helper()
	treeErr := evalModuleError(t, New(), module)
	if treeErr == nil || !strings.Contains(treeErr.Error(), substr) {
		t.Fatalf("expected tree cast error containing %q, got %v", substr, treeErr)
	}
	byteErr := runBytecodeModuleError(t, NewBytecode(), module)
	if byteErr == nil || !strings.Contains(byteErr.Error(), substr) {
		t.Fatalf("expected bytecode cast error containing %q, got %v", substr, byteErr)
	}
}

func evalModuleError(t *testing.T, interp *Interpreter, module *ast.Module) error {
	t.Helper()
	_, _, err := interp.EvaluateModule(module)
	return err
}

func runBytecodeModuleError(t *testing.T, interp *Interpreter, module *ast.Module) error {
	t.Helper()
	program, err := interp.lowerModuleToBytecode(module)
	if err != nil {
		return err
	}
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	_, err = vm.run(program)
	return err
}

func assertIntValue(t *testing.T, value runtime.Value, kind runtime.IntegerType, want int64) {
	t.Helper()
	var iv runtime.IntegerValue
	switch v := value.(type) {
	case runtime.IntegerValue:
		iv = v
	case *runtime.IntegerValue:
		if v == nil {
			t.Fatalf("expected integer value, got nil")
		}
		iv = *v
	default:
		t.Fatalf("expected integer value, got %T", value)
	}
	if iv.TypeSuffix != kind {
		t.Fatalf("expected integer type %s, got %s", kind, iv.TypeSuffix)
	}
	if iv.Val == nil || iv.Val.Cmp(big.NewInt(want)) != 0 {
		t.Fatalf("expected integer %d, got %v", want, iv.Val)
	}
}

func assertFloatValue(t *testing.T, value runtime.Value, kind runtime.FloatType, want float64) {
	t.Helper()
	var fv runtime.FloatValue
	switch v := value.(type) {
	case runtime.FloatValue:
		fv = v
	case *runtime.FloatValue:
		if v == nil {
			t.Fatalf("expected float value, got nil")
		}
		fv = *v
	default:
		t.Fatalf("expected float value, got %T", value)
	}
	if fv.TypeSuffix != kind {
		t.Fatalf("expected float type %s, got %s", kind, fv.TypeSuffix)
	}
	if math.Abs(fv.Val-want) > 1e-9 {
		t.Fatalf("expected float %v, got %v", want, fv.Val)
	}
}
