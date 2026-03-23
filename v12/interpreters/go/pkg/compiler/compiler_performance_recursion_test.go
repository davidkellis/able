package compiler

import (
	"strings"
	"testing"
)

func TestCompilerRecursiveFunctionUsesConditionalEnvSwap(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn fib(n: i32) -> i32 {",
		"  if n <= 2 { return 1 }",
		"  fib(n - 1) + fib(n - 2)",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_fib")
	if !ok {
		t.Fatalf("could not find compiled fib function")
	}
	if !strings.Contains(body, "bridge.SwapEnvIfNeeded(__able_runtime,") {
		t.Fatalf("expected recursive compiled function to use conditional env swapping:\n%s", body)
	}
	for _, fragment := range []string{
		"__able_runtime.SwapEnv(",
		"defer __able_runtime.SwapEnv(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected recursive compiled function to avoid direct env swap fragment %q:\n%s", fragment, body)
		}
	}
}

func TestCompilerRecursiveFunctionExecutes(t *testing.T) {
	source := strings.Join([]string{
		"package demo",
		"",
		"fn fib(n: i32) -> i32 {",
		"  if n <= 2 { return 1 }",
		"  fib(n - 1) + fib(n - 2)",
		"}",
		"",
		"fn main() -> void {",
		"  print(fib(10))",
		"}",
	}, "\n")

	stdout := compileAndRunSourceWithOptions(t, "ablec-fib-recursion-", source, Options{
		PackageName: "main",
		EmitMain:    true,
	})
	if strings.TrimSpace(stdout) != "55" {
		t.Fatalf("expected recursive compiled program to print 55, got %q", stdout)
	}
}
