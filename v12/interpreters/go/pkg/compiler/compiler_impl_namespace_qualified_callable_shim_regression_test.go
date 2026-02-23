package compiler

import (
	"strings"
	"testing"
)

func TestCompilerRemovesImplNamespacePointerQualifiedCallableShim(t *testing.T) {
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
		t.Fatalf("expected value-form ImplementationNamespace branch to remain in qualified callable resolver")
	}
	if got := strings.Count(compiledSrc, "if method, ok := typed.Methods[tail]; ok && method != nil {"); got != 1 {
		t.Fatalf("expected exactly one ImplementationNamespace resolver method branch, got %d", got)
	}
}
