package interpreter

import (
	"fmt"
	"unicode"
	"unicode/utf8"

	"able/interpreter-go/pkg/runtime"
)

func (i *Interpreter) ensureStringHostBuiltins() {
	if i.stringHostReady {
		return
	}
	i.initStringHostBuiltins()
}

func (i *Interpreter) initStringHostBuiltins() {
	if i.stringHostReady {
		return
	}

	int64ValueInfo := func(val runtime.Value) (int64, bool, bool) {
		switch v := val.(type) {
		case runtime.IntegerValue:
			if n, ok := v.ToInt64(); ok {
				return n, true, true
			}
			return 0, true, false
		case *runtime.IntegerValue:
			if v == nil {
				return 0, false, false
			}
			if n, ok := v.ToInt64(); ok {
				return n, true, true
			}
			return 0, true, false
		default:
			return 0, false, false
		}
	}

	int64FromValue := func(val runtime.Value, label string) (int64, error) {
		if n, ok, fits := int64ValueInfo(val); ok {
			if fits {
				return n, nil
			}
			return 0, fmt.Errorf("%s must fit in 64-bit integer", label)
		}
		if val == nil {
			return 0, fmt.Errorf("%s is nil", label)
		}
		return 0, fmt.Errorf("%s must be an integer", label)
	}

	arrayForStringBytes := func(value runtime.Value) (*runtime.ArrayValue, error) {
		for {
			switch v := value.(type) {
			case *runtime.InterfaceValue:
				if v == nil {
					value = nil
				} else {
					value = v.Underlying
				}
			case runtime.InterfaceValue:
				value = v.Underlying
			default:
				goto resolvedArrayValue
			}
		}
	resolvedArrayValue:
		switch v := value.(type) {
		case *runtime.ArrayValue:
			if v == nil {
				return nil, fmt.Errorf("array argument is nil")
			}
			return v, nil
		default:
			return i.toArrayValue(value)
		}
	}

	stringBytesFromStruct := func(inst *runtime.StructInstanceValue) (*runtime.ArrayValue, error) {
		if inst == nil || inst.Definition == nil || inst.Definition.Node == nil || inst.Definition.Node.ID == nil {
			return nil, fmt.Errorf("argument must be a string")
		}
		if inst.Definition.Node.ID.Name != "String" {
			return nil, fmt.Errorf("argument must be a string")
		}
		var bytesVal runtime.Value
		if inst.Fields != nil {
			if field, ok := inst.Fields["bytes"]; ok {
				bytesVal = field
			}
		}
		if bytesVal == nil && len(inst.Positional) > 0 {
			bytesVal = inst.Positional[0]
		}
		if bytesVal == nil {
			return nil, fmt.Errorf("string bytes are missing")
		}
		for {
			switch v := bytesVal.(type) {
			case *runtime.InterfaceValue:
				if v == nil {
					bytesVal = nil
				} else {
					bytesVal = v.Underlying
				}
			case runtime.InterfaceValue:
				bytesVal = v.Underlying
			default:
				goto resolvedBytes
			}
		}
	resolvedBytes:
		if bytesVal == nil {
			return nil, fmt.Errorf("string bytes are missing")
		}
		arr, err := arrayForStringBytes(bytesVal)
		if err != nil {
			kind := "<nil>"
			if bytesVal != nil {
				kind = fmt.Sprintf("%v", bytesVal.Kind())
			}
			return nil, fmt.Errorf("string bytes must be an array (got %T kind=%s): %w", bytesVal, kind, err)
		}
		state, err := i.ensureArrayState(arr, 0)
		if err != nil {
			return nil, err
		}
		cloned := make([]runtime.Value, len(state.Values))
		copy(cloned, state.Values)
		return i.newArrayValue(cloned, len(cloned)), nil
	}

	stringFromBuiltin := runtime.NativeFunctionValue{
		Name:        "__able_String_from_builtin",
		Arity:       1,
		BorrowArgs:  true,
		SkipContext: true,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("__able_String_from_builtin expects one argument")
			}
			val := args[0]
			for {
				if iface, ok := val.(*runtime.InterfaceValue); ok && iface != nil {
					val = iface.Underlying
					continue
				}
				if iface, ok := val.(runtime.InterfaceValue); ok {
					val = iface.Underlying
					continue
				}
				break
			}
			var input string
			switch v := val.(type) {
			case runtime.StringValue:
				input = v.Val
			case *runtime.StringValue:
				if v == nil {
					return nil, fmt.Errorf("string argument is nil")
				}
				input = v.Val
			case *runtime.StructInstanceValue:
				return stringBytesFromStruct(v)
			default:
				return nil, fmt.Errorf("argument must be a string")
			}
			data := []byte(input)
			i.recordBytecodeStringFromBuiltin(len(data))
			elements := make([]runtime.Value, len(data))
			for idx, b := range data {
				elements[idx] = runtime.NewSmallInt(int64(b), runtime.IntegerU8)
			}
			return i.newArrayValue(elements, len(elements)), nil
		},
	}

	stringToBuiltin := runtime.NativeFunctionValue{
		Name:        "__able_String_to_builtin",
		Arity:       1,
		BorrowArgs:  true,
		SkipContext: true,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("__able_String_to_builtin expects one argument")
			}
			arr, err := arrayForStringBytes(args[0])
			if err != nil {
				return nil, fmt.Errorf("argument must be an array: %w", err)
			}
			if bytes, ok, err := runtime.ArrayStoreMonoBorrowedU8BytesIfAvailable(arrayValueHandle(arr)); err != nil {
				return nil, err
			} else if ok {
				valid := utf8.Valid(bytes)
				i.recordBytecodeStringToBuiltin(len(bytes), true, valid)
				if !valid {
					return nil, fmt.Errorf("invalid UTF-8 byte sequence")
				}
				return runtime.StringValue{Val: string(bytes)}, nil
			}
			if _, err := i.ensureArrayState(arr, 0); err != nil {
				return nil, err
			}
			bytes := make([]byte, len(arr.Elements))
			for idx, element := range arr.Elements {
				num, ok, fits := int64ValueInfo(element)
				if !ok {
					return nil, fmt.Errorf("array element %d must be an integer", idx)
				}
				if !fits {
					return nil, fmt.Errorf("array element %d must fit in 64-bit integer", idx)
				}
				if num < 0 || num > 0xff {
					return nil, fmt.Errorf("array element %d must be in range 0..255", idx)
				}
				bytes[idx] = byte(num)
			}
			valid := utf8.Valid(bytes)
			i.recordBytecodeStringToBuiltin(len(bytes), false, valid)
			if !valid {
				return nil, fmt.Errorf("invalid UTF-8 byte sequence")
			}
			return runtime.StringValue{Val: string(bytes)}, nil
		},
	}

	charFromCodepoint := runtime.NativeFunctionValue{
		Name:        "__able_char_from_codepoint",
		Arity:       1,
		BorrowArgs:  true,
		SkipContext: true,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("__able_char_from_codepoint expects one argument")
			}
			codepoint, err := int64FromValue(args[0], "codepoint")
			if err != nil {
				return nil, err
			}
			if codepoint < 0 || codepoint > utf8.MaxRune {
				return nil, fmt.Errorf("codepoint must be within Unicode range 0..0x10FFFF")
			}
			r := rune(codepoint)
			if !utf8.ValidRune(r) {
				return nil, fmt.Errorf("invalid Unicode codepoint %d", codepoint)
			}
			return runtime.CharValue{Val: r}, nil
		},
	}

	charToCodepoint := runtime.NativeFunctionValue{
		Name:        "__able_char_to_codepoint",
		Arity:       1,
		BorrowArgs:  true,
		SkipContext: true,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("__able_char_to_codepoint expects one argument")
			}
			switch v := args[0].(type) {
			case runtime.CharValue:
				return boxedOrSmallIntegerValue(runtime.IntegerI32, int64(v.Val)), nil
			case *runtime.CharValue:
				if v == nil {
					return nil, fmt.Errorf("char argument is nil")
				}
				return boxedOrSmallIntegerValue(runtime.IntegerI32, int64(v.Val)), nil
			default:
				return nil, fmt.Errorf("argument must be a char")
			}
		},
	}

	charSimpleFoldNext := runtime.NativeFunctionValue{
		Name:        "__able_char_simple_fold_next",
		Arity:       1,
		BorrowArgs:  true,
		SkipContext: true,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("__able_char_simple_fold_next expects one argument")
			}
			switch v := args[0].(type) {
			case runtime.CharValue:
				return runtime.CharValue{Val: unicode.SimpleFold(v.Val)}, nil
			case *runtime.CharValue:
				if v == nil {
					return nil, fmt.Errorf("char argument is nil")
				}
				return runtime.CharValue{Val: unicode.SimpleFold(v.Val)}, nil
			default:
				return nil, fmt.Errorf("argument must be a char")
			}
		},
	}

	i.global.Define("__able_String_from_builtin", stringFromBuiltin)
	i.global.Define("__able_String_to_builtin", stringToBuiltin)
	i.global.Define("__able_char_from_codepoint", charFromCodepoint)
	i.global.Define("__able_char_to_codepoint", charToCodepoint)
	i.global.Define("__able_char_simple_fold_next", charSimpleFoldNext)
	i.stringHostReady = true
}
