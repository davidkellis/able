package compiler

import (
	"strings"
	"testing"
)

// TestCompilerStringFromBytesUncheckedFallsBackForBuiltinMapping verifies
// that String.from_bytes_unchecked correctly falls back to the interpreter
// when String maps to the builtin Go string type, since struct-literal
// construction is incompatible with the builtin mapping.
func TestCompilerStringFromBytesUncheckedFallsBackForBuiltinMapping(t *testing.T) {
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

	// String maps to Go string (builtin), so struct-literal construction
	// cannot be compiled — the method should NOT be registered as compiled.
	if strings.Contains(compiledSrc, "__able_register_compiled_method(\"String\", \"from_bytes_unchecked\", false") {
		t.Fatalf("String.from_bytes_unchecked should not be compiled when String maps to builtin string")
	}
}
