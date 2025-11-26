package interpreter

import (
	"strings"
	"testing"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
)

func TestArrayHelpersRequireStdlib(t *testing.T) {
	interp := New()
	interp.ensureArrayBuiltins()
	arr := &runtime.ArrayValue{Elements: []runtime.Value{runtime.NilValue{}}}

	if _, err := interp.arrayMemberWithOverrides(arr, ast.NewIdentifier("size"), interp.GlobalEnvironment()); err == nil {
		t.Fatalf("expected size to be unavailable without stdlib import")
	}
	if _, err := interp.arrayMemberWithOverrides(arr, ast.NewIdentifier("push"), interp.GlobalEnvironment()); err == nil {
		t.Fatalf("expected push to be unavailable without stdlib import")
	}

	iterVal, err := interp.arrayMemberWithOverrides(arr, ast.NewIdentifier("iterator"), interp.GlobalEnvironment())
	if err != nil {
		t.Fatalf("iterator should remain available without stdlib: %v", err)
	}
	if _, ok := iterVal.(*runtime.NativeBoundMethodValue); !ok {
		t.Fatalf("iterator should bind to native method, got %T", iterVal)
	}
}

func TestStringHelpersRequireStdlib(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()

	_, err := interp.stringMemberWithOverrides(runtime.StringValue{Val: "hello"}, ast.NewIdentifier("len_bytes"), env)
	if err == nil {
		t.Fatalf("expected len_bytes to be unavailable without stdlib import")
	}
	if err != nil && !strings.Contains(err.Error(), "able.text.string") {
		t.Fatalf("expected error hinting at stdlib, got %v", err)
	}

	if _, err := interp.stringMemberWithOverrides(runtime.StringValue{Val: "hello"}, ast.NewIdentifier("split"), env); err == nil {
		t.Fatalf("expected split to be unavailable without stdlib import")
	}
}
