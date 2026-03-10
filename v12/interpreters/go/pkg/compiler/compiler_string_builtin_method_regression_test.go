package compiler

import (
	"strings"
	"testing"
)

// Regression test: String methods that access builtin-mapped fields (len_bytes)
// and static methods returning String must remain compilable after the any
// TypeMapper migration. See: stringBuiltinFieldAccess, isBuiltinMappedType.

func TestCompilerStringIsEmptyMethodCompilesWithBuiltinFieldAccess(t *testing.T) {
	_, compiledSrc := compileOutputs(t, "demo", map[string]string{
		"main.able": strings.Join([]string{
			"package demo",
			"",
			"struct String {",
			"  bytes: Array u8,",
			"  len_bytes: u64",
			"}",
			"",
			"methods String {",
			"  fn is_empty(self: Self) -> bool { self.len_bytes == 0 }",
			"}",
			"",
			"fn main() -> void {",
			`  _ = "hello".is_empty()`,
			"}",
			"",
		}, "\n"),
	})

	if !strings.Contains(compiledSrc, "__able_register_compiled_method(\"String\", \"is_empty\", true") {
		t.Fatalf("expected String.is_empty to be registered as a compiled instance method;\ncompiled output:\n%s", compiledSrc)
	}
}

func TestCompilerStringStaticMethodReturnsStringNotStructPointer(t *testing.T) {
	_, compiledSrc := compileOutputs(t, "demo", map[string]string{
		"main.able": strings.Join([]string{
			"package demo",
			"",
			"struct String {",
			"  bytes: Array u8,",
			"  len_bytes: u64",
			"}",
			"",
			"methods String {",
			"  fn concat(a: String, b: String) -> String { a }",
			"}",
			"",
			"fn main() -> void {",
			`  _ = String.concat("hello", "world")`,
			"}",
			"",
		}, "\n"),
	})

	// The method should be registered as compiled — isBuiltinMappedType prevents
	// staticMethodNominalStructReturnType from forcing return to *String.
	if !strings.Contains(compiledSrc, "__able_register_compiled_method(\"String\", \"concat\", false") {
		t.Fatalf("expected String.concat to be registered as a compiled static method;\ncompiled output:\n%s", compiledSrc)
	}
}
