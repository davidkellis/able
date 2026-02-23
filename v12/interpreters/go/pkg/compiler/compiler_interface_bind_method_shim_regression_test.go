package compiler

import (
	"strings"
	"testing"
)

func TestCompilerNormalizesInterfaceBindMethodDispatch(t *testing.T) {
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

	receiverStart := strings.Index(compiledSrc, "func __able_interface_method_receiver(")
	if receiverStart < 0 {
		t.Fatalf("expected __able_interface_method_receiver helper to be emitted")
	}
	receiverSegment := compiledSrc[receiverStart:]
	receiverEnd := strings.Index(receiverSegment, "func __able_interface_bind_method(")
	if receiverEnd < 0 {
		t.Fatalf("expected __able_interface_method_receiver segment terminator")
	}
	receiverSegment = receiverSegment[:receiverEnd]

	if strings.Contains(receiverSegment, "switch fn := method.(type)") {
		t.Fatalf("expected legacy method-kind switch shim to be removed from __able_interface_method_receiver")
	}
	if !strings.Contains(receiverSegment, "if bound, ok, _ := __able_callable_bound_method_value(method); ok {") {
		t.Fatalf("expected __able_interface_method_receiver to use shared bound-method unwrapping helper")
	}

	bindStart := strings.Index(compiledSrc, "func __able_interface_bind_method(")
	if bindStart < 0 {
		t.Fatalf("expected __able_interface_bind_method helper to be emitted")
	}
	bindSegment := compiledSrc[bindStart:]
	bindEnd := strings.Index(bindSegment, "func __able_generic_name_set(")
	if bindEnd < 0 {
		t.Fatalf("expected __able_interface_bind_method segment terminator")
	}
	bindSegment = bindSegment[:bindEnd]

	if strings.Contains(bindSegment, "switch fn := method.(type)") {
		t.Fatalf("expected legacy method-kind switch shim to be removed from __able_interface_bind_method")
	}
	if strings.Contains(bindSegment, "if native, ok, nilPtr := __able_callable_native_function_value(method); ok {") {
		t.Fatalf("expected legacy ok-only native-function helper guard to be removed")
	}
	if !strings.Contains(bindSegment, "if native, ok, nilPtr := __able_callable_native_function_value(method); ok || nilPtr {") {
		t.Fatalf("expected __able_interface_bind_method to use normalized native-function helper guard")
	}
	if strings.Contains(bindSegment, "if nativeBound, ok, nilPtr := __able_callable_native_bound_method_value(method); ok {") {
		t.Fatalf("expected legacy ok-only native-bound helper guard to be removed")
	}
	if !strings.Contains(bindSegment, "if nativeBound, ok, nilPtr := __able_callable_native_bound_method_value(method); ok || nilPtr {") {
		t.Fatalf("expected __able_interface_bind_method to use normalized native-bound helper guard")
	}
	if strings.Contains(bindSegment, "if bound, ok, nilPtr := __able_callable_bound_method_value(method); ok {") {
		t.Fatalf("expected legacy ok-only bound-method helper guard to be removed")
	}
	if !strings.Contains(bindSegment, "if bound, ok, nilPtr := __able_callable_bound_method_value(method); ok || nilPtr {") {
		t.Fatalf("expected __able_interface_bind_method to use normalized bound-method helper guard")
	}
	if !strings.Contains(bindSegment, "if !ok || nilPtr {") {
		t.Fatalf("expected __able_interface_bind_method to preserve explicit typed-nil rejection")
	}
}
