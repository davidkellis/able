package compiler

import (
	"strings"
	"testing"
)

func TestCompilerRemovesImplNamespacePointerMemberGetMethodShim(t *testing.T) {
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

	if !strings.Contains(compiledSrc, "case runtime.ImplementationNamespaceValue:") {
		t.Fatalf("expected value-form ImplementationNamespace member_get_method branch to remain")
	}
	if got := strings.Count(compiledSrc, "if method, ok := typed.Methods[name]; ok && method != nil {"); got != 1 {
		t.Fatalf("expected exactly one member_get_method implementation-namespace method branch, got %d", got)
	}
}
