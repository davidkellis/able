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
		"lengthVal, ok := inst.Fields[\"length\"]",
		"capacityVal, ok := inst.Fields[\"capacity\"]",
		"handleVal, ok := inst.Fields[\"storage_handle\"]",
		"runtime.ArrayStoreState(sourceHandle)",
	} {
		if !strings.Contains(sharedHelper, fragment) {
			t.Fatalf("expected shared Array struct-instance helper to contain %q:\n%s", fragment, sharedHelper)
		}
	}
}
