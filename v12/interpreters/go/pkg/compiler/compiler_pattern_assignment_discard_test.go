package compiler

import (
	"strings"
	"testing"
)

func TestCompilerDiscardedPatternAssignmentsAvoidRuntimeResultMaterialization(t *testing.T) {
	result := compileNoFallbackExecSourceWithOptions(t, "ablec-pattern-assignment-discard-source", strings.Join([]string{
		"package demo",
		"import able.collections.array",
		"import able.kernel.{Array}",
		"",
		"struct Pair { left: String, right: String }",
		"",
		"fn discard_array(values: Array String) -> void {",
		"  [left: String, right: String] := values",
		"  print(left + right)",
		"}",
		"",
		"fn discard_struct(pair: Pair) -> void {",
		"  Pair { left, right } := pair",
		"  print(left + right)",
		"}",
		"",
		"fn observed(values: Array String) -> void {",
		"  if [left: String, right: String] := values {",
		"    print(left + right)",
		"  }",
		"}",
		"",
		"fn main() -> void {}",
		"",
	}, "\n"), Options{
		PackageName:        "main",
		RequireNoFallbacks: true,
	})

	compiled := string(result.Files["compiled.go"])
	arrayBody := extractCompiledFunctionBody(compiled, "fn_discard_array")
	for _, fragment := range []string{
		`__able_raise_control(nil, runtime.ErrorValue{Message: "pattern assignment mismatch"})`,
		"left = ",
		"right = ",
	} {
		if !strings.Contains(arrayBody, fragment) {
			t.Fatalf("discarded Array pattern body missing %q:\n%s", fragment, arrayBody)
		}
	}
	for _, fragment := range []string{
		"__able_array_String_to(",
		"var __able_tmp_",
	} {
		if strings.Contains(arrayBody, fragment) {
			t.Fatalf("discarded Array pattern body unexpectedly materializes result via %q:\n%s", fragment, arrayBody)
		}
	}

	structBody := extractCompiledFunctionBody(compiled, "fn_discard_struct")
	if strings.Contains(structBody, "__able_struct_Pair_to") {
		t.Fatalf("discarded struct pattern body materializes a runtime result:\n%s", structBody)
	}

	observedBody := extractCompiledFunctionBody(compiled, "fn_observed")
	if !strings.Contains(observedBody, "__able_array_String_to(") ||
		!strings.Contains(observedBody, "__able_truthy(") {
		t.Fatalf("observed pattern assignment must preserve its RHS success value:\n%s", observedBody)
	}
}

func TestCompilerDiscardedPatternAssignmentsPreserveSemantics(t *testing.T) {
	source := strings.Join([]string{
		"package demo",
		"import able.collections.array",
		"import able.core.interfaces.{Error}",
		"import able.kernel.{Array}",
		"",
		"struct Pair { left: String, right: String }",
		"struct Bundle { pair: Pair, values: Array String }",
		"",
		"fn make_bundle() -> Bundle {",
		"  print(\"rhs\")",
		"  Bundle { pair: Pair { left: \"a\", right: \"b\" }, values: [\"c\", \"d\"] }",
		"}",
		"",
		"fn discard_nested() -> void {",
		"  Bundle { pair::pair_value: Pair { left, right }, values } := make_bundle()",
		"  [first: String, second: String] := values",
		"  print(left + right + first + second)",
		"}",
		"",
		"fn mismatch() -> void {",
		"  values := [\"only\"]",
		"  [first: String, second: String] := values",
		"  print(\"unreachable\")",
		"}",
		"",
		"fn main() -> void {",
		"  discard_nested()",
		"  observed := [\"left\", \"right\"]",
		"  if [left: String, right: String] := observed {",
		"    print(left + \":\" + right)",
		"  } else {",
		"    print(\"unexpected\")",
		"  }",
		"  do {",
		"    mismatch()",
		"  } rescue {",
		"    case _: Error => print(\"mismatch\")",
		"  }",
		"}",
		"",
	}, "\n")

	stdout := compileAndRunExecSourceWithOptions(t, "ablec-pattern-assignment-discard-exec", source, Options{
		PackageName:        "main",
		EmitMain:           true,
		RequireNoFallbacks: true,
	})
	const expected = "rhs\nabcd\nleft:right\nmismatch\n"
	if stdout != expected {
		t.Fatalf("discarded pattern assignment output = %q, want %q", stdout, expected)
	}
}
