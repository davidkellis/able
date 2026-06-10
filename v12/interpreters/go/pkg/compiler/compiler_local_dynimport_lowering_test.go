package compiler

import (
	"strings"
	"testing"
)

func TestCompilerLocalDynImportUsesExplicitScopedBridge(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn main() -> void {",
		"  dynimport demo.dynamic.{answer}",
		"  print(answer())",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatal("could not find compiled main function")
	}
	for _, fragment := range []string{
		"runtime.NewEnvironment(",
		"bridge.SwapEnvIfNeeded(",
		"bridge.EvaluateStatement(__able_runtime, ast.NewDynImportStatement(",
	} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("local dynimport lowering should contain %q:\n%s", fragment, body)
		}
	}
}
