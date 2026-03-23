package compiler

import (
	"strings"
	"testing"
)

func TestCompilerStructPatternNamedFieldBindingStaysNative(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"struct Box { value: i32 }",
		"",
		"fn main() -> i32 {",
		"  value := Box { value: 7 }",
		"  value match {",
		"    case Box { value::n } => n,",
		"    case _ => 0",
		"  }",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	if strings.Contains(body, "__able_env_get(\"n\"") {
		t.Fatalf("expected named struct-pattern binding to avoid runtime env lookup:\n%s", body)
	}
	if !strings.Contains(body, "var n int32 =") {
		t.Fatalf("expected named struct-pattern binding to stay on a native local:\n%s", body)
	}
	if strings.Contains(body, "var value int32 = __able_tmp_0.Value") {
		t.Fatalf("expected named struct-pattern binding to bind n rather than rebinding the field name:\n%s", body)
	}
}

func TestCompilerStructPatternNamedFieldBindingExecutes(t *testing.T) {
	source := strings.Join([]string{
		"extern go fn __able_os_exit(code: i32) -> void {}",
		"",
		"struct Box { value: i32 }",
		"",
		"fn main() {",
		"  result := Box { value: 7 } match {",
		"    case Box { value::n } => n,",
		"    case _ => 0",
		"  }",
		"  if result == 7 {",
		"    __able_os_exit(0)",
		"  }",
		"  __able_os_exit(1)",
		"}",
		"",
	}, "\n")

	compileAndRunSource(t, "ablec-struct-pattern-native-", source)
}

func TestCompilerStructPatternFieldBindingPreservesGenericTypeExpr(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"struct Node T { value: T }",
		"struct Box T { items: Array (Node T) }",
		"",
		"fn head(box: Box i32) -> i32 {",
		"  box match {",
		"    case Box { items::items } => {",
		"      item := items.read_slot(0) or { Node { value: 0 } }",
		"      item.value",
		"    },",
		"    case _ => 0,",
		"  }",
		"}",
		"",
		"fn main() -> i32 {",
		"  head(Box { items: [Node { value: 7 }] })",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_head")
	if !ok {
		t.Fatalf("could not find compiled head function")
	}
	if !strings.Contains(body, "var items *__able_array_Node_i32 =") {
		t.Fatalf("expected generic struct-pattern field binding to preserve the specialized array carrier:\n%s", body)
	}
	if !strings.Contains(body, "var item *Node_i32 =") {
		t.Fatalf("expected nested read_slot result to preserve the specialized node carrier:\n%s", body)
	}
	if strings.Contains(body, "var item runtime.Value") {
		t.Fatalf("expected nested read_slot result to avoid runtime.Value fallback:\n%s", body)
	}
}
