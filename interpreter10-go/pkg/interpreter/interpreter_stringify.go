package interpreter

import (
	"fmt"
	"sort"
	"strings"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
)

func (i *Interpreter) stringifyValue(val runtime.Value, env *runtime.Environment) (string, error) {
	_ = env
	if inst, ok := val.(*runtime.StructInstanceValue); ok {
		if str, ok := i.invokeStructToString(inst); ok {
			return str, nil
		}
	}
	return valueToString(val), nil
}

func (i *Interpreter) invokeStructToString(inst *runtime.StructInstanceValue) (string, bool) {
	if inst == nil {
		return "", false
	}
	typeName := structTypeName(inst)
	if typeName == "" {
		return "", false
	}
	if bucket, ok := i.inherentMethods[typeName]; ok {
		if method := bucket["to_string"]; method != nil {
			if str, ok := i.callStringMethod(method, inst); ok {
				return str, true
			}
		}
	}
	if method, err := i.selectStructMethod(inst, "to_string"); err == nil && method != nil {
		if str, ok := i.callStringMethod(method, inst); ok {
			return str, true
		}
	}
	return "", false
}

func (i *Interpreter) callStringMethod(fn *runtime.FunctionValue, receiver runtime.Value) (string, bool) {
	if fn == nil {
		return "", false
	}
	result, err := i.invokeFunction(fn, []runtime.Value{receiver}, nil)
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
	case *runtime.RangeValue:
		start := valueToString(v.Start)
		end := valueToString(v.End)
		delim := "..."
		if v.Inclusive {
			delim = ".."
		}
		return fmt.Sprintf("%s%s%s", start, delim, end)
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
	case runtime.NativeFunctionValue:
		return fmt.Sprintf("<native %s>", v.Name)
	case *runtime.NativeFunctionValue:
		return fmt.Sprintf("<native %s>", v.Name)
	case runtime.BoundMethodValue:
		return "<bound method>"
	case *runtime.BoundMethodValue:
		return "<bound method>"
	case runtime.NativeBoundMethodValue:
		return fmt.Sprintf("<native bound %s>", v.Method.Name)
	case *runtime.NativeBoundMethodValue:
		return fmt.Sprintf("<native bound %s>", v.Method.Name)
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
