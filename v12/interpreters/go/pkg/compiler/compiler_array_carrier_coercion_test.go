package compiler

import (
	"regexp"
	"strings"
	"testing"
)

func TestCompilerStaticArrayCarrierCoercionMonoToGenericStaysDirect(t *testing.T) {
	gen := newGenerator(Options{PackageName: "demo"})
	gen.ensureBuiltinArrayStruct()
	ctx := newArrayRestBindingTestContext()
	ctx.returnType = "runtime.Value"

	lines, converted, ok := gen.coerceStaticArrayCarrierLines(ctx, "source", "*__able_array_i32", "*Array")
	if !ok {
		t.Fatalf("expected mono-to-generic array carrier coercion to compile, got reason %q", ctx.reason)
	}
	if converted == "" {
		t.Fatalf("expected mono-to-generic array carrier coercion to return a converted expression")
	}

	joined := strings.Join(lines, "\n")
	for _, fragment := range []string{
		"make([]runtime.Value,",
		"bridge.ToInt(int64(",
		"&Array{Storage_handle:",
		"__able_struct_Array_sync(",
	} {
		if !strings.Contains(joined, fragment) {
			t.Fatalf("expected mono-to-generic coercion to contain %q:\n%s", fragment, joined)
		}
	}
	for _, fragment := range []string{
		"__able_array_i32_to(__able_runtime,",
		"__able_struct_Array_from(",
		"runtime.ArrayValue",
	} {
		if strings.Contains(joined, fragment) {
			t.Fatalf("expected mono-to-generic coercion to avoid %q:\n%s", fragment, joined)
		}
	}
	if !regexp.MustCompile(`if __able_tmp_\d+ != nil \{`).MatchString(joined) {
		t.Fatalf("expected mono-to-generic coercion to preserve nil guarding:\n%s", joined)
	}
}

func TestCompilerStaticArrayCarrierCoercionGenericToMonoStaysDirect(t *testing.T) {
	gen := newGenerator(Options{PackageName: "demo"})
	gen.ensureBuiltinArrayStruct()
	ctx := newArrayRestBindingTestContext()
	ctx.returnType = "runtime.Value"

	lines, converted, ok := gen.coerceStaticArrayCarrierLines(ctx, "source", "*Array", "*__able_array_i32")
	if !ok {
		t.Fatalf("expected generic-to-mono array carrier coercion to compile, got reason %q", ctx.reason)
	}
	if converted == "" {
		t.Fatalf("expected generic-to-mono array carrier coercion to return a converted expression")
	}

	joined := strings.Join(lines, "\n")
	for _, fragment := range []string{
		"make([]int32,",
		"bridge.AsInt(",
		"&__able_array_i32{Storage_handle:",
		"__able_array_i32_sync(",
	} {
		if !strings.Contains(joined, fragment) {
			t.Fatalf("expected generic-to-mono coercion to contain %q:\n%s", fragment, joined)
		}
	}
	for _, fragment := range []string{
		"__able_struct_Array_to(",
		"__able_array_i32_from(",
		"runtime.ArrayValue",
	} {
		if strings.Contains(joined, fragment) {
			t.Fatalf("expected generic-to-mono coercion to avoid %q:\n%s", fragment, joined)
		}
	}
	if !regexp.MustCompile(`if __able_tmp_\d+ != nil \{`).MatchString(joined) {
		t.Fatalf("expected generic-to-mono coercion to preserve nil guarding:\n%s", joined)
	}
}

func TestCompilerStaticArrayCarrierCoercionRuntimeToGenericStaysDirect(t *testing.T) {
	gen := newGenerator(Options{PackageName: "demo"})
	gen.ensureBuiltinArrayStruct()
	ctx := newArrayRestBindingTestContext()
	ctx.returnType = "runtime.Value"

	lines, converted, ok := gen.staticArrayFromRuntimeLines(ctx, "runtimeValue", "*Array")
	if !ok {
		t.Fatalf("expected runtime-to-generic array carrier coercion to compile, got reason %q", ctx.reason)
	}
	if converted == "" {
		t.Fatalf("expected runtime-to-generic array carrier coercion to return a converted expression")
	}

	joined := strings.Join(lines, "\n")
	for _, fragment := range []string{
		"__able_unwrap_interface(runtimeValue)",
		"__able_runtime_array_value(",
		"runtime.ArrayStoreState(raw.Handle)",
		"make([]runtime.Value, len(raw.Elements), cap(raw.Elements))",
		"&Array{Storage_handle: raw.Handle, Elements:",
		"__able_struct_Array_sync(",
	} {
		if !strings.Contains(joined, fragment) {
			t.Fatalf("expected runtime-to-generic coercion to contain %q:\n%s", fragment, joined)
		}
	}
	for _, fragment := range []string{
		"__able_struct_Array_from(",
		"*runtime.StructInstanceValue",
		"inst.Fields[\"storage_handle\"]",
	} {
		if strings.Contains(joined, fragment) {
			t.Fatalf("expected runtime-to-generic coercion to avoid %q:\n%s", fragment, joined)
		}
	}
	if !strings.Contains(joined, "runtime.NilValue") {
		t.Fatalf("expected runtime-to-generic coercion to preserve NilValue handling:\n%s", joined)
	}
}
