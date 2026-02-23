package compiler

import (
	"strings"
	"testing"
)

func TestCompilerNormalizesMemberNameStringUnwrap(t *testing.T) {
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

	if !strings.Contains(compiledSrc, "func __able_runtime_string_value(value runtime.Value) (runtime.StringValue, bool, bool) {") {
		t.Fatalf("expected shared runtime string unwrapping helper to be emitted")
	}

	start := strings.Index(compiledSrc, "func __able_member_name(member runtime.Value) (string, bool) {")
	if start < 0 {
		t.Fatalf("expected __able_member_name helper to be emitted")
	}
	segment := compiledSrc[start:]
	end := strings.Index(segment, "}\n\n")
	if end < 0 {
		t.Fatalf("expected __able_member_name segment terminator")
	}
	segment = segment[:end]

	if strings.Contains(segment, "switch v := val.(type)") {
		t.Fatalf("expected legacy pointer/value switch shim to be removed from __able_member_name")
	}
	if strings.Contains(segment, "if str, ok, nilPtr := __able_runtime_string_value(val); ok {") {
		t.Fatalf("expected legacy ok-only string unwrap guard to be removed from __able_member_name")
	}
	if !strings.Contains(segment, "if str, ok, nilPtr := __able_runtime_string_value(val); ok || nilPtr {") {
		t.Fatalf("expected __able_member_name to use normalized runtime string unwrapping guard")
	}
	if !strings.Contains(segment, "if !ok || nilPtr {") {
		t.Fatalf("expected __able_member_name to treat typed-nil string unwraps as invalid")
	}
}
