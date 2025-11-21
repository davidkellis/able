package interpreter

import (
	"fmt"
	"math"
	"math/big"
	"strings"
	"unicode/utf8"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
)

func (i *Interpreter) stringMember(str runtime.StringValue, member ast.Expression) (runtime.Value, error) {
	ident, ok := member.(*ast.Identifier)
	if !ok {
		return nil, fmt.Errorf("String member access expects identifier")
	}
	switch ident.Name {
	case "len_bytes":
		fn := runtime.NativeFunctionValue{
			Name:  "string.len_bytes",
			Arity: 0,
			Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("len_bytes expects only a receiver")
				}
				receiver, ok := args[0].(runtime.StringValue)
				if !ok {
					return nil, fmt.Errorf("len_bytes receiver must be a string")
				}
				return runtime.IntegerValue{Val: big.NewInt(int64(len(receiver.Val))), TypeSuffix: runtime.IntegerU64}, nil
			},
		}
		return &runtime.NativeBoundMethodValue{Receiver: str, Method: fn}, nil
	case "bytes":
		fn := runtime.NativeFunctionValue{
			Name:  "string.bytes",
			Arity: 0,
			Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("bytes expects only a receiver")
				}
				receiver, ok := args[0].(runtime.StringValue)
				if !ok {
					return nil, fmt.Errorf("bytes receiver must be a string")
				}
				data := []byte(receiver.Val)
				index := 0
				iter := runtime.NewIteratorValue(func() (runtime.Value, bool, error) {
					if index >= len(data) {
						return runtime.IteratorEnd, true, nil
					}
					b := data[index]
					index++
					return runtime.IntegerValue{Val: big.NewInt(int64(b)), TypeSuffix: runtime.IntegerU8}, false, nil
				}, nil)
				return iter, nil
			},
		}
		return &runtime.NativeBoundMethodValue{Receiver: str, Method: fn}, nil
	case "len_chars":
		fn := runtime.NativeFunctionValue{
			Name:  "string.len_chars",
			Arity: 0,
			Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("len_chars expects only a receiver")
				}
				receiver, ok := args[0].(runtime.StringValue)
				if !ok {
					return nil, fmt.Errorf("len_chars receiver must be a string")
				}
				return runtime.IntegerValue{Val: big.NewInt(int64(utf8.RuneCountInString(receiver.Val))), TypeSuffix: runtime.IntegerU64}, nil
			},
		}
		return &runtime.NativeBoundMethodValue{Receiver: str, Method: fn}, nil
	case "len_graphemes":
		fn := runtime.NativeFunctionValue{
			Name:  "string.len_graphemes",
			Arity: 0,
			Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("len_graphemes expects only a receiver")
				}
				receiver, ok := args[0].(runtime.StringValue)
				if !ok {
					return nil, fmt.Errorf("len_graphemes receiver must be a string")
				}
				// Grapheme support is approximated via code points until full segmentation lands.
				return runtime.IntegerValue{Val: big.NewInt(int64(utf8.RuneCountInString(receiver.Val))), TypeSuffix: runtime.IntegerU64}, nil
			},
		}
		return &runtime.NativeBoundMethodValue{Receiver: str, Method: fn}, nil
	case "substring":
		fn := runtime.NativeFunctionValue{
			Name:  "string.substring",
			Arity: 2,
			Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
				if len(args) < 2 || len(args) > 3 {
					return nil, fmt.Errorf("substring expects start and optional length")
				}
				receiver, ok := args[0].(runtime.StringValue)
				if !ok {
					return nil, fmt.Errorf("substring receiver must be a string")
				}
				start, err := toNonNegativeIndex(args[1], "start")
				if err != nil {
					return err, nil
				}
				var length *int
				if len(args) == 3 && !isNilRuntimeValue(args[2]) {
					val, convErr := toNonNegativeIndex(args[2], "length")
					if convErr != nil {
						return convErr, nil
					}
					length = &val
				}
				runes := []rune(receiver.Val)
				if start > len(runes) {
					return makeRangeError("substring start out of range"), nil
				}
				end := len(runes)
				if length != nil {
					if start+*length > len(runes) {
						return makeRangeError("substring range out of bounds"), nil
					}
					end = start + *length
				}
				result := string(runes[start:end])
				return runtime.StringValue{Val: result}, nil
			},
		}
		return &runtime.NativeBoundMethodValue{Receiver: str, Method: fn}, nil
	case "split":
		fn := runtime.NativeFunctionValue{
			Name:  "string.split",
			Arity: 1,
			Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
				if len(args) != 2 {
					return nil, fmt.Errorf("split expects a receiver and delimiter")
				}
				receiver, ok := args[0].(runtime.StringValue)
				if !ok {
					return nil, fmt.Errorf("split receiver must be a string")
				}
				delimiter, ok := args[1].(runtime.StringValue)
				if !ok {
					return nil, fmt.Errorf("split delimiter must be a string")
				}
				var parts []string
				if delimiter.Val == "" {
					for _, r := range receiver.Val {
						parts = append(parts, string(r))
					}
				} else {
					parts = strings.Split(receiver.Val, delimiter.Val)
				}
				elements := make([]runtime.Value, 0, len(parts))
				for _, p := range parts {
					elements = append(elements, runtime.StringValue{Val: p})
				}
				return &runtime.ArrayValue{Elements: elements}, nil
			},
		}
		return &runtime.NativeBoundMethodValue{Receiver: str, Method: fn}, nil
	case "replace":
		fn := runtime.NativeFunctionValue{
			Name:  "string.replace",
			Arity: 2,
			Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
				if len(args) != 3 {
					return nil, fmt.Errorf("replace expects a receiver, search, and replacement")
				}
				receiver, ok := args[0].(runtime.StringValue)
				if !ok {
					return nil, fmt.Errorf("replace receiver must be a string")
				}
				oldVal, ok := args[1].(runtime.StringValue)
				if !ok {
					return nil, fmt.Errorf("replace target must be a string")
				}
				newVal, ok := args[2].(runtime.StringValue)
				if !ok {
					return nil, fmt.Errorf("replace replacement must be a string")
				}
				return runtime.StringValue{Val: strings.ReplaceAll(receiver.Val, oldVal.Val, newVal.Val)}, nil
			},
		}
		return &runtime.NativeBoundMethodValue{Receiver: str, Method: fn}, nil
	case "starts_with":
		fn := runtime.NativeFunctionValue{
			Name:  "string.starts_with",
			Arity: 1,
			Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
				if len(args) != 2 {
					return nil, fmt.Errorf("starts_with expects a receiver and prefix")
				}
				receiver, ok := args[0].(runtime.StringValue)
				if !ok {
					return nil, fmt.Errorf("starts_with receiver must be a string")
				}
				prefix, ok := args[1].(runtime.StringValue)
				if !ok {
					return nil, fmt.Errorf("starts_with prefix must be a string")
				}
				return runtime.BoolValue{Val: strings.HasPrefix(receiver.Val, prefix.Val)}, nil
			},
		}
		return &runtime.NativeBoundMethodValue{Receiver: str, Method: fn}, nil
	case "ends_with":
		fn := runtime.NativeFunctionValue{
			Name:  "string.ends_with",
			Arity: 1,
			Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
				if len(args) != 2 {
					return nil, fmt.Errorf("ends_with expects a receiver and suffix")
				}
				receiver, ok := args[0].(runtime.StringValue)
				if !ok {
					return nil, fmt.Errorf("ends_with receiver must be a string")
				}
				suffix, ok := args[1].(runtime.StringValue)
				if !ok {
					return nil, fmt.Errorf("ends_with suffix must be a string")
				}
				return runtime.BoolValue{Val: strings.HasSuffix(receiver.Val, suffix.Val)}, nil
			},
		}
		return &runtime.NativeBoundMethodValue{Receiver: str, Method: fn}, nil
	case "chars":
		fn := runtime.NativeFunctionValue{
			Name:  "string.chars",
			Arity: 0,
			Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("chars expects only a receiver")
				}
				receiver, ok := args[0].(runtime.StringValue)
				if !ok {
					return nil, fmt.Errorf("chars receiver must be a string")
				}
				data := receiver.Val
				index := 0
				iter := runtime.NewIteratorValue(func() (runtime.Value, bool, error) {
					if index >= len(data) {
						return runtime.IteratorEnd, true, nil
					}
					r, size := utf8.DecodeRuneInString(data[index:])
					if r == utf8.RuneError && size <= 1 {
						return nil, true, fmt.Errorf("invalid UTF-8 in string.chars")
					}
					index += size
					return runtime.CharValue{Val: r}, false, nil
				}, nil)
				return iter, nil
			},
		}
		return &runtime.NativeBoundMethodValue{Receiver: str, Method: fn}, nil
	case "graphemes":
		fn := runtime.NativeFunctionValue{
			Name:  "string.graphemes",
			Arity: 0,
			Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("graphemes expects only a receiver")
				}
				receiver, ok := args[0].(runtime.StringValue)
				if !ok {
					return nil, fmt.Errorf("graphemes receiver must be a string")
				}
				runes := []rune(receiver.Val)
				index := 0
				iter := runtime.NewIteratorValue(func() (runtime.Value, bool, error) {
					if index >= len(runes) {
						return runtime.IteratorEnd, true, nil
					}
					r := runes[index]
					index++
					return runtime.StringValue{Val: string(r)}, false, nil
				}, nil)
				return iter, nil
			},
		}
		return &runtime.NativeBoundMethodValue{Receiver: str, Method: fn}, nil
	default:
		return nil, fmt.Errorf("String has no member '%s'", ident.Name)
	}
}

func makeRangeError(message string) runtime.Value {
	return runtime.ErrorValue{
		TypeName: ast.NewIdentifier("RangeError"),
		Message:  message,
	}
}

func toNonNegativeIndex(val runtime.Value, label string) (int, runtime.Value) {
	switch v := val.(type) {
	case runtime.IntegerValue:
		if v.Val == nil || v.Val.Sign() < 0 {
			return 0, makeRangeError(fmt.Sprintf("%s must be non-negative", label))
		}
		if !v.Val.IsInt64() {
			return 0, makeRangeError(fmt.Sprintf("%s is out of range", label))
		}
		res := v.Val.Int64()
		if res > math.MaxInt {
			return 0, makeRangeError(fmt.Sprintf("%s is out of range", label))
		}
		return int(res), nil
	default:
		return 0, makeRangeError(fmt.Sprintf("%s must be an integer", label))
	}
}
