package compiler

import (
	"strings"
	"testing"
)

func TestCompilerRegistersBuiltinFutureMemberMethods(t *testing.T) {
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

	if !strings.Contains(compiledSrc, "func __able_builtin_future_status(") {
		t.Fatalf("expected builtin Future.status compiled helper to be emitted")
	}
	if !strings.Contains(compiledSrc, "func __able_builtin_future_register(") {
		t.Fatalf("expected builtin Future.register compiled helper to be emitted")
	}
	if !strings.Contains(compiledSrc, "__able_register_compiled_method(\"Future\", \"status\", true, 0, 0, __able_builtin_future_status)") {
		t.Fatalf("expected Future.status builtin method registration")
	}
	if !strings.Contains(compiledSrc, "__able_register_compiled_method(\"Future\", \"register\", true, 1, 1, __able_builtin_future_register)") {
		t.Fatalf("expected Future.register builtin method registration")
	}
	if strings.Contains(compiledSrc, "if val, handled := __able_future_member_value(future, name); handled {") {
		t.Fatalf("expected legacy future member_get/member_get_method shim call sites to be removed")
	}
}
