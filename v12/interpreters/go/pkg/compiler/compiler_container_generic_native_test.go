package compiler

import (
	"strings"
	"testing"
)

func calledFunctionNameFromBody(body string, prefix string) (string, bool) {
	idx := strings.Index(body, prefix)
	if idx < 0 {
		return "", false
	}
	end := strings.Index(body[idx:], "(")
	if end < 0 {
		return "", false
	}
	return body[idx : idx+end], true
}

func TestCompilerLazySeqIteratorCarrierStaysNative(t *testing.T) {
	result := compileNoFallbackExecSource(t, "ablec-lazyseq-iterator-native", strings.Join([]string{
		"package demo",
		"",
		"import able.collections.linked_list.*",
		"import able.collections.lazy_seq.*",
		"import able.core.iteration.{Iterable}",
		"",
		"fn build(src: Iterable i32) -> LazySeq i32 {",
		"  LazySeq.from_iterable(src)",
		"}",
		"",
		"fn main() -> i32 {",
		"  values: LinkedList i32 = LinkedList.new()",
		"  values.push_back(1)",
		"  values.push_back(2)",
		"  seq := build(values)",
		"  seq.pull_next() match {",
		"    case nil => 0,",
		"    case value: i32 => value",
		"  }",
		"}",
		"",
	}, "\n"))
	compiledSrc := string(result.Files["compiled.go"])

	if strings.Contains(compiledSrc, "type LazySeq struct {\n\tSource    any") {
		t.Fatalf("expected LazySeq.Source to stay on a native iterator carrier, not any:\n%s", compiledSrc)
	}
	if !strings.Contains(compiledSrc, "type LazySeq struct {\n\tSource    __able_iface_Iterator_T") {
		t.Fatalf("expected LazySeq.Source to use the native iterator carrier:\n%s", compiledSrc)
	}

	clearBody, ok := findCompiledFunction(result, "__able_compiled_method_LinkedList_clear")
	if !ok {
		t.Fatalf("could not find compiled LinkedList.clear")
	}
	for _, fragment := range []string{"(*ListNode)(nil)"} {
		if !strings.Contains(clearBody, fragment) {
			t.Fatalf("expected LinkedList.clear to use typed nil for native nullable node fields (%q):\n%s", fragment, clearBody)
		}
	}
	if strings.Contains(clearBody, ":= nil") {
		t.Fatalf("expected LinkedList.clear to avoid untyped nil temps:\n%s", clearBody)
	}

	if !strings.Contains(compiledSrc, "__able_iface_Iterator_T(nil)") {
		t.Fatalf("expected LazySeq native iterator paths to use typed nil for source resets:\n%s", compiledSrc)
	}
	if strings.Contains(compiledSrc, "Source = nil") {
		t.Fatalf("expected LazySeq native iterator paths to avoid raw nil source resets:\n%s", compiledSrc)
	}
}

func TestCompilerLinkedListIterableAdapterStaysNative(t *testing.T) {
	result := compileNoFallbackExecSource(t, "ablec-linked-list-iterable-adapter-native", strings.Join([]string{
		"package demo",
		"",
		"import able.collections.linked_list.{LinkedList}",
		"import able.core.iteration.{Iterable}",
		"",
		"fn checksum(values: Iterable i32) -> i64 {",
		"  total: i64 = 0_i64",
		"  for value in values {",
		"    total = total + (value as i64)",
		"  }",
		"  total",
		"}",
		"",
		"fn main() -> i64 {",
		"  values: LinkedList i32 = LinkedList.new()",
		"  values.push_back(1)",
		"  values.push_back(2)",
		"  checksum(values)",
		"}",
		"",
	}, "\n"))
	compiledSrc := string(result.Files["compiled.go"])

	mainBody, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main")
	}
	if !strings.Contains(mainBody, "__able_iface_Iterable_i32_wrap_ptr_LinkedList") {
		t.Fatalf("expected LinkedList arg to wrap directly into the native Iterable carrier:\n%s", mainBody)
	}
	for _, fragment := range []string{
		"__able_any_to_value(values)",
		"__able_iface_Iterable_i32_from_value(",
	} {
		if strings.Contains(mainBody, fragment) {
			t.Fatalf("expected LinkedList Iterable arg to avoid %q:\n%s", fragment, mainBody)
		}
	}

	start := strings.Index(compiledSrc, "func (w __able_iface_Iterable_i32_adapter_ptr_LinkedList_i32) iterator()")
	if start < 0 {
		start = strings.Index(compiledSrc, "func (w __able_iface_Iterable_i32_adapter_ptr_LinkedList) iterator()")
	}
	if start < 0 {
		t.Fatalf("could not find LinkedList Iterable adapter iterator body")
	}
	end := strings.Index(compiledSrc[start:], "\n}\n\n")
	if end < 0 {
		t.Fatalf("could not slice LinkedList Iterable adapter iterator body")
	}
	adapterBody := compiledSrc[start : start+end]
	for _, fragment := range []string{
		"__able_iface_Iterator_A_to_runtime_value(",
		"__able_iface_Iterator_i32_from_value(",
	} {
		if strings.Contains(adapterBody, fragment) {
			t.Fatalf("expected LinkedList Iterable adapter iterator to avoid %q:\n%s", fragment, adapterBody)
		}
	}
	if !strings.Contains(adapterBody, "__able_compiled_impl_Enumerable_iterator_1_spec(") &&
		!strings.Contains(adapterBody, "__able_iface_Iterator_i32_wrap_ptr_LinkedListIterator(") {
		t.Fatalf("expected LinkedList Iterable adapter iterator to stay on the native compiled path:\n%s", adapterBody)
	}
}

func TestCompilerConcreteEnumerableGenericMethodsStayNative(t *testing.T) {
	result := compileNoFallbackExecSource(t, "ablec-linked-list-enumerable-generic-native", strings.Join([]string{
		"package demo",
		"",
		"import able.collections.linked_list.{LinkedList}",
		"",
		"fn main() -> i64 {",
		"  values: LinkedList i32 = LinkedList.new()",
		"  values.push_back(1)",
		"  values.push_back(2)",
		"  values.push_back(3)",
		"  mapped := values.map<i64>({ value => (value as i64) * 3_i64 })",
		"  filtered := mapped.filter({ value => value >= 6_i64 })",
		"  filtered.reduce<i64>(0_i64, { acc, value => acc + value })",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main")
	}
	for _, fragment := range []string{
		"__able_iface_Enumerable_",
		"__able_iface_Iterator_",
		"__able_method_call_node(",
		"__able_call_value(",
		"bridge.MatchType(",
		"__able_fn_runtime_Value_to_bool(",
		"__able_fn_runtime_Value_runtime_Value_to_runtime_Value(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected concrete Enumerable generic methods to stay on compiled native calls without %q:\n%s", fragment, body)
		}
	}
	for _, fragment := range []string{
		"__able_compiled_impl_Enumerable__map_default_",
		"__able_compiled_impl_Enumerable_filter_default_",
		"__able_compiled_impl_Enumerable_reduce_default_",
	} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("expected concrete Enumerable generic methods to call compiled impls directly (%q):\n%s", fragment, body)
		}
	}
	for _, fragment := range []string{
		"var mapped *LinkedList_i64 = ",
		"var filtered *LinkedList_i64 = ",
	} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("expected concrete Enumerable generic methods to keep native LinkedList carriers (%q):\n%s", fragment, body)
		}
	}
	mapName, ok := calledFunctionNameFromBody(body, "__able_compiled_impl_Enumerable__map_default_")
	if !ok {
		t.Fatalf("could not find concrete Enumerable.map helper call in main body")
	}
	mapBody, ok := findCompiledFunction(result, mapName)
	if !ok {
		t.Fatalf("could not find specialized compiled Enumerable.map default impl body")
	}
	filterName, ok := calledFunctionNameFromBody(body, "__able_compiled_impl_Enumerable_filter_default_")
	if !ok {
		t.Fatalf("could not find concrete Enumerable.filter helper call in main body")
	}
	filterBody, ok := findCompiledFunction(result, filterName)
	if !ok {
		t.Fatalf("could not find specialized compiled Enumerable.filter default impl body")
	}
	reduceName, ok := calledFunctionNameFromBody(body, "__able_compiled_impl_Enumerable_reduce_default_")
	if !ok {
		t.Fatalf("could not find concrete Enumerable.reduce helper call in main body")
	}
	reduceBody, ok := findCompiledFunction(result, reduceName)
	if !ok {
		t.Fatalf("could not find specialized compiled Enumerable.reduce default impl body")
	}
	lazyName, ok := calledFunctionNameFromBody(mapBody, "__able_compiled_impl_Enumerable_lazy_default_")
	if !ok {
		t.Fatalf("could not find concrete Enumerable.lazy helper call in map helper body")
	}
	lazyBody, ok := findCompiledFunction(result, lazyName)
	if !ok {
		t.Fatalf("could not find specialized compiled Enumerable.lazy default impl body")
	}
	for name, fnBody := range map[string]string{
		"map":    mapBody,
		"filter": filterBody,
		"reduce": reduceBody,
	} {
		for _, fragment := range []string{
			"__able_iface_Iterator_A_to_runtime_value(",
			"__able_iface_Iterator_T_from_value(",
			"__able_compiled_impl_Iterable_iterator_0(",
			"__able_fn_runtime_Value_to_runtime_Value",
			"__able_fn_runtime_Value_to_bool",
			"__able_fn_runtime_Value_runtime_Value_to_runtime_Value",
		} {
			if strings.Contains(fnBody, fragment) {
				t.Fatalf("expected concrete Enumerable.%s default impl loop to avoid %q:\n%s", name, fragment, fnBody)
			}
		}
		if !strings.Contains(fnBody, ".next()") {
			t.Fatalf("expected concrete Enumerable.%s default impl loop to iterate native iterator carriers directly:\n%s", name, fnBody)
		}
	}
	for _, fragment := range []string{
		"__able_iface_Iterator_A_to_runtime_value(",
		"__able_iface_Iterator_i32_from_value(",
	} {
		if strings.Contains(lazyBody, fragment) {
			t.Fatalf("expected concrete Enumerable.lazy default impl to avoid %q:\n%s", fragment, lazyBody)
		}
	}
	if !strings.Contains(lazyBody, "__able_compiled_impl_Enumerable_iterator_1_spec") {
		t.Fatalf("expected concrete Enumerable.lazy default impl to call the specialized iterator impl directly:\n%s", lazyBody)
	}
}

func TestCompilerConcreteEnumerableGenericMethodsExecute(t *testing.T) {
	output := compileAndRunExecSourceWithOptions(t, "ablec-linked-list-enumerable-generic-exec", strings.Join([]string{
		"package main",
		"",
		"import able.collections.linked_list.{LinkedList}",
		"",
		"fn score() -> i64 {",
		"  values: LinkedList i32 = LinkedList.new()",
		"  values.push_back(1)",
		"  values.push_back(2)",
		"  values.push_back(3)",
		"  mapped := values.map<i64>({ value => (value as i64) * 3_i64 })",
		"  filtered := mapped.filter({ value => value >= 6_i64 })",
		"  filtered.reduce<i64>(0_i64, { acc, value => acc + value })",
		"}",
		"",
		"fn main() -> void {",
		"  print(score())",
		"}",
		"",
	}, "\n"), Options{
		PackageName:        "main",
		EmitMain:           true,
		RequireNoFallbacks: true,
	})
	if strings.TrimSpace(output) != "15" {
		t.Fatalf("expected executable concrete Enumerable generic methods to return 15, got %q", output)
	}
}

func TestCompilerConcreteIteratorGenericMethodsStayNative(t *testing.T) {
	result := compileNoFallbackExecSource(t, "ablec-linked-list-iterator-generic-native", strings.Join([]string{
		"package demo",
		"",
		"import able.collections.linked_list.{LinkedList}",
		"",
		"fn main() -> i64 {",
		"  values: LinkedList i32 = LinkedList.new()",
		"  values.push_back(1)",
		"  values.push_back(2)",
		"  values.push_back(3)",
		"  iter := values.lazy().map<i64>({ value => (value as i64) * 3_i64 }).filter({ value => value >= 6_i64 })",
		"  iter.collect<Array i64>().reduce<i64>(0_i64, { acc, value => acc + value })",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main")
	}
	for _, fragment := range []string{
		"__able_method_call_node(",
		"__able_iface_Iterator_i32_to_runtime_value(",
		"__able_iface_Iterator_i64_to_runtime_value(",
		"__able_iface_Iterator_i64_from_value(",
		"__able_call_value(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected concrete Iterator generic methods to stay on compiled native calls without %q:\n%s", fragment, body)
		}
	}
	for _, fragment := range []string{
		"__able_compiled_iface_Iterator__map_default(",
		"__able_compiled_iface_Iterator_filter_default(",
		"__able_compiled_iface_Iterator_collect_default(",
		"__able_compiled_impl_Enumerable_reduce_default_9_spec(",
	} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("expected concrete Iterator pipeline to call compiled helpers directly (%q):\n%s", fragment, body)
		}
	}

	for _, fnName := range []string{
		"__able_compiled_iface_Iterator__map_default",
		"__able_compiled_iface_Iterator_filter_default",
		"__able_compiled_iface_Iterator_collect_default",
	} {
		fnBody, ok := findCompiledFunction(result, fnName)
		if !ok {
			t.Fatalf("could not find compiled Iterator helper %s", fnName)
		}
		for _, fragment := range []string{
			"__able_iface_Iterator_A_to_runtime_value(",
			"__able_iface_Iterator_T_to_runtime_value(",
			"__able_call_value(",
		} {
			if strings.Contains(fnBody, fragment) {
				t.Fatalf("expected %s to avoid %q:\n%s", fnName, fragment, fnBody)
			}
		}
	}
}

func TestCompilerConcreteIteratorGenericMethodsExecute(t *testing.T) {
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
		"  print(values.lazy().map<i64>({ value => (value as i64) * 3_i64 }).filter({ value => value >= 6_i64 }).collect<Array i64>().reduce<i64>(0_i64, { acc, value => acc + value }))",
		"}",
		"",
	}, "\n")

	stdout := compileAndRunExecSourceWithOptions(t, "ablec-linked-list-iterator-generic-exec", source, Options{
		PackageName: "main",
		EmitMain:    true,
	})
	if strings.TrimSpace(stdout) != "15" {
		t.Fatalf("expected concrete Iterator generic pipeline to print 15, got %q", stdout)
	}
}

func TestCompilerConcreteIteratorGenericMethodsStayNativeWithExperimentalMonoArrays(t *testing.T) {
	result := compileNoFallbackExecSourceWithOptions(t, "ablec-linked-list-iterator-generic-mono-native", strings.Join([]string{
		"package demo",
		"",
		"import able.collections.linked_list.{LinkedList}",
		"",
		"fn main() -> i64 {",
		"  values: LinkedList i32 = LinkedList.new()",
		"  values.push_back(1)",
		"  values.push_back(2)",
		"  values.push_back(3)",
		"  iter := values.lazy().map<i64>({ value => (value as i64) * 3_i64 }).filter({ value => value >= 6_i64 })",
		"  iter.collect<Array i64>().reduce<i64>(0_i64, { acc, value => acc + value })",
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
		"__able_compiled_iface_Iterator__map_default(",
		"__able_compiled_iface_Iterator_filter_default(",
		"__able_compiled_iface_Iterator_collect_",
		"__able_compiled_impl_Enumerable_reduce_default_9_spec",
	} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("expected concrete Iterator mono-array pipeline to contain %q:\n%s", fragment, body)
		}
	}
	compiledSrc := string(result.Files["compiled.go"])
	if !strings.Contains(compiledSrc, "func __able_compiled_iface_Iterator_collect_") || !strings.Contains(compiledSrc, "(*__able_array_i64, *__ableControl)") {
		t.Fatalf("expected compiled mono-array collect helper to return the specialized array carrier:\n%s", compiledSrc)
	}
	helperName := "__able_compiled_iface_Iterator_collect_dispatch"
	helperBody, ok := findCompiledFunction(result, helperName)
	if !ok {
		helperName = "__able_compiled_iface_Iterator_collect_default"
		helperBody, ok = findCompiledFunction(result, helperName)
		if !ok {
			t.Fatalf("could not find compiled mono-array collect helper")
		}
	}
	if strings.Contains(helperBody, "__able_method_call_node(") {
		t.Fatalf("expected %s to keep mono-array collect on the compiled helper path:\n%s", helperName, helperBody)
	}
	if strings.Contains(helperBody, "__able_call_named(") {
		t.Fatalf("expected %s to avoid runtime fallback dispatch:\n%s", helperName, helperBody)
	}
	if helperName == "__able_compiled_iface_Iterator_collect_dispatch" && !strings.Contains(helperBody, "__able_compiled_impl_Iterator_collect_default_") {
		t.Fatalf("expected mono-array collect dispatch helper to call the compiled default helper for every receiver case:\n%s", helperBody)
	}
	if helperName == "__able_compiled_iface_Iterator_collect_default" {
		for _, fragment := range []string{
			"__able_compiled_impl_Default__default_0_spec",
			"__able_compiled_impl_Extend_extend_0_spec",
		} {
			if !strings.Contains(helperBody, fragment) {
				t.Fatalf("expected direct mono-array collect helper to contain %q:\n%s", fragment, helperBody)
			}
		}
	}
	if strings.Contains(compiledSrc, "func __able_compiled_iface_Iterator_collect_dispatch_runtime_adapter") {
		t.Fatalf("expected mono-array collect dispatch to avoid generating a runtime-adapter collect fallback helper:\n%s", compiledSrc)
	}
	for _, fragment := range []string{
		"__able_iface_Iterator_i64_to_runtime_value(",
		"__able_method_call_node(",
		"__able_array_i64_from(",
		"__able_call_value(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected concrete Iterator mono-array pipeline to avoid %q:\n%s", fragment, body)
		}
	}
}

func TestCompilerConcreteIteratorGenericMethodsExecuteWithExperimentalMonoArrays(t *testing.T) {
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
		"  print(values.lazy().map<i64>({ value => (value as i64) * 3_i64 }).filter({ value => value >= 6_i64 }).collect<Array i64>().reduce<i64>(0_i64, { acc, value => acc + value }))",
		"}",
		"",
	}, "\n")

	stdout := compileAndRunExecSourceWithOptions(t, "ablec-linked-list-iterator-generic-mono-exec", source, Options{
		PackageName:            "main",
		EmitMain:               true,
		ExperimentalMonoArrays: true,
	})
	if strings.TrimSpace(stdout) != "15" {
		t.Fatalf("expected concrete Iterator mono-array pipeline to print 15, got %q", stdout)
	}
}

func TestCompilerConcreteIteratorMapFilterFunctionStaysNative(t *testing.T) {
	result := compileNoFallbackExecSource(t, "ablec-linked-list-iterator-map-filter-fn", strings.Join([]string{
		"package demo",
		"",
		"import able.collections.linked_list.{LinkedList}",
		"import able.core.iteration.{IteratorEnd}",
		"",
		"fn score(values: LinkedList i32) -> i64 {",
		"  iter := values.lazy().map<i64>({ value => (value as i64) * 3_i64 }).filter({ value => value >= 6_i64 })",
		"  total: i64 = 0_i64",
		"  loop {",
		"    iter.next() match {",
		"      case IteratorEnd {} => { break },",
		"      case value: i64 => { total = total + value }",
		"    }",
		"  }",
		"  total",
		"}",
		"",
		"fn main() -> i64 {",
		"  values: LinkedList i32 = LinkedList.new()",
		"  values.push_back(1)",
		"  values.push_back(2)",
		"  values.push_back(3)",
		"  score(values)",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_score")
	if !ok {
		t.Fatalf("could not find compiled score helper")
	}
	for _, fragment := range []string{
		"__able_compiled_iface_Iterator__map_default(",
		"__able_compiled_iface_Iterator_filter_default(",
		".next()",
	} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("expected score helper to stay on native iterator helpers (%q):\n%s", fragment, body)
		}
	}
	for _, fragment := range []string{
		"__able_iface_Iterator_i64_to_runtime_value(",
		"__able_method_call_node(",
		"__able_call_value(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected score helper to avoid %q:\n%s", fragment, body)
		}
	}
}

func TestCompilerConcreteIteratorMapFilterFunctionExecutes(t *testing.T) {
	source := strings.Join([]string{
		"package main",
		"",
		"import able.collections.linked_list.{LinkedList}",
		"import able.core.iteration.{IteratorEnd}",
		"",
		"fn score(values: LinkedList i32) -> i64 {",
		"  iter := values.lazy().map<i64>({ value => (value as i64) * 3_i64 }).filter({ value => value >= 6_i64 })",
		"  total: i64 = 0_i64",
		"  loop {",
		"    iter.next() match {",
		"      case IteratorEnd {} => { break },",
		"      case value: i64 => { total = total + value }",
		"    }",
		"  }",
		"  total",
		"}",
		"",
		"fn main() -> void {",
		"  values: LinkedList i32 = LinkedList.new()",
		"  values.push_back(1)",
		"  values.push_back(2)",
		"  values.push_back(3)",
		"  print(score(values))",
		"}",
		"",
	}, "\n")

	stdout := compileAndRunExecSourceWithOptions(t, "ablec-linked-list-iterator-map-filter-fn-exec", source, Options{
		PackageName: "main",
		EmitMain:    true,
	})
	if strings.TrimSpace(stdout) != "15" {
		t.Fatalf("expected concrete Iterator map/filter helper pipeline to print 15, got %q", stdout)
	}
}

func TestCompilerConcreteIteratorCollectGenericNominalAccumulatorStaysNative(t *testing.T) {
	result := compileNoFallbackExecSource(t, "ablec-iterator-collect-generic-nominal-native", strings.Join([]string{
		"package demo",
		"",
		"import able.collections.linked_list.{LinkedList}",
		"import able.core.interfaces.{Default, Extend}",
		"",
		"struct SumCount {",
		"  sum: i64,",
		"  count: i64,",
		"}",
		"",
		"impl Default for SumCount {",
		"  fn default() -> Self {",
		"    SumCount { sum: 0_i64, count: 0_i64 }",
		"  }",
		"}",
		"",
		"impl Extend i64 for SumCount {",
		"  fn extend(self: Self, value: i64) -> Self {",
		"    SumCount { sum: self.sum + value, count: self.count + 1_i64 }",
		"  }",
		"}",
		"",
		"fn main() -> i64 {",
		"  values: LinkedList i32 = LinkedList.new()",
		"  values.push_back(1)",
		"  values.push_back(2)",
		"  values.push_back(3)",
		"  result := values.lazy().map<i64>({ value => (value as i64) * 3_i64 }).collect<SumCount>()",
		"  result.sum + result.count",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main")
	}
	for _, fragment := range []string{
		"__able_method_call_node(",
		"__able_call_value(",
		"__able_iface_Iterator_i32_to_runtime_value(",
		"__able_iface_Iterator_i64_to_runtime_value(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected concrete Iterator.collect<C>() to stay on the shared native path without %q:\n%s", fragment, body)
		}
	}
	if !strings.Contains(body, "__able_compiled_iface_Iterator_collect_default") {
		t.Fatalf("expected Iterator.collect<C>() to use the shared compiled default helper:\n%s", body)
	}
	collectName, ok := calledFunctionNameFromBody(body, "__able_compiled_iface_Iterator_collect_default")
	if !ok {
		t.Fatalf("could not find compiled Iterator.collect helper call in main body")
	}
	collectBody, ok := findCompiledFunction(result, collectName)
	if !ok {
		t.Fatalf("could not find compiled Iterator.collect helper body")
	}
	for _, fragment := range []string{
		"__able_method_call_node(",
		"__able_call_value(",
		"__able_iface_Iterator_i64_to_runtime_value(",
		"__able_any_to_value(",
	} {
		if strings.Contains(collectBody, fragment) {
			t.Fatalf("expected shared Iterator.collect<C>() helper to avoid %q:\n%s", fragment, collectBody)
		}
	}
	for _, fragment := range []string{
		"__able_compiled_impl_Default_",
		"__able_compiled_impl_Extend_extend_",
	} {
		if !strings.Contains(collectBody, fragment) {
			t.Fatalf("expected shared Iterator.collect<C>() helper to resolve %q statically:\n%s", fragment, collectBody)
		}
	}
}

func TestCompilerConcreteIteratorCollectGenericNominalAccumulatorExecutes(t *testing.T) {
	source := strings.Join([]string{
		"package main",
		"",
		"import able.collections.linked_list.{LinkedList}",
		"import able.core.interfaces.{Default, Extend}",
		"",
		"struct SumCount {",
		"  sum: i64,",
		"  count: i64,",
		"}",
		"",
		"impl Default for SumCount {",
		"  fn default() -> Self {",
		"    SumCount { sum: 0_i64, count: 0_i64 }",
		"  }",
		"}",
		"",
		"impl Extend i64 for SumCount {",
		"  fn extend(self: Self, value: i64) -> Self {",
		"    SumCount { sum: self.sum + value, count: self.count + 1_i64 }",
		"  }",
		"}",
		"",
		"fn main() -> void {",
		"  values: LinkedList i32 = LinkedList.new()",
		"  values.push_back(1)",
		"  values.push_back(2)",
		"  values.push_back(3)",
		"  result := values.lazy().map<i64>({ value => (value as i64) * 3_i64 }).collect<SumCount>()",
		"  print(result.sum + result.count)",
		"}",
		"",
	}, "\n")

	stdout := compileAndRunExecSourceWithOptions(t, "ablec-iterator-collect-generic-nominal-exec", source, Options{
		PackageName: "main",
		EmitMain:    true,
	})
	if strings.TrimSpace(stdout) != "21" {
		t.Fatalf("expected concrete Iterator.collect<C>() generic nominal accumulator to print 21, got %q", stdout)
	}
}
