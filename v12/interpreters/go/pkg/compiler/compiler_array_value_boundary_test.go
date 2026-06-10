package compiler

import (
	"strings"
	"testing"
)

func TestCompilerGenericArrayWrapperRawBoundaryStaysDirect(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn cloneish(values: Array) -> Array {",
		"  values",
		"}",
		"",
	}, "\n"))

	compiledSrc := string(result.Files["compiled.go"])
	if !strings.Contains(compiledSrc, "func __able_compiled_fn_cloneish(values *Array) (*Array, *__ableControl)") {
		t.Fatalf("expected cloneish to keep a generic Array carrier signature:\n%s", compiledSrc)
	}

	wrapBody, ok := findCompiledFunction(result, "__able_wrap_fn_cloneish")
	if !ok {
		t.Fatalf("could not find wrapper for cloneish")
	}
	for _, fragment := range []string{
		"values_current := __able_unwrap_interface(arg0Value)",
		"__able_runtime_array_value(values_current)",
		"runtime.ArrayStoreState(raw.Handle)",
		"&Array{Storage_handle: raw.Handle, Elements:",
		"__able_struct_Array_sync(values)",
	} {
		if !strings.Contains(wrapBody, fragment) {
			t.Fatalf("expected generic Array wrapper boundary to contain %q:\n%s", fragment, wrapBody)
		}
	}
	for _, fragment := range []string{
		"__able_struct_Array_from(",
	} {
		if strings.Contains(wrapBody, fragment) {
			t.Fatalf("expected generic Array wrapper to avoid direct helper detour %q:\n%s", fragment, wrapBody)
		}
	}
}

func TestCompilerExpectRuntimeValueExprLinesGenericArrayRawBoundaryStaysDirect(t *testing.T) {
	gen := newGenerator(Options{PackageName: "demo"})
	gen.ensureBuiltinArrayStruct()
	ctx := newArrayRestBindingTestContext()
	ctx.returnType = "runtime.Value"

	lines, converted, ok := gen.expectRuntimeValueExprLines(ctx, "runtimeValue", "*Array")
	if !ok {
		t.Fatalf("expected generic Array runtime-value conversion lines to compile, got reason %q", ctx.reason)
	}
	if converted == "" {
		t.Fatalf("expected generic Array runtime-value conversion lines to return a converted expression")
	}

	joined := strings.Join(lines, "\n")
	for _, fragment := range []string{
		"__able_runtime_array_value(",
		"runtime.ArrayStoreState(raw.Handle)",
		"&Array{Storage_handle: raw.Handle, Elements:",
		"__able_struct_Array_sync(",
		"__able_control_from_error(",
	} {
		if !strings.Contains(joined, fragment) {
			t.Fatalf("expected generic Array runtime-value conversion lines to contain %q:\n%s", fragment, joined)
		}
	}
	if strings.Contains(joined, "__able_struct_Array_from(") {
		t.Fatalf("expected generic Array runtime-value conversion lines to avoid Array_from helper detours:\n%s", joined)
	}
}

func TestCompilerExpectRuntimeValueExprGenericArrayRawBoundaryStaysDirect(t *testing.T) {
	gen := newGenerator(Options{PackageName: "demo"})
	gen.ensureBuiltinArrayStruct()

	expr, ok := gen.expectRuntimeValueExpr("runtimeValue", "*Array")
	if !ok {
		t.Fatalf("expected generic Array panic conversion expression to compile")
	}
	for _, fragment := range []string{
		"__able_unwrap_interface(value)",
		"__able_runtime_array_value(result_current)",
		"runtime.ArrayStoreState(raw.Handle)",
		"&Array{Storage_handle: raw.Handle, Elements:",
		"__able_struct_Array_sync(result)",
	} {
		if !strings.Contains(expr, fragment) {
			t.Fatalf("expected generic Array panic conversion expression to contain %q:\n%s", fragment, expr)
		}
	}
	if strings.Contains(expr, "__able_struct_Array_from(") {
		t.Fatalf("expected generic Array panic conversion expression to avoid Array_from helper detours:\n%s", expr)
	}
}

func TestCompilerArrayFromHelperUsesSharedStructInstanceBoundaryHelper(t *testing.T) {
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
	if !strings.Contains(arrayFrom, "__able_array_struct_instance_state(inst)") {
		t.Fatalf("expected Array_from struct-instance fallback to use shared helper:\n%s", arrayFrom)
	}
	for _, fragment := range []string{
		"lengthVal, ok := inst.Fields[\"length\"]",
		"capacityVal, ok := inst.Fields[\"capacity\"]",
		"handleVal, ok := inst.Fields[\"storage_handle\"]",
	} {
		if strings.Contains(arrayFrom, fragment) {
			t.Fatalf("expected Array_from to avoid inline struct-instance field plumbing %q:\n%s", fragment, arrayFrom)
		}
	}

	sharedHelper, ok := findCompiledFunction(result, "__able_array_struct_instance_state")
	if !ok {
		t.Fatalf("could not find __able_array_struct_instance_state")
	}
	for _, fragment := range []string{
		"__able_struct_named_field_value(inst, \"length\")",
		"__able_struct_named_field_value(inst, \"capacity\")",
		"__able_struct_named_field_value(inst, \"storage_handle\")",
		"runtime.ArrayStoreState(sourceHandle)",
	} {
		if !strings.Contains(sharedHelper, fragment) {
			t.Fatalf("expected shared Array struct-instance helper to contain %q:\n%s", fragment, sharedHelper)
		}
	}
}

func TestCompilerNamedStructBoundaryHelperSupportsPositionalStorage(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"struct Record { value: i32 }",
		"",
		"fn main() -> i32 { Record { value: 1 }.value }",
		"",
	}, "\n"))

	helper, ok := findCompiledFunction(result, "__able_struct_named_field_value")
	if !ok {
		t.Fatal("could not find named-struct boundary helper")
	}
	for _, fragment := range []string{
		"inst.Fields[name]",
		"inst.Positional == nil",
		"ast.StructKindNamed",
		"inst.Definition.NamedFieldIndices[name]",
		"return inst.Positional[idx], true",
	} {
		if !strings.Contains(helper, fragment) {
			t.Fatalf("expected named-struct boundary helper to contain %q:\n%s", fragment, helper)
		}
	}
}

func TestCompilerRuntimeIndexHelpersUseSharedNamedFieldAccessor(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn main() -> i32 { 1 }",
		"",
	}, "\n"))

	for _, helperName := range []string{"__able_index", "__able_index_set"} {
		helper, ok := findCompiledFunction(result, helperName)
		if !ok {
			t.Fatalf("could not find %s", helperName)
		}
		for _, fieldName := range []string{"storage_handle", "handle"} {
			sharedLookup := `__able_struct_named_field_value(inst, "` + fieldName + `")`
			if !strings.Contains(helper, sharedLookup) {
				t.Fatalf("expected %s to use shared lookup %q:\n%s", helperName, sharedLookup, helper)
			}
			directLookup := `inst.Fields["` + fieldName + `"]`
			if strings.Contains(helper, directLookup) {
				t.Fatalf("expected %s to avoid representation-specific lookup %q:\n%s", helperName, directLookup, helper)
			}
		}
	}
}

func TestCompilerRatioRuntimeHelperUsesSharedNamedFieldAccessor(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn main() -> i32 { 1 }",
		"",
	}, "\n"))

	helper, ok := findCompiledFunction(result, "__able_ratio_parts_from_struct")
	if !ok {
		t.Fatal("could not find Ratio runtime conversion helper")
	}
	for _, fieldName := range []string{"num", "den"} {
		sharedLookup := `__able_struct_named_field_value(inst, "` + fieldName + `")`
		if !strings.Contains(helper, sharedLookup) {
			t.Fatalf("expected Ratio conversion to use shared lookup %q:\n%s", sharedLookup, helper)
		}
		directLookup := `inst.Fields["` + fieldName + `"]`
		if strings.Contains(helper, directLookup) {
			t.Fatalf("expected Ratio conversion to avoid representation-specific lookup %q:\n%s", directLookup, helper)
		}
	}
}

func TestCompilerNamedStructToRuntimeUsesSharedPositionalStorage(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"struct Record { value: i32, label: String }",
		"",
		"fn main() -> i32 { Record { value: 1, label: \"one\" }.value }",
		"",
	}, "\n"))

	converter, ok := findCompiledFunction(result, "__able_struct_Record_to_seen")
	if !ok {
		t.Fatal("could not find Record runtime converter")
	}
	for _, fragment := range []string{
		"runtime.NewStructInstancePositionalSized(def, 2, nil)",
		"fields[0] = bridge.ToInt(",
		"fields[1] = bridge.ToString(",
	} {
		if !strings.Contains(converter, fragment) {
			t.Fatalf("expected named-struct runtime converter to contain %q:\n%s", fragment, converter)
		}
	}
	if strings.Contains(converter, "Fields: make(map[string]runtime.Value") {
		t.Fatalf("expected named-struct runtime converter to avoid map-backed field storage:\n%s", converter)
	}
}
