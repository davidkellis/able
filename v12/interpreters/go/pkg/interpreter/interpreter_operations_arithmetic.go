package interpreter

import (
	"fmt"
	"math"
	"math/big"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func applyBinaryOperator(i *Interpreter, op string, left runtime.Value, right runtime.Value) (runtime.Value, error) {
	rawOp := op
	op, dotted := normalizeOperator(op)
	rawLeft := unwrapInterfaceValue(left)
	rawRight := unwrapInterfaceValue(right)
	switch op {
	case "+", "-", "*", "^":
		if op == "^" && dotted {
			if isIntegerValue(rawLeft) && isIntegerValue(rawRight) {
				return evaluateBitwise(op, rawLeft, rawRight)
			}
			if result, ok, err := i.applyOperatorInterface(rawOp, left, right); ok {
				return result, err
			}
			return evaluateBitwise(op, rawLeft, rawRight)
		}
		if isNumericValue(rawLeft) && isNumericValue(rawRight) {
			return evaluateArithmetic(i, op, rawLeft, rawRight)
		}
		if op != "^" {
			if result, ok, err := i.applyOperatorInterface(rawOp, left, right); ok {
				return result, err
			}
		}
		return evaluateArithmetic(i, op, rawLeft, rawRight)
	case "/":
		if isNumericValue(rawLeft) && isNumericValue(rawRight) {
			return evaluateDivision(i, rawLeft, rawRight)
		}
		if result, ok, err := i.applyOperatorInterface(rawOp, left, right); ok {
			return result, err
		}
		return evaluateDivision(i, rawLeft, rawRight)
	case "//", "%", "/%":
		if op == "%" {
			if isIntegerValue(rawLeft) && isIntegerValue(rawRight) {
				return evaluateDivMod(i, op, rawLeft, rawRight)
			}
			if result, ok, err := i.applyOperatorInterface(rawOp, left, right); ok {
				return result, err
			}
			return evaluateDivMod(i, op, rawLeft, rawRight)
		}
		return evaluateDivMod(i, op, rawLeft, rawRight)
	case "<", "<=", ">", ">=":
		if isNumericValue(rawLeft) && isNumericValue(rawRight) {
			return evaluateComparison(op, rawLeft, rawRight)
		}
		if _, ok := stringFromValue(rawLeft); ok {
			if _, ok := stringFromValue(rawRight); ok {
				return evaluateComparison(op, rawLeft, rawRight)
			}
		}
		if result, ok, err := i.applyOrderingInterface(op, left, right); ok {
			return result, err
		}
		return evaluateComparison(op, rawLeft, rawRight)
	case "==":
		if isNumericValue(rawLeft) && isNumericValue(rawRight) {
			return evaluateComparison(op, rawLeft, rawRight)
		}
		if ls, ok := stringFromValue(rawLeft); ok {
			if rs, ok := stringFromValue(rawRight); ok {
				return runtime.BoolValue{Val: ls == rs}, nil
			}
		}
		if _, ok := rawLeft.(runtime.NilValue); ok {
			return runtime.BoolValue{Val: valuesEqual(rawLeft, rawRight)}, nil
		}
		if _, ok := rawRight.(runtime.NilValue); ok {
			return runtime.BoolValue{Val: valuesEqual(rawLeft, rawRight)}, nil
		}
		if _, ok := rawLeft.(*runtime.NilValue); ok {
			return runtime.BoolValue{Val: valuesEqual(rawLeft, rawRight)}, nil
		}
		if _, ok := rawRight.(*runtime.NilValue); ok {
			return runtime.BoolValue{Val: valuesEqual(rawLeft, rawRight)}, nil
		}
		if result, ok, err := i.applyEqualityInterface(op, left, right); ok {
			return result, err
		}
		return runtime.BoolValue{Val: valuesEqual(rawLeft, rawRight)}, nil
	case "!=":
		if isNumericValue(rawLeft) && isNumericValue(rawRight) {
			return evaluateComparison(op, rawLeft, rawRight)
		}
		if ls, ok := stringFromValue(rawLeft); ok {
			if rs, ok := stringFromValue(rawRight); ok {
				return runtime.BoolValue{Val: ls != rs}, nil
			}
		}
		if _, ok := rawLeft.(runtime.NilValue); ok {
			return runtime.BoolValue{Val: !valuesEqual(rawLeft, rawRight)}, nil
		}
		if _, ok := rawRight.(runtime.NilValue); ok {
			return runtime.BoolValue{Val: !valuesEqual(rawLeft, rawRight)}, nil
		}
		if _, ok := rawLeft.(*runtime.NilValue); ok {
			return runtime.BoolValue{Val: !valuesEqual(rawLeft, rawRight)}, nil
		}
		if _, ok := rawRight.(*runtime.NilValue); ok {
			return runtime.BoolValue{Val: !valuesEqual(rawLeft, rawRight)}, nil
		}
		if result, ok, err := i.applyEqualityInterface(op, left, right); ok {
			return result, err
		}
		return runtime.BoolValue{Val: !valuesEqual(rawLeft, rawRight)}, nil
	case "&", "|", "<<", ">>":
		if isIntegerValue(rawLeft) && isIntegerValue(rawRight) {
			return evaluateBitwise(op, rawLeft, rawRight)
		}
		if result, ok, err := i.applyOperatorInterface(rawOp, left, right); ok {
			return result, err
		}
		return evaluateBitwise(op, rawLeft, rawRight)
	default:
		return nil, fmt.Errorf("unsupported binary operator %s", op)
	}
}

// ApplyBinaryOperator exposes binary operator dispatch for compiled/runtime interop.
func (i *Interpreter) ApplyBinaryOperator(op string, left runtime.Value, right runtime.Value) (runtime.Value, error) {
	return applyBinaryOperator(i, op, left, right)
}

func evaluateDivision(i *Interpreter, left runtime.Value, right runtime.Value) (runtime.Value, error) {
	if isRatioValue(left) || isRatioValue(right) {
		leftRatio, err := coerceToRatio(left)
		if err != nil {
			return nil, err
		}
		rightRatio, err := coerceToRatio(right)
		if err != nil {
			return nil, err
		}
		num := new(big.Int).Mul(leftRatio.num, rightRatio.den)
		den := new(big.Int).Mul(leftRatio.den, rightRatio.num)
		normalized, err := normalizeRatioParts(num, den)
		if err != nil {
			return nil, err
		}
		return i.makeRatioValue(normalized)
	}
	if !isNumericValue(left) || !isNumericValue(right) {
		return nil, fmt.Errorf("Arithmetic requires numeric operands")
	}
	_, leftIsFloat := left.(runtime.FloatValue)
	_, rightIsFloat := right.(runtime.FloatValue)
	if leftIsFloat || rightIsFloat {
		targetFloatKind := floatResultKind(left, right)
		leftFloat, err := numericToFloat(left)
		if err != nil {
			return nil, err
		}
		rightFloat, err := numericToFloat(right)
		if err != nil {
			return nil, err
		}
		val := normalizeFloat(targetFloatKind, leftFloat/rightFloat)
		return runtime.FloatValue{Val: val, TypeSuffix: targetFloatKind}, nil
	}
	leftInt, ok := left.(runtime.IntegerValue)
	if !ok {
		return nil, fmt.Errorf("Arithmetic requires numeric operands")
	}
	rightInt, ok := right.(runtime.IntegerValue)
	if !ok {
		return nil, fmt.Errorf("Arithmetic requires numeric operands")
	}
	if rightInt.Val == nil || rightInt.Val.Sign() == 0 {
		return nil, newDivisionByZeroError()
	}
	leftFloat := bigIntToFloat(leftInt.Val)
	rightFloat := bigIntToFloat(rightInt.Val)
	if rightFloat == 0 {
		return nil, newDivisionByZeroError()
	}
	val := normalizeFloat(runtime.FloatF64, leftFloat/rightFloat)
	return runtime.FloatValue{Val: val, TypeSuffix: runtime.FloatF64}, nil
}

func evaluateDivMod(i *Interpreter, op string, left runtime.Value, right runtime.Value) (runtime.Value, error) {
	lv, ok := left.(runtime.IntegerValue)
	if !ok {
		return nil, fmt.Errorf("Arithmetic requires integer operands")
	}
	rv, ok := right.(runtime.IntegerValue)
	if !ok {
		return nil, fmt.Errorf("Arithmetic requires integer operands")
	}
	quotient, remainder, targetType, err := computeDivMod(lv, rv)
	if err != nil {
		return nil, err
	}
	switch op {
	case "//":
		return quotient, nil
	case "%":
		return remainder, nil
	case "/%":
		return i.makeDivModResult(targetType, quotient, remainder)
	default:
		return nil, fmt.Errorf("unsupported div/mod operator %s", op)
	}
}

func evaluateBitwise(op string, left runtime.Value, right runtime.Value) (runtime.Value, error) {
	lv, ok := left.(runtime.IntegerValue)
	if !ok {
		return nil, fmt.Errorf("Bitwise requires integer operands")
	}
	rv, ok := right.(runtime.IntegerValue)
	if !ok {
		return nil, fmt.Errorf("Bitwise requires integer operands")
	}
	lVal := runtime.CloneBigInt(lv.Val)
	rVal := runtime.CloneBigInt(rv.Val)
	var result *big.Int
	switch op {
	case "<<", ">>":
		info, err := getIntegerInfo(lv.TypeSuffix)
		if err != nil {
			return nil, err
		}
		if !rVal.IsInt64() {
			return nil, newShiftOutOfRangeError(0)
		}
		count := int(rVal.Int64())
		var shifted *big.Int
		if op == "<<" {
			shifted, err = shiftValueLeft(lVal, count, info)
		} else {
			shifted, err = shiftValueRight(lVal, count, info)
		}
		if err != nil {
			return nil, err
		}
		result = shifted
		if err := ensureFitsInteger(info, result); err != nil {
			return nil, err
		}
		return runtime.IntegerValue{Val: result, TypeSuffix: lv.TypeSuffix}, nil
	}
	targetType, err := promoteIntegerTypes(lv.TypeSuffix, rv.TypeSuffix)
	if err != nil {
		return nil, err
	}
	info, err := getIntegerInfo(targetType)
	if err != nil {
		return nil, err
	}
	switch op {
	case "&":
		leftPattern := bitPattern(lVal, info)
		rightPattern := bitPattern(rVal, info)
		tmp := new(big.Int).And(leftPattern, rightPattern)
		result = patternToInteger(tmp, info)
	case "|":
		leftPattern := bitPattern(lVal, info)
		rightPattern := bitPattern(rVal, info)
		tmp := new(big.Int).Or(leftPattern, rightPattern)
		result = patternToInteger(tmp, info)
	case "^":
		leftPattern := bitPattern(lVal, info)
		rightPattern := bitPattern(rVal, info)
		tmp := new(big.Int).Xor(leftPattern, rightPattern)
		result = patternToInteger(tmp, info)
	default:
		return nil, fmt.Errorf("unsupported bitwise operator %s", op)
	}
	if err := ensureFitsInteger(info, result); err != nil {
		return nil, err
	}
	return runtime.IntegerValue{Val: result, TypeSuffix: targetType}, nil
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

func evaluateArithmetic(i *Interpreter, op string, left runtime.Value, right runtime.Value) (runtime.Value, error) {
	if isRatioValue(left) || isRatioValue(right) {
		switch op {
		case "+", "-", "*", "/":
			leftRatio, err := coerceToRatio(left)
			if err != nil {
				return nil, err
			}
			rightRatio, err := coerceToRatio(right)
			if err != nil {
				return nil, err
			}
			var num *big.Int
			var den *big.Int
			switch op {
			case "+":
				num = new(big.Int).Add(new(big.Int).Mul(leftRatio.num, rightRatio.den), new(big.Int).Mul(rightRatio.num, leftRatio.den))
				den = new(big.Int).Mul(leftRatio.den, rightRatio.den)
			case "-":
				num = new(big.Int).Sub(new(big.Int).Mul(leftRatio.num, rightRatio.den), new(big.Int).Mul(rightRatio.num, leftRatio.den))
				den = new(big.Int).Mul(leftRatio.den, rightRatio.den)
			case "*":
				num = new(big.Int).Mul(leftRatio.num, rightRatio.num)
				den = new(big.Int).Mul(leftRatio.den, rightRatio.den)
			case "/":
				num = new(big.Int).Mul(leftRatio.num, rightRatio.den)
				den = new(big.Int).Mul(leftRatio.den, rightRatio.num)
			}
			normalized, err := normalizeRatioParts(num, den)
			if err != nil {
				return nil, err
			}
			return i.makeRatioValue(normalized)
		default:
			return nil, fmt.Errorf("unsupported arithmetic operator %s", op)
		}
	}

	leftInt, leftIsInt := left.(runtime.IntegerValue)
	rightInt, rightIsInt := right.(runtime.IntegerValue)
	if leftIsInt && rightIsInt {
		targetType, err := promoteIntegerTypes(leftInt.TypeSuffix, rightInt.TypeSuffix)
		if err != nil {
			return nil, err
		}
		info, err := getIntegerInfo(targetType)
		if err != nil {
			return nil, err
		}
		lv := runtime.CloneBigInt(leftInt.Val)
		rv := runtime.CloneBigInt(rightInt.Val)
		result := new(big.Int)
		switch op {
		case "+":
			result.Add(lv, rv)
		case "-":
			result.Sub(lv, rv)
		case "*":
			result.Mul(lv, rv)
		case "^":
			if rv.Sign() < 0 {
				return nil, fmt.Errorf("Negative integer exponent is not supported")
			}
			result.Exp(lv, rv, nil)
		default:
			return nil, fmt.Errorf("unsupported arithmetic operator %s", op)
		}
		if err := ensureFitsInteger(info, result); err != nil {
			return nil, err
		}
		return runtime.IntegerValue{Val: result, TypeSuffix: targetType}, nil
	}
	if !isNumericValue(left) || !isNumericValue(right) {
		return nil, fmt.Errorf("Arithmetic requires numeric operands")
	}
	targetFloatKind := floatResultKind(left, right)
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
	case "^":
		val = math.Pow(leftFloat, rightFloat)
	default:
		return nil, fmt.Errorf("unsupported arithmetic operator %s", op)
	}
	val = normalizeFloat(targetFloatKind, val)
	return runtime.FloatValue{Val: val, TypeSuffix: targetFloatKind}, nil
}

func computeDivMod(left runtime.IntegerValue, right runtime.IntegerValue) (runtime.IntegerValue, runtime.IntegerValue, runtime.IntegerType, error) {
	if right.Val == nil || right.Val.Sign() == 0 {
		return runtime.IntegerValue{}, runtime.IntegerValue{}, runtime.IntegerI32, newDivisionByZeroError()
	}
	targetType, err := promoteIntegerTypes(left.TypeSuffix, right.TypeSuffix)
	if err != nil {
		return runtime.IntegerValue{}, runtime.IntegerValue{}, runtime.IntegerI32, err
	}
	info, err := getIntegerInfo(targetType)
	if err != nil {
		return runtime.IntegerValue{}, runtime.IntegerValue{}, runtime.IntegerI32, err
	}
	dividend := runtime.CloneBigInt(left.Val)
	divisor := runtime.CloneBigInt(right.Val)
	quotient, remainder, err := euclideanDivModBig(dividend, divisor)
	if err != nil {
		return runtime.IntegerValue{}, runtime.IntegerValue{}, runtime.IntegerI32, err
	}
	if err := ensureFitsInteger(info, quotient); err != nil {
		return runtime.IntegerValue{}, runtime.IntegerValue{}, runtime.IntegerI32, err
	}
	if err := ensureFitsInteger(info, remainder); err != nil {
		return runtime.IntegerValue{}, runtime.IntegerValue{}, runtime.IntegerI32, err
	}
	return runtime.IntegerValue{Val: quotient, TypeSuffix: targetType}, runtime.IntegerValue{Val: remainder, TypeSuffix: targetType}, targetType, nil
}

func euclideanDivModBig(dividend *big.Int, divisor *big.Int) (*big.Int, *big.Int, error) {
	if divisor == nil || divisor.Sign() == 0 {
		return nil, nil, newDivisionByZeroError()
	}
	quotient := new(big.Int).Quo(dividend, divisor)
	remainder := new(big.Int).Rem(dividend, divisor)
	if remainder.Sign() < 0 {
		if divisor.Sign() > 0 {
			quotient.Sub(quotient, big.NewInt(1))
			remainder.Add(remainder, divisor)
		} else {
			quotient.Add(quotient, big.NewInt(1))
			remainder.Sub(remainder, divisor)
		}
	}
	return quotient, remainder, nil
}

func (i *Interpreter) makeDivModResult(kind runtime.IntegerType, quotient runtime.IntegerValue, remainder runtime.IntegerValue) (runtime.Value, error) {
	def, err := i.ensureDivModStruct()
	if err != nil {
		return nil, err
	}
	fields := map[string]runtime.Value{
		"quotient":  quotient,
		"remainder": remainder,
	}
	typeArg := ast.NewSimpleTypeExpression(ast.NewIdentifier(string(kind)))
	return &runtime.StructInstanceValue{
		Definition:    def,
		Fields:        fields,
		TypeArguments: []ast.TypeExpression{typeArg},
	}, nil
}

func (i *Interpreter) ensureDivModStruct() (*runtime.StructDefinitionValue, error) {
	if i.divModStruct != nil {
		return i.divModStruct, nil
	}
	if val, err := i.global.Get("DivMod"); err == nil {
		if def, conv := toStructDefinitionValue(val, "DivMod"); conv == nil {
			i.divModStruct = def
			return def, nil
		}
	}
	typeParam := ast.NewGenericParameter(ast.NewIdentifier("T"), nil)
	quotientField := ast.NewStructFieldDefinition(ast.NewSimpleTypeExpression(ast.NewIdentifier("T")), ast.NewIdentifier("quotient"))
	remainderField := ast.NewStructFieldDefinition(ast.NewSimpleTypeExpression(ast.NewIdentifier("T")), ast.NewIdentifier("remainder"))
	definition := ast.NewStructDefinition(
		ast.NewIdentifier("DivMod"),
		[]*ast.StructFieldDefinition{quotientField, remainderField},
		ast.StructKindNamed,
		[]*ast.GenericParameter{typeParam},
		nil,
		false,
	)
	if _, err := i.evaluateStructDefinition(definition, i.global); err != nil {
		return nil, err
	}
	val, err := i.global.Get("DivMod")
	if err != nil {
		return nil, err
	}
	structDef, conv := toStructDefinitionValue(val, "DivMod")
	if conv != nil {
		return nil, conv
	}
	i.divModStruct = structDef
	return structDef, nil
}
