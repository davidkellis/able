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
	concreteProbe := strings.Index(compiledSrc, "bridge.MatchType(rt, ast.Ty(\"BeWithinMatcher\"), base)")
	siblingProbe := strings.Index(compiledSrc, "bridge.MatchType(rt, ast.Gen(ast.Ty(\"Matcher\"), ast.Ty(\"f64\")), base)")
	if concreteProbe < 0 || siblingProbe < 0 || concreteProbe > siblingProbe {
		t.Fatalf("expected Matcher<Result<f64>> recovery to probe its concrete matcher before the sibling interface")
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

func TestCompilerWideIntegerMatcherAdapterBoxesOnlyAtGenericBoundary(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"interface Matcher T for Self {",
		"  fn matches(self: Self, value: T) -> bool",
		"}",
		"",
		"struct EqMatcher T { expected: T }",
		"",
		"impl Matcher T for EqMatcher T {",
		"  fn matches(self: Self, value: T) -> bool { value == self.expected }",
		"}",
		"",
		"fn accept_i128(matcher: Matcher i128, value: i128) -> bool {",
		"  matcher.matches(value)",
		"}",
		"",
		"fn accept_u128(matcher: Matcher u128, value: u128) -> bool {",
		"  matcher.matches(value)",
		"}",
		"",
		"fn main() -> bool {",
		"  accept_i128(EqMatcher { expected: -170141183460469231731687303715884105728_i128 }, -170141183460469231731687303715884105728_i128) &&",
		"    accept_u128(EqMatcher { expected: 340282366920938463463374607431768211455_u128 }, 340282366920938463463374607431768211455_u128)",
		"}",
		"",
	}, "\n"))

	compiledSrc := string(result.Files["compiled.go"])
	for _, fragment := range []string{
		"__able_iface_Matcher_i128_adapter_ptr_EqMatcher",
		"__able_iface_Matcher_u128_adapter_ptr_EqMatcher",
		"arg0.IntegerValue()",
	} {
		if !strings.Contains(compiledSrc, fragment) {
			t.Fatalf("expected wide primitive matcher boundary to contain %q:\n%s", fragment, compiledSrc)
		}
	}
	if strings.Contains(compiledSrc, "unsupported native interface conversion from runtime.Int128") ||
		strings.Contains(compiledSrc, "unsupported native interface conversion from runtime.Uint128") {
		t.Fatalf("expected wide primitive matcher adapters to use the shared runtime integer encoding")
	}
}

func TestCompilerSpecializedGenericStructBoundaryPreservesRuntimeTypeArguments(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"interface Matcher T for Self {",
		"  fn matches(self: Self, value: T) -> bool",
		"}",
		"",
		"struct EqMatcher T { expected: T }",
		"",
		"impl Matcher T for EqMatcher T {",
		"  fn matches(self: Self, value: T) -> bool { value == self.expected }",
		"}",
		"",
		"fn accept(matcher: Matcher (Result i32), actual: Result i32) -> bool {",
		"  matcher.matches(actual)",
		"}",
		"",
		"fn main() -> bool {",
		"  accept(EqMatcher { expected: 42 }, 42)",
		"}",
		"",
	}, "\n"))

	compiledSrc := string(result.Files["compiled.go"])
	toSeen, ok := findCompiledFunction(result, "__able_struct_EqMatcher_Result_i32_to_seen")
	if !ok {
		t.Fatalf("expected specialized EqMatcher<Result<i32>> runtime converter:\n%s", compiledSrc)
	}
	if !strings.Contains(toSeen, "[]ast.TypeExpression{ast.Result(ast.Ty(\"i32\"))}") {
		t.Fatalf("expected specialized generic struct runtime conversion to retain Result<i32> identity:\n%s", toSeen)
	}
	apply, ok := findCompiledFunction(result, "__able_struct_EqMatcher_Result_i32_apply")
	if !ok {
		t.Fatalf("expected specialized EqMatcher<Result<i32>> apply helper")
	}
	if !strings.Contains(apply, "inst.TypeArguments = updated.TypeArguments") {
		t.Fatalf("expected specialized generic struct apply to retain runtime type arguments:\n%s", apply)
	}
}
