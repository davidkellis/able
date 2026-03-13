package compiler

import (
	"strings"
	"testing"
)

func TestCompilerPrefersInterfaceDispatchBeforeUFCSInMemberGetMethod(t *testing.T) {
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

	start := strings.Index(compiledSrc, "func __able_try_member_get_method(obj runtime.Value, member runtime.Value) (runtime.Value, error) {")
	if start < 0 {
		t.Fatalf("expected __able_member_get_method helper to be emitted")
	}
	segment := compiledSrc[start:]
	end := strings.Index(segment, "func __able_member_get_method(obj runtime.Value, member runtime.Value) (runtime.Value, *__ableControl) {")
	if end < 0 {
		t.Fatalf("expected __able_member_get_method segment terminator")
	}
	segment = segment[:end]

	methodDispatchIdx := strings.Index(segment, "if method, ok := __able_interface_dispatch_member(base, name); ok {")
	if methodDispatchIdx < 0 {
		t.Fatalf("expected interface member dispatch path in __able_member_get_method")
	}
	ufcsFallbackIdx := strings.Index(segment, "if receiverTypeName != \"\" && __able_compiled_ufcs_target_exists(name, receiverTypeName) {")
	if ufcsFallbackIdx < 0 {
		t.Fatalf("expected UFCS fallback path in __able_member_get_method")
	}
	if methodDispatchIdx > ufcsFallbackIdx {
		t.Fatalf("expected interface member dispatch to run before UFCS fallback")
	}
}
