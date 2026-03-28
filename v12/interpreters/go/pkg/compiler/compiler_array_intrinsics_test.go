package compiler

import (
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"able/interpreter-go/pkg/driver"
	"able/interpreter-go/pkg/interpreter"
)

func findCompiledFunction(result *Result, funcName string) (string, bool) {
	compiledSrc := string(result.Files["compiled.go"])
	pattern := "func " + funcName + "("
	idx := strings.Index(compiledSrc, pattern)
	if idx == -1 {
		return "", false
	}
	endMarkers := []string{"\n}\n\nfunc ", "\n}\n\nvar ", "\n}\n\nconst "}
	endIdx := len(compiledSrc)
	for _, marker := range endMarkers {
		if pos := strings.Index(compiledSrc[idx:], marker); pos != -1 && idx+pos < endIdx {
			endIdx = idx + pos + 3
		}
	}
	return compiledSrc[idx:endIdx], true
}

func findCompiledDeclByPrefix(result *Result, prefix string) (string, bool) {
	compiledSrc := string(result.Files["compiled.go"])
	idx := strings.Index(compiledSrc, prefix)
	if idx == -1 {
		return "", false
	}
	endMarkers := []string{"\n}\n\nfunc ", "\n}\n\ntype ", "\n}\n\nvar ", "\n}\n\nconst "}
	endIdx := len(compiledSrc)
	for _, marker := range endMarkers {
		if pos := strings.Index(compiledSrc[idx:], marker); pos != -1 && idx+pos < endIdx {
			endIdx = idx + pos + 3
		}
	}
	return compiledSrc[idx:endIdx], true
}

func compileExecFixtureResult(t *testing.T, rel string) *Result {
	t.Helper()
	dir := filepath.Join(repositoryRoot(), "v12", "fixtures", "exec", filepath.FromSlash(rel))
	manifest, err := interpreter.LoadFixtureManifest(dir)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	entry := manifest.Entry
	if entry == "" {
		entry = "main.able"
	}
	entryPath := filepath.Join(dir, entry)
	searchPaths, err := buildExecSearchPaths(entryPath, dir, manifest)
	if err != nil {
		t.Fatalf("exec search paths: %v", err)
	}
	loader, err := driver.NewLoader(searchPaths)
	if err != nil {
		t.Fatalf("loader init: %v", err)
	}
	t.Cleanup(func() { loader.Close() })
	program, err := loader.Load(entryPath)
	if err != nil {
		t.Fatalf("load program: %v", err)
	}
	result, err := New(Options{
		PackageName:        "main",
		RequireNoFallbacks: requireNoFallbacksForFixtureGates(t),
	}).Compile(program)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	return result
}

func TestCompilerWhileLoopFastPath(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn main() -> i32 {",
		"  i := 0",
		"  while i < 10 {",
		"    i = i + 1",
		"  }",
		"  i",
		"}",
		"",
	}, "\n"))

	fnBody, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	if !strings.Contains(fnBody, "for {") {
		t.Fatalf("expected while-loop fast path to lower into a direct for-loop")
	}
	if strings.Contains(fnBody, "__able_break") || strings.Contains(fnBody, "__able_continue_signal") {
		t.Fatalf("expected while-loop fast path to avoid panic/defer break/continue scaffolding")
	}
}

func TestCompilerCountedLoopFastPath(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn bump(n: i32) -> i32 {",
		"  i := 0",
		"  loop {",
		"    if i >= n { break }",
		"    i = i + 1",
		"  }",
		"  i",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_bump")
	if !ok {
		t.Fatalf("could not find compiled bump function")
	}
	for _, fragment := range []string{
		"for i < n {",
		"i++",
	} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("expected counted-loop fast path to contain %q:\n%s", fragment, body)
		}
	}
	for _, fragment := range []string{
		"for {",
		"if i >= n {",
		"__able_checked_add_signed(",
		"__able_break",
		"__able_continue_signal",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected counted-loop fast path to avoid %q:\n%s", fragment, body)
		}
	}
}

func TestCompilerInlineCheckedIntegerAddSubStayStatic(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn sum_diff(a: i32, b: i32) -> i32 {",
		"  (a + b) - (a - b)",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_sum_diff")
	if !ok {
		t.Fatalf("could not find compiled sum_diff function")
	}
	for _, fragment := range []string{
		"int64(__able_tmp_0) + int64(__able_tmp_1)",
		"int64(__able_tmp_4) - int64(__able_tmp_5)",
		"__able_raise_overflow(",
	} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("expected inline checked integer add/sub lowering to contain %q:\n%s", fragment, body)
		}
	}
	for _, fragment := range []string{
		"__able_checked_add_signed(",
		"__able_checked_sub_signed(",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected inline checked integer add/sub lowering to avoid %q:\n%s", fragment, body)
		}
	}
}

func TestCompilerInlineCheckedSignedSubWithNonNegativeOperandsElidesOverflowBranch(t *testing.T) {
	source := strings.Join([]string{
		"package main",
		"",
		"fn diff() -> i32 {",
		"  a: i32 = 7",
		"  b: i32 = 3",
		"  a - b",
		"}",
	}, "\n")

	result := compileNoFallbackExecSourceWithOptions(t, "ablec-inline-sub-proof", source, Options{
		PackageName: "demo",
	})

	body, ok := findCompiledFunction(result, "__able_compiled_fn_diff")
	if !ok {
		t.Fatalf("could not find compiled diff function")
	}
	if strings.Contains(body, "__able_checked_sub_signed(") {
		t.Fatalf("expected proven non-negative subtraction to avoid helper call:\n%s", body)
	}
	if strings.Contains(body, "__able_raise_overflow(") {
		t.Fatalf("expected proven non-negative subtraction to elide overflow branch:\n%s", body)
	}
	for _, fragment := range []string{
		"__able_tmp_0 := a",
		"__able_tmp_1 := b",
		" := __able_tmp_0 - __able_tmp_1",
	} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("expected proven non-negative subtraction to contain %q:\n%s", fragment, body)
		}
	}
}

func TestCompilerInlineCheckedSignedAddWithCallsiteUpperBoundElidesOverflowBranch(t *testing.T) {
	source := strings.Join([]string{
		"package demo",
		"",
		"fn sum(limit: i32) -> i32 {",
		"  i := 0",
		"  out := 0",
		"  loop {",
		"    if i >= limit { break }",
		"    out = i + i",
		"    i = i + 1",
		"  }",
		"  out",
		"}",
		"",
		"fn main() -> i32 {",
		"  sum(300)",
		"}",
	}, "\n")

	result := compileNoFallbackExecSourceWithOptions(t, "ablec-inline-add-proof", source, Options{
		PackageName: "demo",
	})

	body, ok := findCompiledFunction(result, "__able_compiled_fn_sum")
	if !ok {
		t.Fatalf("could not find compiled sum function")
	}
	if strings.Contains(body, "__able_checked_add_signed(") {
		t.Fatalf("expected proven bounded addition to avoid helper call:\n%s", body)
	}
	if strings.Contains(body, "int64(") {
		t.Fatalf("expected proven bounded addition to avoid widened checked-add lowering:\n%s", body)
	}
	if !strings.Contains(body, " = __able_tmp_") || !strings.Contains(body, " + __able_tmp_") {
		t.Fatalf("expected proven bounded addition to remain direct in compiled body:\n%s", body)
	}
}

func TestCompilerNativeIntegerNarrowingCastStaysStatic(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn cast_byte(value: u16) -> u8 {",
		"  value as u8",
		"}",
		"",
	}, "\n"))

	body, ok := findCompiledFunction(result, "__able_compiled_fn_cast_byte")
	if !ok {
		t.Fatalf("could not find compiled cast_byte function")
	}
	if strings.Contains(body, "__able_cast(") {
		t.Fatalf("expected fixed-width integer narrowing cast to stay native:\n%s", body)
	}
	if !strings.Contains(body, "uint8(") {
		t.Fatalf("expected fixed-width integer narrowing cast to lower directly to a Go uint8 conversion:\n%s", body)
	}
}

func TestCompilerArrayStructKeepsSpecFieldsAndNativeStorage(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn main() -> i32 {",
		"  arr := [1, 2]",
		"  arr.len()",
		"}",
		"",
	}, "\n"))

	compiledSrc := string(result.Files["compiled.go"])
	for _, fragment := range []string{
		"type Array struct {",
		"Length         int32",
		"Capacity       int32",
		"Storage_handle int64",
		"Elements       []runtime.Value",
		"func __able_struct_Array_sync(value *Array) {",
		"&Array{Length: int32(2), Capacity: int32(2), Storage_handle: int64(0), Elements: []runtime.Value{",
	} {
		if !strings.Contains(compiledSrc, fragment) {
			t.Fatalf("expected compiled array lowering to contain %q", fragment)
		}
	}
}

func TestCompilerArrayWithCapacityInGenericReturnContextStaysNative(t *testing.T) {
	result := compileNoFallbackExecSourceWithOptions(t, "ablec-array-with-capacity-generic-return", strings.Join([]string{
		"package demo",
		"",
		"import able.kernel.{Array}",
		"",
		"fn build<T>(count: i32, value: T) -> Array T {",
		"  result := Array.with_capacity(count)",
		"  result.push(value)",
		"  result",
		"}",
		"",
		"fn main() -> i32 {",
		"  values: Array i32 = build(1, 7)",
		"  values.read_slot(0)",
		"}",
		"",
	}, "\n"), Options{
		PackageName:            "main",
		ExperimentalMonoArrays: true,
	})

	body, ok := findCompiledFunction(result, "__able_compiled_fn_build_spec")
	if !ok {
		t.Fatalf("could not find compiled build specialization")
	}
	for _, fragment := range []string{
		"&__able_array_i32{Elements: make([]int32, 0,",
		"__able_tmp_4.Elements = append(__able_tmp_4.Elements, __able_tmp_5)",
	} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("expected generic return-context Array.with_capacity lowering to contain %q:\n%s", fragment, body)
		}
	}
	for _, fragment := range []string{
		"__able_compiled_method_Array_with_capacity_spec(",
		"__able_compiled_method_Array_with_capacity(",
		"__able_method_call_node(",
		"runtime.Value",
	} {
		if strings.Contains(body, fragment) {
			t.Fatalf("expected generic return-context Array.with_capacity lowering to avoid %q:\n%s", fragment, body)
		}
	}
}

func TestCompilerArrayWithCapacityAvoidsAmbientGenericMisinference(t *testing.T) {
	result := compileNoFallbackExecSource(t, "ablec-array-with-capacity-ambient-generic", strings.Join([]string{
		"package demo",
		"",
		"import able.kernel.{Array}",
		"",
		"struct SlotEmpty {}",
		"struct SlotValue T { value: T }",
		"union Slot T = SlotEmpty | SlotValue T",
		"struct Node T { slots: Array (Slot T) }",
		"",
		"fn clone<T>(node: Node T) -> Node T {",
		"  cloned := Array.with_capacity(2)",
		"  idx := 0",
		"  loop {",
		"    if idx >= 2 { break }",
		"    cloned.write_slot(idx, node.slots.read_slot(idx))",
		"    idx = idx + 1",
		"  }",
		"  Node { slots: cloned }",
		"}",
		"",
		"fn main() -> i32 {",
		"  slots: Array (Slot i32) = Array.with_capacity(2)",
		"  slots.write_slot(0, SlotValue { value: 7 })",
		"  copied := clone(Node { slots: slots })",
		"  copied.slots.len()",
		"}",
		"",
	}, "\n"))

	compiledSrc := string(result.Files["compiled.go"])
	if !strings.Contains(compiledSrc, "&__able_array_union__SlotEmpty_or__SlotValue_i32{Elements: make([]__able_union__SlotEmpty_or__SlotValue_i32, 0,") {
		t.Fatalf("expected clone specialization to keep the nested union array carrier:\n%s", compiledSrc)
	}
	for _, fragment := range []string{
		"&__able_array_i32{Elements: make([]int32, 0,",
		"__able_array_i32_to(__able_runtime, cloned)",
	} {
		if strings.Contains(compiledSrc, fragment) {
			t.Fatalf("expected clone specialization to avoid ambient generic Array<T> misinference %q:\n%s", fragment, compiledSrc)
		}
	}
}

func TestCompilerArrayMutationsSyncMetadata(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn main() -> i32 {",
		"  arr := [1, 2]",
		"  arr.push(3)",
		"  arr.write_slot(4, 9)",
		"  arr[0] = 7",
		"  arr.clear()",
		"  arr.capacity()",
		"}",
		"",
	}, "\n"))

	mainBody, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	if count := strings.Count(mainBody, "__able_struct_Array_sync("); count < 4 {
		t.Fatalf("expected static array mutations to sync metadata, found %d sync calls in main", count)
	}
	for _, fragment := range []string{
		"var arr *Array =",
		"append(__able_tmp_",
		".Elements[__able_tmp_",
		".Elements = ",
		".Elements[:0]",
	} {
		if !strings.Contains(mainBody, fragment) {
			t.Fatalf("expected static array lowering to contain %q:\n%s", fragment, mainBody)
		}
	}
	for _, pattern := range []*regexp.Regexp{
		regexp.MustCompile(`__able_index_error\(__able_tmp_\d+, __able_tmp_\d+\)`),
		regexp.MustCompile(`__able_tmp_\d+\.Elements\[__able_tmp_\d+\] = __able_tmp_\d+`),
	} {
		if !pattern.MatchString(mainBody) {
			t.Fatalf("expected static array lowering to match %q:\n%s", pattern.String(), mainBody)
		}
	}
	if strings.Contains(mainBody, "__able_method_call_node(") || strings.Contains(mainBody, "__able_index_set(") {
		t.Fatalf("expected static array lowering to avoid dynamic method/index helpers")
	}
}

func TestCompilerGenericArrayPushKeepsReceiverIdentity(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn main() -> i32 {",
		"  arr := Array.new()",
		"  arr.push(1)",
		"  arr.push(2)",
		"  arr.len()",
		"}",
		"",
	}, "\n"))

	mainBody, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	for _, fragment := range []string{
		"var arr *Array =",
		"append(__able_tmp_",
		"__able_struct_Array_sync(",
	} {
		if !strings.Contains(mainBody, fragment) {
			t.Fatalf("expected generic Array.push lowering to contain %q:\n%s", fragment, mainBody)
		}
	}
	for _, fragment := range []string{
		"__able_array_i32_from(",
		"__able_method_call_node(",
	} {
		if strings.Contains(mainBody, fragment) {
			t.Fatalf("expected generic Array.push lowering to avoid %q:\n%s", fragment, mainBody)
		}
	}
}

func TestCompilerGenericArrayPushExecutes(t *testing.T) {
	source := strings.Join([]string{
		"package main",
		"",
		"fn main() -> void {",
		"  arr := Array.new()",
		"  arr.push(1)",
		"  arr.push(2)",
		"  print(arr.len())",
		"}",
		"",
	}, "\n")

	stdout := compileAndRunExecSourceWithOptions(t, "ablec-generic-array-push-", source, Options{
		PackageName: "main",
		EmitMain:    true,
	})
	if strings.TrimSpace(stdout) != "2" {
		t.Fatalf("expected compiled generic Array.push program to print 2, got %q", stdout)
	}
}

func TestCompilerArrayWrapperUsesExplicitArrayBoundaryConverters(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn cloneish(arr: Array i32) -> Array i32 {",
		"  arr.push(3)",
		"  arr",
		"}",
		"",
	}, "\n"))

	compiledSrc := string(result.Files["compiled.go"])
	if !strings.Contains(compiledSrc, "func __able_compiled_fn_cloneish(arr *Array) (*Array, *__ableControl)") {
		t.Fatalf("expected cloneish to keep a native *Array signature")
	}

	wrapBody, ok := findCompiledFunction(result, "__able_wrap_fn_cloneish")
	if !ok {
		t.Fatalf("could not find wrapper for cloneish")
	}
	if !strings.Contains(wrapBody, "__able_struct_Array_from(arg0Value)") {
		t.Fatalf("expected wrapper arg conversion to use explicit Array_from:\n%s", wrapBody)
	}
	if !strings.Contains(wrapBody, "return __able_struct_Array_to(rt, compiledResult)") {
		t.Fatalf("expected wrapper return to use explicit Array_to:\n%s", wrapBody)
	}
	if strings.Contains(wrapBody, "__able_any_to_value(compiledResult)") {
		t.Fatalf("wrapper should not route Array return through __able_any_to_value:\n%s", wrapBody)
	}
}

func TestCompilerMatchArrayRestBindingStaysNative(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn main() -> i32 {",
		"  arr := [1, 2, 3, 4]",
		"  arr match {",
		"    case [1, 2, ...tail] => tail[0] as i32,",
		"    case _ => 0,",
		"  }",
		"}",
		"",
	}, "\n"))

	mainBody, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	for _, fragment := range []string{
		"var tail *Array",
		"__able_struct_Array_sync(",
	} {
		if !strings.Contains(mainBody, fragment) {
			t.Fatalf("expected native array rest lowering to contain %q:\n%s", fragment, mainBody)
		}
	}
	for _, fragment := range []string{
		"&runtime.ArrayValue{Elements: append([]runtime.Value(nil),",
		"__able_array_values(",
		"var tail runtime.Value =",
	} {
		if strings.Contains(mainBody, fragment) {
			t.Fatalf("expected native array rest lowering to avoid %q:\n%s", fragment, mainBody)
		}
	}
}

func TestCompilerPatternAssignmentArrayRestBindingStaysNative(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn main() -> i32 {",
		"  arr := [1, 2, 3, 4]",
		"  [1, 2, ...tail] := arr",
		"  tail[0] as i32",
		"}",
		"",
	}, "\n"))

	mainBody, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	for _, fragment := range []string{
		"var tail *Array",
		"__able_struct_Array_sync(",
	} {
		if !strings.Contains(mainBody, fragment) {
			t.Fatalf("expected native pattern assignment rest lowering to contain %q:\n%s", fragment, mainBody)
		}
	}
	for _, fragment := range []string{
		"&runtime.ArrayValue{Elements: append([]runtime.Value(nil),",
		"__able_array_values(",
		"var tail runtime.Value =",
	} {
		if strings.Contains(mainBody, fragment) {
			t.Fatalf("expected native pattern assignment rest lowering to avoid %q:\n%s", fragment, mainBody)
		}
	}
}

func TestCompilerArrayHelperFixtureNullableIntrinsicsStayNative(t *testing.T) {
	result := compileExecFixtureResult(t, "06_12_02_stdlib_array_helpers")

	mainBody, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	for _, fragment := range []string{
		"var popped *int32 =",
		"__able_nullable_i32_from_value(",
	} {
		if !strings.Contains(mainBody, fragment) {
			t.Fatalf("expected stdlib array-helper lowering to contain %q:\n%s", fragment, mainBody)
		}
	}
	for _, fragment := range []string{
		"var popped runtime.Value =",
		"func() runtime.Value {",
	} {
		if strings.Contains(mainBody, fragment) {
			t.Fatalf("expected stdlib array-helper lowering to avoid %q:\n%s", fragment, mainBody)
		}
	}
}

func TestCompilerArrayBoundaryHelpersOnlyUseArrayStoreAtExplicitHandleEdges(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn main() -> i32 {",
		"  arr := [1, 2, 3]",
		"  arr.len()",
		"}",
		"",
	}, "\n"))

	arrayFrom, ok := findCompiledFunction(result, "__able_struct_Array_from")
	if !ok {
		t.Fatalf("could not find __able_struct_Array_from")
	}
	if strings.Contains(arrayFrom, "runtime.ArrayStoreEnsure(raw, len(raw.Elements))") {
		t.Fatalf("Array_from should not normalize runtime ArrayValue via ArrayStoreEnsure anymore:\n%s", arrayFrom)
	}
	if !strings.Contains(arrayFrom, "state, err := runtime.ArrayStoreState(raw.Handle)") {
		t.Fatalf("Array_from should read existing handle state directly:\n%s", arrayFrom)
	}

	arrayRuntimeValue, ok := findCompiledFunction(result, "__able_struct_Array_runtime_value")
	if !ok {
		t.Fatalf("could not find __able_struct_Array_runtime_value")
	}
	for _, fragment := range []string{
		"if preferredHandle == 0 {",
		"return &runtime.ArrayValue{Elements: elems}, nil",
		"runtime.ArrayStoreEnsureHandle(preferredHandle, len(elems), cap(elems))",
	} {
		if !strings.Contains(arrayRuntimeValue, fragment) {
			t.Fatalf("expected Array runtime-value helper to contain %q:\n%s", fragment, arrayRuntimeValue)
		}
	}

	arrayTo, ok := findCompiledFunction(result, "__able_struct_Array_to")
	if !ok {
		t.Fatalf("could not find __able_struct_Array_to")
	}
	if !strings.Contains(arrayTo, "arr, err := __able_struct_Array_runtime_value(value, value.Storage_handle)") {
		t.Fatalf("Array_to should route through the shared runtime-value helper:\n%s", arrayTo)
	}
	for _, fragment := range []string{
		"runtime.ArrayStoreEnsure(arr, capHint)",
		"value.Storage_handle = arr.Handle",
		"value.Elements = arr.Elements",
	} {
		if strings.Contains(arrayTo, fragment) {
			t.Fatalf("Array_to should avoid legacy in-place ArrayStore sync fragment %q:\n%s", fragment, arrayTo)
		}
	}

	arrayApply, ok := findCompiledFunction(result, "__able_struct_Array_apply")
	if !ok {
		t.Fatalf("could not find __able_struct_Array_apply")
	}
	for _, fragment := range []string{
		"preferredHandle := raw.Handle",
		"preferredHandle = runtime.ArrayStoreNewWithCapacity(__able_struct_Array_capacity_hint(value))",
		"inst.Fields[\"storage_handle\"] = bridge.ToInt(arr.Handle, runtime.IntegerI64)",
	} {
		if !strings.Contains(arrayApply, fragment) {
			t.Fatalf("expected Array_apply to contain %q:\n%s", fragment, arrayApply)
		}
	}
	for _, fragment := range []string{
		"_, _, _ = runtime.ArrayStoreEnsure(raw, len(value.Elements))",
		"converted, err := __able_struct_Array_to(rt, value)",
	} {
		if strings.Contains(arrayApply, fragment) {
			t.Fatalf("Array_apply should avoid legacy boundary fragment %q:\n%s", fragment, arrayApply)
		}
	}
}
