package compiler

import (
	"strings"
	"testing"
)

func TestCompilerNoFallbacksForLocalMethodsAndImplDefinitions(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn main() -> i32 {",
		"  interface LocalShow {",
		"    fn value(self: Self) -> i32",
		"  }",
		"  struct LocalBox {",
		"    n: i32",
		"  }",
		"  methods LocalBox {",
		"    fn bump(this: Self) -> i32 {",
		"      this.n + 1",
		"    }",
		"  }",
		"  impl LocalShow for LocalBox {",
		"    fn value(this: Self) -> i32 {",
		"      this.bump()",
		"    }",
		"  }",
		"  1",
		"}",
		"",
	}, "\n"))

	compiledSrc := string(result.Files["compiled.go"])
	if !strings.Contains(compiledSrc, "bridge.RegisterMethodsDefinition(__able_runtime, func() *ast.MethodsDefinition") {
		t.Fatalf("expected local methods definition to lower through direct bridge registration without fallback")
	}
	if !strings.Contains(compiledSrc, "bridge.RegisterImplementationDefinition(__able_runtime, func() *ast.ImplementationDefinition") {
		t.Fatalf("expected local implementation definition to lower through direct bridge registration without fallback")
	}
	if strings.Contains(compiledSrc, "bridge.EvaluateStatement(__able_runtime") {
		t.Fatalf("expected local methods/impl path to avoid bridge.EvaluateStatement fallback registration")
	}
	if strings.Contains(compiledSrc, "CallOriginal(\"demo.main\"") {
		t.Fatalf("expected local methods/impl definition path to stay compiled without call_original fallback")
	}
}
