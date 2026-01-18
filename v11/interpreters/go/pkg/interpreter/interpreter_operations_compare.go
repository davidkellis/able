package interpreter

import (
	"fmt"
	"math"
	"math/big"
	"strings"

	"able/interpreter-go/pkg/runtime"
)

func valuesEqual(left runtime.Value, right runtime.Value) bool {
	if iv, ok := left.(runtime.InterfaceValue); ok {
		return valuesEqual(iv.Underlying, right)
	}
	if iv, ok := right.(runtime.InterfaceValue); ok {
		return valuesEqual(left, iv.Underlying)
	}
	switch lv := left.(type) {
	case *runtime.StructDefinitionValue:
		if lv == nil {
			return false
		}
		switch rv := right.(type) {
		case *runtime.StructDefinitionValue:
			if rv == nil {
				return false
			}
			return structDefName(*lv) != "" && structDefName(*lv) == structDefName(*rv)
		case runtime.StructDefinitionValue:
			return structDefName(*lv) != "" && structDefName(*lv) == structDefName(rv)
		case runtime.IteratorEndValue, *runtime.IteratorEndValue:
			return structDefName(*lv) == "IteratorEnd"
		case *runtime.StructInstanceValue:
			return structDefName(*lv) != "" && structDefName(*lv) == structInstanceName(rv) && structInstanceEmpty(rv)
		}
	case runtime.StructDefinitionValue:
		switch rv := right.(type) {
		case runtime.StructDefinitionValue:
			return structDefName(lv) != "" && structDefName(lv) == structDefName(rv)
		case *runtime.StructDefinitionValue:
			if rv == nil {
				return false
			}
			return structDefName(lv) != "" && structDefName(lv) == structDefName(*rv)
		case runtime.IteratorEndValue, *runtime.IteratorEndValue:
			return structDefName(lv) == "IteratorEnd"
		case *runtime.StructInstanceValue:
			return structDefName(lv) != "" && structDefName(lv) == structInstanceName(rv) && structInstanceEmpty(rv)
		}
	case *runtime.StructInstanceValue:
		switch rv := right.(type) {
		case runtime.StructDefinitionValue:
			return structInstanceName(lv) != "" && structInstanceName(lv) == structDefName(rv) && structInstanceEmpty(lv)
		case *runtime.StructDefinitionValue:
			if rv == nil {
				return false
			}
			return structInstanceName(lv) != "" && structInstanceName(lv) == structDefName(*rv) && structInstanceEmpty(lv)
		case runtime.IteratorEndValue, *runtime.IteratorEndValue:
			return structInstanceName(lv) == "IteratorEnd" && structInstanceEmpty(lv)
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
	case runtime.VoidValue:
		_, ok := right.(runtime.VoidValue)
		return ok
	case runtime.IteratorEndValue:
		switch rv := right.(type) {
		case runtime.IteratorEndValue, *runtime.IteratorEndValue:
			return true
		case runtime.StructDefinitionValue:
			return structDefName(rv) == "IteratorEnd"
		case *runtime.StructDefinitionValue:
			if rv == nil {
				return false
			}
			return structDefName(*rv) == "IteratorEnd"
		case *runtime.StructInstanceValue:
			return structInstanceName(rv) == "IteratorEnd" && structInstanceEmpty(rv)
		}
	case *runtime.IteratorEndValue:
		if lv == nil {
			return false
		}
		switch rv := right.(type) {
		case runtime.IteratorEndValue, *runtime.IteratorEndValue:
			return true
		case runtime.StructDefinitionValue:
			return structDefName(rv) == "IteratorEnd"
		case *runtime.StructDefinitionValue:
			if rv == nil {
				return false
			}
			return structDefName(*rv) == "IteratorEnd"
		case *runtime.StructInstanceValue:
			return structInstanceName(rv) == "IteratorEnd" && structInstanceEmpty(rv)
		}
	case runtime.IntegerValue:
		switch rv := right.(type) {
		case runtime.IntegerValue:
			return lv.Val.Cmp(rv.Val) == 0
		case runtime.FloatValue:
			return bigIntToFloat(lv.Val) == rv.Val
		}
	case runtime.FloatValue:
		switch rv := right.(type) {
		case runtime.FloatValue:
			return lv.Val == rv.Val
		case runtime.IntegerValue:
			return lv.Val == bigIntToFloat(rv.Val)
		}
	}
	return false
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
	if math.IsNaN(leftFloat) || math.IsNaN(rightFloat) {
		switch op {
		case "!=":
			return runtime.BoolValue{Val: true}, nil
		case "==":
			return runtime.BoolValue{Val: false}, nil
		default:
			return runtime.BoolValue{Val: false}, nil
		}
	}
	cmp := 0
	if leftFloat < rightFloat {
		cmp = -1
	} else if leftFloat > rightFloat {
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
