package compiler

import (
	"strings"
	"testing"
)

func TestCompilerNormalizesIsNilUnwrap(t *testing.T) {
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

	if !strings.Contains(compiledSrc, "func __able_runtime_nil_value(value runtime.Value) (runtime.NilValue, bool, bool) {") {
		t.Fatalf("expected shared runtime nil unwrapping helper to be emitted")
	}

	start := strings.Index(compiledSrc, "func __able_is_nil(val runtime.Value) bool {")
	if start < 0 {
		t.Fatalf("expected __able_is_nil helper to be emitted")
	}
	segment := compiledSrc[start:]
	end := strings.Index(segment, "func __able_is_error(val runtime.Value) bool {")
	if end < 0 {
		t.Fatalf("expected __able_is_nil segment terminator")
	}
	segment = segment[:end]

	if strings.Contains(segment, "switch val.(type)") {
		t.Fatalf("expected legacy nil pointer/value switch shim to be removed from __able_is_nil")
	}
	if !strings.Contains(segment, "_, ok, nilPtr := __able_runtime_nil_value(val)") {
		t.Fatalf("expected __able_is_nil to use shared runtime nil unwrapping helper")
	}
}
