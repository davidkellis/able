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

func TestCompilerUnaryPrimitiveBranchPreservesCarrierBeforeResultWrap(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn signed(value: i64, negative: bool) -> Result i64 {",
		"  if negative { -value } else { value }",
		"}",
		"",
		"fn flipped(value: u64, flip: bool) -> Result u64 {",
		"  if flip { .~value } else { value }",
		"}",
		"",
		"fn main() -> i64 { 0 }",
		"",
	}, "\n"))

	compiled := string(result.Files["compiled.go"])
	for _, name := range []string{"__able_compiled_fn_signed", "__able_compiled_fn_flipped"} {
		body, ok := findCompiledFunction(result, name)
		if !ok {
			t.Fatalf("could not find compiled function %s", name)
		}
		if strings.Contains(body, "__able_unary_op(") {
			t.Fatalf("%s boxed a statically known primitive unary operand:\n%s", name, body)
		}
	}
	signedBody, _ := findCompiledFunction(result, "__able_compiled_fn_signed")
	if !strings.Contains(signedBody, "__able_checked_sub_signed(") ||
		!strings.Contains(signedBody, "__able_union_int64_or_runtime_ErrorValue_wrap_int64(") {
		t.Fatalf("expected native checked i64 negation followed by Result wrapping:\n%s", signedBody)
	}
	if strings.Contains(compiled, "\"able/interpreter-go/pkg/interpreter\"") {
		t.Fatalf("static primitive unary lowering retained the interpreter package")
	}
}
