package compiler

import (
	"strings"
	"testing"
)

func TestCompilerNativeInterfaceBoundaryCachesRuntimeTypeExpression(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"interface Matcher T for Self {",
		"  fn matches(self: Self, value: T) -> bool",
		"}",
		"",
		"struct ExactMatcher { expected: i64 }",
		"",
		"impl Matcher i64 for ExactMatcher {",
		"  fn matches(self: Self, value: i64) -> bool {",
		"    value == self.expected",
		"  }",
		"}",
		"",
		"fn accept(value: Matcher i64) -> bool {",
		"  value.matches(7)",
		"}",
		"",
		"fn main() -> bool {",
		"  accept(ExactMatcher { expected: 7 })",
		"}",
		"",
	}, "\n"))

	compiledSrc := string(result.Files["compiled.go"])
	declaration := "var __able_iface_Matcher_i64_runtime_type ast.TypeExpression = ast.Gen(ast.Ty(\"Matcher\"), ast.Ty(\"i64\"))"
	if !strings.Contains(compiledSrc, declaration) {
		t.Fatalf("expected the native interface runtime type expression to be initialized once:\n%s", compiledSrc)
	}
	helperStart := strings.Index(compiledSrc, "func __able_iface_Matcher_i64_try_from_value(")
	if helperStart < 0 {
		t.Fatalf("expected Matcher<i64> boundary helper:\n%s", compiledSrc)
	}
	helperEnd := strings.Index(compiledSrc[helperStart:], "\n}\n")
	if helperEnd < 0 {
		t.Fatalf("expected complete Matcher<i64> boundary helper:\n%s", compiledSrc[helperStart:])
	}
	helper := compiledSrc[helperStart : helperStart+helperEnd]
	if !strings.Contains(helper, "bridge.MatchType(rt, __able_iface_Matcher_i64_runtime_type, value)") {
		t.Fatalf("expected boundary matching to reuse cached runtime type metadata:\n%s", helper)
	}
	if strings.Contains(helper, "bridge.MatchType(rt, ast.Gen(ast.Ty(\"Matcher\"), ast.Ty(\"i64\")), value)") {
		t.Fatalf("expected no per-call construction of the interface type expression:\n%s", helper)
	}
}
