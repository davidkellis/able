package compiler

import (
	"strings"
	"testing"
)

func TestCompilerStaticNativePathsAvoidDynamicHelperReachability(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"struct Point { x: i32, y: i32 }",
		"",
		"methods Point {",
		"  fn #sum() -> i32 { #x + #y }",
		"}",
		"",
		"fn main() -> i32 {",
		"  arr := [1, 2, 3]",
		"  arr[0] = 5",
		"  point := Point { x: arr[0]! as i32, y: 4 }",
		"  point.x + point.sum()",
		"}",
		"",
	}, "\n"))

	mainBody, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	for _, fragment := range []string{
		"__able_index(",
		"__able_index_set(",
		"__able_member_get(",
		"__able_member_set(",
		"__able_member_get_method(",
		"__able_call_value(",
	} {
		if strings.Contains(mainBody, fragment) {
			t.Fatalf("static native path should avoid %q:\n%s", fragment, mainBody)
		}
	}
}

func TestCompilerExplicitDynamicPathsUseDynamicHelpers(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"dynimport exec.dynamic_helper_reach::dyn_pkg",
		"pkg := dyn.def_package(\"exec.dynamic_helper_reach\")!",
		"pkg.def(\"fn make() { [4, 5, 6] }\")!",
		"",
		"fn main() -> i32 {",
		"  maker := dyn_pkg.make",
		"  values := maker()",
		"  values.length = 2",
		"  values[1]! as i32",
		"}",
		"",
	}, "\n"))

	mainBody, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	for _, fragment := range []string{
		"__able_member_get(",
		"__able_member_set(",
		"__able_call_value(",
		"__able_index(",
	} {
		if !strings.Contains(mainBody, fragment) {
			t.Fatalf("explicit dynamic path should retain %q:\n%s", fragment, mainBody)
		}
	}
}

func TestCompilerDynamicHelpersRemovePanicBridgeWrappers(t *testing.T) {
	_, compiledSrc := compileOutputs(t, "demo", map[string]string{
		"main.able": strings.Join([]string{
			"package demo",
			"",
			"fn main() -> void {",
			"  _ = 1",
			"}",
			"",
		}, "\n"),
	})

	for _, fragment := range []string{
		"func __able_error_from_panic(",
		"func __able_bridge_call_value_with_node(",
		"func __able_bridge_call_named_with_node(",
	} {
		if strings.Contains(compiledSrc, fragment) {
			t.Fatalf("expected generated runtime to remove %q", fragment)
		}
	}
	for _, fragment := range []string{
		"func __able_try_member_set(",
		"func __able_try_member_get(",
		"func __able_try_member_get_method(",
	} {
		if !strings.Contains(compiledSrc, fragment) {
			t.Fatalf("expected generated runtime to emit %q", fragment)
		}
	}
}
