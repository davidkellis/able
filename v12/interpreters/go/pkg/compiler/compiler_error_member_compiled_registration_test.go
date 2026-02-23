package compiler

import (
	"strings"
	"testing"
)

func TestCompilerRegistersBuiltinErrorMemberMethods(t *testing.T) {
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

	if !strings.Contains(compiledSrc, "func __able_builtin_error_message(") {
		t.Fatalf("expected builtin Error.message compiled helper to be emitted")
	}
	if !strings.Contains(compiledSrc, "func __able_builtin_error_cause(") {
		t.Fatalf("expected builtin Error.cause compiled helper to be emitted")
	}
	if !strings.Contains(compiledSrc, "__able_register_compiled_method(\"Error\", \"message\", true, 0, 0, __able_builtin_error_message)") {
		t.Fatalf("expected Error.message builtin method registration")
	}
	if !strings.Contains(compiledSrc, "__able_register_compiled_method(\"Error\", \"cause\", true, 0, 0, __able_builtin_error_cause)") {
		t.Fatalf("expected Error.cause builtin method registration")
	}
	if strings.Contains(compiledSrc, "messageMethod := runtime.NativeFunctionValue{") {
		t.Fatalf("expected legacy Error.message member_get_method shim branch to be removed")
	}
	if strings.Contains(compiledSrc, "causeMethod := runtime.NativeFunctionValue{") {
		t.Fatalf("expected legacy Error.cause member_get_method shim branch to be removed")
	}
}
