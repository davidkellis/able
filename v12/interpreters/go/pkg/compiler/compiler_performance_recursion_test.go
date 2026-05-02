package compiler

import (
	"regexp"
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
	if regexp.MustCompile(`int64\(__able_tmp_[0-9]+\) - int64\(__able_tmp_[0-9]+\)`).MatchString(body) {
		t.Fatalf("expected recursive compiled body to avoid widened checked subtraction after terminating guard:\n%s", body)
	}
	if !regexp.MustCompile(`__able_tmp_[0-9]+ := __able_tmp_[0-9]+ - __able_tmp_[0-9]+`).MatchString(body) {
		t.Fatalf("expected recursive compiled body to keep guarded decrement direct:\n%s", body)
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

func TestCompilerRecursiveFunctionUsesBoundedReturnFactForKnownCallRange(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn fib(n: i32) -> i32 {",
		"  if n <= 2 { return 1 }",
		"  fib(n - 1) + fib(n - 2)",
		"}",
		"",
		"fn main() -> void {",
		"  print(fib(45))",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_fib")
	if !ok {
		t.Fatalf("could not find compiled fib function")
	}
	if regexp.MustCompile(`int64\(__able_tmp_[0-9]+\) \+ int64\(__able_tmp_[0-9]+\)`).MatchString(body) {
		t.Fatalf("expected bounded recursive return fact to avoid widened checked addition:\n%s", body)
	}
	if strings.Contains(body, "__able_raise_overflow(") {
		t.Fatalf("expected bounded recursive return fact to remove fib overflow transfer:\n%s", body)
	}
	if !regexp.MustCompile(`__able_tmp_[0-9]+ := __able_tmp_[0-9]+ \+ __able_tmp_[0-9]+`).MatchString(body) {
		t.Fatalf("expected bounded recursive return fact to keep the recursive sum direct:\n%s", body)
	}
}

func TestCompilerRecursiveFunctionKeepsCheckedAddWhenReturnRangeExceedsI32(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn grow(n: i32) -> i32 {",
		"  if n <= 0 { return 1 }",
		"  grow(n - 1) + grow(n - 1)",
		"}",
		"",
		"fn main() -> void {",
		"  print(grow(31))",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_grow")
	if !ok {
		t.Fatalf("could not find compiled grow function")
	}
	if !regexp.MustCompile(`int64\(__able_tmp_[0-9]+\) \+ int64\(__able_tmp_[0-9]+\)`).MatchString(body) {
		t.Fatalf("expected overflowing recurrence to keep widened checked addition:\n%s", body)
	}
	if !strings.Contains(body, "__able_raise_overflow(") {
		t.Fatalf("expected overflowing recurrence to retain overflow transfer:\n%s", body)
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
