package interpreter

import (
	"fmt"
	"math/big"
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

	int64FromValue := func(val runtime.Value, label string) (int64, error) {
		switch v := val.(type) {
		case runtime.IntegerValue:
			if !v.Val.IsInt64() {
				return 0, fmt.Errorf("%s must fit in 64-bit integer", label)
			}
			return v.Val.Int64(), nil
		case *runtime.IntegerValue:
			if v == nil || v.Val == nil {
				return 0, fmt.Errorf("%s is nil", label)
			}
			if !v.Val.IsInt64() {
				return 0, fmt.Errorf("%s must fit in 64-bit integer", label)
			}
			return v.Val.Int64(), nil
		default:
			return 0, fmt.Errorf("%s must be an integer", label)
		}
	}

	stringFromBuiltin := runtime.NativeFunctionValue{
		Name:  "__able_String_from_builtin",
		Arity: 1,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("__able_String_from_builtin expects one argument")
			}
			var input string
			switch v := args[0].(type) {
			case runtime.StringValue:
				input = v.Val
			case *runtime.StringValue:
				if v == nil {
					return nil, fmt.Errorf("string argument is nil")
				}
				input = v.Val
			default:
				return nil, fmt.Errorf("argument must be a string")
			}
			data := []byte(input)
			elements := make([]runtime.Value, len(data))
			for idx, b := range data {
				elements[idx] = runtime.IntegerValue{
					Val:        big.NewInt(int64(b)),
					TypeSuffix: runtime.IntegerU8,
				}
			}
			return i.newArrayValue(elements, len(elements)), nil
		},
	}

	stringToBuiltin := runtime.NativeFunctionValue{
		Name:  "__able_String_to_builtin",
		Arity: 1,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("__able_String_to_builtin expects one argument")
			}
			arr, err := i.toArrayValue(args[0])
			if err != nil {
				return nil, fmt.Errorf("argument must be an array: %w", err)
			}
			if _, err := i.ensureArrayState(arr, 0); err != nil {
				return nil, err
			}
			bytes := make([]byte, len(arr.Elements))
			for idx, element := range arr.Elements {
				num, convErr := int64FromValue(element, fmt.Sprintf("array element %d", idx))
				if convErr != nil {
					return nil, convErr
				}
				if num < 0 || num > 0xff {
					return nil, fmt.Errorf("array element %d must be in range 0..255", idx)
				}
				bytes[idx] = byte(num)
			}
			if !utf8.Valid(bytes) {
				return nil, fmt.Errorf("invalid UTF-8 byte sequence")
			}
			return runtime.StringValue{Val: string(bytes)}, nil
		},
	}

	charFromCodepoint := runtime.NativeFunctionValue{
		Name:  "__able_char_from_codepoint",
		Arity: 1,
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
		Name:  "__able_char_to_codepoint",
		Arity: 1,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("__able_char_to_codepoint expects one argument")
			}
			switch v := args[0].(type) {
			case runtime.CharValue:
				return runtime.IntegerValue{Val: big.NewInt(int64(v.Val)), TypeSuffix: runtime.IntegerI32}, nil
			case *runtime.CharValue:
				if v == nil {
					return nil, fmt.Errorf("char argument is nil")
				}
				return runtime.IntegerValue{Val: big.NewInt(int64(v.Val)), TypeSuffix: runtime.IntegerI32}, nil
			default:
				return nil, fmt.Errorf("argument must be a char")
			}
		},
	}

	i.global.Define("__able_String_from_builtin", stringFromBuiltin)
	i.global.Define("__able_String_to_builtin", stringToBuiltin)
	i.global.Define("__able_char_from_codepoint", charFromCodepoint)
	i.global.Define("__able_char_to_codepoint", charToCodepoint)
	i.stringHostReady = true
}
