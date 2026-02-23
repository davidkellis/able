package compiler

import (
	"strings"
	"testing"
)

func TestCompilerNormalizesStringFromBuiltinUnwrap(t *testing.T) {
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

	if !strings.Contains(compiledSrc, "func __able_runtime_struct_instance_value(value runtime.Value) (*runtime.StructInstanceValue, bool, bool) {") {
		t.Fatalf("expected shared runtime struct instance unwrapping helper to be emitted")
	}

	start := strings.Index(compiledSrc, "func __able_string_from_builtin_impl(args []runtime.Value) (runtime.Value, error) {")
	if start < 0 {
		t.Fatalf("expected __able_string_from_builtin_impl helper to be emitted")
	}
	segment := compiledSrc[start:]
	end := strings.Index(segment, "func __able_string_to_builtin_impl(args []runtime.Value) (runtime.Value, error) {")
	if end < 0 {
		t.Fatalf("expected __able_string_from_builtin_impl segment terminator")
	}
	segment = segment[:end]

	if strings.Contains(segment, "switch v := val.(type)") {
		t.Fatalf("expected legacy string pointer/value switch shim to be removed from __able_string_from_builtin_impl")
	}
	if strings.Contains(segment, "if str, ok, nilPtr := __able_runtime_string_value(val); ok {") {
		t.Fatalf("expected legacy ok-only runtime string helper guard to be removed")
	}
	if !strings.Contains(segment, "if str, ok, nilPtr := __able_runtime_string_value(val); ok || nilPtr {") {
		t.Fatalf("expected __able_string_from_builtin_impl to use normalized runtime string unwrapping helper guard")
	}
	if strings.Contains(segment, "if inst, ok := val.(*runtime.StructInstanceValue); ok {") {
		t.Fatalf("expected legacy direct struct instance assertion to be removed from __able_string_from_builtin_impl")
	}
	if !strings.Contains(segment, "if inst, ok, nilPtr := __able_runtime_struct_instance_value(val); ok || nilPtr {") {
		t.Fatalf("expected __able_string_from_builtin_impl to use shared runtime struct instance unwrapping helper")
	}
	if !strings.Contains(segment, "if !ok || nilPtr {") {
		t.Fatalf("expected __able_string_from_builtin_impl to preserve explicit typed-nil rejection")
	}
	if !strings.Contains(segment, "return __able_string_bytes_from_struct(inst)") {
		t.Fatalf("expected __able_string_from_builtin_impl to preserve struct-instance string path")
	}
}
