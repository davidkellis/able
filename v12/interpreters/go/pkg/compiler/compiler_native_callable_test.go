package compiler

import (
	"strings"
	"testing"
)

func TestCompilerPlaceholderLambdaStaysNative(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn add(a: i32, b: i32) -> i32 { a + b }",
		"",
		"fn main() -> i32 {",
		"  add_10 := add(@, 10)",
		"  add_10(5)",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	if !strings.Contains(body, "var add_10 __able_fn_runtime_Value_to_int32 = __able_fn_runtime_Value_to_int32(") {
		t.Fatalf("expected placeholder lambda binding to stay on a native callable carrier:\n%s", body)
	}
	for _, fragment := range []string{"runtime.NativeFunctionValue", "__able_call_value(", "__able_call_value_fast("} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected placeholder lambda to avoid %q:\n%s", fragment, body)
		}
	}
}

func TestCompilerBoundMethodValueStaysNative(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"struct Counter { value: i32 }",
		"",
		"methods Counter {",
		"  fn #inc() -> i32 { #value + 1 }",
		"}",
		"",
		"fn main() -> i32 {",
		"  counter := Counter { value: 9 }",
		"  inc_fn := counter.inc",
		"  inc_fn()",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	if !strings.Contains(body, "var inc_fn __able_fn_void_to_int32 = __able_fn_void_to_int32(") {
		t.Fatalf("expected bound method capture to stay on a native callable carrier:\n%s", body)
	}
	for _, fragment := range []string{"runtime.NativeFunctionValue", "__able_member_get_method(", "__able_call_value(", "__able_call_value_fast("} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected bound method capture to avoid %q:\n%s", fragment, body)
		}
	}
}

func TestCompilerFunctionTypedParamStaysNative(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn twice(f: i32 -> i32, value: i32) -> i32 {",
		"  f(f(value))",
		"}",
		"",
		"fn main() -> i32 {",
		"  inc := fn(value: i32) -> i32 { value + 1 }",
		"  twice(inc, 40)",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_twice")
	if !ok {
		t.Fatalf("could not find compiled twice function")
	}
	if !strings.Contains(body, "var __able_tmp_") || !strings.Contains(body, "__able_fn_int32_to_int32") {
		t.Fatalf("expected function-typed param calls to use the native callable carrier:\n%s", body)
	}
	for _, fragment := range []string{"__able_call_value(", "__able_call_value_fast(", "runtime.NativeFunctionValue"} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected function-typed param call to avoid %q:\n%s", fragment, body)
		}
	}
}

func TestCompilerNativeCallableExecutes(t *testing.T) {
	source := strings.Join([]string{
		"extern go fn __able_os_exit(code: i32) -> void {}",
		"",
		"fn add(a: i32, b: i32) -> i32 { a + b }",
		"",
		"struct Counter { value: i32 }",
		"",
		"methods Counter {",
		"  fn #inc() -> i32 { #value + 1 }",
		"}",
		"",
		"fn twice(f: i32 -> i32, value: i32) -> i32 {",
		"  f(f(value))",
		"}",
		"",
		"fn main() {",
		"  add_10 := add(@, 10)",
		"  counter := Counter { value: 9 }",
		"  inc_fn := counter.inc",
		"  if add_10(5) == 15 && inc_fn() == 10 && twice(fn(value: i32) -> i32 { value + 1 }, 40) == 42 {",
		"    __able_os_exit(0)",
		"  }",
		"  __able_os_exit(1)",
		"}",
		"",
	}, "\n")

	compileAndRunSource(t, "ablec-native-callable-", source)
}

func TestCompilerPipePlaceholderLambdaStaysNative(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn main() -> i32 {",
		"  value := 10 |> (@ + 7)",
		"  value",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	if !strings.Contains(body, "__able_fn_runtime_Value_to_runtime_Value(") {
		t.Fatalf("expected pipe placeholder lambda to stay on a native callable carrier:\n%s", body)
	}
	if strings.Contains(body, "__able_call_value(") || strings.Contains(body, "__able_call_value_fast(") {
		t.Fatalf("expected pipe placeholder lambda call to avoid runtime call helpers:\n%s", body)
	}
}

func TestCompilerPipePlaceholderLambdaExecutes(t *testing.T) {
	source := strings.Join([]string{
		"extern go fn __able_os_exit(code: i32) -> void {}",
		"",
		"fn main() {",
		"  if (10 |> (@ + 7)) == 17 {",
		"    __able_os_exit(0)",
		"  }",
		"  __able_os_exit(1)",
		"}",
		"",
	}, "\n")

	compileAndRunSource(t, "ablec-pipe-placeholder-native-", source)
}
