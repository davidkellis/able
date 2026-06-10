package compiler

import (
	"strings"
	"testing"
)

func TestCompilerSpecNullableExpectationBuilds(t *testing.T) {
	// Keep the generated-Go build guard focused on the matcher/interface and
	// nullable-result shape. Loading the entire canonical able.spec wildcard
	// graph plus building its generated Go exceeds the per-test time limit; the
	// exact canonical import is still lowered in the companion test below.
	_ = compileAndBuildStdlibSource(t, "ablec-spec-nullable-matcher-", strings.Join([]string{
		"package demo",
		"",
		"interface Matcher T for M {",
		"  fn matches(self: Self, actual: T) -> bool",
		"}",
		"",
		"struct Expectation T { actual: T }",
		"",
		"fn expect<T>(value: T) -> Expectation T {",
		"  Expectation T { actual: value }",
		"}",
		"",
		"methods Expectation T {",
		"  fn to(self: Self, matcher: Matcher T) -> void {",
		"    _ = matcher.matches(self.actual)",
		"  }",
		"}",
		"",
		"struct EqMatcher T { expected: T }",
		"",
		"fn eq<T>(expected: T) -> EqMatcher T {",
		"  EqMatcher T { expected }",
		"}",
		"",
		"impl Matcher T for EqMatcher T {",
		"  fn matches(self: Self, actual: T) -> bool { actual == self.expected }",
		"}",
		"",
		"struct BeNilMatcher {}",
		"",
		"fn be_nil() -> BeNilMatcher { BeNilMatcher }",
		"",
		"impl Matcher (?Value) for BeNilMatcher {",
		"  fn matches(self: Self, actual: ?Value) -> bool { actual == nil }",
		"}",
		"",
		"fn fetch(flag: bool) -> ?i32 {",
		"  if flag {",
		"    1",
		"  } else {",
		"    nil",
		"  }",
		"}",
		"",
		"fn main() -> void {",
		"  expect(fetch(true)).to(eq(1))",
		"  expect(fetch(false)).to(be_nil())",
		"}",
		"",
	}, "\n"))
}

func TestCompilerCanonicalSpecNullableExpectationLowers(t *testing.T) {
	// This retains the exact able.spec integration boundary while the test above
	// independently proves that the resulting structural shape builds as Go.
	_ = compileSourceWithStdlibPaths(t, strings.Join([]string{
		"package demo",
		"",
		"import able.spec.*",
		"",
		"fn fetch(flag: bool) -> ?i32 {",
		"  if flag {",
		"    1",
		"  } else {",
		"    nil",
		"  }",
		"}",
		"",
		"fn main() -> void {",
		"  expect(fetch(true)).to(eq(1))",
		"  expect(fetch(false)).to(be_nil())",
		"}",
		"",
	}, "\n"))
}
