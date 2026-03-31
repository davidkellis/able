package compiler

import (
	"strings"
	"testing"
)

func TestCompilerInferredGenericFunctionCallStaysNative(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn id<T>(value: T) -> T { value }",
		"",
		"fn main() -> i32 {",
		"  id(1)",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	if !strings.Contains(body, "__able_compiled_fn_id_spec(int32(1))") {
		t.Fatalf("expected checked generic free-function call to lower through the specialized compiled helper:\n%s", body)
	}
	for _, fragment := range []string{
		"__able_call_named(\"id\"",
		"runtime.Value",
		"bridge.AsInt(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected checked generic free-function call to avoid %q:\n%s", fragment, body)
		}
	}
	compiledSrc := string(result.Files["compiled.go"])
	if !strings.Contains(compiledSrc, "func __able_compiled_fn_id_spec(value int32) (int32, *__ableControl)") {
		t.Fatalf("expected specialized generic free-function signature to stay native:\n%s", compiledSrc)
	}
}

func TestCompilerGenericAliasFunctionCallsStayNative(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"type Pair T = Array T",
		"",
		"fn make_pair<T>(value: T) -> Pair T {",
		"  [value, value]",
		"}",
		"",
		"fn first<T>(values: Pair T) -> T {",
		"  values[0]",
		"}",
		"",
		"fn main() -> i32 {",
		"  first(make_pair(5))",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	for _, fragment := range []string{
		"__able_call_named(\"make_pair\"",
		"__able_call_named(\"first\"",
		"runtime.Value",
		"__able_try_cast(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected generic-alias free-function calls to avoid %q:\n%s", fragment, body)
		}
	}
	if !strings.Contains(body, "__able_compiled_fn_make_pair_spec(") || !strings.Contains(body, "__able_compiled_fn_first_spec(") {
		t.Fatalf("expected generic-alias free-function calls to use specialized compiled helpers:\n%s", body)
	}

	compiledSrc := string(result.Files["compiled.go"])
	for _, fragment := range []string{
		"func __able_compiled_fn_make_pair_spec(value int32) (*__able_array_i32, *__ableControl)",
		"func __able_compiled_fn_first_spec(values *__able_array_i32) (int32, *__ableControl)",
	} {
		if !strings.Contains(compiledSrc, fragment) {
			t.Fatalf("expected generic alias specialization to produce native signatures containing %q:\n%s", fragment, compiledSrc)
		}
	}
}

func TestCompilerInferredGenericNominalReturnStaysNative(t *testing.T) {
	result := compileNoFallbackExecSource(t, "ablec-inferred-generic-nominal-return", strings.Join([]string{
		"package demo",
		"",
		"struct Pair A B { left: A, right: B }",
		"",
		"fn make_pair(left: A, right: B) {",
		"  Pair A B { left: left, right: right }",
		"}",
		"",
		"fn main() -> void {",
		"  pair := make_pair(true, 42)",
		"  print(`${pair.right}`)",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	for _, fragment := range []string{
		"var pair runtime.Value",
		"__able_any_to_value(__able_tmp_",
		"__able_struct_Pair_bool_i32_from(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected inferred generic nominal return to stay native and avoid %q:\n%s", fragment, body)
		}
	}

	compiledSrc := string(result.Files["compiled.go"])
	if !strings.Contains(compiledSrc, "func __able_compiled_fn_make_pair_spec(left bool, right int32) (*Pair_bool_i32, *__ableControl)") {
		t.Fatalf("expected inferred generic nominal specialization to produce a concrete native return signature:\n%s", compiledSrc)
	}

	stdout := strings.TrimSpace(compileAndRunExecSourceWithOptions(t, "ablec-inferred-generic-nominal-return-exec", strings.Join([]string{
		"package demo",
		"",
		"struct Pair A B { left: A, right: B }",
		"",
		"fn make_pair(left: A, right: B) {",
		"  Pair A B { left: left, right: right }",
		"}",
		"",
		"fn main() -> void {",
		"  pair := make_pair(true, 42)",
		"  print(`${pair.right}`)",
		"}",
		"",
	}, "\n"), Options{
		PackageName:        "main",
		RequireNoFallbacks: true,
		EmitMain:           true,
	}))
	if stdout != "42" {
		t.Fatalf("expected compiled inferred generic nominal return path to execute, got %q", stdout)
	}
}
