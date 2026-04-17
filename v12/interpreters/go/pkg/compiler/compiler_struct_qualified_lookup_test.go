package compiler

import (
	"strings"
	"testing"
)

func TestCompilerImportedMatcherStructConvertersUseQualifiedStructLookup(t *testing.T) {
	result := compileSourceWithCanonicalStdlibPaths(t, strings.Join([]string{
		"package demo",
		"",
		"import able.spec.{be_between, matcher}",
		"",
		"fn main() -> void {",
		"  be_between(1, 10)",
		"  matcher(\"ok\", \"not ok\", fn(value: i64) -> bool { value > 0 })",
		"}",
		"",
	}, "\n"))

	beBetweenBody, ok := findCompiledFunction(result, "__able_struct_BeBetweenMatcher_i64_to_seen")
	if !ok {
		t.Fatalf("expected BeBetweenMatcher<i64> struct converter to be generated")
	}
	if !strings.Contains(beBetweenBody, `rt.StructDefinition("able.spec.assertions.BeBetweenMatcher")`) {
		t.Fatalf("expected BeBetweenMatcher<i64> converter to use a qualified struct lookup:\n%s", beBetweenBody)
	}

	customBody, ok := findCompiledFunction(result, "__able_struct_CustomMatcher_i64_to_seen")
	if !ok {
		t.Fatalf("expected CustomMatcher<i64> struct converter to be generated")
	}
	if !strings.Contains(customBody, `rt.StructDefinition("able.spec.assertions.CustomMatcher")`) {
		t.Fatalf("expected CustomMatcher<i64> converter to use a qualified struct lookup:\n%s", customBody)
	}
}
