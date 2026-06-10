package compiler

import (
	"strings"
	"testing"
)

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
		"var tail *__able_array_i32",
	} {
		if !strings.Contains(mainBody, fragment) {
			t.Fatalf("expected native array rest lowering to contain %q:\n%s", fragment, mainBody)
		}
	}
	for _, fragment := range []string{
		"&runtime.ArrayValue{Elements: append([]runtime.Value(nil),",
		"__able_array_values(",
		"var tail runtime.Value =",
		"__able_array_i32_sync(",
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
		"var tail *__able_array_i32",
	} {
		if !strings.Contains(mainBody, fragment) {
			t.Fatalf("expected native pattern assignment rest lowering to contain %q:\n%s", fragment, mainBody)
		}
	}
	for _, fragment := range []string{
		"&runtime.ArrayValue{Elements: append([]runtime.Value(nil),",
		"__able_array_values(",
		"var tail runtime.Value =",
		"__able_array_i32_sync(",
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
		"__able_ptr(__able_tmp_",
	} {
		if !strings.Contains(mainBody, fragment) {
			t.Fatalf("expected stdlib array-helper lowering to contain %q:\n%s", fragment, mainBody)
		}
	}
	for _, fragment := range []string{
		"var popped runtime.Value =",
		"func() runtime.Value {",
		"__able_nullable_i32_from_value(",
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
	if !strings.Contains(arrayFrom, "__able_array_struct_instance_state(inst)") {
		t.Fatalf("Array_from should delegate struct-instance fallback through the shared helper now:\n%s", arrayFrom)
	}
	for _, fragment := range []string{
		"lengthVal, ok := inst.Fields[\"length\"]",
		"capacityVal, ok := inst.Fields[\"capacity\"]",
		"handleVal, ok := inst.Fields[\"storage_handle\"]",
	} {
		if strings.Contains(arrayFrom, fragment) {
			t.Fatalf("Array_from should avoid inline struct-instance field plumbing %q:\n%s", fragment, arrayFrom)
		}
	}

	sharedStructHelper, ok := findCompiledFunction(result, "__able_array_struct_instance_state")
	if !ok {
		t.Fatalf("could not find __able_array_struct_instance_state")
	}
	for _, fragment := range []string{
		"__able_struct_named_field_value(inst, \"length\")",
		"__able_struct_named_field_value(inst, \"capacity\")",
		"__able_struct_named_field_value(inst, \"storage_handle\")",
		"runtime.ArrayStoreState(sourceHandle)",
	} {
		if !strings.Contains(sharedStructHelper, fragment) {
			t.Fatalf("expected shared Array struct-instance helper to contain %q:\n%s", fragment, sharedStructHelper)
		}
	}

	if _, ok := findCompiledFunction(result, "__able_struct_Array_runtime_value"); ok {
		t.Fatalf("expected shared Array runtime-value helper to be removed from generated output")
	}

	arrayTo, ok := findCompiledFunction(result, "__able_struct_Array_to")
	if !ok {
		t.Fatalf("could not find __able_struct_Array_to")
	}
	for _, fragment := range []string{
		"__able_struct_Array_sync(value)",
		"capHint := __able_struct_Array_capacity_hint(value)",
		"elems := __able_struct_Array_clone_elements(value.Elements, capHint)",
		"if value.Storage_handle == 0 {",
		"return &runtime.ArrayValue{Elements: elems}, nil",
		"state, err := runtime.ArrayStoreEnsureHandle(value.Storage_handle, len(elems), cap(elems))",
		"result := &runtime.ArrayValue{Elements: state.Values, Handle: value.Storage_handle}",
		"runtime.ArrayStoreTrackArrayValueLease(result, result.Handle)",
		"return result, nil",
	} {
		if !strings.Contains(arrayTo, fragment) {
			t.Fatalf("expected Array_to to contain %q:\n%s", fragment, arrayTo)
		}
	}
	for _, fragment := range []string{
		"__able_struct_Array_runtime_value(value, value.Storage_handle)",
		"runtime.ArrayStoreEnsure(arr, capHint)",
		"value.Storage_handle = arr.Handle",
		"value.Elements = arr.Elements",
	} {
		if strings.Contains(arrayTo, fragment) {
			t.Fatalf("Array_to should avoid legacy/shared-helper fragment %q:\n%s", fragment, arrayTo)
		}
	}

	arrayApply, ok := findCompiledFunction(result, "__able_struct_Array_apply")
	if !ok {
		t.Fatalf("could not find __able_struct_Array_apply")
	}
	for _, fragment := range []string{
		"preferredHandle := raw.Handle",
		"preferredHandle = runtime.ArrayStoreNewWithCapacity(__able_struct_Array_capacity_hint(value))",
		"state, err := runtime.ArrayStoreEnsureHandle(preferredHandle, len(elems), cap(elems))",
		"inst.Fields[\"storage_handle\"] = bridge.ToInt(preferredHandle, runtime.IntegerI64)",
	} {
		if !strings.Contains(arrayApply, fragment) {
			t.Fatalf("expected Array_apply to contain %q:\n%s", fragment, arrayApply)
		}
	}
	for _, fragment := range []string{
		"_, _, _ = runtime.ArrayStoreEnsure(raw, len(value.Elements))",
		"converted, err := __able_struct_Array_to(rt, value)",
		"arr, err := __able_struct_Array_runtime_value(value, preferredHandle)",
	} {
		if strings.Contains(arrayApply, fragment) {
			t.Fatalf("Array_apply should avoid legacy boundary fragment %q:\n%s", fragment, arrayApply)
		}
	}
}
