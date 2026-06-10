package compiler

import (
	"strings"
	"testing"
)

func TestCompilerStaticPrimitiveFunctionsUseNativeGoCarriersEndToEnd(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn keep_bool(value: bool) -> bool { value }",
		"fn keep_i8(value: i8) -> i8 { value }",
		"fn keep_i16(value: i16) -> i16 { value }",
		"fn keep_i32(value: i32) -> i32 { value }",
		"fn keep_i64(value: i64) -> i64 { value }",
		"fn keep_i128(value: i128) -> i128 { value }",
		"fn keep_isize(value: isize) -> isize { value }",
		"fn keep_u8(value: u8) -> u8 { value }",
		"fn keep_u16(value: u16) -> u16 { value }",
		"fn keep_u32(value: u32) -> u32 { value }",
		"fn keep_u64(value: u64) -> u64 { value }",
		"fn keep_u128(value: u128) -> u128 { value }",
		"fn keep_usize(value: usize) -> usize { value }",
		"fn keep_f32(value: f32) -> f32 { value }",
		"fn keep_f64(value: f64) -> f64 { value }",
		"fn keep_char(value: char) -> char { value }",
		"fn keep_string(value: String) -> String { value }",
		"",
		"fn main() -> i32 {",
		"  _ = keep_bool(true)",
		"  _ = keep_i8(1_i8)",
		"  _ = keep_i16(2_i16)",
		"  _ = keep_i32(3_i32)",
		"  _ = keep_i64(4_i64)",
		"  _ = keep_i128(5_i128)",
		"  _ = keep_isize(6 as isize)",
		"  _ = keep_u8(7_u8)",
		"  _ = keep_u16(8_u16)",
		"  _ = keep_u32(9_u32)",
		"  _ = keep_u64(10_u64)",
		"  _ = keep_u128(11_u128)",
		"  _ = keep_usize(12 as usize)",
		"  _ = keep_f32(1.25_f32)",
		"  _ = keep_f64(2.5_f64)",
		"  _ = keep_char('x')",
		"  _ = keep_string(\"ok\")",
		"  0",
		"}",
		"",
	}, "\n"))

	compiledSrc := string(result.Files["compiled.go"])
	for _, signature := range []string{
		"func __able_compiled_fn_keep_bool(value bool) (bool, *__ableControl)",
		"func __able_compiled_fn_keep_i8(value int8) (int8, *__ableControl)",
		"func __able_compiled_fn_keep_i16(value int16) (int16, *__ableControl)",
		"func __able_compiled_fn_keep_i32(value int32) (int32, *__ableControl)",
		"func __able_compiled_fn_keep_i64(value int64) (int64, *__ableControl)",
		"func __able_compiled_fn_keep_i128(value runtime.Int128) (runtime.Int128, *__ableControl)",
		"func __able_compiled_fn_keep_isize(value int) (int, *__ableControl)",
		"func __able_compiled_fn_keep_u8(value uint8) (uint8, *__ableControl)",
		"func __able_compiled_fn_keep_u16(value uint16) (uint16, *__ableControl)",
		"func __able_compiled_fn_keep_u32(value uint32) (uint32, *__ableControl)",
		"func __able_compiled_fn_keep_u64(value uint64) (uint64, *__ableControl)",
		"func __able_compiled_fn_keep_u128(value runtime.Uint128) (runtime.Uint128, *__ableControl)",
		"func __able_compiled_fn_keep_usize(value uint) (uint, *__ableControl)",
		"func __able_compiled_fn_keep_f32(value float32) (float32, *__ableControl)",
		"func __able_compiled_fn_keep_f64(value float64) (float64, *__ableControl)",
		"func __able_compiled_fn_keep_char(value rune) (rune, *__ableControl)",
		"func __able_compiled_fn_keep_string(value string) (string, *__ableControl)",
	} {
		if !strings.Contains(compiledSrc, signature) {
			t.Fatalf("missing native primitive signature %q", signature)
		}
	}

	for _, function := range []string{
		"__able_compiled_fn_main",
		"__able_compiled_fn_keep_bool",
		"__able_compiled_fn_keep_i8",
		"__able_compiled_fn_keep_i16",
		"__able_compiled_fn_keep_i32",
		"__able_compiled_fn_keep_i64",
		"__able_compiled_fn_keep_i128",
		"__able_compiled_fn_keep_isize",
		"__able_compiled_fn_keep_u8",
		"__able_compiled_fn_keep_u16",
		"__able_compiled_fn_keep_u32",
		"__able_compiled_fn_keep_u64",
		"__able_compiled_fn_keep_u128",
		"__able_compiled_fn_keep_usize",
		"__able_compiled_fn_keep_f32",
		"__able_compiled_fn_keep_f64",
		"__able_compiled_fn_keep_char",
		"__able_compiled_fn_keep_string",
	} {
		body := mustCompiledFunctionBody(t, result, function)
		assertBodyAvoidsFragments(t, function, body, []string{
			"runtime.Value",
			"runtime.ArrayStore",
			"bridge.As",
			"bridge.To",
			"__able_any_to_value(",
			"__able_call_named(",
			"__able_call_value(",
			"__able_try_cast(",
		})
	}
}
