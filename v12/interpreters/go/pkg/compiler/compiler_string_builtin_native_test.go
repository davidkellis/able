package compiler

import (
	"strings"
	"testing"
)

func TestCompilerStaticStringBuiltinHelpersKeepNativeCarriers(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn computed() -> String { \"h\" + \"i\" }",
		"",
		"fn main() -> String {",
		"  bytes := __able_String_from_builtin(computed())",
		"  __able_String_to_builtin(bytes)",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("main body not found")
	}
	for _, fragment := range []string{
		"__able_string_from_builtin_native(",
		"__able_string_to_builtin_native(",
	} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("expected static String/Array u8 conversion to use %q:\n%s", fragment, body)
		}
	}
	for _, fragment := range []string{
		"__able_string_from_builtin_impl(",
		"__able_string_to_builtin_impl(",
		"__able_array_u8_from(",
		"__able_array_u8_to(",
		"bridge.ToString(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected static String/Array u8 conversion to avoid %q:\n%s", fragment, body)
		}
	}
}

func TestCompilerErasedStringBuiltinHelperKeepsRuntimeCompatibilityPath(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn convert(value: any) -> any {",
		"  __able_String_from_builtin(value)",
		"}",
		"",
		"fn restore(value: any) -> any {",
		"  __able_String_to_builtin(value)",
		"}",
		"",
		"fn main() -> void {",
		"  bytes := convert(\"hi\")",
		"  _ = restore(bytes)",
		"}",
		"",
	}, "\n"))

	for _, check := range []struct {
		function string
		runtime  string
		native   string
	}{
		{"__able_compiled_fn_convert", "__able_string_from_builtin_impl(", "__able_string_from_builtin_native("},
		{"__able_compiled_fn_restore", "__able_string_to_builtin_impl(", "__able_string_to_builtin_native("},
	} {
		body, ok := findCompiledFunction(result, check.function)
		if !ok {
			t.Fatalf("%s body not found", check.function)
		}
		if !strings.Contains(body, check.runtime) {
			t.Fatalf("expected erased conversion to retain %q:\n%s", check.runtime, body)
		}
		if strings.Contains(body, check.native) {
			t.Fatalf("expected erased conversion not to assume %q:\n%s", check.native, body)
		}
	}
}

func TestCompilerStaticStringToBuiltinPreservesUTF8Validation(t *testing.T) {
	source := strings.Join([]string{
		"package demo",
		"import able.collections.array",
		"import able.core.interfaces.{Error}",
		"import able.kernel.{Array}",
		"",
		"fn main() -> void {",
		"  invalid: Array u8 = Array.new()",
		"  invalid.push(255_u8)",
		"  result := do {",
		"    _ = __able_String_to_builtin(invalid)",
		"    1",
		"  } rescue {",
		"    case _: Error => 0",
		"  }",
		"  print(result)",
		"}",
		"",
	}, "\n")

	stdout := compileAndRunExecSourceWithOptions(t, "ablec-string-to-builtin-native-validation", source, Options{
		PackageName:        "main",
		EmitMain:           true,
		RequireNoFallbacks: true,
	})
	if got := strings.TrimSpace(stdout); got != "0" {
		t.Fatalf("expected invalid native byte array to preserve UTF-8 error, got %q", stdout)
	}
}
