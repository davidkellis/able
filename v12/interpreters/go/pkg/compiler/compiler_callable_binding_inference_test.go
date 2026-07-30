package compiler

import (
	"strings"
	"testing"
)

func TestCompilerForwardTypedLocalLambdaStaysNativeThroughNestedUse(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"struct Item { value: i64 }",
		"struct Score { value: i64 }",
		"",
		"fn apply(item: Item, sequence: i64, scorer: (Item, i64) -> Score) -> i64 {",
		"  scorer(item, sequence).value",
		"}",
		"",
		"fn main() -> i64 {",
		"  offset := 5_i64",
		"  scorer := { item, sequence => Score { value: item.value + sequence + offset } }",
		"  if true {",
		"    apply(Item { value: 35_i64 }, 2_i64, scorer)",
		"  } else {",
		"    0_i64",
		"  }",
		"}",
		"",
	}, "\n"))

	body := mustCompiledFunctionBody(t, result, "__able_compiled_fn_main")
	if !strings.Contains(body, "var scorer __able_fn__Item_int64_to__Score = __able_fn__Item_int64_to__Score(") {
		t.Fatalf("expected forward typed lambda to keep its concrete native callable carrier:\n%s", body)
	}
	assertBodyAvoidsFragments(t, "__able_compiled_fn_main", body, []string{
		"__able_fn_runtime_Value_runtime_Value_to__Score",
		"__able_fn_runtime_Value_runtime_Value_to_runtime_Value",
		"__able_call_value(",
		"__able_fn__Item_int64_to__Score_to_runtime_value(__able_runtime, scorer)",
		"__able_fn__Item_int64_to__Score_from_runtime_value",
	})
}

func TestCompilerForwardTypedLocalLambdaKeepsDynamicUseErased(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn retain(value: any) -> any { value }",
		"",
		"fn main() -> i32 {",
		"  callback := { value => value }",
		"  retained := retain(callback)",
		"  0",
		"}",
		"",
	}, "\n"))

	body := mustCompiledFunctionBody(t, result, "__able_compiled_fn_main")
	if !strings.Contains(body, "var callback __able_fn_runtime_Value_to_runtime_Value") {
		t.Fatalf("expected explicitly dynamic callable use to retain the erased carrier:\n%s", body)
	}
	if !strings.Contains(body, "__able_fn_runtime_Value_to_runtime_Value_to_runtime_value") {
		t.Fatalf("expected explicitly dynamic callable use to retain its runtime adapter:\n%s", body)
	}
}

func TestCompilerForwardTypedLocalLambdaRejectsConflictingCallableUses(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn apply_i32(callback: i32 -> i32) -> i32 { callback(1) }",
		"fn apply_i64(callback: i64 -> i64) -> i64 { callback(2_i64) }",
		"",
		"fn main() -> i64 {",
		"  callback := { value => value }",
		"  (apply_i32(callback) as i64) + apply_i64(callback)",
		"}",
		"",
	}, "\n"))

	body := mustCompiledFunctionBody(t, result, "__able_compiled_fn_main")
	if !strings.Contains(body, "var callback __able_fn_runtime_Value_to_runtime_Value") {
		t.Fatalf("expected conflicting callable constraints to retain the erased carrier:\n%s", body)
	}
	for _, fragment := range []string{
		"__able_fn_int32_to_int32_from_runtime_value",
		"__able_fn_int64_to_int64_from_runtime_value",
	} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("expected conflicting callable constraints to preserve %q adapter:\n%s", fragment, body)
		}
	}
}

func TestCompilerForwardTypedLocalLambdaUsesImportedCallableConstraint(t *testing.T) {
	result := compileNoFallbackPackage(t, "demo", map[string]string{
		"main.able": strings.Join([]string{
			"package demo",
			"",
			"import demo.remote.{Item, Score, apply}",
			"",
			"fn main() -> i64 {",
			"  scorer := { item, sequence => Score { value: item.value + sequence } }",
			"  apply(Item { value: 40_i64 }, scorer)",
			"}",
			"",
		}, "\n"),
		"remote/module.able": strings.Join([]string{
			"struct Item { value: i64 }",
			"struct Score { value: i64 }",
			"",
			"fn apply(item: Item, scorer: (Item, i64) -> Score) -> i64 {",
			"  scorer(item, 2_i64).value",
			"}",
			"",
		}, "\n"),
	})

	body := mustCompiledFunctionBody(t, result, "__able_compiled_fn_main")
	start := strings.Index(body, "var scorer ")
	if start < 0 {
		t.Fatalf("expected imported callable constraint to type scorer:\n%s", body)
	}
	end := strings.IndexByte(body[start:], '\n')
	if end < 0 {
		end = len(body) - start
	}
	declaration := body[start : start+end]
	if strings.Contains(declaration, "runtime.Value") || !strings.Contains(declaration, "__able_fn_") {
		t.Fatalf("expected imported callable constraint to preserve a concrete carrier:\n%s", declaration)
	}
	assertBodyAvoidsFragments(t, "__able_compiled_fn_main", body, []string{
		"__able_call_value(",
		"_to_runtime_value(__able_runtime, scorer)",
		"_from_runtime_value(__able_runtime",
	})
}

func TestCompilerForwardTypedLocalLambdaInfersDirectInvocation(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"struct Item { value: i64 }",
		"",
		"fn main() -> i64 {",
		"  offset := 5_i64",
		"  scorer := { left, right => left.value + right.value + offset }",
		"  scorer(Item { value: 20_i64 }, Item { value: 17_i64 })",
		"}",
		"",
	}, "\n"))

	body := mustCompiledFunctionBody(t, result, "__able_compiled_fn_main")
	if !strings.Contains(body, "var scorer __able_fn__Item__Item_to_int64 = __able_fn__Item__Item_to_int64(") {
		t.Fatalf("expected the direct monomorphic invocation to infer a native callable carrier:\n%s", body)
	}
	assertBodyAvoidsFragments(t, "__able_compiled_fn_main", body, []string{
		"__able_fn_runtime_Value_runtime_Value_to_runtime_Value",
		"__able_call_value(",
		"__able_binary_op(",
		"__able_struct_Item_to_seen",
	})
}

func TestCompilerForwardTypedLocalLambdaStaysNativeThroughStaticInlineCallback(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"struct Item { value: i64 }",
		"struct Score { value: i64 }",
		"",
		"fn apply(item: Item, callback: Item -> i64) -> i64 { callback(item) }",
		"",
		"fn main() -> i64 {",
		"  offset := 5_i64",
		"  scorer := { item, sequence => Score { value: item.value + sequence + offset } }",
		"  apply(Item { value: 35_i64 }, { current => scorer(current, 2_i64).value })",
		"}",
		"",
	}, "\n"))

	body := mustCompiledFunctionBody(t, result, "__able_compiled_fn_main")
	if !strings.Contains(body, "var scorer __able_fn__Item_int64_to__Score = __able_fn__Item_int64_to__Score(") {
		t.Fatalf("expected the captured callable to stay native through a static inline callback:\n%s", body)
	}
	assertBodyAvoidsFragments(t, "__able_compiled_fn_main", body, []string{
		"__able_fn_runtime_Value_runtime_Value_to__Score",
		"__able_fn_runtime_Value_runtime_Value_to_runtime_Value",
		"__able_call_value(",
		"__able_struct_Item_to_seen",
	})
}

func TestCompilerForwardTypedLocalLambdaStaysNativeThroughGenericUnionCallback(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"struct Item { value: i64 }",
		"struct Score { value: i64 }",
		"union Choice T = nil | T",
		"",
		"methods Choice T {",
		"  fn map<U>(self: Self, f: T -> U) -> Choice U {",
		"    self match {",
		"      case nil => nil,",
		"      case value => f(value),",
		"    }",
		"  }",
		"}",
		"",
		"fn source() -> Choice Item { Item { value: 35_i64 } }",
		"",
		"fn main() -> i64 {",
		"  offset := 5_i64",
		"  scorer := { item, sequence => Score { value: item.value + sequence + offset } }",
		"  mapped := source().map<Score>({ current => scorer(current, 2_i64) })",
		"  mapped match {",
		"    case nil => 0_i64,",
		"    case score => score.value,",
		"  }",
		"}",
		"",
	}, "\n"))

	body := mustCompiledFunctionBody(t, result, "__able_compiled_fn_main")
	if !strings.Contains(body, "var scorer __able_fn__Item_int64_to__Score = __able_fn__Item_int64_to__Score(") {
		t.Fatalf("expected the generic-union callback to preserve the captured callable carrier:\n%s", body)
	}
	assertBodyAvoidsFragments(t, "__able_compiled_fn_main", body, []string{
		"__able_fn_runtime_Value_runtime_Value_to__Score",
		"__able_call_value(",
		"__able_struct_Item_to_seen",
	})
}

func TestCompilerForwardTypedLocalLambdaTracksFutureBlockBinding(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"struct Item { value: i64 }",
		"struct Score { value: i64 }",
		"union Choice T = nil | T",
		"",
		"methods Choice T {",
		"  fn map<U>(self: Self, f: T -> U) -> Choice U {",
		"    self match {",
		"      case nil => nil,",
		"      case value => f(value),",
		"    }",
		"  }",
		"}",
		"",
		"fn source() -> Choice Item { Item { value: 35_i64 } }",
		"",
		"fn main() -> i64 {",
		"  offset := 5_i64",
		"  scorer := { item, sequence => Score { value: item.value + sequence + offset } }",
		"  if true {",
		"    resolved := source()",
		"    mapped := resolved.map<Score>({ current => scorer(current, 2_i64) })",
		"    mapped match {",
		"      case nil => 0_i64,",
		"      case score => score.value,",
		"    }",
		"  } else {",
		"    0_i64",
		"  }",
		"}",
		"",
	}, "\n"))

	body := mustCompiledFunctionBody(t, result, "__able_compiled_fn_main")
	if !strings.Contains(body, "var scorer __able_fn__Item_int64_to__Score = __able_fn__Item_int64_to__Score(") {
		t.Fatalf("expected a future block binding to preserve the captured callable carrier:\n%s", body)
	}
	assertBodyAvoidsFragments(t, "__able_compiled_fn_main", body, []string{
		"__able_fn_runtime_Value_runtime_Value_to__Score",
		"__able_call_value(",
		"__able_struct_Item_to_seen",
	})
}

func TestCompilerForwardTypedLocalLambdaKeepsStoredNestedUseErased(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"struct Item { value: i64 }",
		"struct Score { value: i64 }",
		"",
		"fn retain(value: any) -> any { value }",
		"",
		"fn main() -> i64 {",
		"  offset := 5_i64",
		"  scorer := { item, sequence => Score { value: item.value + sequence + offset } }",
		"  deferred := { current => scorer(current, 2_i64).value }",
		"  retained := retain(deferred)",
		"  0_i64",
		"}",
		"",
	}, "\n"))

	body := mustCompiledFunctionBody(t, result, "__able_compiled_fn_main")
	if !strings.Contains(body, "var scorer __able_fn_runtime_Value_runtime_Value_to__Score") {
		t.Fatalf("expected a stored nested callback to preserve the erased captured carrier:\n%s", body)
	}
	if !strings.Contains(body, "__able_binary_op(") || !strings.Contains(body, "bridge.ToDynamicI64") {
		t.Fatalf("expected the stored nested callback to retain runtime member and arithmetic work:\n%s", body)
	}
}

func TestCompilerForwardTypedLocalLambdaRejectsConflictingNestedInvocations(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn apply_i32(callback: i32 -> i32) -> i32 { callback(1) }",
		"fn apply_i64(callback: i64 -> i64) -> i64 { callback(2_i64) }",
		"",
		"fn main() -> i64 {",
		"  callback := { value => value }",
		"  left := apply_i32({ ignored => callback(1) })",
		"  right := apply_i64({ ignored => callback(2_i64) })",
		"  (left as i64) + right",
		"}",
		"",
	}, "\n"))

	body := mustCompiledFunctionBody(t, result, "__able_compiled_fn_main")
	if !strings.Contains(body, "var callback __able_fn_runtime_Value_to_runtime_Value") {
		t.Fatalf("expected conflicting nested invocation signatures to retain the erased carrier:\n%s", body)
	}
	for _, fragment := range []string{
		"bridge.ToInt(int64(int32(1))",
		"bridge.ToDynamicI64(int64(int64(2)))",
	} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("expected conflicting nested signatures to preserve runtime argument %q:\n%s", fragment, body)
		}
	}
}

func TestCompilerForwardTypedLocalLambdaStaticInlineCallbackExecutes(t *testing.T) {
	source := strings.Join([]string{
		"extern go fn __able_os_exit(code: i32) -> void {}",
		"",
		"struct Item { value: i64 }",
		"struct Score { value: i64 }",
		"",
		"fn apply(item: Item, callback: Item -> i64) -> i64 { callback(item) }",
		"",
		"fn main() {",
		"  offset := 5_i64",
		"  scorer := { item, sequence => Score { value: item.value + sequence + offset } }",
		"  if apply(Item { value: 35_i64 }, { current => scorer(current, 2_i64).value }) == 42_i64 {",
		"    __able_os_exit(0)",
		"  }",
		"  __able_os_exit(1)",
		"}",
		"",
	}, "\n")

	compileAndRunSource(t, "ablec-static-inline-captured-callable-", source)
}
