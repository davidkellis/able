package compiler

import (
	"strings"
	"testing"
)

func TestCompilerResultMatcherBoundaryHelperReusesNativeMatcherAdapter(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"interface Matcher T for Self {",
		"  fn matches(self: Self, value: T) -> bool",
		"}",
		"",
		"struct BeWithinMatcher {}",
		"",
		"impl Matcher f64 for BeWithinMatcher {",
		"  fn matches(self: Self, value: f64) -> bool {",
		"    true",
		"  }",
		"}",
		"",
		"fn accept_plain(value: Matcher f64) -> bool {",
		"  value.matches(1.0)",
		"}",
		"",
		"fn accept_result(value: Matcher (Result f64)) -> bool {",
		"  value.matches(1.0)",
		"}",
		"",
		"fn main() -> bool {",
		"  accept_plain(BeWithinMatcher {}) && accept_result(BeWithinMatcher {})",
		"}",
		"",
	}, "\n"))

	compiledSrc := string(result.Files["compiled.go"])
	if !strings.Contains(compiledSrc, "func __able_iface_Matcher_Result_f64__try_from_value(") {
		t.Fatalf("expected Matcher<Result<f64>> boundary helper to be materialized:\n%s", compiledSrc)
	}
	if !strings.Contains(compiledSrc, "bridge.MatchType(rt, ast.Gen(ast.Ty(\"Matcher\"), ast.Ty(\"f64\")), base)") {
		t.Fatalf("expected Matcher<Result<f64>> boundary helper to probe the native Matcher<f64> carrier first:\n%s", compiledSrc)
	}
	if !strings.Contains(compiledSrc, "__able_iface_Matcher_f64_from_value(__able_runtime, coerced)") {
		t.Fatalf("expected Matcher<Result<f64>> boundary helper path to reuse the native Matcher<f64> carrier path:\n%s", compiledSrc)
	}
	if !strings.Contains(compiledSrc, "__able_iface_Matcher_Result_f64__wrap___able_iface_Matcher_f64(converted)") {
		t.Fatalf("expected Matcher<Result<f64>> boundary helper path to wrap the recovered Matcher<f64> carrier directly:\n%s", compiledSrc)
	}
}

func TestCompilerSiblingMatcherBoundaryHelperProbesConcreteSiblingAdapters(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"interface Matcher T for Self {",
		"  fn matches(self: Self, value: T) -> bool",
		"}",
		"",
		"struct BeBetweenMatcher T { lower: T, upper: T }",
		"",
		"impl Matcher i64 for BeBetweenMatcher i64 {",
		"  fn matches(self: Self, value: i64) -> bool {",
		"    value >= self.lower && value <= self.upper",
		"  }",
		"}",
		"",
		"struct CustomMatcher T {}",
		"",
		"impl Matcher i64 for CustomMatcher i64 {",
		"  fn matches(self: Self, value: i64) -> bool {",
		"    value % 2 == 0",
		"  }",
		"}",
		"",
		"fn accept(value: Matcher i32) -> bool {",
		"  value.matches(4)",
		"}",
		"",
		"fn main() -> bool {",
		"  accept(BeBetweenMatcher { lower: 1, upper: 10 }) && accept(CustomMatcher {})",
		"}",
		"",
	}, "\n"))

	compiledSrc := string(result.Files["compiled.go"])
	if !strings.Contains(compiledSrc, "func __able_iface_Matcher_i32_try_from_value(") {
		t.Fatalf("expected Matcher<i32> boundary helper to be materialized:\n%s", compiledSrc)
	}
	if !strings.Contains(compiledSrc, "bridge.MatchType(rt, ast.Gen(ast.Ty(\"BeBetweenMatcher\"), ast.Ty(\"i64\")), base)") {
		t.Fatalf("expected Matcher<i32> boundary helper to probe concrete BeBetweenMatcher<i64> values:\n%s", compiledSrc)
	}
	if !strings.Contains(compiledSrc, "__able_iface_Matcher_i32_wrap___able_iface_Matcher_i64(__able_iface_Matcher_i64_wrap_ptr_BeBetweenMatcher_i64(converted))") {
		t.Fatalf("expected Matcher<i32> boundary helper to bridge recovered BeBetweenMatcher<i64> values through Matcher<i64>:\n%s", compiledSrc)
	}
	if !strings.Contains(compiledSrc, "bridge.MatchType(rt, ast.Gen(ast.Ty(\"CustomMatcher\"), ast.Ty(\"i64\")), base)") {
		t.Fatalf("expected Matcher<i32> boundary helper to probe concrete CustomMatcher<i64> values:\n%s", compiledSrc)
	}
	if !strings.Contains(compiledSrc, "__able_iface_Matcher_i32_wrap___able_iface_Matcher_i64(__able_iface_Matcher_i64_wrap_ptr_CustomMatcher_i64(converted))") {
		t.Fatalf("expected Matcher<i32> boundary helper to bridge recovered CustomMatcher<i64> values through Matcher<i64>:\n%s", compiledSrc)
	}
}
