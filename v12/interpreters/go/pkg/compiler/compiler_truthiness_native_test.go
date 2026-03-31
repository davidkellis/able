package compiler

import (
	"strings"
	"testing"
)

func TestCompilerErrorTruthinessConditionStaysFalsy(t *testing.T) {
	source := strings.Join([]string{
		"package demo",
		"import able.core.errors.{DivisionByZeroError}",
		"",
		"fn main() -> String {",
		"  err := DivisionByZeroError {}",
		"  if err { \"truthy\" } else { \"falsy\" }",
		"}",
		"",
	}, "\n")

	result := compileNoFallbackExecSource(t, "ablec-error-truthiness-native", source)
	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main body")
	}
	if strings.Contains(body, "if err != nil {") {
		t.Fatalf("expected native Error truthiness to avoid nil-check lowering on the condition value:\n%s", body)
	}
	if !strings.Contains(body, "if false {") {
		t.Fatalf("expected native Error truthiness to lower as statically falsy:\n%s", body)
	}
}

func TestCompilerIfTruthinessValueErrorAndZeroExecutes(t *testing.T) {
	stdout := compileAndRunExecSourceWithOptions(t, "ablec-if-truthiness-value", strings.Join([]string{
		"package demo",
		"import able.core.errors.{DivisionByZeroError}",
		"",
		"fn main() -> void {",
		"  err := DivisionByZeroError {}",
		"  chain := if err { \"error\" } elsif 0 { \"zero\" } else { \"fallback\" }",
		"  result := if err { \"err truthy\" } else { \"err falsy\" }",
		"  print(chain)",
		"  print(result)",
		"}",
		"",
	}, "\n"), Options{
		PackageName: "main",
		EmitMain:    true,
	})

	if strings.TrimSpace(stdout) != "zero\nerr falsy" {
		t.Fatalf("expected compiled truthiness program to print zero/err falsy, got %q", stdout)
	}
}

func TestCompilerIfWithoutElseStringInterpolationKeepsNilCapableCarrier(t *testing.T) {
	result := compileNoFallbackExecSource(t, "ablec-if-without-else-stringify", strings.Join([]string{
		"package demo",
		"",
		"fn main() -> void {",
		"  no_else := if false { 123 }",
		"  print(`no_else ${no_else}`)",
		"}",
		"",
	}, "\n"))
	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main body")
	}
	if strings.Contains(body, "bridge.AsInt(no_else") {
		t.Fatalf("expected nil-capable if result to avoid integer recovery during string interpolation:\n%s", body)
	}
	if !strings.Contains(body, "__able_stringify(") {
		t.Fatalf("expected nil-capable if result to stringify through runtime value conversion:\n%s", body)
	}
}
