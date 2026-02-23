package compiler

import (
	"strings"
	"testing"
)

func TestCompilerNormalizesInt64FromValueIntegerUnwrap(t *testing.T) {
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

	start := strings.Index(compiledSrc, "func __able_int64_from_value(val runtime.Value, label string) (int64, error) {")
	if start < 0 {
		t.Fatalf("expected __able_int64_from_value helper to be emitted")
	}
	segment := compiledSrc[start:]
	end := strings.Index(segment, "func __able_new_array_from_values(values []runtime.Value) (*runtime.ArrayValue, error) {")
	if end < 0 {
		t.Fatalf("expected __able_int64_from_value segment terminator")
	}
	segment = segment[:end]

	if strings.Contains(segment, "switch v := current.(type)") {
		t.Fatalf("expected legacy integer pointer/value switch shim to be removed from __able_int64_from_value")
	}
	if strings.Contains(segment, "if v, ok, nilPtr := __able_runtime_integer_value(current); ok {") {
		t.Fatalf("expected legacy ok-only runtime integer helper guard to be removed")
	}
	if !strings.Contains(segment, "if v, ok, nilPtr := __able_runtime_integer_value(current); ok || nilPtr {") {
		t.Fatalf("expected __able_int64_from_value to use normalized integer unwrapping helper guard")
	}
	if !strings.Contains(segment, "if !ok || nilPtr || v.Val == nil {") {
		t.Fatalf("expected __able_int64_from_value to preserve explicit typed-nil/value-nil rejection")
	}
}
