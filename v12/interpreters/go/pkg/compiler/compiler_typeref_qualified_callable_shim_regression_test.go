package compiler

import (
	"strings"
	"testing"
)

func TestCompilerRemovesTypeRefPointerQualifiedCallableShim(t *testing.T) {
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

	if strings.Contains(segment, "case runtime.TypeRefValue:") {
		t.Fatalf("expected dedicated TypeRef qualified-callable lookup shim branch to be removed")
	}
	if strings.Contains(segment, "if method, ok := lookupStatic(typed.TypeName); ok {") {
		t.Fatalf("expected direct TypeRef lookupStatic shim branch to be removed")
	}
	if !strings.Contains(segment, "candidate, err := __able_try_member_get_method(receiver, runtime.StringValue{Val: tail})") {
		t.Fatalf("expected qualified callable resolver to use shared member_get_method path")
	}
}
