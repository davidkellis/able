package compiler

import (
	"strings"
	"testing"
)

func TestCompilerOutOfRangeExplicitU128LiteralRaisesAtRuntime(t *testing.T) {
	stdout := compileAndRunExecSourceWithOptions(t, "ablec-u128-literal-overflow", strings.Join([]string{
		"package demo",
		"",
		"fn overflow() -> u128 {",
		"  340282366920938463463374607431768211456_u128",
		"}",
		"",
		"fn main() -> void {",
		"  recovered := overflow() rescue {",
		"    case _ => 7_u128",
		"  }",
		"  print(recovered)",
		"}",
		"",
	}, "\n"), Options{
		PackageName:        "main",
		EmitMain:           true,
		RequireNoFallbacks: true,
	})

	if stdout != "7\n" {
		t.Fatalf("expected compiled u128 literal overflow to be recoverable, got %q", stdout)
	}
}

func TestCompilerFloatLiteralMultiplicationOverflowsAtRuntime(t *testing.T) {
	stdout := compileAndRunExecSourceWithOptions(t, "ablec-float-literal-overflow", strings.Join([]string{
		"package demo",
		"",
		"fn main() -> void {",
		"  infinite := 1e308 * 1e308",
		"  print(infinite > 1e308)",
		"}",
		"",
	}, "\n"), Options{
		PackageName:        "main",
		EmitMain:           true,
		RequireNoFallbacks: true,
	})

	if stdout != "true\n" {
		t.Fatalf("expected compiled floating multiplication to produce positive infinity, got %q", stdout)
	}
}
