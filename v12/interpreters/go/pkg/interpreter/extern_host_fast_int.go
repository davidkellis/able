package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func externFastIntegerArg(value runtime.Value, kind string) (int64, error) {
	num, err := toInt64(value)
	if err != nil {
		return 0, fmt.Errorf("extern fast invoker expected %s argument", kind)
	}
	return num, nil
}

func externFastI32Result(value int32) runtime.Value {
	return boxedOrSmallIntegerValue(runtime.IntegerI32, int64(value))
}

func externFastI64Result(value int64) runtime.Value {
	return boxedOrSmallIntegerValue(runtime.IntegerI64, value)
}

func buildExternPrimitiveIntegerFastInvoker(def *ast.ExternFunctionBody, raw any) externHostInvoker {
	if def == nil || def.Signature == nil {
		return nil
	}
	paramKinds := make([]string, len(def.Signature.Params))
	for idx, param := range def.Signature.Params {
		kind := externSimpleTypeName(param.ParamType)
		switch kind {
		case "i32", "i64":
			paramKinds[idx] = kind
		default:
			return nil
		}
	}
	returnKind := externSimpleTypeName(def.Signature.ReturnType)
	matchKinds := func(expected ...string) bool {
		if len(expected) != len(paramKinds) {
			return false
		}
		for idx, kind := range expected {
			if paramKinds[idx] != kind {
				return false
			}
		}
		return true
	}

	switch fn := raw.(type) {
	case func(int32) int32:
		if !matchKinds("i32") || returnKind != "i32" {
			return nil
		}
		return func(_ *Interpreter, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("extern fast invoker expects 1 arg, got %d", len(args))
			}
			value, err := externFastIntegerArg(args[0], "i32")
			if err != nil {
				return nil, err
			}
			return externFastI32Result(fn(int32(value))), nil
		}
	case func(int32) int64:
		if !matchKinds("i32") || returnKind != "i64" {
			return nil
		}
		return func(_ *Interpreter, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("extern fast invoker expects 1 arg, got %d", len(args))
			}
			value, err := externFastIntegerArg(args[0], "i32")
			if err != nil {
				return nil, err
			}
			return externFastI64Result(fn(int32(value))), nil
		}
	case func(int32) string:
		if !matchKinds("i32") || returnKind != "String" {
			return nil
		}
		return func(_ *Interpreter, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("extern fast invoker expects 1 arg, got %d", len(args))
			}
			value, err := externFastIntegerArg(args[0], "i32")
			if err != nil {
				return nil, err
			}
			return runtime.StringValue{Val: fn(int32(value))}, nil
		}
	case func(int64) int32:
		if !matchKinds("i64") || returnKind != "i32" {
			return nil
		}
		return func(_ *Interpreter, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("extern fast invoker expects 1 arg, got %d", len(args))
			}
			value, err := externFastIntegerArg(args[0], "i64")
			if err != nil {
				return nil, err
			}
			return externFastI32Result(fn(value)), nil
		}
	case func(int32, int32) int32:
		if !matchKinds("i32", "i32") || returnKind != "i32" {
			return nil
		}
		return func(_ *Interpreter, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("extern fast invoker expects 2 args, got %d", len(args))
			}
			left, err := externFastIntegerArg(args[0], "i32")
			if err != nil {
				return nil, err
			}
			right, err := externFastIntegerArg(args[1], "i32")
			if err != nil {
				return nil, err
			}
			return externFastI32Result(fn(int32(left), int32(right))), nil
		}
	case func(int32, int64) int32:
		if !matchKinds("i32", "i64") || returnKind != "i32" {
			return nil
		}
		return func(_ *Interpreter, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("extern fast invoker expects 2 args, got %d", len(args))
			}
			left, err := externFastIntegerArg(args[0], "i32")
			if err != nil {
				return nil, err
			}
			right, err := externFastIntegerArg(args[1], "i64")
			if err != nil {
				return nil, err
			}
			return externFastI32Result(fn(int32(left), right)), nil
		}
	case func(int32, int32, int32) int32:
		if !matchKinds("i32", "i32", "i32") || returnKind != "i32" {
			return nil
		}
		return func(_ *Interpreter, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 3 {
				return nil, fmt.Errorf("extern fast invoker expects 3 args, got %d", len(args))
			}
			first, err := externFastIntegerArg(args[0], "i32")
			if err != nil {
				return nil, err
			}
			second, err := externFastIntegerArg(args[1], "i32")
			if err != nil {
				return nil, err
			}
			third, err := externFastIntegerArg(args[2], "i32")
			if err != nil {
				return nil, err
			}
			return externFastI32Result(fn(int32(first), int32(second), int32(third))), nil
		}
	case func(int32, int32, int64) int32:
		if !matchKinds("i32", "i32", "i64") || returnKind != "i32" {
			return nil
		}
		return func(_ *Interpreter, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 3 {
				return nil, fmt.Errorf("extern fast invoker expects 3 args, got %d", len(args))
			}
			first, err := externFastIntegerArg(args[0], "i32")
			if err != nil {
				return nil, err
			}
			second, err := externFastIntegerArg(args[1], "i32")
			if err != nil {
				return nil, err
			}
			third, err := externFastIntegerArg(args[2], "i64")
			if err != nil {
				return nil, err
			}
			return externFastI32Result(fn(int32(first), int32(second), third)), nil
		}
	default:
		return nil
	}
}
