package interpreter

import (
	"fmt"
	"math"
	"math/big"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
)

const numericEpsilon = 1e-9

func binaryOpForAssignment(op ast.AssignmentOperator) (string, bool) {
	switch op {
	case ast.AssignmentAdd:
		return "+", true
	case ast.AssignmentSub:
		return "-", true
	case ast.AssignmentMul:
		return "*", true
	case ast.AssignmentDiv:
		return "/", true
	case ast.AssignmentMod:
		return "%", true
	case ast.AssignmentBitAnd:
		return "&", true
	case ast.AssignmentBitOr:
		return "|", true
	case ast.AssignmentBitXor:
		return "^", true
	case ast.AssignmentShiftL:
		return "<<", true
	case ast.AssignmentShiftR:
		return ">>", true
	default:
		return "", false
	}
}

func applyBinaryOperator(op string, left runtime.Value, right runtime.Value) (runtime.Value, error) {
	switch op {
	case "+", "-", "*", "/":
		return evaluateArithmetic(op, left, right)
	case "%":
		return evaluateModulo(left, right)
	case "<", "<=", ">", ">=":
		return evaluateComparison(op, left, right)
	case "==":
		return runtime.BoolValue{Val: valuesEqual(left, right)}, nil
	case "!=":
		return runtime.BoolValue{Val: !valuesEqual(left, right)}, nil
	case "&", "|", "^", "<<", ">>":
		return evaluateBitwise(op, left, right)
	default:
		return nil, fmt.Errorf("unsupported binary operator %s", op)
	}
}

func evaluateModulo(left runtime.Value, right runtime.Value) (runtime.Value, error) {
	if !isNumericValue(left) || !isNumericValue(right) {
		return nil, fmt.Errorf("Arithmetic requires numeric operands")
	}
	if lv, ok := left.(runtime.IntegerValue); ok {
		if rv, ok := right.(runtime.IntegerValue); ok {
			if rv.Val == nil || rv.Val.Sign() == 0 {
				return nil, fmt.Errorf("division by zero")
			}
			result := new(big.Int).Rem(runtime.CloneBigInt(lv.Val), rv.Val)
			return runtime.IntegerValue{Val: result, TypeSuffix: lv.TypeSuffix}, nil
		}
	}
	leftFloat, err := numericToFloat(left)
	if err != nil {
		return nil, err
	}
	rightFloat, err := numericToFloat(right)
	if err != nil {
		return nil, err
	}
	if rightFloat == 0 {
		return nil, fmt.Errorf("division by zero")
	}
	return runtime.FloatValue{Val: math.Mod(leftFloat, rightFloat), TypeSuffix: runtime.FloatF64}, nil
}

func evaluateBitwise(op string, left runtime.Value, right runtime.Value) (runtime.Value, error) {
	lv, ok := left.(runtime.IntegerValue)
	if !ok {
		return nil, fmt.Errorf("Bitwise requires i32 operands")
	}
	rv, ok := right.(runtime.IntegerValue)
	if !ok {
		return nil, fmt.Errorf("Bitwise requires i32 operands")
	}
	l, err := int32FromIntegerValue(lv)
	if err != nil {
		return nil, err
	}
	r, err := int32FromIntegerValue(rv)
	if err != nil {
		return nil, err
	}
	var result int32
	switch op {
	case "&":
		result = l & r
	case "|":
		result = l | r
	case "^":
		result = l ^ r
	case "<<":
		if r < 0 || r >= 32 {
			return nil, fmt.Errorf("shift out of range")
		}
		result = l << uint(r)
	case ">>":
		if r < 0 || r >= 32 {
			return nil, fmt.Errorf("shift out of range")
		}
		result = l >> uint(r)
	default:
		return nil, fmt.Errorf("unsupported bitwise operator %s", op)
	}
	return runtime.IntegerValue{Val: big.NewInt(int64(result)), TypeSuffix: runtime.IntegerI32}, nil
}

func int32FromIntegerValue(val runtime.IntegerValue) (int32, error) {
	if val.TypeSuffix != runtime.IntegerI32 {
		return 0, fmt.Errorf("Bitwise requires i32 operands")
	}
	if val.Val == nil || !val.Val.IsInt64() {
		return 0, fmt.Errorf("Bitwise requires i32 operands")
	}
	raw := val.Val.Int64()
	if raw < math.MinInt32 || raw > math.MaxInt32 {
		return 0, fmt.Errorf("Bitwise requires i32 operands")
	}
	return int32(raw), nil
}

func bigFromLiteral(val interface{}) *big.Int {
	switch v := val.(type) {
	case int:
		return big.NewInt(int64(v))
	case int64:
		return big.NewInt(v)
	case float64:
		return big.NewInt(int64(v))
	case string:
		if bi, ok := new(big.Int).SetString(v, 10); ok {
			return bi
		}
		return big.NewInt(0)
	case *big.Int:
		return runtime.CloneBigInt(v)
	default:
		return big.NewInt(0)
	}
}

func evaluateArithmetic(op string, left runtime.Value, right runtime.Value) (runtime.Value, error) {
	if !isNumericValue(left) || !isNumericValue(right) {
		return nil, fmt.Errorf("Arithmetic requires numeric operands")
	}
	if lv, ok := left.(runtime.IntegerValue); ok {
		if rv, ok := right.(runtime.IntegerValue); ok {
			result := new(big.Int)
			switch op {
			case "+":
				result.Add(lv.Val, rv.Val)
			case "-":
				result.Sub(lv.Val, rv.Val)
			case "*":
				result.Mul(lv.Val, rv.Val)
			case "/":
				if rv.Val.Sign() == 0 {
					return nil, fmt.Errorf("division by zero")
				}
				result.Quo(lv.Val, rv.Val)
			default:
				return nil, fmt.Errorf("unsupported arithmetic operator %s", op)
			}
			return runtime.IntegerValue{Val: result, TypeSuffix: lv.TypeSuffix}, nil
		}
	}
	leftFloat, err := numericToFloat(left)
	if err != nil {
		return nil, err
	}
	rightFloat, err := numericToFloat(right)
	if err != nil {
		return nil, err
	}
	var val float64
	switch op {
	case "+":
		val = leftFloat + rightFloat
	case "-":
		val = leftFloat - rightFloat
	case "*":
		val = leftFloat * rightFloat
	case "/":
		if rightFloat == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		val = leftFloat / rightFloat
	default:
		return nil, fmt.Errorf("unsupported arithmetic operator %s", op)
	}
	return runtime.FloatValue{Val: val, TypeSuffix: runtime.FloatF64}, nil
}

func valuesEqual(left runtime.Value, right runtime.Value) bool {
	if isNumericValue(left) && isNumericValue(right) {
		lf, err := numericToFloat(left)
		if err != nil {
			return false
		}
		rf, err := numericToFloat(right)
		if err != nil {
			return false
		}
		return math.Abs(lf-rf) < numericEpsilon
	}
	switch lv := left.(type) {
	case runtime.StringValue:
		if rv, ok := right.(runtime.StringValue); ok {
			return lv.Val == rv.Val
		}
	case runtime.BoolValue:
		if rv, ok := right.(runtime.BoolValue); ok {
			return lv.Val == rv.Val
		}
	case runtime.CharValue:
		if rv, ok := right.(runtime.CharValue); ok {
			return lv.Val == rv.Val
		}
	case runtime.NilValue:
		_, ok := right.(runtime.NilValue)
		return ok
	case runtime.IntegerValue:
		if rv, ok := right.(runtime.IntegerValue); ok {
			return lv.Val.Cmp(rv.Val) == 0
		}
	case runtime.FloatValue:
		if rv, ok := right.(runtime.FloatValue); ok {
			return math.Abs(lv.Val-rv.Val) < numericEpsilon
		}
	}
	return false
}

func evaluateComparison(op string, left runtime.Value, right runtime.Value) (runtime.Value, error) {
	if !isNumericValue(left) || !isNumericValue(right) {
		return nil, fmt.Errorf("Arithmetic requires numeric operands")
	}
	leftFloat, err := numericToFloat(left)
	if err != nil {
		return nil, err
	}
	rightFloat, err := numericToFloat(right)
	if err != nil {
		return nil, err
	}
	cmp := 0
	diff := leftFloat - rightFloat
	if math.Abs(diff) < numericEpsilon {
		cmp = 0
	} else if diff < 0 {
		cmp = -1
	} else {
		cmp = 1
	}
	return runtime.BoolValue{Val: comparisonOp(op, cmp)}, nil
}
