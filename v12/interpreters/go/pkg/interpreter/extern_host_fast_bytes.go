package interpreter

import (
	"fmt"
	"reflect"
	"strings"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func externUsesBorrowedU8ArrayArg(def *ast.ExternFunctionBody) bool {
	if def == nil {
		return false
	}
	return strings.Contains(def.Body, "able_borrowed_bytes(")
}

func externU8ArrayArg(i *Interpreter, value runtime.Value, borrowed bool) ([]byte, bool) {
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
	return externU8ArrayBytes(i, arr, borrowed)
}

func externU8ArrayBytes(i *Interpreter, arr *runtime.ArrayValue, borrowed bool) ([]byte, bool) {
	if arr == nil {
		return nil, false
	}
	handle := arr.Handle
	if handle == 0 {
		handle = arr.TrackedHandle
	}
	if handle != 0 {
		var (
			bytes []byte
			ok    bool
			err   error
		)
		if borrowed {
			bytes, ok, err = runtime.ArrayStoreMonoBorrowedU8BytesIfAvailable(handle)
		} else {
			bytes, ok, err = runtime.ArrayStoreMonoU8BytesIfAvailable(handle)
		}
		if err != nil {
			return nil, false
		}
		if ok {
			return bytes, true
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
	bytes := make([]byte, len(elements))
	for idx, elem := range elements {
		num, err := toInt64(elem)
		if err != nil || num < 0 || num > 0xff {
			return nil, false
		}
		bytes[idx] = byte(num)
	}
	return bytes, true
}

func buildExternPrimitiveByteArrayFastInvoker(def *ast.ExternFunctionBody, raw any) externHostInvoker {
	if def == nil || def.Signature == nil || len(def.Signature.Params) != 1 {
		return nil
	}
	if !externIsArrayU8Type(def.Signature.Params[0].ParamType) {
		return nil
	}
	borrowed := externUsesBorrowedU8ArrayArg(def)
	returnType := def.Signature.ReturnType
	switch fn := raw.(type) {
	case func([]byte) string:
		if externSimpleTypeName(returnType) != "String" {
			return nil
		}
		return func(i *Interpreter, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("extern fast invoker expects 1 arg, got %d", len(args))
			}
			value, ok := externU8ArrayArg(i, args[0], borrowed)
			if !ok {
				return nil, fmt.Errorf("extern fast invoker expected Array u8 argument")
			}
			return runtime.StringValue{Val: fn(value)}, nil
		}
	case func([]byte) []byte:
		if !externIsArrayU8Type(returnType) {
			return nil
		}
		return func(i *Interpreter, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("extern fast invoker expects 1 arg, got %d", len(args))
			}
			value, ok := externU8ArrayArg(i, args[0], borrowed)
			if !ok {
				return nil, fmt.Errorf("extern fast invoker expected Array u8 argument")
			}
			result := fn(value)
			if i != nil {
				return i.newOwnedU8ArrayValueFromBytes(result), nil
			}
			return runtime.ArrayStoreMonoValueFromOwnedU8Bytes(result), nil
		}
	case func([]byte) interface{}:
		return func(i *Interpreter, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("extern fast invoker expects 1 arg, got %d", len(args))
			}
			value, ok := externU8ArrayArg(i, args[0], borrowed)
			if !ok {
				return nil, fmt.Errorf("extern fast invoker expected Array u8 argument")
			}
			result := fn(value)
			if externUnionHasArrayU8Member(returnType) {
				if bytes, ok := externReflectU8SliceBytes(reflect.ValueOf(result)); ok {
					if i != nil {
						return i.newOwnedU8ArrayValueFromBytes(bytes), nil
					}
					return runtime.ArrayStoreMonoValueFromOwnedU8Bytes(bytes), nil
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
