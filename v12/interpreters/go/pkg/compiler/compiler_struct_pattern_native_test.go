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
