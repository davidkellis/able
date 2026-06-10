package compiler

import (
	"bytes"
	"strings"
	"testing"
)

// Keep the compiler's handle-bearing generic Array carrier on the same alias
// contract as the interpreters: a copied Array passed to and returned from a
// function carries the same storage_handle, not an independently releasable
// view.
func TestCompilerArrayHandleAliasingThroughFunctionReturnExecutes(t *testing.T) {
	stdout := strings.TrimSpace(compileAndRunExecSourceWithOptions(t, "ablec-array-handle-alias-", strings.Join([]string{
		"package demo",
		"",
		"import able.kernel.{Array}",
		"import able.collections.array",
		"",
		"fn extend(values: Array) -> Array {",
		"  alias := values",
		"  alias.push(30)",
		"  values",
		"}",
		"",
		"fn main() -> void {",
		"  values: Array = Array.new()",
		"  values.push(10)",
		"  alias := values",
		"  returned := extend(alias)",
		"  values.push(20)",
		"  print(values.read_slot(0))",
		"  print(` `)",
		"  print(alias.read_slot(1))",
		"  print(` `)",
		"  print(returned.read_slot(2))",
		"}",
		"",
	}, "\n"), Options{
		PackageName:              "main",
		EmitMain:                 true,
		RequireStaticNoFallbacks: true,
	}))
	if strings.Join(strings.Fields(stdout), " ") != "10 30 20" {
		t.Fatalf("compiled Array alias result = %q, want 10 30 20", stdout)
	}
}

func TestCompilerGenericArrayCarrierEmitsDiagnosticLeaseTracking(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn pass(values: Array) -> Array {",
		"  values",
		"}",
		"",
	}, "\n"))

	compiledSrc := string(result.Files["compiled.go"])
	for _, fragment := range []string{
		"arrayStoreLease runtime.ArrayStoreLease",
		"func __able_struct_Array_track_handle(value *Array) error {",
		"func __able_struct_Array_clone(value *Array) (*Array, error) {",
		"func __able_struct_Array_move(target *Array, source *Array) error {",
		"runtime.ArrayStoreMoveLease(&target.arrayStoreLease, &source.arrayStoreLease)",
		"runtime.ArrayStoreLeaseTracksWithCleanup(&value.arrayStoreLease, value.Storage_handle)",
		"runtime.ArrayStoreTrackLeaseOwner(value, &value.arrayStoreLease, value.Storage_handle)",
		"runtime.ArrayStoreTrackArrayValueLease(result, result.Handle)",
		"runtime.ArrayStoreTrackStructInstanceLease(inst, sourceHandle)",
		"runtime.ArrayStoreTrackStructInstanceLease(inst, preferredHandle)",
	} {
		if !strings.Contains(compiledSrc, fragment) {
			t.Fatalf("expected generic Array carrier to contain %q:\n%s", fragment, compiledSrc)
		}
	}
	if count := strings.Count(compiledSrc, "runtime.ArrayStoreTrackStructInstanceLease(inst, sourceHandle)"); count < 2 {
		t.Fatalf("Array struct-state helpers track source handle %d times, want at least two", count)
	}
}

func TestCompilerGenericArrayRuntimeBoundaryLinesTrackConstructedCarrier(t *testing.T) {
	g := &generator{}
	ctx := newCompileContext(g, nil, nil, nil, "", nil)
	ctx.controlMode = compileControlModeErrorOnly
	directLines, _, ok := g.directRuntimeArrayValueToGenericArrayLines(ctx, "runtimeValue")
	if !ok {
		t.Fatal("direct runtime Array conversion did not lower")
	}
	if count := strings.Count(strings.Join(directLines, "\n"), "__able_struct_Array_track_handle("); count != 1 {
		t.Fatalf("direct runtime Array conversion has %d lease tracking calls, want one:\n%s", count, strings.Join(directLines, "\n"))
	}

	boundaryLines := g.runtimeValueToGenericArrayBoundaryLines("converted", "err", "runtimeValue", true)
	if count := strings.Count(strings.Join(boundaryLines, "\n"), "__able_struct_Array_track_handle("); count != 1 {
		t.Fatalf("runtime Array boundary has %d lease tracking calls, want one:\n%s", count, strings.Join(boundaryLines, "\n"))
	}
}

func TestCompilerArrayHandleBorrowPathsTrackCanonicalStructOwner(t *testing.T) {
	g := newGenerator(Options{PackageName: "main", EmitMain: true, EntryPath: "main.able"})

	var runtimeHelpers bytes.Buffer
	g.renderRuntimeArrayHelpers(&runtimeHelpers)
	if !strings.Contains(runtimeHelpers.String(), "runtime.ArrayStoreTrackStructInstanceLease(inst, rawHandle)") {
		t.Fatalf("compiled Array value extraction does not track its canonical struct owner:\n%s", runtimeHelpers.String())
	}

	mainSrc, err := g.renderMain()
	if err != nil {
		t.Fatalf("render main: %v", err)
	}
	if !strings.Contains(string(mainSrc), "runtime.ArrayStoreTrackStructInstanceLease(v, handle)") {
		t.Fatalf("compiled Array formatting does not track its canonical struct owner:\n%s", mainSrc)
	}
}

func TestCompilerArrayCarrierMoveControlLinesUseCanonicalHelper(t *testing.T) {
	g := &generator{}
	ctx := newCompileContext(g, nil, nil, nil, "", nil)
	ctx.controlMode = compileControlModeErrorOnly
	lines, ok := g.appendArrayCarrierMoveControlLines(ctx, "target", "source")
	if !ok {
		t.Fatal("Array carrier move did not lower")
	}
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "__able_struct_Array_move(target, source)") || !strings.Contains(joined, "__able_control_from_error") {
		t.Fatalf("Array carrier move lines do not use canonical control path:\n%s", joined)
	}
}

func TestCompilerArrayReceiverWritebackUsesLeaseMove(t *testing.T) {
	g := newGenerator(Options{})
	g.structs["Array"] = &structInfo{Name: "Array", GoName: "Array"}
	ctx := newCompileContext(g, nil, nil, nil, "", nil)
	ctx.controlMode = compileControlModeErrorOnly
	lines, ok := g.appendStaticNominalReceiverWriteback(ctx, "target", "*Array", "source", "*Array")
	if !ok {
		t.Fatal("Array receiver writeback did not lower")
	}
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "__able_struct_Array_move(target, source)") || strings.Contains(joined, "*target = *source") {
		t.Fatalf("Array receiver writeback did not use lease move:\n%s", joined)
	}
}

func TestCompilerArrayInterfaceWritebackUsesLeaseMove(t *testing.T) {
	g := newGenerator(Options{})
	g.structs["Array"] = &structInfo{Name: "Array", GoName: "Array"}
	info := &nativeInterfaceInfo{
		ApplyRuntimeHelper: "__able_iface_Relay_apply_runtime_value",
		AdapterVersion:     g.nativeInterfaceAdapterVersion,
		Adapters: []*nativeInterfaceAdapter{{
			GoType:      "*Array",
			AdapterType: "__able_iface_Relay_adapter_ptr_Array",
		}},
	}
	var buf bytes.Buffer
	g.renderNativeInterfaceApplyRuntimeHelper(&buf, info)
	applyBody := buf.String()
	if !strings.Contains(applyBody, "__able_struct_Array_move(typed.Value, converted)") {
		t.Fatalf("Array interface writeback did not use lease move:\n%s", applyBody)
	}
	if strings.Contains(applyBody, "*typed.Value = *converted") {
		t.Fatalf("Array interface writeback retained whole-struct copy:\n%s", applyBody)
	}
}

func TestCompilerArrayFunctionalUpdateUsesLeaseClone(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn duplicate(values: Array) -> Array {",
		"  Array { ...values }",
		"}",
		"",
	}, "\n"))
	compiledSrc := string(result.Files["compiled.go"])
	if count := strings.Count(compiledSrc, "__able_struct_Array_clone("); count < 2 {
		t.Fatalf("functional Array update has %d clone references, want helper and call:\n%s", count, compiledSrc)
	}
}
