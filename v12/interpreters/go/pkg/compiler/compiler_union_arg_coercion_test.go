package compiler

import (
	"strings"
	"testing"
)

func TestCompilerStaticCallCoercesUniqueResultMemberArg(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn expect(value: Result i64) -> i64 {",
		"  value match {",
		"    case n: i64 => n,",
		"    case _: Error => 0",
		"  }",
		"}",
		"",
		"fn main() -> i64 {",
		"  value: i32 = 42",
		"  expect(value)",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	if strings.Contains(body, "__able_compiled_fn_expect(value)") {
		t.Fatalf("expected result-typed call argument to avoid passing the raw i32 binding:\n%s", body)
	}
	expectedWrap := "__able_union_int64_or_runtime_ErrorValue_wrap_int64(int64(value))"
	if !strings.Contains(body, expectedWrap) {
		t.Fatalf("expected result-typed call argument to retarget onto the unique i64 member before wrapping:\n%s", body)
	}
}
