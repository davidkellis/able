package compiler

import (
	"strings"
	"testing"
)

func TestCompilerRangeExpressionAvoidsRangeStructRecoercion(t *testing.T) {
	result := compileNoFallbackExecSource(t, "ablec-range-native-carrier", strings.Join([]string{
		"package demo",
		"",
		"fn main() -> void {",
		"  values := 1..4",
		"  for v in values {",
		"    print(v)",
		"  }",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	if !strings.Contains(body, "__able_range(") {
		t.Fatalf("expected range literal helper to be used on the compiled path:\n%s", body)
	}
	for _, fragment := range []string{
		"__able_struct_Range_from(__able_range(",
		"var values *Range",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected plain range lowering to avoid bogus nominal Range recoercion %q:\n%s", fragment, body)
		}
	}
}

func TestCompilerRangeExpressionExecutesInForLoop(t *testing.T) {
	stdout := strings.TrimSpace(compileAndRunExecSourceWithOptions(t, "ablec-range-for-loop-exec", strings.Join([]string{
		"package demo",
		"",
		"fn main() -> void {",
		"  values := 1..4",
		"  i := 0",
		"  second := 0",
		"  for v in values {",
		"    if i == 0 || i == 2 { print(v) }",
		"    if i == 1 { second = v }",
		"    i = i + 1",
		"  }",
		"  casted := 3.7 as i32",
		"  print(casted)",
		"  print(`sum ${casted + second}`)",
		"}",
		"",
	}, "\n"), Options{
		PackageName:        "main",
		RequireNoFallbacks: true,
		EmitMain:           true,
	}))
	if stdout != "1\n3\n3\nsum 5" {
		t.Fatalf("expected compiled range loop to execute correctly, got %q", stdout)
	}
}
