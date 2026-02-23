package compiler

import (
	"strings"
	"testing"
)

func TestCompilerRegistersBuiltinIteratorMemberMethods(t *testing.T) {
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

	if !strings.Contains(compiledSrc, "func __able_builtin_iterator_next(") {
		t.Fatalf("expected builtin Iterator.next compiled helper to be emitted")
	}
	if !strings.Contains(compiledSrc, "__able_register_compiled_method(\"Iterator\", \"next\", true, 0, 0, __able_builtin_iterator_next)") {
		t.Fatalf("expected Iterator.next builtin method registration")
	}
	if strings.Contains(compiledSrc, "if iter, ok := base.(*runtime.IteratorValue); ok && iter != nil && name == \"next\" {") {
		t.Fatalf("expected legacy Iterator.next member_get_method shim branch to be removed")
	}
	if strings.Contains(compiledSrc, "nextMethod := runtime.NativeFunctionValue{") {
		t.Fatalf("expected legacy Iterator.next native method shim construction to be removed")
	}
}
