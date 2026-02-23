package compiler

import (
	"strings"
	"testing"
)

func TestCompilerTreatsSelfTypedFirstMethodParamAsInstanceReceiver(t *testing.T) {
	_, compiledSrc := compileOutputs(t, "demo", map[string]string{
		"main.able": strings.Join([]string{
			"package demo",
			"",
			"struct Counter {",
			"  n: i32",
			"}",
			"",
			"methods Counter {",
			"  fn bump(this: Self) -> i32 {",
			"    this.n + 1",
			"  }",
			"}",
			"",
			"fn main() -> void {",
			"  _ = Counter { n: 1 }.bump()",
			"}",
			"",
		}, "\n"),
	})

	if !strings.Contains(compiledSrc, "__able_register_compiled_method(\"Counter\", \"bump\", true") {
		t.Fatalf("expected Counter.bump to be registered as a compiled instance method when first param type is Self")
	}
	if strings.Contains(compiledSrc, "__able_register_compiled_method(\"Counter\", \"bump\", false") {
		t.Fatalf("expected Counter.bump not to be registered as a static method")
	}
}
