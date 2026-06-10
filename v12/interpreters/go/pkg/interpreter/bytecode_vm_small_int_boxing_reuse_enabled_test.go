//go:build able_bytecode_box_reuse

package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeDynamicIntBoxCacheReuseRecordsLookupHitAndInsert(t *testing.T) {
	bytecodeResetDynamicIntBoxCacheReuseForTest()
	value := bytecodeI32ExtendedBoxMax + 9_876_543
	if _, ok := bytecodeBoxedIntegerValue(runtime.IntegerI32, value); !ok {
		t.Fatal("first dynamic i32 box failed")
	}
	if _, ok := bytecodeBoxedIntegerValue(runtime.IntegerI32, value); !ok {
		t.Fatal("second dynamic i32 box failed")
	}
	stats := bytecodeDynamicIntBoxCacheReuseForTest()[string(runtime.IntegerI32)]
	if stats.Lookups != 2 || stats.Hits != 1 || stats.Inserts != 1 || stats.CapacityMisses != 0 {
		t.Fatalf("unexpected dynamic i32 reuse snapshot: %#v", stats)
	}
}

func TestBytecodeDynamicIntBoxCacheReuseRecordsI64Bypass(t *testing.T) {
	bytecodeResetDynamicIntBoxCacheReuseForTest()
	if _, ok := bytecodeBoxedIntegerValue(runtime.IntegerI64, bytecodeI32ExtendedBoxMax+1); !ok {
		t.Fatal("large i64 box failed")
	}
	stats := bytecodeDynamicIntBoxCacheReuseForTest()[string(runtime.IntegerI64)]
	if stats.I64Bypasses != 1 || stats.Lookups != 0 || stats.Hits != 0 || stats.Inserts != 0 || stats.CapacityMisses != 0 {
		t.Fatalf("unexpected i64 bypass snapshot: %#v", stats)
	}
}
