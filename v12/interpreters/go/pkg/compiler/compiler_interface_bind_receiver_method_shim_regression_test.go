package compiler

import (
	"strings"
	"testing"
)

func TestCompilerNormalizesInterfaceBindReceiverMethodDispatch(t *testing.T) {
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

	start := strings.Index(compiledSrc, "func __able_interface_bind_receiver_method(")
	if start < 0 {
		t.Fatalf("expected __able_interface_bind_receiver_method helper to be emitted")
	}
	segment := compiledSrc[start:]
	end := strings.Index(segment, "func __able_interface_dispatch_static_receiver(")
	if end < 0 {
		t.Fatalf("expected __able_interface_bind_receiver_method segment terminator")
	}
	segment = segment[:end]

	if strings.Contains(segment, "switch fn := method.(type)") {
		t.Fatalf("expected legacy method pointer/value switch dispatch to be removed from __able_interface_bind_receiver_method")
	}
	if strings.Contains(segment, "if native, ok, nilPtr := __able_callable_native_function_value(method); ok {") {
		t.Fatalf("expected legacy ok-only native-function helper guard to be removed from __able_interface_bind_receiver_method")
	}
	if !strings.Contains(segment, "if native, ok, nilPtr := __able_callable_native_function_value(method); ok || nilPtr {") {
		t.Fatalf("expected __able_interface_bind_receiver_method to use normalized native-function helper guard")
	}
	if strings.Contains(segment, "if nativeBound, ok, nilPtr := __able_callable_native_bound_method_value(method); ok {") {
		t.Fatalf("expected legacy ok-only native-bound helper guard to be removed from __able_interface_bind_receiver_method")
	}
	if !strings.Contains(segment, "if nativeBound, ok, nilPtr := __able_callable_native_bound_method_value(method); ok || nilPtr {") {
		t.Fatalf("expected __able_interface_bind_receiver_method to use normalized native-bound helper guard")
	}
	if strings.Contains(segment, "if bound, ok, nilPtr := __able_callable_bound_method_value(method); ok {") {
		t.Fatalf("expected legacy ok-only bound-method helper guard to be removed from __able_interface_bind_receiver_method")
	}
	if !strings.Contains(segment, "if bound, ok, nilPtr := __able_callable_bound_method_value(method); ok || nilPtr {") {
		t.Fatalf("expected __able_interface_bind_receiver_method to use normalized bound-method helper guard")
	}
	if !strings.Contains(segment, "if !ok || nilPtr {") {
		t.Fatalf("expected __able_interface_bind_receiver_method to preserve explicit typed-nil rejection")
	}
}
