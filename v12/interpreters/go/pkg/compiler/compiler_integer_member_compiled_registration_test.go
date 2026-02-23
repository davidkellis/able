package compiler

import (
	"strings"
	"testing"
)

func TestCompilerRegistersBuiltinIntegerMemberMethods(t *testing.T) {
	_, compiledSrc := compileOutputs(t, "demo", map[string]string{
		"main.able": strings.Join([]string{
			"package demo",
			"",
			"fn main() -> void {",
			"  value := 41_i32",
			"  _ = value.clone()",
			"  _ = value.to_string()",
			"}",
			"",
		}, "\n"),
	})

	if !strings.Contains(compiledSrc, "func __able_builtin_integer_clone(") {
		t.Fatalf("expected builtin integer clone compiled helper to be emitted")
	}
	if !strings.Contains(compiledSrc, "func __able_builtin_integer_to_string(") {
		t.Fatalf("expected builtin integer to_string compiled helper to be emitted")
	}
	if !strings.Contains(compiledSrc, "__able_register_builtin_compiled_methods()") {
		t.Fatalf("expected RegisterIn to invoke builtin compiled method registration")
	}
	if strings.Contains(compiledSrc, "if _, ok := base.(runtime.IntegerValue); ok {") {
		t.Fatalf("expected legacy integer member_get_method shim branch to be removed")
	}
	if strings.Contains(compiledSrc, "if intPtr, ok := base.(*runtime.IntegerValue); ok && intPtr != nil {") {
		t.Fatalf("expected legacy integer pointer shim branch to be removed")
	}
}
