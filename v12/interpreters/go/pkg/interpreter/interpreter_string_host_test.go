package interpreter

import (
	"math/big"
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestStringFromBuiltinProducesBytes(t *testing.T) {
	interp := New()
	global := interp.GlobalEnvironment()

	val, err := interp.evaluateExpression(ast.Call("__able_String_from_builtin", ast.Str("Hi")), global)
	if err != nil {
		t.Fatalf("evaluation failed: %v", err)
	}
	arr, ok := val.(*runtime.ArrayValue)
	if !ok {
		t.Fatalf("expected ArrayValue, got %#v", val)
	}
	if len(arr.Elements) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(arr.Elements))
	}
	for idx, expected := range []int64{72, 105} {
		intVal, ok := arr.Elements[idx].(runtime.IntegerValue)
		if !ok {
			t.Fatalf("element %d type = %T, want IntegerValue", idx, arr.Elements[idx])
		}
		if intVal.BigInt().Int64() != expected {
			t.Fatalf("element %d = %d, want %d", idx, intVal.BigInt().Int64(), expected)
		}
	}
}

func TestStringFromBuiltinAcceptsStructString(t *testing.T) {
	interp := New()
	global := interp.GlobalEnvironment()

	bytes := []runtime.Value{
		runtime.IntegerValue{Val: big.NewInt(72), TypeSuffix: runtime.IntegerU8},
		runtime.IntegerValue{Val: big.NewInt(105), TypeSuffix: runtime.IntegerU8},
	}
	byteArr := interp.newArrayValue(bytes, len(bytes))
	definition := &runtime.StructDefinitionValue{
		Node: ast.StructDef("String", nil, ast.StructKindNamed, nil, nil, false),
	}
	inst := &runtime.StructInstanceValue{
		Definition: definition,
		Fields: map[string]runtime.Value{
			"bytes":     byteArr,
			"len_bytes": runtime.IntegerValue{Val: big.NewInt(2), TypeSuffix: runtime.IntegerI32},
		},
	}
	global.Define("value", inst)

	val, err := interp.evaluateExpression(ast.Call("__able_String_from_builtin", ast.ID("value")), global)
	if err != nil {
		t.Fatalf("evaluation failed: %v", err)
	}
	arr, ok := val.(*runtime.ArrayValue)
	if !ok {
		t.Fatalf("expected ArrayValue, got %#v", val)
	}
	if len(arr.Elements) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(arr.Elements))
	}
	for idx, expected := range []int64{72, 105} {
		intVal, ok := arr.Elements[idx].(runtime.IntegerValue)
		if !ok {
			t.Fatalf("element %d type = %T, want IntegerValue", idx, arr.Elements[idx])
		}
		if intVal.BigInt().Int64() != expected {
			t.Fatalf("element %d = %d, want %d", idx, intVal.BigInt().Int64(), expected)
		}
	}
}

func TestStringToBuiltinDecodesUTF8(t *testing.T) {
	interp := New()
	global := interp.GlobalEnvironment()

	val, err := interp.evaluateExpression(
		ast.Call("__able_String_to_builtin", ast.Arr(ast.Int(0xE2), ast.Int(0x82), ast.Int(0xAC))),
		global,
	)
	if err != nil {
		t.Fatalf("evaluation failed: %v", err)
	}
	strVal, ok := val.(runtime.StringValue)
	if !ok {
		t.Fatalf("expected StringValue, got %#v", val)
	}
	if strVal.Val != "€" {
		t.Fatalf("expected '€', got %q", strVal.Val)
	}
}

func TestStringToBuiltinUsesMonoU8HandleWithoutDeopt(t *testing.T) {
	interp := New()
	global := interp.GlobalEnvironment()

	bytes := interp.newU8ArrayValueFromString("Able")
	global.Define("bytes", bytes)
	val, err := interp.evaluateExpression(ast.Call("__able_String_to_builtin", ast.ID("bytes")), global)
	if err != nil {
		t.Fatalf("evaluation failed: %v", err)
	}
	strVal, ok := val.(runtime.StringValue)
	if !ok {
		t.Fatalf("expected StringValue, got %#v", val)
	}
	if strVal.Val != "Able" {
		t.Fatalf("expected Able, got %q", strVal.Val)
	}
	if bytes.Elements != nil || bytes.State != nil {
		t.Fatalf("mono bytes array was materialized: elements=%#v state=%#v", bytes.Elements, bytes.State)
	}
	got, ok, err := runtime.ArrayStoreMonoBorrowedU8BytesIfAvailable(bytes.Handle)
	if err != nil {
		t.Fatalf("mono bytes after conversion: %v", err)
	}
	if !ok || string(got) != "Able" {
		t.Fatalf("mono bytes after conversion = %q/%v, want %q/true", string(got), ok, "Able")
	}
}

func TestStringifyStdlibStringStructUsesBuiltinBytes(t *testing.T) {
	interp := New()
	value := &runtime.StructInstanceValue{
		Definition: &runtime.StructDefinitionValue{
			Node: ast.StructDef("String", nil, ast.StructKindNamed, nil, nil, false),
		},
		Fields: map[string]runtime.Value{
			"bytes": interp.newArrayValue([]runtime.Value{
				runtime.NewSmallInt(79, runtime.IntegerU8),
				runtime.NewSmallInt(75, runtime.IntegerU8),
			}, 2),
			"len_bytes": runtime.NewSmallInt(2, runtime.IntegerI32),
		},
	}

	got, err := interp.Stringify(value, nil)
	if err != nil {
		t.Fatalf("Stringify failed: %v", err)
	}
	if got != "OK" {
		t.Fatalf("Stringify(stdlib String) = %q, want %q", got, "OK")
	}
}

func TestStringifyMaterializesRawIntegerValue(t *testing.T) {
	interp := NewBytecode()

	got, err := interp.Stringify(bytecodeRawI32SlotCachedValue(7), nil)
	if err != nil {
		t.Fatalf("Stringify failed: %v", err)
	}
	if got != "7" {
		t.Fatalf("Stringify(raw i32) = %q, want %q", got, "7")
	}
}

func TestCharFromCodepoint(t *testing.T) {
	interp := New()
	global := interp.GlobalEnvironment()

	val, err := interp.evaluateExpression(ast.Call("__able_char_from_codepoint", ast.Int(0x1F600)), global)
	if err != nil {
		t.Fatalf("evaluation failed: %v", err)
	}
	charVal, ok := val.(runtime.CharValue)
	if !ok {
		t.Fatalf("expected CharValue, got %#v", val)
	}
	if charVal.Val != '😀' {
		t.Fatalf("expected 😀, got %q", charVal.Val)
	}
}

func TestCharToCodepoint(t *testing.T) {
	interp := New()
	global := interp.GlobalEnvironment()

	for _, test := range []struct {
		name string
		char string
		want int64
	}{
		{name: "ascii", char: "K", want: 'K'},
		{name: "non_ascii", char: "😀", want: '😀'},
	} {
		t.Run(test.name, func(t *testing.T) {
			val, err := interp.evaluateExpression(ast.Call("__able_char_to_codepoint", ast.Chr(test.char)), global)
			if err != nil {
				t.Fatalf("evaluation failed: %v", err)
			}
			integer, ok := val.(runtime.IntegerValue)
			if !ok {
				t.Fatalf("expected IntegerValue, got %#v", val)
			}
			got, ok := integer.ToInt64()
			if !ok || got != test.want {
				t.Fatalf("codepoint(%q) = %d, %v; want %d", test.char, got, ok, test.want)
			}
		})
	}
}

func TestCharSimpleFoldNext(t *testing.T) {
	interp := New()
	global := interp.GlobalEnvironment()

	val, err := interp.evaluateExpression(ast.Call("__able_char_simple_fold_next", ast.Chr("K")), global)
	if err != nil {
		t.Fatalf("evaluation failed: %v", err)
	}
	charVal, ok := val.(runtime.CharValue)
	if !ok {
		t.Fatalf("expected CharValue, got %#v", val)
	}
	if charVal.Val != 'k' {
		t.Fatalf("SimpleFold(K) = %q, want %q", charVal.Val, 'k')
	}
}

func TestStringHostBuiltinCallMetadata(t *testing.T) {
	interp := New()
	global := interp.GlobalEnvironment()

	for _, name := range []string{
		"__able_String_from_builtin",
		"__able_String_to_builtin",
		"__able_char_from_codepoint",
		"__able_char_to_codepoint",
		"__able_char_simple_fold_next",
	} {
		val, err := global.Get(name)
		if err != nil {
			t.Fatalf("missing builtin %s: %v", name, err)
		}
		fn, ok := val.(runtime.NativeFunctionValue)
		if !ok {
			t.Fatalf("builtin %s type = %T, want runtime.NativeFunctionValue", name, val)
		}
		if !fn.BorrowArgs {
			t.Fatalf("%s should borrow args", name)
		}
		if !fn.SkipContext {
			t.Fatalf("%s should skip native call context", name)
		}
	}
}
