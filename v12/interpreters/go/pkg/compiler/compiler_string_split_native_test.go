package compiler

import (
	"strings"
	"testing"
)

func TestCompilerCanonicalStringSplitStaticCallUsesNativeCarriers(t *testing.T) {
	result := compileNoFallbackExecSourceWithOptions(t, "ablec-string-split-native-source", strings.Join([]string{
		"package demo",
		"import able.text.string",
		"",
		"fn value() -> String { \"a|b\" }",
		"fn delimiter() -> String { \"|\" }",
		"",
		"fn main() -> void {",
		"  _ = value().split(delimiter())",
		"}",
		"",
	}, "\n"), Options{
		PackageName:        "main",
		RequireNoFallbacks: true,
	})

	mainBody := extractCompiledFunctionBody(string(result.Files["compiled.go"]), "fn_main")
	for _, fragment := range []string{
		"utf8.ValidString(__able_string_split_value)",
		"utf8.ValidString(__able_string_split_delimiter)",
		"Elements: strings.Split(__able_string_split_value, __able_string_split_delimiter)",
		"__able_compiled_fn_value()",
		"__able_compiled_fn_delimiter()",
		"}(__able_tmp_0, __able_tmp_2)",
	} {
		if !strings.Contains(mainBody, fragment) {
			t.Fatalf("expected static String.split body to contain %q:\n%s", fragment, mainBody)
		}
	}

	splitBody := extractCompiledFunctionBody(string(result.Files["compiled.go"]), "method_String_split")
	if splitBody == "" {
		t.Fatal("canonical String.split body not found")
	}
	if strings.Contains(splitBody, "strings.Split(") {
		t.Fatalf("dynamic canonical String.split body must retain the Able implementation:\n%s", splitBody)
	}
	for _, fragment := range []string{
		"__able_compiled_fn_validated_bytes(",
		"__able_compiled_fn_slice_bytes(",
	} {
		if !strings.Contains(splitBody, fragment) {
			t.Fatalf("dynamic canonical String.split body missing Able fallback %q:\n%s", fragment, splitBody)
		}
	}
}

func TestCompilerCanonicalStringSplitPreservesUnicodeErrorsAndFreshArray(t *testing.T) {
	source := strings.Join([]string{
		"package demo",
		"import able.collections.array",
		"import able.core.interfaces.{Error}",
		"import able.kernel.{Array}",
		"import able.text.string.{String}",
		"",
		"fn main() -> void {",
		"  ascii := \"|a||b|\".split(\"|\")",
		"  print(ascii.size())",
		"  print(ascii[0] + \":\" + ascii[1] + \":\" + ascii[2] + \":\" + ascii[3] + \":\" + ascii[4])",
		"",
		"  unicode := \"é🙂\".split(\"\")",
		"  print(unicode.size())",
		"  print(unicode[0] + \"|\" + unicode[1])",
		"",
		"  multibyte := \"βéββé\".split(\"β\")",
		"  print(multibyte.size())",
		"  print(multibyte[0] + \":\" + multibyte[1] + \":\" + multibyte[2] + \":\" + multibyte[3])",
		"",
		"  first := \"a|b\".split(\"|\")",
		"  second := \"a|b\".split(\"|\")",
		"  first[0] = \"changed\"",
		"  print(first[0] + \":\" + second[0])",
		"",
		"  invalid_bytes: Array u8 = Array.new()",
		"  invalid_bytes.push(255_u8)",
		"  invalid := String.from_bytes_unchecked(invalid_bytes)",
		"  invalid_receiver := do {",
		"    _ = invalid.split(\"|\")",
		"    1",
		"  } rescue {",
		"    case _: Error => 0",
		"  }",
		"  invalid_delimiter := do {",
		"    _ = \"a\".split(invalid)",
		"    1",
		"  } rescue {",
		"    case _: Error => 0",
		"  }",
		"  print(invalid_receiver)",
		"  print(invalid_delimiter)",
		"}",
		"",
	}, "\n")

	stdout := compileAndRunExecSourceWithOptions(t, "ablec-string-split-native-exec", source, Options{
		PackageName:        "main",
		EmitMain:           true,
		RequireNoFallbacks: true,
	})
	const expected = "5\n:a::b:\n2\né|🙂\n4\n:é::é\nchanged:a\n0\n0\n"
	if stdout != expected {
		t.Fatalf("native String.split output = %q, want %q", stdout, expected)
	}
}

func TestCompilerCanonicalStringSplitBoundMethodRetainsDynamicCompatibility(t *testing.T) {
	source := strings.Join([]string{
		"package demo",
		"import able.collections.array",
		"import able.text.string.{String}",
		"",
		"fn main() -> void {",
		"  bound_split := \"x|y\".split",
		"  parts := bound_split(\"|\")",
		"  print(parts[0] + \":\" + parts[1])",
		"}",
		"",
	}, "\n")

	stdout := compileAndRunExecSourceWithOptions(t, "ablec-string-split-bound-method", source, Options{
		PackageName: "main",
		EmitMain:    true,
	})
	if stdout != "x:y\n" {
		t.Fatalf("dynamic String.split output = %q, want %q", stdout, "x:y\n")
	}
}
