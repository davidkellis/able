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

	for _, method := range []struct {
		name   string
		helper string
	}{
		{name: "next", helper: "__able_builtin_iterator_next"},
		{name: "close", helper: "__able_builtin_iterator_close"},
	} {
		if !strings.Contains(compiledSrc, "func "+method.helper+"(") {
			t.Fatalf("expected builtin Iterator.%s compiled helper to be emitted", method.name)
		}
		registration := "__able_register_compiled_method(\"Iterator\", \"" + method.name + "\", true, 0, 0, " + method.helper + ")"
		if !strings.Contains(compiledSrc, registration) {
			t.Fatalf("expected Iterator.%s builtin method registration", method.name)
		}
	}
	if strings.Contains(compiledSrc, "if iter, ok := base.(*runtime.IteratorValue); ok && iter != nil && name == \"next\" {") {
		t.Fatalf("expected legacy Iterator.next member_get_method shim branch to be removed")
	}
	if strings.Contains(compiledSrc, "nextMethod := runtime.NativeFunctionValue{") {
		t.Fatalf("expected legacy Iterator.next native method shim construction to be removed")
	}
}
