package compiler

import (
	"strings"
	"testing"
)

func TestRuntimeLookupPackageNameCollapsesRepeatedLeaf(t *testing.T) {
	for _, test := range []struct {
		input string
		want  string
	}{
		{"dependency_wave_validation.dependency_wave_validation", "dependency_wave_validation"},
		{"able.spec.assertions", "able.spec.assertions"},
		{"demo", "demo"},
	} {
		if got := runtimeLookupPackageName(test.input); got != test.want {
			t.Fatalf("runtimeLookupPackageName(%q) = %q, want %q", test.input, got, test.want)
		}
	}
}

func TestRuntimeStructLookupPackageNamesPreserveLoaderAndRuntimeIdentities(t *testing.T) {
	got := runtimeStructLookupPackageNames("dependency_wave_validation.dependency_wave_validation")
	want := []string{"dependency_wave_validation.dependency_wave_validation", "dependency_wave_validation"}
	if len(got) != len(want) {
		t.Fatalf("runtimeStructLookupPackageNames() = %v, want %v", got, want)
	}
	for idx := range want {
		if got[idx] != want[idx] {
			t.Fatalf("runtimeStructLookupPackageNames()[%d] = %q, want %q", idx, got[idx], want[idx])
		}
	}

	got = runtimeStructLookupPackageNames("able.spec.assertions")
	if len(got) != 1 || got[0] != "able.spec.assertions" {
		t.Fatalf("runtimeStructLookupPackageNames(non-repeated) = %v, want one preserved name", got)
	}
}

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
