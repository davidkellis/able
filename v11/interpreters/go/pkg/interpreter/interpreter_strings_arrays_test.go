package interpreter

import (
	"testing"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
)

func TestArrayHelpersRequireStdlib(t *testing.T) {
	interp := New()
	interp.ensureArrayBuiltins()
	arr := &runtime.ArrayValue{Elements: []runtime.Value{runtime.NilValue{}}}

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

	if _, err := interp.stringMemberWithOverrides(runtime.StringValue{Val: "hello"}, ast.NewIdentifier("split"), env); err == nil {
		t.Fatalf("expected split to be unavailable without stdlib import")
	}
}
