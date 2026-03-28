package compiler

import (
	"strings"
	"testing"
)

func TestCompilerConcreteInterfaceImplSpecificityHonorsConstraints(t *testing.T) {
	source := strings.Join([]string{
		"package demo",
		"",
		"interface Label for T {",
		"  fn label(self: Self) -> String",
		"}",
		"",
		"interface Alpha for T {",
		"  fn alpha(self: Self) -> String",
		"}",
		"",
		"interface Beta for T {",
		"  fn beta(self: Self) -> String",
		"}",
		"",
		"struct Thing {}",
		"struct Solo {}",
		"",
		"impl Alpha for Thing {",
		"  fn alpha(self: Self) -> String { \"a\" }",
		"}",
		"",
		"impl Beta for Thing {",
		"  fn beta(self: Self) -> String { \"b\" }",
		"}",
		"",
		"impl Alpha for Solo {",
		"  fn alpha(self: Self) -> String { \"a\" }",
		"}",
		"",
		"impl<T: Alpha> Label for T {",
		"  fn label(self: Self) -> String { \"alpha\" }",
		"}",
		"",
		"impl<T: Alpha + Beta> Label for T {",
		"  fn label(self: Self) -> String { \"alpha+beta\" }",
		"}",
		"",
		"impl Label for i32 | f32 {",
		"  fn label(self: Self) -> String { \"narrow\" }",
		"}",
		"",
		"impl Label for i32 | f32 | f64 {",
		"  fn label(self: Self) -> String { \"wide\" }",
		"}",
		"",
		"fn main() -> void {",
		"  thing := Thing {}",
		"  solo := Solo {}",
		"  value_i32 := 3",
		"  value_f64 := 3.5",
		"  print(thing.label())",
		"  print(solo.label())",
		"  print(value_i32.label())",
		"  print(value_f64.label())",
		"}",
		"",
	}, "\n")

	result := compileNoFallbackExecSource(t, "ablec-impl-specificity-native", source)
	compiledSrc := string(result.Files["compiled.go"])
	for _, fragment := range []string{
		"func __able_compiled_impl_Label_label_",
		"(self __able_union_float32_or_int32)",
		"(self __able_union_float32_or_float64_or_int32)",
	} {
		if !strings.Contains(compiledSrc, fragment) {
			t.Fatalf("expected compiled Label impl source to contain %q:\n%s", fragment, compiledSrc)
		}
	}
	mainBody, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main")
	}
	for _, fragment := range []string{
		"__able_compiled_impl_Label_label_",
		"thing)",
		"solo)",
		"__able_union_float32_or_int32_wrap_int32(value_i32)",
		"__able_union_float32_or_float64_or_int32_wrap_float64(value_f64)",
	} {
		if !strings.Contains(mainBody, fragment) {
			t.Fatalf("expected impl-specific compiled dispatch to contain %q:\n%s", fragment, mainBody)
		}
	}
	for _, fragment := range []string{
		"__able_method_call_node(",
		"__able_member_get_method(",
		"__able_iface_",
	} {
		if strings.Contains(mainBody, fragment) {
			t.Fatalf("expected static impl-specific dispatch to avoid %q:\n%s", fragment, mainBody)
		}
	}
}
