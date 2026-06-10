package interpreter

import (
	"fmt"
	"reflect"
	"testing"

	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_ConcurrentCallsKeepRawIntegerFramesIndependent(t *testing.T) {
	const source = `
package demo

fn churn(seed: i64) -> i64 {
  value := seed
  index := 0_i32
  loop {
    if index >= 2000_i32 { break }
    value = (value * 31_i64 + (index as i64) * 17_i64 + seed) % 1_000_003_i64
    index = index + 1_i32
  }
  value
}

fn main() -> void {
  first := spawn { churn(11_i64) }
  second := spawn { churn(23_i64) }
  third := spawn { churn(47_i64) }
  fourth := spawn { churn(89_i64) }
  future_flush()
  print(first.value()! as i64)
  print(second.value()! as i64)
  print(third.value()! as i64)
  print(fourth.value()! as i64)
}
`
	want := make([]string, 0, 4)
	for _, seed := range []int64{11, 23, 47, 89} {
		value := seed
		for index := int64(0); index < 2000; index++ {
			value = (value*31 + index*17 + seed) % 1_000_003
		}
		want = append(want, fmt.Sprint(value))
	}

	for run := 0; run < 12; run++ {
		got, err := runStdlibExecSourceWithExecutor(
			t,
			source,
			testExecBytecode,
			NewGoroutineExecutor(nil),
		)
		if err != nil {
			t.Fatalf("run %d failed: %v", run, err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("run %d stdout mismatch: got %v want %v", run, got, want)
		}
	}
}

func TestBytecodeVMMultiThreadRawIntegerCarriersAreImmutableValues(t *testing.T) {
	interp := NewBytecodeWithExecutor(NewGoroutineExecutor(nil))
	interp.ensureMultiThread()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.slots = make([]runtime.Value, 2)

	if got := vm.storeRawI64Slot(0, 41); reflect.TypeOf(got).Kind() == reflect.Pointer {
		t.Fatalf("multi-thread i64 slot carrier is mutable pointer %T", got)
	}
	if got := vm.stackRawI64Value(0, 42); reflect.TypeOf(got).Kind() == reflect.Pointer {
		t.Fatalf("multi-thread i64 stack carrier is mutable pointer %T", got)
	}
	if got := vm.storeRawIntegerSlot(1, runtime.IntegerIsize, 43); reflect.TypeOf(got).Kind() == reflect.Pointer {
		t.Fatalf("multi-thread isize slot carrier is mutable pointer %T", got)
	}
	if got := vm.stackRawI32Value(1, bytecodeRawI32SlotCacheMax+17); reflect.TypeOf(got).Kind() == reflect.Pointer {
		t.Fatalf("multi-thread i32 stack carrier is mutable pointer %T", got)
	}
	if got := vm.rawIntegerReturnValue(runtime.IntegerU16, 44); reflect.TypeOf(got).Kind() == reflect.Pointer {
		t.Fatalf("multi-thread return carrier is mutable pointer %T", got)
	}
}
