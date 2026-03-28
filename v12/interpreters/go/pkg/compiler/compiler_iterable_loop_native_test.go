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
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected interface iterable for-loop to avoid %q:\n%s", fragment, body)
		}
	}
	for _, fragment := range []string{
		"__able_iface_Iterable_i32",
		"__able_iface_Iterator_i32",
		"__able_compiled_iface_Iterable_iterator_default",
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
