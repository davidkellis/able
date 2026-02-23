package compiler

import (
	"strings"
	"testing"
)

func TestCompilerNormalizesStructInstanceErrorUnwrap(t *testing.T) {
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

	if !strings.Contains(compiledSrc, "func __able_runtime_error_value(value runtime.Value) (runtime.ErrorValue, bool, bool) {") {
		t.Fatalf("expected shared runtime error unwrapping helper to be emitted")
	}
	if !strings.Contains(compiledSrc, "func __able_runtime_struct_instance_value(value runtime.Value) (*runtime.StructInstanceValue, bool, bool) {") {
		t.Fatalf("expected shared runtime struct instance unwrapping helper to be emitted")
	}

	start := strings.Index(compiledSrc, "func __able_struct_instance(value runtime.Value) *runtime.StructInstanceValue {")
	if start < 0 {
		t.Fatalf("expected __able_struct_instance helper to be emitted")
	}
	segment := compiledSrc[start:]
	end := strings.Index(segment, "func __able_error_value(value runtime.Value) runtime.Value {")
	if end < 0 {
		t.Fatalf("expected __able_struct_instance segment terminator")
	}
	segment = segment[:end]

	if strings.Contains(segment, "switch v := value.(type)") {
		t.Fatalf("expected legacy error pointer/value switch shim to be removed from __able_struct_instance")
	}
	if strings.Contains(segment, "if inst, ok := value.(*runtime.StructInstanceValue); ok {") {
		t.Fatalf("expected legacy direct struct instance assertion to be removed from __able_struct_instance")
	}
	if !strings.Contains(segment, "if inst, ok, nilPtr := __able_runtime_struct_instance_value(value); ok || nilPtr {") {
		t.Fatalf("expected __able_struct_instance to use shared runtime struct instance unwrapping helper")
	}
	if strings.Contains(segment, "if errVal, ok, nilPtr := __able_runtime_error_value(value); ok {") {
		t.Fatalf("expected legacy ok-only runtime error helper guard to be removed from __able_struct_instance")
	}
	if !strings.Contains(segment, "if errVal, ok, nilPtr := __able_runtime_error_value(value); ok || nilPtr {") {
		t.Fatalf("expected __able_struct_instance to use normalized runtime error unwrapping helper guard")
	}
	if !strings.Contains(segment, "if !ok || nilPtr {") {
		t.Fatalf("expected __able_struct_instance to preserve explicit typed-nil rejection")
	}
}
