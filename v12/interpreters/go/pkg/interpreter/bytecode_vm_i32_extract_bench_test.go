package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/runtime"
)

var (
	benchmarkBytecodeI32ExtractIntSink int64
	benchmarkBytecodeI32ExtractOKSink  bool
	benchmarkBytecodeIntegerKindSink   runtime.IntegerType
)

func BenchmarkBytecodeRawIntegerValueInfo(b *testing.B) {
	for _, tc := range []struct {
		name  string
		value runtime.Value
	}{
		{name: "boxed_i32", value: runtime.NewSmallInt(7, runtime.IntegerI32)},
		{name: "raw_i32", value: bytecodeRawI32SlotCachedValue(7)},
		{name: "raw_u64", value: bytecodeRawIntegerResultValue(runtime.IntegerU64, 7)},
		{name: "string_miss", value: runtime.StringValue{Val: "not an integer"}},
	} {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for idx := 0; idx < b.N; idx++ {
				benchmarkBytecodeIntegerKindSink, benchmarkBytecodeI32ExtractIntSink, benchmarkBytecodeI32ExtractOKSink = bytecodeRawIntegerValueInfo(tc.value)
			}
		})
	}
}

func BenchmarkBytecodeVMDirectSmallI32Value(b *testing.B) {
	b.Run("raw_i32_hit", func(b *testing.B) {
		value := runtime.Value(bytecodeRawI32SlotCachedValue(7))
		if got, ok := bytecodeDirectSmallI32Value(value); !ok || got != 7 {
			b.Fatalf("warm raw i32 = (%d, %v), want (7, true)", got, ok)
		}
		b.ReportAllocs()
		for idx := 0; idx < b.N; idx++ {
			benchmarkBytecodeI32ExtractIntSink, benchmarkBytecodeI32ExtractOKSink = bytecodeDirectSmallI32Value(value)
		}
	})

	b.Run("raw_i32_cell_hit", func(b *testing.B) {
		value := runtime.Value(&bytecodeRawI32StackCell{Val: 7})
		if got, ok := bytecodeDirectSmallI32Value(value); !ok || got != 7 {
			b.Fatalf("warm raw i32 cell = (%d, %v), want (7, true)", got, ok)
		}
		b.ReportAllocs()
		for idx := 0; idx < b.N; idx++ {
			benchmarkBytecodeI32ExtractIntSink, benchmarkBytecodeI32ExtractOKSink = bytecodeDirectSmallI32Value(value)
		}
	})

	b.Run("boxed_i32_hit", func(b *testing.B) {
		value := runtime.Value(runtime.NewSmallInt(7, runtime.IntegerI32))
		if got, ok := bytecodeDirectSmallI32Value(value); !ok || got != 7 {
			b.Fatalf("warm boxed i32 = (%d, %v), want (7, true)", got, ok)
		}
		b.ReportAllocs()
		for idx := 0; idx < b.N; idx++ {
			benchmarkBytecodeI32ExtractIntSink, benchmarkBytecodeI32ExtractOKSink = bytecodeDirectSmallI32Value(value)
		}
	})

	b.Run("boxed_i32_pointer_hit", func(b *testing.B) {
		integer := runtime.NewSmallInt(7, runtime.IntegerI32)
		value := runtime.Value(&integer)
		if got, ok := bytecodeDirectSmallI32Value(value); !ok || got != 7 {
			b.Fatalf("warm boxed i32 pointer = (%d, %v), want (7, true)", got, ok)
		}
		b.ReportAllocs()
		for idx := 0; idx < b.N; idx++ {
			benchmarkBytecodeI32ExtractIntSink, benchmarkBytecodeI32ExtractOKSink = bytecodeDirectSmallI32Value(value)
		}
	})

	b.Run("boxed_i64_miss", func(b *testing.B) {
		value := runtime.Value(runtime.NewSmallInt(7, runtime.IntegerI64))
		if got, ok := bytecodeDirectSmallI32Value(value); ok {
			b.Fatalf("warm boxed i64 = (%d, true), want miss", got)
		}
		b.ReportAllocs()
		for idx := 0; idx < b.N; idx++ {
			benchmarkBytecodeI32ExtractIntSink, benchmarkBytecodeI32ExtractOKSink = bytecodeDirectSmallI32Value(value)
		}
	})

	b.Run("string_miss", func(b *testing.B) {
		value := runtime.Value(runtime.StringValue{Val: "not i32"})
		if got, ok := bytecodeDirectSmallI32Value(value); ok {
			b.Fatalf("warm string = (%d, true), want miss", got)
		}
		b.ReportAllocs()
		for idx := 0; idx < b.N; idx++ {
			benchmarkBytecodeI32ExtractIntSink, benchmarkBytecodeI32ExtractOKSink = bytecodeDirectSmallI32Value(value)
		}
	})
}

func BenchmarkBytecodeVMSlotDirectSmallI32ValueValidated(b *testing.B) {
	b.Run("slot_raw_i32_hit", func(b *testing.B) {
		vm := &bytecodeVM{slots: []runtime.Value{bytecodeRawI32SlotCachedValue(7)}}
		if got, ok := vm.slotDirectSmallI32ValueValidated(0); !ok || got != 7 {
			b.Fatalf("warm slot raw i32 = (%d, %v), want (7, true)", got, ok)
		}
		b.ReportAllocs()
		for idx := 0; idx < b.N; idx++ {
			benchmarkBytecodeI32ExtractIntSink, benchmarkBytecodeI32ExtractOKSink = vm.slotDirectSmallI32ValueValidated(0)
		}
	})

	b.Run("slot_boxed_i32_hit", func(b *testing.B) {
		vm := &bytecodeVM{slots: []runtime.Value{runtime.NewSmallInt(7, runtime.IntegerI32)}}
		if got, ok := vm.slotDirectSmallI32ValueValidated(0); !ok || got != 7 {
			b.Fatalf("warm slot boxed i32 = (%d, %v), want (7, true)", got, ok)
		}
		b.ReportAllocs()
		for idx := 0; idx < b.N; idx++ {
			benchmarkBytecodeI32ExtractIntSink, benchmarkBytecodeI32ExtractOKSink = vm.slotDirectSmallI32ValueValidated(0)
		}
	})

	b.Run("i32_register_hit", func(b *testing.B) {
		vm := &bytecodeVM{
			slots:            []runtime.Value{nil},
			i32Registers:     []int32{7},
			i32RegisterValid: []bool{true},
		}
		if got, ok := vm.slotDirectSmallI32ValueValidated(0); !ok || got != 7 {
			b.Fatalf("warm i32 register = (%d, %v), want (7, true)", got, ok)
		}
		b.ReportAllocs()
		for idx := 0; idx < b.N; idx++ {
			benchmarkBytecodeI32ExtractIntSink, benchmarkBytecodeI32ExtractOKSink = vm.slotDirectSmallI32ValueValidated(0)
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
		if got, ok := vm.slotDirectSmallI32ValueValidated(0); !ok || got != 7 {
			b.Fatalf("warm active value-slot i32 = (%d, %v), want (7, true)", got, ok)
		}
		b.ReportAllocs()
		for idx := 0; idx < b.N; idx++ {
			benchmarkBytecodeI32ExtractIntSink, benchmarkBytecodeI32ExtractOKSink = vm.slotDirectSmallI32ValueValidated(0)
		}
	})

	b.Run("nil_slot_miss", func(b *testing.B) {
		vm := &bytecodeVM{slots: []runtime.Value{nil}}
		if got, ok := vm.slotDirectSmallI32ValueValidated(0); ok {
			b.Fatalf("warm nil slot = (%d, true), want miss", got)
		}
		b.ReportAllocs()
		for idx := 0; idx < b.N; idx++ {
			benchmarkBytecodeI32ExtractIntSink, benchmarkBytecodeI32ExtractOKSink = vm.slotDirectSmallI32ValueValidated(0)
		}
	})
}
