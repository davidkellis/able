package interpreter

import (
	"fmt"
	"math"
	"math/big"
	"strings"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
)

const numericEpsilon = 1e-9

var (
	bigI64Min = big.NewInt(math.MinInt64)
	bigI64Max = big.NewInt(math.MaxInt64)
)

type ratioParts struct {
	num *big.Int
	den *big.Int
}

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
		return ".&", true
	case ast.AssignmentBitOr:
		return ".|", true
	case ast.AssignmentBitXor:
		return ".^", true
	case ast.AssignmentShiftL:
		return ".<<", true
	case ast.AssignmentShiftR:
		return ".>>", true
	default:
		return "", false
	}
}

func normalizeOperator(op string) (string, bool) {
	switch op {
	case ".&":
		return "&", true
	case ".|":
		return "|", true
	case ".^":
		return "^", true
	case ".<<":
		return "<<", true
	case ".>>":
		return ">>", true
	case ".~":
		return "~", true
	case "\\xor":
		return "^", false
	default:
		return op, false
	}
}

func isRatioValue(val runtime.Value) bool {
	_, ok := ratioPartsFromStruct(val)
	return ok
}

func ratioPartsFromStruct(val runtime.Value) (ratioParts, bool) {
	inst, ok := val.(*runtime.StructInstanceValue)
	if !ok || inst == nil || inst.Definition == nil {
		return ratioParts{}, false
	}
	if structInstanceName(inst) != "Ratio" {
		return ratioParts{}, false
	}
	numVal, numOK := inst.Fields["num"].(runtime.IntegerValue)
	denVal, denOK := inst.Fields["den"].(runtime.IntegerValue)
	if !numOK || !denOK {
		return ratioParts{}, false
	}
	return ratioParts{
		num: runtime.CloneBigInt(numVal.Val),
		den: runtime.CloneBigInt(denVal.Val),
	}, true
}

func normalizeRatioParts(num *big.Int, den *big.Int) (ratioParts, error) {
	if den == nil || den.Sign() == 0 {
		return ratioParts{}, fmt.Errorf("division by zero")
	}
	n := runtime.CloneBigInt(num)
	d := runtime.CloneBigInt(den)
	if d.Sign() < 0 {
		n.Neg(n)
		d.Neg(d)
	}
	if n.Sign() != 0 {
		gcd := new(big.Int).GCD(nil, nil, absBigInt(n), absBigInt(d))
		if gcd.Sign() != 0 {
			n.Div(n, gcd)
			d.Div(d, gcd)
		}
	} else {
		d.SetInt64(1)
	}
	if n.Cmp(bigI64Min) < 0 || n.Cmp(bigI64Max) > 0 || d.Cmp(bigI64Max) > 0 || d.Sign() <= 0 {
		return ratioParts{}, fmt.Errorf("ratio overflow")
	}
	return ratioParts{num: n, den: d}, nil
}

func absBigInt(val *big.Int) *big.Int {
	if val == nil {
		return big.NewInt(0)
	}
	return new(big.Int).Abs(val)
}

func ratioFromIntegerValue(val runtime.IntegerValue) (ratioParts, error) {
	return normalizeRatioParts(runtime.CloneBigInt(val.Val), big.NewInt(1))
}

func ratioFromFloatValue(val runtime.FloatValue) (ratioParts, error) {
	r := new(big.Rat).SetFloat64(val.Val)
	if r == nil {
		return ratioParts{}, fmt.Errorf("cannot convert non-finite float to Ratio")
	}
	return normalizeRatioParts(runtime.CloneBigInt(r.Num()), runtime.CloneBigInt(r.Denom()))
}

func coerceToRatio(val runtime.Value) (ratioParts, error) {
	if parts, ok := ratioPartsFromStruct(val); ok {
		return normalizeRatioParts(parts.num, parts.den)
	}
	switch v := val.(type) {
	case runtime.IntegerValue:
		return ratioFromIntegerValue(v)
	case runtime.FloatValue:
		return ratioFromFloatValue(v)
	default:
		return ratioParts{}, fmt.Errorf("Arithmetic requires numeric operands")
	}
}

func applyBinaryOperator(i *Interpreter, op string, left runtime.Value, right runtime.Value) (runtime.Value, error) {
	op, dotted := normalizeOperator(op)
	switch op {
	case "+", "-", "*", "^":
		if op == "^" && dotted {
			return evaluateBitwise(op, left, right)
		}
		return evaluateArithmetic(i, op, left, right)
	case "/":
		return evaluateDivision(i, left, right)
	case "//", "%", "/%":
		return evaluateDivMod(i, op, left, right)
	case "<", "<=", ">", ">=":
		return evaluateComparison(op, left, right)
	case "==":
		return runtime.BoolValue{Val: valuesEqual(left, right)}, nil
	case "!=":
		return runtime.BoolValue{Val: !valuesEqual(left, right)}, nil
	case "&", "|", "<<", ">>":
		return evaluateBitwise(op, left, right)
	default:
		return nil, fmt.Errorf("unsupported binary operator %s", op)
	}
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
		if rightFloat == 0 {
			return nil, fmt.Errorf("division by zero")
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
		return nil, fmt.Errorf("division by zero")
	}
	leftFloat := bigIntToFloat(leftInt.Val)
	rightFloat := bigIntToFloat(rightInt.Val)
	if rightFloat == 0 {
		return nil, fmt.Errorf("division by zero")
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
	targetType, err := promoteIntegerTypes(lv.TypeSuffix, rv.TypeSuffix)
	if err != nil {
		return nil, err
	}
	info, err := getIntegerInfo(targetType)
	if err != nil {
		return nil, err
	}
	lVal := runtime.CloneBigInt(lv.Val)
	rVal := runtime.CloneBigInt(rv.Val)
	var result *big.Int
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
	case "<<":
		if !rVal.IsInt64() {
			return nil, fmt.Errorf("shift out of range")
		}
		count := int(rVal.Int64())
		shifted, err := shiftValueLeft(lVal, count, info)
		if err != nil {
			return nil, err
		}
		result = shifted
	case ">>":
		if !rVal.IsInt64() {
			return nil, fmt.Errorf("shift out of range")
		}
		count := int(rVal.Int64())
		shifted, err := shiftValueRight(lVal, count, info)
		if err != nil {
			return nil, err
		}
		result = shifted
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
		return runtime.IntegerValue{}, runtime.IntegerValue{}, runtime.IntegerI32, fmt.Errorf("division by zero")
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
		return nil, nil, fmt.Errorf("division by zero")
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

func valuesEqual(left runtime.Value, right runtime.Value) bool {
	if iv, ok := left.(runtime.InterfaceValue); ok {
		return valuesEqual(iv.Underlying, right)
	}
	if iv, ok := right.(runtime.InterfaceValue); ok {
		return valuesEqual(left, iv.Underlying)
	}
	switch lv := left.(type) {
	case runtime.StructDefinitionValue:
		switch rv := right.(type) {
		case runtime.StructDefinitionValue:
			return structDefName(lv) != "" && structDefName(lv) == structDefName(rv)
		case *runtime.StructInstanceValue:
			return structDefName(lv) != "" && structDefName(lv) == structInstanceName(rv) && structInstanceEmpty(rv)
		}
	case *runtime.StructInstanceValue:
		switch rv := right.(type) {
		case runtime.StructDefinitionValue:
			return structInstanceName(lv) != "" && structInstanceName(lv) == structDefName(rv) && structInstanceEmpty(lv)
		case *runtime.StructInstanceValue:
			return structInstancesEqual(lv, rv)
		}
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
		switch rv := right.(type) {
		case runtime.IntegerValue:
			return lv.Val.Cmp(rv.Val) == 0
		case runtime.FloatValue:
			return math.Abs(bigIntToFloat(lv.Val)-rv.Val) < numericEpsilon
		}
	case runtime.FloatValue:
		switch rv := right.(type) {
		case runtime.FloatValue:
			return math.Abs(lv.Val-rv.Val) < numericEpsilon
		case runtime.IntegerValue:
			return math.Abs(lv.Val-bigIntToFloat(rv.Val)) < numericEpsilon
		}
	}
	return false
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

func (i *Interpreter) makeRatioValue(parts ratioParts) (runtime.Value, error) {
	def, err := i.ensureRatioStruct()
	if err != nil {
		return nil, err
	}
	fields := map[string]runtime.Value{
		"num": runtime.IntegerValue{Val: parts.num, TypeSuffix: runtime.IntegerI64},
		"den": runtime.IntegerValue{Val: parts.den, TypeSuffix: runtime.IntegerI64},
	}
	return &runtime.StructInstanceValue{
		Definition: def,
		Fields:     fields,
	}, nil
}

func (i *Interpreter) ensureRatioStruct() (*runtime.StructDefinitionValue, error) {
	if i.ratioStruct != nil {
		return i.ratioStruct, nil
	}
	if val, err := i.global.Get("Ratio"); err == nil {
		if def, conv := toStructDefinitionValue(val, "Ratio"); conv == nil {
			i.ratioStruct = def
			return def, nil
		}
	}
	numField := ast.NewStructFieldDefinition(ast.NewSimpleTypeExpression(ast.NewIdentifier("i64")), ast.NewIdentifier("num"))
	denField := ast.NewStructFieldDefinition(ast.NewSimpleTypeExpression(ast.NewIdentifier("i64")), ast.NewIdentifier("den"))
	definition := ast.NewStructDefinition(
		ast.NewIdentifier("Ratio"),
		[]*ast.StructFieldDefinition{numField, denField},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	if _, err := i.evaluateStructDefinition(definition, i.global); err != nil {
		return nil, err
	}
	val, err := i.global.Get("Ratio")
	if err != nil {
		return nil, err
	}
	structDef, conv := toStructDefinitionValue(val, "Ratio")
	if conv != nil {
		return nil, conv
	}
	i.ratioStruct = structDef
	return structDef, nil
}

func structDefName(def runtime.StructDefinitionValue) string {
	if def.Node != nil && def.Node.ID != nil {
		return def.Node.ID.Name
	}
	return ""
}

func structInstanceName(inst *runtime.StructInstanceValue) string {
	if inst == nil || inst.Definition == nil {
		return ""
	}
	return structDefName(*inst.Definition)
}

func structInstanceEmpty(inst *runtime.StructInstanceValue) bool {
	if inst == nil {
		return true
	}
	if inst.Positional != nil {
		return len(inst.Positional) == 0
	}
	if inst.Fields != nil {
		return len(inst.Fields) == 0
	}
	return true
}

func structInstancesEqual(a *runtime.StructInstanceValue, b *runtime.StructInstanceValue) bool {
	if a == nil || b == nil {
		return false
	}
	if structInstanceName(a) == "" || structInstanceName(a) != structInstanceName(b) {
		return false
	}
	if a.Positional != nil || b.Positional != nil {
		if len(a.Positional) != len(b.Positional) {
			return false
		}
		for i := range a.Positional {
			if !valuesEqual(a.Positional[i], b.Positional[i]) {
				return false
			}
		}
		return true
	}
	if len(a.Fields) != len(b.Fields) {
		return false
	}
	for key, av := range a.Fields {
		bv, ok := b.Fields[key]
		if !ok {
			return false
		}
		if !valuesEqual(av, bv) {
			return false
		}
	}
	return true
}

func evaluateComparison(op string, left runtime.Value, right runtime.Value) (runtime.Value, error) {
	if ls, ok := stringFromValue(left); ok {
		if rs, ok := stringFromValue(right); ok {
			cmp := strings.Compare(ls, rs)
			return runtime.BoolValue{Val: comparisonOp(op, cmp)}, nil
		}
	}
	if isRatioValue(left) || isRatioValue(right) {
		leftRatio, err := coerceToRatio(left)
		if err != nil {
			return nil, err
		}
		rightRatio, err := coerceToRatio(right)
		if err != nil {
			return nil, err
		}
		leftCross := new(big.Int).Mul(leftRatio.num, rightRatio.den)
		rightCross := new(big.Int).Mul(rightRatio.num, leftRatio.den)
		cmp := leftCross.Cmp(rightCross)
		return runtime.BoolValue{Val: comparisonOp(op, cmp)}, nil
	}
	if !isNumericValue(left) || !isNumericValue(right) {
		return nil, fmt.Errorf("Arithmetic requires numeric operands")
	}
	if li, ok := left.(runtime.IntegerValue); ok {
		if ri, ok := right.(runtime.IntegerValue); ok {
			cmp := li.Val.Cmp(ri.Val)
			return runtime.BoolValue{Val: comparisonOp(op, cmp)}, nil
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

func stringFromValue(val runtime.Value) (string, bool) {
	switch v := val.(type) {
	case runtime.StringValue:
		return v.Val, true
	case *runtime.StringValue:
		if v != nil {
			return v.Val, true
		}
		return "", false
	default:
		return "", false
	}
}
