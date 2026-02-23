package compiler

import (
	"strings"
	"testing"
)

func TestCompilerRegistersBuiltinDynPackageMemberMethods(t *testing.T) {
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

	if !strings.Contains(compiledSrc, "func __able_builtin_dynpackage_def(") {
		t.Fatalf("expected builtin DynPackage.def compiled helper to be emitted")
	}
	if !strings.Contains(compiledSrc, "func __able_builtin_dynpackage_eval(") {
		t.Fatalf("expected builtin DynPackage.eval compiled helper to be emitted")
	}
	if !strings.Contains(compiledSrc, "__able_register_compiled_method(\"DynPackage\", \"def\", true, 1, 1, __able_builtin_dynpackage_def)") {
		t.Fatalf("expected DynPackage.def builtin method registration")
	}
	if !strings.Contains(compiledSrc, "__able_register_compiled_method(\"DynPackage\", \"eval\", true, 1, 1, __able_builtin_dynpackage_eval)") {
		t.Fatalf("expected DynPackage.eval builtin method registration")
	}
	if strings.Contains(compiledSrc, "if name == \"def\" || name == \"eval\" {") {
		t.Fatalf("expected legacy dynpackage def/eval member_get_method shim branch to be removed")
	}
	if strings.Contains(compiledSrc, "method, err := bridge.MemberGet(__able_runtime, typed, runtime.StringValue{Val: name})") {
		t.Fatalf("expected legacy dynpackage member_get_method bridge lookup shim to be removed")
	}
}
