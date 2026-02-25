package compiler

import (
	"strings"
	"testing"
)

func TestCompilerFormatsCallableValuesAsFunction(t *testing.T) {
	mainSrc, _ := compileOutputs(t, "demo", map[string]string{
		"main.able": strings.Join([]string{
			"package demo",
			"",
			"fn main() -> void {",
			"  print(main)",
			"}",
			"",
		}, "\n"),
	})

	if !strings.Contains(mainSrc, "func isCallableRuntimeValue(val runtime.Value) bool {") {
		t.Fatalf("expected isCallableRuntimeValue helper to be emitted")
	}
	if !strings.Contains(mainSrc, "case runtime.NativeFunctionValue, *runtime.NativeFunctionValue:") {
		t.Fatalf("expected callable helper to normalize native function values")
	}
	if !strings.Contains(mainSrc, "case runtime.NativeBoundMethodValue, *runtime.NativeBoundMethodValue:") {
		t.Fatalf("expected callable helper to normalize native bound method values")
	}

	formatStart := strings.Index(mainSrc, "func formatRuntimeValue(interp *interpreter.Interpreter, val runtime.Value) string {")
	if formatStart < 0 {
		t.Fatalf("expected formatRuntimeValue helper to be emitted")
	}
	formatSegment := mainSrc[formatStart:]
	formatEnd := strings.Index(formatSegment, "func main() {")
	if formatEnd > 0 {
		formatSegment = formatSegment[:formatEnd]
	}
	if !strings.Contains(formatSegment, "if isCallableRuntimeValue(val) {") {
		t.Fatalf("expected formatRuntimeValue to call isCallableRuntimeValue")
	}
	if !strings.Contains(formatSegment, "return \"<function>\"") {
		t.Fatalf("expected formatRuntimeValue to return canonical <function> string for callables")
	}
}
