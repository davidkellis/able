package compiler

import (
	"strings"
	"testing"
)

func raiseNonErrorRescueSource() string {
	return strings.Join([]string{
		"package demo",
		"",
		"fn fail() -> String {",
		"  raise(\"boom\")",
		"  \"ok\"",
		"}",
		"",
		"fn main() -> void {",
		"  fail() rescue {",
		"    case err => print(err.message())",
		"  }",
		"}",
		"",
	}, "\n")
}

func TestCompilerRescueIdentifierBindingFromRaisedStringStaysNativeErrorCarrier(t *testing.T) {
	result := compileNoFallbackSource(t, raiseNonErrorRescueSource())

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	if !strings.Contains(body, "var err runtime.ErrorValue") {
		t.Fatalf("expected rescue binding on raised string payload to stay on the native error carrier:\n%s", body)
	}
	for _, fragment := range []string{
		"var err string",
		"var err runtime.Value",
		"__able_any_to_value(",
		"__able_method_call_node(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected raised-string rescue binding to avoid %q:\n%s", fragment, body)
		}
	}
}

func TestCompilerRescueIdentifierBindingFromRaisedStringExecutes(t *testing.T) {
	stdout := compileAndRunExecSourceWithOptions(t, "ablec-raise-non-error-rescue-", raiseNonErrorRescueSource(), Options{
		PackageName:        "main",
		EmitMain:           true,
		RequireNoFallbacks: true,
	})
	if strings.TrimSpace(stdout) != "boom" {
		t.Fatalf("expected raised-string rescue handler output boom, got %q", stdout)
	}
}
