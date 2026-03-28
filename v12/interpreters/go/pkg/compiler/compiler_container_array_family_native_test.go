package compiler

import (
	"strings"
	"testing"
)

func TestCompilerDequeQueueMethodsStayNative(t *testing.T) {
	result := compileNoFallbackExecSource(t, "ablec-deque-queue-native", strings.Join([]string{
		"package demo",
		"",
		"import able.collections.deque.*",
		"import able.collections.queue.*",
		"",
		"fn main() -> i32 {",
		"  values: Deque i32 = Deque.with_capacity(4)",
		"  values.push_back(1)",
		"  values.push_front(0)",
		"  queue: Queue i32 = Queue.new()",
		"  queue.enqueue(7)",
		"  queue.size() + values.len()",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	for _, fragment := range []string{
		"var values *Deque",
		"var queue *Queue",
	} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("expected native Deque/Queue locals to contain %q:\n%s", fragment, body)
		}
	}

	for _, name := range []string{
		"__able_compiled_method_Deque_push_back",
		"__able_compiled_method_Deque_grow",
		"__able_compiled_method_Queue_enqueue",
	} {
		methodBody, ok := findCompiledFunction(result, name)
		if !ok {
			t.Fatalf("could not find compiled method %s", name)
		}
		for _, fragment := range []string{
			"__able_call_value(",
			"__able_member_get_method(",
			"__able_method_call_node(",
			"bridge.MatchType(",
			"__able_try_cast(",
		} {
			if strings.Contains(methodBody, fragment) {
				t.Fatalf("expected %s to avoid %q:\n%s", name, fragment, methodBody)
			}
		}
	}
}

func TestCompilerUnannotatedQueueLocalPreservesMutationsAcrossRecoveredCalls(t *testing.T) {
	stdout := compileAndRunExecSourceWithOptions(t, "ablec-unannotated-queue-local-writeback", strings.Join([]string{
		"package demo",
		"",
		"import able.collections.queue.*",
		"",
		"fn expect_i32(value: ?i32) -> i32 {",
		"  value match {",
		"    case actual: i32 => actual,",
		"    case nil => 0",
		"  }",
		"}",
		"",
		"fn main() -> void {",
		"  queue := Queue.new()",
		"  queue.enqueue(1)",
		"  queue.enqueue(2)",
		"  queue.enqueue(3)",
		"  print(expect_i32(queue.peek()) + expect_i32(queue.dequeue()) + expect_i32(queue.peek()))",
		"}",
		"",
	}, "\n"), Options{
		PackageName:        "main",
		EmitMain:           true,
		RequireNoFallbacks: true,
	})

	if got := strings.TrimSpace(stdout); got != "4" {
		t.Fatalf("expected output 4, got %q", got)
	}
}

func TestCompilerUnannotatedDequeLocalAvoidsRedundantSpecializationRoundTrip(t *testing.T) {
	source := strings.Join([]string{
		"package demo",
		"",
		"import able.collections.deque.*",
		"",
		"fn expect_i32(value: ?i32) -> i32 {",
		"  value match {",
		"    case actual: i32 => actual,",
		"    case nil => 0",
		"  }",
		"}",
		"",
		"fn main() -> void {",
		"  deque := Deque.new()",
		"  deque.push_back(2)",
		"  deque.push_front(1)",
		"  deque.push_back(3)",
		"  print(expect_i32(deque.front()) + expect_i32(deque.back()))",
		"}",
		"",
	}, "\n")

	result := compileNoFallbackExecSource(t, "ablec-unannotated-deque-roundtrip-audit", source)
	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	for _, fragment := range []string{
		"__able_nominal_coerce_Deque_to_Deque_i32(",
		"__able_nominal_coerce_Deque_i32_to_Deque(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected unannotated deque calls to avoid redundant same-family coercion %q:\n%s", fragment, body)
		}
	}

	stdout := compileAndRunExecSourceWithOptions(t, "ablec-unannotated-deque-roundtrip-exec", source, Options{
		PackageName:        "main",
		EmitMain:           true,
		RequireNoFallbacks: true,
	})
	if got := strings.TrimSpace(stdout); got != "4" {
		t.Fatalf("expected output 4, got %q", got)
	}
}

func TestCompilerUnannotatedQueueIterationAcrossRecoveredCallsExecutes(t *testing.T) {
	stdout := compileAndRunExecSourceWithOptions(t, "ablec-unannotated-queue-iteration", strings.Join([]string{
		"package demo",
		"",
		"import able.collections.queue.*",
		"import able.core.iteration.*",
		"",
		"fn main() -> void {",
		"  queue := Queue.new()",
		"  queue.enqueue(5)",
		"  queue.enqueue(7)",
		"  queue.enqueue(9)",
		"  iter := queue.iterator()",
		"  sum := 0",
		"  loop {",
		"    iter.next() match {",
		"      case value: i32 => sum = sum + value,",
		"      case IteratorEnd {} => { break }",
		"    }",
		"  }",
		"  for_sum := 0",
		"  for value in queue {",
		"    for_sum = for_sum + value",
		"  }",
		"  print(sum + for_sum)",
		"}",
		"",
	}, "\n"), Options{
		PackageName:        "main",
		EmitMain:           true,
		RequireNoFallbacks: true,
	})

	if got := strings.TrimSpace(stdout); got != "42" {
		t.Fatalf("expected output 42, got %q", got)
	}
}

func TestCompilerBitSetHeapMethodsStayNative(t *testing.T) {
	result := compileNoFallbackExecSource(t, "ablec-bitset-heap-native", strings.Join([]string{
		"package demo",
		"",
		"import able.collections.bit_set.*",
		"import able.collections.heap.*",
		"",
		"fn main() -> i32 {",
		"  bits := BitSet.new()",
		"  bits.set(1)",
		"  heap: Heap i32 = Heap.new()",
		"  heap.push(4)",
		"  heap.push(1)",
		"  heap.len() + if bits.contains(1) { 1 } else { 0 }",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	for _, fragment := range []string{
		"var bits *BitSet =",
		"var heap *Heap",
	} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("expected native BitSet/Heap locals to contain %q:\n%s", fragment, body)
		}
	}

	for _, name := range []string{
		"__able_compiled_method_BitSet_set",
		"__able_compiled_method_BitSet_contains",
		"__able_compiled_method_Heap_push",
		"__able_compiled_method_Heap_pop",
	} {
		methodBody, ok := findCompiledFunction(result, name)
		if !ok {
			t.Fatalf("could not find compiled method %s", name)
		}
		for _, fragment := range []string{
			"__able_call_value(",
			"__able_member_get_method(",
			"__able_method_call_node(",
			"bridge.MatchType(",
			"__able_try_cast(",
		} {
			if strings.Contains(methodBody, fragment) {
				t.Fatalf("expected %s to avoid %q:\n%s", name, fragment, methodBody)
			}
		}
	}
}

func TestCompilerPersistentSortedQueueMethodsStayNative(t *testing.T) {
	result := compileNoFallbackExecSource(t, "ablec-persistent-sorted-queue-native", strings.Join([]string{
		"package demo",
		"",
		"import able.collections.persistent_sorted_set.*",
		"import able.collections.persistent_queue.*",
		"",
		"fn main() -> i32 {",
		"  set: PersistentSortedSet i32 = PersistentSortedSet.empty()",
		"  set = set.insert(2).insert(1)",
		"  queue: PersistentQueue i32 = PersistentQueue.empty()",
		"  queue = queue.enqueue(10).enqueue(20)",
		"  set.len() + queue.len()",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	for _, fragment := range []string{
		"var set *PersistentSortedSet_i32 =",
		"var queue *PersistentQueue_i32 =",
		"__able_compiled_method_PersistentSortedSet_insert_spec",
		"__able_compiled_method_PersistentQueue_enqueue_spec",
	} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("expected native persistent container locals to contain %q:\n%s", fragment, body)
		}
	}

	for _, prefix := range []string{
		"func __able_compiled_method_PersistentSortedSet_insert_spec",
		"func __able_compiled_method_PersistentQueue_enqueue_spec",
	} {
		methodBody, ok := findCompiledDeclByPrefix(result, prefix)
		if !ok {
			t.Fatalf("could not find compiled method with prefix %s", prefix)
		}
		for _, fragment := range []string{
			"__able_call_value(",
			"__able_member_get_method(",
			"__able_method_call_node(",
			"bridge.MatchType(",
			"__able_try_cast(",
		} {
			if strings.Contains(methodBody, fragment) {
				t.Fatalf("expected %s to avoid %q:\n%s", prefix, fragment, methodBody)
			}
		}
	}
}
