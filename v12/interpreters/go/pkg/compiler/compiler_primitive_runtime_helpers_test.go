package compiler

import (
	"strings"
	"testing"
)

func TestCompilerPrimitiveKernelHelpersStayNative(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn main() -> void {",
		"  code := __able_char_to_codepoint('A')",
		"  ch := __able_char_from_codepoint(code)",
		"  _ = __able_char_simple_fold_next(ch)",
		"  _ = __able_f32_bits(1.5_f32)",
		"  _ = __able_f64_bits(2.5)",
		"  _ = __able_f64_sqrt(6.25)",
		"  _ = __able_u64_mul(7_u64, 9_u64)",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("main body not found")
	}
	for _, fragment := range []string{
		"__able_char_to_codepoint_native(",
		"__able_char_from_codepoint_native(",
		"__able_char_simple_fold_next_native(",
		"__able_f32_bits_native(",
		"__able_f64_bits_native(",
		"__able_f64_sqrt_native(",
		"__able_u64_mul_native(",
	} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("expected primitive helper call to use %q:\n%s", fragment, body)
		}
	}
	for _, fragment := range []string{
		"[]runtime.Value{",
		"__able_char_to_codepoint_impl(",
		"__able_char_from_codepoint_impl(",
		"__able_char_simple_fold_next_impl(",
		"__able_f32_bits_impl(",
		"__able_f64_bits_impl(",
		"__able_f64_sqrt_impl(",
		"__able_u64_mul_impl(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected primitive helper call to avoid %q:\n%s", fragment, body)
		}
	}
	if !strings.Contains(body, "__able_control_from_error_with_node(") {
		t.Fatalf("expected fallible codepoint conversion to preserve source-aware error control:\n%s", body)
	}
}

func TestCompilerPrimitiveKernelHelperLocalBindingStillWins(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn main() -> void {",
		"  __able_char_to_codepoint := { value => 7_i32 }",
		"  _ = __able_char_to_codepoint('A')",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("main body not found")
	}
	if strings.Contains(body, "__able_char_to_codepoint_native(") {
		t.Fatalf("expected local helper-name binding to retain ordinary callable semantics:\n%s", body)
	}
}

func TestCompilerPrimitiveKernelHelpersExecuteNatively(t *testing.T) {
	stdout := compileAndRunExecSourceWithOptions(t, "ablec-primitive-kernel-helpers-", strings.Join([]string{
		"package main",
		"",
		"fn main() -> void {",
		"  code := __able_char_to_codepoint('A')",
		"  ch := __able_char_from_codepoint(code)",
		"  fold := __able_char_simple_fold_next(ch)",
		"  f32 := __able_f32_bits(1.5_f32)",
		"  f64 := __able_f64_bits(2.5)",
		"  root := __able_f64_sqrt(6.25)",
		"  product := __able_u64_mul(7_u64, 9_u64)",
		"  print(`${code}:${ch}:${fold}:${f32}:${f64}:${root}:${product}`)",
		"}",
		"",
	}, "\n"), Options{PackageName: "main", EmitMain: true})

	const want = "65:A:a:1069547520:4612811918334230528:2.5:63\n"
	if stdout != want {
		t.Fatalf("primitive kernel helper output = %q, want %q", stdout, want)
	}
}
