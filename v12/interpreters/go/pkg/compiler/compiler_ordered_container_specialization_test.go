package compiler

import (
	"strings"
	"testing"
)

func compiledDeclHeadline(result *Result, prefix string) string {
	if decl, ok := findCompiledDeclByPrefix(result, prefix); ok {
		if idx := strings.IndexByte(decl, '\n'); idx != -1 {
			return decl[:idx]
		}
		return decl
	}
	return ""
}

func TestCompilerTreeMapBenchmarkShapePreservesNativeCarriers(t *testing.T) {
	source := strings.Join([]string{
		"package bench_tree_map_i32_small",
		"",
		"import able.collections.tree_map.{TreeMap, TreeEntry}",
		"",
		"fn build(size: i32) -> TreeMap i32 i32 {",
		"  values: TreeMap i32 i32 = TreeMap.new()",
		"  i := 0",
		"  loop {",
		"    if i >= size { break }",
		"    key := ((i * 37) + 11) % (size * 3)",
		"    values.set(key, (i * 5) - 9)",
		"    i = i + 1",
		"  }",
		"  values",
		"}",
		"",
		"fn score(values: TreeMap i32 i32, size: i32) -> i64 {",
		"  total: i64 = 0_i64",
		"  i := 0",
		"  loop {",
		"    if i >= size * 2 { break }",
		"    key := ((i * 19) + 7) % (size * 3)",
		"    values.get(key) match {",
		"      case nil => values.set(key, i - 3),",
		"      case value: i32 => {",
		"        total = total + (value as i64)",
		"        if i % 7 == 0 {",
		"          values.set(key, value + 1)",
		"        }",
		"      }",
		"    }",
		"    if i % 11 == 0 {",
		"      values.remove((i * 13) % (size * 3))",
		"    }",
		"    i = i + 1",
		"  }",
		"  values.each(fn(entry: TreeEntry i32 i32) -> void {",
		"    total = total + (entry.key as i64) + (entry.value as i64)",
		"  })",
		"  total + (values.len() as i64)",
		"}",
		"",
		"fn main() -> void {",
		"  values := build(12)",
		"  print(score(values, 12))",
		"}",
		"",
	}, "\n")

	result := compileNoFallbackExecSource(t, "ablec-ordered-treemap-shape-native", source)
	for _, prefix := range []string{
		"func __able_compiled_fn_build(size int32)",
		"func __able_compiled_fn_score(values *TreeMap_i32_i32, size int32)",
		"func __able_compiled_method_TreeMap_find_index_spec(self *TreeMap_i32_i32, key int32)",
		"func __able_compiled_method_TreeMap_get_spec(self *TreeMap_i32_i32, key int32)",
		"func __able_compiled_method_TreeMap_insertion_index_spec(self *TreeMap_i32_i32, key int32)",
		"func __able_compiled_method_TreeMap_set_spec(self *TreeMap_i32_i32, key int32, value int32)",
		"func __able_compiled_method_TreeMap_remove_spec(self *TreeMap_i32_i32, key int32)",
		"func __able_compiled_method_TreeMap_each_spec(self *TreeMap_i32_i32",
	} {
		if headline := compiledDeclHeadline(result, prefix); headline == "" {
			t.Fatalf("expected TreeMap benchmark shape to preserve native specialization prefix %q", prefix)
		}
	}
	for _, name := range []string{
		"__able_compiled_method_TreeMap_find_index_spec",
		"__able_compiled_method_TreeMap_get_spec",
		"__able_compiled_method_TreeMap_insertion_index_spec",
		"__able_compiled_method_TreeMap_set_spec",
		"__able_compiled_method_TreeMap_remove_spec",
		"__able_compiled_method_TreeMap_each_spec",
	} {
		body, ok := findCompiledFunction(result, name)
		if !ok {
			t.Fatalf("could not find specialized TreeMap method %s", name)
		}
		for _, fragment := range []string{
			"__able_call_value(",
			"__able_member_get_method(",
			"bridge.MatchType(",
			"__able_try_cast(",
			"runtime.Value",
		} {
			if strings.Contains(body, fragment) {
				t.Fatalf("expected specialized TreeMap method %s to avoid %q:\n%s", name, fragment, body)
			}
		}
	}
	for _, name := range []string{
		"__able_compiled_method_TreeMap_find_index_spec",
		"__able_compiled_method_TreeMap_insertion_index_spec",
	} {
		body, ok := findCompiledFunction(result, name)
		if !ok {
			t.Fatalf("could not find specialized TreeMap method %s", name)
		}
		for _, fragment := range []string{
			"__able_union__Equal_or__Greater_or__Less_to_value(",
			"__able_binary_op(\"==\"",
			"bridge.AsBool(",
		} {
			if strings.Contains(body, fragment) {
				t.Fatalf("expected specialized TreeMap ordering method %s to avoid %q:\n%s", name, fragment, body)
			}
		}
	}

	stdout := compileAndRunExecSourceWithOptions(t, "ablec-ordered-treemap-shape-exec", source, Options{
		PackageName:        "main",
		EmitMain:           true,
		RequireNoFallbacks: true,
	})
	if got := strings.TrimSpace(stdout); got != "949" {
		t.Fatalf("expected TreeMap benchmark shape to print 949, got %q", got)
	}
}

func TestCompilerTreeMapBasicExecutes(t *testing.T) {
	stdout := compileAndRunExecSourceWithOptions(t, "ablec-ordered-treemap-basic-exec", strings.Join([]string{
		"package demo",
		"",
		"import able.collections.tree_map.{TreeMap}",
		"",
		"fn expect_i32(value: ?i32) -> i32 {",
		"  value match {",
		"    case actual: i32 => actual,",
		"    case nil => 0",
		"  }",
		"}",
		"",
		"fn main() -> void {",
		"  values: TreeMap i32 i32 = TreeMap.new()",
		"  values.set(1, 2)",
		"  values.set(3, 4)",
		"  print(expect_i32(values.get(1)) + expect_i32(values.get(3)) + values.len())",
		"}",
		"",
	}, "\n"), Options{
		PackageName:        "main",
		EmitMain:           true,
		RequireNoFallbacks: true,
	})
	if got := strings.TrimSpace(stdout); got != "8" {
		t.Fatalf("expected TreeMap basic exec to print 8, got %q", got)
	}
}

func TestCompilerTreeSetBenchmarkShapeExecutes(t *testing.T) {
	source := strings.Join([]string{
		"package bench_tree_set_i32_small",
		"",
		"import able.collections.tree_set.{TreeSet}",
		"",
		"fn build(size: i32) -> TreeSet i32 {",
		"  values: TreeSet i32 = TreeSet.new()",
		"  i := 0",
		"  loop {",
		"    if i >= size { break }",
		"    values.insert(((i * 29) + 5) % (size * 3))",
		"    i = i + 1",
		"  }",
		"  values",
		"}",
		"",
		"fn score(values: TreeSet i32, size: i32) -> i64 {",
		"  total: i64 = 0_i64",
		"  i := 0",
		"  loop {",
		"    if i >= size * 2 { break }",
		"    key := ((i * 17) + 9) % (size * 3)",
		"    if values.contains(key) {",
		"      total = total + (key as i64)",
		"    } else {",
		"      values.insert(key)",
		"    }",
		"    if i % 5 == 0 {",
		"      values.remove((i * 11) % (size * 3))",
		"    }",
		"    i = i + 1",
		"  }",
		"  values.each(fn(value: i32) -> void { total = total + (value as i64) })",
		"  total + (values.len() as i64)",
		"}",
		"",
		"fn main() -> void {",
		"  values := build(12)",
		"  print(score(values, 12))",
		"}",
		"",
	}, "\n")

	stdout := compileAndRunExecSourceWithOptions(t, "ablec-ordered-treeset-shape-exec", source, Options{
		PackageName:        "main",
		EmitMain:           true,
		RequireNoFallbacks: true,
	})
	if got := strings.TrimSpace(stdout); got != "624" {
		t.Fatalf("expected TreeSet benchmark shape to print 624, got %q", got)
	}
}

func TestCompilerPersistentSortedSetBenchmarkShapeExecutes(t *testing.T) {
	source := strings.Join([]string{
		"package bench_persistent_sorted_set_i32_small",
		"",
		"import able.collections.persistent_sorted_set.{PersistentSortedSet}",
		"",
		"fn build(size: i32) -> PersistentSortedSet i32 {",
		"  values: PersistentSortedSet i32 = PersistentSortedSet.empty()",
		"  i := 0",
		"  loop {",
		"    if i >= size { break }",
		"    values = values.insert(((i * 37) + 1) % (size * 3))",
		"    i = i + 1",
		"  }",
		"  values",
		"}",
		"",
		"fn score(initial: PersistentSortedSet i32, size: i32) -> i64 {",
		"  values := initial",
		"  total: i64 = 0_i64",
		"  round := 0",
		"  loop {",
		"    if round >= 2 { break }",
		"    i := 0",
		"    loop {",
		"      if i >= size * 2 { break }",
		"      key := ((i * 41) + (round * 13) + 5) % (size * 4)",
		"      if values.contains(key) {",
		"        total = total + (key as i64)",
		"      } else {",
		"        values = values.insert(key)",
		"      }",
		"      if i % 9 == 0 {",
		"        values = values.remove((i * 17) % (size * 4))",
		"      }",
		"      i = i + 1",
		"    }",
		"    start := size // 3",
		"    stop := (size * 2) // 3",
		"    window := values.range(start, stop)",
		"    j := 0",
		"    loop {",
		"      if j >= window.len() { break }",
		"      total = total + (window.read_slot(j) as i64)",
		"      j = j + 1",
		"    }",
		"    values.for_each(fn(value: i32) -> void {",
		"      if value % 257 == 0 {",
		"        total = total + (value as i64)",
		"      }",
		"    })",
		"    round = round + 1",
		"  }",
		"  total + (values.len() as i64)",
		"}",
		"",
		"fn main() -> void {",
		"  values := build(36)",
		"  print(score(values, 36))",
		"}",
		"",
	}, "\n")

	stdout := compileAndRunExecSourceWithOptions(t, "ablec-ordered-psset-shape-exec", source, Options{
		PackageName:        "main",
		EmitMain:           true,
		RequireNoFallbacks: true,
	})
	if got := strings.TrimSpace(stdout); got != "2466" {
		t.Fatalf("expected PersistentSortedSet benchmark shape to print 2466, got %q", got)
	}
}
