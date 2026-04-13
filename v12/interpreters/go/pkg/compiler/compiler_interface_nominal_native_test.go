package compiler

import (
	"strings"
	"testing"
)

func TestCompilerUnrelatedSameShapeInterfaceJoinStaysNominalUnion(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"interface Reader <T> for Self {",
		"  fn read(self: Self) -> T",
		"}",
		"",
		"interface Source <T> for Self {",
		"  fn read(self: Self) -> T",
		"}",
		"",
		"struct First {}",
		"struct Second {}",
		"",
		"impl Reader i32 for First {",
		"  fn read(self: Self) -> i32 { 1 }",
		"}",
		"",
		"impl Source i32 for Second {",
		"  fn read(self: Self) -> i32 { 2 }",
		"}",
		"",
		"fn main() -> i32 {",
		"  left: Reader i32 = First {}",
		"  right: Source i32 = Second {}",
		"  mixed := if true {",
		"    left",
		"  } else {",
		"    right",
		"  }",
		"  mixed match {",
		"    case reader: Reader i32 => reader.read(),",
		"    case source: Source i32 => source.read()",
		"  }",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	if !strings.Contains(body, "var mixed __able_union_") {
		t.Fatalf("expected unrelated same-shape interface join to stay a native union, not collapse to one interface carrier:\n%s", body)
	}
	for _, fragment := range []string{
		"var mixed runtime.Value",
		"var mixed any",
		"__able_try_cast(",
		"bridge.MatchType(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected unrelated same-shape interface join to avoid %q:\n%s", fragment, body)
		}
	}
	for _, fragment := range []string{
		"var reader __able_iface_Reader_i32",
		"var source __able_iface_Source_i32",
		"reader.read()",
		"source.read()",
	} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("expected unrelated same-shape interface join to keep distinct native interface branches (%q):\n%s", fragment, body)
		}
	}
}
