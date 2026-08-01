package compiler

import (
	"strings"
	"testing"
)

func TestCompilerFixedIntegerArithmeticMethodsStayOnNativeCarriers(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn wrap_i32(a: i32, b: i32) -> i32 { a.wrapping_add(b) }",
		"fn saturate_u64(a: u64, b: u64) -> u64 { a.saturating_mul(b) }",
		"fn check_i16(a: i16, b: i16) -> ?i16 { a.checked_sub(b) }",
		"fn wrap_i128(a: i128, b: i128) -> i128 { a.wrapping_mul(b) }",
		"fn saturate_u128(a: u128, b: u128) -> u128 { a.saturating_add(b) }",
		"fn check_i128(a: i128, b: i128) -> ?i128 { a.checked_mul(b) }",
		"",
	}, "\n"))

	expectations := map[string][]string{
		"__able_compiled_fn_wrap_i32":      {"int32(", " + "},
		"__able_compiled_fn_saturate_u64":  {"__able_saturating_mul_unsigned(", "uint64("},
		"__able_compiled_fn_check_i16":     {"__able_try_sub_signed(", "__able_nullable[int16]"},
		"__able_compiled_fn_wrap_i128":     {".WrappingMul("},
		"__able_compiled_fn_saturate_u128": {".SaturatingAdd("},
		"__able_compiled_fn_check_i128":    {".MulChecked(", "__able_nullable[runtime.Int128]"},
	}
	for function, fragments := range expectations {
		body, ok := findCompiledFunction(result, function)
		if !ok {
			t.Fatalf("could not find %s", function)
		}
		for _, fragment := range fragments {
			if !strings.Contains(body, fragment) {
				t.Fatalf("expected %s to contain %q:\n%s", function, fragment, body)
			}
		}
		for _, dynamic := range []string{
			"runtime.Value",
			"__able_call_member(",
			"__able_builtin_integer_arithmetic(",
			"__able_any_to_value(",
			"interpreter.",
		} {
			if strings.Contains(body, dynamic) {
				t.Fatalf("expected %s to avoid dynamic boundary %q:\n%s", function, dynamic, body)
			}
		}
	}
}

func TestCompilerWideNativeNullableRuntimeHelpersRoundTrip(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn maybe_i128(flag: bool) -> ?i128 {",
		"  if flag { 170141183460469231731687303715884105727_i128 } else { nil }",
		"}",
		"fn maybe_u128(flag: bool) -> ?u128 {",
		"  if flag { 340282366920938463463374607431768211455_u128 } else { nil }",
		"}",
		"",
	}, "\n"))

	compiled := string(result.Files["compiled.go"])
	for _, fragment := range []string{
		"converted, ok := runtime.Int128FromValue(value)",
		"converted, ok := runtime.Uint128FromValue(value)",
		"return value.value.IntegerValue()",
	} {
		if !strings.Contains(compiled, fragment) {
			t.Fatalf("expected wide nullable runtime helpers to contain %q", fragment)
		}
	}
	for _, helper := range []string{"i128", "u128"} {
		body, ok := findCompiledFunction(result, "__able_compiled_fn_maybe_"+helper)
		if !ok {
			t.Fatalf("could not find native %s nullable function", helper)
		}
		if strings.Contains(body, "runtime.Value") || strings.Contains(body, "__able_any_to_value(") {
			t.Fatalf("expected native %s nullable function to avoid dynamic conversion:\n%s", helper, body)
		}
	}
}
