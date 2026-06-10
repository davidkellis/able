package interpreter

import (
	"fmt"
	"reflect"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func externStringArrayArg(i *Interpreter, value runtime.Value) ([]string, bool) {
	for {
		switch v := value.(type) {
		case runtime.InterfaceValue:
			value = v.Underlying
			continue
		case *runtime.InterfaceValue:
			if v == nil {
				return nil, false
			}
			value = v.Underlying
			continue
		}
		break
	}
	arr, ok := value.(*runtime.ArrayValue)
	if !ok || arr == nil {
		if i == nil {
			return nil, false
		}
		var err error
		arr, err = i.toArrayValue(value)
		if err != nil {
			return nil, false
		}
	}
	if i != nil {
		if _, err := i.ensureArrayState(arr, 0); err != nil {
			return nil, false
		}
	}
	elements := arr.Elements
	if arr.State != nil {
		elements = arr.State.Values
	}
	values := make([]string, len(elements))
	for idx, elem := range elements {
		text, ok := externStringArg(elem)
		if !ok {
			return nil, false
		}
		values[idx] = text
	}
	return values, true
}

func externF64SliceResult(i *Interpreter, values []float64) runtime.Value {
	if len(values) == 0 {
		if i != nil {
			return i.newArrayValue([]runtime.Value{}, 0)
		}
		return &runtime.ArrayValue{Elements: []runtime.Value{}}
	}
	boxed := make([]runtime.Value, len(values))
	for idx, value := range values {
		boxed[idx] = runtime.FloatValue{Val: value, TypeSuffix: runtime.FloatF64}
	}
	if i != nil {
		return i.newArrayValue(boxed, len(boxed))
	}
	return &runtime.ArrayValue{Elements: boxed}
}

func buildExternStringArrayFastInvoker(def *ast.ExternFunctionBody, raw any) externHostInvoker {
	if def == nil || def.Signature == nil || len(def.Signature.Params) != 3 {
		return nil
	}
	if externSimpleTypeName(def.Signature.Params[0].ParamType) != "String" ||
		externSimpleTypeName(def.Signature.Params[1].ParamType) != "String" ||
		!externIsArrayStringType(def.Signature.Params[2].ParamType) {
		return nil
	}
	returnType := def.Signature.ReturnType
	switch fn := raw.(type) {
	case func(string, string, []string) interface{}:
		return func(i *Interpreter, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 3 {
				return nil, fmt.Errorf("extern fast invoker expects 3 args, got %d", len(args))
			}
			first, ok := externStringArg(args[0])
			if !ok {
				return nil, fmt.Errorf("extern fast invoker expected String argument")
			}
			second, ok := externStringArg(args[1])
			if !ok {
				return nil, fmt.Errorf("extern fast invoker expected String argument")
			}
			third, ok := externStringArrayArg(i, args[2])
			if !ok {
				return nil, fmt.Errorf("extern fast invoker expected Array String argument")
			}
			result := fn(first, second, third)
			if externUnionHasArrayF64Member(returnType) {
				if floats, ok := externReflectF64SliceValues(reflect.ValueOf(result)); ok {
					return externF64SliceResult(i, floats), nil
				}
			}
			if i == nil {
				return nil, fmt.Errorf("extern fast invoker requires interpreter fallback")
			}
			var reflected reflect.Value
			if result != nil {
				reflected = reflect.ValueOf(result)
			}
			return i.fromHostValue(returnType, reflected)
		}
	}
	return nil
}
