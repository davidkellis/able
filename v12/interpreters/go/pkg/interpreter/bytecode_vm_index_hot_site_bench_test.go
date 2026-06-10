package interpreter

import (
	"path/filepath"
	"testing"

	"able/interpreter-go/pkg/driver"
	"able/interpreter-go/pkg/runtime"
)

var (
	benchmarkBytecodeIndexArraySink  *runtime.ArrayValue
	benchmarkBytecodeIndexErrSink    error
	benchmarkBytecodeIndexHandleSink int64
	benchmarkBytecodeIndexIntSink    int
	benchmarkBytecodeIndexOKSink     bool
)

type benchmarkBytecodeIndexSiteFixture struct {
	vm                 *bytecodeVM
	program            *bytecodeProgram
	left               *runtime.ArrayValue
	right              *runtime.ArrayValue
	alias              *runtime.ArrayValue
	globalRevision     uint64
	methodCacheVersion uint64
	baseDirectEntry    bytecodeInlineIndexMethodCacheEntry
}

func newBenchmarkBytecodeIndexSiteFixture(b *testing.B) benchmarkBytecodeIndexSiteFixture {
	b.Helper()
	interp := NewBytecode()
	preloadArrayStdlibForBenchmark(b, interp)
	left := monoCharArrayValueForBenchmark(b, 'a', 'b')
	right := monoCharArrayValueForBenchmark(b, 'x', 'y')
	alias := monoCharArrayValueForBenchmark(b, 'm', 'n')
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	program := &bytecodeProgram{instructions: []bytecodeInstruction{{}, {}}}
	vm.currentProgram = program

	vm.ip = 0
	if _, fastPath, hasMethod, cacheable, err := vm.resolveCachedIndexMethod(program, 0, left, "get", "Index"); err != nil {
		b.Fatalf("resolveCachedIndexMethod(left): %v", err)
	} else if !cacheable || !hasMethod || fastPath != bytecodeIndexMethodFastPathCanonicalArrayGet {
		b.Fatalf("left cache = fastPath %v hasMethod %v cacheable %v, want canonical get", fastPath, hasMethod, cacheable)
	}
	baseDirectEntry := vm.indexMethodDirect[bytecodeIndexMethodDirectIndex(0)]

	vm.ip = 1
	if _, fastPath, hasMethod, cacheable, err := vm.resolveCachedIndexMethod(program, 1, right, "get", "Index"); err != nil {
		b.Fatalf("resolveCachedIndexMethod(right): %v", err)
	} else if !cacheable || !hasMethod || fastPath != bytecodeIndexMethodFastPathCanonicalArrayGet {
		b.Fatalf("right cache = fastPath %v hasMethod %v cacheable %v, want canonical get", fastPath, hasMethod, cacheable)
	}

	vm.currentProgram = program
	vm.setActiveLookupProgram(program)
	globalRevision, methodCacheVersion := vm.bytecodeGlobalAndMethodVersions()
	return benchmarkBytecodeIndexSiteFixture{
		vm:                 vm,
		program:            program,
		left:               left,
		right:              right,
		alias:              alias,
		globalRevision:     globalRevision,
		methodCacheVersion: methodCacheVersion,
		baseDirectEntry:    baseDirectEntry,
	}
}

func preloadArrayStdlibForBenchmark(b *testing.B, interp *Interpreter) {
	b.Helper()
	loader, err := driver.NewLoader([]driver.SearchPath{
		{Path: stdlibRoot, Kind: driver.RootStdlib},
		{Path: kernelRoot, Kind: driver.RootStdlib},
	})
	if err != nil {
		b.Fatalf("loader init failed: %v", err)
	}
	stdlibProgram, err := loader.Load(filepath.Join(stdlibRoot, "collections", "array.able"))
	if err != nil {
		b.Fatalf("load stdlib array failed: %v", err)
	}
	if _, _, _, err := interp.EvaluateProgram(stdlibProgram, ProgramEvaluationOptions{}); err != nil {
		b.Fatalf("evaluate stdlib array failed: %v", err)
	}
}

func monoCharArrayValueForBenchmark(b *testing.B, values ...rune) *runtime.ArrayValue {
	b.Helper()
	handle := runtime.ArrayStoreMonoNewWithCapacityChar(len(values))
	for idx, value := range values {
		if err := runtime.ArrayStoreMonoWriteChar(handle, idx, value); err != nil {
			b.Fatalf("write mono char value %d: %v", idx, err)
		}
	}
	return &runtime.ArrayValue{Handle: handle, TrackedHandle: handle}
}

func BenchmarkBytecodeVMIndexMethodSiteCache(b *testing.B) {
	b.Run("versions", func(b *testing.B) {
		fixture := newBenchmarkBytecodeIndexSiteFixture(b)

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			fixture.globalRevision, fixture.methodCacheVersion = fixture.vm.bytecodeGlobalAndMethodVersions()
		}
	})

	b.Run("canonical_hot_revision", func(b *testing.B) {
		fixture := newBenchmarkBytecodeIndexSiteFixture(b)
		vm := fixture.vm
		vm.ip = 0
		arr, ok := vm.lookupHotArrayIndexSiteWithVersions(
			bytecodeIndexMethodCacheGet,
			fixture.left,
			bytecodeIndexMethodFastPathCanonicalArrayGet,
			true,
			fixture.globalRevision,
			fixture.methodCacheVersion,
		)
		if !ok || arr != fixture.left {
			b.Fatalf("hot lookup = (%p, %v), want left receiver", arr, ok)
		}

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			benchmarkBytecodeIndexArraySink, benchmarkBytecodeIndexOKSink = vm.lookupHotArrayIndexSiteWithVersions(
				bytecodeIndexMethodCacheGet,
				fixture.left,
				bytecodeIndexMethodFastPathCanonicalArrayGet,
				true,
				fixture.globalRevision,
				fixture.methodCacheVersion,
			)
		}
	})

	b.Run("canonical_direct_exact_alternating", func(b *testing.B) {
		fixture := newBenchmarkBytecodeIndexSiteFixture(b)
		vm := fixture.vm
		vm.activeLookup.indexMethodGetEntries = nil
		receivers := [2]*runtime.ArrayValue{fixture.left, fixture.right}

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			vm.ip = idx & 1
			benchmarkBytecodeIndexArraySink, benchmarkBytecodeIndexOKSink = vm.lookupHotArrayIndexSiteWithVersions(
				bytecodeIndexMethodCacheGet,
				receivers[idx&1],
				bytecodeIndexMethodFastPathCanonicalArrayGet,
				true,
				fixture.globalRevision,
				fixture.methodCacheVersion,
			)
		}
	})

	b.Run("direct_same_identity_validate", func(b *testing.B) {
		fixture := newBenchmarkBytecodeIndexSiteFixture(b)
		vm := fixture.vm
		entry := fixture.baseDirectEntry
		if _, ok := vm.lookupInlineArrayIndexSiteReady(&entry, fixture.alias); !ok {
			b.Fatalf("same-identity validation should succeed")
		}

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			entry := fixture.baseDirectEntry
			benchmarkBytecodeIndexHandleSink, benchmarkBytecodeIndexOKSink = vm.lookupInlineArrayIndexSiteReady(&entry, fixture.alias)
		}
	})

	b.Run("direct_compatible_hot_revision", func(b *testing.B) {
		fixture := newBenchmarkBytecodeIndexSiteFixture(b)
		vm := fixture.vm
		vm.ip = 0
		arr, handle, ok := vm.lookupDirectCompatibleHotArrayIndexSiteForArrayWithVersionsReady(
			bytecodeIndexMethodCacheGet,
			fixture.left,
			bytecodeIndexMethodFastPathCanonicalArrayGet,
			fixture.globalRevision,
			fixture.methodCacheVersion,
		)
		if !ok || arr != fixture.left || handle != fixture.left.Handle {
			b.Fatalf("direct-compatible hot lookup = (%p, %d, %v), want left receiver handle %d", arr, handle, ok, fixture.left.Handle)
		}

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			benchmarkBytecodeIndexArraySink, benchmarkBytecodeIndexHandleSink, benchmarkBytecodeIndexOKSink = vm.lookupDirectCompatibleHotArrayIndexSiteForArrayWithVersionsReady(
				bytecodeIndexMethodCacheGet,
				fixture.left,
				bytecodeIndexMethodFastPathCanonicalArrayGet,
				fixture.globalRevision,
				fixture.methodCacheVersion,
			)
		}
	})

	b.Run("direct_compatible_direct_alternating", func(b *testing.B) {
		fixture := newBenchmarkBytecodeIndexSiteFixture(b)
		vm := fixture.vm
		vm.activeLookup.indexMethodGetEntries = nil
		receivers := [2]*runtime.ArrayValue{fixture.left, fixture.right}

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			vm.ip = idx & 1
			benchmarkBytecodeIndexArraySink, benchmarkBytecodeIndexHandleSink, benchmarkBytecodeIndexOKSink = vm.lookupDirectCompatibleHotArrayIndexSiteForArrayWithVersionsReady(
				bytecodeIndexMethodCacheGet,
				receivers[idx&1],
				bytecodeIndexMethodFastPathCanonicalArrayGet,
				fixture.globalRevision,
				fixture.methodCacheVersion,
			)
		}
	})

	b.Run("canonical_per_program_exact", func(b *testing.B) {
		fixture := newBenchmarkBytecodeIndexSiteFixture(b)
		vm := fixture.vm
		vm.ip = 0
		arr, ok := vm.lookupCachedArrayIndexSiteEntry(
			bytecodeIndexMethodCacheGet,
			fixture.left,
			bytecodeIndexMethodFastPathCanonicalArrayGet,
			true,
			fixture.globalRevision,
			fixture.methodCacheVersion,
		)
		if !ok || arr != fixture.left {
			b.Fatalf("per-program lookup = (%p, %v), want left receiver", arr, ok)
		}

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			benchmarkBytecodeIndexArraySink, benchmarkBytecodeIndexOKSink = vm.lookupCachedArrayIndexSiteEntry(
				bytecodeIndexMethodCacheGet,
				fixture.left,
				bytecodeIndexMethodFastPathCanonicalArrayGet,
				true,
				fixture.globalRevision,
				fixture.methodCacheVersion,
			)
		}
	})

	b.Run("canonical_full_per_program_collision", func(b *testing.B) {
		interp := NewBytecode()
		preloadArrayStdlibForBenchmark(b, interp)
		left := monoCharArrayValueForBenchmark(b, 'a', 'b')
		right := monoCharArrayValueForBenchmark(b, 'x', 'y')
		vm := newBytecodeVM(interp, interp.GlobalEnvironment())
		program := &bytecodeProgram{instructions: make([]bytecodeInstruction, bytecodeIndexMethodDirectEntries+1)}
		vm.currentProgram = program

		vm.ip = 0
		if _, fastPath, hasMethod, cacheable, err := vm.resolveCachedIndexMethod(program, 0, left, "get", "Index"); err != nil {
			b.Fatalf("resolveCachedIndexMethod(left): %v", err)
		} else if !cacheable || !hasMethod || fastPath != bytecodeIndexMethodFastPathCanonicalArrayGet {
			b.Fatalf("left cache = fastPath %v hasMethod %v cacheable %v, want canonical get", fastPath, hasMethod, cacheable)
		}

		collisionIP := bytecodeIndexMethodDirectEntries
		vm.ip = collisionIP
		if _, fastPath, hasMethod, cacheable, err := vm.resolveCachedIndexMethod(program, collisionIP, right, "get", "Index"); err != nil {
			b.Fatalf("resolveCachedIndexMethod(right): %v", err)
		} else if !cacheable || !hasMethod || fastPath != bytecodeIndexMethodFastPathCanonicalArrayGet {
			b.Fatalf("right cache = fastPath %v hasMethod %v cacheable %v, want canonical get", fastPath, hasMethod, cacheable)
		}

		vm.setActiveLookupProgram(program)
		globalRevision, methodCacheVersion := vm.bytecodeGlobalAndMethodVersions()
		receivers := [2]*runtime.ArrayValue{left, right}
		ips := [2]int{0, collisionIP}

		b.ReportAllocs()
		b.ResetTimer()
		for idx := 0; idx < b.N; idx++ {
			site := idx & 1
			vm.ip = ips[site]
			benchmarkBytecodeIndexArraySink, benchmarkBytecodeIndexOKSink = vm.lookupHotArrayIndexSiteWithVersions(
				bytecodeIndexMethodCacheGet,
				receivers[site],
				bytecodeIndexMethodFastPathCanonicalArrayGet,
				true,
				globalRevision,
				methodCacheVersion,
			)
		}
	})
}

func BenchmarkBytecodeVMDirectSmallArrayIndex(b *testing.B) {
	b.Run("raw_i32_hit", func(b *testing.B) {
		value := runtime.Value(bytecodeRawI32SlotCachedValue(7))
		if got, ok := bytecodeDirectSmallArrayIndex(value); !ok || got != 7 {
			b.Fatalf("warm raw i32 index = (%d, %v), want (7, true)", got, ok)
		}
		b.ReportAllocs()
		for idx := 0; idx < b.N; idx++ {
			benchmarkBytecodeIndexIntSink, benchmarkBytecodeIndexOKSink = bytecodeDirectSmallArrayIndex(value)
		}
	})

	b.Run("boxed_i32_hit", func(b *testing.B) {
		value := runtime.Value(runtime.NewSmallInt(7, runtime.IntegerI32))
		if got, ok := bytecodeDirectSmallArrayIndex(value); !ok || got != 7 {
			b.Fatalf("warm boxed i32 index = (%d, %v), want (7, true)", got, ok)
		}
		b.ReportAllocs()
		for idx := 0; idx < b.N; idx++ {
			benchmarkBytecodeIndexIntSink, benchmarkBytecodeIndexOKSink = bytecodeDirectSmallArrayIndex(value)
		}
	})

	b.Run("string_miss", func(b *testing.B) {
		value := runtime.Value(runtime.StringValue{Val: "not an index"})
		if got, ok := bytecodeDirectSmallArrayIndex(value); ok {
			b.Fatalf("warm string index = (%d, true), want miss", got)
		}
		b.ReportAllocs()
		for idx := 0; idx < b.N; idx++ {
			benchmarkBytecodeIndexIntSink, benchmarkBytecodeIndexOKSink = bytecodeDirectSmallArrayIndex(value)
		}
	})

	b.Run("nil_raw_i32_cell_miss", func(b *testing.B) {
		value := runtime.Value((*bytecodeRawI32StackCell)(nil))
		if got, ok := bytecodeDirectSmallArrayIndex(value); ok {
			b.Fatalf("warm nil raw i32 cell index = (%d, true), want miss", got)
		}
		b.ReportAllocs()
		for idx := 0; idx < b.N; idx++ {
			benchmarkBytecodeIndexIntSink, benchmarkBytecodeIndexOKSink = bytecodeDirectSmallArrayIndex(value)
		}
	})
}

func BenchmarkBytecodeVMDirectArrayIndex(b *testing.B) {
	b.Run("raw_i32_hit", func(b *testing.B) {
		value := runtime.Value(bytecodeRawI32SlotCachedValue(7))
		if got, ok, err := bytecodeDirectArrayIndex(value); err != nil || !ok || got != 7 {
			b.Fatalf("warm raw i32 index = (%d, %v, %v), want (7, true, nil)", got, ok, err)
		}
		b.ReportAllocs()
		for idx := 0; idx < b.N; idx++ {
			benchmarkBytecodeIndexIntSink, benchmarkBytecodeIndexOKSink, benchmarkBytecodeIndexErrSink = bytecodeDirectArrayIndex(value)
		}
	})

	b.Run("boxed_i32_hit", func(b *testing.B) {
		value := runtime.Value(runtime.NewSmallInt(7, runtime.IntegerI32))
		if got, ok, err := bytecodeDirectArrayIndex(value); err != nil || !ok || got != 7 {
			b.Fatalf("warm boxed i32 index = (%d, %v, %v), want (7, true, nil)", got, ok, err)
		}
		b.ReportAllocs()
		for idx := 0; idx < b.N; idx++ {
			benchmarkBytecodeIndexIntSink, benchmarkBytecodeIndexOKSink, benchmarkBytecodeIndexErrSink = bytecodeDirectArrayIndex(value)
		}
	})

	b.Run("interface_boxed_i32_hit", func(b *testing.B) {
		value := runtime.Value(runtime.InterfaceValue{Underlying: runtime.NewSmallInt(7, runtime.IntegerI32)})
		if got, ok, err := bytecodeDirectArrayIndex(value); err != nil || !ok || got != 7 {
			b.Fatalf("warm interface boxed i32 index = (%d, %v, %v), want (7, true, nil)", got, ok, err)
		}
		b.ReportAllocs()
		for idx := 0; idx < b.N; idx++ {
			benchmarkBytecodeIndexIntSink, benchmarkBytecodeIndexOKSink, benchmarkBytecodeIndexErrSink = bytecodeDirectArrayIndex(value)
		}
	})

	b.Run("string_miss", func(b *testing.B) {
		value := runtime.Value(runtime.StringValue{Val: "not an index"})
		if got, ok, err := bytecodeDirectArrayIndex(value); err != nil || ok {
			b.Fatalf("warm string index = (%d, %v, %v), want miss", got, ok, err)
		}
		b.ReportAllocs()
		for idx := 0; idx < b.N; idx++ {
			benchmarkBytecodeIndexIntSink, benchmarkBytecodeIndexOKSink, benchmarkBytecodeIndexErrSink = bytecodeDirectArrayIndex(value)
		}
	})
}

func BenchmarkBytecodeVMSlotDirectSmallArrayIndexValidated(b *testing.B) {
	b.Run("slot_raw_i32_hit", func(b *testing.B) {
		vm := &bytecodeVM{slots: []runtime.Value{bytecodeRawI32SlotCachedValue(7)}}
		if got, ok := vm.slotDirectSmallArrayIndexValidated(0); !ok || got != 7 {
			b.Fatalf("warm slot raw i32 index = (%d, %v), want (7, true)", got, ok)
		}
		b.ReportAllocs()
		for idx := 0; idx < b.N; idx++ {
			benchmarkBytecodeIndexIntSink, benchmarkBytecodeIndexOKSink = vm.slotDirectSmallArrayIndexValidated(0)
		}
	})

	b.Run("slot_boxed_i32_hit", func(b *testing.B) {
		vm := &bytecodeVM{slots: []runtime.Value{runtime.NewSmallInt(7, runtime.IntegerI32)}}
		if got, ok := vm.slotDirectSmallArrayIndexValidated(0); !ok || got != 7 {
			b.Fatalf("warm slot boxed i32 index = (%d, %v), want (7, true)", got, ok)
		}
		b.ReportAllocs()
		for idx := 0; idx < b.N; idx++ {
			benchmarkBytecodeIndexIntSink, benchmarkBytecodeIndexOKSink = vm.slotDirectSmallArrayIndexValidated(0)
		}
	})

	b.Run("i32_register_hit", func(b *testing.B) {
		vm := &bytecodeVM{
			slots:            []runtime.Value{nil},
			i32Registers:     []int32{7},
			i32RegisterValid: []bool{true},
		}
		if got, ok := vm.slotDirectSmallArrayIndexValidated(0); !ok || got != 7 {
			b.Fatalf("warm i32 register index = (%d, %v), want (7, true)", got, ok)
		}
		b.ReportAllocs()
		for idx := 0; idx < b.N; idx++ {
			benchmarkBytecodeIndexIntSink, benchmarkBytecodeIndexOKSink = vm.slotDirectSmallArrayIndexValidated(0)
		}
	})

	b.Run("active_value_slot_i32_hit", func(b *testing.B) {
		slots := []runtime.Value{nil}
		vm := &bytecodeVM{
			slots:         slots,
			slotI32Values: []int32{7},
			slotI32Valid:  []bool{true},
			slotI32Owner:  bytecodeSlotFrameOwner(slots),
		}
		if got, ok := vm.slotDirectSmallArrayIndexValidated(0); !ok || got != 7 {
			b.Fatalf("warm active value-slot i32 index = (%d, %v), want (7, true)", got, ok)
		}
		b.ReportAllocs()
		for idx := 0; idx < b.N; idx++ {
			benchmarkBytecodeIndexIntSink, benchmarkBytecodeIndexOKSink = vm.slotDirectSmallArrayIndexValidated(0)
		}
	})

	b.Run("nil_slot_miss", func(b *testing.B) {
		vm := &bytecodeVM{slots: []runtime.Value{nil}}
		if got, ok := vm.slotDirectSmallArrayIndexValidated(0); ok {
			b.Fatalf("warm nil slot index = (%d, true), want miss", got)
		}
		b.ReportAllocs()
		for idx := 0; idx < b.N; idx++ {
			benchmarkBytecodeIndexIntSink, benchmarkBytecodeIndexOKSink = vm.slotDirectSmallArrayIndexValidated(0)
		}
	})
}

func BenchmarkBytecodeVMSlotDirectArrayIndexValidated(b *testing.B) {
	b.Run("slot_raw_i32_hit", func(b *testing.B) {
		vm := &bytecodeVM{slots: []runtime.Value{bytecodeRawI32SlotCachedValue(7)}}
		if got, ok, err := vm.slotDirectArrayIndexValidated(0); err != nil || !ok || got != 7 {
			b.Fatalf("warm slot raw i32 index = (%d, %v, %v), want (7, true, nil)", got, ok, err)
		}
		b.ReportAllocs()
		for idx := 0; idx < b.N; idx++ {
			benchmarkBytecodeIndexIntSink, benchmarkBytecodeIndexOKSink, benchmarkBytecodeIndexErrSink = vm.slotDirectArrayIndexValidated(0)
		}
	})

	b.Run("slot_boxed_i32_hit", func(b *testing.B) {
		vm := &bytecodeVM{slots: []runtime.Value{runtime.NewSmallInt(7, runtime.IntegerI32)}}
		if got, ok, err := vm.slotDirectArrayIndexValidated(0); err != nil || !ok || got != 7 {
			b.Fatalf("warm slot boxed i32 index = (%d, %v, %v), want (7, true, nil)", got, ok, err)
		}
		b.ReportAllocs()
		for idx := 0; idx < b.N; idx++ {
			benchmarkBytecodeIndexIntSink, benchmarkBytecodeIndexOKSink, benchmarkBytecodeIndexErrSink = vm.slotDirectArrayIndexValidated(0)
		}
	})

	b.Run("slot_interface_boxed_i32_hit", func(b *testing.B) {
		vm := &bytecodeVM{slots: []runtime.Value{runtime.InterfaceValue{Underlying: runtime.NewSmallInt(7, runtime.IntegerI32)}}}
		if got, ok, err := vm.slotDirectArrayIndexValidated(0); err != nil || !ok || got != 7 {
			b.Fatalf("warm slot interface boxed i32 index = (%d, %v, %v), want (7, true, nil)", got, ok, err)
		}
		b.ReportAllocs()
		for idx := 0; idx < b.N; idx++ {
			benchmarkBytecodeIndexIntSink, benchmarkBytecodeIndexOKSink, benchmarkBytecodeIndexErrSink = vm.slotDirectArrayIndexValidated(0)
		}
	})

	b.Run("i32_register_hit", func(b *testing.B) {
		vm := &bytecodeVM{
			slots:            []runtime.Value{nil},
			i32Registers:     []int32{7},
			i32RegisterValid: []bool{true},
		}
		if got, ok, err := vm.slotDirectArrayIndexValidated(0); err != nil || !ok || got != 7 {
			b.Fatalf("warm i32 register index = (%d, %v, %v), want (7, true, nil)", got, ok, err)
		}
		b.ReportAllocs()
		for idx := 0; idx < b.N; idx++ {
			benchmarkBytecodeIndexIntSink, benchmarkBytecodeIndexOKSink, benchmarkBytecodeIndexErrSink = vm.slotDirectArrayIndexValidated(0)
		}
	})

	b.Run("active_value_slot_i32_hit", func(b *testing.B) {
		slots := []runtime.Value{nil}
		vm := &bytecodeVM{
			slots:         slots,
			slotI32Values: []int32{7},
			slotI32Valid:  []bool{true},
			slotI32Owner:  bytecodeSlotFrameOwner(slots),
		}
		if got, ok, err := vm.slotDirectArrayIndexValidated(0); err != nil || !ok || got != 7 {
			b.Fatalf("warm active value-slot i32 index = (%d, %v, %v), want (7, true, nil)", got, ok, err)
		}
		b.ReportAllocs()
		for idx := 0; idx < b.N; idx++ {
			benchmarkBytecodeIndexIntSink, benchmarkBytecodeIndexOKSink, benchmarkBytecodeIndexErrSink = vm.slotDirectArrayIndexValidated(0)
		}
	})

	b.Run("nil_slot_miss", func(b *testing.B) {
		vm := &bytecodeVM{slots: []runtime.Value{nil}}
		if got, ok, err := vm.slotDirectArrayIndexValidated(0); err != nil || ok {
			b.Fatalf("warm nil slot index = (%d, %v, %v), want miss", got, ok, err)
		}
		b.ReportAllocs()
		for idx := 0; idx < b.N; idx++ {
			benchmarkBytecodeIndexIntSink, benchmarkBytecodeIndexOKSink, benchmarkBytecodeIndexErrSink = vm.slotDirectArrayIndexValidated(0)
		}
	})
}
