package compiler

import (
	"strings"
	"testing"
)

func TestCompilerNormalizesCharToCodepointUnwrap(t *testing.T) {
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

	if !strings.Contains(compiledSrc, "func __able_runtime_char_value(value runtime.Value) (runtime.CharValue, bool, bool) {") {
		t.Fatalf("expected shared runtime char unwrapping helper to be emitted")
	}

	start := strings.Index(compiledSrc, "func __able_char_to_codepoint_impl(args []runtime.Value) (runtime.Value, error) {")
	if start < 0 {
		t.Fatalf("expected __able_char_to_codepoint_impl helper to be emitted")
	}
	segment := compiledSrc[start:]
	end := strings.Index(segment, "func __able_extern_string_from_builtin(args []runtime.Value, node ast.Node) runtime.Value {")
	if end < 0 {
		t.Fatalf("expected __able_char_to_codepoint_impl segment terminator")
	}
	segment = segment[:end]

	if strings.Contains(segment, "switch v := val.(type)") {
		t.Fatalf("expected legacy char pointer/value switch shim to be removed from __able_char_to_codepoint_impl")
	}
	if strings.Contains(segment, "if ch, ok, nilPtr := __able_runtime_char_value(val); ok {") {
		t.Fatalf("expected legacy ok-only runtime char helper guard to be removed")
	}
	if !strings.Contains(segment, "if ch, ok, nilPtr := __able_runtime_char_value(val); ok || nilPtr {") {
		t.Fatalf("expected __able_char_to_codepoint_impl to use normalized runtime char unwrapping helper guard")
	}
	if !strings.Contains(segment, "if !ok || nilPtr {") {
		t.Fatalf("expected __able_char_to_codepoint_impl to preserve explicit typed-nil rejection")
	}
}
