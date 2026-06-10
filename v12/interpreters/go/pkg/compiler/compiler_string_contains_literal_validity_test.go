package compiler

import (
	"strings"
	"testing"
)

func TestCompilerStringContainsLiteralPreservesInvalidReceiverError(t *testing.T) {
	source := strings.Join([]string{
		"package demo",
		"import able.collections.array",
		"import able.core.interfaces.{Error}",
		"import able.kernel.{Array}",
		"import able.text.string.{String}",
		"",
		"fn contains_literal(value: String) -> i32 {",
		"  do {",
		"    value.contains(\"x\")",
		"    1",
		"  } rescue {",
		"    case _: Error => 0",
		"  }",
		"}",
		"",
		"fn main() -> void {",
		"  print(contains_literal(\"text\"))",
		"  invalid: Array u8 = Array.new()",
		"  invalid.push(255_u8)",
		"  invalid_string := String.from_bytes_unchecked(invalid)",
		"  print(contains_literal(invalid_string))",
		"}",
		"",
	}, "\n")

	stdout := compileAndRunExecSourceWithOptions(t, "ablec-string-contains-literal-validity", source, Options{
		PackageName:        "main",
		EmitMain:           true,
		RequireNoFallbacks: true,
	})
	if got := strings.TrimSpace(stdout); got != "1\n0" {
		t.Fatalf("expected valid receiver success and invalid receiver error, got %q", stdout)
	}
}

func TestCompilerStringLenBytesCallPreservesInvalidReceiverError(t *testing.T) {
	source := strings.Join([]string{
		"package demo",
		"import able.collections.array",
		"import able.core.interfaces.{Error}",
		"import able.kernel.{Array}",
		"import able.text.string.{String}",
		"",
		"fn checked_len(value: String) -> u64 {",
		"  do {",
		"    value.len_bytes()",
		"  } rescue {",
		"    case _: Error => 0_u64",
		"  }",
		"}",
		"",
		"fn main() -> void {",
		"  print(checked_len(\"text\"))",
		"  invalid: Array u8 = Array.new()",
		"  invalid.push(255_u8)",
		"  invalid_string := String.from_bytes_unchecked(invalid)",
		"  print(checked_len(invalid_string))",
		"}",
		"",
	}, "\n")

	stdout := compileAndRunExecSourceWithOptions(t, "ablec-string-len-bytes-call-validity", source, Options{
		PackageName:        "main",
		EmitMain:           true,
		RequireNoFallbacks: true,
	})
	if got := strings.TrimSpace(stdout); got != "4\n0" {
		t.Fatalf("expected native valid length and invalid receiver error, got %q", stdout)
	}
}
