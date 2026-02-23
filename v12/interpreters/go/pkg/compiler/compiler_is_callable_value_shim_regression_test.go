package compiler

import (
	"strings"
	"testing"
)

func TestCompilerNormalizesIsCallableValueDispatch(t *testing.T) {
	_, compiledSrc := compileOutputs(t, "demo", map[string]string{
		"main.able": strings.Join([]string{
			"package demo",
			"",
			"fn main() -> void {",
			"  _ = 1",
			"}",
			"",
		}, "\n"),
	})

	start := strings.Index(compiledSrc, "func __able_is_callable_value(val runtime.Value) bool {")
	if start < 0 {
		t.Fatalf("expected __able_is_callable_value helper to be emitted")
	}
	segment := compiledSrc[start:]
	end := strings.Index(segment, "func __able_member_name(member runtime.Value) (string, bool) {")
	if end < 0 {
		t.Fatalf("expected __able_is_callable_value segment terminator")
	}
	segment = segment[:end]

	if strings.Contains(segment, "switch val.(type)") {
		t.Fatalf("expected legacy callable-kind pointer/value switch shim to be removed from __able_is_callable_value")
	}
	if strings.Contains(segment, "if _, ok := val.(*runtime.FunctionValue); ok {") {
		t.Fatalf("expected direct *runtime.FunctionValue assertion to be removed from __able_is_callable_value")
	}
	if strings.Contains(segment, "if _, ok := val.(*runtime.FunctionOverloadValue); ok {") {
		t.Fatalf("expected direct *runtime.FunctionOverloadValue assertion to be removed from __able_is_callable_value")
	}
	if !strings.Contains(segment, "if _, ok, nilPtr := __able_callable_function_value(val); ok || nilPtr {") {
		t.Fatalf("expected __able_is_callable_value to use shared function unwrapping helper")
	}
	if !strings.Contains(segment, "if _, ok, nilPtr := __able_callable_function_overload_value(val); ok || nilPtr {") {
		t.Fatalf("expected __able_is_callable_value to use shared overload unwrapping helper")
	}
	if !strings.Contains(segment, "if _, ok, nilPtr := __able_callable_native_function_value(val); ok || nilPtr {") {
		t.Fatalf("expected __able_is_callable_value to use shared native-function unwrapping helper")
	}
	if !strings.Contains(segment, "if _, ok, nilPtr := __able_callable_native_bound_method_value(val); ok || nilPtr {") {
		t.Fatalf("expected __able_is_callable_value to use shared native-bound unwrapping helper")
	}
	if !strings.Contains(segment, "if _, ok, nilPtr := __able_callable_bound_method_value(val); ok || nilPtr {") {
		t.Fatalf("expected __able_is_callable_value to use shared bound-method unwrapping helper")
	}
	if !strings.Contains(segment, "if _, ok, nilPtr := __able_callable_partial_value(val); ok || nilPtr {") {
		t.Fatalf("expected __able_is_callable_value to use shared partial unwrapping helper")
	}
	if !strings.Contains(compiledSrc, "func __able_callable_function_value(val runtime.Value) (*runtime.FunctionValue, bool, bool) {") {
		t.Fatalf("expected shared function unwrapping helper to be emitted")
	}
	if !strings.Contains(compiledSrc, "func __able_callable_function_overload_value(val runtime.Value) (*runtime.FunctionOverloadValue, bool, bool) {") {
		t.Fatalf("expected shared function-overload unwrapping helper to be emitted")
	}
}
