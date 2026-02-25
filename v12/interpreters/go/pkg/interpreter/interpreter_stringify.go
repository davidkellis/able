package interpreter

import (
	"fmt"
	"sort"
	"strings"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func (i *Interpreter) stringifyValue(val runtime.Value, env *runtime.Environment) (string, error) {
	_ = env
	if inst, ok := val.(*runtime.StructInstanceValue); ok {
		if str, ok := i.stringifyArrayStruct(inst); ok {
			return str, nil
		}
		if str, ok := i.invokeStructToString(inst); ok {
			return str, nil
		}
	}
	return valueToString(val), nil
}

// Stringify is an exported wrapper for compiled interop.
func (i *Interpreter) Stringify(val runtime.Value, env *runtime.Environment) (string, error) {
	if i == nil {
		return "", fmt.Errorf("interpreter: nil interpreter")
	}
	if env == nil {
		env = i.GlobalEnvironment()
	}
	return i.stringifyValue(val, env)
}

func (i *Interpreter) invokeStructToString(inst *runtime.StructInstanceValue) (string, bool) {
	if inst == nil {
		return "", false
	}
	typeName := structTypeName(inst)
	if typeName == "" {
		return "", false
	}
	for _, candidate := range structTypeNameCandidates(typeName) {
		if bucket, ok := i.inherentMethods[candidate]; ok {
			if method := bucket["to_string"]; method != nil {
				if str, ok := i.callStringMethod(method, inst); ok {
					return str, true
				}
			}
		}
	}
	if method, err := i.selectStructMethod(inst, "to_string"); err == nil && method != nil {
		if str, ok := i.callStringMethod(method, inst); ok {
			return str, true
		}
	}
	if i.interfaceMethodResolver != nil {
		if method, ok := i.interfaceMethodResolver(inst, "Display", "to_string"); ok && method != nil {
			if str, strOk := i.callStringMethod(method, inst); strOk {
				return str, true
			}
		}
	}
	if i.compiledInstanceMethodFn != nil {
		for _, candidate := range structTypeNameCandidates(typeName) {
			if method, ok := i.compiledInstanceMethodFn(candidate, "to_string"); ok && method != nil {
				if str, strOk := i.callStringMethod(method, inst); strOk {
					return str, true
				}
			}
		}
	}
	return "", false
}

func structTypeNameCandidates(typeName string) []string {
	candidates := make([]string, 0, 2)
	if typeName == "" {
		return candidates
	}
	candidates = append(candidates, typeName)
	if idx := strings.LastIndex(typeName, "."); idx >= 0 && idx+1 < len(typeName) {
		short := typeName[idx+1:]
		if short != "" && short != typeName {
			candidates = append(candidates, short)
		}
	}
	return candidates
}

func isArrayStructInstance(inst *runtime.StructInstanceValue) bool {
	if inst == nil {
		return false
	}
	if inst.Definition != nil && inst.Definition.Node != nil && inst.Definition.Node.ID != nil {
		name := inst.Definition.Node.ID.Name
		if name == "Array" || strings.HasSuffix(name, ".Array") {
			return true
		}
	}
	_, hasHandle := inst.Fields["storage_handle"]
	_, hasLength := inst.Fields["length"]
	_, hasCapacity := inst.Fields["capacity"]
	return hasHandle && hasLength && hasCapacity
}

func (i *Interpreter) stringifyArrayStruct(inst *runtime.StructInstanceValue) (string, bool) {
	if !isArrayStructInstance(inst) {
		return "", false
	}
	handleValue, ok := inst.Fields["storage_handle"]
	if !ok {
		return "", false
	}
	handleInt, ok := handleValue.(runtime.IntegerValue)
	if !ok {
		return "", false
	}
	state, err := runtime.ArrayStoreState(handleInt.Val.Int64())
	if err != nil {
		return "", false
	}
	parts := make([]string, 0, len(state.Values))
	for _, item := range state.Values {
		rendered, renderErr := i.stringifyValue(item, nil)
		if renderErr != nil {
			rendered = valueToString(item)
		}
		parts = append(parts, rendered)
	}
	return "[" + strings.Join(parts, ", ") + "]", true
}

func (i *Interpreter) callStringMethod(fn runtime.Value, receiver runtime.Value) (string, bool) {
	if fn == nil {
		return "", false
	}
	var result runtime.Value
	var err error
	if native, ok := fn.(*runtime.NativeFunctionValue); ok && native != nil {
		bound := runtime.NativeBoundMethodValue{Receiver: receiver, Method: *native}
		result, err = i.callCallableValue(bound, nil, nil, nil)
		// The compiled interface method resolver returns arity = N+1 (includes self).
		// NativeBoundMethodValue subtracts injected from provided, producing a partial
		// instead of calling the method. Detect this and retry with self as an explicit arg.
		if err == nil && result != nil {
			if _, isPartial := result.(*runtime.PartialFunctionValue); isPartial {
				result, err = i.callCallableValue(native, []runtime.Value{receiver}, nil, nil)
			}
		}
	} else {
		bound := runtime.BoundMethodValue{Receiver: receiver, Method: fn}
		result, err = i.callCallableValue(bound, nil, nil, nil)
	}
	if err != nil {
		return "", false
	}
	if result == nil {
		return "", false
	}
	if strVal, ok := result.(runtime.StringValue); ok {
		return strVal.Val, true
	}
	return "", false
}

func structTypeName(inst *runtime.StructInstanceValue) string {
	if inst == nil {
		return ""
	}
	if inst.Definition != nil && inst.Definition.Node != nil && inst.Definition.Node.ID != nil {
		return inst.Definition.Node.ID.Name
	}
	return ""
}

func simpleTypeName(expr ast.TypeExpression) (string, bool) {
	switch t := expr.(type) {
	case *ast.SimpleTypeExpression:
		if t.Name != nil {
			return t.Name.Name, true
		}
	case *ast.GenericTypeExpression:
		if base, ok := t.Base.(*ast.SimpleTypeExpression); ok && base != nil && base.Name != nil {
			return base.Name.Name, true
		}
	}
	return "", false
}

func valueToString(val runtime.Value) string {
	switch v := val.(type) {
	case runtime.StringValue:
		return v.Val
	case runtime.BoolValue:
		if v.Val {
			return "true"
		}
		return "false"
	case runtime.CharValue:
		return string(v.Val)
	case runtime.IntegerValue:
		return v.Val.String()
	case runtime.FloatValue:
		return fmt.Sprintf("%g", v.Val)
	case runtime.NilValue:
		return "nil"
	case *runtime.ArrayValue:
		parts := make([]string, 0, len(v.Elements))
		for _, el := range v.Elements {
			parts = append(parts, valueToString(el))
		}
		return fmt.Sprintf("[%s]", strings.Join(parts, ", "))
	case *runtime.StructInstanceValue:
		return structInstanceToString(v)
	case runtime.StructDefinitionValue:
		name := "struct"
		if v.Node != nil && v.Node.ID != nil {
			name = v.Node.ID.Name
		}
		return fmt.Sprintf("<struct %s>", name)
	case *runtime.StructDefinitionValue:
		name := "struct"
		if v.Node != nil && v.Node.ID != nil {
			name = v.Node.ID.Name
		}
		return fmt.Sprintf("<struct %s>", name)
	case *runtime.InterfaceDefinitionValue:
		name := "interface"
		if v.Node != nil && v.Node.ID != nil {
			name = v.Node.ID.Name
		}
		return fmt.Sprintf("<interface %s>", name)
	case runtime.InterfaceDefinitionValue:
		name := "interface"
		if v.Node != nil && v.Node.ID != nil {
			name = v.Node.ID.Name
		}
		return fmt.Sprintf("<interface %s>", name)
	case *runtime.InterfaceValue:
		name := "interface"
		if v.Interface != nil && v.Interface.Node != nil && v.Interface.Node.ID != nil {
			name = v.Interface.Node.ID.Name
		}
		return fmt.Sprintf("<interface %s>", name)
	case *runtime.FunctionValue:
		return "<function>"
	case *runtime.FunctionOverloadValue:
		return "<function overload>"
	case runtime.NativeFunctionValue:
		return fmt.Sprintf("<native %s>", v.Name)
	case *runtime.NativeFunctionValue:
		return fmt.Sprintf("<native %s>", v.Name)
	case runtime.BoundMethodValue:
		methodStr := valueToString(v.Method)
		if methodStr == "" {
			return "<bound method>"
		}
		return fmt.Sprintf("<bound method %s>", methodStr)
	case *runtime.BoundMethodValue:
		if v == nil {
			return "<bound method>"
		}
		target := valueToString(v.Method)
		if target == "" {
			return "<bound method>"
		}
		return fmt.Sprintf("<bound method %s>", target)
	case runtime.NativeBoundMethodValue:
		return fmt.Sprintf("<native bound %s>", v.Method.Name)
	case *runtime.NativeBoundMethodValue:
		return fmt.Sprintf("<native bound %s>", v.Method.Name)
	case runtime.PartialFunctionValue, *runtime.PartialFunctionValue:
		return "<partial>"
	case runtime.PackageValue:
		name := v.Name
		if name == "" {
			name = strings.Join(v.NamePath, "::")
		}
		if name == "" {
			name = "<package>"
		}
		return fmt.Sprintf("<package %s>", name)
	case runtime.ErrorValue:
		return v.Message
	case *runtime.IteratorValue:
		return "<iterator>"
	case runtime.IteratorEndValue:
		return "IteratorEnd"
	case *runtime.IteratorEndValue:
		return "IteratorEnd"
	default:
		if val == nil {
			return "<nil>"
		}
		return fmt.Sprintf("[%s]", val.Kind())
	}
}

func structInstanceToString(inst *runtime.StructInstanceValue) string {
	if inst == nil {
		return "<struct> {}"
	}
	name := structTypeName(inst)
	if name == "" {
		name = "<struct>"
	}
	if inst.Positional != nil {
		parts := make([]string, 0, len(inst.Positional))
		for _, el := range inst.Positional {
			parts = append(parts, valueToString(el))
		}
		return fmt.Sprintf("%s { %s }", name, strings.Join(parts, ", "))
	}
	if inst.Fields != nil {
		keys := make([]string, 0, len(inst.Fields))
		for k := range inst.Fields {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		parts := make([]string, 0, len(keys))
		for _, k := range keys {
			parts = append(parts, fmt.Sprintf("%s: %s", k, valueToString(inst.Fields[k])))
		}
		return fmt.Sprintf("%s { %s }", name, strings.Join(parts, ", "))
	}
	return fmt.Sprintf("%s { }", name)
}
