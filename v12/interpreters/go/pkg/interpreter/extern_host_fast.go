package interpreter

import (
	"fmt"
	"reflect"
	"slices"
	"strings"
	"sync/atomic"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func externStringArg(value runtime.Value) (string, bool) {
	for {
		switch v := value.(type) {
		case runtime.StringValue:
			return v.Val, true
		case *runtime.StringValue:
			if v == nil {
				return "", false
			}
			return v.Val, true
		case runtime.InterfaceValue:
			value = v.Underlying
			continue
		case *runtime.InterfaceValue:
			if v == nil {
				return "", false
			}
			value = v.Underlying
			continue
		default:
			return "", false
		}
	}
}

func externStringSliceResult(values []string) runtime.Value {
	if len(values) == 0 {
		return &runtime.ArrayValue{Elements: []runtime.Value{}}
	}
	elements := make([]runtime.Value, len(values))
	for idx, value := range values {
		elements[idx] = runtime.StringValue{Val: value}
	}
	return &runtime.ArrayValue{Elements: elements}
}

type externStringSliceCacheEntry struct {
	snapshot []string
	boxed    []runtime.Value
}

type externStringSliceCache struct {
	entry atomic.Pointer[externStringSliceCacheEntry]
}

func externCloneValueSlice(values []runtime.Value) []runtime.Value {
	if len(values) == 0 {
		return []runtime.Value{}
	}
	cloned := make([]runtime.Value, len(values))
	copy(cloned, values)
	return cloned
}

func externStringSliceTemplate(values []string) []runtime.Value {
	if len(values) == 0 {
		return []runtime.Value{}
	}
	elements := make([]runtime.Value, len(values))
	for idx, value := range values {
		elements[idx] = runtime.StringValue{Val: value}
	}
	return elements
}

func (c *externStringSliceCache) result(values []string) runtime.Value {
	if len(values) == 0 {
		return &runtime.ArrayValue{Elements: []runtime.Value{}}
	}
	if entry := c.entry.Load(); entry != nil &&
		len(entry.snapshot) == len(values) &&
		slices.Equal(entry.snapshot, values) {
		return &runtime.ArrayValue{Elements: externCloneValueSlice(entry.boxed)}
	}
	snapshot := append([]string(nil), values...)
	boxed := externStringSliceTemplate(snapshot)
	c.entry.Store(&externStringSliceCacheEntry{
		snapshot: snapshot,
		boxed:    boxed,
	})
	return &runtime.ArrayValue{Elements: externCloneValueSlice(boxed)}
}

func externSimpleTypeName(expr ast.TypeExpression) string {
	simple, ok := expr.(*ast.SimpleTypeExpression)
	if !ok || simple == nil || simple.Name == nil {
		return ""
	}
	return normalizeKernelAliasName(simple.Name.Name)
}

func buildExternFastInvoker(def *ast.ExternFunctionBody, raw any) externHostInvoker {
	if def == nil || def.Signature == nil {
		return nil
	}
	paramCount := len(def.Signature.Params)
	returnType := externSimpleTypeName(def.Signature.ReturnType)
	externName := ""
	if def.Signature.ID != nil {
		externName = def.Signature.ID.Name
	}
	var stringSliceCache externStringSliceCache
	switch paramCount {
	case 1:
		if externSimpleTypeName(def.Signature.Params[0].ParamType) != "String" {
			return nil
		}
		switch fn := raw.(type) {
		case func(string) int32:
			if returnType != "i32" {
				return nil
			}
			return func(_ *Interpreter, args []runtime.Value) (runtime.Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("extern fast invoker expects 1 arg, got %d", len(args))
				}
				value, ok := externStringArg(args[0])
				if !ok {
					return nil, fmt.Errorf("extern fast invoker expected String argument")
				}
				return boxedOrSmallIntegerValue(runtime.IntegerI32, int64(fn(value))), nil
			}
		case func(string) bool:
			if returnType != "bool" {
				return nil
			}
			return func(_ *Interpreter, args []runtime.Value) (runtime.Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("extern fast invoker expects 1 arg, got %d", len(args))
				}
				value, ok := externStringArg(args[0])
				if !ok {
					return nil, fmt.Errorf("extern fast invoker expected String argument")
				}
				return runtime.BoolValue{Val: fn(value)}, nil
			}
		case func(string) string:
			if returnType != "String" {
				return nil
			}
			return func(_ *Interpreter, args []runtime.Value) (runtime.Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("extern fast invoker expects 1 arg, got %d", len(args))
				}
				value, ok := externStringArg(args[0])
				if !ok {
					return nil, fmt.Errorf("extern fast invoker expected String argument")
				}
				return runtime.StringValue{Val: fn(value)}, nil
			}
		case func(string) []string:
			base, ok := def.Signature.ReturnType.(*ast.GenericTypeExpression)
			if !ok || externSimpleTypeName(base.Base) != "Array" || len(base.Arguments) != 1 || externSimpleTypeName(base.Arguments[0]) != "String" {
				return nil
			}
			return func(_ *Interpreter, args []runtime.Value) (runtime.Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("extern fast invoker expects 1 arg, got %d", len(args))
				}
				value, ok := externStringArg(args[0])
				if !ok {
					return nil, fmt.Errorf("extern fast invoker expected String argument")
				}
				return stringSliceCache.result(fn(value)), nil
			}
		case func(string) interface{}:
			return func(i *Interpreter, args []runtime.Value) (runtime.Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("extern fast invoker expects 1 arg, got %d", len(args))
				}
				value, ok := externStringArg(args[0])
				if !ok {
					return nil, fmt.Errorf("extern fast invoker expected String argument")
				}
				result := fn(value)
				if externUnionHasArrayStringMember(def.Signature.ReturnType) {
					if lines, ok := result.([]string); ok {
						return stringSliceCache.result(lines), nil
					}
				}
				if i == nil {
					return nil, fmt.Errorf("extern fast invoker requires interpreter fallback")
				}
				var reflected reflect.Value
				if result != nil {
					reflected = reflect.ValueOf(result)
				}
				return i.fromHostValue(def.Signature.ReturnType, reflected)
			}
		}
	case 2:
		if externSimpleTypeName(def.Signature.Params[0].ParamType) != "String" || externSimpleTypeName(def.Signature.Params[1].ParamType) != "String" {
			return nil
		}
		switch fn := raw.(type) {
		case func(string, string) bool:
			if returnType != "bool" {
				return nil
			}
			return func(_ *Interpreter, args []runtime.Value) (runtime.Value, error) {
				if len(args) != 2 {
					return nil, fmt.Errorf("extern fast invoker expects 2 args, got %d", len(args))
				}
				left, ok := externStringArg(args[0])
				if !ok {
					return nil, fmt.Errorf("extern fast invoker expected String argument")
				}
				right, ok := externStringArg(args[1])
				if !ok {
					return nil, fmt.Errorf("extern fast invoker expected String argument")
				}
				return runtime.BoolValue{Val: fn(left, right)}, nil
			}
		case func(string, string) string:
			if returnType != "String" {
				return nil
			}
			return func(_ *Interpreter, args []runtime.Value) (runtime.Value, error) {
				if len(args) != 2 {
					return nil, fmt.Errorf("extern fast invoker expects 2 args, got %d", len(args))
				}
				left, ok := externStringArg(args[0])
				if !ok {
					return nil, fmt.Errorf("extern fast invoker expected String argument")
				}
				right, ok := externStringArg(args[1])
				if !ok {
					return nil, fmt.Errorf("extern fast invoker expected String argument")
				}
				return runtime.StringValue{Val: fn(left, right)}, nil
			}
		}
	case 3:
		if externSimpleTypeName(def.Signature.Params[0].ParamType) != "String" || externSimpleTypeName(def.Signature.Params[1].ParamType) != "String" || externSimpleTypeName(def.Signature.Params[2].ParamType) != "String" {
			return nil
		}
		if fn, ok := raw.(func(string, string, string) string); ok && returnType == "String" {
			return func(_ *Interpreter, args []runtime.Value) (runtime.Value, error) {
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
				third, ok := externStringArg(args[2])
				if !ok {
					return nil, fmt.Errorf("extern fast invoker expected String argument")
				}
				if externName == "string_replace_fast" {
					if second == "" || !strings.Contains(first, second) {
						return runtime.StringValue{Val: first}, nil
					}
				}
				return runtime.StringValue{Val: fn(first, second, third)}, nil
			}
		}
	}
	return nil
}
