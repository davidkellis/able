package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"able/interpreter-go/pkg/interpreter"
	"able/interpreter-go/pkg/runtime"
)

func registerPrint(interp *interpreter.Interpreter) {
	printFn := runtime.NativeFunctionValue{
		Name:  "print",
		Arity: 1,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			var parts []string
			for _, arg := range args {
				parts = append(parts, formatRuntimeValue(arg))
			}
			fmt.Fprintln(os.Stdout, strings.Join(parts, " "))
			return runtime.NilValue{}, nil
		},
	}
	interp.GlobalEnvironment().Define("print", printFn)
}

func formatRuntimeValue(val runtime.Value) string {
	switch v := val.(type) {
	case runtime.StringValue:
		return v.Val
	case runtime.BoolValue:
		if v.Val {
			return "true"
		}
		return "false"
	case runtime.IntegerValue:
		return v.Val.String()
	case runtime.FloatValue:
		return fmt.Sprintf("%g", v.Val)
	case runtime.CharValue:
		return string(v.Val)
	case runtime.NilValue:
		return "nil"
	case *runtime.ArrayValue:
		elems := make([]string, len(v.Elements))
		for i, el := range v.Elements {
			elems[i] = formatRuntimeValue(el)
		}
		return "[" + strings.Join(elems, ", ") + "]"
	case *runtime.StructInstanceValue:
		if v.Definition != nil && v.Definition.Node != nil && v.Definition.Node.ID != nil && v.Definition.Node.ID.Name == "Array" {
			if h, ok := v.Fields["storage_handle"]; ok {
				if hv, ok := h.(runtime.IntegerValue); ok {
					handle := hv.Val.Int64()
					state, err := runtime.ArrayStoreState(handle)
					if err == nil {
						elems := make([]string, len(state.Values))
						for i, el := range state.Values {
							elems[i] = formatRuntimeValue(el)
						}
						return "[" + strings.Join(elems, ", ") + "]"
					}
				}
			}
		}
		name := "Struct"
		if v.Definition != nil && v.Definition.Node != nil && v.Definition.Node.ID != nil {
			name = v.Definition.Node.ID.Name
		}
		keys := make([]string, 0, len(v.Fields))
		for k := range v.Fields {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		fields := make([]string, len(keys))
		for i, k := range keys {
			fields[i] = k + ": " + formatRuntimeValue(v.Fields[k])
		}
		return name + " { " + strings.Join(fields, ", ") + " }"
	case runtime.ErrorValue:
		return v.Message
	case *runtime.InterfaceValue:
		if v.Interface != nil && v.Interface.Node != nil && v.Interface.Node.ID != nil {
			return "<interface " + v.Interface.Node.ID.Name + ">"
		}
		return "<interface>"
	default:
		return fmt.Sprintf("<%s>", v.Kind())
	}
}
