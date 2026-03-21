package compiler

import (
	"strings"
	"testing"
)

func TestCompilerControlToErrorPreservesExitSignals(t *testing.T) {
	result := compileNoFallbackExecSource(t, "ablec-control-bridge-exit", strings.Join([]string{
		"package demo",
		"",
		"fn main() -> void {}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_control_to_error")
	if !ok {
		t.Fatalf("could not find __able_control_to_error helper")
	}
	if !strings.Contains(body, "interpreter.ExitCodeFromError(control.Err)") {
		t.Fatalf("expected control bridge to preserve exit signals before wrapping raised values:\n%s", body)
	}
}
