package compiler

import (
	"strings"
	"testing"
)

func TestCompilerRemovesNilPointerQualifiedCallableShim(t *testing.T) {
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

	start := strings.Index(compiledSrc, "func __able_resolve_qualified_callable(name string, env *runtime.Environment) (runtime.Value, bool, error) {")
	if start < 0 {
		t.Fatalf("expected qualified callable resolver helper")
	}
	segment := compiledSrc[start:]
	end := strings.Index(segment, "func __able_call_named(")
	if end < 0 {
		t.Fatalf("expected qualified callable resolver segment terminator")
	}
	segment = segment[:end]
	if strings.Contains(segment, "switch candidate.(type) {") {
		t.Fatalf("expected qualified callable resolver candidate switch shim to be removed")
	}
	if !strings.Contains(segment, "if !__able_is_nil(candidate) {") {
		t.Fatalf("expected qualified callable resolver to use shared nil-value helper for candidate filtering")
	}
}
