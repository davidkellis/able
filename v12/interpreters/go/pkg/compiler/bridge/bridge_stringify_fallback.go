package bridge

import (
	"fmt"
	"sort"
	"strings"

	"able/interpreter-go/pkg/runtime"
)

func fallbackValueToString(value runtime.Value) string {
	value = unwrapInterface(value)
	switch v := value.(type) {
	case nil:
		return "<nil>"
	case runtime.StringValue:
		return v.Val
	case *runtime.StringValue:
		if v == nil {
			return "<nil>"
		}
		return v.Val
	case runtime.BoolValue:
		if v.Val {
			return "true"
		}
		return "false"
	case *runtime.BoolValue:
		if v == nil {
			return "<nil>"
		}
		if v.Val {
			return "true"
		}
		return "false"
	case runtime.CharValue:
		return string(v.Val)
	case *runtime.CharValue:
		if v == nil {
			return "<nil>"
		}
		return string(v.Val)
	case runtime.IntegerValue:
		return v.String()
	case *runtime.IntegerValue:
		if v == nil {
			return "<nil>"
		}
		return v.String()
	case runtime.FloatValue:
		return fmt.Sprintf("%g", v.Val)
	case *runtime.FloatValue:
		if v == nil {
			return "<nil>"
		}
		return fmt.Sprintf("%g", v.Val)
	case runtime.NilValue, *runtime.NilValue:
		return "nil"
	case runtime.VoidValue, *runtime.VoidValue:
		return "void"
	case *runtime.ArrayValue:
		if v == nil {
			return "<nil>"
		}
		parts := make([]string, 0, len(v.Elements))
		for _, el := range v.Elements {
			parts = append(parts, fallbackValueToString(el))
		}
		return fmt.Sprintf("[%s]", strings.Join(parts, ", "))
	case *runtime.StructInstanceValue:
		if v == nil {
			return "<struct> {}"
		}
		if str, err := AsString(v); err == nil {
			return str
		}
		return fallbackStructInstanceToString(v)
	case runtime.StructDefinitionValue:
		name := "struct"
		if v.Node != nil && v.Node.ID != nil {
			name = v.Node.ID.Name
		}
		return fmt.Sprintf("<struct %s>", name)
	case *runtime.StructDefinitionValue:
		name := "struct"
		if v != nil && v.Node != nil && v.Node.ID != nil {
			name = v.Node.ID.Name
		}
		return fmt.Sprintf("<struct %s>", name)
	case runtime.InterfaceDefinitionValue:
		name := "interface"
		if v.Node != nil && v.Node.ID != nil {
			name = v.Node.ID.Name
		}
		return fmt.Sprintf("<interface %s>", name)
	case *runtime.InterfaceDefinitionValue:
		name := "interface"
		if v != nil && v.Node != nil && v.Node.ID != nil {
			name = v.Node.ID.Name
		}
		return fmt.Sprintf("<interface %s>", name)
	case *runtime.FunctionValue:
		return "<function>"
	case *runtime.FunctionOverloadValue:
		return "<function overload>"
	case runtime.NativeFunctionValue:
		return fmt.Sprintf("<native %s>", v.Name)
	case *runtime.NativeFunctionValue:
		if v == nil {
			return "<native>"
		}
		return fmt.Sprintf("<native %s>", v.Name)
	case runtime.BoundMethodValue:
		method := fallbackValueToString(v.Method)
		if method == "" {
			return "<bound method>"
		}
		return fmt.Sprintf("<bound method %s>", method)
	case *runtime.BoundMethodValue:
		if v == nil {
			return "<bound method>"
		}
		method := fallbackValueToString(v.Method)
		if method == "" {
			return "<bound method>"
		}
		return fmt.Sprintf("<bound method %s>", method)
	case runtime.NativeBoundMethodValue:
		return fmt.Sprintf("<native bound %s>", v.Method.Name)
	case *runtime.NativeBoundMethodValue:
		if v == nil {
			return "<native bound>"
		}
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
	case *runtime.ErrorValue:
		if v == nil {
			return "<nil>"
		}
		return v.Message
	case *runtime.IteratorValue:
		return "<iterator>"
	case runtime.IteratorEndValue, *runtime.IteratorEndValue:
		return "IteratorEnd"
	default:
		return fmt.Sprintf("[%s]", value.Kind())
	}
}

func fallbackStructInstanceToString(inst *runtime.StructInstanceValue) string {
	name := structNameFromValue(inst)
	if name == "" {
		name = "<struct>"
	}
	if inst.Positional != nil {
		parts := make([]string, 0, len(inst.Positional))
		for _, el := range inst.Positional {
			parts = append(parts, fallbackValueToString(el))
		}
		return fmt.Sprintf("%s { %s }", name, strings.Join(parts, ", "))
	}
	if inst.Fields != nil {
		keys := make([]string, 0, len(inst.Fields))
		for key := range inst.Fields {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		parts := make([]string, 0, len(keys))
		for _, key := range keys {
			parts = append(parts, fmt.Sprintf("%s: %s", key, fallbackValueToString(inst.Fields[key])))
		}
		return fmt.Sprintf("%s { %s }", name, strings.Join(parts, ", "))
	}
	return fmt.Sprintf("%s { }", name)
}
