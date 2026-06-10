package compiler

import (
	"strings"
	"testing"
)

func TestCompilerConcreteIterableForLoopStaysNative(t *testing.T) {
	result := compileNoFallbackExecSource(t, "ablec-concrete-iterable-loop-native", strings.Join([]string{
		"package demo",
		"",
		"import able.core.iteration.{Iterable, Iterator}",
		"",
		"struct Counter { stop: i32 }",
		"",
		"impl Iterable i32 for Counter {",
		"  fn iterator(self: Self) -> (Iterator i32) {",
		"    Iterator i32 { gen =>",
		"      i := 0",
		"      while i < self.stop {",
		"        gen.yield(i)",
		"        i = i + 1",
		"      }",
		"    }",
		"  }",
		"}",
		"",
		"fn main() -> i32 {",
		"  counter := Counter { stop: 4 }",
		"  total := 0",
		"  for value in counter {",
		"    total = total + value",
		"  }",
		"  total",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	for _, fragment := range []string{
		"__able_resolve_iterator(",
		"__able_method_call_node(",
		"__able_call_value(",
		"runtime.IteratorValue",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected concrete iterable for-loop to avoid %q:\n%s", fragment, body)
		}
	}
	for _, fragment := range []string{
		"__able_compiled_impl_Iterable_iterator_0_spec(",
		"__able_iface_Iterator_i32",
		".next()",
	} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("expected concrete iterable for-loop to contain %q:\n%s", fragment, body)
		}
	}
}

func TestCompilerInterfaceIterableForLoopStaysNative(t *testing.T) {
	result := compileNoFallbackExecSource(t, "ablec-interface-iterable-loop-native", strings.Join([]string{
		"package demo",
		"",
		"import able.core.iteration.{Iterable, Iterator}",
		"",
		"struct Counter { stop: i32 }",
		"",
		"impl Iterable i32 for Counter {",
		"  fn iterator(self: Self) -> (Iterator i32) {",
		"    Iterator i32 { gen =>",
		"      i := 0",
		"      while i < self.stop {",
		"        gen.yield(i)",
		"        i = i + 1",
		"      }",
		"    }",
		"  }",
		"}",
		"",
		"fn main() -> i32 {",
		"  iterable: Iterable i32 = Counter { stop: 4 }",
		"  total := 0",
		"  for value in iterable {",
		"    total = total + value",
		"  }",
		"  total",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	for _, fragment := range []string{
		"__able_resolve_iterator(",
		"__able_method_call_node(",
		"__able_call_value(",
		"runtime.IteratorValue",
		"__able_compiled_entry_iface_Iterable_iterator_default(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected interface iterable for-loop to avoid %q:\n%s", fragment, body)
		}
	}
	for _, fragment := range []string{
		"__able_iface_Iterable_i32",
		"__able_iface_Iterator_i32",
		"iterable.iterator()",
		".next()",
	} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("expected interface iterable for-loop to contain %q:\n%s", fragment, body)
		}
	}
}

func TestCompilerConcreteIterableArgToInterfaceParamStaysNative(t *testing.T) {
	result := compileNoFallbackExecSource(t, "ablec-iterable-interface-param-native", strings.Join([]string{
		"package demo",
		"",
		"import able.core.iteration.{Iterable, Iterator}",
		"",
		"struct Counter { stop: i32 }",
		"",
		"impl Iterable i32 for Counter {",
		"  fn iterator(self: Self) -> (Iterator i32) {",
		"    Iterator i32 { gen =>",
		"      i := 0",
		"      while i < self.stop {",
		"        gen.yield(i)",
		"        i = i + 1",
		"      }",
		"    }",
		"  }",
		"}",
		"",
		"fn checksum(values: Iterable i32) -> i32 {",
		"  total := 0",
		"  for value in values {",
		"    total = total + value",
		"  }",
		"  total",
		"}",
		"",
		"fn main() -> i32 {",
		"  counter := Counter { stop: 4 }",
		"  checksum(counter)",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	if !strings.Contains(body, "__able_iface_Iterable_i32_wrap_ptr_Counter(") {
		t.Fatalf("expected concrete iterable arg to wrap directly into the native interface carrier:\n%s", body)
	}
	for _, fragment := range []string{
		"__able_any_to_value(counter)",
		"__able_iface_Iterable_i32_from_value(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected concrete iterable arg to avoid %q:\n%s", fragment, body)
		}
	}
}

func TestCompilerIterableForLoopNarrowsIteratorValueBeforeIntegerWidening(t *testing.T) {
	result := compileNoFallbackExecSource(t, "ablec-iterable-loop-value-narrowing", strings.Join([]string{
		"package demo",
		"",
		"import able.core.iteration.{Iterable, Iterator}",
		"",
		"struct Counter { stop: i32 }",
		"",
		"impl Iterable i32 for Counter {",
		"  fn iterator(self: Self) -> (Iterator i32) {",
		"    Iterator i32 { gen =>",
		"      i := 0",
		"      while i < self.stop {",
		"        gen.yield(i)",
		"        i = i + 1",
		"      }",
		"    }",
		"  }",
		"}",
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
		"  checksum(Counter { stop: 4 })",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_checksum")
	if !ok {
		t.Fatalf("could not find compiled checksum function")
	}
	if !strings.Contains(body, "__able_union__IteratorEnd_or_int32_as_int32(") {
		t.Fatalf("expected iterable for-loop to extract the native int32 iterator branch directly:\n%s", body)
	}
	for _, fragment := range []string{
		"__able_union__IteratorEnd_or_int32_to_value(__able_runtime, value)",
		"__able_cast(",
		"bridge.AsInt(__able_tmp_",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected iterable for-loop to avoid %q after IteratorEnd narrowing:\n%s", fragment, body)
		}
	}
}
