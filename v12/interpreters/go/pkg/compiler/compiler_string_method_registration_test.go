package compiler

import (
	"strings"
	"testing"
)

func TestCompilerRegistersCompiledStringFromBytesUncheckedStaticMethod(t *testing.T) {
	_, compiledSrc := compileOutputs(t, "demo", map[string]string{
		"main.able": strings.Join([]string{
			"package demo",
			"",
			"struct Array {",
			"  len_bytes: i32",
			"}",
			"",
			"methods Array {",
			"  fn len(self: Self) -> i32 { self.len_bytes }",
			"}",
			"",
			"struct String {",
			"  bytes: Array,",
			"  len_bytes: i32",
			"}",
			"",
			"methods String {",
			"  fn from_bytes_unchecked(bytes: Array) -> String {",
			"    String { len_bytes: bytes.len(), bytes: bytes }",
			"  }",
			"}",
			"",
			"fn main() -> void {",
			"  _ = String.from_bytes_unchecked(Array { len_bytes: 1 })",
			"}",
			"",
		}, "\n"),
	})

	if !strings.Contains(compiledSrc, "__able_register_compiled_method(\"String\", \"from_bytes_unchecked\", false") {
		t.Fatalf("expected String.from_bytes_unchecked to be registered as a compiled static method")
	}
	if strings.Contains(compiledSrc, "__able_call_named(\"String.from_bytes_unchecked\"") {
		t.Fatalf("expected String.from_bytes_unchecked wrapper to avoid __able_call_named fallback path")
	}
}
