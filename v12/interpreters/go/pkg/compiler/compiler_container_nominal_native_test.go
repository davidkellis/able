package compiler

import (
	"regexp"
	"strings"
	"testing"
)

func TestCompilerStaticSliceBuiltinsIgnoreLenShadowing(t *testing.T) {
	source := strings.Join([]string{
		"package demo",
		"",
		"import able.kernel.{Array}",
		"extern go fn __able_os_exit(code: i32) -> void {}",
		"",
		"fn main() {",
		"  len := 1",
		"  values: Array i32 = Array.with_capacity(2)",
		"  values.push(4)",
		"  values.push(9)",
		"  score := len + values.len() + values.capacity() + values.get(len)!",
		"  if score == 14 {",
		"    __able_os_exit(0)",
		"  } else {",
		"    __able_os_exit(1)",
		"  }",
		"}",
		"",
	}, "\n")

	result := compileNoFallbackExecSource(t, "ablec-builtins-len-shadow", source)
	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	for _, fragment := range []string{
		"__able_slice_len(",
		"__able_slice_cap(",
	} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("expected len/cap builtin shadowing path to use %q:\n%s", fragment, body)
		}
	}
	for _, pattern := range []*regexp.Regexp{
		regexp.MustCompile(`(^|[^[:alnum:]_])len\(`),
		regexp.MustCompile(`(^|[^[:alnum:]_])cap\(`),
	} {
		if pattern.FindStringIndex(body) != nil {
			t.Fatalf("expected compiled body to avoid bare Go builtin calls matched by %q:\n%s", pattern.String(), body)
		}
	}

	compileAndRunExecSourceWithOptions(t, "ablec-builtins-len-shadow-run", source, Options{
		PackageName:        "main",
		EmitMain:           true,
		RequireNoFallbacks: true,
	})
}

func TestCompilerTreeMapStaticCarrierStaysNative(t *testing.T) {
	result := compileNoFallbackExecSource(t, "ablec-treemap-native-static", strings.Join([]string{
		"package demo",
		"",
		"import able.collections.tree_map.*",
		"",
		"fn build() -> TreeMap i32 String {",
		"  values := TreeMap.new()",
		"  values.set(1, \"a\")",
		"  values.set(2, \"b\")",
		"  values",
		"}",
		"",
		"fn size_of(values: TreeMap i32 String) -> i32 {",
		"  values.len()",
		"}",
		"",
	}, "\n"))

	compiledSrc := string(result.Files["compiled.go"])
	if !strings.Contains(compiledSrc, "func __able_compiled_fn_build() (*TreeMap, *__ableControl)") {
		t.Fatalf("expected TreeMap return to stay on the native carrier:\n%s", compiledSrc)
	}
	if !strings.Contains(compiledSrc, "func __able_compiled_fn_size_of(values *TreeMap) (int32, *__ableControl)") {
		t.Fatalf("expected TreeMap param to stay on the native carrier:\n%s", compiledSrc)
	}

	buildBody, ok := findCompiledFunction(result, "__able_compiled_fn_build")
	if !ok {
		t.Fatalf("could not find compiled build function")
	}
	for _, fragment := range []string{
		"var values *TreeMap =",
		"__able_compiled_method_TreeMap_new(",
		"__able_compiled_method_TreeMap_set(values",
	} {
		if !strings.Contains(buildBody, fragment) {
			t.Fatalf("expected native TreeMap build lowering to contain %q:\n%s", fragment, buildBody)
		}
	}
	for _, fragment := range []string{
		"runtime.Value",
		"__able_call_value(",
		"__able_any_to_value(",
	} {
		if strings.Contains(buildBody, fragment) {
			t.Fatalf("expected native TreeMap build lowering to avoid %q:\n%s", fragment, buildBody)
		}
	}
}

func TestCompilerPersistentMapStaticCarrierStaysNative(t *testing.T) {
	result := compileNoFallbackExecSource(t, "ablec-persistentmap-native-static", strings.Join([]string{
		"package demo",
		"",
		"import able.collections.persistent_map.*",
		"",
		"fn build() -> PersistentMap i32 String {",
		"  values := PersistentMap.empty()",
		"  values = values.set(1, \"a\")",
		"  values = values.set(2, \"b\")",
		"  values",
		"}",
		"",
		"fn size_of(values: PersistentMap i32 String) -> i32 {",
		"  values.len()",
		"}",
		"",
	}, "\n"))

	compiledSrc := string(result.Files["compiled.go"])
	if !strings.Contains(compiledSrc, "func __able_compiled_fn_build() (*PersistentMap, *__ableControl)") {
		t.Fatalf("expected PersistentMap return to stay on the native carrier:\n%s", compiledSrc)
	}
	if !strings.Contains(compiledSrc, "func __able_compiled_fn_size_of(values *PersistentMap) (int32, *__ableControl)") {
		t.Fatalf("expected PersistentMap param to stay on the native carrier:\n%s", compiledSrc)
	}

	buildBody, ok := findCompiledFunction(result, "__able_compiled_fn_build")
	if !ok {
		t.Fatalf("could not find compiled build function")
	}
	for _, fragment := range []string{
		"var values *PersistentMap =",
		"__able_compiled_method_PersistentMap_empty(",
		"__able_compiled_method_PersistentMap_set(values",
		"values = __able_tmp_",
	} {
		if !strings.Contains(buildBody, fragment) {
			t.Fatalf("expected native PersistentMap build lowering to contain %q:\n%s", fragment, buildBody)
		}
	}
	for _, fragment := range []string{
		"runtime.Value",
		"__able_call_value(",
		"__able_any_to_value(",
	} {
		if strings.Contains(buildBody, fragment) {
			t.Fatalf("expected native PersistentMap build lowering to avoid %q:\n%s", fragment, buildBody)
		}
	}
}
