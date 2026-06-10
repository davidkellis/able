package compiler

import (
	"strings"
	"testing"
)

func TestCompilerControlToErrorPreservesRuntimeErrors(t *testing.T) {
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
	if !strings.Contains(body, "return __able_value_error{value: __able_control_value(control), err: control.Err}") {
		t.Fatalf("expected control bridge to preserve both the raised value and diagnostic error:\n%s", body)
	}
	if strings.Contains(body, "interpreter.ExitCodeFromError(control.Err)") {
		t.Fatalf("expected control bridge to avoid redundant concrete interpreter inspection:\n%s", body)
	}
}

func TestCompilerRaiseControlPreservesNativeErrorValues(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn main() -> void {",
		"  _ = 1",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_raise_control")
	if !ok {
		t.Fatalf("could not find __able_raise_control helper")
	}
	if !strings.Contains(body, "if _, ok, nilPtr := __able_runtime_error_value(value); !ok && !nilPtr {") {
		t.Fatalf("expected raise control to preserve native runtime.ErrorValue carriers before bridge normalization:\n%s", body)
	}
	if !strings.Contains(body, "value = bridge.ErrorValue(__able_runtime, value)") {
		t.Fatalf("expected raise control to retain bridge normalization for non-native errors:\n%s", body)
	}
}
