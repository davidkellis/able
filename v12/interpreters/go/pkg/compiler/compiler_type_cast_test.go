package compiler

import (
	"strings"
	"testing"
)

func TestCompilerExplicitUnprovenInterfaceCastUsesRuntimeCheck(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"interface Display for Self {",
		"  fn render(self: Self) -> String",
		"}",
		"",
		"struct Opaque {}",
		"",
		"fn main() -> void {",
		"  value := Opaque {} as Display",
		"  value",
		"}",
		"",
	}, "\n"))

	for _, warning := range result.Warnings {
		if strings.Contains(warning, "cannot cast Opaque to Display") {
			t.Fatalf("compiler retained a checker rejection for an explicit interface cast: %q", warning)
		}
	}
	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatal("could not find compiled main function")
	}
	if !strings.Contains(body, "__able_cast(") {
		t.Fatalf("expected explicit unproven interface cast to retain the runtime check:\n%s", body)
	}
}
