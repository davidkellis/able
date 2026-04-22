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

	compiledSrc := string(result.Files["compiled.go"])
	for _, fragment := range []string{
		"func __able_push_call_frame(",
		"func __able_pop_call_frame(",
	} {
		if strings.Contains(compiledSrc, fragment) {
			t.Fatalf("expected compiled output to omit obsolete call-frame wrapper %q", fragment)
		}
	}

	body, ok := findCompiledFunction(result, "__able_compiled_fn_fib")
	if !ok {
		t.Fatalf("could not find compiled fib function")
	}
	for _, fragment := range []string{
		"bridge.SwapEnvIfNeeded(__able_runtime,",
		"__able_runtime.SwapEnv(",
		"defer __able_runtime.SwapEnv(",
		"bridge.PushCallFrame(__able_runtime,",
		"bridge.PopCallFrame(__able_runtime)",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected raw recursive compiled body to avoid env swap fragment %q:\n%s", fragment, body)
		}
	}
	if !strings.Contains(body, "__able_compiled_fn_fib(") {
		t.Fatalf("expected recursive compiled body to recurse directly through the raw body:\n%s", body)
	}
	if !strings.Contains(body, "__able_append_control_call_frame(") {
		t.Fatalf("expected recursive compiled body to append caller frames only on the slow control path:\n%s", body)
	}

	entryBody, ok := findCompiledFunction(result, "__able_compiled_entry_fn_fib")
	if !ok {
		t.Fatalf("could not find compiled fib entry wrapper")
	}
	if !strings.Contains(entryBody, "bridge.SwapEnvIfNeeded(__able_runtime,") {
		t.Fatalf("expected recursive compiled entry wrapper to use conditional env swapping:\n%s", entryBody)
	}
	if !strings.Contains(entryBody, "return __able_compiled_fn_fib(") {
		t.Fatalf("expected recursive compiled entry wrapper to forward to raw body:\n%s", entryBody)
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
