package compiler

import (
	"strings"
	"testing"
)

func compiledSourceText(t *testing.T, result *Result) string {
	t.Helper()
	src, ok := result.Files["compiled.go"]
	if !ok {
		t.Fatalf("compiled.go not found in result files")
	}
	return string(src)
}

func TestCompilerGenericNominalMethodSpecializationStaysNative(t *testing.T) {
	result := compileNoFallbackExecSource(t, "ablec-generic-nominal-method-spec", strings.Join([]string{
		"package demo",
		"",
		"struct Box T {",
		"  value: T",
		"}",
		"",
		"methods Box T {",
		"  fn new(value: T) -> Box T {",
		"    Box { value: value }",
		"  }",
		"",
		"  fn set(self: Self, value: T) -> void {",
		"    self.value = value",
		"  }",
		"",
		"  fn same(a: T, b: T) -> bool {",
		"    a == b",
		"  }",
		"",
		"  fn check(self: Self, other: T) -> bool {",
		"    Box.same(self.value, other)",
		"  }",
		"}",
		"",
		"fn main() -> i32 {",
		"  box: Box i32 = Box.new(1)",
		"  box.set(2)",
		"  if box.check(2) { 1 } else { 0 }",
		"}",
		"",
	}, "\n"))

	mainBody, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	source := compiledSourceText(t, result)
	for _, fragment := range []string{
		"__able_compiled_method_Box_set_spec(",
		"__able_compiled_method_Box_check_spec(",
	} {
		if !strings.Contains(mainBody, fragment) {
			t.Fatalf("expected main to use specialized nominal method %q:\n%s", fragment, mainBody)
		}
	}

	if !strings.Contains(source, "func __able_compiled_method_Box_set_spec(self *Box_i32, value int32)") {
		t.Fatalf("expected specialized Box.set signature to lower Self -> *Box_i32 and T -> int32:\n%s", source)
	}
	if !strings.Contains(source, "func __able_compiled_method_Box_same_spec(a int32, b int32)") {
		t.Fatalf("expected specialized Box.same signature to lower T -> int32:\n%s", source)
	}

	sameBody, ok := findCompiledFunction(result, "__able_compiled_method_Box_same_spec")
	if !ok {
		t.Fatalf("could not find specialized Box.same method")
	}
	for _, fragment := range []string{
		"__able_method_call_node(",
		"__able_call_value(",
		"bridge.MatchType(",
		"__able_try_cast(",
	} {
		if strings.Contains(sameBody, fragment) {
			t.Fatalf("expected specialized Box.same to avoid %q:\n%s", fragment, sameBody)
		}
	}
}

func TestCompilerHeapGenericMethodSpecializationStaysNative(t *testing.T) {
	result := compileNoFallbackExecSource(t, "ablec-heap-generic-method-spec", strings.Join([]string{
		"package demo",
		"",
		"import able.collections.heap.*",
		"",
		"fn main() -> i32 {",
		"  heap: Heap i32 = Heap.new()",
		"  heap.push(4)",
		"  heap.push(1)",
		"  heap.push(3)",
		"  heap.pop()!",
		"}",
		"",
	}, "\n"))

	mainBody, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	source := compiledSourceText(t, result)
	if !strings.Contains(mainBody, "__able_compiled_method_Heap_push_spec(") {
		t.Fatalf("expected main to use specialized Heap.push:\n%s", mainBody)
	}

	if !strings.Contains(source, "func __able_compiled_method_Heap_push_spec(self *Heap_i32, value int32)") {
		t.Fatalf("expected specialized Heap.push signature to lower Self -> *Heap_i32 and T -> int32:\n%s", source)
	}
	if !strings.Contains(source, "func __able_compiled_method_Heap_compare_values_spec(a int32, b int32)") {
		t.Fatalf("expected specialized Heap.compare_values signature to lower T -> int32:\n%s", source)
	}

	compareBody, ok := findCompiledFunction(result, "__able_compiled_method_Heap_compare_values_spec")
	if !ok {
		t.Fatalf("could not find specialized Heap.compare_values")
	}
	for _, fragment := range []string{
		"__able_method_call_node(",
		"bridge.MatchType(",
		"__able_try_cast(",
		"runtime.Value",
	} {
		if strings.Contains(compareBody, fragment) {
			t.Fatalf("expected specialized Heap.compare_values to avoid %q:\n%s", fragment, compareBody)
		}
	}
}

func TestCompilerBoundGenericFieldCarrierSpecializationStaysNative(t *testing.T) {
	result := compileNoFallbackExecSourceWithOptions(t, "ablec-bound-generic-field-carrier-spec", strings.Join([]string{
		"package demo",
		"",
		"struct Bucket T {",
		"  items: Array T",
		"}",
		"",
		"methods Bucket T {",
		"  fn new() -> Bucket T {",
		"    Bucket { items: Array.with_capacity(4) }",
		"  }",
		"",
		"  fn push(self: Self, value: T) -> void {",
		"    self.items.push(value)",
		"  }",
		"",
		"  fn second(self: Self) -> T {",
		"    self.items[1]",
		"  }",
		"}",
		"",
		"fn main() -> i32 {",
		"  bucket: Bucket i32 = Bucket.new()",
		"  bucket.push(1)",
		"  bucket.push(2)",
		"  bucket.second()",
		"}",
		"",
	}, "\n"), Options{
		PackageName:            "main",
		RequireNoFallbacks:     true,
		ExperimentalMonoArrays: true,
	})

	mainBody, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	source := compiledSourceText(t, result)
	for _, fragment := range []string{
		"__able_compiled_method_Bucket_push_spec(",
		"__able_compiled_method_Bucket_second_spec(",
		"type Bucket_i32 struct",
		"Items *__able_array_i32",
	} {
		if !strings.Contains(source, fragment) && !strings.Contains(mainBody, fragment) {
			t.Fatalf("expected bound generic field carrier specialization to contain %q:\n%s", fragment, source)
		}
	}

	for _, name := range []string{
		"__able_compiled_method_Bucket_push_spec",
		"__able_compiled_method_Bucket_second_spec",
	} {
		methodBody, ok := findCompiledFunction(result, name)
		if !ok {
			t.Fatalf("could not find specialized Bucket method %s", name)
		}
		for _, fragment := range []string{
			"__able_method_call_node(",
			"__able_call_value(",
			"bridge.MatchType(",
			"__able_try_cast(",
			"runtime.Value",
			"*Array",
		} {
			if strings.Contains(methodBody, fragment) {
				t.Fatalf("expected %s to avoid %q:\n%s", name, fragment, methodBody)
			}
		}
	}
}

func TestCompilerGenericNominalMethodSpecializationExecutes(t *testing.T) {
	stdout := compileAndRunExecSourceWithOptions(t, "ablec-generic-nominal-method-spec-exec", strings.Join([]string{
		"package demo",
		"",
		"struct Box T {",
		"  value: T",
		"}",
		"",
		"methods Box T {",
		"  fn new(value: T) -> Box T {",
		"    Box { value: value }",
		"  }",
		"",
		"  fn set(self: Self, value: T) -> void {",
		"    self.value = value",
		"  }",
		"",
		"  fn same(a: T, b: T) -> bool {",
		"    a == b",
		"  }",
		"",
		"  fn check(self: Self, other: T) -> bool {",
		"    Box.same(self.value, other)",
		"  }",
		"}",
		"",
		"fn main() -> void {",
		"  box: Box i32 = Box.new(1)",
		"  box.set(2)",
		"  print(if box.check(2) { 1 } else { 0 })",
		"}",
		"",
	}, "\n"), Options{
		PackageName: "main",
		EmitMain:    true,
	})
	if got := strings.TrimSpace(stdout); got != "1" {
		t.Fatalf("expected output 1, got %q", got)
	}
}

func TestCompilerBoundGenericFieldCarrierSpecializationExecutes(t *testing.T) {
	stdout := compileAndRunExecSourceWithOptions(t, "ablec-bound-generic-field-carrier-exec", strings.Join([]string{
		"package demo",
		"",
		"struct Bucket T {",
		"  items: Array T",
		"}",
		"",
		"methods Bucket T {",
		"  fn new() -> Bucket T {",
		"    Bucket { items: Array.with_capacity(4) }",
		"  }",
		"",
		"  fn push(self: Self, value: T) -> void {",
		"    self.items.push(value)",
		"  }",
		"",
		"  fn second(self: Self) -> T {",
		"    self.items[1]",
		"  }",
		"}",
		"",
		"fn main() -> void {",
		"  bucket: Bucket i32 = Bucket.new()",
		"  bucket.push(1)",
		"  bucket.push(2)",
		"  print(bucket.second())",
		"}",
		"",
	}, "\n"), Options{
		PackageName:            "main",
		EmitMain:               true,
		ExperimentalMonoArrays: true,
	})
	if got := strings.TrimSpace(stdout); got != "2" {
		t.Fatalf("expected output 2, got %q", got)
	}
}

func TestCompilerGenericStaticNominalMethodInfersInterfaceParamBindingStaysNative(t *testing.T) {
	result := compileNoFallbackExecSource(t, "ablec-static-nominal-interface-param-spec", strings.Join([]string{
		"package demo",
		"",
		"interface Reader T {",
		"  fn read(self) -> T",
		"}",
		"",
		"struct IntReader {",
		"  value: i32",
		"}",
		"",
		"impl Reader i32 for IntReader {",
		"  fn read(self: Self) -> i32 {",
		"    self.value",
		"  }",
		"}",
		"",
		"struct Holder T {",
		"  value: T",
		"}",
		"",
		"methods Holder T {",
		"  fn from_reader(reader: Reader T) -> Holder T {",
		"    Holder { value: reader.read() }",
		"  }",
		"}",
		"",
		"fn main() -> i32 {",
		"  reader := IntReader { value: 7 }",
		"  holder: Holder i32 = Holder.from_reader(reader)",
		"  holder.value",
		"}",
		"",
	}, "\n"))

	mainBody, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	source := compiledSourceText(t, result)
	if !strings.Contains(mainBody, "__able_compiled_method_Holder_from_reader_spec(") {
		t.Fatalf("expected main to call the specialized static nominal method:\n%s", mainBody)
	}
	for _, fragment := range []string{
		"func __able_compiled_method_Holder_from_reader_spec(reader __able_iface_Reader_i32) (*Holder_i32, *__ableControl)",
		"var holder *Holder_i32 = ",
	} {
		if !strings.Contains(source, fragment) && !strings.Contains(mainBody, fragment) {
			t.Fatalf("expected static nominal interface-param specialization to contain %q:\n%s", fragment, source)
		}
	}
	specBody, ok := findCompiledFunction(result, "__able_compiled_method_Holder_from_reader_spec")
	if !ok {
		t.Fatalf("could not find specialized Holder.from_reader")
	}
	for _, fragment := range []string{
		"__able_method_call_node(",
		"__able_call_value(",
		"bridge.MatchType(",
		"runtime.Value",
	} {
		if strings.Contains(specBody, fragment) {
			t.Fatalf("expected specialized Holder.from_reader to avoid %q:\n%s", fragment, specBody)
		}
	}
}

func TestCompilerGenericStaticNominalMethodInfersInterfaceParamBindingExecutes(t *testing.T) {
	stdout := compileAndRunExecSourceWithOptions(t, "ablec-static-nominal-interface-param-exec", strings.Join([]string{
		"package demo",
		"",
		"interface Reader T {",
		"  fn read(self) -> T",
		"}",
		"",
		"struct IntReader {",
		"  value: i32",
		"}",
		"",
		"impl Reader i32 for IntReader {",
		"  fn read(self: Self) -> i32 {",
		"    self.value",
		"  }",
		"}",
		"",
		"struct Holder T {",
		"  value: T",
		"}",
		"",
		"methods Holder T {",
		"  fn from_reader(reader: Reader T) -> Holder T {",
		"    Holder { value: reader.read() }",
		"  }",
		"}",
		"",
		"fn main() -> void {",
		"  reader := IntReader { value: 7 }",
		"  holder: Holder i32 = Holder.from_reader(reader)",
		"  print(holder.value)",
		"}",
		"",
	}, "\n"), Options{
		PackageName: "main",
		EmitMain:    true,
	})
	if got := strings.TrimSpace(stdout); got != "7" {
		t.Fatalf("expected output 7, got %q", got)
	}
}

func TestCompilerGenericStaticNominalMethodInfersInterfaceParamBindingWithoutExpectedStaysNative(t *testing.T) {
	result := compileNoFallbackExecSource(t, "ablec-static-nominal-interface-param-inferred-local-spec", strings.Join([]string{
		"package demo",
		"",
		"interface Reader T {",
		"  fn read(self) -> T",
		"}",
		"",
		"struct IntReader {",
		"  value: i32",
		"}",
		"",
		"impl Reader i32 for IntReader {",
		"  fn read(self: Self) -> i32 {",
		"    self.value",
		"  }",
		"}",
		"",
		"struct Holder T {",
		"  value: T",
		"}",
		"",
		"methods Holder T {",
		"  fn from_reader(reader: Reader T) -> Holder T {",
		"    Holder { value: reader.read() }",
		"  }",
		"}",
		"",
		"fn main() -> i32 {",
		"  reader := IntReader { value: 9 }",
		"  holder := Holder.from_reader(reader)",
		"  holder.value",
		"}",
		"",
	}, "\n"))

	mainBody, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	source := compiledSourceText(t, result)
	if !strings.Contains(mainBody, "__able_compiled_method_Holder_from_reader_spec(") {
		t.Fatalf("expected inferred-local main to call the specialized static nominal method:\n%s", mainBody)
	}
	for _, fragment := range []string{
		"func __able_compiled_method_Holder_from_reader_spec(reader __able_iface_Reader_i32) (*Holder_i32, *__ableControl)",
		"var holder *Holder_i32 = ",
	} {
		if !strings.Contains(source, fragment) && !strings.Contains(mainBody, fragment) {
			t.Fatalf("expected inferred-local static nominal interface-param specialization to contain %q:\n%s", fragment, source)
		}
	}
	specBody, ok := findCompiledFunction(result, "__able_compiled_method_Holder_from_reader_spec")
	if !ok {
		t.Fatalf("could not find specialized Holder.from_reader")
	}
	for _, fragment := range []string{
		"__able_method_call_node(",
		"__able_call_value(",
		"bridge.MatchType(",
		"runtime.Value",
	} {
		if strings.Contains(specBody, fragment) {
			t.Fatalf("expected inferred-local specialized Holder.from_reader to avoid %q:\n%s", fragment, specBody)
		}
	}
}

func TestCompilerGenericStaticNominalMethodInfersInterfaceParamBindingWithoutExpectedExecutes(t *testing.T) {
	stdout := compileAndRunExecSourceWithOptions(t, "ablec-static-nominal-interface-param-inferred-local-exec", strings.Join([]string{
		"package demo",
		"",
		"interface Reader T {",
		"  fn read(self) -> T",
		"}",
		"",
		"struct IntReader {",
		"  value: i32",
		"}",
		"",
		"impl Reader i32 for IntReader {",
		"  fn read(self: Self) -> i32 {",
		"    self.value",
		"  }",
		"}",
		"",
		"struct Holder T {",
		"  value: T",
		"}",
		"",
		"methods Holder T {",
		"  fn from_reader(reader: Reader T) -> Holder T {",
		"    Holder { value: reader.read() }",
		"  }",
		"}",
		"",
		"fn main() -> void {",
		"  reader := IntReader { value: 9 }",
		"  holder := Holder.from_reader(reader)",
		"  print(holder.value)",
		"}",
		"",
	}, "\n"), Options{
		PackageName: "main",
		EmitMain:    true,
	})
	if got := strings.TrimSpace(stdout); got != "9" {
		t.Fatalf("expected output 9, got %q", got)
	}
}
