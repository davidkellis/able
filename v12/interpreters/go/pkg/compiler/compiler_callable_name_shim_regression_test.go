package compiler

import (
	"strings"
	"testing"
)

func TestCompilerNormalizesCallableNameUnwrapBranches(t *testing.T) {
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

	if !strings.Contains(compiledSrc, "func __able_callable_native_function_value(val runtime.Value) (runtime.NativeFunctionValue, bool, bool) {") {
		t.Fatalf("expected shared native-function unwrap helper for __able_callable_name")
	}
	if !strings.Contains(compiledSrc, "func __able_callable_native_bound_method_value(val runtime.Value) (runtime.NativeBoundMethodValue, bool, bool) {") {
		t.Fatalf("expected shared native-bound unwrap helper for __able_callable_name")
	}
	if !strings.Contains(compiledSrc, "func __able_callable_bound_method_value(val runtime.Value) (runtime.BoundMethodValue, bool, bool) {") {
		t.Fatalf("expected shared bound-method unwrap helper for __able_callable_name")
	}

	start := strings.Index(compiledSrc, "func __able_callable_name(value runtime.Value) string {")
	if start < 0 {
		t.Fatalf("expected __able_callable_name helper to be emitted")
	}
	segment := compiledSrc[start:]
	end := strings.Index(segment, "func __able_mark_boundary_fallback(name string) {")
	if end < 0 {
		t.Fatalf("expected __able_callable_name segment terminator")
	}
	segment = segment[:end]

	if strings.Contains(segment, "switch typed := value.(type)") {
		t.Fatalf("expected legacy pointer/value switch shim to be removed from __able_callable_name")
	}
	if !strings.Contains(segment, "__able_callable_native_function_value(value)") {
		t.Fatalf("expected __able_callable_name to use shared native-function unwrap helper")
	}
	if strings.Contains(segment, "if native, ok, nilPtr := __able_callable_native_function_value(value); nilPtr {") {
		t.Fatalf("expected legacy nilPtr-first native helper branch to be removed from __able_callable_name")
	}
	if !strings.Contains(segment, "if native, ok, nilPtr := __able_callable_native_function_value(value); ok || nilPtr {") {
		t.Fatalf("expected __able_callable_name native helper branch to use normalized ok||nilPtr guard")
	}
	if !strings.Contains(segment, "if !ok || nilPtr {") {
		t.Fatalf("expected __able_callable_name helper branches to use normalized !ok||nilPtr guard")
	}
	if !strings.Contains(segment, "__able_callable_native_bound_method_value(value)") {
		t.Fatalf("expected __able_callable_name to use shared native-bound unwrap helper")
	}
	if strings.Contains(segment, "if nativeBound, ok, nilPtr := __able_callable_native_bound_method_value(value); nilPtr {") {
		t.Fatalf("expected legacy nilPtr-first native-bound helper branch to be removed from __able_callable_name")
	}
	if !strings.Contains(segment, "if nativeBound, ok, nilPtr := __able_callable_native_bound_method_value(value); ok || nilPtr {") {
		t.Fatalf("expected __able_callable_name native-bound helper branch to use normalized ok||nilPtr guard")
	}
	if !strings.Contains(segment, "__able_callable_bound_method_value(value)") {
		t.Fatalf("expected __able_callable_name to use shared bound-method unwrap helper")
	}
	if strings.Contains(segment, "if bound, ok, nilPtr := __able_callable_bound_method_value(value); nilPtr {") {
		t.Fatalf("expected legacy nilPtr-first bound-method helper branch to be removed from __able_callable_name")
	}
	if !strings.Contains(segment, "if bound, ok, nilPtr := __able_callable_bound_method_value(value); ok || nilPtr {") {
		t.Fatalf("expected __able_callable_name bound-method helper branch to use normalized ok||nilPtr guard")
	}
	if !strings.Contains(segment, "__able_callable_partial_value(value)") {
		t.Fatalf("expected __able_callable_name to use shared partial unwrap helper")
	}
	if strings.Contains(segment, "if partial, ok, nilPtr := __able_callable_partial_value(value); nilPtr {") {
		t.Fatalf("expected legacy nilPtr-first partial helper branch to be removed from __able_callable_name")
	}
	if !strings.Contains(segment, "if partial, ok, nilPtr := __able_callable_partial_value(value); ok || nilPtr {") {
		t.Fatalf("expected __able_callable_name partial helper branch to use normalized ok||nilPtr guard")
	}
	if !strings.Contains(segment, "__able_callable_interface_value(value)") {
		t.Fatalf("expected __able_callable_name to use shared interface unwrap helper")
	}
	if strings.Contains(segment, "if iface, ok, nilPtr := __able_callable_interface_value(value); nilPtr {") {
		t.Fatalf("expected legacy nilPtr-first interface helper branch to be removed from __able_callable_name")
	}
	if !strings.Contains(segment, "if iface, ok, nilPtr := __able_callable_interface_value(value); ok || nilPtr {") {
		t.Fatalf("expected __able_callable_name interface helper branch to use normalized ok||nilPtr guard")
	}
}
