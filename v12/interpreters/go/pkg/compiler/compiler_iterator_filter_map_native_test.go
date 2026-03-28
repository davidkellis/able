package compiler

import (
	"strings"
	"testing"
)

func TestCompilerConcreteIteratorFilterMapStayNative(t *testing.T) {
	result := compileNoFallbackExecSource(t, "ablec-linked-list-iterator-filter-map-native", strings.Join([]string{
		"package demo",
		"",
		"import able.collections.linked_list.{LinkedList}",
		"",
		"fn main() -> i64 {",
		"  values: LinkedList i32 = LinkedList.new()",
		"  values.push_back(1)",
		"  values.push_back(2)",
		"  values.push_back(3)",
		"  values.push_back(4)",
		"  values.lazy().filter_map<i64>({ value => if value % 2 == 0 { (value as i64) * 3_i64 } else { nil } }).collect<Array i64>().reduce<i64>(0_i64, { acc, value => acc + value })",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main")
	}
	for _, fragment := range []string{
		"__able_compiled_iface_Iterator_filter_map_default(",
		"__able_compiled_iface_Iterator_collect_default(",
		"__able_compiled_impl_Enumerable_reduce_default_9_spec",
	} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("expected concrete Iterator filter_map pipeline to call compiled helpers directly (%q):\n%s", fragment, body)
		}
	}
	for _, fragment := range []string{
		"__able_method_call_node(",
		"__able_call_value(",
		"__able_iface_Iterator_i32_to_runtime_value(",
		"__able_iface_Iterator_i64_to_runtime_value(",
		"__able_iface_Iterator_i64_from_value(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected concrete Iterator filter_map pipeline to avoid %q:\n%s", fragment, body)
		}
	}

	filterMapBody, ok := findCompiledFunction(result, "__able_compiled_iface_Iterator_filter_map_default")
	if !ok {
		t.Fatalf("could not find compiled Iterator.filter_map helper")
	}
	for _, fragment := range []string{
		"__able_iface_Iterator_A_to_runtime_value(",
		"__able_iface_Iterator_T_to_runtime_value(",
		"__able_call_value(",
		"__able_method_call_node(",
	} {
		if strings.Contains(filterMapBody, fragment) {
			t.Fatalf("expected compiled Iterator.filter_map helper to avoid %q:\n%s", fragment, filterMapBody)
		}
	}
	for _, fragment := range []string{
		".next()",
		"if ",
	} {
		if !strings.Contains(filterMapBody, fragment) {
			t.Fatalf("expected compiled Iterator.filter_map helper to contain %q:\n%s", fragment, filterMapBody)
		}
	}
}

func TestCompilerConcreteIteratorFilterMapExecutes(t *testing.T) {
	source := strings.Join([]string{
		"package main",
		"",
		"import able.collections.linked_list.{LinkedList}",
		"",
		"fn main() -> void {",
		"  values: LinkedList i32 = LinkedList.new()",
		"  values.push_back(1)",
		"  values.push_back(2)",
		"  values.push_back(3)",
		"  values.push_back(4)",
		"  print(values.lazy().filter_map<i64>({ value => if value % 2 == 0 { (value as i64) * 3_i64 } else { nil } }).collect<Array i64>().reduce<i64>(0_i64, { acc, value => acc + value }))",
		"}",
		"",
	}, "\n")

	stdout := compileAndRunExecSourceWithOptions(t, "ablec-linked-list-iterator-filter-map-exec", source, Options{
		PackageName: "main",
		EmitMain:    true,
	})
	if strings.TrimSpace(stdout) != "18" {
		t.Fatalf("expected concrete Iterator filter_map pipeline to print 18, got %q", stdout)
	}
}

func TestCompilerConcreteIteratorFilterMapStayNativeWithExperimentalMonoArrays(t *testing.T) {
	result := compileNoFallbackExecSourceWithOptions(t, "ablec-linked-list-iterator-filter-map-mono-native", strings.Join([]string{
		"package demo",
		"",
		"import able.collections.linked_list.{LinkedList}",
		"",
		"fn main() -> i64 {",
		"  values: LinkedList i32 = LinkedList.new()",
		"  values.push_back(1)",
		"  values.push_back(2)",
		"  values.push_back(3)",
		"  values.push_back(4)",
		"  values.lazy().filter_map<i64>({ value => if value % 2 == 0 { (value as i64) * 3_i64 } else { nil } }).collect<Array i64>().reduce<i64>(0_i64, { acc, value => acc + value })",
		"}",
		"",
	}, "\n"), Options{
		PackageName:            "main",
		ExperimentalMonoArrays: true,
	})

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main")
	}
	for _, fragment := range []string{
		"__able_compiled_iface_Iterator_filter_map_default(",
		"__able_compiled_iface_Iterator_collect_",
		"__able_compiled_impl_Enumerable_reduce_default_9_spec",
	} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("expected concrete Iterator mono-array filter_map pipeline to contain %q:\n%s", fragment, body)
		}
	}
	compiledSrc := string(result.Files["compiled.go"])
	if !strings.Contains(compiledSrc, "func __able_compiled_iface_Iterator_collect_") || !strings.Contains(compiledSrc, "(*__able_array_i64, *__ableControl)") {
		t.Fatalf("expected compiled mono-array collect helper to return the specialized array carrier:\n%s", compiledSrc)
	}
	for _, fragment := range []string{
		"__able_iface_Iterator_i64_to_runtime_value(",
		"__able_method_call_node(",
		"__able_array_i64_from(",
		"__able_call_value(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected concrete Iterator mono-array filter_map pipeline to avoid %q:\n%s", fragment, body)
		}
	}
}

func TestCompilerConcreteIteratorFilterMapExecutesWithExperimentalMonoArrays(t *testing.T) {
	source := strings.Join([]string{
		"package main",
		"",
		"import able.collections.linked_list.{LinkedList}",
		"",
		"fn main() -> void {",
		"  values: LinkedList i32 = LinkedList.new()",
		"  values.push_back(1)",
		"  values.push_back(2)",
		"  values.push_back(3)",
		"  values.push_back(4)",
		"  print(values.lazy().filter_map<i64>({ value => if value % 2 == 0 { (value as i64) * 3_i64 } else { nil } }).collect<Array i64>().reduce<i64>(0_i64, { acc, value => acc + value }))",
		"}",
		"",
	}, "\n")

	stdout := compileAndRunExecSourceWithOptions(t, "ablec-linked-list-iterator-filter-map-mono-exec", source, Options{
		PackageName:            "main",
		EmitMain:               true,
		ExperimentalMonoArrays: true,
	})
	if strings.TrimSpace(stdout) != "18" {
		t.Fatalf("expected concrete Iterator mono-array filter_map pipeline to print 18, got %q", stdout)
	}
}
