package compiler

import (
	"strings"
	"testing"
)

func TestCompilerNormalizesCallValueUnwrapBranches(t *testing.T) {
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

	if !strings.Contains(compiledSrc, "func __able_callable_interface_value(val runtime.Value) (runtime.InterfaceValue, bool, bool) {") {
		t.Fatalf("expected shared interface unwrap helper for __able_call_value")
	}
	if !strings.Contains(compiledSrc, "func __able_callable_partial_value(val runtime.Value) (runtime.PartialFunctionValue, bool, bool) {") {
		t.Fatalf("expected shared partial unwrap helper for __able_call_value")
	}
	if !strings.Contains(compiledSrc, "func __able_merge_bound_args(bound []runtime.Value, args []runtime.Value) []runtime.Value {") {
		t.Fatalf("expected shared bound-arg merge helper for __able_call_value")
	}

	start := strings.Index(compiledSrc, "func __able_call_value(")
	if start < 0 {
		t.Fatalf("expected __able_call_value helper to be emitted")
	}
	segment := compiledSrc[start:]
	end := strings.Index(segment, "type __able_compiled_call_entry struct {")
	if end < 0 {
		t.Fatalf("expected __able_call_value segment terminator")
	}
	segment = segment[:end]

	if strings.Contains(segment, "switch typed := fn.(type) {") {
		t.Fatalf("expected legacy interface/partial unwrapping switch to be removed from __able_call_value")
	}
	if !strings.Contains(segment, "__able_callable_interface_value(fn)") {
		t.Fatalf("expected __able_call_value to use shared interface unwrap helper")
	}
	if strings.Contains(segment, "if iface, ok, nilPtr := __able_callable_interface_value(fn); nilPtr {") {
		t.Fatalf("expected legacy nilPtr-first interface branch to be removed from __able_call_value")
	}
	if !strings.Contains(segment, "if iface, ok, nilPtr := __able_callable_interface_value(fn); ok || nilPtr {") {
		t.Fatalf("expected __able_call_value to use normalized interface helper guard")
	}
	if !strings.Contains(segment, "__able_callable_partial_value(fn)") {
		t.Fatalf("expected __able_call_value to use shared partial unwrap helper")
	}
	if strings.Contains(segment, "if partial, ok, nilPtr := __able_callable_partial_value(fn); nilPtr {") {
		t.Fatalf("expected legacy nilPtr-first partial branch to be removed from __able_call_value")
	}
	if !strings.Contains(segment, "if partial, ok, nilPtr := __able_callable_partial_value(fn); ok || nilPtr {") {
		t.Fatalf("expected __able_call_value to use normalized partial helper guard")
	}
	if !strings.Contains(segment, "if !ok || nilPtr {") {
		t.Fatalf("expected __able_call_value helper branches to enforce normalized typed-nil handling")
	}
	if !strings.Contains(segment, "args = __able_merge_bound_args(partial.BoundArgs, args)") {
		t.Fatalf("expected __able_call_value to use shared bound-arg merge helper")
	}
}
