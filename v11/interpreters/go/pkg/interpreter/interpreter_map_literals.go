package interpreter

import (
	"fmt"
	"math/big"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func (i *Interpreter) evaluateMapLiteral(lit *ast.MapLiteral, env *runtime.Environment) (runtime.Value, error) {
	i.ensureHashMapBuiltins()
	defVal, err := env.Get("HashMap")
	if err != nil {
		return nil, err
	}
	structDef, err := toStructDefinitionValue(defVal, "HashMap")
	if err != nil {
		return nil, err
	}
	elementCount := 0
	if lit != nil {
		elementCount = len(lit.Elements)
	}
	handle := i.newHashMapHandle(elementCount)
	state, err := i.hashMapStateForHandle(handle)
	if err != nil {
		return nil, err
	}
	handleValue := runtime.IntegerValue{Val: big.NewInt(handle), TypeSuffix: runtime.IntegerI64}
	instance := &runtime.StructInstanceValue{
		Definition: structDef,
		Fields:     map[string]runtime.Value{"handle": handleValue},
	}
	if lit == nil || elementCount == 0 {
		instance.TypeArguments = []ast.TypeExpression{
			ast.NewWildcardTypeExpression(),
			ast.NewWildcardTypeExpression(),
		}
		return instance, nil
	}
	mergeType := func(current, next ast.TypeExpression) ast.TypeExpression {
		if current == nil {
			return next
		}
		if next == nil {
			return current
		}
		if _, ok := current.(*ast.WildcardTypeExpression); ok {
			return current
		}
		if _, ok := next.(*ast.WildcardTypeExpression); ok {
			return next
		}
		if typeExpressionsEqual(current, next) {
			return current
		}
		return ast.NewWildcardTypeExpression()
	}

	var keyType ast.TypeExpression
	var valueType ast.TypeExpression
	for _, element := range lit.Elements {
		switch entry := element.(type) {
		case *ast.MapLiteralEntry:
			keyVal, err := i.evaluateExpression(entry.Key, env)
			if err != nil {
				return nil, err
			}
			valueVal, err := i.evaluateExpression(entry.Value, env)
			if err != nil {
				return nil, err
			}
			if err := i.hashMapInsertEntry(state, keyVal, valueVal); err != nil {
				return nil, err
			}
			keyType = mergeType(keyType, i.typeExpressionFromValue(keyVal))
			valueType = mergeType(valueType, i.typeExpressionFromValue(valueVal))
		case *ast.MapLiteralSpread:
			spreadVal, err := i.evaluateExpression(entry.Expression, env)
			if err != nil {
				return nil, err
			}
			source, ok := i.hashMapInstanceFromValue(spreadVal)
			if !ok {
				return nil, fmt.Errorf("map literal spread expects HashMap value")
			}
			sourceHandle, err := i.hashMapHandleFromInstance(source)
			if err != nil {
				return nil, err
			}
			sourceState, err := i.hashMapStateForHandle(sourceHandle)
			if err != nil {
				return nil, err
			}
			for _, existing := range sourceState.Entries {
				if err := i.hashMapInsertEntry(state, existing.Key, existing.Value); err != nil {
					return nil, err
				}
			}
			if len(source.TypeArguments) >= 2 {
				keyType = mergeType(keyType, source.TypeArguments[0])
				valueType = mergeType(valueType, source.TypeArguments[1])
			}
		default:
			return nil, fmt.Errorf("unsupported map literal element %T", element)
		}
	}
	if keyType == nil {
		keyType = ast.NewWildcardTypeExpression()
	}
	if valueType == nil {
		valueType = ast.NewWildcardTypeExpression()
	}
	instance.TypeArguments = []ast.TypeExpression{keyType, valueType}
	return instance, nil
}

func (i *Interpreter) hashMapInstanceFromValue(value runtime.Value) (*runtime.StructInstanceValue, bool) {
	switch v := value.(type) {
	case *runtime.StructInstanceValue:
		if v != nil && structInstanceName(v) == "HashMap" {
			return v, true
		}
	case *runtime.InterfaceValue:
		if v == nil {
			return nil, false
		}
		return i.hashMapInstanceFromValue(v.Underlying)
	case runtime.InterfaceValue:
		return i.hashMapInstanceFromValue(&v)
	}
	return nil, false
}

func (i *Interpreter) hashMapHandleFromInstance(inst *runtime.StructInstanceValue) (int64, error) {
	if inst == nil || inst.Fields == nil {
		return 0, fmt.Errorf("hash map handle missing")
	}
	raw, ok := inst.Fields["handle"]
	if !ok {
		return 0, fmt.Errorf("hash map handle missing")
	}
	switch v := raw.(type) {
	case runtime.IntegerValue:
		if v.Val == nil || !v.Val.IsInt64() {
			return 0, fmt.Errorf("hash map handle is invalid")
		}
		return v.Val.Int64(), nil
	case *runtime.IntegerValue:
		if v == nil || v.Val == nil || !v.Val.IsInt64() {
			return 0, fmt.Errorf("hash map handle is invalid")
		}
		return v.Val.Int64(), nil
	default:
		return 0, fmt.Errorf("hash map handle must be an integer")
	}
}
