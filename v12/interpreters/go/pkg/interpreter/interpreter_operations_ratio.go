package interpreter

import (
	"fmt"
	"math"
	"math/big"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

var (
	bigI64Min = big.NewInt(math.MinInt64)
	bigI64Max = big.NewInt(math.MaxInt64)
)

type ratioParts struct {
	num *big.Int
	den *big.Int
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
		return ratioParts{}, newDivisionByZeroError()
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
		return ratioParts{}, newOverflowError("ratio overflow")
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
	for _, key := range []string{"able.kernel.Ratio", "kernel.Ratio"} {
		if val, err := i.global.Get(key); err == nil {
			if def, conv := toStructDefinitionValue(val, "Ratio"); conv == nil {
				i.ratioStruct = def
				return def, nil
			}
		}
	}
	for _, bucket := range i.packageRegistry {
		if val, ok := bucket["Ratio"]; ok {
			if def, conv := toStructDefinitionValue(val, "Ratio"); conv == nil {
				i.ratioStruct = def
				return def, nil
			}
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
