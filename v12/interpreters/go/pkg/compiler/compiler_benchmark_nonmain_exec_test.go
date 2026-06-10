package compiler

import (
	"strings"
	"testing"
)

func TestCompilerHashMapBenchmarkShapeExecutesFromNonMainSourcePackage(t *testing.T) {
	source := strings.Join([]string{
		"package bench_hashmap_i32_small",
		"",
		"import able.collections.hash_map.*",
		"import able.collections.map.*",
		"",
		"fn build(size: i32) -> HashMap i32 i32 {",
		"  values: HashMap i32 i32 = HashMap.with_capacity(size)",
		"  i := 0",
		"  loop {",
		"    if i >= size { break }",
		"    values.raw_set(i, (i * 3) + 1)",
		"    i = i + 1",
		"  }",
		"  values",
		"}",
		"",
		"fn checksum(values: Map i32 i32, size: i32) -> i64 {",
		"  total: i64 = 0_i64",
		"  i := 0",
		"  loop {",
		"    if i >= size { break }",
		"    maybe_value := values.get(i)",
		"    value: i32 := maybe_value or { 0 }",
		"    total = total + (value as i64)",
		"    i = i + 1",
		"  }",
		"  total",
		"}",
		"",
		"fn main() -> void {",
		"  size := 8",
		"  values := build(size)",
		"  rounds := 0",
		"  total: i64 = 0_i64",
		"  loop {",
		"    if rounds >= 3 { break }",
		"    total = total + checksum(values, size)",
		"    maybe_current := values.raw_get(rounds)",
		"    current: i32 := maybe_current or { 0 }",
		"    values.raw_set(rounds, current + 1)",
		"    rounds = rounds + 1",
		"  }",
		"  print(total)",
		"}",
		"",
	}, "\n")

	stdout := compileAndRunExecSourceWithOptions(t, "ablec-bench-hashmap-nonmain-exec", source, Options{
		PackageName: "main",
		EmitMain:    true,
	})
	if got := strings.TrimSpace(stdout); got != "279" {
		t.Fatalf("expected HashMap benchmark shape to print 279, got %q", got)
	}
}

func TestCompilerIteratorCollectBenchmarkShapeExecutesFromNonMainSourcePackage(t *testing.T) {
	source := strings.Join([]string{
		"package bench_linked_list_iterator_collect_i64_small",
		"",
		"import able.collections.linked_list.{LinkedList}",
		"",
		"fn build(size: i32) -> LinkedList i32 {",
		"  values: LinkedList i32 = LinkedList.new()",
		"  i := 0",
		"  loop {",
		"    if i >= size { break }",
		"    values.push_back(i)",
		"    i = i + 1",
		"  }",
		"  values",
		"}",
		"",
		"fn score(values: LinkedList i32) -> i64 {",
		"  values",
		"    .lazy()",
		"    .map<i64>({ value => (value as i64) * 3_i64 })",
		"    .filter({ value => value >= 6_i64 })",
		"    .collect<Array i64>()",
		"    .reduce<i64>(0_i64, { acc, value => acc + value })",
		"}",
		"",
		"fn main() -> void {",
		"  values := build(8)",
		"  rounds := 0",
		"  total: i64 = 0_i64",
		"  loop {",
		"    if rounds >= 2 { break }",
		"    total = total + score(values)",
		"    rounds = rounds + 1",
		"  }",
		"  print(total)",
		"}",
		"",
	}, "\n")

	stdout := compileAndRunExecSourceWithOptions(t, "ablec-bench-iter-collect-nonmain-exec", source, Options{
		PackageName: "main",
		EmitMain:    true,
	})
	if got := strings.TrimSpace(stdout); got != "162" {
		t.Fatalf("expected Iterator.collect benchmark shape to print 162, got %q", got)
	}
}

func TestCompilerIteratorFilterMapBenchmarkShapeExecutesFromNonMainSourcePackage(t *testing.T) {
	source := strings.Join([]string{
		"package bench_linked_list_iterator_filter_map_i64_small",
		"",
		"import able.collections.linked_list.{LinkedList}",
		"",
		"fn build(size: i32) -> LinkedList i32 {",
		"  values: LinkedList i32 = LinkedList.new()",
		"  i := 0",
		"  loop {",
		"    if i >= size { break }",
		"    values.push_back(i)",
		"    i = i + 1",
		"  }",
		"  values",
		"}",
		"",
		"fn score(values: LinkedList i32) -> i64 {",
		"  values",
		"    .lazy()",
		"    .filter_map<i64>({ value => if value % 2 == 0 { (value as i64) * 3_i64 } else { nil } })",
		"    .collect<Array i64>()",
		"    .reduce<i64>(0_i64, { acc, value => acc + value })",
		"}",
		"",
		"fn main() -> void {",
		"  values := build(8)",
		"  rounds := 0",
		"  total: i64 = 0_i64",
		"  loop {",
		"    if rounds >= 2 { break }",
		"    total = total + score(values)",
		"    rounds = rounds + 1",
		"  }",
		"  print(total)",
		"}",
		"",
	}, "\n")

	stdout := compileAndRunExecSourceWithOptions(t, "ablec-bench-iter-filtermap-nonmain-exec", source, Options{
		PackageName: "main",
		EmitMain:    true,
	})
	if got := strings.TrimSpace(stdout); got != "72" {
		t.Fatalf("expected Iterator.filter_map benchmark shape to print 72, got %q", got)
	}
}

func TestCompilerIteratorPipelineBenchmarkShapeExecutesFromNonMainSourcePackage(t *testing.T) {
	source := strings.Join([]string{
		"package bench_linked_list_iterator_pipeline_i64_small",
		"",
		"import able.collections.linked_list.{LinkedList}",
		"import able.core.iteration.{IteratorEnd}",
		"",
		"fn build(size: i32) -> LinkedList i32 {",
		"  values: LinkedList i32 = LinkedList.new()",
		"  i := 0",
		"  loop {",
		"    if i >= size { break }",
		"    values.push_back(i)",
		"    i = i + 1",
		"  }",
		"  values",
		"}",
		"",
		"fn score(values: LinkedList i32) -> i64 {",
		"  iter := values",
		"    .lazy()",
		"    .map<i64>({ value => (value as i64) * 3_i64 })",
		"    .filter({ value => value >= 6_i64 })",
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
		"  values := build(8)",
		"  rounds := 0",
		"  total: i64 = 0_i64",
		"  loop {",
		"    if rounds >= 2 { break }",
		"    total = total + score(values)",
		"    rounds = rounds + 1",
		"  }",
		"  print(total)",
		"}",
		"",
	}, "\n")

	stdout := compileAndRunExecSourceWithOptions(t, "ablec-bench-iter-pipeline-nonmain-exec", source, Options{
		PackageName: "main",
		EmitMain:    true,
	})
	if got := strings.TrimSpace(stdout); got != "162" {
		t.Fatalf("expected Iterator pipeline benchmark shape to print 162, got %q", got)
	}
}
