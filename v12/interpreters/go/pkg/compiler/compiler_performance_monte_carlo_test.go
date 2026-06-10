package compiler

import (
	"strings"
	"testing"
)

func TestCompilerSignedIntegerHelpersRenderPositiveFastPaths(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn main() -> void {",
		"  print((42_i64 * 48271_i64) % 2147483647_i64)",
		"}",
	}, "\n"))

	compiledSrc := string(result.Files["compiled.go"])
	for _, fragment := range []string{
		"if a >= 0 && b >= 0 {",
		"if a >= 0 && b > 0 {",
	} {
		if !strings.Contains(compiledSrc, fragment) {
			t.Fatalf("expected compiled runtime helper fragment %q:\n%s", fragment, compiledSrc)
		}
	}
}

func TestCompilerMonteCarloRecurrenceExecutes(t *testing.T) {
	source := strings.Join([]string{
		"package demo",
		"",
		"fn step(state: i64) -> i64 {",
		"  (state * 48271_i64) % 2147483647_i64",
		"}",
		"",
		"fn main() -> void {",
		"  print(`${step(1_i64)} ${step(2_i64)}`)",
		"}",
	}, "\n")

	stdout := compileAndRunSourceWithOptions(t, "ablec-monte-carlo-recurrence-", source, Options{
		PackageName: "main",
		EmitMain:    true,
	})
	if strings.TrimSpace(stdout) != "48271 96542" {
		t.Fatalf("expected Monte Carlo recurrence output, got %q", stdout)
	}
}
