package interpreter

import (
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
		if intVal.Val.Int64() != expected {
			t.Fatalf("element %d = %d, want %d", idx, intVal.Val.Int64(), expected)
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
