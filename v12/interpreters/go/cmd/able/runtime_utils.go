package main

import (
	"fmt"
	"os"
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
	default:
		return fmt.Sprintf("[%s]", v.Kind())
	}
}
